---
plan-id: 10A-additional-npcs
status: queued
owner: unassigned
branch: plan/10A-additional-npcs
depends-on: [9D-polish-hud-navigation, 9C-polish-audio]
last-touched: 2026-04-30
---

# Plan: Phase 10A — Additional Major NPCs (Seris, Mira, Kael)

## Goal

Implement the three remaining major NPCs — Seris the Ember, Mira the Veil, and Kael the Root — each with a full 4-phase dialogue arc, a unique 3-phase boss form, and a distinct philosophical theme. These NPCs follow the same infrastructure built in Phase 5 (NPCPhaseTracker, trust system, BossSelectionEngine), requiring only data and content, not new engine code.

When done, the boss selection engine can pick from four possible final bosses (Varn + three new ones) based on player choices, making each run's endgame narratively unique.

## Scope

**In scope:**
- `MajorNPCDef` entries for Seris, Mira, and Kael in `npc_data.go`.
- Full dialogue arcs (4 phases × 3 NPCs = 12 new JSON files).
- Boss entities for each NPC (`entities/bosses/seris.go`, `mira.go`, `kael.go`).
- Boss pre/post dialogue (6 new JSON files).
- Hub dialogue (3 new JSON files).
- Boss selection engine updated to include the three new boss types.

**Out of scope (do not change in this plan):**
- Abaddon or alignment system (Phase 10B).
- Biome expansion (Phase 10C).
- New engine mechanics — these NPCs use the existing phase tracker, trust, and boss systems.

## File envelope

**Touched:**
- `src/game/npc_data.go` — add Seris, Mira, Kael `MajorNPCDef` entries
- `src/entities/bosses/seris.go` *(new)*
- `src/entities/bosses/mira.go` *(new)*
- `src/entities/bosses/kael.go` *(new)*
- `src/game/boss_selection.go` — add `BossSeris`, `BossMira`, `BossKael` types and selection logic
- `src/dialogues/seris_phase0-3.json` *(4 new files)*
- `src/dialogues/seris_boss_pre.json`, `seris_boss_post.json` *(2 new files)*
- `src/dialogues/seris_hub.json` *(1 new file)*
- `src/dialogues/mira_phase0-3.json` *(4 new files)*
- `src/dialogues/mira_boss_pre.json`, `mira_boss_post.json`, `mira_hub.json` *(3 new files)*
- `src/dialogues/kael_phase0-3.json` *(4 new files)*
- `src/dialogues/kael_boss_pre.json`, `kael_boss_post.json`, `kael_hub.json` *(3 new files)*
- `design-docs/roadmap.md` — mark 9.1–9.4 ✅ on completion
- `CLAUDE.md` — update Phase 10 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/dialogue/loader.go`, `src/dialogue/types.go` — no new action/condition types needed
- `src/entities/varn_boss.go` — Varn arc is complete
- `src/game/npc_phases.go` — phase tracker infrastructure is complete

## Acceptance criteria

- [ ] Seris appears in fire-biome sanctuary rooms; her arc concerns destruction as renewal.
- [ ] Mira appears in mid-dungeon sanctuary rooms; her arc concerns truth vs. comfortable illusion.
- [ ] Kael appears in nature-adjacent rooms; his arc concerns growth vs. stagnation.
- [ ] Each NPC has 4 phase dialogue trees matching the quality and branching depth of Varn's arc.
- [ ] Each NPC's boss form has 3 distinct phases with unique attack patterns (no copy-paste from Varn).
- [ ] `BossSelectionEngine` picks the highest-phase NPC across all 4 major NPCs; tie-broken by trust level.
- [ ] Hub dialogue for all three appears when `NPCMeta[id].Met == true`.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: NPC definitions and boss selection update
- [ ] 1.1 Add `MajorNPCDef` for Seris (sprite `npc_ember`, sanctuary rooms in fire biomes, min floor 2), Mira (sprite `npc_veil`, any sanctuary, min floor 3), Kael (sprite `npc_root`, nature rooms, min floor 2).
- [ ] 1.2 Update `boss_selection.go`: add `BossSeris`, `BossMira`, `BossKael`; `SelectBoss()` evaluates all four NPCs and picks the one at the highest phase (tie → highest trust).
- [ ] 1.3 `cd src && go build ./...` passes.

### Phase 2: Boss entities
- [ ] 2.1 Seris boss (`seris.go`): Phase 1 — fire trail melee; Phase 2 (66% HP) — flame pillar AoE columns; Phase 3 (33% HP) — conflagration (ring of fire expanding outward).
- [ ] 2.2 Mira boss (`mira.go`): Phase 1 — illusory clones (2–3 decoys that vanish on hit); Phase 2 (66%) — reality inversion (reverses player movement briefly); Phase 3 (33%) — veil collapse (rapid multi-hit across the arena).
- [ ] 2.3 Kael boss (`kael.go`): Phase 1 — root spike melee; Phase 2 (66%) — vine eruptions (ground-targeted, delayed); Phase 3 (33%) — canopy surge (rains nature projectiles from above).
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Dialogue arcs — Seris
- [ ] 3.1 Write `seris_phase0.json` through `seris_phase3.json`. Theme: fire as both destruction and purification. Phase 2 choice: does the player accept that some things must burn to allow rebirth?
- [ ] 3.2 Write `seris_boss_pre.json`, `seris_boss_post.json`, `seris_hub.json`.
- [ ] 3.3 `cd src && go build ./...` passes.

### Phase 4: Dialogue arcs — Mira and Kael
- [ ] 4.1 Write `mira_phase0-3.json`. Theme: illusion as mercy vs. truth as cruelty. Phase 2 choice: does the player prefer uncomfortable truths or kind fictions?
- [ ] 4.2 Write `mira_boss_pre.json`, `mira_boss_post.json`, `mira_hub.json`.
- [ ] 4.3 Write `kael_phase0-3.json`. Theme: growth requires decay. Phase 2 choice: does the player embrace letting old things die to make room for new ones?
- [ ] 4.4 Write `kael_boss_pre.json`, `kael_boss_post.json`, `kael_hub.json`.
- [ ] 4.5 `cd src && go build ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 Mark roadmap rows 9.1–9.4 ✅ in `design-docs/roadmap.md`.
- [ ] 5.2 Update `CLAUDE.md` Phase 10 status block.
- [ ] 5.3 Move this plan to `plans/COMPLETED/10A-additional-npcs.md`.
- [ ] 5.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Mira's reality inversion mechanic.** Reversing player movement is a novel mechanic not in the existing input system. Recommend: add a `MovementInverted bool` flag to the player that the movement controller checks. Simple, surgical. Affects 2.2.
- **Biome gating for Seris.** Fire biome doesn't exist yet (Phase 10C). Until then, Seris can appear in any sanctuary room. Add a `BiomeRestriction []string` field to `MajorNPCDef` and leave it empty for Seris until the biome exists. Affects 1.1.
