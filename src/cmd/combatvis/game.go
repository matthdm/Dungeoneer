package main

import (
	"dungeoneer/combat"
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	arenaW  = 680
	panelW  = 500
	totalW  = arenaW + panelW
	totalH  = 700
	playerX = float32(190)
	enemyX  = float32(490)
	dotY    = float32(300)

	maxLogLines = 12

	lineH = 15 // pixels per text line (DebugPrint internal font height ≈ 13px)
)

var (
	colBG        = color.RGBA{14, 14, 22, 255}
	colArenaBG   = color.RGBA{20, 20, 34, 255}
	colPanelBG   = color.RGBA{16, 16, 26, 255}
	colHeader    = color.RGBA{30, 30, 52, 255}
	colDivider   = color.RGBA{50, 50, 80, 255}
	colText      = color.RGBA{210, 210, 230, 255}
	colDim       = color.RGBA{130, 130, 160, 255}
	colGreen     = color.RGBA{80, 200, 120, 255}
	colRed       = color.RGBA{200, 70, 70, 255}
	colOrange    = color.RGBA{220, 140, 50, 255}
	colBlue      = color.RGBA{100, 160, 255, 255}
	colPurple    = color.RGBA{160, 90, 220, 255}
	colYellow    = color.RGBA{240, 210, 60, 255}
	colCDReady   = color.RGBA{70, 190, 110, 255}
	colCDActive  = color.RGBA{180, 120, 40, 255}
	colCDBar     = color.RGBA{55, 55, 80, 255}
)

// domainColor maps domain name to a color for item display.
func domainColor(domain string) color.RGBA {
	switch domain {
	case "iron":
		return color.RGBA{180, 190, 200, 255}
	case "shadow":
		return color.RGBA{130, 80, 200, 255}
	case "flame":
		return color.RGBA{220, 100, 40, 255}
	case "void":
		return color.RGBA{110, 60, 170, 255}
	case "nature":
		return color.RGBA{80, 180, 100, 255}
	case "arcane":
		return color.RGBA{80, 160, 230, 255}
	default:
		return color.RGBA{160, 160, 180, 255}
	}
}

// effectDesc produces a human-readable description of an ArtifactEffect.
func effectDesc(eff combat.ArtifactEffect) string {
	var parts []string
	if eff.IsPassive {
		if eff.HealOnKillPct > 0 {
			parts = append(parts, fmt.Sprintf("heal %d%% maxHP on kill", eff.HealOnKillPct))
		}
		if eff.DamageCapPct > 0 {
			parts = append(parts, fmt.Sprintf("cap incoming dmg at %d%% maxHP", eff.DamageCapPct))
		}
		if eff.CooldownReductionPct > 0 {
			parts = append(parts, fmt.Sprintf("-%d%% all cooldowns", eff.CooldownReductionPct))
		}
		if eff.SkillDurationPct > 0 {
			parts = append(parts, fmt.Sprintf("+%d%% skill duration", eff.SkillDurationPct))
		}
		if eff.AttackSpeedPct > 0 {
			parts = append(parts, fmt.Sprintf("+%d%% attack speed", eff.AttackSpeedPct))
		}
		if eff.ShroudCooldownReset > 0 {
			parts = append(parts, fmt.Sprintf("shadow hit: -%.0fs shroud CD", eff.ShroudCooldownReset))
		}
		if eff.BurnDPS > 0 {
			parts = append(parts, fmt.Sprintf("auto hits: burn %d dps for %.0fs", eff.BurnDPS, eff.BurnDurationSec))
		}
		if len(parts) == 0 {
			parts = append(parts, "passive")
		}
		return strings.Join(parts, " · ")
	}
	// Active skill effects
	if eff.DamageMultiplier > 0 {
		parts = append(parts, fmt.Sprintf("%.1f× weapon dmg", eff.DamageMultiplier))
	}
	if eff.AOERadius > 0 {
		parts = append(parts, fmt.Sprintf("AoE r%.1f", eff.AOERadius))
	}
	if eff.IsBlinkStrike {
		parts = append(parts, "blink + guarantee crit")
	}
	if eff.IsTaunt {
		parts = append(parts, fmt.Sprintf("taunt: %d%% DR for %.0fs", eff.DamageReductionPct, eff.DurationSec))
	}
	if eff.IsRoot {
		parts = append(parts, fmt.Sprintf("root enemy for %.0fs", eff.DurationSec))
	}
	if eff.IsExecute {
		parts = append(parts, fmt.Sprintf("execute enemy below %d%% HP", eff.ExecuteThresholdPct))
	}
	if eff.SurgeDmgPerCooldown > 0 {
		parts = append(parts, fmt.Sprintf("%d dmg per artifact on CD", eff.SurgeDmgPerCooldown))
	}
	if eff.HPCostPct > 0 && eff.DamageFlat > 0 {
		parts = append(parts, fmt.Sprintf("spend %d%% HP → %d void dmg", eff.HPCostPct, eff.DamageFlat))
	}
	if eff.SpellDamageBase > 0 {
		s := fmt.Sprintf("%d base", eff.SpellDamageBase)
		if eff.SpellDamagePerINT > 0 {
			s += fmt.Sprintf(" +%.1f/INT", eff.SpellDamagePerINT)
		}
		if eff.SpellDamagePerSTR > 0 {
			s += fmt.Sprintf(" +%.1f/STR", eff.SpellDamagePerSTR)
		}
		if eff.IsAOEField {
			s += fmt.Sprintf(" field (×%.0fs)", eff.DurationSec)
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		parts = append(parts, "no registered effect")
	}
	return strings.Join(parts, " · ")
}

type logEntry struct {
	text  string
	age   int
	isHit bool
}

type visState struct {
	state         combat.CombatState
	enemyAtkTimer float64
	kills         int
	dmgDealt      int
	dmgTaken      int
	simTick       int
	dead          bool
	deathTick     int
	enemyFlash    int
	playerFlash   int
}

type spark struct {
	x, y   float32
	vx, vy float32
	life   int
	col    color.RGBA
}

// VisGame is the Ebiten game.
type VisGame struct {
	scenarios []combat.Scenario
	scenIdx   int
	engine    combat.CombatEngine
	rotIdx    int

	vs        visState
	log       []logEntry
	tickAccum float64
	speed     float64
	paused    bool
	sparks    []spark

	showScenList bool // scenario picker overlay open
	scenListHov  int  // hovered row index (-1 = none)
}

func newVisGame(scenarios []combat.Scenario) *VisGame {
	g := &VisGame{
		scenarios: scenarios,
		engine:    combat.NewDefaultCombatEngine(42),
		speed:     1.0,
	}
	g.reset()
	return g
}

func (g *VisGame) currentScenario() combat.Scenario {
	return g.scenarios[g.scenIdx]
}

func (g *VisGame) reset() {
	s := g.currentScenario()
	g.vs = visState{
		state:         combat.BuildStateExported(s),
		enemyAtkTimer: combat.EnemyAttackInterval,
	}
	g.log = nil
	g.rotIdx = 0
	g.tickAccum = 0
	g.sparks = nil
}

func (g *VisGame) addLog(text string, isHit bool) {
	g.log = append(g.log, logEntry{text: text, isHit: isHit})
	if len(g.log) > maxLogLines*4 {
		g.log = g.log[len(g.log)-maxLogLines*4:]
	}
}

func (g *VisGame) Update() error {
	// Toggle scenario picker with Tab.
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showScenList = !g.showScenList
		g.scenListHov = -1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.showScenList {
		g.showScenList = false
	}

	// When the scenario picker is open, handle mouse interaction only.
	if g.showScenList {
		mx, my := ebiten.CursorPosition()
		g.scenListHov = g.scenListRowAt(mx, my)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.scenListHov >= 0 {
			g.selectScenario(g.scenListHov)
		}
		// Arrow keys still navigate while list is open.
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			g.scenListHov = clampI(g.scenListHov+1, 0, len(g.scenarios)-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			g.scenListHov = clampI(g.scenListHov-1, 0, len(g.scenarios)-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && g.scenListHov >= 0 {
			g.selectScenario(g.scenListHov)
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.engine = combat.NewDefaultCombatEngine(42)
		g.reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		g.speed = math.Min(g.speed*2, 8.0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		g.speed = math.Max(g.speed/2, 0.125)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && len(g.scenarios) > 1 {
		g.scenIdx = (g.scenIdx + 1) % len(g.scenarios)
		g.engine = combat.NewDefaultCombatEngine(42)
		g.reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && len(g.scenarios) > 1 {
		g.scenIdx = (g.scenIdx - 1 + len(g.scenarios)) % len(g.scenarios)
		g.engine = combat.NewDefaultCombatEngine(42)
		g.reset()
	}

	// Age sparks.
	alive := g.sparks[:0]
	for _, sp := range g.sparks {
		sp.x += sp.vx
		sp.y += sp.vy
		sp.life--
		if sp.life > 0 {
			alive = append(alive, sp)
		}
	}
	g.sparks = alive

	for i := range g.log {
		g.log[i].age++
	}

	if g.paused || g.vs.dead {
		return nil
	}

	g.tickAccum += g.speed
	for g.tickAccum >= 1.0 {
		g.tickAccum--
		g.simStep()
	}
	return nil
}

func (g *VisGame) simStep() {
	vs := &g.vs
	if vs.dead {
		return
	}

	s := g.currentScenario()
	rot := s.SkillRotation
	rotLen := len(rot)

	var actions []combat.Action
	if rotLen > 0 {
		entry := rot[g.rotIdx%rotLen]
		g.rotIdx++
		if entry != "auto" {
			slot := combat.ParseSlotIdxExported(entry)
			if slot >= 0 {
				actions = append(actions, combat.Action{
					Type:    combat.ActionActivateSkill,
					SlotIdx: slot,
				})
			}
		}
	}

	newState, events := g.engine.Tick(vs.state, actions)
	vs.state = newState
	vs.simTick++

	for _, ev := range events {
		switch ev.Type {
		case combat.EventDamageDealt:
			vs.dmgDealt += ev.Value
			vs.enemyFlash = 6
			tag := ev.Tag
			if tag != "" {
				tag = " [" + tag + "]"
			}
			suffix := ""
			if ev.IsCrit {
				suffix = " CRIT"
			}
			g.addLog(fmt.Sprintf("→ %d dmg%s%s", ev.Value, suffix, tag), false)
			g.spawnSparks(enemyX, dotY, color.RGBA{255, 180, 50, 255}, 6)

		case combat.EventTargetDied:
			vs.kills++
			// Award EXP in a separate slice to avoid mutating the range variable.
			var lvlEvents []combat.Event
			combat.AwardSimEXPExported(&vs.state, s.Floor, s, &lvlEvents)
			for _, lev := range lvlEvents {
				if lev.Type == combat.EventLevelUp {
					g.addLog(fmt.Sprintf("▲ LEVEL %d  +%s", lev.Value, lev.Tag), false)
					g.spawnSparks(playerX, dotY, color.RGBA{100, 220, 255, 255}, 30)
				}
			}
			vs.state.TargetHP = combat.EnemyHPExported(s.Floor)
			vs.state.TargetMaxHP = vs.state.TargetHP
			vs.state.TargetIsDead = false
			vs.state.HasTarget = true
			vs.state.IsAutoAttacking = true
			vs.enemyAtkTimer = combat.EnemyAttackInterval
			g.addLog(fmt.Sprintf("★ KILL #%d (streak ×%d)", vs.kills, ev.Value), false)
			g.spawnSparks(enemyX, dotY, color.RGBA{255, 80, 80, 255}, 20)

		case combat.EventSkillFired:
			g.addLog(fmt.Sprintf("⚡ %s", ev.Tag), false)

		case combat.EventHPSpent:
			if vs.state.EnemySacrificeLeech && !vs.state.TargetIsDead {
				vs.state.TargetHP += ev.Value
				if vs.state.TargetHP > vs.state.TargetMaxHP {
					vs.state.TargetHP = vs.state.TargetMaxHP
				}
			}
		}
	}

	// Enemy attack.
	vs.enemyAtkTimer -= combat.TickDurationExported
	if vs.enemyAtkTimer <= 0 && vs.state.HasTarget && !vs.state.TargetIsDead {
		vs.enemyAtkTimer = combat.EnemyAttackInterval
		dmg := combat.EnemyDmgPerAttackExported(s.Floor)

		if vs.state.TargetRooted {
			dmg = 0
		}
		if dmg > 0 {
			if vs.state.DamageReductionPct > 0 {
				dmg = int(float64(dmg) * (1.0 - float64(vs.state.DamageReductionPct)/100.0))
				if dmg < 0 {
					dmg = 0
				}
			}
			rawDmg := dmg
			for _, id := range vs.state.EquippedArtifacts {
				if eff, ok := combat.ArtifactEffects[id]; ok && eff.DamageCapPct > 0 {
					cap := int(float64(vs.state.PlayerMaxHP) * float64(eff.DamageCapPct) / 100.0)
					if cap < 1 {
						cap = 1
					}
					if dmg > cap {
						dmg = cap
					}
					break
				}
			}
			if vs.state.EnemyDamageCapBypass && vs.state.PlayerHP*2 < vs.state.PlayerMaxHP {
				dmg = rawDmg
			}
			if vs.state.EnemyPackBonusPct > 0 {
				dmg += int(float64(dmg) * float64(vs.state.EnemyPackBonusPct) / 100.0)
			}

			vs.state.PlayerHP -= dmg
			vs.dmgTaken += dmg
			vs.playerFlash = 8
			g.addLog(fmt.Sprintf("! enemy hits for %d  [HP: %d/%d]", dmg, vs.state.PlayerHP, vs.state.PlayerMaxHP), true)
			g.spawnSparks(playerX, dotY, color.RGBA{255, 60, 60, 255}, 6)

			if vs.state.EnemyInstakillPct > 0 && vs.state.PlayerHP > 0 {
				threshold := vs.state.PlayerMaxHP * vs.state.EnemyInstakillPct / 100
				if vs.state.PlayerHP <= threshold {
					vs.state.PlayerHP = 0
				}
			}
			if vs.state.KillStreak > 0 {
				vs.state.KillStreak = 0
				vs.state.StreakTimer = 0
			}
		}
	}

	// Regenerator.
	if vs.state.EnemyHPRegenPerSec > 0 && !vs.state.TargetIsDead {
		regen := int(float64(vs.state.EnemyHPRegenPerSec) * combat.TickDurationExported)
		if regen > 0 {
			vs.state.TargetHP += regen
			if vs.state.TargetHP > vs.state.TargetMaxHP {
				vs.state.TargetHP = vs.state.TargetMaxHP
			}
		}
	}

	if vs.enemyFlash > 0 {
		vs.enemyFlash--
	}
	if vs.playerFlash > 0 {
		vs.playerFlash--
	}

	if vs.state.PlayerHP <= 0 {
		vs.dead = true
		vs.deathTick = vs.simTick
		g.addLog(fmt.Sprintf("✝ PLAYER DIED  (kills: %d, t=%.1fs)", vs.kills, float64(vs.simTick)/60.0), true)
	}
}

func (g *VisGame) spawnSparks(x, y float32, c color.RGBA, n int) {
	for i := 0; i < n; i++ {
		angle := float64(i) / float64(n) * 2 * math.Pi
		speed := float32(1.5 + float64(i%3)*0.7)
		g.sparks = append(g.sparks, spark{
			x: x, y: y,
			vx:  float32(math.Cos(angle)) * speed,
			vy:  float32(math.Sin(angle)) * speed,
			life: 14 + i%8,
			col: c,
		})
	}
}

// ── Drawing ──────────────────────────────────────────────────────────────────

func (g *VisGame) Draw(screen *ebiten.Image) {
	screen.Fill(colBG)
	g.drawArena(screen)
	g.drawPanel(screen)
	if g.showScenList {
		g.drawScenarioList(screen)
	}
}

func (g *VisGame) drawArena(screen *ebiten.Image) {
	vs := &g.vs
	s := g.currentScenario()

	ebitenutil.DrawRect(screen, 0, 0, arenaW, totalH, colArenaBG)

	// Title bar.
	ebitenutil.DrawRect(screen, 0, 0, arenaW, 26, colHeader)
	name := s.Name
	if len(g.scenarios) > 1 {
		name = fmt.Sprintf("[%d/%d] %s", g.scenIdx+1, len(g.scenarios), name)
	}
	ebitenutil.DebugPrintAt(screen, "  "+name, 0, 7)

	// Ground line.
	ebitenutil.DrawRect(screen, 30, float64(dotY)+36, arenaW-60, 1, colDivider)

	// Sparks.
	for _, sp := range g.sparks {
		alpha := uint8(clampI(sp.life*15, 0, 255))
		c := color.RGBA{sp.col.R, sp.col.G, sp.col.B, alpha}
		vector.DrawFilledRect(screen, sp.x-2, sp.y-2, 4, 4, c, false)
	}

	// ── ENEMY ────────────────────────────────────────────────────────────
	eCol := colRed
	if vs.enemyFlash > 0 {
		eCol = colYellow
	}
	if vs.state.TargetIsDead {
		eCol = color.RGBA{60, 20, 20, 255}
	}
	vector.DrawFilledCircle(screen, enemyX, dotY, 26, eCol, true)

	// Root ring.
	if vs.state.TargetRooted {
		vector.StrokeCircle(screen, enemyX, dotY, 30, 3, colGreen, true)
	}

	// Enemy HP bar.
	drawBar(screen, enemyX-54, dotY-52, 108, 10, vs.state.TargetHP, vs.state.TargetMaxHP, colRed)
	// Enemy attack charge bar (how close to next attack).
	atkCharge := 1.0 - vs.enemyAtkTimer/combat.EnemyAttackInterval
	drawBarF(screen, enemyX-54, dotY-38, 108, 6, atkCharge, colOrange)

	ebitenutil.DebugPrintAt(screen, "ENEMY", int(enemyX)-16, int(dotY)+32)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d / %d HP", vs.state.TargetHP, vs.state.TargetMaxHP), int(enemyX)-36, int(dotY)+46)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("atk: %d dmg / %.0fs", combat.EnemyDmgPerAttackExported(s.Floor), combat.EnemyAttackInterval), int(enemyX)-50, int(dotY)+60)

	// ── PLAYER ───────────────────────────────────────────────────────────
	pCol := colBlue
	if vs.state.InShadow {
		pCol = colPurple
	}
	if vs.playerFlash > 0 {
		pCol = colRed
	}
	if vs.dead {
		pCol = color.RGBA{80, 30, 30, 255}
	}
	vector.DrawFilledCircle(screen, playerX, dotY, 28, pCol, true)

	// Shadow aura.
	if vs.state.InShadow {
		vector.StrokeCircle(screen, playerX, dotY, 34, 3, colPurple, true)
	}
	// Taunt ring.
	if vs.state.TauntTimer > 0 {
		vector.StrokeCircle(screen, playerX, dotY, 36, 3, colBlue, true)
	}

	// Player HP bar.
	drawBar(screen, playerX-54, dotY-52, 108, 12, vs.state.PlayerHP, vs.state.PlayerMaxHP, colGreen)

	ebitenutil.DebugPrintAt(screen, "PLAYER", int(playerX)-20, int(dotY)+32)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d / %d HP", vs.state.PlayerHP, vs.state.PlayerMaxHP), int(playerX)-36, int(dotY)+46)

	// Auto-attack timer bar (progress toward next auto).
	autoTimer := vs.state.AutoAttackTimer
	maxInterval := vs.state.PlayerAttackInterval
	autoPct := 0.0
	if maxInterval > 0 {
		autoPct = 1.0 - autoTimer/maxInterval
		if autoPct < 0 {
			autoPct = 0
		}
	}
	drawBarF(screen, playerX-54, dotY-38, 108, 6, autoPct, colBlue)

	// ── Active effects summary (between dots) ─────────────────────────
	mx := float64(arenaW)/2 - 40
	my := float64(dotY) - 80.0
	effects := g.activeEffectLines(vs)
	for _, e := range effects {
		ebitenutil.DebugPrintAt(screen, e, int(mx), int(my))
		my += 14
	}

	// ── Status bar ───────────────────────────────────────────────────────
	ebitenutil.DrawRect(screen, 0, float64(totalH)-46, arenaW, 46, colHeader)
	ebitenutil.DrawRect(screen, 0, float64(totalH)-47, arenaW, 1, colDivider)

	simSec := float64(vs.simTick) / 60.0
	kpm := 0.0
	if simSec > 0 {
		kpm = float64(vs.kills) / simSec * 60.0
	}
	statusLine1 := fmt.Sprintf("  t=%.1fs   kills=%d   KPM=%.1f   speed=%.2g×", simSec, vs.kills, kpm, g.speed)
	if g.paused {
		statusLine1 += "  [PAUSED]"
	}
	if vs.dead {
		statusLine1 = fmt.Sprintf("  ✝ DIED at t=%.1fs after %d kills  (KPM was %.1f)", float64(vs.deathTick)/60.0, vs.kills, kpm)
	}
	ebitenutil.DebugPrintAt(screen, statusLine1, 0, totalH-44)
	ebitenutil.DebugPrintAt(screen, "  Space=pause   +/-=speed   R=restart   ←/→=cycle   Tab=pick scenario", 0, totalH-28)
}

func (g *VisGame) activeEffectLines(vs *visState) []string {
	var lines []string
	if vs.state.InShadow {
		lines = append(lines, fmt.Sprintf("◈ SHADOW  %.1fs", vs.state.ShadowTimer))
	}
	if vs.state.TauntTimer > 0 {
		lines = append(lines, fmt.Sprintf("◈ TAUNT   %.1fs  %d%%DR", vs.state.TauntTimer, vs.state.DamageReductionPct))
	}
	if vs.state.BurnActive {
		lines = append(lines, fmt.Sprintf("◈ BURN    %d dps  %.1fs", vs.state.BurnDPS, vs.state.BurnTimer))
	}
	if vs.state.TargetRooted {
		lines = append(lines, fmt.Sprintf("◈ ROOTED  %.1fs", vs.state.RootTimer))
	}
	if vs.state.KillStreak > 0 {
		lines = append(lines, fmt.Sprintf("◈ STREAK ×%d  +%.0f%%spd", vs.state.KillStreak, float64(vs.state.KillStreak)*5))
	}
	return lines
}

// ── Panel ─────────────────────────────────────────────────────────────────

func (g *VisGame) drawPanel(screen *ebiten.Image) {
	vs := &g.vs
	s := g.currentScenario()

	ebitenutil.DrawRect(screen, arenaW, 0, panelW, totalH, colPanelBG)
	ebitenutil.DrawRect(screen, arenaW, 0, 1, totalH, colDivider)

	// We'll use a simple cursor-based layout.
	c := &panelCursor{screen: screen, x: arenaW + 8, y: 6}

	// ── Section: Stats ─────────────────────────────────────────────────
	c.header("PLAYER STATS")
	// Level + EXP bar
	lvlStr := fmt.Sprintf("Lv %d  %d / %d EXP", vs.state.PlayerLevel, vs.state.PlayerEXP, vs.state.PlayerEXPToNext)
	c.stat("Level", lvlStr, colYellow)
	expPct := 0.0
	if vs.state.PlayerEXPToNext > 0 {
		expPct = float64(vs.state.PlayerEXP) / float64(vs.state.PlayerEXPToNext)
	}
	drawBarF(c.screen, float32(c.x+6), float32(c.y), float32(panelW-30), 5, expPct, colYellow)
	c.y += 8
	c.stat("HP", fmt.Sprintf("%d / %d", vs.state.PlayerHP, vs.state.PlayerMaxHP), hpColor(vs))
	c.stat("Damage", fmt.Sprintf("%d per hit", vs.state.PlayerDamage), colText)
	c.stat("Atk Speed", fmt.Sprintf("%.2fs interval", vs.state.PlayerAttackInterval), colText)
	c.stat("STR / INT", fmt.Sprintf("%d / %d", vs.state.PlayerStrength, vs.state.PlayerIntelligence), colText)
	c.stat("CDR", fmt.Sprintf("%d%%", vs.state.CooldownReductionPct), colText)
	c.stat("Atk Spd Bonus", fmt.Sprintf("+%d%%", vs.state.AttackSpeedPct), colText)
	c.gap()

	// ── Section: Enemy ─────────────────────────────────────────────────
	c.header("ENEMY")
	c.stat("Floor", fmt.Sprintf("%d", s.Floor), colText)
	c.stat("HP", fmt.Sprintf("%d / %d", vs.state.TargetHP, vs.state.TargetMaxHP), colRed)
	rawAtk := combat.EnemyDmgPerAttackExported(s.Floor)
	c.stat("Base Atk", fmt.Sprintf("%d dmg every %.0fs", rawAtk, combat.EnemyAttackInterval), colText)
	if s.EnemyBuildID != "" {
		c.stat("Counter", s.EnemyBuildID, colPurple)
	}
	c.stat("Next attack", fmt.Sprintf("%.2fs", vs.enemyAtkTimer), colOrange)
	c.gap()

	// ── Section: Combat Stats ──────────────────────────────────────────
	c.header("COMBAT STATS")
	simSec := float64(vs.simTick) / 60.0
	kpm := 0.0
	if simSec > 0 {
		kpm = float64(vs.kills) / simSec * 60.0
	}
	c.stat("Time", fmt.Sprintf("%.1fs", simSec), colText)
	c.stat("Kills", fmt.Sprintf("%d  (%.1f KPM)", vs.kills, kpm), colYellow)
	c.stat("Dmg dealt", fmt.Sprintf("%d", vs.dmgDealt), colGreen)
	c.stat("Dmg taken", fmt.Sprintf("%d", vs.dmgTaken), colRed)
	if vs.dead {
		c.statColored("Status", "DEAD", color.RGBA{255, 60, 60, 255})
	} else if g.paused {
		c.statColored("Status", "PAUSED", colOrange)
	} else {
		c.statColored("Status", "ALIVE", colGreen)
	}
	c.gap()

	// ── Section: Equipped Items ────────────────────────────────────────
	c.header("EQUIPPED ITEMS")
	g.drawItemTable(c, vs, s)
	c.gap()

	// ── Section: Event Log ────────────────────────────────────────────
	c.header("EVENT LOG")
	start := len(g.log) - maxLogLines
	if start < 0 {
		start = 0
	}
	for _, entry := range g.log[start:] {
		col := colText
		if entry.isHit {
			col = colRed
		} else if strings.HasPrefix(entry.text, "★") {
			col = colYellow
		} else if strings.HasPrefix(entry.text, "⚡") {
			col = colBlue
		}
		c.textCol(entry.text, col)
	}
}

func (g *VisGame) drawItemTable(c *panelCursor, vs *visState, _ combat.Scenario) {
	for i, id := range vs.state.EquippedArtifacts {
		if id == "" {
			continue
		}
		eff, known := combat.ArtifactEffects[id]
		isElite := i == 6

		// Slot label.
		slotLabel := fmt.Sprintf("s%d", i+1)
		if isElite {
			slotLabel = "EL"
		}

		// CD status.
		cdStatus := "PASSIVE"
		cdCol := colDim
		if known && eff.Cooldown > 0 {
			remaining := vs.state.ArtifactCooldowns[i]
			if remaining <= 0 {
				cdStatus = "READY"
				cdCol = colCDReady
			} else {
				cdStatus = fmt.Sprintf("%.1fs / %.0fs", remaining, eff.Cooldown)
				cdCol = colCDActive
			}
		}

		// Domain color.
		dc := color.RGBA{160, 160, 160, 255}
		if known {
			dc = domainColor(eff.Domain)
		}

		// Row: slot  name  [domain]  CD status
		nameStr := id
		// Truncate long names to fit.
		maxNameLen := 24
		if len(nameStr) > maxNameLen {
			nameStr = nameStr[:maxNameLen-1] + "…"
		}
		c.itemRow(slotLabel, nameStr, dc, cdStatus, cdCol)

		// Indent: effect description.
		if known {
			desc := effectDesc(eff)
			// Wrap at ~55 chars.
			for _, line := range wrapText(desc, 55) {
				c.textCol("       "+line, colDim)
			}
		}

		// CD progress bar for active skills.
		if known && eff.Cooldown > 0 {
			remaining := vs.state.ArtifactCooldowns[i]
			pct := 1.0 - remaining/eff.Cooldown
			if pct < 0 {
				pct = 0
			}
			drawBarF(c.screen, float32(c.x+6), float32(c.y), float32(panelW-30), 4, pct, cdCol)
			c.y += 7
		}

		c.y += 3 // small gap between items
	}
}

// ── Helper types and functions ────────────────────────────────────────────

type panelCursor struct {
	screen *ebiten.Image
	x, y   int
}

func (c *panelCursor) gap() { c.y += 6 }

func (c *panelCursor) header(title string) {
	ebitenutil.DrawRect(c.screen, float64(c.x-8), float64(c.y), float64(panelW), 1, colDivider)
	c.y += 3
	ebitenutil.DebugPrintAt(c.screen, title, c.x, c.y)
	c.y += lineH + 2
	ebitenutil.DrawRect(c.screen, float64(c.x-8), float64(c.y-3), float64(panelW), 1, colDivider)
}

func (c *panelCursor) text(s string) {
	ebitenutil.DebugPrintAt(c.screen, s, c.x, c.y)
	c.y += lineH
}

func (c *panelCursor) textCol(s string, _ color.RGBA) {
	// DebugPrintAt doesn't support per-call color; use as-is.
	ebitenutil.DebugPrintAt(c.screen, s, c.x, c.y)
	c.y += lineH
}

func (c *panelCursor) stat(label, value string, _ color.RGBA) {
	row := fmt.Sprintf("  %-18s %s", label, value)
	ebitenutil.DebugPrintAt(c.screen, row, c.x, c.y)
	c.y += lineH
}

func (c *panelCursor) statColored(label, value string, _ color.RGBA) {
	c.stat(label, value, colText)
}

func (c *panelCursor) itemRow(slot, name string, _ color.RGBA, cdStatus string, _ color.RGBA) {
	row := fmt.Sprintf("  [%s] %-26s %s", slot, name, cdStatus)
	ebitenutil.DebugPrintAt(c.screen, row, c.x, c.y)
	c.y += lineH
}

func drawBar(screen *ebiten.Image, x, y, w, h float32, cur, max int, col color.RGBA) {
	if max <= 0 {
		return
	}
	pct := float64(cur) / float64(max)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), colCDBar)
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w)*pct, float64(h), col)
}

func drawBarF(screen *ebiten.Image, x, y, w, h float32, pct float64, col color.RGBA) {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), colCDBar)
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w)*pct, float64(h), col)
}

func hpColor(vs *visState) color.RGBA {
	if vs.dead {
		return colRed
	}
	pct := float64(vs.state.PlayerHP) / float64(vs.state.PlayerMaxHP)
	if pct > 0.5 {
		return colGreen
	}
	if pct > 0.25 {
		return colOrange
	}
	return colRed
}

func clampI(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func wrapText(s string, maxChars int) []string {
	if len(s) <= maxChars {
		return []string{s}
	}
	var lines []string
	for len(s) > maxChars {
		cut := maxChars
		// Try to break at a space.
		for i := maxChars - 1; i > maxChars/2; i-- {
			if s[i] == ' ' || s[i] == '·' {
				cut = i + 1
				break
			}
		}
		lines = append(lines, s[:cut])
		s = s[cut:]
	}
	if len(s) > 0 {
		lines = append(lines, s)
	}
	return lines
}

func (g *VisGame) Layout(outsideW, outsideH int) (int, int) {
	return totalW, totalH
}

// ── Scenario picker ───────────────────────────────────────────────────────

const (
	scenListX    = 40
	scenListY    = 44
	scenRowH     = 18
	scenRowPad   = 6  // horizontal text padding inside a row
	scenListW    = totalW - 80
)

func (g *VisGame) selectScenario(idx int) {
	g.scenIdx = idx
	g.engine = combat.NewDefaultCombatEngine(42)
	g.reset()
	g.showScenList = false
	g.scenListHov = -1
}

// scenListRowAt returns the scenario index under pixel (mx, my), or -1.
func (g *VisGame) scenListRowAt(mx, my int) int {
	if mx < scenListX || mx > scenListX+scenListW {
		return -1
	}
	row := (my - scenListY) / scenRowH
	if row < 0 || row >= len(g.scenarios) {
		return -1
	}
	return row
}

func (g *VisGame) drawScenarioList(screen *ebiten.Image) {
	// Dim the background.
	overlay := ebiten.NewImage(totalW, totalH)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	// Panel background.
	listH := scenRowH*len(g.scenarios) + 60
	ebitenutil.DrawRect(screen, float64(scenListX-10), 20, float64(scenListW+20), float64(listH), color.RGBA{18, 18, 30, 255})
	ebitenutil.DrawRect(screen, float64(scenListX-10), 20, float64(scenListW+20), 1, colDivider)
	ebitenutil.DrawRect(screen, float64(scenListX-10), float64(20+listH), float64(scenListW+20), 1, colDivider)

	// Header.
	ebitenutil.DrawRect(screen, float64(scenListX-10), 20, float64(scenListW+20), 22, colHeader)
	ebitenutil.DebugPrintAt(screen, "  SELECT SCENARIO   (Tab / Esc to close  ·  ↑↓ + Enter to keyboard-select)", scenListX-4, 27)

	// Column header.
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %-3s  %-7s  %-40s  %s", "#", "CLASS", "NAME", "FLOOR / PRIORITY"), scenListX, scenListY-16)
	ebitenutil.DrawRect(screen, float64(scenListX-2), float64(scenListY-2), float64(scenListW+4), 1, colDivider)

	// Rows.
	for i, s := range g.scenarios {
		y := scenListY + i*scenRowH

		// Row background.
		rowBG := color.RGBA{0, 0, 0, 0}
		if i == g.scenIdx {
			rowBG = color.RGBA{30, 60, 30, 200}
		}
		if i == g.scenListHov {
			rowBG = color.RGBA{50, 50, 90, 220}
		}
		if i == g.scenIdx && i == g.scenListHov {
			rowBG = color.RGBA{40, 80, 40, 230}
		}
		if rowBG.A > 0 {
			ebitenutil.DrawRect(screen, float64(scenListX-2), float64(y), float64(scenListW+4), float64(scenRowH), rowBG)
		}

		// Truncate name to fit.
		name := s.Name
		if len(name) > 44 {
			name = name[:43] + "…"
		}

		priority := "default"
		if len(s.StatPriority) > 0 {
			priority = strings.Join(s.StatPriority, "+")
		}

		classTag := s.Class
		if len(classTag) > 6 {
			classTag = classTag[:6]
		}

		marker := "  "
		if i == g.scenIdx {
			marker = "▶ "
		}

		row := fmt.Sprintf("%s%-3d  %-7s  %-44s  fl.%-2d  [%s]",
			marker, i+1, classTag, name, s.Floor, priority)
		ebitenutil.DebugPrintAt(screen, row, scenListX, y+2)
	}
}
