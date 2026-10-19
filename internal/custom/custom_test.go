package custom

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"SideQuest/internal/quest"
	"SideQuest/internal/store"
)

func open(t *testing.T, path string) *Library {
	t.Helper()
	l, err := Open(store.File{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func draft() Draft {
	return Draft{Title: "  Sock   puppet show ", Category: quest.Wildcard, Minutes: 3, Difficulty: 2}
}

func TestSaveListDeletePersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.json")
	l := open(t, path)

	q, err := l.Save(draft())
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^my-sock-puppet-show-[0-9a-f]{6}$`).MatchString(q.ID) {
		t.Fatalf("unexpected id %q", q.ID)
	}
	if q.Title != "Sock puppet show" || q.Description != DefaultDescription {
		t.Fatalf("title/description not normalised: %+v", q)
	}
	if q.Source != quest.SourceCustom || q.XP != Reward(3, 2) || !q.Repeatable {
		t.Fatalf("unexpected quest: %+v", q)
	}

	second, _ := l.Save(Draft{Title: "Hum a tune", Description: "Any tune.", Category: quest.Play, Minutes: 1, Difficulty: 1})
	if got := l.List(); len(got) != 2 || got[0].ID != second.ID {
		t.Fatalf("expected newest first, got %+v", got)
	}

	// Edit keeps the id.
	d := draft()
	d.ID, d.Title = q.ID, "Sock puppet opera"
	edited, err := l.Save(d)
	if err != nil || edited.ID != q.ID || edited.Title != "Sock puppet opera" {
		t.Fatalf("edit failed: %+v %v", edited, err)
	}

	// Reopen from disk.
	l2 := open(t, path)
	if got := l2.List(); len(got) != 2 || got[1].Title != "Sock puppet opera" {
		t.Fatalf("not persisted: %+v", got)
	}

	if err := l2.Delete(q.ID); err != nil {
		t.Fatal(err)
	}
	if err := l2.Delete(q.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if got := open(t, path).List(); len(got) != 1 {
		t.Fatalf("delete not persisted: %+v", got)
	}
}

func TestSaveRejectsBadDrafts(t *testing.T) {
	l := open(t, filepath.Join(t.TempDir(), "c.json"))
	cases := map[string]struct {
		mut  func(*Draft)
		want error
	}{
		"empty title":   {func(d *Draft) { d.Title = "   " }, ErrNoTitle},
		"long title":    {func(d *Draft) { d.Title = strings.Repeat("a", 61) }, ErrTitleLong},
		"long details":  {func(d *Draft) { d.Description = strings.Repeat("a", 221) }, ErrDescLong},
		"bad category":  {func(d *Draft) { d.Category = "hacking" }, ErrBadCategory},
		"zero minutes":  {func(d *Draft) { d.Minutes = 0 }, ErrBadMinutes},
		"too difficult": {func(d *Draft) { d.Difficulty = 9 }, ErrBadDifficult},
		"unknown id":    {func(d *Draft) { d.ID = "my-nope" }, ErrNotFound},
	}
	for name, c := range cases {
		d := draft()
		c.mut(&d)
		if _, err := l.Save(d); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v want %v", name, err, c.want)
		}
	}
}

func TestLimit(t *testing.T) {
	l := open(t, filepath.Join(t.TempDir(), "c.json"))
	for i := 0; i < MaxQuests; i++ {
		if _, err := l.Save(draft()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Save(draft()); !errors.Is(err, ErrTooMany) {
		t.Fatalf("expected ErrTooMany, got %v", err)
	}
}

func TestOpenDropsTamperedQuests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.json")
	body := `{"format":1,"id":"custom","name":"My Quests","quests":[
		{"id":"my-ok","title":"Fine","description":"Ok","category":"chill","minutes":1,"difficulty":1,"xp":10,"repeatable":true},
		{"id":"my-greedy","title":"Free XP","description":"Ha","category":"chill","minutes":1,"difficulty":1,"xp":5000}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := Open(store.File{Path: path})
	if err == nil {
		t.Fatal("expected a problem to be reported")
	}
	got := l.List()
	if len(got) != 1 || got[0].ID != "my-ok" || got[0].Source != quest.SourceCustom {
		t.Fatalf("unexpected quests: %+v", got)
	}
}

func TestReward(t *testing.T) {
	if Reward(1, 1) != 10 || Reward(15, 3) != 40 || Reward(5, 2) != 24 {
		t.Fatal("unexpected rewards")
	}
}
