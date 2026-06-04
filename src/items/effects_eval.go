package items

import "math/rand/v2"

// EvalPassiveDamageReduction returns the total incoming-damage reduction factor
// (0.0–0.9) from all equipped passive "damage_reduction" effects.
// Example: two items each granting 5% → 0.10 → caller multiplies dmg by 0.90.
func EvalPassiveDamageReduction(eq map[string]*Item) float64 {
	total := 0.0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if (e.Trigger == "passive") && e.Type == "damage_reduction" {
			total += float64(e.MagnitudePct) / 100.0
		}
	}
	// Cap at 75% reduction to prevent invincibility.
	if total > 0.75 {
		total = 0.75
	}
	return total
}

// EvalPassiveAllResistance returns a combined resistance bonus from
// "all_resistance" passive effects. This is applied additively with damage
// reduction: a 2% all_resistance means 2% less damage from all sources.
func EvalPassiveAllResistance(eq map[string]*Item) float64 {
	total := 0.0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger == "passive" && e.Type == "all_resistance" {
			total += float64(e.MagnitudePct) / 100.0
		}
	}
	if total > 0.60 {
		total = 0.60
	}
	return total
}

// EvalPassiveManaCostReduction returns the mana-cost reduction multiplier
// (0.0–0.5 max). Caller multiplies base cost by (1 - result).
func EvalPassiveManaCostReduction(eq map[string]*Item) float64 {
	total := 0.0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger == "passive" && e.Type == "mana_cost_reduction" {
			total += float64(e.MagnitudePct) / 100.0
		}
	}
	if total > 0.50 {
		total = 0.50
	}
	return total
}

// EvalPassiveCooldownReduction returns the cooldown reduction multiplier
// (0.0–0.50 max). Caller multiplies base cooldown by (1 - result).
func EvalPassiveCooldownReduction(eq map[string]*Item) float64 {
	total := 0.0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger == "passive" && e.Type == "cooldown_reduction" {
			total += float64(e.MagnitudePct) / 100.0
		}
	}
	if total > 0.50 {
		total = 0.50
	}
	return total
}

// EvalPassiveGoldFind returns the gold-find bonus multiplier (0.0–2.0 max).
// Caller multiplies base gold drop by (1 + result).
func EvalPassiveGoldFind(eq map[string]*Item) float64 {
	total := 0.0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger == "passive" && e.Type == "gold_find" {
			total += float64(e.MagnitudePct) / 100.0
		}
	}
	if total > 2.0 {
		total = 2.0
	}
	return total
}

// EvalOnKillLifesteal returns the HP percentage to restore on each kill.
// Returns 0 if no lifesteal is equipped.
func EvalOnKillLifesteal(eq map[string]*Item) float64 {
	total := 0.0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger == "on_kill" && e.Type == "lifesteal" {
			total += float64(e.MagnitudePct) / 100.0
		}
	}
	return total
}

// EvalOnHitCrit checks all equipped "on_hit crit_multiplier" effects.
// Returns (multiplier, true) when a crit should occur, (1.0, false) otherwise.
// MagnitudePct is the extra damage multiplier (e.g. 100 → double damage).
// ChancePct is the roll threshold (e.g. 5 → 5% chance per hit).
func EvalOnHitCrit(eq map[string]*Item) (multiplier float64, triggered bool) {
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger == "on_hit" && e.Type == "crit_multiplier" {
			chance := float64(e.ChancePct) / 100.0
			if chance <= 0 {
				continue
			}
			if rand.Float64() < chance {
				mult := 1.0 + float64(e.MagnitudePct)/100.0
				return mult, true
			}
		}
	}
	return 1.0, false
}

// EvalOnLowHP checks all equipped "on_low_hp" triggered effects.
// hpPct must be in [0, 1]. Returns matching effects whose cooldown is ready.
// Callers must call eff.PutOnCooldown() after consuming each returned effect.
func EvalOnLowHP(eq map[string]*Item, hpPct float64) []*ItemEffect {
	var triggered []*ItemEffect
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger != "on_low_hp" {
			continue
		}
		threshold := float64(e.ThresholdPct) / 100.0
		if threshold <= 0 {
			threshold = 0.20 // default: 20%
		}
		if hpPct <= threshold && e.IsReady() {
			triggered = append(triggered, e)
		}
	}
	return triggered
}

// TickEquippedEffectCooldowns advances the cooldown timer for every equipped
// item's effect by dt seconds. Call once per frame from Player.Update.
func TickEquippedEffectCooldowns(eq map[string]*Item, dt float64) {
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		it.Effect.TickCooldown(dt)
	}
}

// TickRegenEffects advances any "regen_hp" periodic effects and returns
// the total HP to restore this frame (may be 0).
func TickRegenEffects(eq map[string]*Item, dt float64) int {
	total := 0
	for _, it := range eq {
		if it == nil || it.Effect == nil {
			continue
		}
		e := it.Effect
		if e.Trigger != "passive" || e.Type != "regen_hp" {
			continue
		}
		interval := e.IntervalSec
		if interval <= 0 {
			interval = 10.0 // default: 1 HP every 10 seconds
		}
		e.intervalAcc += dt
		for e.intervalAcc >= interval {
			e.intervalAcc -= interval
			heal := e.MagnitudeFlat
			if heal <= 0 {
				heal = 1
			}
			total += heal
		}
	}
	return total
}
