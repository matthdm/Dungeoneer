---
plan-id: options-menu
status: queued
owner: unassigned
branch: plan/options-menu
depends-on: []
last-touched: 2026-04-30
---

# Plan: Options Menu

## Goal

Wire up the Options menu that is currently stubbed in `src/game/handlers.game.go`. The options screen lets the player toggle fullscreen, adjust master volume (placeholder slider until audio ships in Phase 9), and rebind a minimal set of keys. Settings persist to a `options.json` file alongside `meta.json`. The screen is accessible from the main menu and from a pause menu during a run.

## Scope

**In scope:**
- `OptionsData` struct with: `Fullscreen bool`, `MasterVolume float64` (0.0–1.0), `KeyBindings map[string]ebiten.Key` for the 4 most critical actions (Move Up/Down/Left/Right or equivalent).
- Save/load `options.json` to the same directory as `meta.json`.
- Options screen UI: toggle, slider (non-functional volume knob until audio ships), key-rebind rows.
- Apply fullscreen toggle immediately via `ebiten.SetFullscreen`.
- Accessible from main menu ("Options" button) and from an in-run pause menu (Escape key).

**Out of scope (do not change in this plan):**
- Actual audio volume wiring (that is Phase 9C).
- Full key rebinding for all actions (just the critical 4 to prove the pattern).
- Controller/gamepad support.
- Graphics quality settings.

## File envelope

**Touched:**
- `src/game/handlers.game.go` — replace Options stub with real transition
- `src/game/options.go` *(new)* — `OptionsData` struct, load/save, apply
- `src/ui/options_screen.go` *(new)* — options screen UI
- `src/game/game.go` — register new state, store `Options *OptionsData`
- `src/game/draw.game.go` — draw dispatch for options state
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — owned by offset plan
- `src/levels/`, `src/entities/`, `src/items/`, `src/dialogue/` — not relevant
- `src/audio/` — does not exist yet; do not create it here

## Acceptance criteria

- [ ] "Options" from main menu opens the options screen.
- [ ] Fullscreen toggle works immediately and is saved to `options.json`.
- [ ] Master volume slider renders and persists its value (actual audio hookup deferred to Phase 9C).
- [ ] Key rebinding UI shows current binding for 4 actions; clicking a row enters "press a key" mode and updates the binding on keypress.
- [ ] Settings are loaded on game startup and applied (fullscreen state restored).
- [ ] Pressing Escape during a run opens the options screen (pause behavior: game update loop halted while options are open).
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: OptionsData and persistence
- [ ] 1.1 Create `src/game/options.go` with `OptionsData` struct: `Fullscreen bool`, `MasterVolume float64`, `KeyBindings map[string]ebiten.Key`.
- [ ] 1.2 `DefaultOptions() *OptionsData` returns sensible defaults (windowed, volume 0.8, WASD).
- [ ] 1.3 `LoadOptions() (*OptionsData, error)` reads `options.json`; returns defaults on missing or corrupt file.
- [ ] 1.4 `(o *OptionsData) Save() error` writes to `options.json`.
- [ ] 1.5 `(o *OptionsData) Apply()` calls `ebiten.SetFullscreen(o.Fullscreen)`.
- [ ] 1.6 In `src/game/game.go` `New()`, call `LoadOptions()` and store on `Game`, call `Apply()`.
- [ ] 1.7 `cd src && go build ./...` passes.

### Phase 2: Options screen UI
- [ ] 2.1 Create `src/ui/options_screen.go` with `OptionsScreen` struct. Fields: `Options *OptionsData`, `OnBack func()`.
- [ ] 2.2 Draw rows: Fullscreen (toggle checkbox), Master Volume (horizontal slider bar, no audio wiring), then one row per key binding.
- [ ] 2.3 Key rebinding: clicking a row sets `rebindingAction string`; next key event updates `Options.KeyBindings[rebindingAction]` and clears rebinding state.
- [ ] 2.4 Back button calls `Options.Save()` then `OnBack()`.
- [ ] 2.5 `cd src && go build ./...` passes.

### Phase 3: State wiring
- [ ] 3.1 In `src/game/handlers.game.go`, replace Options stub with: construct `OptionsScreen`, set `OnBack` to return to previous state, transition.
- [ ] 3.2 Add Escape handler in the in-run update path to push options screen onto state (pause). On back, pop and resume run.
- [ ] 3.3 Add draw and update dispatch in `src/game/draw.game.go` and `src/game/game.go`.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Move this plan to `plans/COMPLETED/options-menu.md`.
- [ ] 4.2 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **State stack vs. flat state machine.** Pushing options over the in-run state requires either a state stack or a `previousState` field. Recommendation: add a `previousState GameState` field to `Game` for now; a full stack is overkill until there are 3+ overlay states. Affects 3.2.
- **Save file path.** Wherever `meta.json` is saved, put `options.json` beside it. Confirm the path from `src/game/metasave.go` before hardcoding. Affects 1.3–1.4.
