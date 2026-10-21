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
	"SideQuest/internal/store"
	"SideQuest/internal/tray"
)

// Window sizes in CSS pixels. They include a transparent margin so the
// capsule and panel shadows are not clipped; keep in sync with style.css.
const (
	capsuleW, capsuleH = 156, 62
	panelW, panelH     = 380, 520
)

// App is the bridge between the Vue frontend and the game.
type App struct {
	ctx      context.Context
	game     *game.Game
	window   desktop.Window
	expanded atomic.Bool

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
	runtime.WindowHide(a.ctx)
}
