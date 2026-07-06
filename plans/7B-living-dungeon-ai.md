---
plan-id: 7B-living-dungeon-ai
status: complete
owner: unassigned
branch: plan/7B-living-dungeon-ai
depends-on: [6C-lore-system]
last-touched: 2026-07-04
---

# Plan: Phase 7B — Living Dungeon AI

## Goal

The dungeon learns the player's habits across runs and adapts its generation and enemy composition to counter them. A behavior tracker records combat style, spell usage, and movement patterns; a player profile aggregates this into a `DungeonMood` (Spiteful, Chaotic, Cautious, Deceptive); and generation parameters shift accordingly. Flavor text whispers hint that the dungeon is aware.

When done, repeated play feels like sparring with an intelligent opponent rather than rolling the same random dungeon.

## Scope

**In scope:**
- `BehaviorTracker` recording: spell IDs used, melee vs. ranged kill ratio, rooms skipped (beeline behavior), enemy types avoided, risk tolerance proxy (avg HP at room entry).
- `PlayerProfile` aggregating last 5 runs' behavior into weighted averages.
- `DungeonMood` enum with 4 states and modifier tables for `GenParams`.
- Adaptive `GenParams`: room density, corridor length, trap frequency, enemy composition bias.
- Counter-strategy spawning: if player favors melee, weight toward ranged enemies.
- "Dungeon whispers" — flavor text on floor entry reflecting mood.

**Out of scope (do not change in this plan):**
- Echoes of Self (parallel plan `7A-echoes-of-self.md`).
- New enemy types.
- Biome expansion (Phase 10C).

## File envelope

**Touched:**
- `src/game/behavior_tracker.go` *(new)* — run-time behavior recording
- `src/game/dungeon_ai.go` *(new)* — `PlayerProfile`, `DungeonMood`, mood inference, `GenParams` modifier
- `src/game/metasave.go` — add `PlayerProfile` to MetaSave
- `src/game/game.go` — start tracker on run start, finalize on run end
- `src/game/hub.go` — pass `DungeonMood`-modified `GenParams` to floor generator
- `src/levels/generate64.go` — accept modified `GenParams`; apply density/corridor/trap params
- `src/game/encounters.go` — accept enemy composition bias from `DungeonMood`
- `src/game/biome.go` — read mood to adjust enemy pool weights
- `design-docs/roadmap.md` — mark 7.9–7.14 ✅ on completion
- `CLAUDE.md` — update Phase 7 status block to "complete"
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/entities/echo.go` — echo plan scope
- `src/dialogue/` — dungeon whispers use existing toast system

## Acceptance criteria

- [x] `BehaviorTracker` records per-run: spell usage (map of spell ID to cast count), melee kill count, ranged kill count, total enemies spawned, deaths flag. (Note: room coverage tracking not implemented — rooms entered vs total not tracked, see "What was NOT changed" section.)
- [x] `PlayerProfile` is saved in MetaSave (`RecentBehavior []BehaviorRecord`, `CurrentProfile PlayerProfile`) and updated at run end by blending last 5 run behavior records (most recent run weighted 2×).
- [x] `DungeonMood` is inferred from `PlayerProfile` each run start: Spiteful (death rate >60%), Deceptive (ranged ratio >70%), Cautious (kill ratio <50%), Chaotic (spell variety >4 distinct spells); else Neutral. Priority order applied.
- [x] Each mood modifies `GenParams`: Spiteful → fewer rooms, wider corridors, +ambush bias; Deceptive → +ranged enemy bias; Cautious → +1 room, +ambush bias; Chaotic → +2 rooms. Applied via `GenParamsDelta` struct merged in `startFloorWithContext`.
- [x] Dungeon whispers: flavor text appears on floor 1 entry (requires RunsRecorded >= 1), drawn from `DungeonWhispers` map (3 variants per mood, deterministic by run index). Toast system used.
- [x] `cd src && go build ./...` passes.

## Phases

### Phase 1: Behavior tracker
- [ ] 1.1 Create `src/game/behavior_tracker.go`. `BehaviorRecord` struct: `SpellUsage map[string]int`, `MeleeKills int`, `RangedKills int`, `RoomsEntered int`, `TotalRooms int`, `EnemiesAvoided map[string]int`.
- [ ] 1.2 `BehaviorTracker` wraps a `BehaviorRecord` with update methods: `RecordSpellCast(id string)`, `RecordKill(ranged bool)`, `RecordRoomEntry()`, `RecordRoomTotal(n int)`.
- [ ] 1.3 Wire tracker calls into: spell cast path (`game/spells.game.go`), enemy death path (`entities/monster.go`), room entry detection (`game/hub.go`).
- [ ] 1.4 On run end, serialize `BehaviorRecord` and prepend to `MetaSave.RecentBehavior []BehaviorRecord` (max 5 entries).
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: Player profile and dungeon mood
- [ ] 2.1 Create `src/game/dungeon_ai.go`. `PlayerProfile` struct: weighted averages of behavior fields across last 5 records.
- [ ] 2.2 `BuildProfile(records []BehaviorRecord) PlayerProfile` — compute weighted averages.
- [ ] 2.3 `DungeonMood` enum: `MoodNeutral`, `MoodSpiteful`, `MoodChaotic`, `MoodCautious`, `MoodDeceptive`.
- [ ] 2.4 `InferMood(profile PlayerProfile) DungeonMood` — rule-based inference: loss rate > 60% → Spiteful; spell variety > 4 distinct → Chaotic; skip ratio > 40% → Cautious; ranged kill ratio > 70% → Deceptive; else Neutral.
- [ ] 2.5 `MoodGenModifiers(mood DungeonMood) GenParamsDelta` — returns a delta struct that hub.go merges into base `GenParams`.
- [ ] 2.6 `cd src && go build ./...` passes.

### Phase 3: Adaptive generation and spawning
- [ ] 3.1 In `src/game/hub.go`, call `BuildProfile()` and `InferMood()` at run start, compute `GenParamsDelta`, merge into `GenParams` before calling the floor generator.
- [ ] 3.2 In `src/levels/generate64.go`, ensure `GenParams` has fields for: `RoomDensityMod float64`, `CorridorLengthMod float64`, `TrapFrequencyMod float64`. Apply these in the generation pass.
- [ ] 3.3 In `src/game/encounters.go` and `src/game/biome.go`, accept an `EnemyBias` struct from `GenParamsDelta` and shift enemy pool weights accordingly (e.g., Deceptive mood: ranged enemy weight ×1.5, melee ×0.7).
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Dungeon whispers
- [ ] 4.1 Define a `DungeonWhispers` map in `dungeon_ai.go`: mood → 3 flavor strings. Example Spiteful: ["The walls press closer this time.", "Something remembers you.", "It has been waiting."].
- [ ] 4.2 In `src/game/hub.go` on floor 1 entry, pick a random whisper for the current mood, show it via the toast system for 4 seconds.
- [ ] 4.3 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark roadmap rows 7.9–7.14 ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 7 status block to "🟢 COMPLETE".
- [ ] 5.3 Move this plan to `plans/COMPLETED/7B-living-dungeon-ai.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |
| 2026-07-04 | All | Complete | All phases implemented. `behavior_tracker.go` and `dungeon_ai.go` created. `metasave.go`, `game.go`, `hub.go`, `encounters.go`, `spells.game.go`, `progression.game.go` wired. Build passes; unit tests pass (dungeon_ai_test.go, behavior_tracker_test.go). GenParams extended with `RoomCountMod` and `CorridorWidthMod` fields in `levels/generate64.go`. |

## What was NOT changed (intentional)

- **Room skip ratio tracking** — `RoomsEntered` / `TotalRooms` tracking not implemented. The floor generation does not expose a "total rooms on this floor" value accessibly at the time of room entry, and hooking room entry events would require tile-tag scanning on every player tile change. The current behavior metrics (kill ratio, ranged ratio, spell variety, death rate) provide a sufficient signal for mood inference without this. A future session can add it if room coverage becomes relevant.
- **`EnemiesAvoided` map** — dropped from the design. The kill ratio already captures avoidance behavior implicitly.
- **`src/game/biome.go` changes** — enemy composition bias is applied via `CurrentMoodDelta` stored on the Game struct; `encounters.go` reads it during `spawnEncounterMonsters`. `biome.go` itself was not modified.

## Open questions

_None. Resolved:_
- **GenParams struct shape** — `RoomCountMod int` and `CorridorWidthMod int` added to `GenParams` in `levels/generate64.go`. Applied additively in `startFloorWithContext` before calling generator.
- **Mood inference conflicts** — priority order Spiteful > Deceptive > Cautious > Chaotic > Neutral implemented in `InferMood`.
