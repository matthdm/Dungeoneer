package entities

import (
	"dungeoneer/levels"
)

// PendingSpellCast is queued by caster monsters for the game loop to process
// into actual spell instances (fireballs, etc.).
type PendingSpellCast struct {
	SpellName        string
	OriginX, OriginY float64
	TargetX, TargetY float64
	Damage           int
}

// CasterBehavior keeps distance and casts spells (fireballs) at the player.
type CasterBehavior struct {
	AttackRange   int
	FleeRange     int
	Triggered     bool
	TriggerRadius int
	CastCooldown  int // ticks between casts
	CastCounter   int
	SpellName     string // "fireball" etc.
}

// NewCasterBehavior creates a CasterBehavior with sensible defaults.
func NewCasterBehavior(triggerRadius int) *CasterBehavior {
	return &CasterBehavior{
		AttackRange:   7,
		FleeRange:     3,
		TriggerRadius: triggerRadius,
		CastCooldown:  90, // ~1.5 seconds between casts
		SpellName:     "fireball",
	}
}

func (c *CasterBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead || m.Moving {
		return
	}

	// Use continuous world-space distance so behaviour matches the visual
	// positions on screen, not lagging integer tile coords.
	dist := m.Pos().DistTo(p.Pos())

	// Trigger check.
	if !c.Triggered {
		if dist <= float64(c.TriggerRadius) {
			c.Triggered = true
		} else {
			return
		}
	}

	c.CastCounter++

	// Too close — retreat.
	if dist < float64(c.FleeRange) {
		c.retreat(m, p, level)
		return
	}

	// In range — cast spell.
	if dist <= float64(c.AttackRange) {
		if c.CastCounter >= c.CastCooldown {
			c.CastCounter = 0
			// Use continuous positions for origin and target so spells fire
			// from where the sprite is drawn and track the player's visual position.
			m.PendingSpells = append(m.PendingSpells, PendingSpellCast{
				SpellName: c.SpellName,
				OriginX:   m.InterpX,
				OriginY:   m.InterpY,
				TargetX:   p.MoveController.InterpX,
				TargetY:   p.MoveController.InterpY,
				Damage:    m.Damage,
			})
		}
		return
	}

	// Out of range — chase to get in range.
	m.BasicChaseLogic(p, level)
}

// retreat tries to move the caster away from the player.
func (c *CasterBehavior) retreat(m *Monster, p *Player, level *levels.Level) {
	bestDist := -1.0
	bestX, bestY := m.TileX, m.TileY
	for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := m.TileX+d[0], m.TileY+d[1]
		if !level.IsWalkable(nx, ny) {
			continue
		}
		ddx := float64(nx) - p.MoveController.InterpX
		ddy := float64(ny) - p.MoveController.InterpY
		dd := ddx*ddx + ddy*ddy
		if dd > bestDist {
			bestDist = dd
			bestX = nx
			bestY = ny
		}
	}
	if bestX != m.TileX || bestY != m.TileY {
		m.MoveTo(bestX, bestY)
	}
}
