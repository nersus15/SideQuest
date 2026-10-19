package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"SideQuest/internal/desktop"
	"SideQuest/internal/game"
	"SideQuest/internal/quest"
	"SideQuest/internal/tray"
)

// Window sizes in CSS pixels. They include a transparent margin so the
// capsule and panel shadows are not clipped; keep in sync with style.css.
const (
	capsuleW, capsuleH = 208, 84
	panelW, panelH     = 380, 470
)

// App is the bridge between the Vue frontend and the game.
type App struct {
	ctx      context.Context
	game     *game.Game
	window   desktop.Window
	expanded atomic.Bool

	spotMu   sync.Mutex
	lastSpot desktop.Point
}

// NewApp wires the game to its save file and quest content.
func NewApp(dataDir string, window desktop.Window) (*App, error) {
	quests, err := loadQuests(filepath.Join(dataDir, "packs"))
	if err != nil {
		return nil, err
	}
	g, err := game.New(quests, newStore(dataDir), game.Options{Logf: log.Printf})
	if err != nil {
		return nil, err
	}
	return &App{game: g, window: window}, nil
}

// loadQuests returns the built-in quests plus any valid packs dropped into
// dir. Packs are data only; invalid ones are skipped and logged.
func loadQuests(dir string) ([]quest.Quest, error) {
	builtin, err := quest.Builtin()
	if err != nil {
		return nil, err
	}
	extra, problems := quest.LoadDir(dir, quest.SourceImported)
	for _, p := range problems {
		log.Printf("skipped quest pack %v", p)
	}
	return quest.Merge(builtin, extra), nil
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
	a.window.Place(ctx, saved, capsuleW, capsuleH)
	a.window.Show(ctx)

	tray.Start(tray.Options{
		Tooltip: "SideQuest",
		Icon:    trayIcon,
		OnClick: a.open,
		Items: []tray.Item{
			{Label: "Open SideQuest", OnClick: a.open},
			{Label: "Hide Boo", OnClick: func() { runtime.EventsEmit(a.ctx, "sidequest:hide") }},
			{},
			{Label: "Quit SideQuest", OnClick: func() { runtime.Quit(a.ctx) }},
		},
	})
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
	p, ok := a.window.Spot(a.ctx, capsuleW, capsuleH)
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

// Complete finishes the accepted quest.
func (a *App) Complete() (game.Completion, error) { return a.game.Complete() }

// SetPrefs updates quest preferences.
func (a *App) SetPrefs(p quest.Prefs) (game.View, error) { return a.game.SetPrefs(p) }

// Expand grows the window into the quest panel.
func (a *App) Expand() {
	if a.expanded.Swap(true) {
		return
	}
	a.window.Resize(a.ctx, panelW, panelH)
}

// Collapse shrinks the window back into the capsule.
func (a *App) Collapse() {
	if !a.expanded.Swap(false) {
		return
	}
	a.window.Resize(a.ctx, capsuleW, capsuleH)
	a.rememberSpot()
}

// Hide tucks the window away; the tray or menu bar brings it back.
func (a *App) Hide() {
	runtime.WindowHide(a.ctx)
}
