package game

import (
	"math"

	"dungeoneer/entities"
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
		if ls, ok := sp.(*spells.LightningStrike); ok {
			if !ls.DamageApplied {
				g.applyLightningDamage(ls, int(math.Floor(ls.X)), int(math.Floor(ls.Y)))
				ls.DamageApplied = true
			}
		}
		if storm, ok := sp.(*spells.LightningStorm); ok {
			for _, ns := range storm.TakeSpawns() {
				remaining = append(remaining, ns)
			}
		}
		if bloom, ok := sp.(*spells.FractalBloom); ok {
			for _, n := range bloom.TakeSpawns() {
				remaining = append(remaining, n)
			}
		}
		if fc, ok := sp.(*spells.FractalCanopy); ok {
			g.applyFractalCanopyHealing(fc)
		}
		if node, ok := sp.(*spells.FractalNode); ok {
			if !node.DamageApplied {
				g.applyFractalDamage(node, int(math.Floor(node.X)), int(math.Floor(node.Y)))
				node.DamageApplied = true
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
                               if m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers) {
                                       g.awardEXP(m)
                               }
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
                               if m.TakeDamage(cr.Info.Damage, &g.HitMarkers, &g.DamageNumbers) {
                                       g.awardEXP(m)
                               }
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

func (g *Game) applyLightningDamage(l *spells.LightningStrike, cx, cy int) {
	radius := 1
	dmg := l.Info.Damage
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		dx := int(math.Abs(float64(m.TileX - cx)))
		dy := int(math.Abs(float64(m.TileY - cy)))
		if dx <= radius && dy <= radius {
			if g.hasLineOfSight(cx, cy, m.TileX, m.TileY) {
                               if m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers) {
                                       g.awardEXP(m)
                               }
                       }
               }
	}
}

func (g *Game) castLightningStrike(targetX, targetY float64, c *spells.Caster) {
	info := spells.SpellInfo{Name: "lightning", Level: 1, Cooldown: 0.01, Damage: 8}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	ls := spells.NewLightningStrike(info, targetX, targetY, g.spriteSheet.ArcaneBurst)
	g.ActiveSpells = append(g.ActiveSpells, ls)
}

func (g *Game) castLightningStorm(centerX, centerY float64, c *spells.Caster) {
	info := spells.SpellInfo{Name: "lightningstorm", Level: 1, Cooldown: 3.0, Damage: 8}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)

	// Do not apply any offset hack here — use center-aligned tile coordinates
	tx := float64(g.hoverTileX)
	ty := float64(g.hoverTileY)

	storm := spells.NewLightningStorm(info, tx, ty, 3, 0.2, 3.0, c, g.spriteSheet.ArcaneBurst, g.currentLevel)
	g.ActiveSpells = append(g.ActiveSpells, storm)
}

func (g *Game) applyFractalDamage(n *spells.FractalNode, cx, cy int) {
	radius := n.Radius
	dmg := n.Damage
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		dx := int(math.Abs(float64(m.TileX - cx)))
		dy := int(math.Abs(float64(m.TileY - cy)))
		if dx <= radius && dy <= radius {
			if g.hasLineOfSight(cx, cy, m.TileX, m.TileY) {
                               if m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers) {
                                       g.awardEXP(m)
                               }
                       }
               }
       }
}

func (g *Game) castFractalBloom(centerX, centerY float64, c *spells.Caster) {
	info := spells.SpellInfo{Name: "fractalbloom", Level: 1, Cooldown: 4.0, Damage: 6}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	bloom := spells.NewFractalBloom(info, centerX, centerY, c, g.spriteSheet.ArcaneBurst, g.currentLevel, 3, 0.7, 0.2)
	g.ActiveSpells = append(g.ActiveSpells, bloom)
}
func (g *Game) applyFractalCanopyHealing(fc *spells.FractalCanopy) {
	if g.player == nil || g.player.IsDead {
		return
	}

	dx := g.player.MoveController.InterpX - fc.X
	dy := g.player.MoveController.InterpY - fc.Y
	dist := math.Hypot(dx, dy)
	if dist > fc.Radius {
		return
	}

	// Scale heal based on growth stage
	frac := fc.Age / fc.MaxGrowTime
	if frac > 1 {
		frac = 1
	}
	healPerSec := fc.HealingMin + (fc.HealingMax-fc.HealingMin)*frac

	// Heal every full second
	fc.HealingTickTimer += g.DeltaTime
	if fc.HealingTickTimer < 1.0 {
		return
	}
	fc.HealingTickTimer = 0

	healAmt := int(healPerSec)
	if healAmt <= 0 {
		return
	}

	g.player.HP += healAmt
	if g.player.HP > g.player.MaxHP {
		g.player.HP = g.player.MaxHP
	}

	g.HealNumbers = append(g.HealNumbers, entities.DamageNumber{
		X:        g.player.MoveController.InterpX,
		Y:        g.player.MoveController.InterpY,
		Value:    healAmt,
		Ticks:    0,
		MaxTicks: 40,
	})
}

func (g *Game) castFractalCanopy(centerX, centerY float64, c *spells.Caster) {
	info := spells.SpellInfo{Name: "fractalcanopy", Level: 1, Cooldown: 5.0, Damage: 0}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	fc := &spells.FractalCanopy{
		MaxGrowTime: 5,
		MaxDuration: 10,
		MaxRadius:   5,
		HealingMin:  3,
		HealingMax:  15,
		X:           centerX,
		Y:           centerY,
		Visual:      spells.NewFractalCanopyVisual(centerX, centerY, 10),
	}
	g.ActiveSpells = append(g.ActiveSpells, fc)
}
