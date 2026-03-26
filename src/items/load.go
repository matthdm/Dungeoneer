package items

import (
	"dungeoneer/images"
	"encoding/json"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// sheetEntry represents a single icon entry in the reverse map JSON.
type sheetEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Pos  struct {
		Row int `json:"row"`
		Col int `json:"col"`
	} `json:"position"`
	Effects struct {
		Description   string `json:"description"`
		StatModifiers struct {
			Strength     int `json:"strength_mod"`
			Dexterity    int `json:"dexterity_mod"`
			Vitality     int `json:"vitality_mod"`
			Intelligence int `json:"intelligence_mod"`
			Luck         int `json:"luck_mod"`
		} `json:"stat_modifiers"`
		Effect *struct {
			Trigger      string `json:"trigger"`
			Type         string `json:"type"`
			MagnitudePct int    `json:"magnitude_pct"`
			ChancePct    int    `json:"chance_pct"`
		} `json:"effect"`
	} `json:"effects"`
}

// LoadItemSheet registers items from the provided sheet and mapping.
func LoadItemSheet(img *ebiten.Image, entries []sheetEntry) {
	for _, e := range entries {
		x0 := e.Pos.Col * 32
		y0 := e.Pos.Row * 32
		sub := img.SubImage(image.Rect(x0, y0, x0+32, y0+32)).(*ebiten.Image)
		scaled := ebiten.NewImage(64, 64)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2)
		scaled.DrawImage(sub, op)

		stats := map[string]int{}
		if v := e.Effects.StatModifiers.Strength; v != 0 {
			stats["Strength"] = v
		}
		if v := e.Effects.StatModifiers.Dexterity; v != 0 {
			stats["Dexterity"] = v
		}
		if v := e.Effects.StatModifiers.Vitality; v != 0 {
			stats["Vitality"] = v
		}
		if v := e.Effects.StatModifiers.Intelligence; v != 0 {
			stats["Intelligence"] = v
		}
		if v := e.Effects.StatModifiers.Luck; v != 0 {
			stats["Luck"] = v
		}
		var eff *ItemEffect
		if e.Effects.Effect != nil {
			eff = &ItemEffect{
				Trigger:      e.Effects.Effect.Trigger,
				Type:         e.Effects.Effect.Type,
				MagnitudePct: e.Effects.Effect.MagnitudePct,
				ChancePct:    e.Effects.Effect.ChancePct,
			}
		}
		equippable := len(stats) > 0 || eff != nil
		tmpl := &ItemTemplate{
			ID:          e.ID,
			Name:        e.Name,
			Type:        ItemMisc,
			Description: e.Effects.Description,
			Stackable:   false,
			MaxStack:    1,
			Usable:      false,
			Equippable:  equippable,
			Stats:       stats,
			Effect:      eff,
			Icon:        scaled,
			OnUse:       nil,
			OnEquip:     nil,
			OnUnequip:   nil,
		}
		RegisterItem(tmpl)
	}
}

// LoadDefaultItems loads the bundled item sheet and mapping, then applies
// ability overrides to starter/quest items.
func LoadDefaultItems() error {
	img, err := images.LoadEmbeddedImage(images.Item_subset_png)
	if err != nil {
		return err
	}
	var entries []sheetEntry
	if err := json.Unmarshal(images.Items_structured_effects_json, &entries); err != nil {
		return err
	}
	LoadItemSheet(img, entries)
	applyAbilityOverrides()
	return nil
}

// abilityOverride patches a registered item with ability-granting fields.
type abilityOverride struct {
	ID            string
	GrantsAbility string
	AbilitySlot   AbilitySlotType
	ItemType      ItemType // override generic Misc type
}

// applyAbilityOverrides patches known items with their ability grants.
// Called once after LoadItemSheet so the items already exist in the registry.
func applyAbilityOverrides() {
	overrides := []abilityOverride{
		// Knight starters
		{ID: "item_0_1", GrantsAbility: "slash_combo", AbilitySlot: AbilitySlotPrimary, ItemType: ItemWeapon},   // Iron Emblem → melee combo
		{ID: "item_0_60", GrantsAbility: "dash", AbilitySlot: AbilitySlotDash, ItemType: ItemArmor},              // Leather Boots → dash

		// Mage starters
		{ID: "item_2_44", GrantsAbility: "arcane_bolt", AbilitySlot: AbilitySlotPrimary, ItemType: ItemWeapon},   // Grey Wizard Hat → arcane bolt
		{ID: "item_0_2", GrantsAbility: "arcane_spray", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon},      // Arcane Emblem → arcane spray
		{ID: "item_0_9", GrantsAbility: "blink", AbilitySlot: AbilitySlotDash, ItemType: ItemArmor},               // Sapphire Amulet → blink
	}
	for _, o := range overrides {
		tmpl, ok := Registry[o.ID]
		if !ok {
			continue
		}
		tmpl.GrantsAbility = o.GrantsAbility
		tmpl.AbilitySlot = o.AbilitySlot
		if o.ItemType != "" {
			tmpl.Type = o.ItemType
		}
		tmpl.Equippable = true
	}
}
