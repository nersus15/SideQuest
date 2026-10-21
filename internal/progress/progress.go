// Package progress turns completed quests into XP, levels and badges.
//
// The curve is deliberately gentle and there are no streaks or decay:
// progress only ever goes up, and taking a break never costs anything.
package progress

import "SideQuest/internal/quest"

// Level describes the player's position on the XP curve.
type Level struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	XPInLevel int    `json:"xpInLevel"` // XP earned since this level started
	XPToNext  int    `json:"xpToNext"`  // XP span of this level
}

var titles = []string{
	"Curious Blob",
	"Wanderer",
	"Side Kick",
	"Trailblazer",
	"Quest Goblin",
	"Daydream Ranger",
	"Chaos Wizard",
	"Legend of the Desk",
}

// levelSpan is the XP needed to go from level n to n+1.
func levelSpan(n int) int {
	return 50 + 25*(n-1)
}

// LevelFor computes the level for a total XP amount.
func LevelFor(xp int) Level {
	if xp < 0 {
		xp = 0
	}
	n := 1
	for xp >= levelSpan(n) {
		xp -= levelSpan(n)
		n++
	}
	title := titles[len(titles)-1]
	if n-1 < len(titles) {
		title = titles[n-1]
	}
	return Level{Number: n, Title: title, XPInLevel: xp, XPToNext: levelSpan(n)}
}

// Stats is the history summary badges are computed from.
type Stats struct {
	Completed       int
	ByCategory      map[quest.Category]int
	TwistsCompleted int
	SpicyCompleted  int
}

// Badge is a small, unpressured achievement.
type Badge struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Hint  string `json:"hint"`
}

type badgeRule struct {
	Badge
	earned func(Stats) bool
}

func category(c quest.Category, n int) func(Stats) bool {
	return func(s Stats) bool { return s.ByCategory[c] >= n }
}

var rules = []badgeRule{
	{Badge{"first-step", "First Step", "🐣", "Finish your first quest"}, func(s Stats) bool { return s.Completed >= 1 }},
	{Badge{"doodler", "Doodle Wizard", "🎨", "3 creative quests"}, category(quest.Creative, 3)},
	{Badge{"zen-bean", "Zen Bean", "🌿", "3 chill quests"}, category(quest.Chill, 3)},
	{Badge{"butterfly", "Social Butterfly", "🦋", "3 social quests"}, category(quest.Social, 3)},
	{Badge{"pathfinder", "Pathfinder", "🔭", "3 explore quests"}, category(quest.Explore, 3)},
	{Badge{"player-one", "Player One", "🎮", "3 play quests"}, category(quest.Play, 3)},
	{Badge{"gremlin", "Chaos Gremlin", "🦄", "3 wildcard quests"}, category(quest.Wildcard, 3)},
	{Badge{"twisty", "Twist Enjoyer", "🌀", "5 quests with a twist"}, func(s Stats) bool { return s.TwistsCompleted >= 5 }},
	{Badge{"spicy", "Brave Heart", "🦁", "A bold quest"}, func(s Stats) bool { return s.SpicyCompleted >= 1 }},
	{Badge{"all-rounder", "All-Rounder", "🌈", "One of every kind"}, func(s Stats) bool {
		for _, c := range quest.Categories {
			if s.ByCategory[c] == 0 {
				return false
			}
		}
		return true
	}},
	{Badge{"main-character", "Main Character", "⭐", "25 quests total"}, func(s Stats) bool { return s.Completed >= 25 }},
}

// AllBadges lists every badge in display order.
func AllBadges() []Badge {
	out := make([]Badge, len(rules))
	for i, r := range rules {
		out[i] = r.Badge
	}
	return out
}

// NewlyEarned returns badges earned with stats that are not in owned.
func NewlyEarned(s Stats, owned map[string]bool) []Badge {
	var out []Badge
	for _, r := range rules {
		if !owned[r.ID] && r.earned(s) {
			out = append(out, r.Badge)
		}
	}
	return out
}

// MaxDifficulty is the hardest quest difficulty unlocked at a level.
// New players start with easy quests only, so nothing feels like work.
func MaxDifficulty(level int) int {
	switch {
	case level >= 5:
		return 3
	case level >= 3:
		return 2
	}
	return 1
}

// DifficultyLevel is the level at which difficulty d unlocks.
func DifficultyLevel(d int) int {
	switch d {
	case 2:
		return 3
	case 3:
		return 5
	}
	return 1
}

// Look is something Boo can wear, unlocked by levelling up.
type Look struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"` // "hat" or "extra"
	Level int    `json:"level"`
}

var looks = []Look{
	{"hat-violet", "hat", 1},
	{"hat-mint", "hat", 2},
	{"glasses", "extra", 3},
	{"hat-coral", "hat", 4},
	{"scarf", "extra", 5},
	{"hat-gold", "hat", 6},
	{"flower", "extra", 7},
	{"crown", "hat", 8},
}

// DefaultHat is what Boo wears from the start.
const DefaultHat = "hat-violet"

// Looks lists everything Boo can wear, in unlock order.
func Looks() []Look { return append([]Look(nil), looks...) }

// LookByID finds a look.
func LookByID(id string) (Look, bool) {
	for _, l := range looks {
		if l.ID == id {
			return l, true
		}
	}
	return Look{}, false
}

// LooksUnlocked returns the looks unlocked after level from, up to and
// including level to.
func LooksUnlocked(from, to int) []Look {
	var out []Look
	for _, l := range looks {
		if l.Level > from && l.Level <= to {
			out = append(out, l)
		}
	}
	return out
}
