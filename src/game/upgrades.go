package game

import "dungeoneer/entities"

// UpgradeDef defines a purchasable permanent upgrade.
type UpgradeDef struct {
	ID           string
	Name         string
	Description  string
	MaxLevel     int
	CostPerLevel []int // Remnant cost for each level (index 0 = level 1)
	Apply        func(p *entities.Player, level int)
}

// UpgradeRegistry holds all purchasable upgrades.
var UpgradeRegistry = []*UpgradeDef{
	{
		ID: "iron_constitution", Name: "Iron Constitution",
		Description:  "+2 Vitality per level (increases max HP)",
		MaxLevel:     3,
		CostPerLevel: []int{50, 100, 200},
		Apply: func(p *entities.Player, level int) {
			p.Stats.Vitality += 2 * level
			p.RecalculateStats()
		},
	},
	{
		ID: "sharpened_edge", Name: "Sharpened Edge",
		Description:  "+2 Strength per level (increases melee damage)",
		MaxLevel:     3,
		CostPerLevel: []int{60, 120, 240},
		Apply: func(p *entities.Player, level int) {
			p.Stats.Strength += 2 * level
			p.RecalculateStats()
		},
	},
	{
		ID: "mana_well", Name: "Mana Well",
		Description:  "+2 Intelligence per level (increases max mana)",
		MaxLevel:     3,
		CostPerLevel: []int{50, 100, 200},
		Apply: func(p *entities.Player, level int) {
			p.Stats.Intelligence += 2 * level
			p.RecalculateStats()
		},
	},
	{
		ID: "swift_strikes", Name: "Swift Strikes",
		Description:  "+2 Dexterity per level (increases attack speed)",
		MaxLevel:     3,
		CostPerLevel: []int{70, 140, 280},
		Apply: func(p *entities.Player, level int) {
			p.Stats.Dexterity += 2 * level
			p.RecalculateStats()
		},
	},
	{
		ID: "scavenger", Name: "Scavenger",
		Description:  "+1 Luck per level (improves loot drops)",
		MaxLevel:     3,
		CostPerLevel: []int{80, 160, 320},
		Apply: func(p *entities.Player, level int) {
			p.Stats.Luck += 1 * level
			p.RecalculateStats()
		},
	},
	{
		ID: "innate_dash", Name: "Innate Dash",
		Description:  "Start every run with dash ability",
		MaxLevel:     1,
		CostPerLevel: []int{150},
		Apply: func(p *entities.Player, level int) {
			if p.Abilities == nil {
				p.Abilities = make(map[string]bool)
			}
			p.Abilities["dash"] = true
		},
	},
}

// GetUpgradeDef returns an upgrade definition by ID.
func GetUpgradeDef(id string) *UpgradeDef {
	for _, u := range UpgradeRegistry {
		if u.ID == id {
			return u
		}
	}
	return nil
}

// UpgradeCost returns the Remnant cost to purchase the NEXT level of an upgrade.
// currentLevel is the player's current level (0 = not purchased).
func UpgradeCost(def *UpgradeDef, currentLevel int) int {
	if currentLevel >= def.MaxLevel {
		return 0 // maxed out
	}
	if currentLevel < len(def.CostPerLevel) {
		return def.CostPerLevel[currentLevel]
	}
	return 999
}
