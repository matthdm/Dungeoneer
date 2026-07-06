package items

import (
	"dungeoneer/images"
	"encoding/json"
	"image"
	"image/color"

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
			Trigger       string  `json:"trigger"`
			Type          string  `json:"type"`
			MagnitudePct  int     `json:"magnitude_pct"`
			ChancePct     int     `json:"chance_pct"`
			MagnitudeFlat int     `json:"magnitude_flat"`
			ThresholdPct  int     `json:"threshold_pct"`
			CooldownSec   float64 `json:"cooldown_s"`
			DurationSec   float64 `json:"duration_s"`
			IntervalSec   float64 `json:"interval_s"`
			Element       string  `json:"element"`
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
				Trigger:       e.Effects.Effect.Trigger,
				Type:          e.Effects.Effect.Type,
				MagnitudePct:  e.Effects.Effect.MagnitudePct,
				ChancePct:     e.Effects.Effect.ChancePct,
				MagnitudeFlat: e.Effects.Effect.MagnitudeFlat,
				ThresholdPct:  e.Effects.Effect.ThresholdPct,
				CooldownSec:   e.Effects.Effect.CooldownSec,
				DurationSec:   e.Effects.Effect.DurationSec,
				IntervalSec:   e.Effects.Effect.IntervalSec,
				Element:       e.Effects.Effect.Element,
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
	registerKeyItems()
	registerNewArtifacts()
	return nil
}

// registerKeyItems registers dungeon key items that have no icon sprite.
func registerKeyItems() {
	RegisterItem(&ItemTemplate{
		ID:          "iron_key",
		Name:        "Iron Key",
		Type:        ItemConsumable,
		Description: "Opens locked iron and gold chests.",
		Stackable:   true,
		MaxStack:    5,
		Usable:      false,
		Equippable:  false,
		QuestLocked: false,
	})
}

// abilityOverride patches a registered item with ability-granting fields.
type abilityOverride struct {
	ID             string
	GrantsAbility  string
	AbilitySlot    AbilitySlotType
	ItemType       ItemType // override generic Misc type
	QuestLocked    bool     // true = class starter, excluded from random loot
	Quality        string   // "common", "uncommon", "rare", "legendary"
	IsArtifact     bool
	ArtifactDomain string
	IsElite        bool
}

// applyAbilityOverrides patches known items with their ability grants.
// Called once after LoadItemSheet so the items already exist in the registry.
func applyAbilityOverrides() {
	overrides := []abilityOverride{
		// Knight starters — QuestLocked, Uncommon: class-defining gear given at run start.
		{ID: "item_0_1", GrantsAbility: "slash_combo", AbilitySlot: AbilitySlotPrimary, ItemType: ItemWeapon, QuestLocked: true, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "iron"},   // Iron Emblem → melee combo
		{ID: "item_0_60", GrantsAbility: "dash", AbilitySlot: AbilitySlotDash, ItemType: ItemArmor, QuestLocked: true, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "shadow"},            // Leather Boots → dash

		// Mage starters — QuestLocked, Uncommon: class-defining gear given at run start.
		{ID: "item_2_44", GrantsAbility: "arcane_bolt", AbilitySlot: AbilitySlotPrimary, ItemType: ItemWeapon, QuestLocked: true, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "arcane"}, // Grey Wizard Hat → arcane bolt
		{ID: "item_0_2", GrantsAbility: "arcane_spray", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, QuestLocked: true, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "arcane"},   // Arcane Emblem → arcane spray
		{ID: "item_0_9", GrantsAbility: "blink", AbilitySlot: AbilitySlotDash, ItemType: ItemArmor, QuestLocked: true, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "shadow"},            // Sapphire Amulet → blink

		// Droppable ability items — Uncommon unless otherwise noted.
		{ID: "item_2_24", GrantsAbility: "fireball", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "flame"},                            // Fireball Emblem → fireball
		{ID: "item_0_3", GrantsAbility: "chaos_ray", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, Quality: RarityRare, IsArtifact: true, ArtifactDomain: "void", IsElite: true},                  // Chaos Emblem → chaos ray (elite)
		{ID: "item_0_26", GrantsAbility: "lightning", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "arcane"},                          // Rage Emblem → lightning
		{ID: "item_0_35", GrantsAbility: "lightning_storm", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, Quality: RarityRare, IsArtifact: true, ArtifactDomain: "arcane", IsElite: true},         // Azazel's Pentagram → lightning storm (Rare, elite)
		{ID: "item_2_63", GrantsAbility: "fractal_bloom", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "nature"},                      // Verdant Tome → fractal bloom
		{ID: "item_2_55", GrantsAbility: "fractal_canopy", AbilitySlot: AbilitySlotSpell, ItemType: ItemWeapon, Quality: RarityRare, IsArtifact: true, ArtifactDomain: "nature", IsElite: true},          // Necromancer's Tome → fractal canopy (Rare, elite)
		{ID: "item_0_63", GrantsAbility: "dash", AbilitySlot: AbilitySlotDash, ItemType: ItemArmor, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "shadow"},                                 // Boots of Speed → dash (cross-class)
		{ID: "item_2_35", GrantsAbility: "blink", AbilitySlot: AbilitySlotDash, ItemType: ItemArmor, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "shadow"},                                // Haste Carriers → blink (cross-class)
		{ID: "item_1_12", GrantsAbility: "grapple", AbilitySlot: AbilitySlotGrapple, ItemType: ItemMisc, Quality: RarityUncommon, IsArtifact: true, ArtifactDomain: "iron"},                              // Grips of the Buried Flame → grapple
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
		tmpl.QuestLocked = o.QuestLocked
		if o.Quality != "" {
			tmpl.Quality = o.Quality
		}
		tmpl.IsArtifact = o.IsArtifact
		tmpl.ArtifactDomain = o.ArtifactDomain
		tmpl.IsElite = o.IsElite
	}
}

// artifactIcon generates a 16×16 placeholder icon for a combat artifact.
// Each domain gets a distinct solid colour so items are visually distinguishable
// in-world until real art ships.
func artifactIcon(domain string) *ebiten.Image {
	var fill color.NRGBA
	switch domain {
	case "iron":
		fill = color.NRGBA{R: 160, G: 185, B: 210, A: 255}
	case "shadow":
		fill = color.NRGBA{R: 110, G: 60, B: 180, A: 255}
	case "flame":
		fill = color.NRGBA{R: 220, G: 90, B: 30, A: 255}
	case "void":
		fill = color.NRGBA{R: 60, G: 20, B: 120, A: 255}
	case "nature":
		fill = color.NRGBA{R: 50, G: 170, B: 80, A: 255}
	case "arcane":
		fill = color.NRGBA{R: 60, G: 160, B: 230, A: 255}
	default:
		fill = color.NRGBA{R: 160, G: 140, B: 100, A: 255}
	}
	border := color.NRGBA{R: fill.R / 3, G: fill.G / 3, B: fill.B / 3, A: 255}

	img := ebiten.NewImage(16, 16)
	img.Fill(fill)
	for i := range 16 {
		img.Set(i, 0, border)
		img.Set(i, 15, border)
		img.Set(0, i, border)
		img.Set(15, i, border)
	}
	return img
}

// registerNewArtifacts registers all artifacts not sourced from the item sheet:
//   - Wave 1: 6 melee artifacts (combat redesign)
//   - Wave 2: 7 meta-build skill artifacts (the 55 tank, perma-shadow, surge nuke, sacrifice)
//   - Build enablers: 10 stat-modifier items (the +20%-enchant-duration layer)
//
// Every item here carries at least one stat bonus in addition to any skill grant.
// Effect logic lives in src/combat/skills.go.
func registerNewArtifacts() {
	// ── Wave 1: melee skill artifacts ─────────────────────────────────────
	RegisterItem(&ItemTemplate{
		ID:             "ironbreaker_gauntlets",
		Name:           "Ironbreaker Gauntlets",
		Type:           ItemWeapon,
		Description:    "Slam the ground, sending a shockwave through nearby enemies. +2 Strength.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"Strength": 2},
		GrantsAbility:  "ironbreaker_slam",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "iron",
	})
	RegisterItem(&ItemTemplate{
		ID:             "shroud_cloak",
		Name:           "Shroud Cloak",
		Type:           ItemArmor,
		Description:    "Blink behind your target. Your next strike is a guaranteed critical. +3 Dexterity.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"Dexterity": 3},
		GrantsAbility:  "shadowstep",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "shadow",
	})
	RegisterItem(&ItemTemplate{
		ID:             "wardens_medallion",
		Name:           "Warden's Medallion",
		Type:           ItemArmor,
		Description:    "Taunt all nearby enemies. Reduces incoming damage by 30% while active. +8 Max HP.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxHP": 8},
		GrantsAbility:  "warden_taunt",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "nature",
	})
	RegisterItem(&ItemTemplate{
		ID:             "ashbound_chain",
		Name:           "Ashbound Chain",
		Type:           ItemWeapon,
		Description:    "Bind your target in place for 2 seconds. +4 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 4},
		GrantsAbility:  "void_bind",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "void",
	})
	RegisterItem(&ItemTemplate{
		ID:             "grave_reaper",
		Name:           "Grave Reaper",
		Type:           ItemWeapon,
		Description:    "Execute an enemy below 20% HP. Elite. +6 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 6},
		GrantsAbility:  "execute",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityLegendary,
		IsArtifact:     true,
		ArtifactDomain: "void",
		IsElite:        true,
	})
	RegisterItem(&ItemTemplate{
		ID:             "ember_mantle",
		Name:           "Ember Mantle",
		Type:           ItemArmor,
		Description:    "Auto-attacks apply burn (5 dmg/s for 3s). Passive. +1 Intelligence.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"Intelligence": 1},
		GrantsAbility:  "ember_burn",
		AbilitySlot:    AbilitySlotPrimary,
		Quality:        RarityUncommon,
		IsArtifact:     true,
		ArtifactDomain: "flame",
	})

	// ── Wave 2: meta-build skill artifacts ────────────────────────────────

	// "The 55" tank core — incoming damage capped at 12% max HP
	RegisterItem(&ItemTemplate{
		ID:             "stone_skin_idol",
		Name:           "Stone Skin Idol",
		Type:           ItemArmor,
		Description:    "Passive: incoming damage cannot exceed 12% of your maximum HP. +8 Max HP.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxHP": 8},
		GrantsAbility:  "stone_skin",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "nature",
	})
	// Perma-Shadow enabler — hitting during shadow reduces Shroud Cloak CD by 3s
	RegisterItem(&ItemTemplate{
		ID:             "shadows_return",
		Name:           "Shadow's Return",
		Type:           ItemArmor,
		Description:    "Passive: striking an enemy while in shadow reduces Shroud Cloak's cooldown by 3s. +4 Dexterity.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"Dexterity": 4},
		GrantsAbility:  "shadows_return_passive",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "shadow",
	})
	// Duration extender — +20% to all skill durations
	RegisterItem(&ItemTemplate{
		ID:             "void_mirror_pendant",
		Name:           "Void Mirror Pendant",
		Type:           ItemArmor,
		Description:    "Passive: all skill durations increased by 20%. +6 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 6, "SkillDuration": 20},
		GrantsAbility:  "void_mirror_passive",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "void",
	})
	// Surge Nuke elite — 12 damage per artifact currently on cooldown
	RegisterItem(&ItemTemplate{
		ID:             "arcane_surge",
		Name:           "Arcane Surge",
		Type:           ItemWeapon,
		Description:    "Active: deal 12 × (artifacts on cooldown) arcane damage to nearby enemies. Elite. +10 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 10},
		GrantsAbility:  "arcane_surge_blast",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityLegendary,
		IsArtifact:     true,
		ArtifactDomain: "arcane",
		IsElite:        true,
	})
	// CDR ring — passive -10% to all active cooldowns
	RegisterItem(&ItemTemplate{
		ID:             "arcane_tempo_ring",
		Name:           "Arcane Tempo Ring",
		Type:           ItemMisc,
		Description:    "Passive: all active skill cooldowns reduced by 10%. +4 Intelligence.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"Intelligence": 4, "CooldownReduction": 10},
		GrantsAbility:  "arcane_tempo_passive",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "arcane",
	})
	// Sacrifice nuke — spend 25% HP to deal 200 void damage
	RegisterItem(&ItemTemplate{
		ID:             "blood_price",
		Name:           "Blood Price",
		Type:           ItemWeapon,
		Description:    "Active: sacrifice 25% of your current HP to deal 200 void damage. +6 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 6},
		GrantsAbility:  "blood_price_strike",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "void",
	})
	// Sacrifice sustain — heal 20% max HP on kill
	RegisterItem(&ItemTemplate{
		ID:             "soul_harvest",
		Name:           "Soul Harvest",
		Type:           ItemWeapon,
		Description:    "Passive: each kill restores 20% of your maximum HP. +8 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 8},
		GrantsAbility:  "soul_harvest_passive",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "void",
	})

	// ── Build enablers: stat-modifier items ───────────────────────────────
	// These are the "+20% enchant duration" layer — no active skill, but their
	// stats are required to hit specific build breakpoints.

	// +20% skill duration: enables perma-uptime builds when stacked
	RegisterItem(&ItemTemplate{
		ID:          "voidweave_wraps",
		Name:        "Voidweave Wraps",
		Type:        ItemArmor,
		Description: "+12 Max HP. All skill durations increased by 20%.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"MaxHP": 12, "SkillDuration": 20},
		Quality:     RarityUncommon,
	})
	// -10% CDR: accelerates cooldown cycling
	RegisterItem(&ItemTemplate{
		ID:          "arcane_tempo_belt",
		Name:        "Arcane Tempo Belt",
		Type:        ItemArmor,
		Description: "+8 Max Mana. All active skill cooldowns reduced by 10%.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"MaxMana": 8, "CooldownReduction": 10},
		Quality:     RarityUncommon,
	})
	// Damage amp while low HP: enables sacrifice builds
	RegisterItem(&ItemTemplate{
		ID:          "blood_vow_amulet",
		Name:        "Blood Vow Amulet",
		Type:        ItemMisc,
		Description: "Drain 5 HP/s. Deal 20% more damage while below 50% HP.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"DamagePct": 20},
		Effect: &ItemEffect{
			Trigger:      "on_low_hp",
			Type:         "damage_bonus_pct",
			ThresholdPct: 50,
			MagnitudePct: 20,
			IntervalSec:  1.0,
			MagnitudeFlat: -5, // HP drain per second (negative = drain)
		},
		Quality: RarityRare,
	})
	// +INT + DoT amp: enables flame nuke builds
	RegisterItem(&ItemTemplate{
		ID:          "resonance_crystal",
		Name:        "Resonance Crystal",
		Type:        ItemMisc,
		Description: "+2 Intelligence. Spells deal 8% more damage per active DoT on the target.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"Intelligence": 2},
		Quality:     RarityUncommon,
	})
	// Kill-sustain: momentum + healing for clear builds
	RegisterItem(&ItemTemplate{
		ID:          "lifedrinker_robe",
		Name:        "Lifedrinker Robe",
		Type:        ItemArmor,
		Description: "+10 Max HP. Restore 6 HP on each kill.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"MaxHP": 10},
		Effect: &ItemEffect{
			Trigger:       "on_kill",
			Type:          "heal",
			MagnitudeFlat: 6,
		},
		Quality: RarityUncommon,
	})
	// STR ring with adrenaline charge potential
	RegisterItem(&ItemTemplate{
		ID:          "iron_will_band",
		Name:        "Iron Will Band",
		Type:        ItemMisc,
		Description: "+2 Strength. Auto-attacks build Adrenaline (max 5 charges; future: converts to bonus damage).",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"Strength": 2},
		Quality:     RarityUncommon,
	})
	// Intentional HP reduction ring for "The 55" build
	RegisterItem(&ItemTemplate{
		ID:          "marrow_ring",
		Name:        "Marrow Ring",
		Type:        ItemMisc,
		Description: "-15 Max HP. Deal 12% more damage. +1 Mana/s regen.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"MaxHP": -15, "ManaRegen": 1},
		Effect: &ItemEffect{
			Trigger:      "passive",
			Type:         "damage_bonus_pct",
			MagnitudePct: 12,
		},
		Quality: RarityRare,
	})
	// Mana item that routes void skill costs to HP
	RegisterItem(&ItemTemplate{
		ID:          "hollow_sigil",
		Name:        "Hollow Sigil",
		Type:        ItemMisc,
		Description: "+6 Max Mana. Void skill costs are paid with HP instead of mana.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"MaxMana": 6},
		Quality:     RarityRare,
	})
	// Tank vest with reflect damage
	RegisterItem(&ItemTemplate{
		ID:          "thornweave_vest",
		Name:        "Thornweave Vest",
		Type:        ItemArmor,
		Description: "+8 Max HP. Return 4 damage to any enemy that strikes you in melee.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"MaxHP": 8},
		Effect: &ItemEffect{
			Trigger:       "on_hit",
			Type:          "reflect_damage",
			MagnitudeFlat: 4,
		},
		Quality: RarityUncommon,
	})
	// DEX talisman with attack speed — general cleave/duelist item
	RegisterItem(&ItemTemplate{
		ID:          "quicksilver_talisman",
		Name:        "Quicksilver Talisman",
		Type:        ItemMisc,
		Description: "+3 Dexterity. Attack speed increased by 15%.",
		Stackable:   false,
		MaxStack:    1,
		Equippable:  true,
		Stats:       map[string]int{"Dexterity": 3, "AttackSpeed": 15},
		Quality:     RarityUncommon,
	})

	// Assign placeholder domain icons to all artifact items that have no icon yet.
	// These coloured squares are visible in-world until real art ships.
	for _, tmpl := range Registry {
		if tmpl.IsArtifact && tmpl.Icon == nil {
			tmpl.Icon = artifactIcon(tmpl.ArtifactDomain)
		}
	}
}
