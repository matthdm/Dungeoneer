package game

import (
	"math"

	"dungeoneer/fov"
	"dungeoneer/spells"
)

func (g *Game) updateSpells() {
	var remaining []spells.Spell
	for _, sp := range g.ActiveSpells {
		sp.Update(g.currentLevel, g.DeltaTime)
		if fb, ok := sp.(*spells.Fireball); ok {
			if !fb.Impact {
				for _, m := range g.Monsters {
					if m.IsDead {
						continue
					}

					dx := m.InterpX - fb.X
					dy := m.InterpY - fb.Y
					distSq := dx*dx + dy*dy

					if distSq <= fb.Radius*fb.Radius {
						fb.Impact = true
						tx := int(math.Floor(fb.X))
						ty := int(math.Floor(fb.Y))
						g.applyFireballDamage(fb, tx, ty)
						break
					}
				}
			} else if fb.IsFinished() {
				// impact done
			}
		}
		if !sp.IsFinished() {
			remaining = append(remaining, sp)
		}
	}
	g.ActiveSpells = remaining
}

func (g *Game) applyFireballDamage(fb *spells.Fireball, cx, cy int) {
	level := fb.Info.Level
	var radius int
	mult := 1.0
	switch level {
	case 1:
		radius = 0
		fb.ImpactImg = g.spriteSheet.FireBurst
		mult = 1.0
	case 2:
		radius = 1
		fb.ImpactImg = g.spriteSheet.FireBurst2
		mult = 2.0
	case 3:
		radius = 2
		fb.ImpactImg = g.spriteSheet.FireBurst3
		mult = 4.0
	}

	dmg := int(float64(fb.Info.Damage) * mult)
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		dx := int(math.Abs(float64(m.TileX - cx)))
		dy := int(math.Abs(float64(m.TileY - cy)))
		if dx <= radius && dy <= radius {
			if g.hasLineOfSight(cx, cy, m.TileX, m.TileY) {
				m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers)
			}
		}
	}
}

func (g *Game) hasLineOfSight(x1, y1, x2, y2 int) bool {
	pts := fov.TraceLineToTiles(float64(x1)+0.5, float64(y1)+0.5, float64(x2)+0.5, float64(y2)+0.5)
	for _, p := range pts {
		if p.X == x1 && p.Y == y1 {
			continue
		}
		if p.X == x2 && p.Y == y2 {
			return true
		}
		if !g.currentLevel.IsWalkable(p.X, p.Y) {
			return false
		}
	}
	return true
}

func (g *Game) castFireball(casterX, casterY, targetX, targetY float64, c *spells.Caster) {
	info := spells.SpellInfo{Name: "fireball", Level: 3, Cooldown: 1.0, Damage: 6}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	fb := spells.NewFireball(info, casterX, casterY, targetX, targetY, g.fireballSprites, g.spriteSheet.FireBurst)
	g.ActiveSpells = append(g.ActiveSpells, fb)
}
