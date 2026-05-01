---
plan-id: 5D-hub-npc-quarter
status: queued
owner: unassigned
branch: plan/5D-hub-npc-quarter
depends-on: [5B-varn-arc]
last-touched: 2026-04-30
---

# Plan: Phase 5D — Hub NPC Quarter

## Goal

Major NPCs the player has met appear in the hub between runs, positioned at fixed spots in an NPC quarter area of the hub. Each has between-run dialogue that reflects quest progress and hints at what the next encounter will bring. This makes the hub feel alive and gives the player a reason to walk through it rather than rush to the portal.

When done, Varn (the only major NPC at this phase) appears in the hub after the player first meets him in the dungeon, at a fixed tile position, with context-appropriate dialogue trees.

## Scope

**In scope:**
- Fixed tile positions for up to 6 major NPCs defined in `levels/hub.json`.
- Hub spawning logic: check `NPCMeta[id].Met` for each registered major NPC, spawn at their assigned hub position.
- Varn's hub dialogue tree (`dialogues/varn_hub.json`) with 3–4 nodes reflecting phase progress.
- Visual separation of the hub NPC quarter (a region of the hub map, not a new level).

**Out of scope (do not change in this plan):**
- Hub shop and upgrade station NPCs (Phase 8B).
- Other major NPCs' hub dialogue (Seris, Mira, Kael — Phase 10).
- Hub visual evolution based on milestones (Phase 10C).

## File envelope

**Touched:**
- `src/levels/hub.json` — add NPC quarter positions for up to 6 major NPCs
- `src/game/hub.go` — hub NPC spawn pass using `NPCMeta.Met` checks
- `src/game/npc.game.go` — ensure hub NPC dialogue triggers work in hub state
- `src/dialogues/varn_hub.json` *(new)*
- `design-docs/roadmap.md` — mark 5.15, 5.16, 5.17 ✅ on completion
- `CLAUDE.md` — update Phase 5 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/entities/varn_boss.go` — boss plan scope
- `src/dialogues/varn_phase*.json` — Varn arc plan scope
- Shop or upgrade station code — Phase 8B scope

## Acceptance criteria

- [ ] `hub.json` contains named tile positions for major NPCs: `npc_varn`, `npc_seris`, `npc_mira`, `npc_kael`, plus 2 reserved slots.
- [ ] After first meeting Varn in the dungeon (`NPCMeta["varn"].Met == true`), Varn appears in the hub at his designated position on subsequent runs.
- [ ] Varn's hub dialogue (`varn_hub.json`) has at least 3 phase-conditional root nodes: pre-Phase-1 ("I wondered if you'd come back"), post-Phase-1 ("You proved yourself"), post-Phase-3 ("What happens next is written").
- [ ] Hub NPC dialogue trigger works correctly (interaction range, "[E] Talk" prompt, dialogue panel).
- [ ] NPCs not yet met do not appear in the hub.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Hub NPC positions
- [ ] 1.1 In `src/levels/hub.json`, add a `"npc_positions"` object mapping NPC IDs to `{x, y}` tile coords. Place them in a visually distinct area of the hub (e.g., near a hearth or alcove, away from the portal).
- [ ] 1.2 Add 6 slots: `varn`, `seris`, `mira`, `kael`, `reserved_1`, `reserved_2`. Only Varn will be used this phase; others are placeholders.
- [ ] 1.3 `cd src && go build ./...` passes (no logic change yet, just data).

### Phase 2: Hub NPC spawn pass
- [ ] 2.1 In `src/game/hub.go`, after loading the hub level, iterate all entries in `npc_positions`. For each, check `MetaSave.NPCMeta[id].Met`; if true, spawn an NPC entity at that position with the NPC's hub dialogue ID.
- [ ] 2.2 Ensure the hub NPC uses the correct sprite for their current phase (e.g., ascended Varn sprite if `varn_p3_done`).
- [ ] 2.3 `cd src && go build ./...` passes.

### Phase 3: Varn hub dialogue
- [ ] 3.1 Write `src/dialogues/varn_hub.json`. Three root nodes selected by `varn_p3_done` and `varn_p1_done`:
  - Default (phase 0 met): reflective greeting, philosophical observation, hint about next encounter.
  - Phase 1+ done: acknowledges the task completed, anticipates the conflict ahead.
  - Phase 3 done: quiet, resolute. Acknowledges what is coming (the boss fight).
- [ ] 3.2 No trust-gated actions needed in hub dialogue (hub is a rest space — trust builds in the dungeon).
- [ ] 3.3 Validate the dialogue JSON loads without error by running the game and triggering the hub NPC.
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark roadmap rows 5.15, 5.16, 5.17 ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 5 status block to "complete" if 5A–5D are all done.
- [ ] 4.3 Move this plan to `plans/COMPLETED/5D-hub-npc-quarter.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Hub map bounds.** Confirm the hub tile dimensions from `hub.json` before choosing NPC positions. They should not overlap with the portal, chest spawns, or wall tiles. Affects 1.1.
- **Ascended sprite in hub.** If `varn_p3_done` is set, should the hub show the ascended Varn sprite? Recommendation: yes, it reinforces that something has changed. This requires `hub.go` to read the phase when assigning the sprite. Affects 2.2.
