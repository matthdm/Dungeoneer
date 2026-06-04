package items

import "github.com/hajimehoshi/ebiten/v2"

// ItemType categorizes items for basic behavior.
type ItemType string

const (
	ItemWeapon     ItemType = "Weapon"
	ItemArmor      ItemType = "Armor"
	ItemConsumable ItemType = "Consumable"
	ItemKey        ItemType = "Key"
	ItemQuest      ItemType = "Quest"
	ItemMisc       ItemType = "Misc"
)

// AbilitySlotType determines where a granted ability is placed.
type AbilitySlotType string

const (
	AbilitySlotSpell   AbilitySlotType = "spell"   // fills spell bar (keys 1-6)
	AbilitySlotDash    AbilitySlotType = "dash"     // enables Shift dash
	AbilitySlotGrapple AbilitySlotType = "grapple"  // enables F grapple
	AbilitySlotPrimary AbilitySlotType = "primary"  // replaces left-click attack
)

// ItemTemplate defines common data shared across item instances.
type ItemTemplate struct {
	ID          string
	Name        string
	Type        ItemType
	Description string
	Stackable   bool
	MaxStack    int
	Usable      bool
	Equippable  bool
	Stats       map[string]int
	Effect      *ItemEffect
	Icon        *ebiten.Image
	OnUse       func(p interface{})
	OnEquip     func(p interface{})
	OnUnequip   func(p interface{})

	// Ability gating: equipping this item grants an ability.
	GrantsAbility string          // ability ID: "fireball", "chaos_ray", "dash", etc.
	AbilitySlot   AbilitySlotType // where the ability goes: "spell", "dash", "grapple"
	SetID         string          // item set membership (empty = no set)
	QuestLocked   bool            // true = excluded from loot tables, quest reward only
	Quality       string          // "common", "uncommon", "rare", "legendary" — controls border/title color
}

// Item represents an inventory instance.
type Item struct {
	*ItemTemplate
	Count int
}

// ItemEffect describes a special effect an item grants when equipped.
// Fields map directly to items_structured_effects.json.
type ItemEffect struct {
	// Trigger determines when the effect fires.
	// Values: "passive", "on_kill", "on_hit", "on_low_hp", "on_block",
	//         "on_potion_use", "regen_hp"
	Trigger string

	// Type identifies the kind of effect.
	// Values: "damage_reduction", "all_resistance", "lifesteal",
	//         "crit_multiplier", "gold_find", "mana_cost_reduction",
	//         "cooldown_reduction", "regen_hp", "heal", "bonus_healing",
	//         "counterpulse", "damage_reduction_buff"
	Type string

	// MagnitudePct is the primary percentage magnitude (0–100+).
	MagnitudePct int

	// ChancePct is the probability the effect fires on a trigger (0–100).
	// 0 means always fires when triggered.
	ChancePct int

	// MagnitudeFlat is a raw HP/mana value for flat effects (e.g. heal 10 HP).
	MagnitudeFlat int

	// ThresholdPct is an HP-percentage threshold for conditional triggers
	// such as "on_low_hp" (e.g. 20 = fires when HP < 20%).
	ThresholdPct int

	// CooldownSec is the minimum seconds between activations for triggered effects.
	CooldownSec float64

	// DurationSec is how long a temporary buff/debuff lasts.
	DurationSec float64

	// IntervalSec is the tick interval for periodic effects (e.g. regen_hp every 10s).
	IntervalSec float64

	// Element is an optional element tag (e.g. "physical", "fire").
	Element string

	// --- runtime-only fields (not serialised) ---

	// activeCooldown counts down (seconds) before this effect can fire again.
	activeCooldown float64

	// intervalAcc accumulates dt for periodic regen ticks.
	intervalAcc float64
}

// TickCooldown advances the active cooldown timer by dt seconds.
func (e *ItemEffect) TickCooldown(dt float64) {
	if e.activeCooldown > 0 {
		e.activeCooldown -= dt
		if e.activeCooldown < 0 {
			e.activeCooldown = 0
		}
	}
}

// IsReady returns true when the effect is not on cooldown.
func (e *ItemEffect) IsReady() bool { return e.activeCooldown <= 0 }

// PutOnCooldown resets the activation cooldown.
func (e *ItemEffect) PutOnCooldown() {
	if e.CooldownSec > 0 {
		e.activeCooldown = e.CooldownSec
	}
}

// ItemSave is a minimal representation for serialization.
type ItemSave struct {
	ID    string
	Count int
}

// ToSave converts an item instance to its save form.
func (i *Item) ToSave() ItemSave {
	return ItemSave{ID: i.ID, Count: i.Count}
}

// FromSave recreates an item from saved data.
func FromSave(data ItemSave) *Item {
	it := NewItem(data.ID)
	it.Count = data.Count
	return it
}

// SerializeEquipment converts an equipment map into savable data.
func SerializeEquipment(eq map[string]*Item) map[string]ItemSave {
	res := make(map[string]ItemSave)
	for slot, it := range eq {
		if it != nil {
			res[slot] = it.ToSave()
		}
	}
	return res
}

// DeserializeEquipment reconstructs an equipment map from saved data.
func DeserializeEquipment(data map[string]ItemSave) map[string]*Item {
	eq := make(map[string]*Item)
	for slot, sv := range data {
		if sv.ID != "" {
			eq[slot] = FromSave(sv)
		}
	}
	return eq
}
