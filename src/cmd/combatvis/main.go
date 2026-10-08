// combatvis — watch, edit and benchmark combat scenarios in a window.
//
// Usage:
//
//	go run ./cmd/combatvis --scenario cmd/benchmarker/scenarios/iron_flurry.json
//	go run ./cmd/combatvis --scenario cmd/benchmarker/scenarios/the_55.json
//
// It runs the same simulation as cmd/benchmarker (combat.Sim), so a fight
// shown here is one of the iterations behind a benchmark number.
//
// Controls:
//
//	Space      pause / unpause
//	+/-        speed up / slow down (0.125x to 64x)
//	R          restart the fight
//	N          next seed: a different fight of the same scenario
//	Left/Right previous / next scenario (when no --scenario flag)
//	Tab        pick a scenario from a list
//	E          setup screen: edit the build (class, stats, artifacts) and the
//	           fight (floor, wave size, ramp, enemy archetype), then watch it or
//	           run benchmarks without leaving the window
//	V          reopen the last benchmark report
package main

import (
	"dungeoneer/combat"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	scenarioPath := flag.String("scenario", "", "Path to a single JSON scenario file (default: cycle all ./cmd/benchmarker/scenarios/*.json)")
	flag.Parse()

	var scenarios []combat.Scenario
	if *scenarioPath != "" {
		s, err := combat.LoadScenario(*scenarioPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "combatvis: %v\n", err)
			os.Exit(1)
		}
		scenarios = []combat.Scenario{s}
	} else {
		// Default: load all scenarios from the benchmarker directory.
		entries, _ := filepath.Glob("cmd/benchmarker/scenarios/*.json")
		sort.Strings(entries)
		for _, p := range entries {
			s, err := combat.LoadScenario(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "combatvis: skipping %s: %v\n", p, err)
				continue
			}
			scenarios = append(scenarios, s)
		}
		if len(scenarios) == 0 {
			fmt.Fprintln(os.Stderr, "combatvis: no scenarios found — run from src/ or pass --scenario")
			os.Exit(1)
		}
	}

	g := newVisGame(scenarios)

	ebiten.SetWindowSize(totalW, totalH)
	ebiten.SetWindowTitle("Combat Visualizer")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
