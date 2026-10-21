package game

import (
	"errors"
	"math/rand/v2"
	"path/filepath"
	"testing"
	"time"

	"SideQuest/internal/quest"
	"SideQuest/internal/schedule"
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

func TestSetQuestsAndPlay(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	builtin, _ := quest.Builtin()

	mine := quest.Quest{ID: "my-hum", Title: "Hum", Description: "Hum it", Category: quest.Play,
		Minutes: 1, Difficulty: 1, XP: 10, Repeatable: true, Source: quest.SourceCustom}
	if err := g.SetQuests(append(builtin, mine)); err != nil {
		t.Fatal(err)
	}
	v, err := g.Play("my-hum")
	if err != nil || v.Quest.ID != "my-hum" || v.Accepted {
		t.Fatalf("play failed: %+v %v", v.Quest, err)
	}
	if _, err := g.Play("nope"); !errors.Is(err, ErrUnknownQuest) {
		t.Fatalf("expected ErrUnknownQuest, got %v", err)
	}

	// Removing the offered quest replaces the offer.
	if err := g.SetQuests(builtin); err != nil {
		t.Fatal(err)
	}
	if v := g.State(); v.Quest == nil || v.Quest.ID == "my-hum" {
		t.Fatalf("stale offer kept: %+v", v.Quest)
	}

	// An accepted quest is never swapped out by Play.
	g.SetQuests(append(builtin, mine))
	accepted := g.Accept().Quest.ID
	if _, err := g.Play("my-hum"); !errors.Is(err, ErrQuestActive) {
		t.Fatalf("expected ErrQuestActive, got %v", err)
	}
	if g.State().Quest.ID != accepted {
		t.Fatal("accepted quest was replaced")
	}
}

func TestNewPlayersOnlyGetEasyQuests(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	if v := g.State(); v.MaxDifficulty != 1 || v.NextDifficultyAt != 3 {
		t.Fatalf("unexpected unlock state: %d next %d", v.MaxDifficulty, v.NextDifficultyAt)
	}
	for i := 0; i < 40; i++ {
		if d := g.Reroll().Quest.Difficulty; d != 1 {
			t.Fatalf("level 1 got a difficulty %d quest", d)
		}
	}
}

func TestLevelUpUnlocksLooksAndDifficulty(t *testing.T) {
	c := &clock{t: time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, filepath.Join(t.TempDir(), "s.json"), c)
	var looks []string
	newDifficulty := 0
	for i := 0; i < 30 && newDifficulty == 0; i++ {
		g.Accept()
		res, err := g.Complete()
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range res.NewLooks {
			looks = append(looks, l.ID)
		}
		newDifficulty = res.NewDifficulty
		c.t = c.t.Add(time.Hour)
	}
	if newDifficulty != 2 {
		t.Fatalf("expected medium quests to unlock, got %d", newDifficulty)
	}
	if len(looks) < 2 || looks[0] != "hat-mint" || looks[1] != "glasses" {
		t.Fatalf("expected mint hat then glasses, got %v", looks)
	}

	v, err := g.SetLook(Look{Hat: "hat-mint", Extra: "glasses"})
	if err != nil || v.Look.Hat != "hat-mint" {
		t.Fatalf("could not wear unlocked look: %v", err)
	}
	if _, err := g.SetLook(Look{Hat: "crown"}); !errors.Is(err, ErrLookLocked) {
		t.Fatalf("crown should be locked, got %v", err)
	}
	if _, err := g.SetLook(Look{Hat: "glasses"}); !errors.Is(err, ErrLookLocked) {
		t.Fatalf("an extra is not a hat, got %v", err)
	}
}

func TestLanguageAndOnboardingPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	c := &clock{t: time.Now()}
	g := newGame(t, path, c)
	if v := g.State(); v.Onboarded || v.Lang != "" || v.Look.Hat != "hat-violet" {
		t.Fatalf("unexpected fresh state: %+v", v)
	}
	if _, err := g.SetLanguage("fr"); !errors.Is(err, ErrBadLanguage) {
		t.Fatalf("expected ErrBadLanguage, got %v", err)
	}
	if _, err := g.SetLanguage("id"); err != nil {
		t.Fatal(err)
	}
	g.FinishOnboarding()

	v := newGame(t, path, c).State()
	if v.Lang != "id" || !v.Onboarded {
		t.Fatalf("settings lost across restart: lang %q onboarded %v", v.Lang, v.Onboarded)
	}
}

func TestRatingsPersistAndShowOnCompletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	c := &clock{t: time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)}
	g := newGame(t, path, c)
	id := g.Accept().Quest.ID

	if err := g.RateQuest(id, "meh"); !errors.Is(err, ErrBadRating) {
		t.Fatalf("expected ErrBadRating, got %v", err)
	}
	if err := g.RateQuest("not-a-quest", quest.RatingUp); !errors.Is(err, ErrUnknownQuest) {
		t.Fatalf("expected ErrUnknownQuest, got %v", err)
	}
	if err := g.RateQuest(id, quest.RatingUp); err != nil {
		t.Fatal(err)
	}
	res, err := g.Complete()
	if err != nil {
		t.Fatal(err)
	}
	if res.Rating != quest.RatingUp {
		t.Fatalf("completion should carry the existing rating, got %q", res.Rating)
	}

	g2 := newGame(t, path, c)
	if g2.save.Ratings[id] != quest.RatingUp {
		t.Fatal("rating lost across restart")
	}
	if err := g2.RateQuest(id, ""); err != nil || len(g2.save.Ratings) != 0 {
		t.Fatalf("clearing a rating failed: %v %v", err, g2.save.Ratings)
	}
}

func TestNudgeSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	c := &clock{t: time.Now()}
	g := newGame(t, path, c)
	if n := g.State().Nudges; n.BreakMinutes != 0 || n.StayWhenBusy {
		t.Fatalf("defaults should be off and hide-when-busy, got %+v", n)
	}
	if _, err := g.SetNudges(Nudges{BreakMinutes: 5}); !errors.Is(err, ErrBadNudges) {
		t.Fatalf("expected ErrBadNudges, got %v", err)
	}
	if _, err := g.SetNudges(Nudges{BreakMinutes: 90, StayWhenBusy: true}); err != nil {
		t.Fatal(err)
	}
	if n := newGame(t, path, c).NudgeSettings(); n.BreakMinutes != 90 || !n.StayWhenBusy {
		t.Fatalf("nudge settings lost across restart: %+v", n)
	}
}

func TestScheduleSettingsAndSlots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	c := &clock{t: time.Now()}
	g := newGame(t, path, c)
	if s := g.State().Schedule; s.Days != "daily" || len(s.Times) != 0 {
		t.Fatalf("unexpected default schedule: %+v", s)
	}
	if _, err := g.SetSchedule(schedule.Schedule{Times: []string{"99:00"}}); err == nil {
		t.Fatal("expected an invalid schedule error")
	}
	v, err := g.SetSchedule(schedule.Schedule{Times: []string{"15:00", "10:00"}, Days: "weekdays", Notify: true})
	if err != nil || v.Schedule.Times[0] != "10:00" {
		t.Fatalf("got %+v %v", v.Schedule, err)
	}
	if !g.MarkScheduled("2026-10-21 10:00") || g.MarkScheduled("2026-10-21 10:00") {
		t.Fatal("a slot must fire exactly once")
	}
	g2 := newGame(t, path, c)
	if g2.MarkScheduled("2026-10-21 10:00") {
		t.Fatal("a fired slot must not fire again after a restart")
	}
	if s := g2.ScheduleSettings(); !s.Notify || s.Days != "weekdays" {
		t.Fatalf("schedule lost across restart: %+v", s)
	}
}
