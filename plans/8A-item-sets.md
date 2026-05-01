---
plan-id: 8A-item-sets
status: queued
owner: unassigned
branch: plan/8A-item-sets
depends-on: [7A-echoes-of-self, 7B-living-dungeon-ai]
last-touched: 2026-04-30
---

# Plan: Phase 8A — Item Sets & Set Bonuses

## Goal

Add item set mechanics: groups of 2–3 items that, when equipped together, grant a bonus stat effect or ability. Three starter sets are defined: Stormcaller (lightning focus), Fractalist (bloom+canopy synergy), and Chainbreaker (Varn quest-locked, melee AoE). The equipment panel shows set name, pieces owned, and active bonuses.

When done, players have a compelling reason to build around specific item combinations, adding depth to loot decisions.

## Scope

**In scope:**
- `ItemSet` struct: ID, member item IDs, `[]SetBonus` (pieces required, stat bonuses, bonus ability).
- Set bonus tracking: on equip/unequip, recalculate active set bonuses, apply to player stats.
- Set bonus UI in the equipment/inventory panel: set name, owned/required pip count, active bonus highlighted.
- Three starter sets with defined bonuses.

**Out of scope (do not change in this plan):**
- Hub shop (plan `8B-hub-shop-upgrades.md`).
- New ability items beyond what already exists.
- Crafting system (backlog B.5).

## File envelope

**Touched:**
- `src/items/sets.go` *(new)* — `ItemSet`, `SetBonus`, registry, active-set calculation
- `src/entities/player.go` — call `RecalculateSetBonuses()` on equip/unequip; apply set stat bonuses
- `src/ui/tooltip.go` — show set membership and active bonuses in item tooltip
- `src/hud/hud.go` — set bonus panel in equipment screen
- `design-docs/roadmap.md` — mark 8.1–8.4 ✅ on completion
- `CLAUDE.md` — update Phase 8 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/items/load.go` — only add set definitions in `sets.go`; do not restructure item loading
- `src/dialogue/` — not relevant

## Acceptance criteria

- [ ] `ItemSet` struct exists with ID, member item IDs, and `[]SetBonus` (each bonus has `PiecesRequired int`, `StatBonus StatModifier`, optional `BonusAbility string`).
- [ ] `RecalculateSetBonuses(equippedItems []Item) []ActiveSetBonus` returns the currently active set bonuses based on equipped items.
- [ ] Set bonuses are applied to player stats on equip/unequip (same path as existing `StatModifiers`).
- [ ] Stormcaller set (3pc): +20% lightning damage, bonus ability `lightning_aura`. Chainbreaker set (3pc, quest-locked): +30% melee damage, bonus ability `chain_sweep`. Fractalist set (2pc): fractal bloom summons an additional canopy fragment.
- [ ] Item tooltip shows set membership: "Part of Stormcaller Set (1/3)". Active bonuses shown in green.
- [ ] Equipment panel has a Set Bonuses section showing all sets the player has partial or full completion of.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Set data model and registry
- [ ] 1.1 Create `src/items/sets.go`. Define `SetBonus` struct and `ItemSet` struct.
- [ ] 1.2 Define `ActiveSetBonus` struct: set ID, bonus tier, applied stat modifier.
- [ ] 1.3 `RecalculateSetBonuses(equippedIDs []string) []ActiveSetBonus` — for each registered set, count owned pieces, return bonuses for met thresholds.
- [ ] 1.4 Register the three starter sets: Stormcaller, Fractalist, Chainbreaker. Mark Chainbreaker as `QuestLocked: true` (won't drop in loot tables; granted via Varn arc).
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: Player integration
- [ ] 2.1 In `src/entities/player.go` `RecalculateStats()`, also call `RecalculateSetBonuses()` and apply returned `StatModifier` bonuses on top of individual item bonuses.
- [ ] 2.2 Apply `BonusAbility` from active set bonuses to `LearnedAbilities` (alongside item-granted abilities).
- [ ] 2.3 `cd src && go build ./...` passes.

### Phase 3: UI — tooltip and equipment panel
- [ ] 3.1 In `src/ui/tooltip.go`, if the item belongs to a set, add a set line: "Part of [Set Name] ([owned]/[total])". List active bonuses in green if threshold met.
- [ ] 3.2 In `src/hud/hud.go` equipment screen, add a "Set Bonuses" section below the stats panel. Show each set the player has ≥1 piece of: name, pip row (filled/unfilled), active bonus text.
- [ ] 3.3 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark roadmap rows 8.1–8.4 ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 8 status block.
- [ ] 4.3 Move this plan to `plans/COMPLETED/8A-item-sets.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Stormcaller item IDs.** Confirm which existing lightning items form the Stormcaller set (Storm Rod + Tempest Tome + ?). Check `src/items/load.go` for the exact item IDs before hardcoding. Affects 1.4.
- **Set bonuses and ability slots.** If the player already has all spell slots filled and a set grants a bonus ability, where does it go? Recommendation: bonus set abilities go into an overflow slot and are always-active (no hotkey), distinct from equipped spell slots. Affects 2.2.
