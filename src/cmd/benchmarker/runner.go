package main

import (
	"dungeoneer/combat"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// scenarioDirs are tried in order when --scenario is not given, so the tool
// works from its own directory, from src/ and from the repository root.
var scenarioDirs = []string{
	"scenarios",
	filepath.Join("cmd", "benchmarker", "scenarios"),
	filepath.Join("src", "cmd", "benchmarker", "scenarios"),
}

// findScenarioDir returns the first default scenario directory that exists.
func findScenarioDir() (string, error) {
	for _, dir := range scenarioDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("no scenario directory found (looked for %s); pass --scenario", strings.Join(scenarioDirs, ", "))
}

// loadScenarios loads one scenario file, or every *.json file in a directory.
// It returns the directory the scenarios came from. Scenario names must be
// unique: they key the baseline and seed the random stream.
func loadScenarios(path string) ([]combat.Scenario, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}
	files := []string{path}
	dir := filepath.Dir(path)
	if info.IsDir() {
		dir = path
		files, err = filepath.Glob(filepath.Join(path, "*.json"))
		if err != nil {
			return nil, "", fmt.Errorf("globbing %s: %w", path, err)
		}
		if len(files) == 0 {
			return nil, "", fmt.Errorf("no *.json files found in %s", path)
		}
		sort.Strings(files)
	}

	scenarios := make([]combat.Scenario, 0, len(files))
	seen := make(map[string]string, len(files))
	for _, file := range files {
		s, err := combat.LoadScenario(file)
		if err != nil {
			return nil, "", err
		}
		if prev, dup := seen[s.Name]; dup {
			return nil, "", fmt.Errorf("%s and %s both use the scenario name %q", prev, file, s.Name)
		}
		seen[s.Name] = file
		scenarios = append(scenarios, s)
	}
	return scenarios, dir, nil
}

// parseFloors parses a floor list such as "1-10", "3,5,8" or "1-3,8".
func parseFloors(spec string) ([]int, error) {
	var floors []int
	for _, part := range strings.Split(spec, ",") {
		lo, hi, isRange := strings.Cut(strings.TrimSpace(part), "-")
		first, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("floors %q: %q is not a number", spec, lo)
		}
		last := first
		if isRange {
			if last, err = strconv.Atoi(hi); err != nil {
				return nil, fmt.Errorf("floors %q: %q is not a number", spec, hi)
			}
		}
		if first < 1 || last < first {
			return nil, fmt.Errorf("floors %q: %q is not a valid range of floors from 1 up", spec, part)
		}
		for f := first; f <= last; f++ {
			floors = append(floors, f)
		}
	}
	return floors, nil
}

// sweepFloors returns every scenario repeated at each floor.
func sweepFloors(scenarios []combat.Scenario, floors []int) []combat.Scenario {
	out := make([]combat.Scenario, 0, len(scenarios)*len(floors))
	for _, s := range scenarios {
		for _, f := range floors {
			s.Floor = f
			out = append(out, s)
		}
	}
	return out
}

// enemyBuildIDs returns "" (a plain enemy) followed by every registered enemy
// archetype in name order.
func enemyBuildIDs() []string {
	ids := make([]string, 0, len(combat.EnemyBuilds)+1)
	for id := range combat.EnemyBuilds {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return append([]string{""}, ids...)
}

// enemyMatrix returns every build scenario paired with every enemy archetype.
// Scenarios that already name an enemy build are single matchups, not builds,
// and are left out.
func enemyMatrix(scenarios []combat.Scenario) []combat.Scenario {
	enemies := enemyBuildIDs()
	var out []combat.Scenario
	for _, s := range scenarios {
		if s.EnemyBuildID != "" {
			continue
		}
		for _, id := range enemies {
			s.EnemyBuildID = id
			out = append(out, s)
		}
	}
	return out
}

// sortResults orders results by the named column. Ties keep their input order.
func sortResults(results []combat.SimResult, by string) error {
	var less func(a, b combat.SimResult) bool
	switch by {
	case "", "file":
		return nil
	case "name":
		less = func(a, b combat.SimResult) bool { return a.ScenarioName < b.ScenarioName }
	case "dps":
		less = func(a, b combat.SimResult) bool { return a.DPS > b.DPS }
	case "survival":
		less = func(a, b combat.SimResult) bool { return a.SurvivalRate > b.SurvivalRate }
	case "kpm":
		less = func(a, b combat.SimResult) bool { return a.KillsPerMinute > b.KillsPerMinute }
	default:
		return fmt.Errorf("--sort %q is not file, name, dps, survival or kpm", by)
	}
	sort.SliceStable(results, func(i, j int) bool { return less(results[i], results[j]) })
	return nil
}
