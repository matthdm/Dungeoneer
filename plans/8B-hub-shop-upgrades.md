---
plan-id: 8B-hub-shop-upgrades
status: queued
owner: unassigned
branch: plan/8B-hub-shop-upgrades
depends-on: [8A-item-sets]
last-touched: 2026-04-30
---

# Plan: Phase 8B — Hub Shop & Upgrade Station

## Goal

Add the two permanent hub meta-economy stations: a shop where players spend Remnants to buy consumables and starter ability items, and an upgrade station where players spend Remnants to permanently improve their runs (HP, damage, mana, inventory size, etc.). Both are milestone-gated via `HubState` (shop unlocks after run 1, upgrades after run 3).

When done, Remnants have meaningful spend options and players feel progress between runs even when they die early.

## Scope

**In scope:**
- Shop NPC entity in hub with a scrollable item list UI, Remnant balance display, stock limits per run.
- 8 permanent upgrades as defined in the roadmap (Iron Constitution through Innate Dash).
- Upgrade station NPC entity in hub with category tabs and level pip UI.
- Both entities respect `HubState` milestone gates.
- Purchased items appear in player inventory at run start. Upgrades apply via `Apply()` before dungeon entry.

**Out of scope (do not change in this plan):**
- Item sets (plan `8A-item-sets.md`).
- New ability items.
- Shop restocking system (restock is per-run; stock is fixed per run, not dynamic).

## File envelope

**Touched:**
- `src/game/shop_data.go` *(new)* — shop inventory definitions, per-run stock
- `src/ui/shop.go` *(new)* — shop UI
- `src/game/upgrades.go` *(new)* — `UpgradeDef`, `Apply()`, 8 starter upgrades
- `src/ui/upgrade_station.go` *(new)* — upgrade station UI
- `src/game/hub.go` — spawn shop and upgrade station NPCs when HubState flags are set; apply upgrades at run start; inject purchased items into player inventory
- `src/game/npc_data.go` — shop and upgrade station NPC definitions
- `src/game/metasave.go` — add `Upgrades map[string]int` and `ShopPurchases []string` (per-run resets)
- `design-docs/roadmap.md` — mark 8.5–8.14 ✅ on completion
- `CLAUDE.md` — update Phase 8 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/items/sets.go` — set plan scope
- `src/dialogue/` — shop/upgrade use their own UI, not the dialogue engine

## Acceptance criteria

- [ ] Shop NPC appears in hub when `HubState["shop_unlocked"]`. Opening shows a scrollable item list with Remnant prices. Buying deducts Remnants and adds item to `ShopPurchases`.
- [ ] Shop stocks 3–5 consumables and 1–2 basic ability items per run. Stock resets each run.
- [ ] Purchased shop items appear in player inventory when the run begins.
- [ ] Upgrade Station NPC appears when `HubState["upgrades_unlocked"]`. Shows 8 upgrades with current level, max level, cost, and description.
- [ ] All 8 upgrades are implemented: Iron Constitution (+15 MaxHP/level, max 5), Sharpened Edge (+5% damage/level, max 5), Deep Pockets (+1 inventory row/level, max 3), Spell Affinity (-5% mana cost/level, max 5), Mana Well (+10 MaxMana/level, max 5), Quick Draw (-0.1s cooldown/level, max 5), Scavenger (+2 Remnants/kill/level, max 5), Innate Dash (1 level only — grants dash without requiring item).
- [ ] Upgrades persist in MetaSave and are applied to the player at run start.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Upgrade system
- [ ] 1.1 Create `src/game/upgrades.go`. `UpgradeDef` struct: ID, Name, Description, CostPerLevel `[]int`, MaxLevel, `Apply(player *Player, level int)` function.
- [ ] 1.2 Define all 8 starter upgrades with their `Apply()` implementations (modify `player.MaxHP`, `player.DamageBonus`, etc.).
- [ ] 1.3 In `src/game/metasave.go`, add `Upgrades map[string]int`.
- [ ] 1.4 In `src/game/hub.go` run-start path, iterate `MetaSave.Upgrades`, call `Apply()` for each at its current level.
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: Upgrade Station UI
- [ ] 2.1 Create `src/ui/upgrade_station.go`. `UpgradeStation` struct with upgrade list, selected upgrade, Remnant balance.
- [ ] 2.2 Draw: category list (or simple scrollable list for 8 items), selected upgrade detail panel (name, description, level pips, cost, upgrade button).
- [ ] 2.3 Purchase: check `MetaSave.Remnants >= cost`, deduct, increment `MetaSave.Upgrades[id]`, save MetaSave.
- [ ] 2.4 In `src/game/hub.go`, spawn Upgrade Station NPC when `HubState["upgrades_unlocked"]`.
- [ ] 2.5 `cd src && go build ./...` passes.

### Phase 3: Shop data and UI
- [ ] 3.1 Create `src/game/shop_data.go`. Define `ShopItem` struct and the per-run stock list (item ID + Remnant price + max stock).
- [ ] 3.2 Create `src/ui/shop.go`. `ShopUI` struct: stock list, Remnant balance, purchase confirmation.
- [ ] 3.3 Draw: scrollable item list with name, description, price, stock remaining. Buy button. Remnant balance in corner.
- [ ] 3.4 Purchase: deduct Remnants, add item ID to `MetaSave.ShopPurchases`, update stock.
- [ ] 3.5 In `src/game/hub.go` run-start path, move `ShopPurchases` into player inventory and clear the list.
- [ ] 3.6 In `src/game/hub.go`, spawn Shop NPC when `HubState["shop_unlocked"]`.
- [ ] 3.7 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark roadmap rows 8.5–8.14 ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 8 status block to "🟢 COMPLETE".
- [ ] 4.3 Move this plan to `plans/COMPLETED/8B-hub-shop-upgrades.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Innate Dash upgrade vs. item-gated dash.** If the player has the Innate Dash upgrade, they can dash without equipping the Windrunner Boots. This means `HasAbility("dash")` must check both `LearnedAbilities` (from items) and `MetaSave.Upgrades["innate_dash"] > 0`. Affects 1.2.
- **Shop NPC sprite.** Use an existing NPC sprite until a dedicated merchant sprite is created. Document the gap. Affects 3.6.
