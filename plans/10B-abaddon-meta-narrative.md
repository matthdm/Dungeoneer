---
plan-id: 10B-abaddon-meta-narrative
status: queued
owner: unassigned
branch: plan/10B-abaddon-meta-narrative
depends-on: [10A-additional-npcs]
last-touched: 2026-04-30
---

# Plan: Phase 10B — Abaddon, Alignment System & Endgame Narrative

## Goal

Implement the meta-narrative capstone: Abaddon appears in the hub after 10+ runs as a philosophical observer who cannot be fought, speaks in riddles, and gradually becomes more direct as the player accumulates runs, NPC defeats, and lore. The alignment system tracks player choices across all NPC arcs on two axes (Order↔Chaos, Creation↔Destruction) and gates Abaddon's dialogue tiers and final endgame resolution text.

When done, the game has a meaningful philosophical conclusion that reflects the player's accumulated choices across all runs and NPC arcs.

## Scope

**In scope:**
- `Alignment` struct in `game/alignment.go`: two axes, each [-100, 100].
- Alignment-affecting actions tagged to existing NPC phase choices (Varn, Seris, Mira, Kael).
- Abaddon NPC entity: appears in hub after 10 completed runs, has 5+ dialogue tiers unlocked by run count, NPC defeats, lore collected, and alignment.
- Alignment-gated dialogue conditions for all NPCs (responses shift based on player's philosophical lean).
- Endgame narrative dialogue for each NPC archetype (post-all-boss-defeat).

**Out of scope (do not change in this plan):**
- World expansion (Phase 10C).
- New gameplay mechanics.

## File envelope

**Touched:**
- `src/game/alignment.go` *(new)* — `Alignment` struct, axis update, tier classification
- `src/game/metasave.go` — add `Alignment Alignment`
- `src/game/runstate.go` — add `add_alignment` action wiring
- `src/dialogue/types.go` — add `add_alignment`, `alignment_gte`, `alignment_lte` action/conditions
- `src/game/npc.game.go` — execute `add_alignment`; evaluate alignment conditions
- `src/game/npc_data.go` — add Abaddon `MajorNPCDef`
- `src/dialogues/abaddon_tier0-4.json` *(5 new files)*
- `src/dialogues/endgame_varn.json`, `endgame_seris.json`, `endgame_mira.json`, `endgame_kael.json` *(4 new files)*
- `design-docs/roadmap.md` — mark 9.5–9.9 ✅ on completion
- `CLAUDE.md` — update Phase 10 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- Boss entities — all boss fights are complete
- `src/entities/varn_boss.go` and analogues

## Acceptance criteria

- [ ] `Alignment` has `Order int` (-100 to 100) and `Creation int` (-100 to 100). Updates are clamped.
- [ ] `add_alignment "order" 10` action works in dialogue; `alignment_gte "order" 50` condition works.
- [ ] Existing Varn phase 2 choice sets alignment (supporting Varn → +10 Order; challenging → +10 Chaos). Seris, Mira, Kael choices similarly tagged.
- [ ] Abaddon appears in hub after `MetaSave.CompletedRuns >= 10`. Has 5 dialogue tiers unlocked by: run count, total NPC defeats, lore entries collected, and alignment score.
- [ ] Abaddon's lowest tier is cryptic and impersonal; his highest tier directly acknowledges the loop and the player's accumulated choices.
- [ ] After all 4 bosses are defeated (tracked in MetaSave), each NPC's endgame dialogue becomes available in the hub — a final conversation that closes their arc.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Alignment system
- [ ] 1.1 Create `src/game/alignment.go`. `Alignment` struct with `Order, Creation int`. `Update(axis string, delta int)` clamps to [-100, 100].
- [ ] 1.2 `Tier() AlignmentTier` classifies into 5 tiers per axis for dialogue gating.
- [ ] 1.3 In `src/dialogue/types.go`, add `add_alignment` action (axis string, value int) and `alignment_gte` / `alignment_lte` conditions.
- [ ] 1.4 In `src/game/npc.game.go`, handle the new action/conditions using `MetaSave.Alignment`.
- [ ] 1.5 In `src/game/metasave.go`, add `Alignment Alignment`.
- [ ] 1.6 `cd src && go build ./...` passes.

### Phase 2: Tag existing NPC choices with alignment
- [ ] 2.1 In `varn_phase2.json`: support branch → `add_alignment "order" 10`; challenge branch → `add_alignment "order" -10`.
- [ ] 2.2 In `seris_phase2.json`: accept destruction branch → `add_alignment "creation" -10`; reject branch → `add_alignment "creation" 10`.
- [ ] 2.3 In `mira_phase2.json` and `kael_phase2.json`: tag appropriately per their philosophical axis.
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Abaddon
- [ ] 3.1 Add Abaddon to `npc_data.go`: sprite `npc_abaddon`, fixed hub position (center hub, near the portal), appears when `CompletedRuns >= 10`, no phase tracker (his tiers are directly meta-flag-gated).
- [ ] 3.2 Write 5 dialogue tiers (`abaddon_tier0.json` through `abaddon_tier4.json`). Each unlocks by a combination of run count, defeat count, lore count, and alignment extremity. Tone progresses from cryptic → philosophical → direct → intimate → resolving.
- [ ] 3.3 In `src/game/hub.go`, spawn Abaddon when `CompletedRuns >= 10`.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Endgame narrative
- [ ] 4.1 Track `AllBossesDefeated bool` in MetaSave: true when all 4 NPCs have `DefeatCount >= 1`.
- [ ] 4.2 Write `endgame_varn.json`, `endgame_seris.json`, `endgame_mira.json`, `endgame_kael.json` — final closing conversations for each NPC in the hub. These are only available after `AllBossesDefeated`.
- [ ] 4.3 In `src/game/hub.go`, check `AllBossesDefeated`; if true, load endgame dialogue for any NPC with `DefeatCount >= 1` and `NPCMeta[id].Met`.
- [ ] 4.4 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark roadmap rows 9.5–9.9 ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 10 status block.
- [ ] 5.3 Move this plan to `plans/COMPLETED/10B-abaddon-meta-narrative.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Abaddon's sprite.** If no Abaddon sprite exists, use a heavily tinted existing NPC sprite (pure white or black with opacity). Document the gap. Affects 3.1.
- **Alignment display.** Should the player be able to see their alignment scores? Recommendation: no explicit display — let it manifest only through Abaddon's dialogue reactions. Preserve the mystery. Affects Phase 1 scope.
