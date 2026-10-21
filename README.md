# SideQuest

A tiny floating desktop companion that occasionally offers a playful, low-pressure side quest.
Boo the ghost lives in a small capsule. Click it to get a quest, then **Let's go!**, **Reroll** or **Skip**.
Many quests are played right in the app: write a short answer, tick a checklist, follow a timer or breathing guide,
tap a counter or doodle. Finishing a quest earns XP. Levelling up unlocks outfits for Boo and bolder quests.
There are no streaks and no penalties. The app speaks Indonesian and English.

Built with Go + [Wails v2](https://wails.io) + Vue 3. Works fully offline.

## Run & build

```bash
wails dev                               # live development (any OS)
wails build -platform windows/amd64     # Windows .exe in build/bin
go test ./internal/...                  # unit tests (the root package needs frontend/dist to exist)
```

To work on the UI in a normal browser, run `npm run dev` in `frontend/`. Without Wails, the UI uses an
in-memory mock backend (`frontend/src/lib/mock.js`) that plays the real built-in quests. Preview helpers:
`?fresh` shows the first-run intro, `?level=5` starts at a level, `?quest=<id>` starts on a quest, and
`?anchor=br` puts the capsule in another corner. The mock is never included in release builds.

## Layout

| Path | What lives there |
| --- | --- |
| `main.go` | Wails options: frameless, transparent, always-on-top, single instance |
| `app.go` | Bindings for the frontend, startup and shutdown, window sizes |
| `internal/quest` | Quest model, JSON pack parsing and validation, weighted quest selection |
| `internal/quest/builtin/*.json` | 739 built-in quests in English and Indonesian, one pack per category (embedded) |
| `internal/custom` | The player's own quests: create, edit, delete, saved as a pack |
| `internal/progress` | XP curve, badges, difficulty unlocks, Boo's outfits |
| `internal/journal` | Answers, doodles and mood check-ins, saved in `journal.json` |
| `internal/game` | State machine (offer → accept → complete) and save data |
| `internal/store` | Atomic JSON file persistence |
| `internal/desktop` | Window geometry; Win32 controller (`window_windows.go`) and dev fallback |
| `internal/tray` | Menu bar item (macOS, Objective-C) and notification-area icon (Windows, Win32) |
| `frontend/src` | Vue UI: `Capsule`, `QuestCard`, `QuestPlay`, `Celebration`, `Backpack`, `Onboarding`, `Mascot` |
| `frontend/src/lib/i18n.js` | UI text in English and Indonesian |

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

After finishing a quest you can tap 👍 or 👎. A liked quest comes up 1.5× as often, and a disliked one 0.2× as often:
rarer, but never gone for good. Ratings also nudge their whole category, by at most ±20–30%. That way you get more
of the kinds you enjoy without a few ratings taking over. Ratings stay on the device. Tap the same thumb again to
clear it.

New players only get easy quests. Medium quests unlock at level 3 and bold ones at level 5. Bold means a little
out of your comfort zone, not hard thinking. Your own quests are always available.

## When Boo speaks up

By default, Boo perks up now and then (every 20–45 minutes) with a gentle glow on the capsule. In Backpack →
Settings you can switch to a **break reminder** every 60 or 90 minutes instead. The timer restarts whenever you
open and close the panel, so a break you just took counts.

Boo never nudges while a fullscreen app, game, video or presentation is in front, or while the screen is locked. A
reminder that comes due then waits until the screen is free. On Windows this uses `SHQueryUserNotificationState`,
checked every 5 seconds. By default the capsule also steps aside until the fullscreen app is gone; Settings has a
toggle to keep it visible. It stays put while the panel is open. If you hid Boo yourself, it stays hidden. macOS
doesn't detect a busy screen yet (`internal/desktop/busy_other.go`).

### Quest schedule

In Backpack → Settings, pick up to four times a day: every day, weekdays or weekends. At each time the capsule lights
up with **"Quest time!"**, and Boo comes back if it was hidden. An optional system notification (off by default)
opens the panel when clicked. Each time fires once, even across restarts. If the screen is busy, it waits up to 15
minutes; a time that was missed (busy, asleep, app closed) is skipped, never piled up.

## Sound effects

Small sounds are synthesised with Web Audio, so there are no audio files and they work offline:
- a pop when you start a quest;
- ticks for checklists, counters and the twist toggle;
- a swoosh for reroll and skip;
- a calm tone with each breath in the breathing guide, and a ding when a timer ends;
- a chime when you finish a quest, or a short jingle when you level up.

Sounds are on by default at 60% and can be turned off or adjusted in Settings. The sound when Boo speaks up from the
capsule is the only one that interrupts, so it is opt-in. Everything is silent while a fullscreen app is in front.
Webviews only allow audio after the first click, so a greeting sound before you have touched Boo at all may stay
silent.

## Playing in the app

A quest's optional `interaction` says how it is played after **Let's go!**. Until that part is done, the
**I did it!** button stays off and a hint says what's missing.

| `kind` | Played as | Fields |
| --- | --- | --- |
| *(none)* | Do it, then tap Done | – |
| `answer` | 1–5 short text boxes | `prompts` |
| `checklist` | 2–8 steps to tick off | `prompts` |
| `timer` | Countdown; `"guide": "breath"` adds box breathing | `seconds` (5–300), `guide` |
| `counter` | Tap button, optionally up to a target | `target` (0–50) |
| `doodle` | Small drawing canvas | – |

What you write, count and draw goes to the **Journal** (Backpack → Journal), along with an optional mood check-in
after each quest. Once a kind of quest has three check-ins, the Journal gently says which kind tends to leave you
feeling best. It is an observation, not an assessment. The Journal keeps the last 120 entries, and drawings for the
30 most recent. It never leaves the device.

## Progress and outfits

XP fills the level bar. Each level unlocks something for Boo, which you can choose in Backpack → Boo:

| Level | Unlock |
| --- | --- |
| 2 | Mint hat |
| 3 | Round glasses, medium quests |
| 4 | Coral hat |
| 5 | Cozy scarf, bold quests |
| 6 | Golden hat |
| 7 | Little flower |
| 8 | Crown |

## Language

The first-run intro asks for Indonesian or English. You can change it later in Backpack → Settings, and the
tray or menu bar menu follows. Quests carry their translations in an `i18n` block:

```json
"i18n": { "id": { "title": "…", "description": "…", "prompts": ["…"], "twists": { "twist-id": "…" } } }
```

Missing translations fall back to the English text.

## Your own quests

Open the 🎒 Backpack and tap **+ Make a quest**. Give it a name, optional details, a kind, how long it takes and how
hard it is. XP is worked out from time and difficulty (10–40 XP), so making quests is never an XP shortcut.

Your quests are saved to `custom.json` in the data folder. They use the normal pack format, so they can be exported
later. They join the normal rotation and come up about twice as often as built-in quests. **Try it now**, or ▶ in the
list, offers one right away. Deleting needs two taps. You can have up to 100 quests.

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

## Tray and quitting

SideQuest keeps running in the background like a little companion:

- **macOS:** no Dock icon. Boo lives in the menu bar. Cmd+Q is not bound.
- **Windows:** no taskbar button. Boo lives in the system tray. Alt+F4 only hides the window.

The tray/menu bar has **Open SideQuest**, **Hide SideQuest** and **Quit SideQuest**. Quitting is only possible from there.
On Windows, a left-click on the tray icon opens the panel. "Hide SideQuest for now" in the Backpack tucks the window away.

Wails v2 has no tray API, so `internal/tray` is a small native implementation for each platform.
On macOS, Wails makes the app a regular Dock app at launch, so the tray switches it to an accessory app once the
window is ready. The Dock icon may show for a moment during launch.

## Windows notes

- On Windows 11 22H2 and later the window is truly transparent (`WindowIsTranslucent` with no backdrop, so
  `WS_EX_NOREDIRECTIONBITMAP`), and the shadows show. Older Windows can't do that: Wails falls back to a blur-behind,
  and without it the margin is painted dark. There the window region follows the rounded panel and the pill exactly
  (`CreateRoundRectRgn`), so no box is ever visible, at the cost of the shadow.
- The window is frameless with a transparent WebView2, so only the capsule or panel is visible. The DWM frame is disabled.
- `WS_EX_TOOLWINDOW` keeps SideQuest out of the taskbar and Alt+Tab.
- The window never resizes. It always has the panel's size, and a window region (`SetWindowRgn`) trims it down to
  the capsule. Clicks outside the capsule reach the desktop underneath. Opening the panel just removes the region,
  so WebView2 never has to re-layout mid-animation and the capsule can't jump.
- The capsule sits in the window corner nearest its screen corner, so the panel always opens toward the middle of
  the screen. If the capsule has been dragged to another part of the screen, the window is re-laid out on the next
  open.
- Positioning uses Win32 directly and is per-monitor DPI-aware. Wails' own position APIs mix work-area-relative and
  absolute coordinates.
- The capsule position is saved every 15 seconds, on collapse, and on close. It is clamped back on screen at startup
  if a monitor was removed.

Save data: `%AppData%\SideQuest\state.json`. It is written atomically. A corrupt file is moved aside and the app
starts fresh.
