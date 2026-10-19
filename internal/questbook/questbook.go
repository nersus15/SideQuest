package questbook

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

const defaultCooldown = 2 * time.Hour

// Quest describes a single playable challenge in the app.
type Quest struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Duration      int       `json:"durationMinutes"`
	Difficulty    int       `json:"difficulty"`
	Context       string    `json:"context,omitempty"`
	XP            int       `json:"xp"`
	Repeatable    bool      `json:"repeatable"`
	Cooldown      int       `json:"cooldownMinutes,omitempty"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
	LastCompleted time.Time `json:"lastCompleted,omitempty"`
}

// PlayerProgress tracks local progression state.
type PlayerProgress struct {
	XP          int                  `json:"xp"`
	Level       int                  `json:"level"`
	Completed   map[string]time.Time `json:"completed"`
	Badges      map[string]bool      `json:"badges"`
	LastUpdated time.Time            `json:"lastUpdated"`
}

// HistoryEntry stores lightweight quest usage history.
type HistoryEntry struct {
	QuestID   string    `json:"questId"`
	Completed bool      `json:"completed"`
	Timestamp time.Time `json:"timestamp"`
}

// Manager handles quest selection, state transitions, and progression.
type Manager struct {
	quests    []Quest
	history   []HistoryEntry
	completed map[string]time.Time
	current   string
	player    PlayerProgress
	now       func() time.Time
}

// NewManager instantiates a quest manager with built-in content.
func NewManager(quests []Quest, now func() time.Time) *Manager {
	if now == nil {
		now = time.Now
	}
	if len(quests) == 0 {
		quests = defaultQuests()
	}
	completed := make(map[string]time.Time)
	for _, q := range quests {
		if !q.LastCompleted.IsZero() {
			completed[q.ID] = q.LastCompleted
		}
	}
	manager := &Manager{
		quests:    append([]Quest(nil), quests...),
		completed: completed,
		player: PlayerProgress{
			Completed: make(map[string]time.Time),
			Badges:    make(map[string]bool),
		},
		now: now,
	}
	if len(manager.player.Completed) == 0 {
		manager.player.Completed = map[string]time.Time{}
	}
	return manager
}

func defaultQuests() []Quest {
	now := time.Now().UTC()
	return []Quest{
		{ID: "creative-doodle", Title: "Doodle your weirdest idea", Description: "Sketch a tiny bizarre object in three lines or less.", Category: "Creative", Duration: 2, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
		{ID: "creative-mashup", Title: "Make a tiny mashup", Description: "Combine two random objects and invent a silly use for them.", Category: "Creative", Duration: 3, Difficulty: 2, XP: 20, Source: "built-in", CreatedAt: now},
		{ID: "creative-silhouette", Title: "Silhouette test", Description: "Trace one object in your room with a single continuous line.", Category: "Creative", Duration: 3, Difficulty: 2, XP: 22, Source: "built-in", CreatedAt: now},
		{ID: "creative-badge", Title: "Tiny icon design", Description: "Design a little icon or symbol for a made-up game.", Category: "Creative", Duration: 4, Difficulty: 2, XP: 24, Source: "built-in", CreatedAt: now},
		{ID: "creative-spell", Title: "Weird spell naming", Description: "Invent a silly magical spell name for a harmless everyday task.", Category: "Creative", Duration: 2, Difficulty: 2, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "creative-funny-quote", Title: "Tiny caption", Description: "Write a one-line caption for the strangest object nearby.", Category: "Creative", Duration: 2, Difficulty: 1, XP: 17, Source: "built-in", CreatedAt: now},
		{ID: "creative-emoji-story", Title: "Emoji story", Description: "Tell a 3-frame emoji story about your desk or room.", Category: "Creative", Duration: 2, Difficulty: 1, XP: 19, Source: "built-in", CreatedAt: now},
		{ID: "creative-lunch-art", Title: "Snack portrait", Description: "Turn a snack or drink into a playful little portrait.", Category: "Creative", Duration: 3, Difficulty: 2, XP: 22, Source: "built-in", CreatedAt: now},
		{ID: "chill-breathe", Title: "Slow breath reset", Description: "Take five slow breaths and notice one sound around you.", Category: "Chill", Duration: 1, Difficulty: 1, XP: 12, Source: "built-in", CreatedAt: now},
		{ID: "chill-stretch", Title: "Tiny stretch break", Description: "Reach overhead and loosen your shoulders for ten seconds.", Category: "Chill", Duration: 1, Difficulty: 1, XP: 12, Source: "built-in", CreatedAt: now},
		{ID: "chill-sensor", Title: "Body check-in", Description: "Notice one tense spot in your body and soften it gently.", Category: "Chill", Duration: 2, Difficulty: 1, XP: 14, Source: "built-in", CreatedAt: now},
		{ID: "chill-ambient", Title: "Ambient reset", Description: "Sit for thirty seconds and watch a single object around you.", Category: "Chill", Duration: 1, Difficulty: 1, XP: 13, Source: "built-in", CreatedAt: now},
		{ID: "chill-heating", Title: "Warm up your hands", Description: "Rub your palms together until they feel awake again.", Category: "Chill", Duration: 1, Difficulty: 1, XP: 11, Source: "built-in", CreatedAt: now},
		{ID: "chill-quiet", Title: "Quiet minute", Description: "Close your eyes for a minute and listen for the smallest sound.", Category: "Chill", Duration: 1, Difficulty: 2, XP: 16, Source: "built-in", CreatedAt: now},
		{ID: "chill-knee-tilt", Title: "Knee reset", Description: "Shift your knees and ankles into a more comfortable position.", Category: "Chill", Duration: 1, Difficulty: 1, XP: 10, Source: "built-in", CreatedAt: now},
		{ID: "chill-window", Title: "Window gaze", Description: "Look out a window and notice the light or sky for a moment.", Category: "Chill", Duration: 2, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
		{ID: "social-wave", Title: "Friendly wave", Description: "Say hi or wave to someone nearby in a cheerful way.", Category: "Social", Duration: 1, Difficulty: 1, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "social-compliment", Title: "Tiny compliment", Description: "Give someone a genuine compliment that makes them smile.", Category: "Social", Duration: 2, Difficulty: 2, XP: 20, Source: "built-in", CreatedAt: now},
		{ID: "social-story", Title: "Mini story swap", Description: "Ask someone for a tiny, funny story from their day.", Category: "Social", Duration: 2, Difficulty: 2, XP: 19, Source: "built-in", CreatedAt: now},
		{ID: "social-prop", Title: "Object intro", Description: "Show someone an interesting object in your room and explain it.", Category: "Social", Duration: 2, Difficulty: 2, XP: 21, Source: "built-in", CreatedAt: now},
		{ID: "social-name-game", Title: "Name game", Description: "Say the name of three things you like about the room.", Category: "Social", Duration: 1, Difficulty: 1, XP: 12, Source: "built-in", CreatedAt: now},
		{ID: "social-smile", Title: "Smile check", Description: "Smile at one person and notice if it changes the mood.", Category: "Social", Duration: 1, Difficulty: 1, XP: 14, Source: "built-in", CreatedAt: now},
		{ID: "social-ask", Title: "Curious question", Description: "Ask one person a low-stakes question you’ve always wondered.", Category: "Social", Duration: 2, Difficulty: 2, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "social-favor", Title: "Tiny favor", Description: "Do one small helpful thing for someone without making a big deal of it.", Category: "Social", Duration: 3, Difficulty: 2, XP: 24, Source: "built-in", CreatedAt: now},
		{ID: "explore-window", Title: "Window watch", Description: "Notice one interesting thing outside your window or door.", Category: "Explore", Duration: 2, Difficulty: 1, XP: 14, Source: "built-in", CreatedAt: now},
		{ID: "explore-sound", Title: "Listen for a secret sound", Description: "Pause and identify the most surprising sound you can hear.", Category: "Explore", Duration: 2, Difficulty: 1, XP: 16, Source: "built-in", CreatedAt: now},
		{ID: "explore-shadow", Title: "Shadow tracker", Description: "Find the strangest shadow in your room and watch it for ten seconds.", Category: "Explore", Duration: 2, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
		{ID: "explore-oddity", Title: "Oddity hunt", Description: "Find one weirdly shaped object and point it out to yourself.", Category: "Explore", Duration: 2, Difficulty: 1, XP: 17, Source: "built-in", CreatedAt: now},
		{ID: "explore-light", Title: "Light check", Description: "Notice where the light is strongest in your room right now.", Category: "Explore", Duration: 1, Difficulty: 1, XP: 13, Source: "built-in", CreatedAt: now},
		{ID: "explore-texture", Title: "Texture scout", Description: "Find the smoothest, roughest, or oddest texture within reach.", Category: "Explore", Duration: 2, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
		{ID: "explore-nearest", Title: "Nearest object", Description: "Pick the nearest object you’ve never really noticed before.", Category: "Explore", Duration: 1, Difficulty: 1, XP: 12, Source: "built-in", CreatedAt: now},
		{ID: "explore-color", Title: "Color hunt", Description: "Name every color you can spot in a single glance.", Category: "Explore", Duration: 2, Difficulty: 1, XP: 16, Source: "built-in", CreatedAt: now},
		{ID: "play-finger-tap", Title: "Beat challenge", Description: "Tap a rhythm on your desk or knee for 20 seconds.", Category: "Play", Duration: 1, Difficulty: 1, XP: 13, Source: "built-in", CreatedAt: now},
		{ID: "play-coin-flip", Title: "Make a tiny bet", Description: "Flip a coin or choose a random object to settle a tiny decision.", Category: "Play", Duration: 1, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
		{ID: "play-one-word", Title: "One-word challenge", Description: "Say a one-word answer to a silly question like “What is your chair?”", Category: "Play", Duration: 1, Difficulty: 1, XP: 12, Source: "built-in", CreatedAt: now},
		{ID: "play-gesture", Title: "Mystery gesture", Description: "Make one ridiculous gesture and then reset your face.", Category: "Play", Duration: 1, Difficulty: 2, XP: 17, Source: "built-in", CreatedAt: now},
		{ID: "play-hide-and-seek", Title: "Object hide", Description: "Hide a small object and then try to find it in under a minute.", Category: "Play", Duration: 2, Difficulty: 2, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "play-guess", Title: "Guess the object", Description: "Pick up a random object and ask yourself what it was “for” before using it.", Category: "Play", Duration: 2, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
		{ID: "play-pretend", Title: "Pretend mode", Description: "Pretend your keyboard has a personality and greet it for ten seconds.", Category: "Play", Duration: 1, Difficulty: 2, XP: 16, Source: "built-in", CreatedAt: now},
		{ID: "play-balance", Title: "Balance challenge", Description: "Stand on one foot for as long as you can while smiling.", Category: "Play", Duration: 1, Difficulty: 2, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "fun-absurd", Title: "Absurd little mission", Description: "Pretend your keyboard is a spaceship for thirty seconds.", Category: "Random Fun", Duration: 1, Difficulty: 2, XP: 17, Source: "built-in", CreatedAt: now},
		{ID: "fun-shape", Title: "Spot the shape", Description: "Find the most unusual shape in your room and point it out.", Category: "Random Fun", Duration: 2, Difficulty: 1, XP: 14, Source: "built-in", CreatedAt: now},
		{ID: "fun-random-fact", Title: "Tiny fact crown", Description: "Invent a completely fake fact about a nearby object and tell it proudly.", Category: "Random Fun", Duration: 2, Difficulty: 2, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "fun-silly-walk", Title: "Silly stroll", Description: "Walk to the next room like a dramatic movie character for ten steps.", Category: "Random Fun", Duration: 1, Difficulty: 2, XP: 18, Source: "built-in", CreatedAt: now},
		{ID: "fun-hat", Title: "Hat check", Description: "Put on a harmless object as a ridiculous hat and keep it for a minute.", Category: "Random Fun", Duration: 1, Difficulty: 2, XP: 17, Source: "built-in", CreatedAt: now},
		{ID: "fun-voice", Title: "Voice warp", Description: "Say one sentence in a dramatic voice that doesn’t match your mood.", Category: "Random Fun", Duration: 1, Difficulty: 1, XP: 14, Source: "built-in", CreatedAt: now},
		{ID: "fun-mirror", Title: "Mirror wink", Description: "Wink at your reflection or a shiny object for a second.", Category: "Random Fun", Duration: 1, Difficulty: 1, XP: 13, Source: "built-in", CreatedAt: now},
		{ID: "fun-rename", Title: "Rename time", Description: "Give your current setup a ridiculous new title like a tiny kingdom.", Category: "Random Fun", Duration: 2, Difficulty: 1, XP: 15, Source: "built-in", CreatedAt: now},
	}
}

// All returns all available quests.
func (m *Manager) All() []Quest {
	return append([]Quest(nil), m.quests...)
}

// NextQuest picks a quest while respecting recent history, cooldowns, and repeatability.
func (m *Manager) NextQuest() (Quest, error) {
	if len(m.quests) == 0 {
		return Quest{}, errors.New("no quests available")
	}

	now := m.now()
	eligible := make([]Quest, 0, len(m.quests))
	for _, q := range m.quests {
		if q.ID == m.current {
			continue
		}
		if !m.isAvailable(q, now) {
			continue
		}
		eligible = append(eligible, q)
	}
	if len(eligible) == 0 {
		for _, q := range m.quests {
			if m.isAvailable(q, now) {
				m.current = q.ID
				return q, nil
			}
		}
		return Quest{}, errors.New("no quest matched the current availability rules")
	}

	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].Category != eligible[j].Category {
			return eligible[i].Category < eligible[j].Category
		}
		if eligible[i].Difficulty != eligible[j].Difficulty {
			return eligible[i].Difficulty < eligible[j].Difficulty
		}
		return eligible[i].XP > eligible[j].XP
	})

	choice := eligible[0]
	m.current = choice.ID
	return choice, nil
}

func (m *Manager) isAvailable(q Quest, now time.Time) bool {
	for _, entry := range m.history {
		if entry.QuestID != q.ID {
			continue
		}
		if !entry.Completed && now.Sub(entry.Timestamp) < 3*time.Hour {
			return false
		}
		if entry.Completed {
			cooldown := q.Cooldown
			if cooldown <= 0 {
				cooldown = int(defaultCooldown.Minutes())
			}
			if now.Sub(entry.Timestamp) < time.Duration(cooldown)*time.Minute {
				return false
			}
		}
	}
	if !q.Repeatable {
		if last, ok := m.completed[q.ID]; ok {
			if now.Sub(last) < defaultCooldown {
				return false
			}
		}
	}
	return true
}

// SetCurrent sets the active quest by ID.
func (m *Manager) SetCurrent(id string) error {
	for _, q := range m.quests {
		if q.ID == id {
			m.current = id
			return nil
		}
	}
	return fmt.Errorf("quest %q not found", id)
}

// Current returns the active quest if set.
func (m *Manager) Current() (Quest, bool) {
	for _, q := range m.quests {
		if q.ID == m.current {
			return q, true
		}
	}
	return Quest{}, false
}

// CurrentQuestID returns the currently selected quest ID.
func (m *Manager) CurrentQuestID() string {
	return m.current
}

// SetProgress restores saved data from disk.
func (m *Manager) SetProgress(progress PlayerProgress) {
	if progress.Completed == nil {
		progress.Completed = make(map[string]time.Time)
	}
	if progress.Badges == nil {
		progress.Badges = make(map[string]bool)
	}
	m.player = progress
	m.completed = make(map[string]time.Time)
	for id, ts := range progress.Completed {
		m.completed[id] = ts
	}
}

// RecordHistory appends recent quest activity.
func (m *Manager) RecordHistory(id string, completed bool) {
	m.history = append(m.history, HistoryEntry{QuestID: id, Completed: completed, Timestamp: m.now()})
}

// HasCompleted checks if a quest was completed before.
func (m *Manager) HasCompleted(id string) bool {
	_, ok := m.completed[id]
	return ok
}

// CompleteCurrent marks the active quest complete and returns XP + unlocked badges.
func (m *Manager) CompleteCurrent() (int, []string, error) {
	if m.current == "" {
		return 0, nil, errors.New("no quest selected")
	}
	var quest Quest
	found := false
	for _, q := range m.quests {
		if q.ID == m.current {
			quest = q
			found = true
			break
		}
	}
	if !found {
		return 0, nil, fmt.Errorf("quest %q not found", m.current)
	}

	now := m.now()
	m.completed[quest.ID] = now
	if m.player.Completed == nil {
		m.player.Completed = make(map[string]time.Time)
	}
	m.player.Completed[quest.ID] = now
	m.player.XP += quest.XP
	m.player.LastUpdated = now
	m.player.Level = (m.player.XP / 100) + 1
	m.RecordHistory(quest.ID, true)

	badges := []string{}
	if m.player.XP >= 100 {
		if m.player.Badges == nil {
			m.player.Badges = make(map[string]bool)
		}
		if !m.player.Badges["first-100"] {
			m.player.Badges["first-100"] = true
			badges = append(badges, "First 100 XP")
		}
	}
	if quest.Category == "Creative" {
		if m.player.Badges == nil {
			m.player.Badges = make(map[string]bool)
		}
		if !m.player.Badges["creative-spark"] {
			m.player.Badges["creative-spark"] = true
			badges = append(badges, "Creative spark")
		}
	}
	return quest.XP, badges, nil
}

// SkipCurrent records the quest as skipped and picks another one.
func (m *Manager) SkipCurrent() (Quest, error) {
	if m.current != "" {
		m.RecordHistory(m.current, false)
	}
	m.current = ""
	return m.NextQuest()
}

// Snapshot returns the current player state.
func (m *Manager) Snapshot() PlayerProgress {
	clone := PlayerProgress{
		XP:          m.player.XP,
		Level:       m.player.Level,
		Completed:   map[string]time.Time{},
		Badges:      map[string]bool{},
		LastUpdated: m.player.LastUpdated,
	}
	if m.player.Completed != nil {
		for id, ts := range m.player.Completed {
			clone.Completed[id] = ts
		}
	}
	if m.player.Badges != nil {
		for id, ok := range m.player.Badges {
			clone.Badges[id] = ok
		}
	}
	return clone
}
