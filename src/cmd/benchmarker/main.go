// Dungeoneer Build Benchmarker — standalone headless simulation tool.
//
// It runs each scenario (a build, a floor and optionally an enemy archetype)
// through the combat engine many times and reports how the build performs.
// Runs are seeded, so the same scenarios and flags always print the same
// numbers.
//
// Run from src/, from this directory or from the repository root:
//
//	go run ./cmd/benchmarker                     one summary row per scenario
//	go run ./cmd/benchmarker --format detail     every metric, with per-skill figures
//	go run ./cmd/benchmarker --check             compare against baseline.json; exit 1 on drift
//	go run ./cmd/benchmarker --update-baseline   record the current numbers as the baseline
//	go run ./cmd/benchmarker --floors 1-10       find the floor where each build falls over
//	go run ./cmd/benchmarker --matrix            every build against every enemy archetype
//	go run ./cmd/benchmarker --pack 3 --ramp 30 --duration 1800
//	                                             wave stress test: how long each build lasts
//	go run ./cmd/benchmarker --scenario scenarios/the_55.json --trace
//
// Flags:
//
//	--scenario <path>    A scenario file or a directory of them (default: the scenarios directory).
//	--format <name>      table (default), detail, json or csv.
//	--sort <column>      file (default), name, dps, survival or kpm.
//	--seed <n>           Vary the random stream (default 0).
//	--iterations <n>     Override every scenario's iteration count.
//	--duration <sec>     Length of one iteration (default 120).
//	--leveling           Let the player level up mid-run in every scenario.
//	--pack <n>           Enemies attacking at once in every scenario (a wave). Area and chain
//	                     skills hit the whole wave; a root only holds the enemy being fought.
//	--ramp <sec>         Raise the enemy floor by one every <sec> seconds in every scenario.
//	--floors <list>      Run every scenario at each floor, e.g. 1-10 or 3,5,8.
//	--matrix             Run every build against every enemy archetype.
//	--baseline <path>    Baseline file (default: baseline.json beside the scenarios directory).
//	--check              Fail if results drift from the baseline by more than --tolerance.
//	--update-baseline    Write the results to the baseline file.
//	--tolerance <pct>    Allowed drift for --check (default 5).
//	--trace              Print the event log of one iteration of a single scenario.
//	--trace-iter <n>     Which iteration --trace follows (default 0).
package main

import (
	"dungeoneer/combat"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// errDrift is returned when --check finds results outside tolerance.
var errDrift = errors.New("results drifted from the baseline")

func run() error {
	scenarioPath := flag.String("scenario", "", "scenario file or directory (default: the scenarios directory)")
	format := flag.String("format", "table", "output format: table, detail, json or csv")
	jsonOutput := flag.Bool("json", false, "shorthand for --format json")
	sortBy := flag.String("sort", "file", "sort rows by: file, name, dps, survival or kpm")
	seed := flag.Uint64("seed", 0, "vary the random stream")
	iterations := flag.Int("iterations", 0, "override every scenario's iteration count")
	duration := flag.Float64("duration", 120, "length of one iteration in seconds")
	leveling := flag.Bool("leveling", false, "let the player level up mid-run in every scenario")
	pack := flag.Int("pack", 0, "enemies attacking at once in every scenario")
	ramp := flag.Float64("ramp", 0, "raise the enemy floor by one every this many seconds in every scenario")
	floors := flag.String("floors", "", "run every scenario at each floor, e.g. 1-10 or 3,5,8")
	matrix := flag.Bool("matrix", false, "run every build against every enemy archetype")
	baselinePath := flag.String("baseline", "", "baseline file (default: baseline.json beside the scenarios directory)")
	check := flag.Bool("check", false, "fail if results drift from the baseline by more than --tolerance")
	update := flag.Bool("update-baseline", false, "write the results to the baseline file")
	tolerance := flag.Float64("tolerance", 5, "allowed drift for --check, in percent")
	trace := flag.Bool("trace", false, "print the event log of one iteration of a single scenario")
	traceIter := flag.Int("trace-iter", 0, "which iteration --trace follows")
	flag.Parse()

	if flag.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", flag.Arg(0))
	}
	if *jsonOutput {
		*format = "json"
	}
	switch *format {
	case "table", "detail", "json", "csv":
	default:
		return fmt.Errorf("--format %q is not table, detail, json or csv", *format)
	}
	if *duration <= 0 {
		return fmt.Errorf("--duration %g must be positive", *duration)
	}
	if *iterations < 0 || *traceIter < 0 {
		return errors.New("--iterations and --trace-iter must not be negative")
	}
	usesBaseline := *check || *update
	if *check && *update {
		return errors.New("--check and --update-baseline cannot be combined")
	}
	if usesBaseline && (*floors != "" || *matrix || *iterations > 0 || *leveling || *trace || *pack > 0 || *ramp > 0) {
		return errors.New("the baseline covers the scenarios as written; drop --floors, --matrix, --iterations, --leveling, --pack, --ramp and --trace")
	}
	if *pack < 0 || *ramp < 0 {
		return errors.New("--pack and --ramp must not be negative")
	}
	if *floors != "" && *matrix {
		return errors.New("--floors and --matrix cannot be combined")
	}

	if *scenarioPath == "" {
		dir, err := findScenarioDir()
		if err != nil {
			return err
		}
		*scenarioPath = dir
	}
	scenarios, scenarioDir, err := loadScenarios(*scenarioPath)
	wholeSet := filepath.Clean(scenarioDir) == filepath.Clean(*scenarioPath)
	if err != nil {
		return err
	}
	for i := range scenarios {
		if *iterations > 0 {
			scenarios[i].Iterations = *iterations
		}
		if *leveling {
			scenarios[i].Leveling = true
		}
		if *pack > 0 {
			scenarios[i].Pack = *pack
		}
		if *ramp > 0 {
			scenarios[i].RampSec = *ramp
		}
		if err := scenarios[i].Validate(); err != nil {
			return fmt.Errorf("scenario %q: %w", scenarios[i].Name, err)
		}
	}
	opts := combat.SimOptions{Seed: *seed, DurationSec: *duration}

	if *trace {
		if len(scenarios) != 1 || *floors != "" || *matrix {
			return errors.New("--trace follows one iteration of one scenario; pass --scenario <file> without --floors or --matrix")
		}
		// Only the traced iteration needs to run; iterations are independently seeded.
		scenarios[0].Iterations = *traceIter + 1
		opts.Trace = os.Stdout
		opts.TraceIteration = *traceIter
		fmt.Printf("Trace of %q, iteration %d, seed %d\n\n", scenarios[0].Name, *traceIter, *seed)
		combat.Simulate(scenarios[0], opts)
		return nil
	}

	var floorList []int
	if *floors != "" {
		if floorList, err = parseFloors(*floors); err != nil {
			return err
		}
		scenarios = sweepFloors(scenarios, floorList)
	}
	if *matrix {
		if scenarios = enemyMatrix(scenarios); len(scenarios) == 0 {
			return errors.New("--matrix needs at least one scenario without an enemy_build")
		}
	}

	results := combat.SimulateAll(scenarios, opts)
	if err := sortResults(results, *sortBy); err != nil {
		return err
	}

	if *baselinePath == "" {
		*baselinePath = filepath.Join(filepath.Dir(filepath.Clean(scenarioDir)), baselineName)
	}
	if *update {
		if err := saveBaseline(*baselinePath, Baseline{Seed: *seed, DurationSec: *duration, Results: results}); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "baseline written to %s (%d scenarios)\n", *baselinePath, len(results))
	}

	switch {
	case *format == "json":
		err = printJSON(os.Stdout, results)
	case *format == "csv":
		err = printCSV(os.Stdout, results)
	case *format == "detail":
		printDetail(os.Stdout, results)
	case floorList != nil:
		columns := make([]string, len(floorList))
		for i, f := range floorList {
			columns[i] = "F" + strconv.Itoa(f)
		}
		printPivot(os.Stdout, results, columns, func(r combat.SimResult) string { return "F" + strconv.Itoa(r.Floor) })
	case *matrix:
		ids := enemyBuildIDs()
		columns := make([]string, len(ids))
		for i, id := range ids {
			columns[i] = enemyLabel(id)
		}
		printPivot(os.Stdout, results, columns, func(r combat.SimResult) string { return enemyLabel(r.EnemyBuild) })
	default:
		printSummary(os.Stdout, results)
	}
	if err != nil {
		return err
	}

	if *check {
		base, err := loadBaseline(*baselinePath)
		if err != nil {
			return err
		}
		drifts := compareBaseline(base, results, *seed, *duration, *tolerance, wholeSet)
		if len(drifts) > 0 {
			fmt.Fprintf(os.Stderr, "\n%d difference(s) from %s beyond %.1f%%:\n", len(drifts), *baselinePath, *tolerance)
			printDrifts(os.Stderr, drifts)
			return errDrift
		}
		fmt.Fprintf(os.Stderr, "\nbaseline check passed: %d scenarios within %.1f%% of %s\n", len(results), *tolerance, *baselinePath)
	}
	return nil
}
