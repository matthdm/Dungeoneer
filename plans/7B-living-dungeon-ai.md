---
plan-id: 7B-living-dungeon-ai
status: queued
owner: unassigned
branch: plan/7B-living-dungeon-ai
depends-on: [6C-lore-system]
last-touched: 2026-04-30
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

- [ ] `BehaviorTracker` records per-run: spell usage (map of spell ID to cast count), melee kill count, ranged kill count, rooms entered vs. rooms on floor (skip ratio), distinct enemy types avoided.
- [ ] `PlayerProfile` is saved in MetaSave and updated at run end by blending last 5 run behavior records (exponential decay — most recent run weighted 2×).
- [ ] `DungeonMood` is inferred from `PlayerProfile` each run start: Spiteful (player loses often), Chaotic (player uses many different spells), Cautious (player skips many rooms), Deceptive (player avoids melee, uses ranged heavily).
- [ ] Each mood modifies `GenParams`: Spiteful → more dead ends, tighter corridors; Chaotic → more open rooms, more enemies; Cautious → more ambush placements; Deceptive → more ranged enemies, fewer melee.
- [ ] Dungeon whispers: a 4-second flavor text appears on floor 1 entry, text drawn from a mood-keyed table (3 variants per mood).
- [ ] `cd src && go build ./...` passes.

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

## What was NOT changed (intentional)

_None yet._

## Open questions

- **GenParams struct shape.** Confirm what fields `GenParams` already has in `levels/generate64.go` before adding delta fields. The delta approach (merge, don't replace) avoids breaking baseline generation. Affects 3.2.
- **Mood inference conflicts.** A player can simultaneously qualify for Chaotic and Deceptive. Recommendation: priority order — Spiteful > Deceptive > Cautious > Chaotic > Neutral. The highest-priority qualifying mood wins. Affects 2.4.
