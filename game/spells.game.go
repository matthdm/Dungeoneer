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
		radius = 1
		fb.ImpactImg = g.spriteSheet.FireBurst
		mult = 1.0
	case 2:
		radius = 2
		fb.ImpactImg = g.spriteSheet.FireBurst2
		mult = 2.0
	case 3:
		radius = 3
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
	info := spells.SpellInfo{Name: "fireball", Level: 1, Cooldown: 1.0, Damage: 5}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	fb := spells.NewFireball(info, casterX, casterY, targetX, targetY, g.fireballSprites, g.spriteSheet.FireBurst)
	g.ActiveSpells = append(g.ActiveSpells, fb)
}

func (g *Game) applyChaosRayDamage(cr *spells.ChaosRay) {
	radius := 0.6
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		px := m.InterpX
		py := m.InterpY
		for i := 0; i < len(cr.Path)-1; i++ {
			p1 := cr.Path[i]
			p2 := cr.Path[i+1]
			if pointSegmentDistance(px, py, p1.X, p1.Y, p2.X, p2.Y) <= radius {
				m.TakeDamage(cr.Info.Damage, &g.HitMarkers, &g.DamageNumbers)
				break
			}
		}
	}
}

func pointSegmentDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	if dx == 0 && dy == 0 {
		return math.Hypot(px-x1, py-y1)
	}
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	projX := x1 + t*dx
	projY := y1 + t*dy
	return math.Hypot(px-projX, py-projY)
}

func (g *Game) castChaosRay(casterX, casterY, targetX, targetY float64, c *spells.Caster) {
	info := spells.SpellInfo{Name: "chaosray", Level: 1, Cooldown: 1.0, Damage: 8}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	cr := spells.NewChaosRay(info, casterX, casterY, targetX, targetY)
	g.applyChaosRayDamage(cr)
	g.ActiveSpells = append(g.ActiveSpells, cr)
}
