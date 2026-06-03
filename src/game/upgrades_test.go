package game

import (
	"dungeoneer/entities"
	"testing"
)

func makeTestPlayer() *entities.Player {
	// Minimal player for testing — zero-value is safe for stat tests.
	return &entities.Player{
		Stats: entities.BaseStats{
			Strength: 1, Dexterity: 1, Vitality: 1, Intelligence: 1, Luck: 1,
		},
		Abilities: make(map[string]bool),
	}
}

func TestUpgrade_IronConstitution_AppliesVitality(t *testing.T) {
	p := makeTestPlayer()
	before := p.Stats.Vitality

	def := GetUpgradeDef("iron_constitution")
	if def == nil {
		t.Fatal("iron_constitution not found in registry")
	}
	def.Apply(p, 2) // level 2 = +4 Vitality

	if p.Stats.Vitality != before+4 {
		t.Errorf("Vitality want %d, got %d", before+4, p.Stats.Vitality)
	}
	// MaxHP should also have increased via RecalculateStats.
	if p.MaxHP <= 0 {
		t.Error("MaxHP should be positive after RecalculateStats")
	}
}

func TestUpgrade_InnateDash_GrantsAbility(t *testing.T) {
	p := makeTestPlayer()

	def := GetUpgradeDef("innate_dash")
	if def == nil {
		t.Fatal("innate_dash not found in registry")
	}
	def.Apply(p, 1)

	if !p.Abilities["dash"] {
		t.Error("innate_dash should grant dash ability")
	}
}

func TestUpgradeCost_MaxedOutReturnsZero(t *testing.T) {
	def := GetUpgradeDef("innate_dash") // max level 1
	cost := UpgradeCost(def, 1)
	if cost != 0 {
		t.Errorf("maxed upgrade should cost 0, got %d", cost)
	}
}

func TestUpgradeCost_FirstLevel(t *testing.T) {
	def := GetUpgradeDef("iron_constitution")
	cost := UpgradeCost(def, 0)
	if cost != 50 {
		t.Errorf("iron_constitution level 1 cost want 50, got %d", cost)
	}
}

func TestGetUpgradeDef_UnknownID(t *testing.T) {
	def := GetUpgradeDef("nonexistent_upgrade")
	if def != nil {
		t.Error("expected nil for unknown upgrade ID")
	}
}

func TestUpgradeCost_SecondLevel(t *testing.T) {
	def := GetUpgradeDef("iron_constitution")
	cost := UpgradeCost(def, 1)
	if cost != 100 {
		t.Errorf("iron_constitution level 2 cost want 100, got %d", cost)
	}
}

func TestUpgradeRegistry_AllDefsValid(t *testing.T) {
	for _, def := range UpgradeRegistry {
		if def.ID == "" {
			t.Error("upgrade def has empty ID")
		}
		if def.MaxLevel <= 0 {
			t.Errorf("upgrade %q has MaxLevel <= 0", def.ID)
		}
		if len(def.CostPerLevel) != def.MaxLevel {
			t.Errorf("upgrade %q: CostPerLevel len %d != MaxLevel %d",
				def.ID, len(def.CostPerLevel), def.MaxLevel)
		}
		if def.Apply == nil {
			t.Errorf("upgrade %q has nil Apply func", def.ID)
		}
	}
}
