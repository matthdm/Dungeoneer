package game

import (
	"math"

	"dungeoneer/audio"
	"dungeoneer/controls"
	"dungeoneer/coords"
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
			cost := g.spellManaCost(abilityID)
			g.HUD.SkillSlots[i].Active = true
			g.HUD.SkillSlots[i].ManaCost = cost
			g.HUD.SkillSlots[i].Enabled = g.player.Mana >= cost
			g.HUD.SkillSlots[i].Name = abilityID
			g.HUD.SkillSlots[i].Icon = g.abilityIcon(abilityID)
			g.HUD.SkillSlots[i].Locked = g.player.IsAbilityLocked(abilityID)
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
		"lightning":       "Blue Tome",
		"lightning_storm": "Verdant Tome",
		"fractal_bloom":   "Crypt Tome",
		"fractal_canopy":  "Verdant Tome",
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
		prevX, prevY := 0.0, 0.0
		isArcaneBolt := false
		if ab, ok := sp.(*spells.ArcaneBolt); ok && ab != nil {
			isArcaneBolt = true
			prevX, prevY = ab.X, ab.Y
		}

		sp.Update(g.currentLevel, g.DeltaTime)
		if fb, ok := sp.(*spells.Fireball); ok {
			if !fb.Impact {
				if fb.MonsterCast {
					// Monster-cast fireball: check player collision.
					if g.player != nil && !g.player.IsDead {
						dx := g.player.BodyX() - fb.X
						dy := g.player.BodyY() - fb.Y
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

						dx := m.BodyX() - fb.X
						dy := m.BodyY() - fb.Y
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
			g.checkArcaneBoltHits(ab, prevX, prevY, isArcaneBolt)
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
		g.sprayManaDrainAcc = 0
		g.ActiveSpray = nil
		return
	}
	if g.player == nil {
		spray.StopChannel()
		g.sprayManaDrainAcc = 0
		g.ActiveSpray = nil
		return
	}

	// Update origin and direction to follow the player/cursor.
	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY
	spray.UpdateChannel(px, py, float64(g.hoverTileX), float64(g.hoverTileY))

	// Stop spray when the player releases the spell key.
	sprayHeld := false
	spellActions := []controls.ActionID{
		controls.ActionSpell1, controls.ActionSpell2, controls.ActionSpell3,
		controls.ActionSpell4, controls.ActionSpell5, controls.ActionSpell6,
	}
	for i, action := range spellActions {
		if i < len(g.player.SpellSlots) && g.player.SpellSlots[i] == "arcane_spray" {
			if g.isActionPressed(action) {
				sprayHeld = true
				break
			}
		}
	}
	if !sprayHeld {
		spray.StopChannel()
		g.sprayManaDrainAcc = 0
		g.ActiveSpray = nil
		return
	}

	// Drain mana using an accumulator so per-second drain is stable across frames.
	if !g.InfMana {
		g.sprayManaDrainAcc += spray.ManaDrain * g.DeltaTime
		if g.sprayManaDrainAcc >= 1 {
			spent := int(g.sprayManaDrainAcc)
			g.player.Mana -= spent
			g.sprayManaDrainAcc -= float64(spent)
		}
	}
	if g.player.Mana <= 0 && !g.InfMana {
		g.player.Mana = 0
		spray.StopChannel()
		g.sprayManaDrainAcc = 0
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
			if spray.IsInCone(m.BodyX(), m.BodyY()) {
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
				if spells.OnSpellImpact != nil {
					spells.OnSpellImpact(m.InterpX, m.InterpY, "fireball")
				}
			}
		}
	}
	// Also emit at the impact center even if no monsters were hit.
	if spells.OnSpellImpact != nil {
		spells.OnSpellImpact(float64(cx), float64(cy), "fireball")
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

func (g *Game) getSpellDamage(baseDmg int) int {
	if g.player == nil {
		return baseDmg
	}
	intBonus := g.player.EffectiveStats().Intelligence / 2
	return baseDmg + intBonus
}

func (g *Game) getSpellCooldown(baseCD float64) float64 {
	if g.player == nil {
		return baseCD
	}
	itemCDR := items.EvalPassiveCooldownReduction(g.player.Equipment)
	dexCDR := float64(g.player.EffectiveStats().Dexterity) * 0.01
	if dexCDR > 0.30 {
		dexCDR = 0.30
	}
	cdr := itemCDR + dexCDR
	if cdr > 0.80 {
		cdr = 0.80
	}
	return baseCD * (1.0 - cdr)
}

// spellManaCost returns the mana cost for a spell ability ID, scaling down with item effects.
func (g *Game) spellManaCost(abilityID string) int {
	var baseCost int
	switch abilityID {
	case "fireball":
		baseCost = 8
	case "chaos_ray":
		baseCost = 12
	case "lightning":
		baseCost = 6
	case "lightning_storm":
		baseCost = 25
	case "fractal_bloom":
		baseCost = 20
	case "fractal_canopy":
		baseCost = 15
	case "arcane_spray":
		baseCost = 5
	default:
		baseCost = 0
	}
	if g.player == nil {
		return baseCost
	}
	reduction := items.EvalPassiveManaCostReduction(g.player.Equipment)
	cost := int(float64(baseCost) * (1.0 - reduction))
	if cost < 0 {
		cost = 0
	}
	return cost
}

// castSpellSlot dispatches a spell cast for the given spell bar index (0-5).
// The slot maps to player.SpellSlots[index], which is populated by equipped items.
func (g *Game) castSpellSlot(index int) {
	if g.player == nil || index < 0 || index >= len(g.player.SpellSlots) {
		return
	}
	abilityID := g.player.SpellSlots[index]
	if g.player.IsAbilityLocked(abilityID) {
		return
	}
	cost := g.spellManaCost(abilityID)
	// Channeled spray startup cost check is done inside tryCastArcaneSpray.
	if abilityID != "arcane_spray" && g.player.Mana < cost {
		return
	}

	// Cast from the player's body center so projectiles and rays visually
	// originate from the sprite rather than the invisible feet anchor.
	gx := g.player.BodyX()
	gy := g.player.BodyY()
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
		if g.BehaviorTracker != nil {
			g.BehaviorTracker.RecordSpellCast(abilityID)
		}
		if g.Audio != nil {
			switch abilityID {
			case "fireball":
				g.Audio.PlaySFX(audio.SFXSpellFireball)
			case "lightning", "lightning_storm":
				g.Audio.PlaySFX(audio.SFXSpellLightning)
			case "chaos_ray":
				g.Audio.PlaySFX(audio.SFXSpellChaos)
			default:
				g.Audio.PlaySFX(audio.SFXSpellGeneric)
			}
		}
	}
}

// handlePrimaryAttack dispatches left-click based on the player's primary ability.
// tx, ty are the cursor position in fractional cartesian space.
// cx, cy are the snapped tile coords (for fallback melee).
func (g *Game) handlePrimaryAttack(tx, ty float64, cx, cy int) {
	if g.player == nil {
		return
	}

	switch {
	case g.player.HasAbility("slash_combo"):
		if g.player.IsAbilityLocked("slash_combo") {
			return
		}
		g.handleSlashCombo(tx, ty)
	case g.player.HasAbility("arcane_bolt"):
		if g.player.IsAbilityLocked("arcane_bolt") {
			return
		}
		g.handleArcaneBolt(tx, ty)
	default:
		// Fallback: basic click-on-enemy melee (no ability needed).
		g.handleBasicMelee(cx, cy)
	}
}

func (g *Game) handleSlashCombo(tx, ty float64) {
	// Use the player's body center as the arc origin so the detection radius
	// is measured from the same point the visual arc is drawn from. Using feet
	// (px, py) made the effective origin ~1 tile away from the sprite body,
	// causing the monster's body center to fall outside the radius even when
	// visually adjacent.
	bx := g.player.BodyX()
	by := g.player.BodyY()
	dirAngle := math.Atan2(ty-by, tx-bx)

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

	slash := spells.NewSlashArc(info, bx, by, dirAngle, hit)
	g.ActiveSpells = append(g.ActiveSpells, slash)
	g.applySlashDamage(slash)

	// Advance combo.
	g.player.ComboHit = (hit + 1) % 3
	g.player.ComboTimer = 0.5
	g.player.AttackTick = 0
}

func (g *Game) applySlashDamage(slash *spells.SlashArc) {
	dmg := int(float64(slash.Info.Damage) * spells.SlashComboHits[slash.ComboHit].DamageMult)
	critMult, isCrit := items.EvalOnHitCrit(g.player.Equipment)
	if isCrit {
		dmg = int(float64(dmg) * critMult)
	}
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		if slash.IsInArc(m.BodyX(), m.BodyY()) {
			if m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers) {
				g.handleMonsterDeath(m)
			}
			if spells.OnSpellImpact != nil {
				spells.OnSpellImpact(m.InterpX, m.InterpY, "arcane")
			}
		}
	}
}

func (g *Game) handleArcaneBolt(tx, ty float64) {
	reduction := items.EvalPassiveManaCostReduction(g.player.Equipment)
	cost := int(float64(2) * (1.0 - reduction))
	if cost < 0 {
		cost = 0
	}
	info := spells.SpellInfo{Name: "arcane_bolt", Level: 1, Cooldown: g.getSpellCooldown(0.3), Damage: g.getSpellDamage(3), Cost: cost}
	c := g.player.Caster
	if !c.Ready(info) {
		return
	}
	if g.player.Mana < info.Cost {
		return
	}
	c.PutOnCooldown(info)
	g.player.Mana -= info.Cost

	// Emit from the player's body center so the bolt travels from the
	// character's visual position, not the feet anchor.
	bx := g.player.BodyX()
	by := g.player.BodyY()
	bolt := spells.NewArcaneBolt(info, bx, by, tx, ty)
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
			dmg := g.player.Damage
			critMult, isCrit := items.EvalOnHitCrit(g.player.Equipment)
			if isCrit {
				dmg = int(float64(dmg) * critMult)
			}
			died := m.TakeDamage(dmg, &g.HitMarkers, &g.DamageNumbers)
			g.player.AttackTick = 0
			if died {
				g.handleMonsterDeath(m)
			}
		}
	}
}

func (g *Game) tryCastArcaneSpray(casterX, casterY, targetX, targetY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "arcane_spray", Level: 1, Cooldown: g.getSpellCooldown(0.15), Damage: g.getSpellDamage(3), Cost: g.spellManaCost("arcane_spray")}

	// Already channeling — key is held, updateChanneledSpray handles it.
	if g.ActiveSpray != nil && g.ActiveSpray.Channeling {
		return false
	}
	if !c.Ready(info) {
		return false
	}
	if !g.InfMana && g.player.Mana < info.Cost {
		return false
	}
	c.PutOnCooldown(info)
	spray := spells.NewArcaneSpray(info, casterX, casterY, targetX, targetY)
	spray.ManaDrain = float64(info.Cost)
	g.sprayManaDrainAcc = 0
	g.ActiveSpray = spray
	g.ActiveSpells = append(g.ActiveSpells, spray)
	return true // upfront slot cost + ongoing per-second channel drain
}

// Arcane bolt collision — checked each frame in updateSpells.
func (g *Game) checkArcaneBoltHits(ab *spells.ArcaneBolt, prevX, prevY float64, hasPrev bool) {
	if ab.IsFinished() {
		return
	}
	// If already in post-impact animation and the bolt did not advance this frame,
	// do not re-apply collision.
	if ab.Impact && (!hasPrev || (math.Abs(ab.X-prevX) < 1e-6 && math.Abs(ab.Y-prevY) < 1e-6)) {
		return
	}

	segStartX, segStartY := ab.X, ab.Y
	if hasPrev {
		segStartX, segStartY = prevX, prevY
	}
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		hitX := m.BodyX()
		hitY := m.BodyY()
		r := ab.Radius + m.HitRadius
		if pointSegmentDistance(hitX, hitY, segStartX, segStartY, ab.X, ab.Y) <= r {
			ab.Impact = true
			ab.X = hitX
			ab.Y = hitY
			if m.TakeDamage(ab.Info.Damage, &g.HitMarkers, &g.DamageNumbers) {
				g.handleMonsterDeath(m)
			}
			if spells.OnSpellImpact != nil {
				spells.OnSpellImpact(m.InterpX, m.InterpY, "arcane_bolt")
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
	info := spells.SpellInfo{Name: "fireball", Level: 1, Cooldown: g.getSpellCooldown(1.0), Damage: g.getSpellDamage(5), Cost: g.spellManaCost("fireball")}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	fb := spells.NewFireball(info, casterX, casterY, targetX+coords.BodyDX, targetY+coords.BodyDY, g.fireballSprites, g.spriteSheet.FireBurst)
	g.ActiveSpells = append(g.ActiveSpells, fb)
	return true
}

func (g *Game) tryCastChaosRay(casterX, casterY, targetX, targetY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "chaosray", Level: 1, Cooldown: g.getSpellCooldown(1.0), Damage: g.getSpellDamage(8), Cost: g.spellManaCost("chaos_ray")}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	cr := spells.NewChaosRay(info, casterX, casterY, targetX+coords.BodyDX, targetY+coords.BodyDY)
	g.applyChaosRayDamage(cr)
	g.ActiveSpells = append(g.ActiveSpells, cr)
	return true
}

func (g *Game) tryCastLightningStrike(targetX, targetY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "lightning", Level: 1, Cooldown: g.getSpellCooldown(0.01), Damage: g.getSpellDamage(8), Cost: g.spellManaCost("lightning")}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	ls := spells.NewLightningStrike(info, targetX, targetY, g.spriteSheet.ArcaneBurst)
	g.ActiveSpells = append(g.ActiveSpells, ls)
	return true
}

func (g *Game) tryCastLightningStorm(centerX, centerY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "lightningstorm", Level: 1, Cooldown: g.getSpellCooldown(3.0), Damage: g.getSpellDamage(8), Cost: g.spellManaCost("lightning_storm")}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	storm := spells.NewLightningStorm(info, centerX, centerY, 3, 0.2, 3.0, c, g.spriteSheet.ArcaneBurst, g.currentLevel)
	g.ActiveSpells = append(g.ActiveSpells, storm)
	return true
}

func (g *Game) tryCastFractalBloom(centerX, centerY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "fractalbloom", Level: 1, Cooldown: g.getSpellCooldown(4.0), Damage: g.getSpellDamage(6), Cost: g.spellManaCost("fractal_bloom")}
	if !c.Ready(info) {
		return false
	}
	c.PutOnCooldown(info)
	bloom := spells.NewFractalBloom(info, centerX, centerY, c, g.spriteSheet.ArcaneBurst, g.currentLevel, 3, 0.7, 0.2)
	g.ActiveSpells = append(g.ActiveSpells, bloom)
	return true
}

func (g *Game) tryCastFractalCanopy(centerX, centerY float64, c *spells.Caster) bool {
	info := spells.SpellInfo{Name: "fractalcanopy", Level: 1, Cooldown: g.getSpellCooldown(5.0), Damage: 0, Cost: g.spellManaCost("fractal_canopy")}
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


func (g *Game) applyChaosRayDamage(cr *spells.ChaosRay) {
	radius := 0.6
	for _, m := range g.Monsters {
		if m.IsDead {
			continue
		}
		px := m.BodyX()
		py := m.BodyY()
		for i := 0; i < len(cr.Path)-1; i++ {
			p1 := cr.Path[i]
			p2 := cr.Path[i+1]
			if pointSegmentDistance(px, py, p1.X, p1.Y, p2.X, p2.Y) <= radius {
				if m.TakeDamage(cr.Info.Damage, &g.HitMarkers, &g.DamageNumbers) {
					g.handleMonsterDeath(m)
				}
				if spells.OnSpellImpact != nil {
					spells.OnSpellImpact(m.InterpX, m.InterpY, "chaos_ray")
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
	if spells.OnSpellImpact != nil {
		spells.OnSpellImpact(float64(cx), float64(cy), "lightning")
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
					g.handleMonsterDeath(m)
				}
			}
		}
	}
	if spells.OnSpellImpact != nil {
		spells.OnSpellImpact(float64(cx), float64(cy), "nature")
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

