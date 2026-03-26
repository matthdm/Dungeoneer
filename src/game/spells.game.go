package game

import (
	"math"

	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/hud"
	"dungeoneer/items"
	"dungeoneer/spells"

	"github.com/hajimehoshi/ebiten/v2"
)

// syncHUDSpellSlots updates the HUD skill bar from the player's current SpellSlots.
func (g *Game) syncHUDSpellSlots() {
	if g.HUD == nil || g.player == nil {
		return
	}
	for i := range g.HUD.SkillSlots {
		if i < len(g.player.SpellSlots) {
			abilityID := g.player.SpellSlots[i]
			cost := spellManaCost(abilityID)
			g.HUD.SkillSlots[i].Active = true
			g.HUD.SkillSlots[i].ManaCost = cost
			g.HUD.SkillSlots[i].Enabled = g.player.Mana >= cost
			g.HUD.SkillSlots[i].Name = abilityID
			g.HUD.SkillSlots[i].Icon = g.abilityIcon(abilityID)
			// Sync cooldown from caster.
			if g.player.Caster != nil {
				g.HUD.SkillSlots[i].Cooldown = g.player.Caster.Cooldowns[abilitySpellName(abilityID)]
			}
		} else {
			g.HUD.SkillSlots[i] = hud.SkillSlot{}
		}
	}
}

// abilityIcon returns the icon image for an ability, looking up the first item
// in the registry that grants this ability. Falls back to SpriteMap.
func (g *Game) abilityIcon(abilityID string) *ebiten.Image {
	for _, tmpl := range items.Registry {
		if tmpl.GrantsAbility == abilityID && tmpl.Icon != nil {
			return tmpl.Icon
		}
	}
	// Fallback: use existing tome icons by ability name.
	fallback := map[string]string{
		"fireball":        "Red Tome",
		"chaos_ray":       "Teal Tome",
		"lightning":        "Blue Tome",
		"lightning_storm":  "Verdant Tome",
		"fractal_bloom":    "Crypt Tome",
		"fractal_canopy":   "Verdant Tome",
	}
	if name, ok := fallback[abilityID]; ok {
		for _, tmpl := range items.Registry {
			if tmpl.Name == name && tmpl.Icon != nil {
				return tmpl.Icon
			}
		}
	}
	return nil
}

// abilitySpellName maps an ability ID to the spell Name used in Caster.Cooldowns.
func abilitySpellName(abilityID string) string {
	switch abilityID {
	case "fireball":
		return "fireball"
	case "chaos_ray":
		return "chaosray"
	case "lightning":
		return "lightning"
	case "lightning_storm":
		return "lightningstorm"
	case "fractal_bloom":
		return "fractalbloom"
	case "fractal_canopy":
		return "fractalcanopy"
	case "arcane_spray":
		return "arcane_spray"
	case "arcane_bolt":
		return "arcane_bolt"
	default:
		return abilityID
	}
}

func (g *Game) updateSpells() {
	var remaining []spells.Spell
	for _, sp := range g.ActiveSpells {
		sp.Update(g.currentLevel, g.DeltaTime)
		if fb, ok := sp.(*spells.Fireball); ok {
			if !fb.Impact {
				if fb.MonsterCast {
					// Monster-cast fireball: check player collision.
					if g.player != nil && !g.player.IsDead {
						dx := g.player.MoveController.InterpX - fb.X
						dy := g.player.MoveController.InterpY - fb.Y
						if dx*dx+dy*dy <= fb.Radius*fb.Radius {
							fb.Impact = true
							g.player.TakeDamage(fb.Info.Damage)
						}
					}
				} else {
					// Player-cast fireball: check monster collision.
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
		if ab, ok := sp.(*spells.ArcaneBolt); ok {
			g.checkArcaneBoltHits(ab)
		}
		if !sp.IsFinished() {
			remaining = append(remaining, sp)
		}
	}
	g.ActiveSpells = remaining
	g.updateChanneledSpray()
}

// updateChanneledSpray handles the per-frame update for the active arcane spray:
// tracks cursor, drains mana, applies damage ticks, stops when mana runs out.
func (g *Game) updateChanneledSpray() {
	spray := g.ActiveSpray
	if spray == nil || !spray.Channeling {
		g.ActiveSpray = nil
		return
	}
	if g.player == nil {
		spray.StopChannel()
		g.ActiveSpray = nil
		return
	}

	// Update origin and direction to follow the player/cursor.
	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY
	spray.UpdateChannel(px, py, float64(g.hoverTileX), float64(g.hoverTileY))

	// Drain mana.
	drain := spray.ManaDrain * g.DeltaTime
	if !g.InfMana {
		g.player.Mana -= int(math.Ceil(drain))
	}
	if g.player.Mana <= 0 && !g.InfMana {
		g.player.Mana = 0
		spray.StopChannel()
		g.ActiveSpray = nil
		return
	}

	// Damage tick: apply every 0.25s.
	spray.DmgTimer += g.DeltaTime
	if spray.DmgTimer >= 0.25 {
		spray.DmgTimer -= 0.25
		for _, m := range g.Monsters {
			if m.IsDead {
				continue
			}
			if spray.IsInCone(m.InterpX, m.InterpY) {
				if m.TakeDamage(spray.Info.Damage, &g.HitMarkers, &g.DamageNumbers) {
					g.handleMonsterDeath(m)
				}
			}
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
                                       g.handleMonsterDeath(m)
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

// spellManaCost returns the mana cost for a spell ability ID.
func spellManaCost(abilityID string) int {
	switch abilityID {
	case "fireball":
		return 8
	case "chaos_ray":
		return 12
	case "lightning":
		return 6
	case "lightning_storm":
		return 25
	case "fractal_bloom":
		return 20
	case "fractal_canopy":
		return 15
	case "arcane_spray":
		return 5
	default:
		return 0
	}
}

// castSpellSlot dispatches a spell cast for the given spell bar index (0-5).
// The slot maps to player.SpellSlots[index], which is populated by equipped items.
func (g *Game) castSpellSlot(index int) {
	if g.player == nil || index < 0 || index >= len(g.player.SpellSlots) {
		return
	}
	abilityID := g.player.SpellSlots[index]
	cost := spellManaCost(abilityID)
	if g.player.Mana < cost {
		return
	}

	gx := g.player.MoveController.InterpX
	gy := g.player.MoveController.InterpY
	tx := float64(g.hoverTileX)
	ty := float64(g.hoverTileY)
	c := g.player.Caster

	cast := false
	switch abilityID {
	case "fireball":
		cast = g.tryCastFireball(gx, gy, tx, ty, c)
	case "chaos_ray":
		cast = g.tryCastChaosRay(gx, gy, tx, ty, c)
	case "lightning":
		cast = g.tryCastLightningStrike(tx, ty, c)
	case "lightning_storm":
		cast = g.tryCastLightningStorm(tx, ty, c)
	case "fractal_bloom":
		cast = g.tryCastFractalBloom(tx, ty, c)
	case "fractal_canopy":
		cast = g.tryCastFractalCanopy(tx, ty, c)
	case "arcane_spray":
		cast = g.tryCastArcaneSpray(gx, gy, tx, ty, c)
	}

	if cast {
		g.player.Mana -= cost
	}
}

// handlePrimaryAttack dispatches left-click based on the player's primary ability.
// tx, ty are the cursor position in fractional cartesian space.
// cx, cy are the snapped tile coords (for fallback melee).
func (g *Game) handlePrimaryAttack(tx, ty float64, cx, cy int) {
	if g.player == nil {
		return
	}

	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY

	switch {
	case g.player.HasAbility("slash_combo"):
		g.handleSlashCombo(px, py, tx, ty)
	case g.player.HasAbility("arcane_bolt"):
		g.handleArcaneBolt(px, py, tx, ty)
	default:
		// Fallback: basic click-on-enemy melee (no ability needed).
		g.handleBasicMelee(cx, cy)
	}
}

func (g *Game) handleSlashCombo(px, py, tx, ty float64) {
	if !g.player.CanAttack() {
		return
	}
	// Direction from player to cursor in cartesian space.
	dirAngle := math.Atan2(ty-py, tx-px)

	hit := g.player.ComboHit
	if hit > 2 {
		hit = 0
	}

	info := spells.SpellInfo{
		Name: "slash_combo", Level: 1,
		Cooldown: spells.SlashComboHits[hit].SweepTime + spells.SlashComboHits[hit].FadeTime,
		Damage:   g.player.Damage,
	}
	c := g.player.Caster
	if !c.Ready(info) {
		return
	}
	c.PutOnCooldown(info)

	slash := spells.NewSlashArc(info, px, py, dirAngle, hit)
	g.ActiveSpells = append(g.ActiveSpells, slash)
	g.applySlashDamage(slash)

	// Advance combo.
	g.player.ComboHit = (hit + 1) % 3
	g.player.ComboTimer = 0.5
	g.player.AttackTick = 0
}

func (g *Game) applySlashDamage(slash *spells.SlashArc) {
	dmg := int(float64(slash.Info.Damage) * spells.SlashComboHits[slash.ComboHit].DamageMult)
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		if slash.IsInArc(m.InterpX, m.InterpY) {
			if m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers) {
				g.handleMonsterDeath(m)
			}
		}
	}
}

func (g *Game) handleArcaneBolt(px, py, tx, ty float64) {
	info := spells.SpellInfo{Name: "arcane_bolt", Level: 1, Cooldown: 0.3, Damage: 3, Cost: 2}
	c := g.player.Caster
	if !c.Ready(info) {
		return
	}
	if g.player.Mana < info.Cost {
		return
	}
	c.PutOnCooldown(info)
	g.player.Mana -= info.Cost

	bolt := spells.NewArcaneBolt(info, px, py, tx, ty)
	g.ActiveSpells = append(g.ActiveSpells, bolt)
}

func (g *Game) handleBasicMelee(cx, cy int) {
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		if m.TileX == cx && m.TileY == cy &&
			entities.IsAdjacentRanged(g.player.TileX, g.player.TileY, m.TileX, m.TileY, 2) &&
			g.player.CanAttack() {
			died := m.TakeDamage(g.player.Damage, &g.HitMarkers, &g.DamageNumbers)
			g.player.AttackTick = 0
			if died {
				g.handleMonsterDeath(m)
			}
		}
	}
}

func (g *Game) tryCastArcaneSpray(casterX, casterY, targetX, targetY float64, _ *spells.Caster) bool {
	// Toggle: if already channeling, stop.
	if g.ActiveSpray != nil && g.ActiveSpray.Channeling {
		g.ActiveSpray.StopChannel()
		g.ActiveSpray = nil
		return false
	}
	if g.player.Mana < 5 {
		return false
	}
	info := spells.SpellInfo{Name: "arcane_spray", Level: 1, Cooldown: 0, Damage: 3, Cost: 0}
	spray := spells.NewArcaneSpray(info, casterX, casterY, targetX, targetY)
	g.ActiveSpray = spray
	g.ActiveSpells = append(g.ActiveSpells, spray)
	return false // mana drained per-tick, not on cast
}

// Arcane bolt collision — checked each frame in updateSpells.
func (g *Game) checkArcaneBoltHits(ab *spells.ArcaneBolt) {
	if ab.Impact || ab.IsFinished() {
		return
	}
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		dx := m.InterpX - ab.X
		dy := m.InterpY - ab.Y
		if dx*dx+dy*dy <= ab.Radius*ab.Radius {
			ab.Impact = true
			ab.X = m.InterpX
			ab.Y = m.InterpY
			if m.TakeDamage(ab.Info.Damage, &g.HitMarkers, &g.DamageNumbers) {
				g.handleMonsterDeath(m)
			}
			return
		}
	}
}

// handleBlink teleports the player along a line, stopping at walls.
func (g *Game) handleBlink(px, py, tx, ty float64) {
	if g.player == nil || g.currentLevel == nil {
		return
	}
	destX, destY := spells.FindBlinkTarget(g.currentLevel, px, py, tx, ty)
	// Only blink if we'd actually move.
	if math.Hypot(destX-px, destY-py) < 0.5 {
		return
	}

	// Teleport the player.
	g.player.MoveController.Stop()
	g.player.MoveController.InterpX = destX
	g.player.MoveController.InterpY = destY
	g.player.TileX = int(math.Floor(destX))
	g.player.TileY = int(math.Floor(destY))
	g.player.CollisionBox.X = destX
	g.player.CollisionBox.Y = destY - (g.player.CollisionBox.Height / 2)

	// Spawn visual effect.
	effect := spells.NewBlinkEffect(px, py, destX, destY)
	g.ActiveSpells = append(g.ActiveSpells, effect)
}

func (g *Game) tryCastFireball(casterX, casterY, targetX, targetY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "fireball", Level: 1, Cooldown: 1.0, Damage: 5, Cost: 8}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	fb := spells.NewFireball(info, casterX, casterY, targetX, targetY, g.fireballSprites, g.spriteSheet.FireBurst)
	g.ActiveSpells = append(g.ActiveSpells, fb)
	return true
}

func (g *Game) tryCastChaosRay(casterX, casterY, targetX, targetY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "chaosray", Level: 1, Cooldown: 1.0, Damage: 8, Cost: 12}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	cr := spells.NewChaosRay(info, casterX, casterY, targetX, targetY)
	g.applyChaosRayDamage(cr)
	g.ActiveSpells = append(g.ActiveSpells, cr)
	return true
}

func (g *Game) tryCastLightningStrike(targetX, targetY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "lightning", Level: 1, Cooldown: 0.01, Damage: 8, Cost: 6}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	ls := spells.NewLightningStrike(info, targetX, targetY, g.spriteSheet.ArcaneBurst)
	g.ActiveSpells = append(g.ActiveSpells, ls)
	return true
}

func (g *Game) tryCastLightningStorm(centerX, centerY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "lightningstorm", Level: 1, Cooldown: 3.0, Damage: 8, Cost: 25}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	storm := spells.NewLightningStorm(info, centerX, centerY, 3, 0.2, 3.0, c, g.spriteSheet.ArcaneBurst, g.currentLevel)
	g.ActiveSpells = append(g.ActiveSpells, storm)
	return true
}

func (g *Game) tryCastFractalBloom(centerX, centerY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "fractalbloom", Level: 1, Cooldown: 4.0, Damage: 6, Cost: 20}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	bloom := spells.NewFractalBloom(info, centerX, centerY, c, g.spriteSheet.ArcaneBurst, g.currentLevel, 3, 0.7, 0.2)
	g.ActiveSpells = append(g.ActiveSpells, bloom)
	return true
}

func (g *Game) tryCastFractalCanopy(centerX, centerY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "fractalcanopy", Level: 1, Cooldown: 5.0, Damage: 0, Cost: 15}
	if !c.Ready(info) {
		return false
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
	return true
}

// Legacy wrappers — used by monster-cast spells and any non-player casters.
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
                                       g.handleMonsterDeath(m)
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
                                       g.handleMonsterDeath(m)
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
                                       g.handleMonsterDeath(m)
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
