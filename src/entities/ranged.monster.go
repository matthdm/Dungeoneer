package entities

import (
	"dungeoneer/levels"
)

// RangedBehavior keeps distance from the player and fires projectiles.
type RangedBehavior struct {
	AttackRange   int // max tiles to shoot from
	FleeRange     int // retreat if player closer than this
	Triggered     bool
	TriggerRadius int
	ShootCooldown int // ticks between shots
	ShootCounter  int
	ProjectileDmg int // overridden by monster Damage if 0
}

// NewRangedBehavior creates a RangedBehavior with sensible defaults.
func NewRangedBehavior(triggerRadius int) *RangedBehavior {
	return &RangedBehavior{
		AttackRange:   6,
		FleeRange:     2,
		TriggerRadius: triggerRadius,
		ShootCooldown: 60,
	}
}

func (r *RangedBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead || m.Moving {
		return
	}

	// Use continuous world-space distance so behaviour matches the visual
	// positions on screen, not lagging integer tile coords.
	dist := m.Pos().DistTo(p.Pos())

	// Trigger check.
	if !r.Triggered {
		if dist <= float64(r.TriggerRadius) {
			r.Triggered = true
		} else {
			return
		}
	}

	r.ShootCounter++

	// Too close — retreat.
	if dist < float64(r.FleeRange) {
		r.retreat(m, p, level)
		return
	}

	// In range — shoot.
	if dist <= float64(r.AttackRange) {
		if r.ShootCounter >= r.ShootCooldown {
			r.ShootCounter = 0
			dmg := r.ProjectileDmg
			if dmg == 0 {
				dmg = m.Damage
			}
			// Body to body, in world space: the same points HitsPlayer tests
			// against and the same points the projectile is drawn between.
			from, to := m.Pos().BodyCenter(), p.Pos().BodyCenter()
			proj := NewMonsterProjectile(from.X, from.Y, to.X, to.Y, 0.15, dmg)
			m.PendingProjectiles = append(m.PendingProjectiles, proj)
		}
		return
	}

	// Out of range — chase to get in range.
	m.BasicChaseLogic(p, level)
}

// retreat tries to move the monster away from the player.
func (r *RangedBehavior) retreat(m *Monster, p *Player, level *levels.Level) {
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

