package main

import (
	"dungeoneer/combat"
	"fmt"
	"image/color"
	"slices"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// The setup screen edits a copy of the watched scenario: the build (class,
// stats, artifacts) and what it fights (floor, wave size, ramp, enemy
// archetype). From it the edited scenario can be watched, benchmarked on its
// own, or its wave settings applied to a benchmark of every loaded scenario.

const (
	setupX      = 60
	setupY      = 58
	setupRowH   = 17
	setupValueX = setupX + 190 // where a row's value starts
	setupW      = totalW - 120
)

var (
	batchDurations  = []float64{60, 120, 300, 600, 1800, 3600}
	batchIterations = []int{20, 50, 100, 200, 500, 1000}
)

type setupScreen struct {
	open bool
	edit combat.Scenario
	sel  int    // selected row
	err  string // why the edited scenario cannot run, or ""

	artifactIDs []string // "" (empty slot) followed by every artifact, sorted
	enemyIDs    []string // "" (plain enemy) followed by every archetype, sorted

	// Benchmark settings. They apply to runs started from this screen only.
	durIdx, iterIdx int
}

func (s *setupScreen) init() {
	s.artifactIDs = []string{""}
	for id := range combat.ArtifactEffects {
		s.artifactIDs = append(s.artifactIDs, id)
	}
	sort.Strings(s.artifactIDs[1:])
	s.enemyIDs = []string{""}
	for id := range combat.EnemyBuilds {
		s.enemyIDs = append(s.enemyIDs, id)
	}
	sort.Strings(s.enemyIDs[1:])
	s.durIdx = slices.Index(batchDurations, 600)
	s.iterIdx = slices.Index(batchIterations, 100)
}

// cloneScenario copies a scenario so edits never reach the loaded original.
// Artifacts is padded to the full seven slots so any slot can be edited.
func cloneScenario(s combat.Scenario) combat.Scenario {
	arts := make([]string, 7)
	copy(arts, s.Artifacts)
	s.Artifacts = arts
	s.SkillRotation = slices.Clone(s.SkillRotation)
	s.StatPriority = slices.Clone(s.StatPriority)
	s.EnemyPool = slices.Clone(s.EnemyPool)
	return s
}

// isActive reports whether a slot holds something that can be cast.
func isActive(s *combat.Scenario, slot int) bool {
	if slot < 0 || slot >= len(s.Artifacts) || s.Artifacts[slot] == "" {
		return false
	}
	eff, ok := combat.ArtifactEffects[s.Artifacts[slot]]
	return ok && !(eff.IsPassive && eff.Cooldown == 0)
}

// fixRotation keeps the skill rotation consistent with the artifacts after an
// edit: presses of a slot that can no longer be cast become "auto", and a
// newly castable slot is added to the end.
func fixRotation(s *combat.Scenario) {
	var present [7]bool
	for i, entry := range s.SkillRotation {
		if entry == "auto" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(entry, "slot_%d", &n); err != nil || !isActive(s, n-1) {
			s.SkillRotation[i] = "auto"
			continue
		}
		present[n-1] = true
	}
	for slot := range present {
		if isActive(s, slot) && !present[slot] {
			s.SkillRotation = append(s.SkillRotation, fmt.Sprintf("slot_%d", slot+1))
		}
	}
	if len(s.SkillRotation) == 0 {
		s.SkillRotation = []string{"auto"}
	}
}

// setupRow is one editable line. adjust moves the value by dir (±1) steps;
// big asks for a larger step where that makes sense.
type setupRow struct {
	section string // heading drawn above this row, if any
	label   string
	value   string
	adjust  func(dir int, big bool)
}

// cycle moves an index through n options, wrapping around.
func cycle(idx, dir, n int) int { return ((idx+dir)%n + n) % n }

func intRow(section, label string, v *int, lo, hi, bigStep int, show func(int) string) setupRow {
	return setupRow{section: section, label: label, value: show(*v), adjust: func(dir int, big bool) {
		step := 1
		if big {
			step = bigStep
		}
		*v = clampI(*v+dir*step, lo, hi)
	}}
}

func plain(v int) string { return fmt.Sprint(v) }

func (g *VisGame) setupRows() []setupRow {
	s := &g.setup
	e := &s.edit

	classes := []string{"knight", "mage"}
	rows := []setupRow{
		{section: "BUILD", label: "Class", value: e.Class, adjust: func(dir int, _ bool) {
			e.Class = classes[cycle(slices.Index(classes, e.Class), dir, len(classes))]
		}},
		intRow("", "Strength", &e.Stats.Strength, -20, 500, 10, plain),
		intRow("", "Vitality", &e.Stats.Vitality, -20, 500, 10, plain),
		intRow("", "Intelligence", &e.Stats.Intelligence, -20, 500, 10, plain),
	}
	for slot := range e.Artifacts {
		label := fmt.Sprintf("Slot %d", slot+1)
		if slot == 6 {
			label = "Slot 7 (elite)"
		}
		value := e.Artifacts[slot]
		if value == "" {
			value = "(empty)"
		} else {
			value += "   " + effectDesc(combat.ArtifactEffects[e.Artifacts[slot]])
		}
		rows = append(rows, setupRow{label: label, value: truncate(value, 130), adjust: func(dir int, _ bool) {
			idx := max(slices.Index(s.artifactIDs, e.Artifacts[slot]), 0)
			e.Artifacts[slot] = s.artifactIDs[cycle(idx, dir, len(s.artifactIDs))]
		}})
	}

	pack := max(e.Pack, 1)
	ramp := int(e.RampSec)
	dist := int(e.Distance)
	enemy := e.EnemyBuildID
	if enemy == "" {
		enemy = "(plain enemy)"
	}
	leveling := "off"
	if e.Leveling {
		leveling = "on"
	}
	seed := int(g.seed)
	rows = append(rows,
		intRow("FIGHT", "Floor", &e.Floor, 1, 99, 5, plain),
		setupRow{label: "Wave size", value: fmt.Sprintf("%d enemies attacking at once", pack), adjust: func(dir int, _ bool) {
			e.Pack = clampI(pack+dir, 1, 20)
		}},
		setupRow{label: "Ramp", value: rampLabel(ramp), adjust: func(dir int, big bool) {
			step := 5
			if big {
				step = 30
			}
			e.RampSec = float64(clampI(ramp+dir*step, 0, 600))
		}},
		setupRow{label: "Enemy archetype", value: enemy, adjust: func(dir int, _ bool) {
			e.EnemyBuildID = s.enemyIDs[cycle(slices.Index(s.enemyIDs, e.EnemyBuildID), dir, len(s.enemyIDs))]
		}},
		setupRow{label: "Distance", value: distanceLabel(dist), adjust: func(dir int, _ bool) {
			e.Distance = float64(clampI(dist+dir, 0, 12))
		}},
		setupRow{label: "Mid-fight leveling", value: leveling, adjust: func(int, bool) { e.Leveling = !e.Leveling }},
		setupRow{label: "Seed", value: fmt.Sprintf("%d   (which fight is shown; benchmarks use it too)", seed), adjust: func(dir int, big bool) {
			step := 1
			if big {
				step = 10
			}
			g.seed = uint64(max(seed+dir*step, 0))
		}},
		setupRow{section: "BENCHMARK", label: "Length of each run", value: fmt.Sprintf("%.0fs", batchDurations[s.durIdx]), adjust: func(dir int, _ bool) {
			s.durIdx = clampI(s.durIdx+dir, 0, len(batchDurations)-1)
		}},
		setupRow{label: "Runs per build", value: fmt.Sprint(batchIterations[s.iterIdx]), adjust: func(dir int, _ bool) {
			s.iterIdx = clampI(s.iterIdx+dir, 0, len(batchIterations)-1)
		}},
	)
	return rows
}

func rampLabel(sec int) string {
	if sec == 0 {
		return "off (enemies stay at the floor above)"
	}
	return fmt.Sprintf("enemies gain a floor every %ds", sec)
}

func distanceLabel(tiles int) string {
	if tiles == 0 {
		return "class default (knight 1 tile, mage 5)"
	}
	return fmt.Sprintf("%d tiles from the enemy", tiles)
}

// setupRowY returns the y of each row, leaving room for section headings.
func setupRowY(rows []setupRow) []int {
	ys := make([]int, len(rows))
	y := setupY
	for i, r := range rows {
		if r.section != "" {
			y += setupRowH + 4
		}
		ys[i] = y
		y += setupRowH
	}
	return ys
}

// adjustSetup applies one change and re-checks that the scenario can run.
func (g *VisGame) adjustSetup(row setupRow, dir int, big bool) {
	row.adjust(dir, big)
	fixRotation(&g.setup.edit)
	g.setup.err = ""
	if err := g.setup.edit.Validate(); err != nil {
		g.setup.err = strings.ReplaceAll(err.Error(), "\n", "; ")
	}
}

func (g *VisGame) updateSetup() {
	s := &g.setup
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
		s.open = false
		return
	}
	rows := g.setupRows()
	ys := setupRowY(rows)
	big := ebiten.IsKeyPressed(ebiten.KeyShift)

	if keyRepeat(ebiten.KeyArrowDown) {
		s.sel = cycle(s.sel, 1, len(rows))
	}
	if keyRepeat(ebiten.KeyArrowUp) {
		s.sel = cycle(s.sel, -1, len(rows))
	}
	if keyRepeat(ebiten.KeyArrowRight) {
		g.adjustSetup(rows[s.sel], 1, big)
	}
	if keyRepeat(ebiten.KeyArrowLeft) {
		g.adjustSetup(rows[s.sel], -1, big)
	}

	// Mouse: hovering selects a row; left click or wheel up raises its value,
	// right click or wheel down lowers it.
	mx, my := ebiten.CursorPosition()
	if mx >= setupX && mx <= setupX+setupW {
		for i, y := range ys {
			if my < y || my >= y+setupRowH {
				continue
			}
			s.sel = i
			_, wheel := ebiten.Wheel()
			switch {
			case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || wheel > 0:
				g.adjustSetup(rows[i], 1, big)
			case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || wheel < 0:
				g.adjustSetup(rows[i], -1, big)
			}
		}
	}

	if s.err != "" {
		return // nothing below may run an invalid scenario
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
		g.active = cloneScenario(s.edit)
		g.reset()
		g.paused = false
		s.open = false
	case inpututil.IsKeyJustPressed(ebiten.KeyB):
		g.benchmarkEdited()
	case inpututil.IsKeyJustPressed(ebiten.KeyA):
		g.benchmarkAll()
	case inpututil.IsKeyJustPressed(ebiten.KeyF):
		g.benchmarkFloors()
	case inpututil.IsKeyJustPressed(ebiten.KeyM):
		g.benchmarkEnemies()
	}
}

func (g *VisGame) drawSetup(screen *ebiten.Image) {
	s := &g.setup
	dimScreen(screen)
	ebitenutil.DrawRect(screen, setupX-20, 14, setupW+40, totalH-28, color.RGBA{18, 18, 30, 255})
	ebitenutil.DrawRect(screen, setupX-20, 14, setupW+40, 22, colHeader)
	ebitenutil.DebugPrintAt(screen, "SIMULATION SETUP  -  "+truncate(s.edit.Name, 70), setupX-8, 18)
	ebitenutil.DebugPrintAt(screen, "Up/Down = select   Left/Right or click / right-click or wheel = change   Shift = bigger steps", setupX-8, 40)

	rows := g.setupRows()
	ys := setupRowY(rows)
	for i, r := range rows {
		y := ys[i]
		if r.section != "" {
			ebitenutil.DebugPrintAt(screen, r.section, setupX-8, y-setupRowH)
			ebitenutil.DrawRect(screen, setupX-8, float64(y-2), setupW+16, 1, colDivider)
		}
		if i == s.sel {
			ebitenutil.DrawRect(screen, setupX-8, float64(y), setupW+16, setupRowH, color.RGBA{50, 50, 90, 220})
		}
		ebitenutil.DebugPrintAt(screen, r.label, setupX, y)
		ebitenutil.DebugPrintAt(screen, "< "+r.value+" >", setupValueX, y)
	}

	y := ys[len(ys)-1] + setupRowH + 12
	ebitenutil.DrawRect(screen, setupX-8, float64(y-6), setupW+16, 1, colDivider)
	for _, line := range wrapText("Rotation: "+strings.Join(s.edit.SkillRotation, " "), 170) {
		ebitenutil.DebugPrintAt(screen, line, setupX, y)
		y += lineH
	}
	ebitenutil.DebugPrintAt(screen, "(one entry is pressed per tick; it follows the artifacts automatically. Edit the JSON for a custom order.)", setupX, y)
	y += lineH + 10

	if s.err != "" {
		for _, line := range wrapText("Cannot run: "+s.err, 170) {
			ebitenutil.DebugPrintAt(screen, line, setupX, y)
			y += lineH
		}
		ebitenutil.DebugPrintAt(screen, "Esc = close", setupX, totalH-34)
		return
	}
	if g.batch.running {
		ebitenutil.DebugPrintAt(screen, "A benchmark is running...", setupX, y)
	}
	ebitenutil.DebugPrintAt(screen, "Enter = watch this fight      B = benchmark this build      F = this build on floors 1-15      M = this build vs every enemy archetype", setupX, totalH-52)
	ebitenutil.DebugPrintAt(screen, "A = benchmark every loaded scenario with this wave size and ramp      Esc = close without applying", setupX, totalH-34)
}
