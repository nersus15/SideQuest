package journal

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"SideQuest/internal/quest"
	"SideQuest/internal/store"
)

var at = time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)

func open(t *testing.T, path string) *Journal {
	t.Helper()
	j, err := Open(store.File{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func sample(cat quest.Category) quest.Quest {
	return quest.Quest{
		ID: "q-" + string(cat), Title: "Hello", Category: cat,
		Interaction: quest.Interaction{Kind: quest.KindAnswer, Prompts: []string{"A"}},
		I18n:        map[string]quest.Translation{"id": {Title: "Halo"}},
	}
}

func TestAddListAndPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.json")
	j := open(t, path)
	e, err := j.Add(sample(quest.Chill), 12, at, Input{Answers: []string{"  sunny  ", ""}, Count: 3})
	if err != nil {
		t.Fatal(err)
	}
	if e.Titles["en"] != "Hello" || e.Titles["id"] != "Halo" || len(e.Answers) != 1 || e.Answers[0] != "sunny" {
		t.Fatalf("unexpected entry: %+v", e)
	}
	if err := j.SetMood(e.ID, "great"); err != nil {
		t.Fatal(err)
	}
	got := open(t, path).List()
	if len(got) != 1 || got[0].Mood != "great" || got[0].Count != 3 {
		t.Fatalf("journal not persisted: %+v", got)
	}
}

func TestRejectsBadInput(t *testing.T) {
	j := open(t, filepath.Join(t.TempDir(), "j.json"))
	bad := []Input{
		{Answers: []string{strings.Repeat("a", 141)}},
		{Answers: make([]string, 6)},
		{Doodle: "javascript:alert(1)"},
		{Doodle: doodlePrefix + strings.Repeat("A", maxDoodleBytes)},
		{Count: -1},
	}
	for i, in := range bad {
		if _, err := j.Add(sample(quest.Play), 10, at, in); !errors.Is(err, ErrBadInput) {
			t.Errorf("case %d: expected ErrBadInput, got %v", i, err)
		}
	}
	if err := j.SetMood("nope", "great"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	e, _ := j.Add(sample(quest.Play), 10, at, Input{})
	if err := j.SetMood(e.ID, "ecstatic"); !errors.Is(err, ErrBadMood) {
		t.Errorf("expected ErrBadMood, got %v", err)
	}
}

func TestOldEntriesAndDoodlesAreTrimmed(t *testing.T) {
	j := open(t, filepath.Join(t.TempDir(), "j.json"))
	for i := 0; i < MaxEntries+5; i++ {
		if _, err := j.Add(sample(quest.Creative), 10, at, Input{Doodle: doodlePrefix + "AAAA"}); err != nil {
			t.Fatal(err)
		}
	}
	list := j.List()
	if len(list) != MaxEntries {
		t.Fatalf("expected %d entries, got %d", MaxEntries, len(list))
	}
	withDoodle := 0
	for _, e := range list {
		if e.Doodle != "" {
			withDoodle++
		}
	}
	if withDoodle != MaxDoodles || list[0].Doodle == "" {
		t.Fatalf("expected the %d newest doodles kept, got %d", MaxDoodles, withDoodle)
	}
}

func TestInsightNeedsEnoughCheckIns(t *testing.T) {
	j := open(t, filepath.Join(t.TempDir(), "j.json"))
	add := func(cat quest.Category, mood string) {
		e, _ := j.Add(sample(cat), 10, at, Input{})
		_ = j.SetMood(e.ID, mood)
	}
	add(quest.Chill, "great")
	add(quest.Chill, "great")
	if got := j.Insight(); got.Category != "" || got.Moods != 2 {
		t.Fatalf("two check-ins should not be enough: %+v", got)
	}
	add(quest.Chill, "good")
	for i := 0; i < 3; i++ {
		add(quest.Play, "meh")
	}
	if got := j.Insight(); got.Category != quest.Chill {
		t.Fatalf("expected chill, got %+v", got)
	}
}
