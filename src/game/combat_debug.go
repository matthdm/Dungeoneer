package game

import (
	"dungeoneer/combat"
	"dungeoneer/entities"
	"dungeoneer/images"
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// devLoadBuild applies a combat.Scenario to the live player so the new combat
// engine sees the correct artifact loadout and base stats. Artifact IDs are
// written directly into SpellSlots (slots 0-5) and Meta.ArtifactLoadout[6],
// mirroring the benchmarker's approach so ArtifactEffects lookups resolve.
func (g *Game) devLoadBuild(s combat.Scenario) {
	if g.player == nil {
		return
	}

	// 1. Class + sprite.
	switch s.Class {
	case "mage":
		g.player.Class = entities.ClassMage
		if img, err := images.LoadEmbeddedImage(images.Black_Mage_Full_png); err == nil {
			g.player.Sprite = img
		}
	default:
		g.player.Class = entities.ClassKnight
		if g.spriteSheet != nil {
			g.player.Sprite = g.spriteSheet.GreyKnight
		}
	}

	// 2. Base stats.
	g.player.Stats.Strength = s.Stats.Strength
	g.player.Stats.Dexterity = s.Stats.Dexterity
	g.player.Stats.Vitality = s.Stats.Vitality
	g.player.Stats.Intelligence = s.Stats.Intelligence

	// 3. Clear current abilities and spell slots.
	g.player.ClearAbilities()
	g.player.SpellSlots = g.player.SpellSlots[:0]

	// 4. Write artifact IDs directly into SpellSlots[0:6] so the new combat
	// adapter can find them in ArtifactEffects by their original ID.
	for i, id := range s.Artifacts {
		if id == "" {
			continue
		}
		if i < 6 {
			g.player.SpellSlots = append(g.player.SpellSlots, id)
			g.player.Abilities[id] = true
		} else if i == 6 && g.Meta != nil {
			g.Meta.ArtifactLoadout[6] = id
		}
	}

	// 5. Recalculate derived stats (MaxHP, MaxMana, Damage, AttackRate).
	g.player.RecalculateStats()

	// 6. Restore HP and Mana to new maximums.
	g.player.HP = g.player.MaxHP
	g.player.Mana = g.player.MaxMana

	// 7. Sync HUD spell bar.
	if g.HUD != nil {
		g.syncHUDSpellSlots()
	}

	// 8. Reset combat adapter state so old cooldowns don't carry over.
	if na, ok := g.CombatAdapt.(*NewCombatAdapter); ok {
		na.combatState = combat.CombatState{}
		na.pendingActions = na.pendingActions[:0]
	}
	g.TargetedMonster = nil
	g.IsAutoAttacking = false
}

// devSpawnCombatEnemy places a single enemy near the player for combat testing.
// role is "melee", "ranged", or "elite". Stats are scaled to the provided floor.
func (g *Game) devSpawnCombatEnemy(role string, floor int) {
	if g.player == nil || g.currentLevel == nil {
		return
	}
	if g.Monsters == nil {
		g.Monsters = g.Monsters[:0]
	}

	// Pick a spawn tile 3 tiles away from the player (diagonal).
	tx := g.player.TileX + 3
	ty := g.player.TileY
	// Clamp to level bounds.
	if tx >= g.currentLevel.W {
		tx = g.player.TileX - 3
	}

	hp := 30 + floor*15
	dmg := 5 + floor*3
	name := "Training Dummy"

	switch role {
	case "ranged":
		name = "Ranged Target"
		hp = 20 + floor*10
		dmg = 4 + floor*2
	case "elite":
		name = "Elite Target"
		hp = 80 + floor*25
		dmg = 8 + floor*4
	}

	var sprite *ebiten.Image
	if g.spriteSheet != nil {
		sprite = g.spriteSheet.GreyKnight // reuse as placeholder
	}

	m := &entities.Monster{
		Name:   name,
		TileX:  tx,
		TileY:  ty,
		HP:     hp,
		MaxHP:  hp,
		Damage: dmg,
		Level:  floor,
		Sprite: sprite,
	}
	m.InterpX = float64(tx)
	m.InterpY = float64(ty)

	g.Monsters = append(g.Monsters, m)
}

// devResetPlayerState restores HP and Mana to max and clears combat conditions.
func (g *Game) devResetPlayerState() {
	if g.player == nil {
		return
	}
	g.player.HP = g.player.MaxHP
	g.player.Mana = g.player.MaxMana
	if na, ok := g.CombatAdapt.(*NewCombatAdapter); ok {
		na.combatState.InShadow = false
		na.combatState.ShadowTimer = 0
		na.combatState.TargetRooted = false
		na.combatState.RootTimer = 0
		na.combatState.TauntTimer = 0
		na.combatState.BurnActive = false
		na.combatState.BurnTimer = 0
		na.combatState.HPDrainAccum = 0
		na.combatState.PlayerHP = g.player.MaxHP
	}
	g.TargetedMonster = nil
	g.IsAutoAttacking = false
	g.KillStreak = 0
}

// drawCombatDebugOverlay renders a live CombatState panel anchored to the
// top-left of the screen (below the status effect icons). Only shown when
// g.ShowCombatDebug is true.
func (g *Game) drawCombatDebugOverlay(screen *ebiten.Image) {
	if !g.ShowCombatDebug {
		return
	}

	na, ok := g.CombatAdapt.(*NewCombatAdapter)
	if !ok {
		ebitenutil.DebugPrintAt(screen, "[COMBAT DEBUG] legacy adapter active", 8, 8)
		return
	}
	cs := na.combatState

	const (
		panelX = 8
		panelY = 8
		panelW = 260
		lineH  = 13
		lines  = 14
		padH   = 6
	)
	panelH := lines*lineH + padH*2

	// Background.
	bg := ebiten.NewImage(panelW, panelH)
	bg.Fill(color.NRGBA{R: 8, G: 8, B: 18, A: 210})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(panelX), float64(panelY))
	screen.DrawImage(bg, op)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), float32(panelW), float32(panelH), 1,
		color.NRGBA{R: 80, G: 80, B: 140, A: 200}, false)

	y := panelY + padH
	lx := panelX + 6

	active := color.NRGBA{R: 80, G: 220, B: 100, A: 255}
	inactive := color.NRGBA{R: 120, G: 120, B: 120, A: 255}
	label := color.NRGBA{R: 200, G: 200, B: 220, A: 255}

	drawLine := func(text string, col color.NRGBA) {
		ebitenutil.DebugPrintAt(screen, text, lx, y)
		_ = col // DebugPrintAt uses white; colour used for badge only
		y += lineH
	}
	drawBadge := func(text string, col color.NRGBA, val string) {
		ebitenutil.DebugPrintAt(screen, text, lx, y)
		bx := float32(panelX + panelW - 78)
		by := float32(y)
		bCol := inactive
		if col == active {
			bCol = active
		}
		vector.DrawFilledRect(screen, bx, by, 72, float32(lineH-1), color.NRGBA{R: bCol.R / 4, G: bCol.G / 4, B: bCol.B / 4, A: 200}, false)
		ebitenutil.DebugPrintAt(screen, val, int(bx)+4, int(by))
		y += lineH
	}

	_ = label
	_ = drawLine

	drawBadge("HP", active, fmt.Sprintf("%d / %d", cs.PlayerHP, cs.PlayerMaxHP))

	drainStr := "0.00"
	if cs.HPDrainAccum > 0 {
		drainStr = fmt.Sprintf("%.2f", cs.HPDrainAccum)
	}
	drainCol := inactive
	if cs.HPDrainAccum > 0 {
		drainCol = active
	}
	drawBadge("HP Drain Accum", drainCol, drainStr)

	shadowCol := inactive
	shadowVal := fmt.Sprintf("OFF  %.1fs", cs.ShadowTimer)
	if cs.InShadow {
		shadowCol = active
		shadowVal = fmt.Sprintf("ON   %.1fs", cs.ShadowTimer)
	}
	drawBadge("Shadow", shadowCol, shadowVal)

	rootCol := inactive
	rootVal := fmt.Sprintf("OFF  %.1fs", cs.RootTimer)
	if cs.TargetRooted {
		rootCol = active
		rootVal = fmt.Sprintf("ON   %.1fs", cs.RootTimer)
	}
	drawBadge("Root", rootCol, rootVal)

	tauntCol := inactive
	tauntVal := fmt.Sprintf("OFF  %.1fs", cs.TauntTimer)
	if cs.TauntTimer > 0 {
		tauntCol = active
		tauntVal = fmt.Sprintf("ON   %.1fs", cs.TauntTimer)
	}
	drawBadge("Taunt", tauntCol, tauntVal)

	burnCol := inactive
	burnVal := fmt.Sprintf("OFF  %.1fs", cs.BurnTimer)
	if cs.BurnActive {
		burnCol = active
		burnVal = fmt.Sprintf("ON   %.1fs", cs.BurnTimer)
	}
	drawBadge("Burn", burnCol, burnVal)

	atkInterval := cs.PlayerAttackInterval
	if cs.AttackSpeedPct > 0 {
		atkInterval *= 1.0 - float64(cs.AttackSpeedPct)/100.0
	}
	drawBadge("Atk Interval", active, fmt.Sprintf("%.2fs", math.Max(0.1, atkInterval)))

	drawBadge("Kill Streak", active, fmt.Sprintf("%d", cs.KillStreak))

	// Cooldown slots 0-6.
	ebitenutil.DebugPrintAt(screen, "Cooldowns:", lx, y)
	y += lineH
	cdStr := ""
	for i, cd := range cs.ArtifactCooldowns {
		id := ""
		if i < len(cs.EquippedArtifacts) {
			id = cs.EquippedArtifacts[i]
		}
		if id == "" {
			continue
		}
		// Truncate artifact ID to 12 chars for display.
		short := id
		if len(short) > 12 {
			short = short[:12]
		}
		_ = cdStr
		col := inactive
		cdVal := "ready"
		if cd > 0 {
			col = active
			cdVal = fmt.Sprintf("%.1fs", cd)
		}
		drawBadge(fmt.Sprintf("  [%d] %s", i+1, short), col, cdVal)
	}
}
