---
plan-id: 7A-echoes-of-self
status: complete
owner: claude-sonnet-4-6
branch: plan/7A-echoes-of-self
depends-on: [6C-lore-system]
last-touched: 2026-04-30
---

# Plan: Phase 7A — Echoes of Self

## Goal

Record each run's player path, equipment, and death location, then manifest those recordings as echo entities in future runs. Echoes come in three forms: Wicked Echo (enemy, drops Remnants), Hero Echo (temporary ally), and Memory Fragment (static ghost NPC with lore dialogue). All share a ghost visual style. An Echo Shrine in the hub lets the player view and manage stored echoes.

When done, the dungeon is haunted by the player's past selves in a way that creates both tactical variety and emotional resonance.

## Scope

**In scope:**
- Echo recorder capturing position path, HP history, equipped items, and cause of death each run.
- `EchoRecord` data model serialized to `echoes/run_N.json`.
- Echo spawner: places up to 3 echoes per floor near historical death locations.
- Three echo entity types: Wicked (enemy), Hero (ally), Memory Fragment (NPC ghost).
- Ghost visual style: blue/purple tint, reduced opacity, flicker.
- Echo Shrine hub UI for viewing and banishing echoes.

**Out of scope (do not change in this plan):**
- Living Dungeon AI (parallel plan `7B-living-dungeon-ai.md`).
- Upgrade that increases echo count (Phase 8B).

## File envelope

**Touched:**
- `src/game/echo_recorder.go` *(new)* — tick-by-tick recorder, run-end serializer
- `src/entities/echo.go` *(new)* — `WickedEcho`, `HeroEcho`, `MemoryFragment` entity types
- `src/game/metasave.go` — add echo file path list to MetaSave
- `src/game/hub.go` — echo spawner pass; echo shrine spawn when unlocked
- `src/game/game.go` — start/stop recorder; store active echoes
- `src/game/draw.game.go` — ghost shader pass for echo entities (tint + alpha)
- `src/ui/echo_shrine.go` *(new)* — echo shrine UI
- `design-docs/roadmap.md` — mark 7.1–7.8 ✅ on completion
- `CLAUDE.md` — update Phase 7 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/game/dungeon_ai.go` — living dungeon plan scope
- `src/dialogue/` — Memory Fragment dialogue is simple; use existing dialogue engine

## Acceptance criteria

- [x] After every run (death or victory), an `EchoRecord` is written to `echoes/run_N.json` capturing: floor-by-floor position snapshots (sampled every 2 seconds), equipped item IDs at death, cause of death string, floor of death.
- [x] On floor generation, the echo spawner reads stored `EchoRecord` files and spawns 0–3 echo entities near recorded death locations on matching floors.
- [x] `WickedEcho` behaves as a melee enemy with the original run's stats; drops 5–15 Remnants on defeat.
- [x] `HeroEcho` follows the player and attacks nearby enemies for 60 seconds, then dissolves.
- [x] `MemoryFragment` is a static ghost NPC; interaction shows a 1–2 line contextual message ("I fell here. Watch the shadows.").
- [x] All echo entities render with a blue/purple tint and reduced opacity (alpha 0.6, flicker ±0.1 at 2Hz).
- [x] Echo Shrine in hub (when unlocked) lists stored echo runs with date, floor of death, and cause; player can banish (delete) individual echo records.
- [x] `cd src && go build ./...` passes.

## Phases

### Phase 1: Echo recorder and data model
- [x] 1.1 Create `src/game/echo_recorder.go`. `EchoRecorder` struct with `Snapshots []PositionSnapshot`, `EquippedItems []string`, `DeathCause string`, `DeathFloor int`.
- [x] 1.2 `PositionSnapshot`: `Floor int`, `X, Y float64`, `HP int`, `Timestamp float64`.
- [x] 1.3 `Record(dt float64)` called each update tick; appends a snapshot every 2 seconds.
- [x] 1.4 `Finalize(cause string, floor int)` called on run end; serializes to `echoes/run_N.json` where N is `MetaSave.CompletedRuns`.
- [x] 1.5 In `src/game/metasave.go`, add `EchoFiles []string` to track echo record paths.
- [x] 1.6 `cd src && go build ./...` passes.

### Phase 2: Echo entity types
- [x] 2.1 Create `src/entities/echo.go`. Define `EchoType` enum and base `Echo` struct (position, type, source record reference).
- [x] 2.2 `WickedEcho`: implements monster interface, melee behavior, HP scaled from source run's floor, drops Remnants on death.
- [x] 2.3 `HeroEcho`: follows player, attacks nearest enemy in range, 60-second lifetime, no drops.
- [x] 2.4 `MemoryFragment`: static, interaction triggers a short dialogue panel with a 1–2 line contextual message drawn from the death cause string.
- [x] 2.5 `cd src && go build ./...` passes.

### Phase 3: Ghost visual style
- [x] 3.1 In `src/game/draw.game.go`, detect echo entities in the draw list. Apply `ColorScale` to set RGB to blue/purple tint (R: 0.4, G: 0.5, B: 1.0) and alpha to 0.6.
- [x] 3.2 Flicker: use a sine wave on game tick to oscillate alpha between 0.5 and 0.7 at 2Hz.
- [x] 3.3 `cd src && go build ./...` passes.

### Phase 4: Echo spawner and shrine
- [x] 4.1 In `src/game/hub.go` floor generation, after room layout: load echo records, for each record where `DeathFloor == currentFloor`, pick 1 echo type (weighted: 50% Wicked, 30% Hero, 20% Memory), spawn near the death position. Cap at 3 echoes per floor.
- [x] 4.2 In `src/game/hub.go` hub load, spawn Echo Shrine entity when `HubState["echo_shrine_unlocked"]`.
- [x] 4.3 Create `src/ui/echo_shrine.go`. List stored echo runs (date, floor, cause). Button to banish (delete file and remove from `MetaSave.EchoFiles`). Close button.
- [x] 4.4 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark roadmap rows 7.1–7.8 ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 7 status block.
- [ ] 5.3 Move this plan to `plans/COMPLETED/7A-echoes-of-self.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |
| 2026-07-05 | 1–4 | Complete | All files found already implemented. Added Remnant drop handling in handleMonsterDeath (progression.game.go): WickedEcho awards 5–15 Remnants on death, skips normal XP/gold/loot. Build passes, game tests pass. |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Snapshot frequency vs. file size.** 2-second sampling on a 10-minute run is ~300 snapshots per floor × 7 floors = ~2100 records. At ~50 bytes each, that's ~100KB per run — acceptable. If echo files grow unwieldy, cap at 50 runs. Affects 1.3.
- **Echo floor matching.** Death floor matching is exact (floor 3 echo appears only on floor 3). If the player dies on floor 1 repeatedly, floor 1 fills with echoes. Recommendation: cap 3 echoes per floor regardless of available records; randomly sample from available records for that floor. Affects 4.1.
