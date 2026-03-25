# Ability-Item System

Abilities are not innate. The player starts each run with basic movement and melee attack only. All spells, dash, and grapple must be obtained through items found in the dungeon.

---

## Design Philosophy

The player earns their power. Every ability in the game is tied to an item or item set that must be discovered, looted, or earned through quests. This creates:

1. **Run identity** — each run feels different depending on what you find
2. **Meaningful loot** — finding a Grimoire of Fire isn't just +stats, it unlocks Fireball
3. **Quest motivation** — the best ability items are quest-locked behind NPC relationships
4. **Build diversity** — players develop preferences and seek specific loadouts
5. **Progression clarity** — new abilities are always tied to a concrete in-world object

---

## Baseline Player (No Items)

At the start of every run, the player has only:

| Ability | Source | Always Available |
|---|---|---|
| Movement (WASD) | Innate | Yes |
| Melee Attack (Left Click) | Innate | Yes |
| Interact (E) | Innate | Yes |
| Inventory (Tab) | Innate | Yes |

Everything else requires an item.

---

## Ability Items

### Ability Slots

The player has **6 ability slots** (keys 1-6) that can be filled by equipping ability-granting items. The slots are universal — any ability item fills the next available slot.

Additionally, two **utility ability slots** exist:
- **Dash** (Shift) — requires a dash-granting item
- **Grapple** (F) — requires a grapple-granting item

### Item → Ability Mapping

Each ability-granting item specifies which ability it provides:

```go
type ItemTemplate struct {
    // ... existing fields ...
    GrantsAbility  string   // ability ID: "fireball", "chaos_ray", "dash", etc.
    AbilitySlot    string   // "spell" (fills 1-6), "dash", "grapple"
    SetID          string   // item set membership (empty = no set)
}
```

When an ability item is equipped:
1. Check `AbilitySlot` to determine slot type
2. If "spell" — add to next available spell slot (1-6)
3. If "dash" — enable dash ability
4. If "grapple" — enable grapple ability
5. `OnEquip` callback registers the ability with the player

When unequipped:
1. Remove ability from the corresponding slot
2. Shift remaining spell slots down (no gaps)
3. `OnUnequip` callback deregisters the ability

### Ability Source Items

#### Spell-Granting Items

| Item | Ability Granted | Acquisition | Rarity |
|---|---|---|---|
| **Grimoire of Fire** | Fireball | Drop / Chest | Uncommon |
| **Wand of Chaos** | Chaos Ray | Drop / Chest | Uncommon |
| **Storm Rod** | Lightning Strike | Drop / Chest | Uncommon |
| **Tempest Tome** | Lightning Storm | Quest reward (Forgotten Soldier) | Rare |
| **Fractal Seed** | Fractal Bloom | Quest reward (Mad Scholar) | Rare |
| **Living Branch** | Fractal Canopy (heal) | Quest reward (Hollow Monk) | Rare |

#### Utility-Granting Items

| Item | Ability Granted | Acquisition | Rarity |
|---|---|---|---|
| **Windrunner Boots** | Dash (3 charges) | Drop / Chest | Uncommon |
| **Shadowstep Cloak** | Dash (5 charges, faster recharge) | Quest reward | Rare |
| **Iron Grapple** | Grapple (8 tiles) | Drop / Chest | Uncommon |
| **Voidhook** | Grapple (12 tiles, pulls enemies) | Quest reward (Weeping Shade) | Rare |

#### Stat-Only Items (No Abilities)

These provide stat bonuses, buffs, and passive effects but do NOT grant abilities. They are the majority of loot:

| Category | Examples | Acquisition |
|---|---|---|
| Weapons | Swords, axes, bows — damage stats | Drop / Chest / Shop |
| Armor | Helmets, plate, robes — defense stats | Drop / Chest / Shop |
| Rings | +Luck, +Vitality, minor procs | Drop / Chest |
| Consumables | Potions, scrolls, keys | Drop / Chest / Shop |

---

## Item Sets

An item set is a group of 2-4 items that provide bonus effects when multiple pieces are equipped simultaneously.

### Set Structure

```go
type ItemSet struct {
    ID      string
    Name    string
    Items   []string          // item template IDs in this set
    Bonuses []SetBonus
}

type SetBonus struct {
    PiecesRequired int       // 2, 3, or 4
    Description    string
    StatBonuses    map[string]int
    GrantsAbility  string    // bonus ability at this tier (optional)
    EffectID       string    // special passive effect (optional)
}
```

### Example Sets

**Stormcaller Set** (3 pieces)
| Piece | Item | Slot |
|---|---|---|
| 1 | Storm Rod | Weapon (grants Lightning Strike) |
| 2 | Stormweave Robe | Chest |
| 3 | Conduit Ring | Ring |

| Pieces | Bonus |
|---|---|
| 2/3 | +20% Lightning damage, -0.5s Lightning cooldown |
| 3/3 | Grants **Lightning Storm** as bonus ability (no extra item needed) |

**Fractalist Set** (2 pieces)
| Piece | Item | Slot |
|---|---|---|
| 1 | Fractal Seed | Weapon (grants Fractal Bloom) |
| 2 | Living Branch | Offhand (grants Fractal Canopy) |

| Pieces | Bonus |
|---|---|
| 2/2 | Fractal Bloom explosions also heal player for 10% of damage dealt |

**Chainbreaker Set** (3 pieces — quest-locked from Varn's arc)
| Piece | Item | Slot |
|---|---|---|
| 1 | Varn's Broken Chain | Weapon |
| 2 | Shackle Guard | Offhand |
| 3 | Iron Will Amulet | Ring |

| Pieces | Bonus |
|---|---|
| 2/3 | +3 Strength, +15 MaxHP |
| 3/3 | Grants **Chain Lash** ability (melee AoE in front arc) |

---

## Quest-Locked vs. World-Dropped

### Quest-Locked Items

The most powerful ability items come from NPC quest completions. These use the existing dialogue action system:

```json
{
    "type": "give_item",
    "item_id": "tempest_tome",
    "amount": 1
}
```

Quest-locked items are:
- **Always Rare or Legendary quality**
- **Cannot drop from enemies or chests**
- **Tied to specific NPC quest phases** (completing a phase = reward)
- **May be part of an item set** whose other pieces drop in the world

This creates a compelling loop: find a Stormweave Robe (world drop) → realize it's part of the Stormcaller set → seek out the NPC who gives the Storm Rod → complete their quest → unlock the full set bonus.

### World-Dropped Items

Stat items, consumables, basic ability items (Grimoire of Fire, Windrunner Boots, Iron Grapple):
- Drop from enemies, chests, and loot rooms
- Quality tier determined by floor depth and room tag
- Available in hub shop (basic versions only, purchased with Remnants)

---

## Mana System Integration

With abilities now item-gated, mana becomes meaningful:

| Spell | Mana Cost | Notes |
|---|---|---|
| Fireball | 8 | Low cost, bread-and-butter |
| Chaos Ray | 12 | Medium cost, instant damage |
| Lightning Strike | 6 | Low cost, fast cooldown |
| Lightning Storm | 25 | High cost, area denial |
| Fractal Bloom | 20 | High cost, cascading damage |
| Fractal Canopy | 15 | Medium cost, healing |
| Dash | 0 | No mana cost (charge-based) |
| Grapple | 0 | No mana cost (cooldown-based) |

Mana regenerates passively (rate scales with Intelligence stat). Mana Elixirs restore mana instantly.

---

## Spell Bar UI Changes

The HUD spell bar currently shows all 6 spells. Changes needed:

1. **Empty by default** — slots show as dark/locked
2. **Populate on equip** — when an ability item is equipped, its icon appears in the next slot
3. **Show cooldown overlay** — existing cooldown display works, just applied per-slot
4. **Show mana cost** — small number below each spell icon
5. **Gray out if insufficient mana** — visual feedback when player can't cast
6. **Set bonus indicator** — subtle glow or border when set bonus is active

---

## Upgrade Station Interaction

Permanent upgrades (purchased with Remnants between runs) can enhance the ability system:

| Upgrade | Effect |
|---|---|
| Spell Affinity I-III | -10% mana cost per level |
| Mana Well I-III | +15 max mana per level |
| Quick Draw I-II | -0.3s cooldown reduction on all spells |
| Pack Rat I-II | +1 spell slot (beyond base 6) |
| Innate Dash I | Start runs with basic dash (no item needed) |

The "Innate Dash" upgrade is a deliberate late-game convenience — after enough runs, the player earns the right to start with dash built-in.

---

## Implementation Priority

### Phase 4A: Core Wiring (Must Have)

1. Add `GrantsAbility`, `AbilitySlot`, `SetID` fields to `ItemTemplate`
2. Add `LearnedAbilities []string` to Player — populated from equipped items
3. Gate spell casting: handler checks `LearnedAbilities` before allowing cast
4. Gate dash: check if dash ability is learned
5. Gate grapple: check if grapple ability is learned
6. Wire `OnEquip`/`OnUnequip` to add/remove abilities
7. Implement mana costs on spell cast
8. Update spell bar HUD to show only learned spells

### Phase 4B: Ability Items (Must Have)

9. Create ability-granting item templates (6 spell items + 2 dash + 2 grapple)
10. Add ability items to loot tables (Uncommon+ tier)
11. Ensure at least 1 spell item drops on Floor 1 (guaranteed early ability)
12. Add basic ability items to hub shop (Grimoire of Fire, Iron Grapple, Windrunner Boots)

### Phase 4C: Item Sets (Phase 5+)

13. Implement `ItemSet` registry and set bonus tracking
14. Create 3-4 item sets with 2-3 piece bonuses
15. Set bonus UI indicator on equipment panel
16. Quest-locked set pieces tied to NPC rewards

### Phase 4D: Quest Item Rewards (Phase 5)

17. Create rare ability items as quest rewards
18. Wire `give_item` dialogue action to inventory
19. NPC dialogue hints at item existence ("I could teach you the storm's fury...")

---

## Migration Path

Since spells currently work, the transition should be gradual:

1. **First**: Add the gating check but give the player a "Starter Grimoire" item in inventory that grants Fireball — so the player isn't empty-handed on Floor 1
2. **Then**: Remove free abilities one at a time as corresponding items are added to loot tables
3. **Finally**: Full item-gated system where baseline player has only melee

For development/testing, a debug key can temporarily grant all abilities.
