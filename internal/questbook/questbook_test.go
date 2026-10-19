package questbook

import (
	"testing"
	"time"
)

func TestNextQuestAvoidsImmediateRepeats(t *testing.T) {
	quests := []Quest{
		{ID: "q1", Title: "Look up", Category: "Explore", Duration: 2, Difficulty: 1, XP: 15},
		{ID: "q2", Title: "Sing a line", Category: "Creative", Duration: 1, Difficulty: 1, XP: 18},
		{ID: "q3", Title: "Stretch and reset", Category: "Chill", Duration: 2, Difficulty: 1, XP: 14},
	}

	mgr := NewManager(quests, func() time.Time { return time.Unix(1700000000, 0).UTC() })
	mgr.RecordHistory("q1", false)
	mgr.RecordHistory("q2", false)

	quest, err := mgr.NextQuest()
	if err != nil {
		t.Fatalf("expected quest, got error: %v", err)
	}

	if quest.ID == "q1" || quest.ID == "q2" {
		t.Fatalf("selected recent quest %q despite cooldown rules", quest.ID)
	}
}

func TestCompleteQuestRewardsXPAndTracksBadges(t *testing.T) {
	quests := []Quest{
		{ID: "q1", Title: "Quick sketch", Category: "Creative", Duration: 3, Difficulty: 1, XP: 30},
	}

	mgr := NewManager(quests, func() time.Time { return time.Unix(1700000000, 0).UTC() })
	if err := mgr.SetCurrent("q1"); err != nil {
		t.Fatalf("set current quest: %v", err)
	}

	reward, badges, err := mgr.CompleteCurrent()
	if err != nil {
		t.Fatalf("complete current quest: %v", err)
	}
	if reward != 30 {
		t.Fatalf("expected 30 xp, got %d", reward)
	}
	if len(badges) == 0 {
		t.Fatal("expected badge unlocks for an accepted quest")
	}
	if !mgr.HasCompleted("q1") {
		t.Fatal("expected quest to be marked complete")
	}
}
