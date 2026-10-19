package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"SideQuest/internal/questbook"
)

// CompletionPayload is returned after a quest is finished.
type CompletionPayload struct {
	Quest    questbook.Quest          `json:"quest"`
	XP       int                      `json:"xp"`
	Level    int                      `json:"level"`
	Badges   []string                 `json:"badges"`
	Progress questbook.PlayerProgress `json:"progress"`
}

// App holds the desktop shell and quest manager.
type App struct {
	ctx       context.Context
	manager   *questbook.Manager
	statePath string
}

type persistedState struct {
	CurrentQuestID string                   `json:"currentQuestId"`
	Progress       questbook.PlayerProgress `json:"progress"`
}

// NewApp creates a new App application struct.
func NewApp() *App {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	statePath := filepath.Join(configDir, "SideQuest", "state.json")
	manager := questbook.NewManager(nil, time.Now)
	return &App{manager: manager, statePath: statePath}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadState()
}

func (a *App) loadState() {
	if _, err := os.Stat(a.statePath); err != nil {
		if a.manager != nil {
			_, _ = a.manager.NextQuest()
		}
		return
	}
	data, err := os.ReadFile(a.statePath)
	if err != nil {
		return
	}
	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	if a.manager != nil {
		a.manager.SetProgress(state.Progress)
		if state.CurrentQuestID != "" {
			_ = a.manager.SetCurrent(state.CurrentQuestID)
		}
	}
}

func (a *App) persistState() error {
	if err := os.MkdirAll(filepath.Dir(a.statePath), 0o755); err != nil {
		return err
	}
	state := persistedState{
		CurrentQuestID: a.manager.CurrentQuestID(),
		Progress:       a.manager.Snapshot(),
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.statePath, data, 0o644)
}

// GetState exposes the current player state and active quest.
func (a *App) GetState() (map[string]interface{}, error) {
	state := map[string]interface{}{}
	progress := a.manager.Snapshot()
	state["xp"] = progress.XP
	state["level"] = progress.Level
	state["badges"] = sortedBadgeNames(progress.Badges)
	state["completedCount"] = len(progress.Completed)
	quest, ok := a.manager.Current()
	if ok {
		state["quest"] = quest
	} else {
		quest, err := a.manager.NextQuest()
		if err == nil {
			state["quest"] = quest
		} else {
			state["quest"] = nil
		}
	}
	return state, nil
}

// NextQuest selects a new recommended quest.
func (a *App) NextQuest() (questbook.Quest, error) {
	quest, err := a.manager.NextQuest()
	if err != nil {
		return questbook.Quest{}, err
	}
	if err := a.persistState(); err != nil {
		return quest, err
	}
	return quest, nil
}

// AcceptCurrent keeps the active quest for later completion.
func (a *App) AcceptCurrent() (questbook.Quest, error) {
	quest, ok := a.manager.Current()
	if !ok {
		var err error
		quest, err = a.manager.NextQuest()
		if err != nil {
			return questbook.Quest{}, err
		}
	}
	if err := a.persistState(); err != nil {
		return quest, err
	}
	return quest, nil
}

// SkipCurrent marks the quest as skipped and picks a replacement.
func (a *App) SkipCurrent() (questbook.Quest, error) {
	quest, err := a.manager.SkipCurrent()
	if err != nil {
		return questbook.Quest{}, err
	}
	if err := a.persistState(); err != nil {
		return quest, err
	}
	return quest, nil
}

// CompleteCurrent completes the current quest and rewards XP.
func (a *App) CompleteCurrent() (CompletionPayload, error) {
	xp, badges, err := a.manager.CompleteCurrent()
	if err != nil {
		return CompletionPayload{}, err
	}
	quest, _ := a.manager.Current()
	progress := a.manager.Snapshot()
	payload := CompletionPayload{
		Quest:    quest,
		XP:       xp,
		Level:    progress.Level,
		Badges:   badges,
		Progress: progress,
	}
	if err := a.persistState(); err != nil {
		return payload, err
	}
	return payload, nil
}

// RerollQuest picks a different quest when possible.
func (a *App) RerollQuest() (questbook.Quest, error) {
	quest, err := a.manager.NextQuest()
	if err != nil {
		return questbook.Quest{}, err
	}
	if err := a.persistState(); err != nil {
		return quest, err
	}
	return quest, nil
}

func sortedBadgeNames(badges map[string]bool) []string {
	keys := make([]string, 0, len(badges))
	for key := range badges {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Greet returns a greeting for the given name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, it’s show time!", name)
}
