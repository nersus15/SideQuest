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
	"SideQuest/internal/schedule"
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
	Lang            string                 `json:"lang,omitempty"` // "" until chosen
	Onboarded       bool                   `json:"onboarded"`
	Look            Look                   `json:"look"`
	// Ratings are thumbs up or down by quest id; they tune selection.
	Ratings map[string]quest.Rating `json:"ratings"`
	Nudges  Nudges                  `json:"nudges"`
	// Schedule offers quests at set times; LastScheduled is the last slot
	// that fired, so each one fires once even across restarts.
	Schedule      schedule.Schedule `json:"schedule"`
	LastScheduled string            `json:"lastScheduled,omitempty"`
	Sound         Sound             `json:"sound"`
}

// Sound controls the little sound effects.
type Sound struct {
	Volume int  `json:"volume"` // 10–100
	Muted  bool `json:"muted"`
	// Nudge also plays a soft sound when Boo speaks up from the capsule.
	// Off by default: it is the one sound that interrupts.
	Nudge bool `json:"nudge"`
}

// DefaultVolume is used until the player picks one.
const DefaultVolume = 60

// Nudges control when Boo speaks up.
type Nudges struct {
	// BreakMinutes turns on a break reminder every N minutes (0 = off,
	// Boo then perks up now and then at random).
	BreakMinutes int `json:"breakMinutes"`
	// StayWhenBusy keeps the capsule visible over fullscreen apps, games
	// and presentations. By default it hides until they are gone.
	StayWhenBusy bool `json:"stayWhenBusy"`
}

// BreakChoices are the allowed break reminder intervals, in minutes.
var BreakChoices = []int{0, 60, 90}

// Look is what Boo is wearing.
type Look struct {
	Hat   string `json:"hat"`
	Extra string `json:"extra"` // "" for nothing
}

func (s *Save) fillDefaults() {
	s.Version = saveVersion
	if s.Sound.Volume == 0 {
		s.Sound.Volume = DefaultVolume
	}
	if s.Schedule.Days == "" {
		s.Schedule.Days = schedule.Daily
	}
	if s.Schedule.Times == nil {
		s.Schedule.Times = []string{}
	}
	if s.Look.Hat == "" {
		s.Look.Hat = progress.DefaultHat
	}
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
	if s.Ratings == nil {
		s.Ratings = map[string]quest.Rating{}
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
	ErrUnknownQuest  = errors.New("that quest no longer exists")
	ErrQuestActive   = errors.New("finish or put back your current quest first")
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
	Quest     *quest.Quest      `json:"quest"`
	Twist     *quest.Twist      `json:"twist"`
	TwistOn   bool              `json:"twistOn"`
	Accepted  bool              `json:"accepted"`
	Reward    int               `json:"reward"`
	XP        int               `json:"xp"`
	Level     progress.Level    `json:"level"`
	Completed int               `json:"completed"`
	Badges    []BadgeView       `json:"badges"`
	Prefs     quest.Prefs       `json:"prefs"`
	Recent    []LogEntry        `json:"recent"`
	Lang      string            `json:"lang"`
	Onboarded bool              `json:"onboarded"`
	Look      Look              `json:"look"`
	Looks     []LookView        `json:"looks"`
	Nudges    Nudges            `json:"nudges"`
	Schedule  schedule.Schedule `json:"schedule"`
	Sound     Sound             `json:"sound"`
	// MaxDifficulty is the hardest unlocked difficulty; NextDifficultyAt is
	// the level that unlocks the next one (0 when all are open).
	MaxDifficulty    int `json:"maxDifficulty"`
	NextDifficultyAt int `json:"nextDifficultyAt"`
}

// LookView is a look plus whether it is unlocked.
type LookView struct {
	progress.Look
	Unlocked bool `json:"unlocked"`
}

// BadgeView is a badge plus whether it is earned.
type BadgeView struct {
	progress.Badge
	Earned bool `json:"earned"`
}

// Completion is returned after finishing a quest.
type Completion struct {
	View      View             `json:"view"`
	Done      quest.Quest      `json:"done"`   // the quest just completed
	Rating    quest.Rating     `json:"rating"` // its current rating, if any
	Gained    int              `json:"gained"`
	LeveledUp bool             `json:"leveledUp"`
	NewBadges []progress.Badge `json:"newBadges"`
	// Things the level-up unlocked.
	NewLooks      []progress.Look `json:"newLooks"`
	NewDifficulty int             `json:"newDifficulty"` // 0 if none
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

	res := Completion{
		Done:      q,
		Rating:    s.Ratings[q.ID],
		View:      g.view(),
		Gained:    gained,
		LeveledUp: after.Number > before.Number,
		NewBadges: newBadges,
		NewLooks:  progress.LooksUnlocked(before.Number, after.Number),
	}
	if d := progress.MaxDifficulty(after.Number); d > progress.MaxDifficulty(before.Number) {
		res.NewDifficulty = d
	}
	if res.NewBadges == nil {
		res.NewBadges = []progress.Badge{}
	}
	if res.NewLooks == nil {
		res.NewLooks = []progress.Look{}
	}
	return res, nil
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

// SetQuests swaps the quest content (e.g. after the player edits their own
// quests). Progress is kept; an offer whose quest disappeared is replaced.
func (g *Game) SetQuests(quests []quest.Quest) error {
	if len(quests) == 0 {
		return quest.ErrNoQuests
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.quests = quests
	g.byID = make(map[string]quest.Quest, len(quests))
	for _, q := range quests {
		g.byID[q.ID] = q
	}
	if c := g.save.Current; c != nil {
		if _, ok := g.byID[c.QuestID]; !ok {
			g.save.Current = nil
			g.offer(c.QuestID)
			g.persist()
		}
	}
	return nil
}

// Play offers a specific quest right away, e.g. one the player just wrote.
// It never replaces a quest the player has already accepted.
func (g *Game) Play(id string) (View, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	q, ok := g.byID[id]
	if !ok {
		return View{}, ErrUnknownQuest
	}
	if c := g.save.Current; c != nil && c.Accepted && c.QuestID != id {
		return View{}, ErrQuestActive
	}
	if c := g.save.Current; c == nil || c.QuestID != id {
		g.save.Current = &Current{QuestID: q.ID, OfferedAt: g.now()}
		g.save.LastCategory = q.Category
	}
	g.persist()
	return g.view(), nil
}

// ErrBadNudges is returned for an unsupported break interval.
var ErrBadNudges = errors.New("invalid nudge settings")

// SetNudges updates when Boo speaks up.
func (g *Game) SetNudges(n Nudges) (View, error) {
	ok := false
	for _, m := range BreakChoices {
		ok = ok || n.BreakMinutes == m
	}
	if !ok {
		return View{}, ErrBadNudges
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Nudges = n
	g.persist()
	return g.view(), nil
}

// NudgeSettings returns the current nudge settings.
func (g *Game) NudgeSettings() Nudges {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.save.Nudges
}

// ErrBadSound is returned for a volume outside 10–100.
var ErrBadSound = errors.New("invalid sound settings")

// SetSound updates the sound effect settings.
func (g *Game) SetSound(s Sound) (View, error) {
	if s.Volume < 10 || s.Volume > 100 {
		return View{}, ErrBadSound
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Sound = s
	g.persist()
	return g.view(), nil
}

// SetSchedule updates the quest schedule.
func (g *Game) SetSchedule(s schedule.Schedule) (View, error) {
	s, err := schedule.Normalize(s)
	if err != nil {
		return View{}, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Schedule = s
	g.persist()
	return g.view(), nil
}

// ScheduleSettings returns the quest schedule.
func (g *Game) ScheduleSettings() schedule.Schedule {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.save.Schedule
}

// MarkScheduled records that the slot with this key fired. It returns false
// if it already had, so a slot never fires twice.
func (g *Game) MarkScheduled(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.save.LastScheduled == key {
		return false
	}
	g.save.LastScheduled = key
	g.persist()
	return true
}

// ErrBadRating is returned for anything but "up", "down" or "".
var ErrBadRating = errors.New("invalid rating")

// RateQuest stores a thumbs up or down for a quest ("" clears it). Ratings
// only change how often quests come up; they never hide one for good.
func (g *Game) RateQuest(id string, r quest.Rating) error {
	if !r.Valid() {
		return ErrBadRating
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.byID[id]; !ok {
		return ErrUnknownQuest
	}
	if r == "" {
		delete(g.save.Ratings, id)
	} else {
		g.save.Ratings[id] = r
	}
	g.persist()
	return nil
}

// Errors for appearance and language settings.
var (
	ErrBadLanguage = errors.New("unknown language")
	ErrLookLocked  = errors.New("that look is not unlocked yet")
)

// SetLanguage stores the player's language ("en" or "id").
func (g *Game) SetLanguage(lang string) (View, error) {
	ok := false
	for _, l := range quest.Languages {
		ok = ok || l == lang
	}
	if !ok {
		return View{}, ErrBadLanguage
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Lang = lang
	g.persist()
	return g.view(), nil
}

// Language returns the stored language ("" if never chosen).
func (g *Game) Language() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.save.Lang
}

// FinishOnboarding remembers that the intro was shown.
func (g *Game) FinishOnboarding() View {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.save.Onboarded = true
	g.persist()
	return g.view()
}

// SetLook dresses Boo. Both items must be unlocked; extra may be empty.
func (g *Game) SetLook(l Look) (View, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	level := progress.LevelFor(g.save.XP).Number
	check := func(id, kind string) error {
		look, ok := progress.LookByID(id)
		if !ok || look.Kind != kind || look.Level > level {
			return ErrLookLocked
		}
		return nil
	}
	if err := check(l.Hat, "hat"); err != nil {
		return View{}, err
	}
	if l.Extra != "" {
		if err := check(l.Extra, "extra"); err != nil {
			return View{}, err
		}
	}
	g.save.Look = l
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
		Ratings:       s.Ratings,
	}
	prefs := s.Prefs
	prefs.MaxDifficulty = progress.MaxDifficulty(progress.LevelFor(s.XP).Number)
	o, err := quest.Pick(g.quests, prefs, mem, g.now(), exclude, g.rng)
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
	lvl := progress.LevelFor(s.XP)
	v := View{
		XP:            s.XP,
		Level:         lvl,
		Completed:     s.TotalCompleted,
		Prefs:         s.Prefs,
		Lang:          s.Lang,
		Onboarded:     s.Onboarded,
		Look:          s.Look,
		Nudges:        s.Nudges,
		Schedule:      s.Schedule,
		Sound:         s.Sound,
		MaxDifficulty: progress.MaxDifficulty(lvl.Number),
	}
	if v.MaxDifficulty < 3 {
		v.NextDifficultyAt = progress.DifficultyLevel(v.MaxDifficulty + 1)
	}
	for _, l := range progress.Looks() {
		v.Looks = append(v.Looks, LookView{Look: l, Unlocked: l.Level <= lvl.Number})
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
