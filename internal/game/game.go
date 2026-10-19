// Package game is the SideQuest state machine: it offers quests, tracks the
// active one, awards progress and persists everything after each change.
package game

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"SideQuest/internal/progress"
	"SideQuest/internal/quest"
)

const (
	saveVersion   = 1
	recentLimit   = 24
	logLimit      = 100
	maxPrefsItems = 16
)

// Current is the quest on the table right now.
type Current struct {
	QuestID    string    `json:"questId"`
	TwistID    string    `json:"twistId,omitempty"` // suggested twist, if any
	TwistOn    bool      `json:"twistOn"`
	Accepted   bool      `json:"accepted"`
	OfferedAt  time.Time `json:"offeredAt"`
	AcceptedAt time.Time `json:"acceptedAt,omitzero"`
}

// LogEntry records a completed quest.
type LogEntry struct {
	QuestID  string         `json:"questId"`
	Title    string         `json:"title"`
	Category quest.Category `json:"category"`
	XP       int            `json:"xp"`
	At       time.Time      `json:"at"`
}

// Position is the remembered window position, in screen pixels.
type Position struct {
	X     int  `json:"x"`
	Y     int  `json:"y"`
	Valid bool `json:"valid"`
}

// Save is everything persisted between sessions.
type Save struct {
	Version         int                    `json:"version"`
	XP              int                    `json:"xp"`
	TotalCompleted  int                    `json:"totalCompleted"`
	ByCategory      map[quest.Category]int `json:"byCategory"`
	TwistsCompleted int                    `json:"twistsCompleted"`
	SpicyCompleted  int                    `json:"spicyCompleted"`
	Badges          map[string]time.Time   `json:"badges"`
	LastCompleted   map[string]time.Time   `json:"lastCompleted"`
	LastSkipped     map[string]time.Time   `json:"lastSkipped"`
	RecentlyShown   []string               `json:"recentlyShown"`
	LastCategory    quest.Category         `json:"lastCategory,omitempty"`
	Log             []LogEntry             `json:"log"`
	Current         *Current               `json:"current,omitempty"`
	Prefs           quest.Prefs            `json:"prefs"`
	Window          Position               `json:"window"`
}

func (s *Save) fillDefaults() {
	s.Version = saveVersion
	if s.ByCategory == nil {
		s.ByCategory = map[quest.Category]int{}
	}
	if s.Badges == nil {
		s.Badges = map[string]time.Time{}
	}
	if s.LastCompleted == nil {
		s.LastCompleted = map[string]time.Time{}
	}
	if s.LastSkipped == nil {
		s.LastSkipped = map[string]time.Time{}
	}
}

// Persister stores the save data.
type Persister interface {
	Load(v any) (bool, error)
	Save(v any) error
}

// Errors returned to the UI.
var (
	ErrNoActiveQuest = errors.New("there is no accepted quest to complete")
	ErrBadPrefs      = errors.New("invalid preferences")
)

// Game is safe for concurrent use.
type Game struct {
	mu     sync.Mutex
	quests []quest.Quest
	byID   map[string]quest.Quest
	save   Save
	store  Persister
	now    func() time.Time
	rng    *rand.Rand
	logf   func(format string, args ...any)
}

// Options configure a Game. Zero values get sensible defaults.
type Options struct {
	Now  func() time.Time
	Rand *rand.Rand
	Logf func(format string, args ...any)
}

// New loads saved state and prepares the game.
func New(quests []quest.Quest, store Persister, opts Options) (*Game, error) {
	if len(quests) == 0 {
		return nil, quest.ErrNoQuests
	}
	g := &Game{
		quests: quests,
		byID:   make(map[string]quest.Quest, len(quests)),
		store:  store,
		now:    opts.Now,
		rng:    opts.Rand,
		logf:   opts.Logf,
	}
	if g.now == nil {
		g.now = time.Now
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	if g.logf == nil {
		g.logf = func(string, ...any) {}
	}
	for _, q := range quests {
		g.byID[q.ID] = q
	}

	if _, err := store.Load(&g.save); err != nil {
		g.logf("load state: %v", err)
		g.save = Save{}
	}
	g.save.fillDefaults()

	// Drop a current quest whose content disappeared (e.g. a removed pack).
	if c := g.save.Current; c != nil {
		if _, ok := g.byID[c.QuestID]; !ok {
			g.save.Current = nil
		}
	}
	return g, nil
}

// View is what the UI renders.
type View struct {
	Quest     *quest.Quest   `json:"quest"`
	Twist     *quest.Twist   `json:"twist"`
	TwistOn   bool           `json:"twistOn"`
	Accepted  bool           `json:"accepted"`
	Reward    int            `json:"reward"`
	XP        int            `json:"xp"`
	Level     progress.Level `json:"level"`
	Completed int            `json:"completed"`
	Badges    []BadgeView    `json:"badges"`
	Prefs     quest.Prefs    `json:"prefs"`
	Recent    []LogEntry     `json:"recent"`
}

// BadgeView is a badge plus whether it is earned.
type BadgeView struct {
	progress.Badge
	Earned bool `json:"earned"`
}

// Completion is returned after finishing a quest.
type Completion struct {
	View      View             `json:"view"`
	Gained    int              `json:"gained"`
	LeveledUp bool             `json:"leveledUp"`
	NewBadges []progress.Badge `json:"newBadges"`
}

// State returns the current view, making sure a quest is on offer.
func (g *Game) State() View {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.save.Current == nil {
		g.offer("")
		g.persist()
	}
	return g.view()
}

// Reroll swaps the offered quest for a different one, without penalty.
func (g *Game) Reroll() View {
	g.mu.Lock()
	defer g.mu.Unlock()
	exclude := ""
	if g.save.Current != nil {
		exclude = g.save.Current.QuestID
	}
	g.offer(exclude)
	g.persist()
	return g.view()
}

// Skip rests the current quest for a while and offers another.
func (g *Game) Skip() View {
	g.mu.Lock()
	defer g.mu.Unlock()
	exclude := ""
	if c := g.save.Current; c != nil {
		exclude = c.QuestID
		g.save.LastSkipped[c.QuestID] = g.now()
	}
	g.offer(exclude)
	g.persist()
	return g.view()
}

// Accept commits to the offered quest.
func (g *Game) Accept() View {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.save.Current == nil {
		g.offer("")
	}
	if !g.save.Current.Accepted {
		g.save.Current.Accepted = true
		g.save.Current.AcceptedAt = g.now()
	}
	g.persist()
	return g.view()
}

// Release puts an accepted quest back on offer. No penalty, no guilt.
func (g *Game) Release() View {
	g.mu.Lock()
	defer g.mu.Unlock()
	if c := g.save.Current; c != nil {
		c.Accepted = false
		c.AcceptedAt = time.Time{}
	}
	g.persist()
	return g.view()
}

// SetTwist toggles the suggested twist on or off.
func (g *Game) SetTwist(on bool) View {
	g.mu.Lock()
	defer g.mu.Unlock()
	if c := g.save.Current; c != nil && c.TwistID != "" {
		c.TwistOn = on
	}
	g.persist()
	return g.view()
}

// Complete finishes the accepted quest and awards XP and badges.
func (g *Game) Complete() (Completion, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	c := g.save.Current
	if c == nil || !c.Accepted {
		return Completion{}, ErrNoActiveQuest
	}
	q := g.byID[c.QuestID]
	gained := g.reward(c)
	now := g.now()
	before := progress.LevelFor(g.save.XP)

	s := &g.save
	s.XP += gained
	s.TotalCompleted++
	s.ByCategory[q.Category]++
	if c.TwistOn && c.TwistID != "" {
		s.TwistsCompleted++
	}
	if q.Difficulty >= 3 {
		s.SpicyCompleted++
	}
	s.LastCompleted[q.ID] = now
	s.Log = append(s.Log, LogEntry{QuestID: q.ID, Title: q.Title, Category: q.Category, XP: gained, At: now})
	if len(s.Log) > logLimit {
		s.Log = s.Log[len(s.Log)-logLimit:]
	}

	owned := make(map[string]bool, len(s.Badges))
	for id := range s.Badges {
		owned[id] = true
	}
	newBadges := progress.NewlyEarned(g.stats(), owned)
	for _, b := range newBadges {
		s.Badges[b.ID] = now
	}

	after := progress.LevelFor(s.XP)
	g.offer(q.ID)
	g.persist()

	return Completion{
		View:      g.view(),
		Gained:    gained,
		LeveledUp: after.Number > before.Number,
		NewBadges: newBadges,
	}, nil
}

// SetPrefs updates preferences. An un-accepted offer that no longer fits is
// replaced right away.
func (g *Game) SetPrefs(p quest.Prefs) (View, error) {
	if len(p.Muted) > maxPrefsItems || len(p.Avoid) > maxPrefsItems {
		return View{}, ErrBadPrefs
	}
	for _, c := range p.Muted {
		if !c.Valid() {
			return View{}, ErrBadPrefs
		}
	}
	for _, n := range p.Avoid {
		if !quest.KnownNeed(n) {
			return View{}, ErrBadPrefs
		}
	}
	if p.Muted == nil {
		p.Muted = []quest.Category{}
	}
	if p.Avoid == nil {
		p.Avoid = []quest.Need{}
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Prefs = p
	if c := g.save.Current; c != nil && !c.Accepted && !p.Allows(g.byID[c.QuestID]) {
		g.offer(c.QuestID)
	}
	g.persist()
	return g.view(), nil
}

// WindowPosition returns the remembered window position.
func (g *Game) WindowPosition() Position {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.save.Window
}

// SetWindowPosition remembers the window position and saves.
func (g *Game) SetWindowPosition(x, y int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Window = Position{X: x, Y: y, Valid: true}
	g.persist()
}

// --- internals (callers hold g.mu) ---

func (g *Game) offer(exclude string) {
	s := &g.save
	mem := quest.Memory{
		LastCompleted: s.LastCompleted,
		LastSkipped:   s.LastSkipped,
		RecentlyShown: s.RecentlyShown,
		LastCategory:  s.LastCategory,
	}
	o, err := quest.Pick(g.quests, s.Prefs, mem, g.now(), exclude, g.rng)
	if err != nil {
		// Unreachable with a non-empty quest list; keep the current one.
		g.logf("pick quest: %v", err)
		return
	}
	c := &Current{QuestID: o.Quest.ID, OfferedAt: g.now()}
	if o.Twist != nil {
		c.TwistID = o.Twist.ID
	}
	s.Current = c
	s.LastCategory = o.Quest.Category
	s.RecentlyShown = append(s.RecentlyShown, o.Quest.ID)
	if len(s.RecentlyShown) > recentLimit {
		s.RecentlyShown = s.RecentlyShown[len(s.RecentlyShown)-recentLimit:]
	}
}

func (g *Game) reward(c *Current) int {
	q := g.byID[c.QuestID]
	xp := q.XP
	if c.TwistOn {
		if t, ok := q.TwistByID(c.TwistID); ok {
			xp += t.Bonus
		}
	}
	return xp
}

func (g *Game) stats() progress.Stats {
	return progress.Stats{
		Completed:       g.save.TotalCompleted,
		ByCategory:      g.save.ByCategory,
		TwistsCompleted: g.save.TwistsCompleted,
		SpicyCompleted:  g.save.SpicyCompleted,
	}
}

func (g *Game) view() View {
	s := &g.save
	v := View{
		XP:        s.XP,
		Level:     progress.LevelFor(s.XP),
		Completed: s.TotalCompleted,
		Prefs:     s.Prefs,
	}
	if v.Prefs.Muted == nil {
		v.Prefs.Muted = []quest.Category{}
	}
	if v.Prefs.Avoid == nil {
		v.Prefs.Avoid = []quest.Need{}
	}
	if c := s.Current; c != nil {
		q := g.byID[c.QuestID]
		v.Quest = &q
		if t, ok := q.TwistByID(c.TwistID); ok {
			v.Twist = &t
		}
		v.TwistOn = c.TwistOn
		v.Accepted = c.Accepted
		v.Reward = g.reward(c)
	}
	for _, b := range progress.AllBadges() {
		_, earned := s.Badges[b.ID]
		v.Badges = append(v.Badges, BadgeView{Badge: b, Earned: earned})
	}
	for i := len(s.Log) - 1; i >= 0 && len(v.Recent) < 5; i-- {
		v.Recent = append(v.Recent, s.Log[i])
	}
	if v.Recent == nil {
		v.Recent = []LogEntry{}
	}
	return v
}

func (g *Game) persist() {
	if err := g.store.Save(&g.save); err != nil {
		g.logf("save state: %v", err)
	}
}
