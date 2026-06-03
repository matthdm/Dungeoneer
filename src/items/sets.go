package items

// SetStatBonus defines the stat gains for a set bonus tier.
// Keys match the stat map convention used by ItemTemplate.Stats:
// "Strength", "Dexterity", "Vitality", "Intelligence", "Luck".
type SetStatBonus map[string]int

// SetBonus defines a reward for equipping N pieces of a set.
type SetBonus struct {
	PiecesRequired int
	StatBonus      SetStatBonus // nil or empty = no stat bonus
	BonusAbility   string       // empty = no ability bonus
}

// ItemSet groups related items with tiered bonuses.
type ItemSet struct {
	ID          string
	Name        string
	MemberIDs   []string   // item IDs that belong to this set
	Bonuses     []SetBonus // sorted by PiecesRequired ascending
	QuestLocked bool       // if true, bonus only activates with Varn quest complete
}

// ActiveSetBonus describes a bonus currently in effect (or partially met).
type ActiveSetBonus struct {
	SetID       string
	SetName     string
	PiecesOwned int
	PiecesTotal int // total members in set
	Tier        SetBonus
}

// SetRegistry holds all registered item sets.
var SetRegistry []*ItemSet

func init() {
	SetRegistry = []*ItemSet{
		{
			// Stormcaller: lightning-themed spell weapons.
			// item_0_26 = Rage Emblem (grants lightning)
			// item_0_35 = Azazel's Pentagram (grants lightning_storm)
			ID:        "stormcaller",
			Name:      "Stormcaller",
			MemberIDs: []string{"item_0_26", "item_0_35"},
			Bonuses: []SetBonus{
				{PiecesRequired: 2, StatBonus: SetStatBonus{"Intelligence": 20}, BonusAbility: "lightning_aura"},
			},
		},
		{
			// Fractalist: nature/fractal tome weapons.
			// item_2_63 = Verdant Tome (grants fractal_bloom)
			// item_2_55 = Necromancer's Tome (grants fractal_canopy)
			ID:        "fractalist",
			Name:      "Fractalist",
			MemberIDs: []string{"item_2_63", "item_2_55"},
			Bonuses: []SetBonus{
				{PiecesRequired: 2, StatBonus: nil, BonusAbility: "fractal_echo"},
			},
		},
		{
			// Chainbreaker: melee/grapple focused gear.
			// item_1_12 = Grips of the Buried Flame (grants grapple)
			// item_0_3  = Chaos Emblem (grants chaos_ray)
			// item_0_1  = Iron Emblem (grants slash_combo) — closest confirmed melee item.
			// Note: item_0_15 was specified but not present in ability overrides;
			// item_0_1 (Iron Emblem / slash_combo) substituted as the melee anchor.
			ID:          "chainbreaker",
			Name:        "Chainbreaker",
			MemberIDs:   []string{"item_1_12", "item_0_3", "item_0_1"},
			QuestLocked: true,
			Bonuses: []SetBonus{
				{PiecesRequired: 2, StatBonus: SetStatBonus{"Strength": 15}, BonusAbility: ""},
				{PiecesRequired: 3, StatBonus: SetStatBonus{"Strength": 30}, BonusAbility: "chain_sweep"},
			},
		},
	}
}

// RecalculateSetBonuses returns the active set bonuses for the given equipped item IDs.
// questComplete should be true when Varn's questline has been completed (DefeatCount > 0).
func RecalculateSetBonuses(equippedIDs []string, questComplete bool) []ActiveSetBonus {
	equipped := make(map[string]bool, len(equippedIDs))
	for _, id := range equippedIDs {
		equipped[id] = true
	}

	var active []ActiveSetBonus
	for _, set := range SetRegistry {
		if set.QuestLocked && !questComplete {
			continue
		}
		owned := 0
		for _, mid := range set.MemberIDs {
			if equipped[mid] {
				owned++
			}
		}
		if owned == 0 {
			continue
		}
		// Find highest qualifying bonus tier.
		var best *SetBonus
		for i := range set.Bonuses {
			b := &set.Bonuses[i]
			if owned >= b.PiecesRequired {
				best = b
			}
		}
		if best == nil {
			// Has pieces but no tier met yet — still show in UI as partial.
			active = append(active, ActiveSetBonus{
				SetID:       set.ID,
				SetName:     set.Name,
				PiecesOwned: owned,
				PiecesTotal: len(set.MemberIDs),
			})
			continue
		}
		active = append(active, ActiveSetBonus{
			SetID:       set.ID,
			SetName:     set.Name,
			PiecesOwned: owned,
			PiecesTotal: len(set.MemberIDs),
			Tier:        *best,
		})
	}
	return active
}
