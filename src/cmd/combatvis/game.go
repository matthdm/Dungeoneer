package main

import (
	"dungeoneer/combat"
	"fmt"
	"image/color"
	"math"
	"reflect"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type logEntry struct {
	text  string
	isHit bool
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

	// active is the scenario being watched: the loaded one, or an edited copy
	// applied from the setup screen. seed picks which fight of it is shown.
	active combat.Scenario
	seed   uint64
	sim    *combat.Sim
	wave   []combat.WaveEnemy // scratch buffer for drawing the rest of the wave

	enemyFlash  int
	playerFlash int
	log         []logEntry
	tickAccum   float64
	speed       float64
	paused      bool
	sparks      []spark

	showScenList bool // scenario picker overlay open
	scenListHov  int  // hovered row index (-1 = none)

	setup setupScreen
	batch batchRunner
}

func newVisGame(scenarios []combat.Scenario) *VisGame {
	g := &VisGame{
		scenarios: scenarios,
		speed:     1.0,
	}
	g.setup.init()
	g.loadScenario(0)
	return g
}

// loadScenario switches to one of the loaded scenarios, discarding any edits.
func (g *VisGame) loadScenario(idx int) {
	g.scenIdx = idx
	g.active = cloneScenario(g.scenarios[idx])
	g.setup.edit = cloneScenario(g.active)
	g.setup.err = ""
	g.reset()
}

// reset restarts the fight for the active scenario and seed.
func (g *VisGame) reset() {
	g.log = g.log[:0]
	g.sim = combat.NewSim(g.active, g.seed, 0, logSink{g})
	g.tickAccum = 0
	g.sparks = g.sparks[:0]
	g.enemyFlash, g.playerFlash = 0, 0
}

// modified reports whether the watched scenario differs from the loaded file.
func (g *VisGame) modified() bool {
	return !reflect.DeepEqual(g.active, g.scenarios[g.scenIdx])
}

// logSink feeds the simulator's trace lines into the on-screen event log.
type logSink struct{ g *VisGame }

func (l logSink) Write(p []byte) (int, error) {
	line := strings.TrimRight(string(p), "\n")
	isHit := strings.Contains(line, "enemy hits") || strings.Contains(line, "PLAYER DIED") || strings.Contains(line, "enemy executes")
	l.g.log = append(l.g.log, logEntry{text: line, isHit: isHit})
	if len(l.g.log) > maxLogLines*4 {
		l.g.log = l.g.log[len(l.g.log)-maxLogLines*4:]
	}
	return len(p), nil
}

func (g *VisGame) Update() error {
	g.batch.poll()

	// Overlays take the input while they are open, topmost first.
	switch {
	case g.batch.visible:
		g.batch.update()
		return nil
	case g.setup.open:
		g.updateSetup()
		return nil
	case g.showScenList:
		g.updateScenarioList()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showScenList = true
		g.scenListHov = g.scenIdx
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.setup.open = true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) && g.batch.hasResults() {
		g.batch.visible = true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) { // next seed: a different fight of the same scenario
		g.seed++
		g.reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		g.speed = math.Min(g.speed*2, 64.0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		g.speed = math.Max(g.speed/2, 0.125)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && len(g.scenarios) > 1 {
		g.loadScenario((g.scenIdx + 1) % len(g.scenarios))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && len(g.scenarios) > 1 {
		g.loadScenario((g.scenIdx - 1 + len(g.scenarios)) % len(g.scenarios))
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
	if g.enemyFlash > 0 {
		g.enemyFlash--
	}
	if g.playerFlash > 0 {
		g.playerFlash--
	}

	if g.paused || g.sim.Dead() {
		return nil
	}

	g.tickAccum += g.speed
	for g.tickAccum >= 1.0 {
		g.tickAccum--
		g.simStep()
	}
	return nil
}

// simStep advances the fight one tick and turns what changed into flashes and
// sparks. The numbers and the log come from the simulator itself.
func (g *VisGame) simStep() {
	st := g.sim.State()
	playerHP, targetHP, kills := st.PlayerHP, st.TargetHP, g.sim.Kills()

	g.sim.Step()

	switch {
	case g.sim.Kills() > kills:
		g.enemyFlash = 6
		g.spawnSparks(enemyX, dotY, color.RGBA{255, 80, 80, 255}, 20)
	case st.TargetHP < targetHP:
		g.enemyFlash = 6
		g.spawnSparks(enemyX, dotY, color.RGBA{255, 180, 50, 255}, 6)
	}
	if st.PlayerHP < playerHP {
		g.playerFlash = 8
		g.spawnSparks(playerX, dotY, color.RGBA{255, 60, 60, 255}, 6)
	}
}

func (g *VisGame) spawnSparks(x, y float32, c color.RGBA, n int) {
	if len(g.sparks) > 400 { // high speeds would otherwise flood the arena
		return
	}
	for i := 0; i < n; i++ {
		angle := float64(i) / float64(n) * 2 * math.Pi
		speed := float32(1.5 + float64(i%3)*0.7)
		g.sparks = append(g.sparks, spark{
			x: x, y: y,
			vx:   float32(math.Cos(angle)) * speed,
			vy:   float32(math.Sin(angle)) * speed,
			life: 14 + i%8,
			col:  c,
		})
	}
}

// ── Drawing ──────────────────────────────────────────────────────────────────

func (g *VisGame) Draw(screen *ebiten.Image) {
	screen.Fill(colBG)
	g.drawArena(screen)
	g.drawPanel(screen)
	switch {
	case g.batch.visible:
		g.batch.draw(screen)
	case g.setup.open:
		g.drawSetup(screen)
	case g.showScenList:
		g.drawScenarioList(screen)
	}
}

// dimScreen darkens everything drawn so far, behind an overlay.
func dimScreen(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, totalW, totalH, color.RGBA{0, 0, 0, 190}, false)
}

func (g *VisGame) kpm() float64 {
	if t := g.sim.Time(); t > 0 {
		return float64(g.sim.Kills()) / t * 60.0
	}
	return 0
}

func (g *VisGame) drawArena(screen *ebiten.Image) {
	st := g.sim.State()
	s := g.active
	dead := g.sim.Dead()
	timers := g.sim.AttackTimers()

	ebitenutil.DrawRect(screen, 0, 0, arenaW, totalH, colArenaBG)

	// Title bar.
	ebitenutil.DrawRect(screen, 0, 0, arenaW, 42, colHeader)
	name := s.Name
	if len(g.scenarios) > 1 {
		name = fmt.Sprintf("[%d/%d] %s", g.scenIdx+1, len(g.scenarios), name)
	}
	if g.modified() {
		name += "  (modified)"
	}
	ebitenutil.DebugPrintAt(screen, "  "+name, 0, 5)
	ebitenutil.DebugPrintAt(screen, "  "+scenarioSummary(s)+fmt.Sprintf("   seed %d", g.seed), 0, 22)

	// Ground line.
	ebitenutil.DrawRect(screen, 30, float64(dotY)+36, arenaW-60, 1, colDivider)

	// Sparks.
	for _, sp := range g.sparks {
		alpha := uint8(clampI(sp.life*15, 0, 255))
		c := color.RGBA{sp.col.R, sp.col.G, sp.col.B, alpha}
		vector.DrawFilledRect(screen, sp.x-2, sp.y-2, 4, 4, c, false)
	}

	// ── REST OF THE WAVE ─────────────────────────────────────────────────
	g.wave = g.sim.Wave(g.wave[:0])
	for i, m := range g.wave {
		x := enemyX + 62 + float32(i%4)*34
		y := dotY - 100 + float32(i/4)*40
		vector.DrawFilledCircle(screen, x, y, 11, color.RGBA{150, 60, 60, 255}, true)
		drawBar(screen, x-13, y+14, 26, 4, m.HP, m.MaxHP, colRed)
		if i+1 < len(timers) {
			drawBarF(screen, x-13, y+19, 26, 3, 1.0-timers[i+1]/combat.EnemyAttackInterval, colOrange)
		}
	}

	// ── ENEMY ────────────────────────────────────────────────────────────
	eCol := colRed
	if g.enemyFlash > 0 {
		eCol = colYellow
	}
	vector.DrawFilledCircle(screen, enemyX, dotY, 26, eCol, true)

	// Root ring.
	if st.TargetRooted {
		vector.StrokeCircle(screen, enemyX, dotY, 30, 3, colGreen, true)
	}

	// Enemy HP bar.
	drawBar(screen, enemyX-54, dotY-52, 108, 10, st.TargetHP, st.TargetMaxHP, colRed)
	// Enemy attack charge bar (how close to next attack).
	drawBarF(screen, enemyX-54, dotY-38, 108, 6, 1.0-timers[0]/combat.EnemyAttackInterval, colOrange)

	label := "ENEMY"
	if len(g.wave) > 0 {
		label = fmt.Sprintf("ENEMY (wave of %d)", len(g.wave)+1)
	}
	ebitenutil.DebugPrintAt(screen, label, int(enemyX)-16, int(dotY)+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d / %d HP", st.TargetHP, st.TargetMaxHP), int(enemyX)-36, int(dotY)+54)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("floor %d: ~%d dmg / %.0fs each", g.sim.Floor(), g.sim.EnemyBaseDamage(), combat.EnemyAttackInterval), int(enemyX)-80, int(dotY)+68)

	// ── PLAYER ───────────────────────────────────────────────────────────
	pCol := colBlue
	if st.InShadow {
		pCol = colPurple
	}
	if g.playerFlash > 0 {
		pCol = colRed
	}
	if dead {
		pCol = color.RGBA{80, 30, 30, 255}
	}
	vector.DrawFilledCircle(screen, playerX, dotY, 28, pCol, true)

	// Shadow aura.
	if st.InShadow {
		vector.StrokeCircle(screen, playerX, dotY, 34, 3, colPurple, true)
	}
	// Taunt ring.
	if st.TauntTimer > 0 {
		vector.StrokeCircle(screen, playerX, dotY, 36, 3, colBlue, true)
	}

	// Player HP and mana bars.
	drawBar(screen, playerX-54, dotY-58, 108, 12, st.PlayerHP, st.PlayerMaxHP, colGreen)
	drawBar(screen, playerX-54, dotY-45, 108, 5, st.PlayerMana, st.PlayerMaxMana, colBlue)

	ebitenutil.DebugPrintAt(screen, "PLAYER", int(playerX)-20, int(dotY)+40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d / %d HP", max(st.PlayerHP, 0), st.PlayerMaxHP), int(playerX)-36, int(dotY)+54)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d / %d mana", st.PlayerMana, st.PlayerMaxMana), int(playerX)-40, int(dotY)+68)

	// Auto-attack timer bar (progress toward next auto).
	autoPct := 0.0
	if st.PlayerAttackInterval > 0 {
		autoPct = 1.0 - st.AutoAttackTimer/st.PlayerAttackInterval
	}
	drawBarF(screen, playerX-54, dotY-38, 108, 6, autoPct, colYellow)

	// ── Active effects summary (between dots) ─────────────────────────
	mx := float64(arenaW)/2 - 60
	my := float64(dotY) - 150.0
	for _, e := range activeEffectLines(st) {
		ebitenutil.DebugPrintAt(screen, e, int(mx), int(my))
		my += 14
	}

	// ── Status bar ───────────────────────────────────────────────────────
	ebitenutil.DrawRect(screen, 0, float64(totalH)-62, arenaW, 62, colHeader)
	ebitenutil.DrawRect(screen, 0, float64(totalH)-63, arenaW, 1, colDivider)

	status := fmt.Sprintf("  t=%.1fs   kills=%d   KPM=%.1f   speed=%.3gx", g.sim.Time(), g.sim.Kills(), g.kpm(), g.speed)
	if g.paused {
		status += "  [PAUSED]"
	}
	if dead {
		status = fmt.Sprintf("  DIED at t=%.1fs after %d kills  (KPM was %.1f)", g.sim.Time(), g.sim.Kills(), g.kpm())
	}
	if g.batch.running {
		status += "   [benchmark running...]"
	}
	ebitenutil.DebugPrintAt(screen, status, 0, totalH-60)
	ebitenutil.DebugPrintAt(screen, "  Space=pause  +/-=speed  R=restart  N=next seed  Left/Right=cycle  Tab=pick scenario", 0, totalH-44)
	hint := "  E=edit build and wave settings, run benchmarks"
	if g.batch.hasResults() {
		hint += "   V=view last benchmark"
	}
	ebitenutil.DebugPrintAt(screen, hint, 0, totalH-28)
}

// scenarioSummary is a one-line description of what a scenario fights.
func scenarioSummary(s combat.Scenario) string {
	out := fmt.Sprintf("%s  floor %d", s.Class, s.Floor)
	if s.RampSec > 0 {
		out += fmt.Sprintf(" (+1 every %.0fs)", s.RampSec)
	}
	if s.Pack > 1 {
		out += fmt.Sprintf("  wave of %d", s.Pack)
	}
	if s.EnemyBuildID != "" {
		out += "  vs " + s.EnemyBuildID
	}
	if s.Leveling {
		out += "  leveling"
	}
	return out
}

func activeEffectLines(st *combat.CombatState) []string {
	var lines []string
	if st.InShadow {
		lines = append(lines, fmt.Sprintf("* SHADOW  %.1fs", st.ShadowTimer))
	}
	if st.TauntTimer > 0 {
		lines = append(lines, fmt.Sprintf("* TAUNT   %.1fs  %d%%DR", st.TauntTimer, st.DamageReductionPct))
	}
	if st.BurnActive {
		lines = append(lines, fmt.Sprintf("* BURN    %d dps  %.1fs", st.BurnDPS, st.BurnTimer))
	}
	if st.TargetRooted {
		lines = append(lines, fmt.Sprintf("* ROOTED  %.1fs", st.RootTimer))
	}
	if st.KillStreak > 0 {
		lines = append(lines, fmt.Sprintf("* STREAK x%d", st.KillStreak))
	}
	return lines
}

// ── Panel ─────────────────────────────────────────────────────────────────

func (g *VisGame) drawPanel(screen *ebiten.Image) {
	st := g.sim.State()
	s := g.active

	ebitenutil.DrawRect(screen, arenaW, 0, panelW, totalH, colPanelBG)
	ebitenutil.DrawRect(screen, arenaW, 0, 1, totalH, colDivider)

	// We'll use a simple cursor-based layout.
	c := &panelCursor{screen: screen, x: arenaW + 8, y: 6}

	// ── Section: Stats ─────────────────────────────────────────────────
	c.header("PLAYER STATS")
	if s.Leveling {
		c.stat("Level", fmt.Sprintf("Lv %d  %d / %d EXP", st.PlayerLevel, st.PlayerEXP, st.PlayerEXPToNext), colYellow)
	}
	c.stat("HP", fmt.Sprintf("%d / %d", max(st.PlayerHP, 0), st.PlayerMaxHP), colText)
	c.stat("Mana", fmt.Sprintf("%d / %d", st.PlayerMana, st.PlayerMaxMana), colText)
	c.stat("Damage", fmt.Sprintf("%d per hit, every %.2fs", st.PlayerDamage, st.PlayerAttackInterval), colText)
	c.stat("STR / INT", fmt.Sprintf("%d / %d", st.PlayerStrength, st.PlayerIntelligence), colText)
	c.stat("CDR / Atk speed", fmt.Sprintf("%d%% / +%d%%", st.CooldownReductionPct, st.AttackSpeedPct), colText)
	c.gap()

	// ── Section: Combat Stats ──────────────────────────────────────────
	c.header("THIS FIGHT")
	t := g.sim.Time()
	c.stat("Time", fmt.Sprintf("%.1fs", t), colText)
	c.stat("Kills", fmt.Sprintf("%d  (%.1f per minute)", g.sim.Kills(), g.kpm()), colYellow)
	if t > 0 {
		c.stat("DPS", fmt.Sprintf("%.1f", float64(g.sim.DamageDealt())/t), colGreen)
		c.stat("Damage in / heal", fmt.Sprintf("%.1f/s / %.1f/s", float64(g.sim.DamageTaken())/t, float64(g.sim.Healed())/t), colRed)
	}
	switch {
	case g.sim.Dead():
		c.stat("Status", "DEAD", colRed)
	case g.paused:
		c.stat("Status", "PAUSED", colOrange)
	default:
		c.stat("Status", "ALIVE", colGreen)
	}
	c.gap()

	// ── Section: Equipped Items ────────────────────────────────────────
	c.header("EQUIPPED ITEMS")
	g.drawItemTable(c, st)
	c.gap()

	// ── Section: Event Log ────────────────────────────────────────────
	c.header("EVENT LOG")
	lines := (totalH - c.y - 4) / lineH
	start := max(len(g.log)-lines, 0)
	for _, entry := range g.log[start:] {
		c.text(entry.text)
	}
}

func (g *VisGame) drawItemTable(c *panelCursor, st *combat.CombatState) {
	for i, id := range st.EquippedArtifacts {
		if id == "" {
			continue
		}
		eff, known := combat.ArtifactEffects[id]

		// Slot label.
		slotLabel := fmt.Sprintf("s%d", i+1)
		if i == 6 {
			slotLabel = "EL"
		}

		// CD status.
		cdStatus := "PASSIVE"
		cdCol := colDim
		if known && eff.Cooldown > 0 {
			remaining := st.ArtifactCooldowns[i]
			if remaining <= 0 {
				cdStatus = "READY"
				cdCol = colCDReady
			} else {
				cdStatus = fmt.Sprintf("%.1fs / %.0fs", remaining, eff.Cooldown)
				cdCol = colCDActive
			}
		}

		c.itemRow(slotLabel, truncate(id, 24), domainColor(eff.Domain), cdStatus, cdCol)

		// Indent: effect description.
		if known {
			for _, line := range wrapText(effectDesc(eff), 55) {
				c.textCol("       "+line, colDim)
			}
		}

		// CD progress bar for active skills.
		if known && eff.Cooldown > 0 {
			drawBarF(c.screen, float32(c.x+6), float32(c.y), float32(panelW-30), 4, 1.0-st.ArtifactCooldowns[i]/eff.Cooldown, cdCol)
			c.y += 7
		}

		c.y += 3 // small gap between items
	}
}

// truncate shortens s to at most n characters.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "~"
}

func (g *VisGame) Layout(outsideW, outsideH int) (int, int) {
	return totalW, totalH
}

// ── Scenario picker ───────────────────────────────────────────────────────

const (
	scenListX = 40
	scenListY = 60
	scenRowH  = 18
	scenListW = totalW - 80
)

func (g *VisGame) updateScenarioList() {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.showScenList = false
		return
	}
	mx, my := ebiten.CursorPosition()
	if row := g.scenListRowAt(mx, my); row >= 0 {
		g.scenListHov = row
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.selectScenario(row)
			return
		}
	}
	if keyRepeat(ebiten.KeyArrowDown) {
		g.scenListHov = clampI(g.scenListHov+1, 0, len(g.scenarios)-1)
	}
	if keyRepeat(ebiten.KeyArrowUp) {
		g.scenListHov = clampI(g.scenListHov-1, 0, len(g.scenarios)-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && g.scenListHov >= 0 {
		g.selectScenario(g.scenListHov)
	}
}

func (g *VisGame) selectScenario(idx int) {
	g.loadScenario(idx)
	g.showScenList = false
}

// scenListRowAt returns the scenario index under pixel (mx, my), or -1.
func (g *VisGame) scenListRowAt(mx, my int) int {
	if mx < scenListX || mx > scenListX+scenListW || my < scenListY {
		return -1
	}
	row := (my - scenListY) / scenRowH
	if row >= len(g.scenarios) {
		return -1
	}
	return row
}

func (g *VisGame) drawScenarioList(screen *ebiten.Image) {
	dimScreen(screen)

	// Panel background.
	listH := scenRowH*len(g.scenarios) + 50
	ebitenutil.DrawRect(screen, float64(scenListX-10), 20, float64(scenListW+20), float64(listH), color.RGBA{18, 18, 30, 255})

	// Header.
	ebitenutil.DrawRect(screen, float64(scenListX-10), 20, float64(scenListW+20), 22, colHeader)
	ebitenutil.DebugPrintAt(screen, "  SELECT SCENARIO   (click, or Up/Down + Enter; Tab or Esc to close)", scenListX-4, 24)

	// Column header.
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  %-3s  %-7s  %-60s  %s", "#", "CLASS", "NAME", "FIGHTS"), scenListX, scenListY-16)
	ebitenutil.DrawRect(screen, float64(scenListX-2), float64(scenListY-1), float64(scenListW+4), 1, colDivider)

	// Rows.
	for i, s := range g.scenarios {
		y := scenListY + i*scenRowH

		// Row background.
		rowBG := color.RGBA{}
		if i == g.scenIdx {
			rowBG = color.RGBA{30, 60, 30, 200}
		}
		if i == g.scenListHov {
			rowBG = color.RGBA{50, 50, 90, 220}
		}
		if rowBG.A > 0 {
			ebitenutil.DrawRect(screen, float64(scenListX-2), float64(y), float64(scenListW+4), float64(scenRowH), rowBG)
		}

		marker := "  "
		if i == g.scenIdx {
			marker = "> "
		}
		row := fmt.Sprintf("%s%-3d  %-7s  %-60s  %s", marker, i+1, s.Class, truncate(s.Name, 60), scenarioSummary(s))
		ebitenutil.DebugPrintAt(screen, row, scenListX, y+1)
	}
}

// keyRepeat reports a key press, repeating while the key is held.
func keyRepeat(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	return d == 1 || (d > 24 && d%3 == 0)
}
