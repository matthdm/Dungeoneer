package main

import (
	"dungeoneer/combat"
	"fmt"
	"image/color"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Benchmarks started from the setup screen run combat.SimulateAll, the same
// call the benchmarker CLI makes, on a background goroutine so the window
// stays responsive. One runs at a time; its report stays viewable with V.

const (
	batchX     = 30
	batchY     = 64
	batchLines = (totalH - batchY - 44) / lineH
)

type batchReport struct {
	title string
	lines []string
}

type batchRunner struct {
	running bool
	visible bool
	title   string // of the run in progress, or of the report shown
	lines   []string
	scroll  int
	started time.Time
	took    time.Duration
	done    chan batchReport
}

func (b *batchRunner) hasResults() bool { return len(b.lines) > 0 }

// start runs the scenarios in the background and opens the report overlay.
// render turns the results into the lines to show.
func (b *batchRunner) start(title string, scenarios []combat.Scenario, opts combat.SimOptions, render func([]combat.SimResult) []string) {
	if b.running {
		b.visible = true
		return
	}
	b.running, b.visible = true, true
	b.title, b.lines, b.scroll = title, nil, 0
	b.started = time.Now()
	b.done = make(chan batchReport, 1)
	done := b.done
	go func() {
		done <- batchReport{title: title, lines: render(combat.SimulateAll(scenarios, opts))}
	}()
}

// poll collects a finished run. Called every frame.
func (b *batchRunner) poll() {
	if !b.running {
		return
	}
	select {
	case rep := <-b.done:
		b.running = false
		b.title, b.lines = rep.title, rep.lines
		b.took = time.Since(b.started)
	default:
	}
}

func (b *batchRunner) update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyV) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		b.visible = false
		return
	}
	_, wheel := ebiten.Wheel()
	switch {
	case keyRepeat(ebiten.KeyArrowDown) || wheel < 0:
		b.scroll++
	case keyRepeat(ebiten.KeyArrowUp) || wheel > 0:
		b.scroll--
	case inpututil.IsKeyJustPressed(ebiten.KeyPageDown):
		b.scroll += batchLines
	case inpututil.IsKeyJustPressed(ebiten.KeyPageUp):
		b.scroll -= batchLines
	}
	b.scroll = clampI(b.scroll, 0, max(len(b.lines)-batchLines, 0))
}

func (b *batchRunner) draw(screen *ebiten.Image) {
	dimScreen(screen)
	ebitenutil.DrawRect(screen, batchX-16, 14, totalW-2*(batchX-16), totalH-28, color.RGBA{18, 18, 30, 255})
	ebitenutil.DrawRect(screen, batchX-16, 14, totalW-2*(batchX-16), 22, colHeader)
	ebitenutil.DebugPrintAt(screen, "BENCHMARK  -  "+b.title, batchX-6, 18)

	if b.running {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Running... %.0fs", time.Since(b.started).Seconds()), batchX, batchY)
		ebitenutil.DebugPrintAt(screen, "Esc = close and keep watching (the benchmark carries on; press V to come back)", batchX, totalH-34)
		return
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Finished in %.1fs.", b.took.Seconds()), batchX, 42)
	end := min(b.scroll+batchLines, len(b.lines))
	for i, line := range b.lines[b.scroll:end] {
		ebitenutil.DebugPrintAt(screen, line, batchX, batchY+i*lineH)
	}
	footer := "Esc = close"
	if len(b.lines) > batchLines {
		footer += fmt.Sprintf("   Up/Down/PgUp/PgDn/wheel = scroll (%d-%d of %d)", b.scroll+1, end, len(b.lines))
	}
	ebitenutil.DebugPrintAt(screen, footer, batchX, totalH-34)
}

// ── What can be benchmarked from the setup screen ─────────────────────────

// batchSettings returns the edited scenario prepared for a benchmark, the
// options for it, and a description of the run settings.
func (g *VisGame) batchSettings() (combat.Scenario, combat.SimOptions, string) {
	s := &g.setup
	sc := cloneScenario(s.edit)
	sc.Iterations = batchIterations[s.iterIdx]
	opts := combat.SimOptions{Seed: g.seed, DurationSec: batchDurations[s.durIdx]}
	return sc, opts, fmt.Sprintf("%d runs of %.0fs, seed %d", sc.Iterations, opts.DurationSec, g.seed)
}

func (g *VisGame) benchmarkEdited() {
	sc, opts, settings := g.batchSettings()
	g.batch.start(truncate(sc.Name, 60)+"  ("+settings+")", []combat.Scenario{sc}, opts, func(res []combat.SimResult) []string {
		return detailLines(res[0])
	})
}

func (g *VisGame) benchmarkFloors() {
	sc, opts, settings := g.batchSettings()
	var scenarios []combat.Scenario
	for floor := 1; floor <= 15; floor++ {
		s := sc
		s.Floor = floor
		scenarios = append(scenarios, s)
	}
	g.batch.start(truncate(sc.Name, 50)+" on floors 1-15  ("+settings+")", scenarios, opts, func(res []combat.SimResult) []string {
		return tableLines(res, false)
	})
}

func (g *VisGame) benchmarkEnemies() {
	sc, opts, settings := g.batchSettings()
	var scenarios []combat.Scenario
	for _, id := range g.setup.enemyIDs {
		s := sc
		s.EnemyBuildID = id
		scenarios = append(scenarios, s)
	}
	g.batch.start(truncate(sc.Name, 50)+" against every enemy archetype  ("+settings+")", scenarios, opts, func(res []combat.SimResult) []string {
		return tableLines(res, false)
	})
}

// benchmarkAll runs every loaded scenario as written, except for the wave
// size and ramp chosen on the setup screen.
func (g *VisGame) benchmarkAll() {
	sc, opts, settings := g.batchSettings()
	scenarios := make([]combat.Scenario, len(g.scenarios))
	for i, loaded := range g.scenarios {
		s := cloneScenario(loaded)
		s.Pack, s.RampSec, s.Iterations = sc.Pack, sc.RampSec, sc.Iterations
		scenarios[i] = s
	}
	wave := fmt.Sprintf("wave of %d", max(sc.Pack, 1))
	if sc.RampSec > 0 {
		wave += fmt.Sprintf(", +1 floor every %.0fs", sc.RampSec)
	}
	g.batch.start("every scenario, "+wave+"  ("+settings+")", scenarios, opts, func(res []combat.SimResult) []string {
		return tableLines(res, true)
	})
}

// ── Report text ───────────────────────────────────────────────────────────

func enemyLabel(id string) string {
	if id == "" {
		return "-"
	}
	return id
}

// tableLines renders one row per result. byEndurance sorts the longest-lived
// builds first; otherwise rows keep their order (a floor or enemy sweep).
func tableLines(results []combat.SimResult, byEndurance bool) []string {
	if byEndurance {
		sort.SliceStable(results, func(i, j int) bool {
			a, b := results[i], results[j]
			if a.AvgTimeAliveSec != b.AvgTimeAliveSec {
				return a.AvgTimeAliveSec > b.AvgTimeAliveSec
			}
			return a.DamageTakenPerSec-a.HealingPerSec < b.DamageTakenPerSec-b.HealingPerSec
		})
	}
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "BUILD\tFLOOR\tENEMY\tWAVE\tSURVIVAL\tALIVE\tDPS\tKILLS/MIN\tDMG IN/S\tHEAL/S\tLOW HP")
	for _, r := range results {
		floor := fmt.Sprint(r.Floor)
		if r.RampSec > 0 {
			floor += "+"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%.0f%%\t%.0fs\t%.1f\t%.1f\t%.1f\t%.1f\t%.0f%%\n",
			truncate(r.ScenarioName, 46), floor, enemyLabel(r.EnemyBuild), r.Pack,
			r.SurvivalRate*100, r.AvgTimeAliveSec, r.DPS, r.KillsPerMinute,
			r.DamageTakenPerSec, r.HealingPerSec, r.MinHPPctAvg*100)
	}
	tw.Flush()
	lines := strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
	return append(lines, "",
		"ALIVE = average seconds survived. DPS excludes overkill and is measured over time alive.",
		"LOW HP = average lowest HP reached, as a share of max HP. FLOOR n+ = enemies ramp up from floor n.")
}

// detailLines renders every metric of one result, with per-skill figures.
func detailLines(r combat.SimResult) []string {
	lines := []string{
		fmt.Sprintf("Floor %d, enemy %s, wave of %d", r.Floor, enemyLabel(r.EnemyBuild), r.Pack),
	}
	if r.RampSec > 0 {
		lines = append(lines, fmt.Sprintf("Enemies gain a floor every %.0fs (floor %d at the average time of death or end)",
			r.RampSec, r.Floor+int(r.AvgTimeAliveSec/r.RampSec)))
	}
	lines = append(lines, "",
		fmt.Sprintf("Survival          %.1f%%   (average %.1fs alive of %.0fs)", r.SurvivalRate*100, r.AvgTimeAliveSec, r.DurationSec))
	if r.FirstDeathIteration >= 0 {
		lines = append(lines, fmt.Sprintf("Deaths            median at %.1fs", r.MedianDeathSec))
	}
	lines = append(lines,
		fmt.Sprintf("DPS               %.1f   (p10 %.1f, p50 %.1f, p90 %.1f)", r.DPS, r.DPSP10, r.DPSP50, r.DPSP90),
		fmt.Sprintf("Raw DPS           %.1f   (%.0f%% overkill)", r.RawDPS, r.OverkillPct*100),
		fmt.Sprintf("Kills per minute  %.1f   (%.1fs per kill)", r.KillsPerMinute, r.AvgClearTimeSec),
		fmt.Sprintf("Damage taken      %.1f/s", r.DamageTakenPerSec),
		fmt.Sprintf("Healing           %.1f/s", r.HealingPerSec),
		fmt.Sprintf("Lowest HP         average %.0f%%, worst 5%% of runs %.0f%%", r.MinHPPctAvg*100, r.MinHPPctP05*100),
		fmt.Sprintf("Longest streak    %.1f", r.StreakAvg),
	)
	if r.AvgFinalLevel > 1 {
		lines = append(lines, fmt.Sprintf("Final level       %.1f", r.AvgFinalLevel))
	}

	ids := make([]string, 0, len(r.Skills))
	for id := range r.Skills {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := r.Skills[ids[i]], r.Skills[ids[j]]
		if a.DamageShare != b.DamageShare {
			return a.DamageShare > b.DamageShare
		}
		return ids[i] < ids[j]
	})
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SKILL\tDAMAGE\tCASTS/MIN\tREADY BUT NOT CAST (seconds per run)")
	for _, id := range ids {
		sk := r.Skills[id]
		casts := "-"
		if id != combat.AutoAttack {
			casts = fmt.Sprintf("%.1f", sk.CastsPerMin)
		}
		reasons := make([]string, 0, len(sk.BlockedSec))
		for reason := range sk.BlockedSec {
			reasons = append(reasons, reason)
		}
		sort.Slice(reasons, func(i, j int) bool { return sk.BlockedSec[reasons[i]] > sk.BlockedSec[reasons[j]] })
		blocked := "-"
		for i, reason := range reasons {
			if i == 0 {
				blocked = ""
			} else {
				blocked += ", "
			}
			blocked += fmt.Sprintf("%s %.1f", reason, sk.BlockedSec[reason])
		}
		fmt.Fprintf(tw, "%s\t%.1f%%\t%s\t%s\n", id, sk.DamageShare*100, casts, blocked)
	}
	tw.Flush()
	lines = append(lines, "")
	return append(lines, strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")...)
}
