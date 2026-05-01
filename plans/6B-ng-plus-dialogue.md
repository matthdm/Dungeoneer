---
plan-id: 6B-ng-plus-dialogue
status: queued
owner: unassigned
branch: plan/6B-ng-plus-dialogue
depends-on: [6A-full-meta-save]
last-touched: 2026-04-30
---

# Plan: Phase 6B — NG+ Dialogue & Meta-Flag Conditions

## Goal

Make NPCs remember across runs. Varn, after being defeated, reacts to subsequent encounters differently — recognition on the first NG+ run, self-doubt on the second, meta-awareness on the third. Betrayal in a previous run makes him hostile on re-encounter. Trust accumulates across runs. This plan delivers the infrastructure for meta-flag conditions in dialogue and Varn's full NG+ tree set.

## Scope

**In scope:**
- Two new dialogue condition types: `meta_flag_gte` and `meta_flag_equals` that read `NPCMeta` fields (defeat count, betrayed, highest phase).
- Varn NG+ dialogue trees for runs 1, 2, and 3+ post-defeat.
- Betrayal memory tree (hostile/suspicious variant for players who set `varn_betrayed`).
- Trust accumulation: `NPCMeta.TotalTrust` sums across runs; high total trust unlocks deeper lore dialogue.

**Out of scope (do not change in this plan):**
- Other NPCs' NG+ dialogue (Phase 10).
- Lore unlock content (Phase 6C).
- Abaddon (Phase 10B).

## File envelope

**Touched:**
- `src/dialogue/types.go` — add `meta_flag_gte`, `meta_flag_equals` condition types
- `src/game/npc.game.go` — evaluate meta-flag conditions; accumulate `TotalTrust` at run end
- `src/game/runstate.go` — add `TotalTrust int` to `NPCMeta` if not present
- `src/dialogues/varn_ng1.json` *(new)* — post-defeat run 1
- `src/dialogues/varn_ng2.json` *(new)* — post-defeat run 2
- `src/dialogues/varn_ng3.json` *(new)* — post-defeat run 3+
- `src/dialogues/varn_betrayed.json` *(new)* — hostile re-encounter variant
- `design-docs/roadmap.md` — mark 6.4–6.7 ✅ on completion
- `CLAUDE.md` — update Phase 6 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/entities/varn_boss.go` — boss plan scope
- `src/dialogues/varn_phase*.json` — Varn arc plan scope

## Acceptance criteria

- [ ] `meta_flag_gte "npc_id" "field" N` condition evaluates `NPCMeta[npc_id][field] >= N` where field can be `defeat_count`, `total_trust`, or `highest_phase`.
- [ ] `meta_flag_equals "npc_id" "field" N` evaluates equality.
- [ ] After Varn is defeated for the first time, his dungeon dialogue on the next run loads `varn_ng1.json` (recognition — he knows the player has beaten him before).
- [ ] Second defeat leads to `varn_ng2.json` (doubt — cracks in his certainty).
- [ ] Third+ defeat leads to `varn_ng3.json` (meta-awareness — he knows he is caught in a loop).
- [ ] If `varn_betrayed = 1` in MetaSave, Varn loads `varn_betrayed.json` variant — hostile opening, requires higher trust to progress.
- [ ] `NPCMeta["varn"].TotalTrust` accumulates across runs; a run's ending trust is added to TotalTrust in the death/victory handler.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Meta-flag condition types
- [ ] 1.1 In `src/dialogue/types.go`, add condition kinds `"meta_flag_gte"` and `"meta_flag_equals"`. Each has fields: `NPCID string`, `Field string`, `Value int`.
- [ ] 1.2 In `src/game/npc.game.go` condition evaluator, handle both: read `MetaSave.NPCMeta[NPCID]`, switch on `Field` (`"defeat_count"` → `DefeatCount`, `"total_trust"` → `TotalTrust`, `"highest_phase"` → `HighestPhase`).
- [ ] 1.3 In `src/game/runstate.go`, add `TotalTrust int` and `HighestPhase int` to `NPCMeta` if absent.
- [ ] 1.4 In death/victory handler, for each NPC: add run trust to `NPCMeta[id].TotalTrust`, update `HighestPhase` if current phase is higher.
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: NG+ dialogue trees
- [ ] 2.1 Write `src/dialogues/varn_ng1.json`. Varn opens with quiet recognition ("You again. I remember the weight of your blade."). 2–3 nodes. No quest progression — this is atmosphere and lore.
- [ ] 2.2 Write `src/dialogues/varn_ng2.json`. Varn shows cracks ("I was so certain. But here you are, and here I am, again."). Deeper philosophical doubt. One lore-gated branch for `total_trust >= 60`.
- [ ] 2.3 Write `src/dialogues/varn_ng3.json`. Varn is quietly meta-aware ("How many times have you walked this dungeon? How many times have I forgotten?"). One unique response if `total_trust >= 80` unlocks a deeper lore entry (prep for lore system in 6C).
- [ ] 2.4 Write `src/dialogues/varn_betrayed.json`. Varn opens hostile ("Betrayer. You dare return."). Higher trust threshold required to advance dialogue. Eventually reaches the same phase arc but with scarred text.
- [ ] 2.5 Update `dialogue/loader.go` `SelectTree()` to check `defeat_count` and `betrayed` meta flags when selecting Varn's tree.
- [ ] 2.6 `cd src && go build ./...` passes.

### Phase 3: Cleanup
- [ ] 3.1 Mark roadmap rows 6.4–6.7 ✅ in `design-docs/roadmap.md`.
- [ ] 3.2 Update `CLAUDE.md` Phase 6 status block.
- [ ] 3.3 Move this plan to `plans/COMPLETED/6B-ng-plus-dialogue.md`.
- [ ] 3.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **MetaSave vs. RunState for betrayal flag.** `varn_betrayed` is set in `RunState.QuestFlags` during a run. It must be persisted to `MetaSave.NPCMeta["varn"]` at run end. Confirm the run-end handler copies relevant quest flags to NPCMeta before clearing RunState. Affects 1.4.
- **Tree selection priority.** When Varn has been defeated AND betrayed, which tree wins? Recommendation: betrayed takes priority over NG+ defeat count — betrayal is the more emotionally charged state. Affects 2.5.
