---
plan-id: 10C-world-expansion
status: queued
owner: unassigned
branch: plan/10C-world-expansion
depends-on: [10B-abaddon-meta-narrative]
last-touched: 2026-04-30
---

# Plan: Phase 10C — World Expansion (Biomes, Hub Evolution & NPC Cross-Talk)

## Goal

Add four new biomes (Gehenna, Pandemonium, Cocytus, Lapis), biome-specific encounter templates, hub visual evolution based on progression milestones, and cross-NPC dialogue in the hub where major NPCs comment on each other and react to player alignment. This is the final content pass that makes the world feel complete and lived-in.

## Scope

**In scope:**
- 4 new `BiomeConfig` entries: Gehenna (fire), Pandemonium (chaos), Cocytus (ice), Lapis (crystal).
- Biome-specific encounter templates for each new biome (2–3 encounter patterns per biome).
- Hub visual evolution: conditional tile/decoration overlays based on `HubState` flags (torches light as NPCs arrive, banners appear after boss defeats).
- Cross-NPC hub dialogue: Varn comments on Seris, Mira reacts to Kael, etc. Triggered when both NPCs are present in the hub.
- Hub NPC reactions to player alignment.

**Out of scope (do not change in this plan):**
- New enemy type art assets (use existing enemy types with biome-themed color variants for now).
- New map generation algorithms (use existing `generate64.go` with new biome configs).

## File envelope

**Touched:**
- `src/game/biome.go` — add 4 new `BiomeConfig` entries
- `src/game/encounters.go` — add biome-specific encounter templates for new biomes
- `src/levels/hub.json` — add conditional decoration layers for hub evolution
- `src/game/hub.go` — apply hub decoration overlays based on `HubState`; spawn cross-NPC dialogue triggers
- `src/dialogues/hub_cross_varn_seris.json` *(new)* — Varn and Seris cross-talk
- `src/dialogues/hub_cross_mira_kael.json` *(new)* — Mira and Kael cross-talk
- `src/dialogues/hub_cross_abaddon_alignment.json` *(new)* — Abaddon reacts to alignment
- `design-docs/roadmap.md` — mark 9.10–9.13 ✅ on completion
- `CLAUDE.md` — update Phase 10 status block to "🟢 COMPLETE"
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/levels/generate64.go` — biome configs are consumed by existing generation, not changed
- Boss entities — complete
- `src/dialogue/` engine — no new condition/action types needed

## Acceptance criteria

- [ ] Four new `BiomeConfig` entries exist: Gehenna (fire-red palette, more caster/ranged enemies), Pandemonium (chaotic purple, swarm-heavy), Cocytus (ice-blue, patrol-heavy with slow-applying enemies), Lapis (crystal-teal, ambush-heavy).
- [ ] Each new biome has at least 2 encounter templates registered in `encounters.go`.
- [ ] Hub visuals change as milestones unlock: no torches on first run; torches appear after first NPC met; banner appears after first boss defeated; full decoration after all bosses defeated.
- [ ] When both Varn and Seris are in the hub, an ambient cross-NPC dialogue trigger appears (small indicator); interacting plays their exchange.
- [ ] When both Mira and Kael are present, similar cross-talk is available.
- [ ] Abaddon comments on player alignment in `hub_cross_abaddon_alignment.json` — separate from his main dialogue tier, triggered by returning to the hub after an alignment shift.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: New biomes
- [ ] 1.1 In `src/game/biome.go`, add `BiomeConfig` for Gehenna, Pandemonium, Cocytus, Lapis. Each needs: wall/floor tile IDs (use existing tiles with tint if no new tiles exist), enemy pool, music ID (use existing ambient for now), loot bias.
- [ ] 1.2 In `src/game/encounters.go`, add 2 encounter templates per new biome using existing enemy types.
- [ ] 1.3 Register new biomes in the floor assignment logic so they appear on appropriate floors (e.g., Gehenna on floors 5+).
- [ ] 1.4 `cd src && go build ./...` passes.

### Phase 2: Hub visual evolution
- [ ] 2.1 In `src/levels/hub.json`, add a `"decorations"` array: each entry has a tile position, tile ID, and a `"requires_hub_state"` key.
- [ ] 2.2 In `src/game/hub.go` hub load, iterate `decorations`; render those whose `requires_hub_state` is set in `MetaSave.HubState`.
- [ ] 2.3 Define decoration states: `"first_npc_met"` → torches; `"first_boss_defeated"` → banner; `"all_bosses_defeated"` → full decoration set.
- [ ] 2.4 Add corresponding `HubState` flag updates in the run-end handler.
- [ ] 2.5 `cd src && go build ./...` passes.

### Phase 3: Cross-NPC dialogue
- [ ] 3.1 Write `hub_cross_varn_seris.json`: Varn and Seris discuss order vs. destruction — a short 4-node exchange that acknowledges both philosophies.
- [ ] 3.2 Write `hub_cross_mira_kael.json`: Mira and Kael discuss illusion vs. organic truth.
- [ ] 3.3 Write `hub_cross_abaddon_alignment.json`: Abaddon comments on the player's alignment axis (3 variants per axis extreme).
- [ ] 3.4 In `src/game/hub.go`, when both NPC pairs are present, spawn a cross-talk trigger entity between them. Interaction plays the appropriate JSON tree via the existing dialogue engine.
- [ ] 3.5 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark roadmap rows 9.10–9.13 ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 10 status block to "🟢 COMPLETE".
- [ ] 4.3 Update `README.md` to reflect the game is feature-complete.
- [ ] 4.4 Move this plan to `plans/COMPLETED/10C-world-expansion.md`.
- [ ] 4.5 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **New biome tile art.** If no distinct tile art exists for the new biomes, use existing dungeon tiles with `ColorScale` tinting (red for Gehenna, purple for Pandemonium, etc.). Document the gap. Affects 1.1.
- **Cross-talk trigger entity.** This is a new entity type (non-NPC, non-item, non-monster). Recommend reusing a simple interactable entity pattern similar to the chest entity but triggering a dialogue tree. Affects 3.4.
