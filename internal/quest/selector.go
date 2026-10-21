package quest

import (
	"math/rand/v2"
	"time"
)

// Prefs are the player's (optional) quest preferences.
type Prefs struct {
	// Muted categories are never offered unless nothing else is left.
	Muted []Category `json:"muted"`
	// Avoid lists needs the player cannot meet right now (e.g. "people").
	Avoid []Need `json:"avoid"`
	// MaxDifficulty hides harder quests until the player has levelled up.
	// It is set by the game, not the player; 0 means no limit.
	MaxDifficulty int `json:"-"`
}

// Unlocked reports whether q's difficulty is unlocked. The player's own
// quests are always allowed: they chose how hard to make them.
func (p Prefs) Unlocked(q Quest) bool {
	return p.MaxDifficulty == 0 || q.Difficulty <= p.MaxDifficulty || q.Source == SourceCustom
}

// Allows reports whether q fits the preferences.
func (p Prefs) Allows(q Quest) bool {
	for _, c := range p.Muted {
		if q.Category == c {
			return false
		}
	}
	for _, avoid := range p.Avoid {
		for _, n := range q.Needs {
			if n == avoid {
				return false
			}
		}
	}
	return true
}

// Memory is what the selector needs to know about the player's history.
type Memory struct {
	LastCompleted map[string]time.Time
	LastSkipped   map[string]time.Time
	// RecentlyShown holds recently offered quest ids, oldest first.
	RecentlyShown []string
	LastCategory  Category
	// Ratings are the player's thumbs up or down, by quest id.
	Ratings map[string]Rating
}

// Rating is a player's opinion of a quest.
type Rating string

const (
	RatingUp   Rating = "up"
	RatingDown Rating = "down"
)

// Valid reports whether r is a known rating ("" clears a rating).
func (r Rating) Valid() bool { return r == "" || r == RatingUp || r == RatingDown }

// Weights applied from ratings. A disliked quest is rare, never gone:
// moods change, and the tiers above still decide what is allowed.
const (
	likedWeight    = 1.5
	dislikedWeight = 0.2
)

// categoryAffinity turns ratings into a gentle per-category factor between
// 0.8 and 1.3. The +2 in the denominator keeps a single rating from
// swinging a whole category.
func categoryAffinity(quests []Quest, ratings map[string]Rating) map[Category]float64 {
	if len(ratings) == 0 {
		return nil
	}
	cat := make(map[string]Category, len(quests))
	for _, q := range quests {
		cat[q.ID] = q.Category
	}
	up := map[Category]int{}
	down := map[Category]int{}
	for id, r := range ratings {
		c, ok := cat[id]
		if !ok {
			continue
		}
		switch r {
		case RatingUp:
			up[c]++
		case RatingDown:
			down[c]++
		}
	}
	out := map[Category]float64{}
	for _, c := range Categories {
		score := float64(up[c]-down[c]) / float64(up[c]+down[c]+2)
		out[c] = min(1.3, max(0.8, 1+0.4*score))
	}
	return out
}

// SkipRest is how long a skipped quest stays out of rotation.
const SkipRest = 12 * time.Hour

// Offer is a quest picked for the player plus a suggested twist.
type Offer struct {
	Quest Quest  `json:"quest"`
	Twist *Twist `json:"twist,omitempty"`
}

// Pick chooses a quest. It never returns `exclude` unless it is the only
// quest left, and it progressively relaxes its rules so the player always
// gets something when content exists.
func Pick(quests []Quest, prefs Prefs, mem Memory, now time.Time, exclude string, rng *rand.Rand) (Offer, error) {
	if len(quests) == 0 {
		return Offer{}, ErrNoQuests
	}

	recentWindow := min(8, len(quests)/2)
	recent := map[string]bool{}
	for i := len(mem.RecentlyShown) - 1; i >= 0 && len(recent) < recentWindow; i-- {
		recent[mem.RecentlyShown[i]] = true
	}

	resting := func(q Quest) bool {
		last, done := mem.LastCompleted[q.ID]
		if !done {
			return false
		}
		if !q.Repeatable {
			return true
		}
		return within(now, last, q.Cooldown())
	}
	skipped := func(q Quest) bool {
		at, ok := mem.LastSkipped[q.ID]
		return ok && within(now, at, SkipRest)
	}

	// Each tier is a filter; the first tier with candidates wins.
	affinity := categoryAffinity(quests, mem.Ratings)
	open := prefs.Unlocked
	tiers := []func(Quest) bool{
		func(q Quest) bool { return open(q) && prefs.Allows(q) && !resting(q) && !skipped(q) && !recent[q.ID] },
		func(q Quest) bool { return open(q) && prefs.Allows(q) && !resting(q) && !skipped(q) },
		func(q Quest) bool { return open(q) && prefs.Allows(q) && !resting(q) },
		func(q Quest) bool { return open(q) && !resting(q) },
		func(q Quest) bool { return open(q) },
		func(q Quest) bool { return true },
	}

	for _, ok := range tiers {
		var pool []Quest
		var weights []float64
		total := 0.0
		for _, q := range quests {
			if q.ID == exclude || !ok(q) {
				continue
			}
			w := weight(q, mem, affinity)
			pool = append(pool, q)
			weights = append(weights, w)
			total += w
		}
		if len(pool) == 0 {
			continue
		}
		r := rng.Float64() * total
		chosen := pool[len(pool)-1]
		for i, w := range weights {
			if r < w {
				chosen = pool[i]
				break
			}
			r -= w
		}
		return makeOffer(chosen, rng), nil
	}

	// Only the excluded quest exists.
	for _, q := range quests {
		if q.ID == exclude {
			return makeOffer(q, rng), nil
		}
	}
	return Offer{}, ErrNoQuests
}

// within reports whether t happened less than d before now. A timestamp in
// the future (the clock was moved back) counts as long ago, so a clock change
// can never lock quests away.
func within(now, t time.Time, d time.Duration) bool {
	since := now.Sub(t)
	return since >= 0 && since < d
}

func weight(q Quest, mem Memory, affinity map[Category]float64) float64 {
	w := 1.0
	switch mem.Ratings[q.ID] {
	case RatingUp:
		w *= likedWeight
	case RatingDown:
		w *= dislikedWeight
	}
	if f, ok := affinity[q.Category]; ok {
		w *= f
	}
	if mem.LastCategory != "" && q.Category != mem.LastCategory {
		w *= 1.6 // nudge toward variety
	}
	if _, done := mem.LastCompleted[q.ID]; !done {
		w *= 1.3 // a little extra love for unseen quests
	}
	if q.Source == SourceCustom {
		w *= 2 // players want to see the quests they wrote
	}
	for _, id := range mem.RecentlyShown {
		if id == q.ID {
			w *= 0.5
			break
		}
	}
	return w
}

func makeOffer(q Quest, rng *rand.Rand) Offer {
	o := Offer{Quest: q}
	if len(q.Twists) > 0 {
		t := q.Twists[rng.IntN(len(q.Twists))]
		o.Twist = &t
	}
	return o
}
