package items

import (
	"testing"
)

func TestEffectsEval(t *testing.T) {
	// Create mock equipment map
	eq := make(map[string]*Item)

	// Test case 1: Empty equipment
	if val := EvalPassiveDamageReduction(eq); val != 0 {
		t.Errorf("expected 0 damage reduction, got %f", val)
	}
	if val := EvalPassiveAllResistance(eq); val != 0 {
		t.Errorf("expected 0 all resistance, got %f", val)
	}
	if val := EvalPassiveManaCostReduction(eq); val != 0 {
		t.Errorf("expected 0 mana cost reduction, got %f", val)
	}
	if val := EvalPassiveCooldownReduction(eq); val != 0 {
		t.Errorf("expected 0 cooldown reduction, got %f", val)
	}
	if val := EvalPassiveGoldFind(eq); val != 0 {
		t.Errorf("expected 0 gold find, got %f", val)
	}
	if val := EvalOnKillLifesteal(eq); val != 0 {
		t.Errorf("expected 0 lifesteal, got %f", val)
	}
	if mult, ok := EvalOnHitCrit(eq); ok || mult != 1.0 {
		t.Errorf("expected no crit and 1.0 mult, got %t and %f", ok, mult)
	}

	// Test case 2: Equipped items with passive effects
	eq["Weapon"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_weapon",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "damage_reduction",
				MagnitudePct: 10,
			},
		},
	}
	eq["Shield"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_shield",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "damage_reduction",
				MagnitudePct: 15,
			},
		},
	}
	if val := EvalPassiveDamageReduction(eq); val != 0.25 {
		t.Errorf("expected 0.25 damage reduction, got %f", val)
	}

	// Test capping of damage reduction
	eq["Chest"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_chest",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "damage_reduction",
				MagnitudePct: 60,
			},
		},
	}
	if val := EvalPassiveDamageReduction(eq); val != 0.75 { // cap is 0.75
		t.Errorf("expected capped 0.75 damage reduction, got %f", val)
	}
	delete(eq, "Chest")

	// Test passive all resistance
	eq["Amulet"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_amulet",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "all_resistance",
				MagnitudePct: 12,
			},
		},
	}
	if val := EvalPassiveAllResistance(eq); val != 0.12 {
		t.Errorf("expected 0.12 all resistance, got %f", val)
	}

	// Test mana cost reduction
	eq["Ring1"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_ring1",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "mana_cost_reduction",
				MagnitudePct: 20,
			},
		},
	}
	if val := EvalPassiveManaCostReduction(eq); val != 0.20 {
		t.Errorf("expected 0.20 mana cost reduction, got %f", val)
	}

	// Test cooldown reduction
	eq["Ring2"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_ring2",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "cooldown_reduction",
				MagnitudePct: 15,
			},
		},
	}
	if val := EvalPassiveCooldownReduction(eq); val != 0.15 {
		t.Errorf("expected 0.15 cooldown reduction, got %f", val)
	}

	// Test gold find
	eq["Feet"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_feet",
			Effect: &ItemEffect{
				Trigger:      "passive",
				Type:         "gold_find",
				MagnitudePct: 50,
			},
		},
	}
	if val := EvalPassiveGoldFind(eq); val != 0.50 {
		t.Errorf("expected 0.50 gold find, got %f", val)
	}

	// Test lifesteal
	eq["Helm"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_helm",
			Effect: &ItemEffect{
				Trigger:      "on_kill",
				Type:         "lifesteal",
				MagnitudePct: 5,
			},
		},
	}
	if val := EvalOnKillLifesteal(eq); val != 0.05 {
		t.Errorf("expected 0.05 lifesteal, got %f", val)
	}

	// Test Crit on hit (high chance for reliable test)
	eq["Gloves"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID: "test_gloves",
			Effect: &ItemEffect{
				Trigger:      "on_hit",
				Type:         "crit_multiplier",
				ChancePct:    100,
				MagnitudePct: 50,
			},
		},
	}
	if mult, ok := EvalOnHitCrit(eq); !ok || mult != 1.50 {
		t.Errorf("expected crit with 1.50 multiplier, got %t and %f", ok, mult)
	}

	// Test OnLowHP trigger
	lowHPEff := &ItemEffect{
		Trigger:      "on_low_hp",
		Type:         "counterpulse",
		ThresholdPct: 30,
		CooldownSec:  10.0,
	}
	eq["Shoulders"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID:     "test_shoulders",
			Effect: lowHPEff,
		},
	}

	// Should not trigger at 50% HP
	if triggered := EvalOnLowHP(eq, 0.50); len(triggered) != 0 {
		t.Errorf("expected no triggers at 50%% HP, got %d", len(triggered))
	}

	// Should trigger at 25% HP
	triggered := EvalOnLowHP(eq, 0.25)
	if len(triggered) != 1 || triggered[0] != lowHPEff {
		t.Errorf("expected 1 trigger matching shoulder effect at 25%% HP, got %d", len(triggered))
	}

	// Put on cooldown and check it doesn't trigger again
	lowHPEff.PutOnCooldown()
	if triggered2 := EvalOnLowHP(eq, 0.25); len(triggered2) != 0 {
		t.Errorf("expected no triggers when effect is on cooldown, got %d", len(triggered2))
	}

	// Tick cooldown down and check it becomes ready
	TickEquippedEffectCooldowns(eq, 5.0)
	if lowHPEff.IsReady() {
		t.Errorf("expected effect not ready after 5s tick of 10s cooldown")
	}
	TickEquippedEffectCooldowns(eq, 5.0)
	if !lowHPEff.IsReady() {
		t.Errorf("expected effect ready after total 10s tick")
	}

	// Test periodic regen
	regenEff := &ItemEffect{
		Trigger:       "passive",
		Type:          "regen_hp",
		IntervalSec:   2.0,
		MagnitudeFlat: 3,
	}
	eq["Belt"] = &Item{
		ItemTemplate: &ItemTemplate{
			ID:     "test_belt",
			Effect: regenEff,
		},
	}

	// Ticking less than interval should return 0 heal
	if heal := TickRegenEffects(eq, 1.0); heal != 0 {
		t.Errorf("expected 0 heal after 1s tick, got %d", heal)
	}
	// Ticking rest of interval should return 3 heal
	if heal := TickRegenEffects(eq, 1.0); heal != 3 {
		t.Errorf("expected 3 heal after another 1s tick, got %d", heal)
	}
}
