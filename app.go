package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"SideQuest/internal/custom"
	"SideQuest/internal/desktop"
	"SideQuest/internal/game"
	"SideQuest/internal/journal"
	"SideQuest/internal/quest"
	"SideQuest/internal/schedule"
	"SideQuest/internal/store"
	"SideQuest/internal/tray"
)

// Window sizes in CSS pixels. They include a transparent margin so the
// capsule and panel shadows are not clipped; keep in sync with style.css.
const (
	capsuleW, capsuleH = 156, 62
	panelW, panelH     = 380, 520

	// Shapes inside the window, in CSS pixels; keep in sync with style.css
	// (--gutter, --radius) and Capsule.vue.
	gutter        = 8
	capsuleRadius = 23
	panelRadius   = 26
)

// App is the bridge between the Vue frontend and the game.
type App struct {
	ctx      context.Context
	game     *game.Game
	window   desktop.Window
	expanded atomic.Bool

	// busy is true while a fullscreen app, game or presentation is in
	// front; hiddenForBusy remembers that Boo stepped aside because of it.
	busy          atomic.Bool
	hiddenForBusy atomic.Bool

	notifyInit sync.Once

	spotMu   sync.Mutex
	lastSpot desktop.Point

	// Quest content: built-in and imported packs are fixed at startup; the
	// player's own quests change while the app runs.
	base   []quest.Quest
	extra  []quest.Quest
	custom *custom.Library

	journal *journal.Journal
}

// NewApp wires the game to its save file and quest content.
func NewApp(dataDir string, window desktop.Window) (*App, error) {
	base, err := quest.Builtin()
	if err != nil {
		return nil, err
	}
	extra, problems := quest.LoadDir(filepath.Join(dataDir, "packs"), quest.SourceImported)
	for _, p := range problems {
		log.Printf("skipped quest pack %v", p)
	}
	lib, err := custom.Open(store.File{Path: filepath.Join(dataDir, "custom.json")})
	if err != nil {
		log.Printf("custom quests: %v", err)
	}

	jr, err := journal.Open(store.File{Path: filepath.Join(dataDir, "journal.json")})
	if err != nil {
		log.Printf("journal: %v", err)
	}

	a := &App{window: window, base: base, extra: extra, custom: lib, journal: jr}
	a.game, err = game.New(a.allQuests(), newStore(dataDir), game.Options{Logf: log.Printf})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// allQuests merges the content. Built-ins win id collisions, then the
// player's quests, then imported packs.
func (a *App) allQuests() []quest.Quest {
	return quest.Merge(a.base, a.custom.List(), a.extra)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Native window drags never reach the frontend, so check the position
	// now and then; this also covers the app being killed abruptly.
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				a.rememberSpot()
			}
		}
	}()
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				a.checkBusy()
				a.checkSchedule()
			}
		}
	}()
}

// checkSchedule offers a quest when a scheduled time comes up. A slot that
// arrives while the screen is busy waits (up to schedule.Grace), and one
// that is missed is simply skipped.
func (a *App) checkSchedule() {
	s := a.game.ScheduleSettings()
	key, due := s.Due(time.Now())
	if !due || a.busy.Load() || !a.game.MarkScheduled(key) {
		return
	}
	// The player set this time on purpose, so Boo comes back if hidden.
	a.hiddenForBusy.Store(false)
	a.window.Show(a.ctx)
	runtime.EventsEmit(a.ctx, "sidequest:scheduled")
	if s.Notify {
		a.notify()
	}
}

// initNotifications sets up system notifications once; clicking one opens
// the panel.
func (a *App) initNotifications() {
	a.notifyInit.Do(func() {
		if err := runtime.InitializeNotifications(a.ctx); err != nil {
			log.Printf("notifications: %v", err)
			return
		}
		runtime.OnNotificationResponse(a.ctx, func(runtime.NotificationResult) { a.open() })
	})
}

// notify sends a gentle system notification; clicking it opens Boo.
func (a *App) notify() {
	a.initNotifications()
	title, body := "SideQuest", "Boo has a tiny quest for you 👻"
	if a.game.Language() == "id" {
		body = "Boo punya quest kecil buat kamu 👻"
	}
	err := runtime.SendNotification(a.ctx, runtime.NotificationOptions{
		ID: "sidequest-schedule", Title: title, Body: body,
	})
	if err != nil {
		log.Printf("notification: %v", err)
	}
}

// checkBusy keeps Boo out of the way during fullscreen apps, games and
// presentations: no nudges, and the capsule steps aside unless the player
// asked it to stay (or has the panel open).
func (a *App) checkBusy() {
	busy := desktop.Busy()
	if a.busy.Swap(busy) == busy {
		return
	}
	runtime.EventsEmit(a.ctx, "sidequest:busy", busy)
	if busy {
		if !a.game.NudgeSettings().StayWhenBusy && !a.expanded.Load() {
			a.hiddenForBusy.Store(true)
			runtime.WindowHide(a.ctx)
		}
		return
	}
	if a.hiddenForBusy.Swap(false) {
		a.window.Show(a.ctx)
	}
}

// domReady positions the hidden window, then reveals it.
func (a *App) domReady(ctx context.Context) {
	var saved *desktop.Point
	if pos := a.game.WindowPosition(); pos.Valid {
		saved = &desktop.Point{X: pos.X, Y: pos.Y}
	}
	a.window.Place(ctx, saved)
	// The UI pins the capsule to this corner; tell it before showing.
	runtime.EventsEmit(ctx, "sidequest:anchor", a.window.Anchor(ctx))
	a.window.Show(ctx)

	labels := trayLabels(a.game.Language())
	tray.Start(tray.Options{
		Tooltip: "SideQuest",
		Icon:    trayIcon,
		OnClick: a.open,
		Items: []tray.Item{
			{Label: labels[0], OnClick: a.open},
			{Label: labels[1], OnClick: func() { runtime.EventsEmit(a.ctx, "sidequest:hide") }},
			{},
			{Label: labels[3], OnClick: func() { runtime.Quit(a.ctx) }},
		},
	})
}

// trayLabels returns the tray menu text in the player's language.
func trayLabels(lang string) []string {
	if lang == "id" {
		return []string{"Buka SideQuest", "Sembunyikan SideQuest", "", "Keluar dari SideQuest"}
	}
	return []string{"Open SideQuest", "Hide SideQuest", "", "Quit SideQuest"}
}

// open shows the window (if hidden) and asks the UI to open the panel.
func (a *App) open() {
	if a.ctx == nil {
		return
	}
	a.hiddenForBusy.Store(false)
	runtime.WindowShow(a.ctx)
	runtime.EventsEmit(a.ctx, "sidequest:poke")
}

func (a *App) shutdown(context.Context) {
	tray.Stop()
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.rememberSpot()
	return false
}

// rememberSpot saves where the capsule sits (or will sit once collapsed).
func (a *App) rememberSpot() {
	if a.ctx == nil {
		return
	}
	p, ok := a.window.CapsuleSpot(a.ctx)
	if !ok {
		return
	}
	a.spotMu.Lock()
	defer a.spotMu.Unlock()
	if pos := a.game.WindowPosition(); pos.Valid && p == a.lastSpot {
		return
	}
	a.lastSpot = p
	a.game.SetWindowPosition(p.X, p.Y)
}

// --- bindings ---

// State returns the current quest and progress.
func (a *App) State() game.View { return a.game.State() }

// Reroll offers a different quest.
func (a *App) Reroll() game.View { return a.game.Reroll() }

// Skip rests the offered quest for a while and offers another.
func (a *App) Skip() game.View { return a.game.Skip() }

// Accept commits to the offered quest.
func (a *App) Accept() game.View { return a.game.Accept() }

// Release un-accepts the quest without any penalty.
func (a *App) Release() game.View { return a.game.Release() }

// SetTwist toggles the suggested twist.
func (a *App) SetTwist(on bool) game.View { return a.game.SetTwist(on) }

// CompleteResult is a completion plus the journal entry it created.
type CompleteResult struct {
	game.Completion
	Entry string `json:"entry"` // journal entry id, for the mood check-in
}

// Complete finishes the accepted quest and records what the player wrote,
// counted or drew in the journal.
func (a *App) Complete(in journal.Input) (CompleteResult, error) {
	// Check the input first so a bad payload never half-completes a quest.
	if _, err := journal.Clean(in); err != nil {
		return CompleteResult{}, err
	}
	res, err := a.game.Complete()
	if err != nil {
		return CompleteResult{}, err
	}
	out := CompleteResult{Completion: res}
	if e, err := a.journal.Add(res.Done, res.Gained, time.Now(), in); err != nil {
		log.Printf("journal: %v", err) // progress is saved either way
	} else {
		out.Entry = e.ID
	}
	return out, nil
}

// JournalView is the journal plus a gentle mood observation.
type JournalView struct {
	Entries []journal.Entry `json:"entries"`
	Insight journal.Insight `json:"insight"`
}

// Journal lists what the player has done, newest first.
func (a *App) Journal() JournalView {
	return JournalView{Entries: a.journal.List(), Insight: a.journal.Insight()}
}

// RateQuest stores a thumbs up ("up"), down ("down") or clears it ("").
func (a *App) RateQuest(id string, rating quest.Rating) error { return a.game.RateQuest(id, rating) }

// SetNudges updates the break reminder and busy-screen behaviour.
func (a *App) SetNudges(n game.Nudges) (game.View, error) { return a.game.SetNudges(n) }

// SetSchedule updates when Boo offers quests on its own. Turning
// notifications on asks the system for permission where needed (macOS).
func (a *App) SetSchedule(s schedule.Schedule) (game.View, error) {
	before := a.game.ScheduleSettings()
	v, err := a.game.SetSchedule(s)
	if err == nil && s.Notify && !before.Notify {
		go func() {
			a.initNotifications()
			if _, err := runtime.RequestNotificationAuthorization(a.ctx); err != nil {
				log.Printf("notification permission: %v", err)
			}
		}()
	}
	return v, err
}

// Busy reports whether a fullscreen app, game or presentation is in front.
func (a *App) Busy() bool { return a.busy.Load() }

// SetMood records how the player felt after a quest.
func (a *App) SetMood(entry, mood string) error { return a.journal.SetMood(entry, mood) }

// SetLanguage switches the app language, including the tray menu.
func (a *App) SetLanguage(lang string) (game.View, error) {
	v, err := a.game.SetLanguage(lang)
	if err == nil {
		tray.SetLabels(trayLabels(lang))
	}
	return v, err
}

// FinishOnboarding remembers that the intro was shown.
func (a *App) FinishOnboarding() game.View { return a.game.FinishOnboarding() }

// SetLook dresses Boo with unlocked items.
func (a *App) SetLook(l game.Look) (game.View, error) { return a.game.SetLook(l) }

// SetPrefs updates quest preferences.
func (a *App) SetPrefs(p quest.Prefs) (game.View, error) { return a.game.SetPrefs(p) }

// Anchor returns the window corner the capsule sits in and the panel grows
// from. The UI pins the capsule to it so nothing jumps.
func (a *App) Anchor() desktop.Anchor { return a.window.Anchor(a.ctx) }

// Expand makes room for the quest panel.
func (a *App) Expand() {
	if a.expanded.Swap(true) {
		return
	}
	a.window.Expand(a.ctx)
}

// Collapse goes back to the capsule.
func (a *App) Collapse() {
	if !a.expanded.Swap(false) {
		return
	}
	a.window.Collapse(a.ctx)
	a.rememberSpot()
}

// CustomQuests lists the player's own quests, newest first.
func (a *App) CustomQuests() []quest.Quest { return a.custom.List() }

// CustomReward previews the XP a custom quest would give.
func (a *App) CustomReward(minutes, difficulty int) int {
	return custom.Reward(minutes, difficulty)
}

// SaveCustom creates or edits one of the player's quests.
func (a *App) SaveCustom(d custom.Draft) (quest.Quest, error) {
	q, err := a.custom.Save(d)
	if err != nil {
		return quest.Quest{}, err
	}
	return q, a.game.SetQuests(a.allQuests())
}

// DeleteCustom removes one of the player's quests.
func (a *App) DeleteCustom(id string) error {
	if err := a.custom.Delete(id); err != nil {
		return err
	}
	return a.game.SetQuests(a.allQuests())
}

// PlayQuest offers a specific quest right away ("Try it now").
func (a *App) PlayQuest(id string) (game.View, error) { return a.game.Play(id) }

// Hide tucks the window away; the tray or menu bar brings it back.
func (a *App) Hide() {
	a.hiddenForBusy.Store(false) // hidden on purpose: stay hidden
	runtime.WindowHide(a.ctx)
}
