// Package custom manages the quests a player writes themselves. They are
// stored as a regular quest pack, so they can later be exported and shared.
package custom

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"SideQuest/internal/quest"
)

const (
	packID    = "custom"
	packName  = "My Quests"
	MaxQuests = 100

	maxTitle = 60
	maxDesc  = 220

	// DefaultDescription is used when the player leaves the details empty.
	DefaultDescription = "A quest you made yourself. Go for it!"
)

// Draft is what the player fills in. ID is empty for a new quest.
type Draft struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Category    quest.Category `json:"category"`
	Minutes     int            `json:"minutes"`
	Difficulty  int            `json:"difficulty"`
}

// Errors shown in the editor. The text is a translation key for the UI.
var (
	ErrNoTitle      = errors.New("custom.noTitle")
	ErrTitleLong    = errors.New("custom.titleLong")
	ErrDescLong     = errors.New("custom.descLong")
	ErrBadCategory  = errors.New("custom.badCategory")
	ErrBadMinutes   = errors.New("custom.badMinutes")
	ErrBadDifficult = errors.New("custom.badDifficulty")
	ErrTooMany      = errors.New("custom.tooMany")
	ErrNotFound     = errors.New("custom.notFound")
)

// Persister stores the pack.
type Persister interface {
	Load(v any) (bool, error)
	Save(v any) error
}

// Library holds the player's quests. It is safe for concurrent use.
type Library struct {
	mu    sync.Mutex
	store Persister
	pack  quest.Pack
}

// Open loads the library. A missing or invalid file starts an empty library
// (and returns the problem for logging).
func Open(store Persister) (*Library, error) {
	l := &Library{store: store, pack: emptyPack()}
	var p quest.Pack
	ok, err := store.Load(&p)
	if err != nil || !ok {
		return l, err
	}
	// Re-validate on load: the file is user-editable.
	var valid []quest.Quest
	var bad error
	for _, q := range p.Quests {
		q.Source, q.Pack = quest.SourceCustom, packID
		if verr := quest.Validate(&q); verr != nil {
			bad = verr
			continue
		}
		valid = append(valid, q)
	}
	l.pack.Quests = valid
	return l, bad
}

func emptyPack() quest.Pack {
	return quest.Pack{Format: quest.PackFormat, ID: packID, Name: packName, Quests: []quest.Quest{}}
}

// List returns the player's quests, newest first.
func (l *Library) List() []quest.Quest {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]quest.Quest, len(l.pack.Quests))
	for i, q := range l.pack.Quests {
		out[len(out)-1-i] = q
	}
	return out
}

// Save creates or updates a quest and persists the library.
func (l *Library) Save(d Draft) (quest.Quest, error) {
	q, err := build(d)
	if err != nil {
		return quest.Quest{}, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if d.ID == "" {
		if len(l.pack.Quests) >= MaxQuests {
			return quest.Quest{}, ErrTooMany
		}
		q.ID = newID(q.Title)
		l.pack.Quests = append(l.pack.Quests, q)
	} else {
		i := l.index(d.ID)
		if i < 0 {
			return quest.Quest{}, ErrNotFound
		}
		q.ID = d.ID
		l.pack.Quests[i] = q
	}
	return q, l.store.Save(&l.pack)
}

// Delete removes a quest and persists the library.
func (l *Library) Delete(id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	i := l.index(id)
	if i < 0 {
		return ErrNotFound
	}
	l.pack.Quests = append(l.pack.Quests[:i], l.pack.Quests[i+1:]...)
	return l.store.Save(&l.pack)
}

func (l *Library) index(id string) int {
	for i, q := range l.pack.Quests {
		if q.ID == id {
			return i
		}
	}
	return -1
}

// Reward is the XP a custom quest gives. It is derived, not chosen, so
// writing quests never becomes an XP shortcut.
func Reward(minutes, difficulty int) int {
	return min(40, 8+2*minutes+6*(difficulty-1))
}

func build(d Draft) (quest.Quest, error) {
	title := strings.Join(strings.Fields(d.Title), " ")
	desc := strings.TrimSpace(d.Description)
	switch {
	case title == "":
		return quest.Quest{}, ErrNoTitle
	case utf8.RuneCountInString(title) > maxTitle:
		return quest.Quest{}, ErrTitleLong
	case utf8.RuneCountInString(desc) > maxDesc:
		return quest.Quest{}, ErrDescLong
	case !d.Category.Valid():
		return quest.Quest{}, ErrBadCategory
	case d.Minutes < 1 || d.Minutes > 15:
		return quest.Quest{}, ErrBadMinutes
	case d.Difficulty < 1 || d.Difficulty > 3:
		return quest.Quest{}, ErrBadDifficult
	}
	if desc == "" {
		desc = DefaultDescription
	}
	q := quest.Quest{
		ID:            "my-placeholder",
		Title:         title,
		Description:   desc,
		Category:      d.Category,
		Minutes:       d.Minutes,
		Difficulty:    d.Difficulty,
		XP:            Reward(d.Minutes, d.Difficulty),
		Repeatable:    true,
		CooldownHours: 12,
		Source:        quest.SourceCustom,
		Pack:          packID,
	}
	// Belt and braces: custom quests obey the same rules as any pack.
	if err := quest.Validate(&q); err != nil {
		return quest.Quest{}, err
	}
	return q, nil
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// newID makes a readable, unique id such as "my-sock-puppet-3f9a1c".
func newID(title string) string {
	slug := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if len(slug) > 24 {
		slug = strings.TrimRight(slug[:24], "-")
	}
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	if slug == "" {
		return "my-" + hex.EncodeToString(b)
	}
	return "my-" + slug + "-" + hex.EncodeToString(b)
}
