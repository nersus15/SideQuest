# SideQuest

A tiny floating desktop companion that occasionally offers a playful, low-pressure side quest.
Boo the ghost lives in a small capsule. Click it to get a quest, then **Let's go!**, **Reroll** or **Skip**.
Finishing a quest earns XP, levels and a few badges. There are no streaks, timers or penalties.

Built with Go + [Wails v2](https://wails.io) + Vue 3. Works fully offline.

## Run & build

```bash
wails dev                               # live development (any OS)
wails build -platform windows/amd64     # Windows .exe in build/bin
go test ./internal/...                  # unit tests (the root package needs frontend/dist to exist)
```

To work on the UI in a normal browser, run `npm run dev` in `frontend/`. Without Wails, the UI uses an
in-memory mock backend (`frontend/src/lib/mock.js`).

## Layout

| Path | What lives there |
| --- | --- |
| `main.go` | Wails options: frameless, transparent, always-on-top, single instance |
| `app.go` | Bindings for the frontend, startup and shutdown, window sizes |
| `internal/quest` | Quest model, JSON pack parsing and validation, weighted quest selection |
| `internal/quest/builtin/quests.json` | The 54 built-in quests (embedded) |
| `internal/progress` | XP curve, level titles, badges |
| `internal/game` | State machine (offer → accept → complete) and save data |
| `internal/store` | Atomic JSON file persistence |
| `internal/desktop` | Window geometry; Win32 controller (`window_windows.go`) and dev fallback |
| `frontend/src` | Vue UI: `Capsule`, `QuestCard`, `Celebration`, `Backpack`, `Mascot` |

## How quests are picked

Each quest has a category, duration, difficulty, needs (`paper`, `people`, `move`, `outside`, `phone`, `sound`),
XP, repeatability, cooldown and optional **twists**. A twist is a real variation you can opt into for bonus XP.

`quest.Pick` runs weighted random selection over tiers that it relaxes one by one, so there is always a quest:

1. Matches preferences, is not cooling down, was not skipped in the last 12 hours, and was not among the last few shown
2. The same, but recently shown quests are allowed
3. The same, but recently skipped quests are allowed
4. Preferences are ignored
5. Anything

The quest currently on offer is never offered again by a reroll unless it is the only quest left. The weights favour
a different category from the last one and quests you have never finished.

**Reroll** has no cost. **Skip** hides that quest for 12 hours. **Maybe later** un-accepts a quest with no penalty.

## Quest packs

Drop `*.json` packs into the SideQuest data folder's `packs/` directory and restart:

- Windows: `%AppData%\SideQuest\packs\`
- macOS: `~/Library/Application Support/SideQuest/packs/`

```json
{
  "format": 1,
  "id": "my-pack",
  "name": "My Pack",
  "quests": [
    { "id": "sock-puppet", "title": "Sock puppet", "description": "Make a sock puppet and give it opinions.",
      "category": "wildcard", "minutes": 3, "difficulty": 1, "xp": 15, "repeatable": true, "cooldownHours": 48,
      "twists": [{ "id": "accent", "text": "It has a dramatic accent.", "bonus": 5 }] }
  ]
}
```

Packs are data only and are strictly validated:

- Unknown fields are rejected.
- IDs, categories and needs must come from a whitelist.
- Lengths and numbers are bounded, and a pack can be at most 512 KB.
- The `source` field is set by the app, not taken from the file.
- Packs can never override built-in quests. Invalid packs are skipped.

Nothing in a pack is ever executed.

## Windows notes

- The window is frameless with a transparent WebView2, so only the capsule or panel is visible. The DWM frame is disabled.
- `WS_EX_TOOLWINDOW` keeps SideQuest out of the taskbar and Alt+Tab. You can close it from the Backpack.
- Wails' `WindowSetPosition` is relative to the work area while `WindowGetPosition` is absolute.
  So on Windows, resizing and positioning use Win32 directly and are DPI-aware.
  When the panel opens, it grows away from the nearest screen corner and always stays inside the work area.
- The capsule position is saved every 15 seconds, on collapse, and on close. It is clamped back on screen at startup
  if a monitor was removed.

Save data: `%AppData%\SideQuest\state.json`. It is written atomically. A corrupt file is moved aside and the app
starts fresh.
