// Dungeoneer Build Benchmarker — standalone headless simulation tool.
//
// Usage:
//
//	benchmarker [flags]
//
// Flags:
//
//	--scenario <path>   Path to a single JSON scenario file.
//	                    Omit to run all *.json files in ./scenarios/.
//	--json              Output raw JSON instead of a human-readable table.
//
// Build:
//
//	cd src && go build ./cmd/benchmarker/
//
// Run:
//
//	./benchmarker
//	./benchmarker --scenario scenarios/knight_slash_build.json
//	./benchmarker --json
package main

import (
	"dungeoneer/combat"
	"flag"
	"fmt"
	"os"
)

// noopEngine is a compile-time fallback used when DefaultCombatEngine is not
// yet wired (e.g. during development of parallel sub-systems). It accepts all
// state transitions without modifying them so the benchmarker binary always
// compiles.
type noopEngine struct{}

func (n *noopEngine) Tick(state combat.CombatState, actions []combat.Action) (combat.CombatState, []combat.Event) {
	return state, nil
}

func main() {
	scenarioPath := flag.String("scenario", "", "Path to a single JSON scenario file (default: all files in ./scenarios/)")
	jsonOutput := flag.Bool("json", false, "Output raw JSON instead of a human-readable table")
	flag.Parse()

	// Resolve which engine to use.
	// DefaultCombatEngine exists in src/combat/engine.go; use it.
	// The noopEngine is kept as a compile-time safety net above.
	var engine combat.CombatEngine = combat.NewDefaultCombatEngine(0)

	// Load scenarios.
	var scenarios []combat.Scenario
	if *scenarioPath != "" {
		s, err := loadScenario(*scenarioPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		scenarios = []combat.Scenario{s}
	} else {
		var err error
		scenarios, err = loadAllScenarios("./scenarios")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	if !*jsonOutput {
		fmt.Println("=== Dungeoneer Build Benchmarker ===")
		fmt.Println()
	}

	hasError := false
	for _, s := range scenarios {
		if s.Iterations <= 0 {
			fmt.Fprintf(os.Stderr, "warning: scenario %q has zero iterations — skipping\n", s.Name)
			continue
		}

		result := runSimulation(engine, s)

		if *jsonOutput {
			if err := printJSON(result); err != nil {
				fmt.Fprintf(os.Stderr, "error encoding JSON for %q: %v\n", s.Name, err)
				hasError = true
			}
		} else {
			printTable(result)
			fmt.Println()
		}
	}

	if hasError {
		os.Exit(1)
	}
}
