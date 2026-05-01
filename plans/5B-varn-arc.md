---
plan-id: 5B-varn-arc
status: queued
owner: unassigned
branch: plan/5B-varn-arc
depends-on: [5A-npc-phase-tracker]
last-touched: 2026-04-30
---

# Plan: Phase 5B — Varn's Full Dialogue Arc

## Goal

Write and wire Varn's complete four-phase dialogue arc — Phase 0 (Introduction), Phase 1 (First Task), Phase 2 (Conflict / moral choice), and Phase 3 (Ascension Trigger). When done, a player who repeatedly encounters Varn across a run experiences a coherent character arc where their choices determine whether Varn ascends peacefully or violently, which in turn shapes the boss fight.

This plan is content-heavy: most of the work is JSON dialogue trees, not Go code. The Go wiring touches only `npc_data.go` (Varn's definition) and `npc.game.go` (ascension visual change).

## Scope

**In scope:**
- Varn's `MajorNPCDef` entry in `npc_data.go`: sprite, placement tag, 4 dialogue tree IDs, phase tracker conditions.
- Four phase dialogue JSON files with full branching, conditions, and actions.
- Phase 3 ascension visual: Varn's sprite changes and a brief animation/effect on ascension trigger.
- Varn's spawning logic (floors 2+ in sanctuary rooms).

**Out of scope (do not change in this plan):**
- The boss fight itself (that is `5C-boss-selection.md`).
- Hub NPC dialogue (that is `5D-hub-npc-quarter.md`).
- Trust system mechanics (built in `5A-npc-phase-tracker.md`).

## File envelope

**Touched:**
- `src/game/npc_data.go` — add Varn `MajorNPCDef` with full phase tracker config
- `src/game/npc.game.go` — Varn spawning on floors 2+; ascension visual trigger
- `src/dialogues/varn_phase0.json` *(new)*
- `src/dialogues/varn_phase1.json` *(new)*
- `src/dialogues/varn_phase2.json` *(new)*
- `src/dialogues/varn_phase3.json` *(new)*
- `design-docs/roadmap.md` — mark 5.4–5.9 ✅ on completion
- `CLAUDE.md` — update Phase 5 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/entities/varn_boss.go` — boss plan scope
- `src/entities/boss.go` — boss plan scope
- `src/dialogue/loader.go`, `src/dialogue/types.go` — phase tracker plan owns these
- Any other NPC's dialogue files

## Acceptance criteria

- [ ] Varn appears in sanctuary-tagged rooms on floors 2 and above.
- [ ] First encounter triggers `varn_phase0.json` (Introduction). Player learns his philosophy and backstory.
- [ ] After meeting the Phase 0 quest flag threshold, Varn loads `varn_phase1.json` on next encounter (First Task). He requests a specific item or action.
- [ ] After Phase 1 completion, Varn loads `varn_phase2.json` (Conflict). Player faces a meaningful binary choice that sets a `varn_betrayed` flag.
- [ ] After Phase 2 completion (either path), Varn loads `varn_phase3.json` (Ascension Trigger). His sprite visually changes; a line of dialogue acknowledges the player's Phase 2 choice.
- [ ] All four trees use `add_trust`, `trust_gte`, `give_item`, `has_item`, and `set_flag` actions correctly.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Varn NPC definition
- [ ] 1.1 In `src/game/npc_data.go`, define Varn as a `MajorNPCDef`: ID `"varn"`, sprite `"npc_chainkeeper"`, placement tag `sanctuary`, min floor 2, `PhaseTrackerConfig` with 4 phase conditions drawn from quest flags `varn_met`, `varn_p1_done`, `varn_p2_done`, `varn_p3_done`.
- [ ] 1.2 In `src/game/npc.game.go`, add Varn to the major NPC spawn pass: check `NPCMeta["varn"].Phase` and `currentFloor >= 2`, spawn in a sanctuary room if not yet spawned this floor.
- [ ] 1.3 `cd src && go build ./...` passes.

### Phase 2: Phase 0 — Introduction
- [ ] 2.1 Write `src/dialogues/varn_phase0.json`. Nodes: greeting (sets `varn_met = 1`), philosophy monologue (2 player responses — curious / skeptical), backstory reveal (unlocks on second visit via `varn_met >= 1`), farewell.
- [ ] 2.2 Trust actions: curious response `add_trust 10`, skeptical response `add_trust 3`.
- [ ] 2.3 Ensure `SelectTree()` picks this tree when `varn` phase == 0.
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Phase 1 — First Task
- [ ] 3.1 Write `src/dialogues/varn_phase1.json`. Varn asks for a specific item (e.g., `iron_key` or a quest item). Nodes: request, player has item (branch: give / refuse), player lacks item (hint at where to find it), completion (sets `varn_p1_done = 1`, `add_trust 20`), refusal (sets `varn_refused = 1`, `trust_decay 15`).
- [ ] 3.2 Use `has_item` condition and `give_item` / `take_item` actions.
- [ ] 3.3 `cd src && go build ./...` passes.

### Phase 4: Phase 2 — Conflict
- [ ] 4.1 Write `src/dialogues/varn_phase2.json`. Varn reveals something morally questionable about his methods. Nodes: revelation, player choice (support / challenge). Support: `add_trust 15`, sets `varn_supported = 1`. Challenge: `trust_decay 10`, sets `varn_betrayed = 1`. Both paths set `varn_p2_done = 1`.
- [ ] 4.2 Both branches must feel like genuine choices — neither is obviously "good." The text must acknowledge prior Phase 1 outcome if `varn_refused` is set.
- [ ] 4.3 `cd src && go build ./...` passes.

### Phase 5: Phase 3 — Ascension
- [ ] 5.1 Write `src/dialogues/varn_phase3.json`. Two root nodes gated by `varn_betrayed`: one for allies (grateful, resolute), one for betrayers (cold, resigned). Both set `varn_p3_done = 1` and trigger ascension.
- [ ] 5.2 In `src/game/npc.game.go`, on `varn_p3_done` flag set: swap Varn's sprite to `"npc_chainkeeper_ascended"`, play a brief visual pulse effect (reuse any existing effect emitter).
- [ ] 5.3 `cd src && go build ./...` passes.

### Phase 6: Cleanup
- [ ] 6.1 Mark roadmap rows 5.4–5.9 ✅ in `design-docs/roadmap.md`.
- [ ] 6.2 Update `CLAUDE.md` Phase 5 status block.
- [ ] 6.3 Move this plan to `plans/COMPLETED/5B-varn-arc.md`.
- [ ] 6.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Varn's Phase 1 quest item.** The existing `items/load.go` already injects Grips for Varn phase 1 and Chaos Emblem for phase 2. Confirm the exact item IDs before writing `has_item` conditions. Check `src/items/load.go` and `src/game/progression.game.go`. Affects 3.1.
- **Ascended sprite.** If `npc_chainkeeper_ascended` does not exist in the spritesheet, use the existing sprite with a color tint until an art pass. Document this in "What was NOT changed." Affects 5.2.
- **Phase 2 moral scenario.** Varn's morally questionable act is not yet specified in the design docs. Suggested: he has been secretly draining the life force of the dungeon's weaker entities to maintain the chains that bind him (a form of necessary evil). The player can endorse this as pragmatic or reject it as monstrous. Lock in the scenario before writing the JSON. Affects 4.1.
