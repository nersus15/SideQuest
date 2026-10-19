package progress

import (
	"testing"

	"SideQuest/internal/quest"
)

func TestLevelFor(t *testing.T) {
	cases := []struct {
		xp, level, in, span int
	}{
		{0, 1, 0, 50},
		{49, 1, 49, 50},
		{50, 2, 0, 75},
		{124, 2, 74, 75},
		{125, 3, 0, 100},
		{-5, 1, 0, 50},
	}
	for _, c := range cases {
		l := LevelFor(c.xp)
		if l.Number != c.level || l.XPInLevel != c.in || l.XPToNext != c.span {
			t.Errorf("LevelFor(%d) = %+v, want level %d in %d span %d", c.xp, l, c.level, c.in, c.span)
		}
		if l.Title == "" {
			t.Errorf("LevelFor(%d) has no title", c.xp)
		}
	}
	if LevelFor(1_000_000).Title == "" {
		t.Error("high levels need a title")
	}
}

func TestNewlyEarned(t *testing.T) {
	s := Stats{Completed: 1, ByCategory: map[quest.Category]int{quest.Creative: 1}}
	got := NewlyEarned(s, map[string]bool{})
	if len(got) != 1 || got[0].ID != "first-step" {
		t.Fatalf("expected first-step, got %+v", got)
	}
	if again := NewlyEarned(s, map[string]bool{"first-step": true}); len(again) != 0 {
		t.Fatalf("badge awarded twice: %+v", again)
	}

	all := Stats{Completed: 6, ByCategory: map[quest.Category]int{}}
	for _, c := range quest.Categories {
		all.ByCategory[c] = 1
	}
	found := false
	for _, b := range NewlyEarned(all, map[string]bool{}) {
		found = found || b.ID == "all-rounder"
	}
	if !found {
		t.Fatal("expected all-rounder badge")
	}
}

func TestBadgeIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range AllBadges() {
		if seen[b.ID] {
			t.Fatalf("duplicate badge id %s", b.ID)
		}
		seen[b.ID] = true
	}
}
