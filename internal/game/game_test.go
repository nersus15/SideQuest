package game

import (
	"errors"
	"math/rand/v2"
	"path/filepath"
	"testing"
	"time"

	"SideQuest/internal/quest"
	"SideQuest/internal/store"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newGame(t *testing.T, path string, c *clock) *Game {
	t.Helper()
	qs, err := quest.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	g, err := New(qs, store.File{Path: path}, Options{Now: c.now, Rand: rand.New(rand.NewPCG(7, 7))})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestFullLoopPersistsAcrossRestarts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, path, c)

	v := g.State()
	if v.Quest == nil || v.Accepted {
		t.Fatalf("expected a fresh offer, got %+v", v)
	}
	first := v.Quest.ID

	if _, err := g.Complete(); !errors.Is(err, ErrNoActiveQuest) {
		t.Fatalf("completing an unaccepted quest should fail, got %v", err)
	}

	v = g.Accept()
	if !v.Accepted || v.Quest.ID != first {
		t.Fatalf("accept changed the quest: %+v", v)
	}
	if v.Twist != nil {
		v = g.SetTwist(true)
		if !v.TwistOn || v.Reward != v.Quest.XP+v.Twist.Bonus {
			t.Fatalf("twist reward wrong: %+v", v)
		}
	}
	reward := v.Reward

	res, err := g.Complete()
	if err != nil {
		t.Fatal(err)
	}
	if res.Gained != reward || res.View.XP != reward || res.View.Completed != 1 {
		t.Fatalf("unexpected completion: %+v", res)
	}
	if len(res.NewBadges) == 0 || res.NewBadges[0].ID != "first-step" {
		t.Fatalf("expected first-step badge, got %+v", res.NewBadges)
	}
	if res.View.Quest == nil || res.View.Quest.ID == first || res.View.Accepted {
		t.Fatal("a new, different quest should be offered after completion")
	}

	// Restart: everything should come back.
	g2 := newGame(t, path, c)
	v2 := g2.State()
	if v2.XP != reward || v2.Completed != 1 || v2.Quest.ID != res.View.Quest.ID {
		t.Fatalf("state lost across restart: %+v", v2)
	}
	earned := 0
	for _, b := range v2.Badges {
		if b.Earned {
			earned++
		}
	}
	if earned != 1 {
		t.Fatalf("expected 1 earned badge after restart, got %d", earned)
	}
}

func TestRerollAndSkipAvoidRepeats(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)

	seen := map[string]int{}
	prev := g.State().Quest.ID
	seen[prev]++
	for i := 0; i < 20; i++ {
		v := g.Reroll()
		if v.Quest.ID == prev {
			t.Fatalf("reroll returned the same quest %s", prev)
		}
		prev = v.Quest.ID
		seen[prev]++
	}
	for id, n := range seen {
		if n > 2 {
			t.Fatalf("quest %s shown %d times in 21 rolls", id, n)
		}
	}

	if v := g.Skip(); v.Quest.ID == prev {
		t.Fatal("skip did not change the quest")
	}
}

func TestSkippedQuestRests(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	skipped := g.State().Quest.ID
	g.Skip()
	for i := 0; i < 60; i++ {
		if g.Reroll().Quest.ID == skipped {
			t.Fatalf("skipped quest %s came back within the rest period", skipped)
		}
	}
}

func TestReleaseKeepsQuestWithoutPenalty(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	id := g.Accept().Quest.ID
	v := g.Release()
	if v.Accepted || v.Quest.ID != id || v.XP != 0 {
		t.Fatalf("release should un-accept only: %+v", v)
	}
}

func TestPrefsReplaceUnfittingOffer(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	cat := g.State().Quest.Category
	v, err := g.SetPrefs(quest.Prefs{Muted: []quest.Category{cat}})
	if err != nil {
		t.Fatal(err)
	}
	if v.Quest.Category == cat {
		t.Fatalf("offer in muted category %s was kept", cat)
	}
	if _, err := g.SetPrefs(quest.Prefs{Muted: []quest.Category{"nope"}}); !errors.Is(err, ErrBadPrefs) {
		t.Fatalf("expected ErrBadPrefs, got %v", err)
	}
}

func TestLevelUpReported(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	leveled := false
	for i := 0; i < 10 && !leveled; i++ {
		g.Accept()
		res, err := g.Complete()
		if err != nil {
			t.Fatal(err)
		}
		leveled = res.LeveledUp
		c.t = c.t.Add(time.Hour)
	}
	if !leveled {
		t.Fatal("expected a level up within 10 quests")
	}
}

func TestUnknownSavedQuestIsDropped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	if err := (store.File{Path: path}).Save(Save{Current: &Current{QuestID: "gone-forever"}}); err != nil {
		t.Fatal(err)
	}
	c := &clock{t: time.Now()}
	g := newGame(t, path, c)
	if v := g.State(); v.Quest == nil || v.Quest.ID == "gone-forever" {
		t.Fatalf("stale quest not replaced: %+v", v.Quest)
	}
}
