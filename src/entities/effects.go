package entities

// EffectType identifies a status effect category.
type EffectType string

const (
	EffectPoison EffectType = "poison"
	EffectBurn   EffectType = "burn"
	EffectSlow   EffectType = "slow"
	EffectShield EffectType = "shield"
	EffectWeaken EffectType = "weaken"
	EffectHaste  EffectType = "haste"
)

// StatusEffect represents an active buff or debuff on an entity.
type StatusEffect struct {
	Type      EffectType
	Duration  float64 // seconds remaining
	TickRate  float64 // seconds between damage ticks (for DoTs)
	TickTimer float64 // accumulator
	Value     int     // damage per tick, speed modifier %, shield HP, etc.
	Source    string  // identifier for stacking rules
}

// EffectHolder manages a set of active status effects.
type EffectHolder struct {
	Effects []*StatusEffect
}

// AddEffect appends an effect. Duplicate sources refresh duration rather than stack.
func (h *EffectHolder) AddEffect(e *StatusEffect) {
	for _, existing := range h.Effects {
		if existing.Type == e.Type && existing.Source == e.Source {
			existing.Duration = e.Duration
			existing.Value = e.Value
			return
		}
	}
	h.Effects = append(h.Effects, e)
}

// HasEffect returns true if any effect of the given type is active.
func (h *EffectHolder) HasEffect(t EffectType) bool {
	for _, e := range h.Effects {
		if e.Type == t && e.Duration > 0 {
			return true
		}
	}
	return false
}

// UpdateEffects advances all effects by dt seconds. DoTs call takeDamage for
// each tick. Expired effects are removed.
func (h *EffectHolder) UpdateEffects(dt float64, takeDamage func(int)) {
	alive := h.Effects[:0]
	for _, e := range h.Effects {
		e.Duration -= dt
		if e.Duration <= 0 {
			continue // expired
		}

		// Process damage-over-time effects.
		switch e.Type {
		case EffectPoison, EffectBurn:
			e.TickTimer += dt
			if e.TickRate > 0 && e.TickTimer >= e.TickRate {
				e.TickTimer -= e.TickRate
				if takeDamage != nil {
					takeDamage(e.Value)
				}
			}
		}

		alive = append(alive, e)
	}
	// Nil out removed slots for GC.
	for i := len(alive); i < len(h.Effects); i++ {
		h.Effects[i] = nil
	}
	h.Effects = alive
}

// SpeedModifier returns a multiplier (1.0 = normal) considering slow/haste.
func (h *EffectHolder) SpeedModifier() float64 {
	mod := 1.0
	for _, e := range h.Effects {
		if e.Duration <= 0 {
			continue
		}
		switch e.Type {
		case EffectSlow:
			// Value is the slow percentage (e.g. 50 = 50% slower).
			mod *= 1.0 - float64(e.Value)/100.0
		case EffectHaste:
			mod *= 1.0 + float64(e.Value)/100.0
		}
	}
	if mod < 0.1 {
		mod = 0.1 // floor to prevent frozen entities
	}
	return mod
}

// DamageModifier returns a multiplier for outgoing damage.
func (h *EffectHolder) DamageModifier() float64 {
	mod := 1.0
	for _, e := range h.Effects {
		if e.Duration <= 0 {
			continue
		}
		if e.Type == EffectWeaken {
			mod *= 1.0 - float64(e.Value)/100.0
		}
	}
	if mod < 0.1 {
		mod = 0.1
	}
	return mod
}

// ShieldAmount returns the total remaining shield HP across all shield effects.
func (h *EffectHolder) ShieldAmount() int {
	total := 0
	for _, e := range h.Effects {
		if e.Type == EffectShield && e.Duration > 0 {
			total += e.Value
		}
	}
	return total
}

// AbsorbDamage reduces incoming damage by consuming shield effects.
// Returns the remaining damage after shields.
func (h *EffectHolder) AbsorbDamage(dmg int) int {
	for _, e := range h.Effects {
		if dmg <= 0 {
			break
		}
		if e.Type == EffectShield && e.Duration > 0 && e.Value > 0 {
			if e.Value >= dmg {
				e.Value -= dmg
				dmg = 0
			} else {
				dmg -= e.Value
				e.Value = 0
				e.Duration = 0 // spent
			}
		}
	}
	return dmg
}
