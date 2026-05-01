---
plan-id: 4F-loot-refinement
status: active
owner: unassigned
branch: plan/4F-loot-refinement
depends-on: []
last-touched: 2026-04-27
---

# Plan: Phase 4F — Loot Refinement

## Goal

Ship the three open Phase 4F items from `design-docs/roadmap.md`: chest variants (Wooden / Iron / Gold) with quality-tiered loot, treasure-room loot population (rooms tagged `treasure` always contain a chest), and floor-depth-weighted loot rarity. Locked Iron and Gold chests require an Iron Key to open.

When done, every run produces a clear visual loot economy: the player learns to recognize chest tiers, hunts for keys to unlock the better ones, and feels rarity escalation across floors.

## Scope

**In scope:**
- New `ChestVariant` enum and `Locked` field on `entities/chest.go`.
- Iron Key item registered in `items/load.go`; consumed when opening a locked chest.
- Treasure-room chest spawning in floor generation (after room tagging).
- Floor-depth scaling for loot rarity weights, matching the table in `design-docs/biome-system.md`.
- Hint text tweaks for locked-chest interaction.

**Out of scope (do not change in this plan):**
- Hub shop and run loadout (Phase 4E, deferred to Phase 8).
- Item set bonuses (Phase 8).
- New ability items (4B is shipped; do not add more here).
- Anything in `src/coords/` or `src/collision/` (offset plan owns these).
- Dialogue trees, NPCs, boss work.

## File envelope

**Touched:**
- `src/entities/chest.go`
- `src/items/loot.go`
- `src/items/load.go` (Iron Key registration only)
- `src/game/handlers.game.go` (locked chest interaction)
- `src/game/hub.go` (treasure-room chest spawning during floor gen)
- `src/items/loot_test.go` *(new)* — rarity weight unit tests
- `design-docs/roadmap.md` (mark 4.32–4.34 ✅)
- `CLAUDE.md` (Phase 4 status block update on completion)
- `plans/_QUEUE.md` (move to Completed on finish)

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/`, `src/fov/`
- `src/dialogue/`, `src/dialogues/`
- `src/spells/`, `src/entities/varn_boss.go`, `src/entities/boss.go`
- Anything in `src/levels/` other than reading from existing room tags

## Acceptance criteria

- [ ] Three chest variants exist as a typed enum and render with distinct sprites: Wooden, Iron, Gold.
- [ ] `Locked bool` on Chest; locked chests require Iron Key in inventory to open. Key is consumed.
- [ ] Approaching a locked chest without a key shows hint `"[E] Open (Iron Key required)"`.
- [ ] Every room with `levels.TagTreasure` contains at least one chest after floor generation.
- [ ] Chest variant on a floor scales with depth: Wooden (1–2), Iron (3–5), Gold (6+). Configurable via constants in `entities/chest.go`.
- [ ] Loot rarity weights scale with floor depth per `design-docs/biome-system.md` "Rarity weights scale with floor depth" table.
- [ ] `cd src && go build ./...` passes.
- [ ] `cd src && go test ./...` passes — at minimum a new test in `src/items/loot_test.go` covering rarity weight calculation across floors 1, 5, and 10.

## Phases

### Phase 1: Chest variants and Locked field
- [ ] 1.1 Add `ChestVariant` enum (`ChestWooden`, `ChestIron`, `ChestGold`) to `src/entities/chest.go`.
- [ ] 1.2 Add `Variant ChestVariant` and `Locked bool` fields to `Chest` struct.
- [ ] 1.3 Map each variant to a sprite ID via a `chestVariantSprite()` helper. Use existing chest sprites from the spritesheet; if only one chest sprite exists, document that in **What was NOT changed** and tint by variant for now.
- [ ] 1.4 `cd src && go build ./...` passes.

### Phase 2: Iron Key + locked-chest interaction
- [ ] 2.1 Register `iron_key` template in `src/items/load.go` — `ItemConsumable`, stackable up to 5, no ability grant.
- [ ] 2.2 In `src/game/handlers.game.go` chest interaction path: if `chest.Locked`, check player inventory for `iron_key`, consume one on open; otherwise display the locked hint and do nothing.
- [ ] 2.3 In `src/game/npc.game.go` `updateChestHints` (or wherever the chest hint is composed), branch on `chest.Locked` + key presence to choose the hint string.
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Treasure-room chest spawning
- [ ] 3.1 In `src/game/hub.go` after room tagging during floor generation, walk rooms with `levels.TagTreasure` and spawn a chest near the room center via the existing chest spawn helper.
- [ ] 3.2 Variant chosen by floor depth (constants from Phase 1.1).
- [ ] 3.3 Locked-chance scales with floor: 0% on Wooden, 40% on Iron, 70% on Gold (constants).
- [ ] 3.4 `cd src && go build ./...` passes; manually verify with a few seeded runs.

### Phase 4: Floor-depth weighted rarity
- [ ] 4.1 Update `src/items/loot.go` rarity selection to apply per-floor weight modifiers from `design-docs/biome-system.md`:
      Common base 60% (-3%/floor), Uncommon 30% (+1%/floor), Rare 9% (+1.5%/floor), Legendary 1% (+0.5%/floor). Clamp to [0,100].
- [ ] 4.2 Add `src/items/loot_test.go` covering the weight function at floors 1, 5, 10.
- [ ] 4.3 `cd src && go build ./... && go test ./...` passes.

### Phase 5: Cleanup
- [ ] 5.1 In `design-docs/roadmap.md`, prefix tasks 4.32, 4.33, 4.34 with ✅.
- [ ] 5.2 Update `CLAUDE.md` Phase 4 status block: 4F → "shipped".
- [ ] 5.3 Move this plan to `plans/COMPLETED/4F-loot-refinement.md` and update `plans/_QUEUE.md`.
- [ ] 5.4 Update `README.md` "Remaining polish" checklist (remove the 4F line).

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-27 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Chest sprite count.** If the spritesheet only has one chest sprite, do we tint Iron green and Gold yellow, or commission a sprite first? Recommendation: tint for now, file a sprite request as a separate one-line `art-request` plan. Affects 1.3.
- **Multiple chests per treasure room.** Always one, or scale 1–3 with room area? Recommendation: always 1, keep the loot economy predictable in this pass; revisit later if treasure rooms feel thin. Affects 3.1.
