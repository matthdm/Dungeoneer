package game

import (
	"math"
	"sort"

	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/spells"

	"github.com/hajimehoshi/ebiten/v2"
)

// updateSpells is responsible for:
// - advancing each active spell instance
// - applying on-hit/aoe effects (fireball, lightning, fractal nodes, canopy healing)
// - adopting spawned children (storm → strikes, bloom → nodes)
// - pruning finished spells
//
// IMPORTANT: While this handles per-instance updates, do NOT also call
// g.SpellCtrl.Update(...) for instances elsewhere or you'll double-tick.
// Keep the controller for cooldown timing + HUD + slotting.
func (g *Game) updateSpells() {
	if g.SpellCtrl == nil {
		return
	}
	var remaining []spells.Spell

	for _, sp := range g.SpellCtrl.Active {
		// 1) advance
		sp.Update(g.currentLevel, g.DeltaTime)

		// 2) spell-specific side-effects
		if fb, ok := sp.(*spells.Fireball); ok {
			if !fb.Impact {
				for _, m := range g.Monsters {
					if m.IsDead {
						continue
					}
					dx := m.InterpX - fb.X
					dy := m.InterpY - fb.Y
					if dx*dx+dy*dy <= fb.Radius*fb.Radius {
						fb.Impact = true
						tx := int(math.Floor(fb.X))
						ty := int(math.Floor(fb.Y))
						g.applyFireballDamage(fb, tx, ty)
						break
					}
				}
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

		// 3) keep if still alive
		if !sp.IsFinished() {
			remaining = append(remaining, sp)
		}
	}

	// write back
	g.SpellCtrl.Active = remaining
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

// --- Casting helpers that your input can call directly.
// These now append into the controller's Active list and use either the
// passed-in caster or the controller's caster. ---

func (g *Game) castFireball(casterX, casterY, targetX, targetY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "fireball", Level: 1, Cooldown: 1.0, Damage: 5} // :contentReference[oaicite:5]{index=5}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	fb := spells.NewFireball(info, casterX, casterY, targetX, targetY, g.fireballSprites, g.spriteSheet.FireBurst) // :contentReference[oaicite:6]{index=6}
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, fb)
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
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "chaosray", Level: 1, Cooldown: 1.0, Damage: 8} // :contentReference[oaicite:7]{index=7}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	cr := spells.NewChaosRay(info, casterX, casterY, targetX, targetY) // :contentReference[oaicite:8]{index=8}
	g.applyChaosRayDamage(cr)
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, cr)
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
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "lightning", Level: 1, Cooldown: 0.01, Damage: 8} // :contentReference[oaicite:9]{index=9}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	ls := spells.NewLightningStrike(info, targetX, targetY, g.spriteSheet.ArcaneBurst) // :contentReference[oaicite:10]{index=10}
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, ls)
}

func (g *Game) castLightningStorm(centerX, centerY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "lightningstorm", Level: 1, Cooldown: 3.0, Damage: 8} // :contentReference[oaicite:11]{index=11}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)

	// Target from current hover (tile-aligned)
	tx := float64(g.hoverTileX)
	ty := float64(g.hoverTileY)

	storm := spells.NewLightningStorm(info, tx, ty, 3, 0.2, 3.0, c, g.spriteSheet.ArcaneBurst, g.currentLevel) // :contentReference[oaicite:12]{index=12}
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, storm)
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
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "fractalbloom", Level: 1, Cooldown: 4.0, Damage: 6} // :contentReference[oaicite:13]{index=13}
	if c == nil {
		c = g.SpellCtrl.Caster
	}
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)
	bloom := spells.NewFractalBloom(info, centerX, centerY, c, g.spriteSheet.ArcaneBurst, g.currentLevel, 3, 0.7, 0.2) // :contentReference[oaicite:14]{index=14}
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, bloom)
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

func (g *Game) castFractalCanopy(centerX, centerY float64, c *spells.Caster) {
	if g.SpellCtrl == nil {
		return
	}
	info := spells.SpellInfo{Name: "fractalcanopy", Level: 1, Cooldown: 5.0, Damage: 0} // :contentReference[oaicite:15]{index=15}
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
		Visual:      spells.NewFractalCanopyVisual(centerX, centerY, 10), // :contentReference[oaicite:16]{index=16}
	}
	g.SpellCtrl.Active = append(g.SpellCtrl.Active, fc)
}

// mapTomeNameToSpell returns the spell registry key for a tome item name.
func mapTomeNameToSpell(tome string) string {
	switch tome {
	case "Red Tome":
		return "fireball"
	case "Teal Tome":
		return "lightning"
	case "Crypt Tome":
		return "chaosray"
	case "Blue Tome":
		return "fractalbloom"
	case "Verdant Tome":
		return "fractalcanopy"
	default:
		return ""
	}
}

// syncEquippedTomesToSpells mirrors the player's equipped tomes (any 5) to the spell controller.
func (g *Game) syncEquippedTomesToSpells() {
	if g.player == nil || g.player.Equipment == nil || g.SpellCtrl == nil || g.HUD == nil {
		return
	}

	// 1) Collect up to 5 tome-derived spell names from Equipment in a stable order.
	// We don't rely on a specific slot name; any equipped tome counts.
	keys := make([]string, 0, len(g.player.Equipment))
	for k := range g.player.Equipment {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	equippedSpells := make([]string, 0, 5)
	for _, k := range keys {
		if it := g.player.Equipment[k]; it != nil {
			if name := mapTomeNameToSpell(it.Name); name != "" {
				equippedSpells = append(equippedSpells, name)
				if len(equippedSpells) == 5 {
					break
				}
			}
		}
	}

	// 2) Revoke any item-granted spells no longer present.
	//    (Anything known as SourceItem but not in equippedSpells gets revoked.)
	isEquipped := map[string]bool{}
	for _, n := range equippedSpells {
		isEquipped[n] = true
	}
	// controller.known is private, so revoke defensively by checking every slot + “expected” map
	// Start by clearing all controller slots, then re-equip from equippedSpells.
	for i := 0; i < spells.SlotCount; i++ {
		g.SpellCtrl.Unequip(i)
	}
	// Revoke any previously item-granted spell that isn't equipped now.
	// We can discover candidates by asking every registry entry we might have mapped from tomes:
	for _, candidate := range []string{"fireball", "lightning", "chaosray", "fractalbloom", "fractalcanopy"} {
		if !isEquipped[candidate] {
			g.SpellCtrl.RevokeFromItem(candidate)
		}
	}

	// 3) Ensure everything currently equipped is granted from item + equip into slots 0..N
	for i, name := range equippedSpells {
		g.SpellCtrl.GrantFromItem(name)
		_ = g.SpellCtrl.Equip(i, name)
	}

	// 4) Mirror controller → HUD (icons + cooldown)
	//    Read each equipped slot’s SpellDef and the caster cooldowns.
	for i := 0; i < 5; i++ {
		var icon *ebiten.Image
		var cd float64

		if i < spells.SlotCount {
			slot := g.SpellCtrl.Slot(i) // add a tiny helper if you don’t have one; see below
			if slot != "" {
				if def := g.SpellCtrl.Def(slot); def != nil {
					icon = def.Icon
					// cooldown fraction is shown inside HUD; HUD expects raw seconds remaining.
					cd = g.SpellCtrl.Caster.Cooldowns[def.Name]
				}
			}
		}
		g.HUD.SkillSlots[i].Icon = icon
		g.HUD.SkillSlots[i].Cooldown = cd
	}
}
