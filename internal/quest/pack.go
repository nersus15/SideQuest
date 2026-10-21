package quest

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// PackFormat is the only pack format version this build understands.
const PackFormat = 1

// Limits that keep imported packs small and harmless.
const (
	MaxPackBytes      = 512 * 1024
	MaxQuestsPerPack  = 500
	MaxTwistsPerQuest = 6
	maxTitleLen       = 60
	maxDescLen        = 220
	maxTwistLen       = 120
	maxPackNameLen    = 60
)

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,47}$`)

// Pack is a named collection of quests stored as JSON.
type Pack struct {
	Format int     `json:"format"`
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Quests []Quest `json:"quests"`
}

//go:embed builtin/quests.json
var builtinJSON []byte

// Builtin returns the quests that ship with the app.
func Builtin() ([]Quest, error) {
	p, err := ParsePack(bytes.NewReader(builtinJSON), SourceBuiltin)
	if err != nil {
		return nil, fmt.Errorf("builtin pack: %w", err)
	}
	return p.Quests, nil
}

// ParsePack reads, validates and normalises a pack. The source is assigned
// by the caller and never trusted from the file itself.
func ParsePack(r io.Reader, source Source) (Pack, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxPackBytes+1))
	if err != nil {
		return Pack{}, err
	}
	if len(data) > MaxPackBytes {
		return Pack{}, fmt.Errorf("pack is larger than %d KB", MaxPackBytes/1024)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Pack
	if err := dec.Decode(&p); err != nil {
		return Pack{}, fmt.Errorf("invalid pack JSON: %w", err)
	}
	if err := p.normalise(source); err != nil {
		return Pack{}, err
	}
	return p, nil
}

// ValidationError collects every problem found in a pack.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return "invalid quest pack: " + strings.Join(e.Problems, "; ")
}

func (p *Pack) normalise(source Source) error {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if p.Format != PackFormat {
		add("unsupported format %d (expected %d)", p.Format, PackFormat)
	}
	if !idPattern.MatchString(p.ID) {
		add("pack id %q must be 2-48 lowercase letters, digits or dashes", p.ID)
	}
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" || utf8.RuneCountInString(p.Name) > maxPackNameLen {
		add("pack name must be 1-%d characters", maxPackNameLen)
	}
	if len(p.Quests) == 0 {
		add("pack has no quests")
	}
	if len(p.Quests) > MaxQuestsPerPack {
		add("pack has %d quests (max %d)", len(p.Quests), MaxQuestsPerPack)
	}

	seen := map[string]bool{}
	for i := range p.Quests {
		q := &p.Quests[i]
		q.Source = source
		q.Pack = p.ID
		for _, msg := range validateQuest(q) {
			add("quest %d (%s): %s", i+1, q.ID, msg)
		}
		if seen[q.ID] {
			add("duplicate quest id %q", q.ID)
		}
		seen[q.ID] = true
	}

	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}

func validateQuest(q *Quest) []string {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	q.Title = strings.TrimSpace(q.Title)
	q.Description = strings.TrimSpace(q.Description)

	if !idPattern.MatchString(q.ID) {
		add("id must be 2-48 lowercase letters, digits or dashes")
	}
	if n := utf8.RuneCountInString(q.Title); n == 0 || n > maxTitleLen {
		add("title must be 1-%d characters", maxTitleLen)
	}
	if n := utf8.RuneCountInString(q.Description); n == 0 || n > maxDescLen {
		add("description must be 1-%d characters", maxDescLen)
	}
	if !q.Category.Valid() {
		add("unknown category %q", q.Category)
	}
	if q.Minutes < 1 || q.Minutes > 15 {
		add("minutes must be 1-15")
	}
	if q.Difficulty < 1 || q.Difficulty > 3 {
		add("difficulty must be 1-3")
	}
	if q.XP < 5 || q.XP > 100 {
		add("xp must be 5-100")
	}
	if q.CooldownHours < 0 || q.CooldownHours > 24*30 {
		add("cooldownHours must be 0-720")
	}
	needs := map[Need]bool{}
	for _, n := range q.Needs {
		if !knownNeeds[n] {
			add("unknown need %q", n)
		}
		if needs[n] {
			add("duplicate need %q", n)
		}
		needs[n] = true
	}
	if len(q.Twists) > MaxTwistsPerQuest {
		add("too many twists (max %d)", MaxTwistsPerQuest)
	}
	twistIDs := map[string]bool{}
	for j := range q.Twists {
		t := &q.Twists[j]
		t.Text = strings.TrimSpace(t.Text)
		if !idPattern.MatchString(t.ID) {
			add("twist %d id is invalid", j+1)
		}
		if twistIDs[t.ID] {
			add("duplicate twist id %q", t.ID)
		}
		twistIDs[t.ID] = true
		if n := utf8.RuneCountInString(t.Text); n == 0 || n > maxTwistLen {
			add("twist %q text must be 1-%d characters", t.ID, maxTwistLen)
		}
		if t.Bonus < 0 || t.Bonus > 50 {
			add("twist %q bonus must be 0-50", t.ID)
		}
	}
	problems = append(problems, validateInteraction(&q.Interaction)...)
	problems = append(problems, validateI18n(q)...)
	return problems
}

const (
	maxPrompts   = 8
	maxPromptLen = 48
)

func validateInteraction(in *Interaction) []string {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}
	for i := range in.Prompts {
		in.Prompts[i] = strings.TrimSpace(in.Prompts[i])
		if n := utf8.RuneCountInString(in.Prompts[i]); n == 0 || n > maxPromptLen {
			add("prompt %d must be 1-%d characters", i+1, maxPromptLen)
		}
	}
	np := len(in.Prompts)
	switch in.Kind {
	case KindDo, KindDoodle:
		if np > 0 || in.Seconds != 0 || in.Target != 0 || in.Guide != "" {
			add("%q quests take no prompts, seconds or target", in.Kind)
		}
	case KindAnswer:
		if np < 1 || np > 5 {
			add("answer quests need 1-5 prompts")
		}
	case KindChecklist:
		if np < 2 || np > maxPrompts {
			add("checklist quests need 2-%d prompts", maxPrompts)
		}
	case KindTimer:
		if in.Seconds < 5 || in.Seconds > 300 {
			add("timer seconds must be 5-300")
		}
		if in.Guide != "" && in.Guide != "breath" {
			add("unknown timer guide %q", in.Guide)
		}
	case KindCounter:
		if in.Target < 0 || in.Target > 50 {
			add("counter target must be 0-50")
		}
	default:
		add("unknown interaction kind %q", in.Kind)
	}
	if in.Kind != KindAnswer && in.Kind != KindChecklist && np > 0 {
		add("only answer and checklist quests take prompts")
	}
	if in.Kind != KindTimer && (in.Seconds != 0 || in.Guide != "") {
		add("only timer quests take seconds or a guide")
	}
	if in.Kind != KindCounter && in.Target != 0 {
		add("only counter quests take a target")
	}
	return dedupe(problems)
}

func validateI18n(q *Quest) []string {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}
	for lang, tr := range q.I18n {
		if !knownLanguage(lang) {
			add("unknown language %q", lang)
			continue
		}
		tr.Title = strings.TrimSpace(tr.Title)
		tr.Description = strings.TrimSpace(tr.Description)
		if utf8.RuneCountInString(tr.Title) > maxTitleLen {
			add("%s title must be at most %d characters", lang, maxTitleLen)
		}
		if utf8.RuneCountInString(tr.Description) > maxDescLen {
			add("%s description must be at most %d characters", lang, maxDescLen)
		}
		if len(tr.Prompts) > 0 && len(tr.Prompts) != len(q.Interaction.Prompts) {
			add("%s needs exactly %d prompts", lang, len(q.Interaction.Prompts))
		}
		for i := range tr.Prompts {
			tr.Prompts[i] = strings.TrimSpace(tr.Prompts[i])
			if n := utf8.RuneCountInString(tr.Prompts[i]); n == 0 || n > maxPromptLen {
				add("%s prompt %d must be 1-%d characters", lang, i+1, maxPromptLen)
			}
		}
		for id, text := range tr.Twists {
			if _, ok := q.TwistByID(id); !ok {
				add("%s translates unknown twist %q", lang, id)
			}
			text = strings.TrimSpace(text)
			if n := utf8.RuneCountInString(text); n == 0 || n > maxTwistLen {
				add("%s twist %q must be 1-%d characters", lang, id, maxTwistLen)
			}
			tr.Twists[id] = text
		}
		q.I18n[lang] = tr
	}
	return problems
}

func knownLanguage(lang string) bool {
	for _, l := range Languages {
		if l == lang {
			return true
		}
	}
	return false
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// Merge combines quest lists. Earlier lists win when ids collide, so
// imported packs can never replace built-in quests.
func Merge(lists ...[]Quest) []Quest {
	seen := map[string]bool{}
	var out []Quest
	for _, list := range lists {
		for _, q := range list {
			if seen[q.ID] {
				continue
			}
			seen[q.ID] = true
			out = append(out, q)
		}
	}
	return out
}

// ErrNoQuests is returned when no quest can be offered at all.
var ErrNoQuests = errors.New("no quests available")

// LoadDir reads every *.json pack in dir. Invalid packs are skipped and
// reported in problems; a missing dir simply yields no quests.
func LoadDir(dir string, source Source) (quests []Quest, problems []error) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	sort.Strings(files)
	for _, path := range files {
		p, err := loadFile(path, source)
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", filepath.Base(path), err))
			continue
		}
		quests = append(quests, p.Quests...)
	}
	return quests, problems
}

func loadFile(path string, source Source) (Pack, error) {
	f, err := os.Open(path)
	if err != nil {
		return Pack{}, err
	}
	defer f.Close()
	return ParsePack(f, source)
}

// Validate normalises q and reports every problem with it.
func Validate(q *Quest) error {
	if problems := validateQuest(q); len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}
