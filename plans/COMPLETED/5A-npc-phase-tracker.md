---
plan-id: 5A-npc-phase-tracker
status: queued
owner: unassigned
branch: plan/5A-npc-phase-tracker
depends-on: [4F-loot-refinement]
last-touched: 2026-04-30
---

# Plan: Phase 5A — NPC Phase Tracker & Trust System

## Goal

Build the generalized NPC phase progression infrastructure that all major NPC arcs (Varn, Seris, Mira, Kael) will use. Right now Varn's phase is tracked ad-hoc through `QuestFlags`; this plan introduces `NPCPhaseTracker`, a `SelectTree()` upgrade that reads phase and defeat count, and the `add_trust` / trust-threshold dialogue system.

When done, the dialogue engine can select the correct dialogue tree for any major NPC based on their phase and trust level, and a new NPC arc requires only: a data entry in `npc_data.go`, a set of JSON dialogue trees, and phase-advancement conditions — no new Go code per NPC.

## Scope

**In scope:**
- `NPCPhaseTracker` struct in a new `game/npc_phases.go`: holds phase index, advancement conditions (quest flag threshold + floor threshold), and exposes `ShouldAdvance(RunState) bool` + `Advance()`.
- Hook `NPCPhaseTracker` into the per-floor NPC update so phases advance automatically when conditions are met.
- Update `dialogue/loader.go` `SelectTree()` to consult `npc_phase` flag and `NPCMeta.DefeatCount`.
- `add_trust` dialogue action (integer delta), `trust_gte` condition for gating responses.
- `trust_decay` on betrayal action.
- Store trust per NPC in `NPCMeta` (already in MetaSave).

**Out of scope (do not change in this plan):**
- Varn's actual dialogue content (that is `5B-varn-arc.md`).
- Boss selection logic (that is `5C-boss-selection.md`).
- Hub NPC spawning positions (that is `5D-hub-npc-quarter.md`).
- Other major NPC definitions (Seris, Mira, Kael).

## File envelope

**Touched:**
- `src/game/npc_phases.go` *(new)* — `NPCPhaseTracker`, advancement logic
- `src/entities/npc.go` — add `PhaseTracker *NPCPhaseTracker` field to major NPC type
- `src/game/runstate.go` — ensure `NPCMeta` has `Trust int` field; add if missing
- `src/game/npc.game.go` — call `PhaseTracker.ShouldAdvance()` on floor entry; execute `add_trust` / `trust_decay` actions
- `src/dialogue/loader.go` — update `SelectTree()` to use phase + defeat count
- `src/dialogue/types.go` — add `add_trust`, `trust_decay` actions; `trust_gte`, `trust_lte` conditions
- `src/game/npc_phases_test.go` *(new)* — unit tests for advancement logic and trust system
- `plans/_QUEUE.md` — move to Completed on finish
- `design-docs/roadmap.md` — mark 5.1, 5.2, 5.3 ✅ on completion
- `CLAUDE.md` — update Phase 5 status block

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — owned by offset plan
- `src/dialogues/*.json` — content belongs to `5B-varn-arc.md`
- `src/entities/varn_boss.go`, `src/entities/boss.go` — boss plan scope
- `src/levels/` — not relevant

## Acceptance criteria

- [ ] `NPCPhaseTracker` exists with `ShouldAdvance(rs RunState) bool` and `Advance()` methods.
- [ ] `ShouldAdvance` checks configurable quest flag threshold AND optional floor threshold; both must be met to advance.
- [ ] Phase tracker is evaluated on each floor entry for all major NPCs; phase increments when conditions are met.
- [ ] `dialogue/loader.go` `SelectTree()` selects the tree whose `requires_phase` matches the NPC's current phase, falling back to the highest available phase tree if no exact match.
- [ ] `add_trust N` action increments `NPCMeta[id].Trust` by N (capped at configurable max, default 100).
- [ ] `trust_decay N` action decrements trust by N (floored at 0).
- [ ] `trust_gte N` condition returns true when `NPCMeta[id].Trust >= N`.
- [ ] `trust_lte N` condition returns true when `NPCMeta[id].Trust <= N`.
- [ ] Unit tests in `src/game/npc_phases_test.go` cover: advance when conditions met, no advance when quest flag below threshold, no advance when floor below threshold, trust cap, trust floor.
- [ ] `cd src && go build ./...` passes.
- [ ] `cd src && go test ./...` passes.

## Phases

### Phase 1: NPCPhaseTracker
- [ ] 1.1 Create `src/game/npc_phases.go`. Define `PhaseCondition` struct: `RequiredFlag string`, `RequiredFlagValue int`, `RequiredFloor int`.
- [ ] 1.2 Define `NPCPhaseTracker` struct: `NPCID string`, `CurrentPhase int`, `MaxPhase int`, `Conditions []PhaseCondition` (one per phase transition).
- [ ] 1.3 Implement `ShouldAdvance(rs RunState) bool`: checks `rs.QuestFlags[RequiredFlag] >= RequiredFlagValue && rs.Floor >= RequiredFloor` for the condition at `CurrentPhase`.
- [ ] 1.4 Implement `Advance()`: increments `CurrentPhase` up to `MaxPhase`.
- [ ] 1.5 Write unit tests in `src/game/npc_phases_test.go` covering advance conditions.
- [ ] 1.6 `cd src && go build ./... && go test ./...` passes.

### Phase 2: Trust system
- [ ] 2.1 In `src/game/runstate.go`, confirm `NPCMeta` struct has `Trust int` field; add if absent.
- [ ] 2.2 In `src/dialogue/types.go`, add action types `"add_trust"` (value int) and `"trust_decay"` (value int); add condition types `"trust_gte"` and `"trust_lte"` (value int).
- [ ] 2.3 In `src/game/npc.game.go`, handle `add_trust` and `trust_decay` in the action executor: read NPC ID from context, update `RunState.NPCMeta[id].Trust`, clamp to [0, 100].
- [ ] 2.4 In dialogue condition evaluator, handle `trust_gte` and `trust_lte` by reading `RunState.NPCMeta[id].Trust`.
- [ ] 2.5 Add trust tests to `src/game/npc_phases_test.go`.
- [ ] 2.6 `cd src && go build ./... && go test ./...` passes.

### Phase 3: SelectTree upgrade and phase hook
- [ ] 3.1 In `src/dialogue/loader.go`, update `SelectTree()`: iterate trees for the NPC, find the one whose `"requires_phase"` metadata field (add this to `DialogueTree` if not present) equals the NPC's current phase; fallback to the highest-phase tree available.
- [ ] 3.2 Add `RequiresPhase int` field to `DialogueTree` struct if not present.
- [ ] 3.3 In `src/game/npc.game.go` floor-entry path, iterate all major NPCs, call `PhaseTracker.ShouldAdvance(RunState)`, call `Advance()` if true, update `QuestFlags` to reflect new phase.
- [ ] 3.4 `cd src && go build ./... && go test ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark roadmap tasks 5.1, 5.2, 5.3 ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 5 status block.
- [ ] 4.3 Move this plan to `plans/COMPLETED/5A-npc-phase-tracker.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **`RequiresPhase` on DialogueTree.** If the existing JSON trees use a different mechanism to express phase selection (e.g., a top-level condition), prefer that over adding a new field. Check existing `dialogues/varn_*.json` before adding the field. Affects 3.1–3.2.
- **Phase tracker storage.** `NPCPhaseTracker` needs to survive across floors within a run. It can live on `RunState` (serialized) or be reconstructed from `QuestFlags` each floor. Recommendation: reconstruct from QuestFlags to avoid a new serialization surface until `runstate-serialization` ships. Affects 1.2 and 3.3.
