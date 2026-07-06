// combatvis — visual replay of a single combat scenario.
//
// Usage:
//
//	go run ./cmd/combatvis --scenario cmd/benchmarker/scenarios/iron_flurry.json
//	go run ./cmd/combatvis --scenario cmd/benchmarker/scenarios/the_55.json
//
// Controls:
//
//	Space      pause / unpause
//	+/-        speed up / slow down (0.1× to 8×)
//	R          restart current scenario
//	←/→        previous / next scenario (when no --scenario flag)
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
