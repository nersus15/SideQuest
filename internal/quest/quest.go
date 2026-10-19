// Package quest defines quest content, quest packs, validation and selection.
//
// Quest packs are plain JSON data. Nothing inside a pack is ever executed:
// every field is validated against a strict whitelist before it is used.
package quest

import "time"

// Category groups quests by the kind of fun they offer.
type Category string

const (
	Creative Category = "creative"
	Chill    Category = "chill"
	Social   Category = "social"
	Explore  Category = "explore"
	Play     Category = "play"
	Wildcard Category = "wildcard"
)

// Categories lists every known category in display order.
var Categories = []Category{Creative, Chill, Social, Explore, Play, Wildcard}

// Valid reports whether c is a known category.
func (c Category) Valid() bool {
	for _, known := range Categories {
		if c == known {
			return true
		}
	}
	return false
}

// Need describes something a quest requires from the player's surroundings.
type Need string

const (
	NeedPaper   Need = "paper"   // something to draw or write on
	NeedPeople  Need = "people"  // another person nearby or reachable
	NeedMove    Need = "move"    // room to stand up and move around
	NeedOutside Need = "outside" // a window, door or outdoor space
	NeedPhone   Need = "phone"   // a phone or camera
	NeedSound   Need = "sound"   // ok to make a little noise
)

var knownNeeds = map[Need]bool{
	NeedPaper: true, NeedPeople: true, NeedMove: true,
	NeedOutside: true, NeedPhone: true, NeedSound: true,
}

// Source tells where a quest came from.
type Source string

const (
	SourceBuiltin  Source = "builtin"
	SourceCustom   Source = "custom"
	SourceImported Source = "imported"
)

// Twist is a meaningful variation of a quest that changes how it is played.
type Twist struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Bonus int    `json:"bonus,omitempty"` // extra XP for taking the twist
}

// Quest is a single playable challenge.
type Quest struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    Category `json:"category"`
	Minutes     int      `json:"minutes"`
	Difficulty  int      `json:"difficulty"` // 1 easy, 2 medium, 3 spicy
	Needs       []Need   `json:"needs,omitempty"`
	XP          int      `json:"xp"`
	// Repeatable quests come back after their cooldown; one-offs never repeat
	// once completed.
	Repeatable    bool    `json:"repeatable"`
	CooldownHours int     `json:"cooldownHours,omitempty"`
	Twists        []Twist `json:"twists,omitempty"`
	Source        Source  `json:"source"`
	Pack          string  `json:"pack"`
}

// Cooldown returns how long a completed quest rests before it can return.
func (q Quest) Cooldown() time.Duration {
	if q.CooldownHours > 0 {
		return time.Duration(q.CooldownHours) * time.Hour
	}
	return DefaultCooldown
}

// TwistByID finds a twist by id.
func (q Quest) TwistByID(id string) (Twist, bool) {
	for _, t := range q.Twists {
		if t.ID == id {
			return t, true
		}
	}
	return Twist{}, false
}

// DefaultCooldown applies to repeatable quests that do not set their own.
const DefaultCooldown = 24 * time.Hour

// KnownNeed reports whether n is a recognised need.
func KnownNeed(n Need) bool { return knownNeeds[n] }
