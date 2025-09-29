package game

import (
	"fmt"
	"math"

	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/items"
	"dungeoneer/levels"
	"dungeoneer/spells"
)

// configureTomeItems rewires tome templates so equipping them keeps the spell
// controller and HUD in sync.
func configureTomeItems() {
	for name, info := range items.TomeSpells {
		tmpl, ok := items.Registry[name]
		if !ok || tmpl == nil {
			continue
		}
		tmpl.Description = fmt.Sprintf("Grants the %s spell.", info.SpellName)
		tmpl.OnEquip = func(p interface{}) {
			if loadout, ok := p.(items.SpellLoadout); ok {
				loadout.RebuildSpellLoadout()
			}
		}
		tmpl.OnUnequip = func(p interface{}) {
			if loadout, ok := p.(items.SpellLoadout); ok {
				loadout.RebuildSpellLoadout()
			}
		}
	}
}

// ResolveSpellEffects satisfies spells.EffectResolver. It applies damage,
// healing, and other side-effects exactly once per frame after the controller
// advances each spell.
func (g *Game) ResolveSpellEffects(sp spells.Spell, level *levels.Level) {
	switch s := sp.(type) {
	case *spells.Fireball:
		if !s.Impact {
			for _, m := range g.Monsters {
				if m.IsDead {
					continue
				}
				dx := m.InterpX - s.X
				dy := m.InterpY - s.Y
				if dx*dx+dy*dy <= s.Radius*s.Radius {
					s.Impact = true
					s.X = m.InterpX
					s.Y = m.InterpY
					break
				}
			}
		}
		if s.Impact && !s.DamageApplied {
			s.DamageApplied = true
			tx := int(math.Floor(s.X))
			ty := int(math.Floor(s.Y))
			g.applyFireballDamage(s, tx, ty)
		}
	case *spells.LightningStrike:
		if !s.DamageApplied {
			g.applyLightningDamage(s, int(math.Floor(s.X)), int(math.Floor(s.Y)))
			s.DamageApplied = true
		}
	case *spells.ChaosRay:
		g.applyChaosRayDamage(s)
	case *spells.FractalCanopy:
		g.applyFractalCanopyHealing(s)
	case *spells.FractalNode:
		if !s.DamageApplied {
			g.applyFractalDamage(s, int(math.Floor(s.X)), int(math.Floor(s.Y)))
			s.DamageApplied = true
		}
	}
}

func (g *Game) applyFireballDamage(fb *spells.Fireball, cx, cy int) {
	level := fb.Info.Level
	var radius int
	mult := 1.0
	switch level {
	case 1:
		radius = 1
		fb.ImpactImg = g.spriteSheet.FireBurst
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

func (g *Game) applyChaosRayDamage(cr *spells.ChaosRay) {
	radius := 0.6
	for idx, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		if cr.HitIndices[idx] {
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
				cr.HitIndices[idx] = true
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
	frac := fc.Age / fc.MaxGrowTime
	if frac > 1 {
		frac = 1
	}
	healPerSec := fc.HealingMin + (fc.HealingMax-fc.HealingMin)*frac
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
		X: g.player.MoveController.InterpX, Y: g.player.MoveController.InterpY,
		Value: healAmt, Ticks: 0, MaxTicks: 40,
	})
}

// Casting helpers used by debug tools and menus.
func (g *Game) castFireball(casterX, casterY, targetX, targetY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "fireball", DisplayName: "Fireball", Level: 1, Cooldown: 1.0, Damage: 5}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	fb := spells.NewFireball(info, casterX, casterY, targetX, targetY, g.fireballSprites, g.spriteSheet.FireBurst)
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, fb)
}

func (g *Game) castChaosRay(casterX, casterY, targetX, targetY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "chaosray", DisplayName: "Chaos Ray", Level: 1, Cooldown: 1.0, Damage: 8}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	cr := spells.NewChaosRay(info, casterX, casterY, targetX, targetY)
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, cr)
}

func (g *Game) castLightningStrike(targetX, targetY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "lightning", DisplayName: "Lightning", Level: 1, Cooldown: 0.01, Damage: 8}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	ls := spells.NewLightningStrike(info, targetX, targetY, g.spriteSheet.ArcaneBurst)
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, ls)
}

func (g *Game) castLightningStorm(centerX, centerY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "lightningstorm", DisplayName: "Lightning Storm", Level: 1, Cooldown: 3.0, Damage: 8}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	tx := float64(g.hoverTileX)
	ty := float64(g.hoverTileY)
	storm := spells.NewLightningStorm(info, tx, ty, 3, 0.2, 3.0, c, g.spriteSheet.ArcaneBurst, g.currentLevel)
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, storm)
}

func (g *Game) castFractalBloom(centerX, centerY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "fractalbloom", DisplayName: "Fractal Bloom", Level: 1, Cooldown: 4.0, Damage: 6}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	bloom := spells.NewFractalBloom(info, centerX, centerY, c, g.spriteSheet.ArcaneBurst, g.currentLevel, 3, 0.7, 0.2)
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, bloom)
}

func (g *Game) castFractalCanopy(centerX, centerY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "fractalcanopy", DisplayName: "Fractal Canopy", Level: 1, Cooldown: 5.0, Damage: 0}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
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
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, fc)
}

// refreshSpellHUD mirrors controller cooldown state to the HUD each frame.
func (g *Game) refreshSpellHUD() {
	if g.SpellCtrl == nil || g.HUD == nil {
		return
	}
	icons, remaining, totals, infos := g.SpellCtrl.SlotsForHUD()
	g.HUD.SetSkillSlots(icons, remaining, totals, infos)
}
