package game

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"testing"
)

func TestStatsScaling_SpellDamageBonus(t *testing.T) {
	g := &Game{
		player: &entities.Player{
			Stats: entities.BaseStats{
				Intelligence: 10, // should give +5 spell damage
			},
		},
	}
	g.player.RecalculateStats()

	// Base 10 spell damage + 10/2 intelligence bonus = 15 damage
	gotDmg := g.getSpellDamage(10)
	if gotDmg != 15 {
		t.Errorf("expected 15 spell damage, got %d", gotDmg)
	}
}

func TestStatsScaling_SpellCDR(t *testing.T) {
	g := &Game{
		player: &entities.Player{
			Stats: entities.BaseStats{
				Dexterity: 10, // should give 10% CDR
			},
			Equipment: make(map[string]*items.Item),
		},
	}
	g.player.RecalculateStats()

	// Base 2.0 cooldown, 10% CDR -> 1.8 cooldown
	gotCD := g.getSpellCooldown(2.0)
	if gotCD != 1.8 {
		t.Errorf("expected 1.8 cooldown, got %f", gotCD)
	}

	// Add item CDR passive (e.g. 20% CDR)
	g.player.Equipment["Ring1"] = &items.Item{
		ItemTemplate: &items.ItemTemplate{
			ID: "test_ring_cdr",
			Effect: &items.ItemEffect{
				Trigger:      "passive",
				Type:         "cooldown_reduction",
				MagnitudePct: 20,
			},
		},
	}

	// Total CDR: 10% (dex) + 20% (items) = 30% CDR
	// Base 2.0 cooldown -> 1.4 cooldown
	gotCD = g.getSpellCooldown(2.0)
	if gotCD != 1.4 {
		t.Errorf("expected 1.4 cooldown with items, got %f", gotCD)
	}
}

func TestStatsScaling_ManaCostReduction(t *testing.T) {
	g := &Game{
		player: &entities.Player{
			Stats: entities.BaseStats{},
			Equipment: make(map[string]*items.Item),
		},
	}
	g.player.RecalculateStats()

	// Base cost of fireball is 8. With no reduction, should be 8.
	cost := g.spellManaCost("fireball")
	if cost != 8 {
		t.Errorf("expected fireball cost to be 8, got %d", cost)
	}

	// Add 25% mana cost reduction item
	g.player.Equipment["Ring1"] = &items.Item{
		ItemTemplate: &items.ItemTemplate{
			ID: "test_ring_mcr",
			Effect: &items.ItemEffect{
				Trigger:      "passive",
				Type:         "mana_cost_reduction",
				MagnitudePct: 25,
			},
		},
	}

	// Fireball cost 8 -> should be 6 (8 * 0.75)
	cost = g.spellManaCost("fireball")
	if cost != 6 {
		t.Errorf("expected fireball cost to be 6, got %d", cost)
	}
}

func TestStatsScaling_VitalityArmor(t *testing.T) {
	p := &entities.Player{
		Stats: entities.BaseStats{
			Vitality: 20, // should give 10% armor (20 * 0.5%)
		},
		HP:    100,
		MaxHP: 100,
		Equipment: make(map[string]*items.Item),
	}
	p.RecalculateStats()

	// Take 10 raw damage -> should be reduced by 10% -> 9 damage
	p.TakeDamage(10)
	if p.HP != 91 { // 100 - 9 = 91
		t.Errorf("expected player HP to be 91 after 10 percent armor reduction, got %d", p.HP)
	}
}
