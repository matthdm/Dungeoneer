---
plan-id: runstate-serialization
status: queued
owner: unassigned
branch: plan/runstate-serialization
depends-on: [load-game-menu]
last-touched: 2026-04-30
---

# Plan: RunState Serialization

## Goal

Implement full mid-run save/load so the player can quit during a run and resume from the same floor, same position, same inventory. Currently only `MetaSave` (Remnants, run count, best floor) persists across sessions; `RunState`, the current floor layout, entity positions, and inventory are all lost on exit. This plan makes the game fully resumable.

When done: quitting the game mid-run and relaunching resumes on the same floor with the same enemies, items, and quest flags intact. Death or victory still clears the run save normally.

## Scope

**In scope:**
- Serialize and deserialize `RunState` (quest flags, floor number, kill count, Remnants-in-run).
- Serialize and deserialize `FloorContext` (current floor layout seed, room list, entity spawn list — enough to regenerate the same floor deterministically, or store full entity state).
- Serialize player inventory and equipped items.
- Serialize player position and HP/mana.
- Auto-save on exit (graceful quit and alt-F4) and on floor transition.
- Auto-load on "Continue" (hooking into `load-game-menu` plan's Continue button).
- Clear run save on death or victory.

**Out of scope (do not change in this plan):**
- Multiple save slots.
- Cloud saves.
- Saving NPC state beyond what's already in `RunState.QuestFlags` (NPC positions are re-derived from room tags on floor gen; that's sufficient for now).
- MetaSave changes (handled in `6A-full-meta-save.md`).

## File envelope

**Touched:**
- `src/game/runsave.go` *(new)* — `RunSave` struct, marshal/unmarshal, file path helper, `ClearRunSave()`
- `src/game/runstate.go` — add JSON tags if missing; no structural changes
- `src/entities/player.go` — add `ToSnapshot()` / `FromSnapshot()` for position + HP + mana + inventory
- `src/inventory/inventory.go` — add `ToSnapshot()` / `FromSnapshot()`
- `src/game/hub.go` — call `SaveRun()` on floor transition; call `ClearRunSave()` on death/victory
- `src/game/metasave.go` — call `ClearRunSave()` on death (it already handles MetaSave update)
- `src/game/game.go` — hook `SaveRun()` to `Layout()` / window close signal
- `src/ui/load_screen.go` — detect run save in addition to meta save; show "resume floor N" if present (coordinate with `load-game-menu` plan)
- `plans/_QUEUE.md` — move to Completed on finish
- `design-docs/roadmap.md` — no specific row, but note in CLAUDE.md

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — owned by offset plan
- `src/levels/generate64.go` — floor generation itself; we regenerate from seed, not patch it
- `src/dialogue/`, `src/dialogues/` — not relevant
- `src/entities/monster.go`, `src/entities/npc.go` — entity positions are re-derived from floor gen seed

## Acceptance criteria

- [ ] Quitting mid-run and relaunching shows "Resume Run (Floor N)" on the load screen.
- [ ] Resuming restores: current floor number, player position, player HP/mana, full inventory (equipped and carried), `RunState.QuestFlags`, kill count, Remnants-in-run.
- [ ] The floor layout is regenerated from the saved seed — rooms, corridors, and chest positions match the original.
- [ ] Enemies are re-spawned from the saved encounter list — not re-randomized. (Enemies killed before save do not respawn.)
- [ ] Dying after loading a run save clears the run save and behaves normally (death screen, Remnants to meta).
- [ ] Completing a floor after loading clears the run save for that floor and writes a new one for the next.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: RunSave data model
- [ ] 1.1 Create `src/game/runsave.go`. Define `RunSave` struct with: `RunState`, `FloorSeed int64`, `PlayerSnapshot PlayerSnapshot`, `KilledEntityIDs []string` (to suppress respawn).
- [ ] 1.2 Define `PlayerSnapshot` struct: `TileX, TileY int`, `InterpX, InterpY float64`, `HP, MaxHP, Mana, MaxMana int`, `InventorySnapshot`, `EquippedSnapshot`.
- [ ] 1.3 `SaveRunSave(rs *RunSave) error` — marshal to `runsave.json` beside `meta.json`.
- [ ] 1.4 `LoadRunSave() (*RunSave, error)` — unmarshal; return `nil, nil` if file absent.
- [ ] 1.5 `ClearRunSave() error` — delete `runsave.json`.
- [ ] 1.6 `cd src && go build ./...` passes.

### Phase 2: Player and inventory snapshots
- [ ] 2.1 In `src/entities/player.go`, add `ToSnapshot() PlayerSnapshot` capturing position, HP/mana, inventory state.
- [ ] 2.2 Add `ApplySnapshot(PlayerSnapshot)` to restore the above fields from a snapshot.
- [ ] 2.3 In `src/inventory/inventory.go`, add `ToSnapshot() InventorySnapshot` (slice of item IDs + quantities + slot positions) and `FromSnapshot(InventorySnapshot)`.
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Save triggers
- [ ] 3.1 In `src/game/hub.go` floor-transition path, call `SaveRunSave()` after generating the new floor and before the player enters.
- [ ] 3.2 In `src/game/metasave.go` death/victory path, call `ClearRunSave()` after writing MetaSave.
- [ ] 3.3 In `src/game/game.go`, hook `ebiten`'s window-close signal (or `Layout()` if no close hook exists) to call `SaveRunSave()` if a run is active.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Load and restore
- [ ] 4.1 Update `src/ui/load_screen.go` (from `load-game-menu` plan): if `LoadRunSave()` returns a non-nil save, show "Resume Run — Floor N" as an additional option above "New Run".
- [ ] 4.2 On resume, call `LoadRunSave()`, regenerate the floor from `FloorSeed`, suppress entities in `KilledEntityIDs`, then call `player.ApplySnapshot()`.
- [ ] 4.3 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Note in `CLAUDE.md` that RunState is now fully serialized.
- [ ] 5.2 Move this plan to `plans/COMPLETED/runstate-serialization.md`.
- [ ] 5.3 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Regenerate from seed vs. store full entity list.** Regenerating from seed is simpler and handles the room layout, but enemies that moved significantly before save will re-spawn at their original positions. Recommendation: store `KilledEntityIDs` only (dead enemies don't respawn); living enemies re-spawn in place from the encounter list. Acceptable for the first pass. Affects 1.1 and 4.2.
- **Window-close signal in Ebiten.** Ebiten v2.8 does not expose an `OnClose` callback; the closest is checking `inpututil.IsKeyJustPressed(ebiten.KeyEscape)` in the update loop or hooking the OS signal in `main.go`. Recommendation: save on every floor transition (which covers most cases) and accept that alt-F4 mid-floor loses the in-floor delta. Document in "What was NOT changed." Affects 3.3.
