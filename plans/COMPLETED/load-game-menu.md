---
plan-id: load-game-menu
status: queued
owner: unassigned
branch: plan/load-game-menu
depends-on: []
last-touched: 2026-04-30
---

# Plan: Load Game Menu

## Goal

Wire up the Load Game menu that is currently stubbed in `src/game/handlers.game.go`. When the player selects "Load Game" from the main menu, the game reads `meta.json` and any existing run state, presents a minimal confirmation screen, and resumes the run (or shows "No saved game found" and returns to the menu). This is a self-contained UX fix that requires no new gameplay systems.

## Scope

**In scope:**
- Detecting whether a valid saved game exists (`meta.json` present and readable).
- A `LoadGameState` game state (or reuse of an existing menu state) that displays: run summary (floor reached, Remnants), a "Continue" button, and a "Back" button.
- Loading meta state on confirm and transitioning to the hub or active floor.
- Error handling: corrupted or missing save shows "No saved game found" and returns to main menu.

**Out of scope (do not change in this plan):**
- Full RunState serialization (that is `runstate-serialization.md`). This plan only handles loading what already persists (Remnants, run count, best floor via `meta.json`).
- Options menu (separate plan `options-menu.md`).
- New save slots or cloud saves.
- Any gameplay system changes.

## File envelope

**Touched:**
- `src/game/handlers.game.go` — replace Load Game stub with real state transition
- `src/game/metasave.go` — add `Exists()` helper and load-with-error-return variant
- `src/ui/load_screen.go` *(new)* — load game summary UI panel
- `src/game/game.go` — register new `LoadGameState` if needed
- `src/game/draw.game.go` — draw dispatch for new state
- `plans/_QUEUE.md` — move to Completed on finish
- `CLAUDE.md` — no phase status change needed (this is housekeeping)
- `design-docs/roadmap.md` — no entry to check off (this is a TODO fix, not a roadmap row)

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — owned by offset plan
- `src/levels/`, `src/entities/`, `src/items/` — not relevant
- `src/dialogue/`, `src/dialogues/` — not relevant
- `src/game/runstate.go` — do not add full serialization here; that is a separate plan

## Acceptance criteria

- [ ] Selecting "Load Game" from the main menu when `meta.json` exists transitions to a load screen showing: floors cleared, Remnants carried, run count.
- [ ] Pressing "Continue" on the load screen loads MetaSave and enters the hub.
- [ ] Selecting "Load Game" when no valid `meta.json` exists shows "No saved game found" and returns to main menu after 2 seconds or a keypress.
- [ ] Corrupted `meta.json` (invalid JSON) is treated the same as missing: graceful error, return to menu.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: MetaSave existence check
- [ ] 1.1 Add `MetaSaveExists() bool` to `src/game/metasave.go` — returns true if the save file is present and valid JSON.
- [ ] 1.2 Add `LoadMetaSaveWithError() (*MetaSave, error)` variant (or update existing loader to return error) so the caller can handle corruption gracefully.
- [ ] 1.3 `cd src && go build ./...` passes.

### Phase 2: Load screen UI
- [ ] 2.1 Create `src/ui/load_screen.go` with a `LoadScreen` struct. Fields: `MetaSave *MetaSave`, `OnContinue func()`, `OnBack func()`.
- [ ] 2.2 `Draw(*ebiten.Image)` renders a centered panel with run summary and two buttons (Continue / Back), matching the existing UI style from `src/ui/death_screen.go` as a reference.
- [ ] 2.3 `Update()` handles mouse click and keyboard (Enter = continue, Escape = back).
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: State wiring
- [ ] 3.1 In `src/game/handlers.game.go`, replace the Load Game stub: call `MetaSaveExists()`; if true, construct `LoadScreen` and transition to it; if false, show the "No saved game" message (can be a toast or a simple timed text overlay — match whatever `ui/toast.go` or equivalent already provides).
- [ ] 3.2 In `src/game/game.go` / `src/game/draw.game.go`, add draw and update dispatch for the load screen state.
- [ ] 3.3 On "Continue" confirmation, call `LoadMetaSaveWithError()`, apply to `g.Meta`, and transition to hub state.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Move this plan to `plans/COMPLETED/load-game-menu.md`.
- [ ] 4.2 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Toast vs. overlay for "no save found."** If `ui/toast.go` doesn't exist yet, a simple 2-second timed text drawn in the main menu state is sufficient. Don't create a toast system just for this. Affects 3.1.
- **Hub vs. active floor on load.** Current MetaSave only persists between-run state (Remnants, run count). Always loading into the hub is correct until `runstate-serialization.md` ships. Affects 3.3.
