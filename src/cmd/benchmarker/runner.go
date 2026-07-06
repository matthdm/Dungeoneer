package main

import (
	"dungeoneer/combat"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// loadScenario reads a combat.Scenario from a JSON file.
func loadScenario(path string) (combat.Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return combat.Scenario{}, fmt.Errorf("reading %s: %w", path, err)
	}
	var s combat.Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return combat.Scenario{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return s, nil
}

// loadAllScenarios loads all *.json files from a directory.
func loadAllScenarios(dir string) ([]combat.Scenario, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("globbing %s: %w", dir, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no *.json files found in %s", dir)
	}
	var scenarios []combat.Scenario
	for _, path := range entries {
		s, err := loadScenario(path)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}
	return scenarios, nil
}

// runSimulation delegates to combat.RunSimulation so the benchmarker and
// the game use identical engine logic and enemy scaling.
func runSimulation(engine combat.CombatEngine, s combat.Scenario) combat.SimResult {
	return combat.RunSimulation(engine, s)
}

// printTable prints a human-readable SimResult summary to stdout.
func printTable(result combat.SimResult) {
	fmt.Printf("Scenario: %s\n", result.ScenarioName)
	fmt.Printf("  Iterations:       %d\n", result.Iterations)
	fmt.Printf("  Survival Rate:    %.1f%%\n", result.SurvivalRate*100)
	fmt.Printf("  Avg Clear Time:   %.1fs\n", result.AvgClearTimeSec)
	fmt.Printf("  DPS (avg):        %.1f\n", result.AvgDPSTick)
	fmt.Printf("  Kills/min:        %.1f\n", result.KillsPerMinute)
	fmt.Printf("  Avg Dmg Taken:    %.0f\n", result.AvgDamageTaken)
	fmt.Printf("  Streak Avg:       %.1f\n", result.StreakAvg)
	if len(result.SkillUsageRatio) > 0 {
		fmt.Println("  Skill Usage:")
		keys := make([]string, 0, len(result.SkillUsageRatio))
		for k := range result.SkillUsageRatio {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("    %-20s%.1f%%\n", k, result.SkillUsageRatio[k]*100)
		}
	}
	fmt.Println("---")
}

// printJSON prints a SimResult as indented JSON to stdout.
func printJSON(result combat.SimResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
