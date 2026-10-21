// Package journal keeps what the player wrote, drew and felt while doing
// quests. It lives in its own file so the main save stays small, and it
// never leaves the device.
package journal

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"SideQuest/internal/quest"
)

const (
	// MaxEntries is how many entries are kept; the oldest go first.
	MaxEntries = 120
	// MaxDoodles is how many recent entries keep their drawing.
	MaxDoodles = 30

	maxAnswers     = 5
	maxAnswerLen   = 140
	maxDoodleBytes = 200 * 1024
	maxStoryLen    = 400
	maxMillis      = 60 * 60 * 1000
	doodlePrefix   = "data:image/png;base64,"
)

// Moods, from low to great.
var moods = map[string]int{"low": 0, "meh": 1, "good": 2, "great": 3}

// Input is what the player produced while doing a quest.
type Input struct {
	Answers []string `json:"answers"`
	Count   int      `json:"count"`
	Doodle  string   `json:"doodle"` // PNG data URL
	Story   string   `json:"story"`  // a filled-in madlib
	Millis  int      `json:"millis"` // a stopwatch time
}

// Entry is one completed quest.
type Entry struct {
	ID       string            `json:"id"`
	QuestID  string            `json:"questId"`
	Titles   map[string]string `json:"titles"` // by language
	Category quest.Category    `json:"category"`
	Kind     string            `json:"kind"`
	XP       int               `json:"xp"`
	At       time.Time         `json:"at"`
	Answers  []string          `json:"answers,omitempty"`
	Count    int               `json:"count,omitempty"`
	Doodle   string            `json:"doodle,omitempty"`
	Story    string            `json:"story,omitempty"`
	Millis   int               `json:"millis,omitempty"`
	Mood     string            `json:"mood,omitempty"`
}

// Errors shown to the player as translation keys.
var (
	ErrBadInput = errors.New("journal.badInput")
	ErrBadMood  = errors.New("journal.badMood")
	ErrNotFound = errors.New("journal.notFound")
)

// Persister stores the journal.
type Persister interface {
	Load(v any) (bool, error)
	Save(v any) error
}

type file struct {
	Entries []Entry `json:"entries"`
}

// Journal is safe for concurrent use.
type Journal struct {
	mu    sync.Mutex
	store Persister
	data  file
}

// Open loads the journal; a missing or unreadable file starts empty.
func Open(store Persister) (*Journal, error) {
	j := &Journal{store: store}
	_, err := store.Load(&j.data)
	if j.data.Entries == nil {
		j.data.Entries = []Entry{}
	}
	return j, err
}

// Clean validates and trims player input. Unknown or oversized content is
// rejected rather than silently stored.
func Clean(in Input) (Input, error) {
	if len(in.Answers) > maxAnswers || in.Count < 0 || in.Count > 999 {
		return Input{}, ErrBadInput
	}
	var answers []string
	for _, a := range in.Answers {
		a = strings.TrimSpace(a)
		if utf8.RuneCountInString(a) > maxAnswerLen {
			return Input{}, ErrBadInput
		}
		answers = append(answers, a)
	}
	if in.Doodle != "" && (!strings.HasPrefix(in.Doodle, doodlePrefix) || len(in.Doodle) > maxDoodleBytes) {
		return Input{}, ErrBadInput
	}
	story := strings.TrimSpace(in.Story)
	if utf8.RuneCountInString(story) > maxStoryLen || in.Millis < 0 || in.Millis > maxMillis {
		return Input{}, ErrBadInput
	}
	return Input{Answers: answers, Count: in.Count, Doodle: in.Doodle, Story: story, Millis: in.Millis}, nil
}

// Add records a completed quest and returns the new entry.
func (j *Journal) Add(q quest.Quest, xp int, at time.Time, in Input) (Entry, error) {
	in, err := Clean(in)
	if err != nil {
		return Entry{}, err
	}
	titles := map[string]string{"en": q.Title}
	for lang, tr := range q.I18n {
		if tr.Title != "" {
			titles[lang] = tr.Title
		}
	}
	e := Entry{
		ID:       newID(),
		QuestID:  q.ID,
		Titles:   titles,
		Category: q.Category,
		Kind:     q.Interaction.Kind,
		XP:       xp,
		At:       at,
		Count:    in.Count,
		Doodle:   in.Doodle,
		Story:    in.Story,
		Millis:   in.Millis,
	}
	for _, a := range in.Answers {
		if a != "" {
			e.Answers = append(e.Answers, a)
		}
	}

	j.mu.Lock()
	defer j.mu.Unlock()
	j.data.Entries = append(j.data.Entries, e)
	if n := len(j.data.Entries); n > MaxEntries {
		j.data.Entries = j.data.Entries[n-MaxEntries:]
	}
	// Drawings are the bulk of the file; keep only recent ones.
	for i := 0; i < len(j.data.Entries)-MaxDoodles; i++ {
		j.data.Entries[i].Doodle = ""
	}
	return e, j.store.Save(&j.data)
}

// SetMood attaches how the player felt after a quest ("" clears it).
func (j *Journal) SetMood(id, mood string) error {
	if _, ok := moods[mood]; !ok && mood != "" {
		return ErrBadMood
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	for i := range j.data.Entries {
		if j.data.Entries[i].ID == id {
			j.data.Entries[i].Mood = mood
			return j.store.Save(&j.data)
		}
	}
	return ErrNotFound
}

// List returns entries, newest first.
func (j *Journal) List() []Entry {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]Entry, len(j.data.Entries))
	for i, e := range j.data.Entries {
		out[len(out)-1-i] = e
	}
	return out
}

// Insight is a gentle observation from the player's own mood check-ins.
// It is not an assessment of any kind.
type Insight struct {
	Category quest.Category `json:"category"` // "" when there is not enough data
	Moods    int            `json:"moods"`    // check-ins so far
}

// minMoods is how many check-ins a category needs before it counts.
const minMoods = 3

// Insight finds the kind of quest after which the player tends to feel best.
func (j *Journal) Insight() Insight {
	j.mu.Lock()
	defer j.mu.Unlock()
	sum := map[quest.Category]int{}
	count := map[quest.Category]int{}
	total := 0
	for _, e := range j.data.Entries {
		score, ok := moods[e.Mood]
		if !ok {
			continue
		}
		sum[e.Category] += score
		count[e.Category]++
		total++
	}
	best := Insight{Moods: total}
	bestAvg := 0.0
	for _, c := range quest.Categories { // fixed order keeps ties stable
		if count[c] < minMoods {
			continue
		}
		avg := float64(sum[c]) / float64(count[c])
		if avg >= 2 && avg > bestAvg {
			best.Category, bestAvg = c, avg
		}
	}
	return best
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
