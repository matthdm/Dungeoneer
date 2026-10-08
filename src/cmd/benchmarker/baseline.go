package main

import (
	"dungeoneer/combat"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"text/tabwriter"
)

// baselineName is the baseline file kept next to the scenarios directory.
const baselineName = "baseline.json"

// Baseline is a saved set of results that later runs are compared against.
type Baseline struct {
	Seed        uint64             `json:"seed"`
	DurationSec float64            `json:"duration_sec"`
	Results     []combat.SimResult `json:"results"`
}

func loadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, fmt.Errorf("reading baseline: %w", err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, fmt.Errorf("parsing baseline %s: %w", path, err)
	}
	return b, nil
}

func saveBaseline(path string, b Baseline) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// drift is one metric of one scenario that moved outside tolerance, or a
// scenario present on only one side.
type drift struct {
	Scenario string
	Metric   string
	Was, Now float64
	Note     string // set instead of the numbers for a scenario-level problem
}

// compareBaseline returns every way the results differ from the baseline by
// more than tolerancePct. Survival is compared in percentage points; the other
// metrics are compared relative to their baseline value. wholeSet says the run
// covered the full scenario directory, so a baseline entry with no result is
// itself a difference.
func compareBaseline(base Baseline, results []combat.SimResult, seed uint64, durationSec, tolerancePct float64, wholeSet bool) []drift {
	var drifts []drift
	if base.Seed != seed || base.DurationSec != durationSec {
		return []drift{{
			Scenario: "(all)",
			Note: fmt.Sprintf("baseline was recorded with seed %d and duration %.0fs, this run used seed %d and %.0fs",
				base.Seed, base.DurationSec, seed, durationSec),
		}}
	}

	was := make(map[string]combat.SimResult, len(base.Results))
	for _, r := range base.Results {
		was[r.ScenarioName] = r
	}
	seen := make(map[string]bool, len(results))
	for _, now := range results {
		seen[now.ScenarioName] = true
		old, ok := was[now.ScenarioName]
		if !ok {
			drifts = append(drifts, drift{Scenario: now.ScenarioName, Note: "not in the baseline"})
			continue
		}
		if old.Iterations != now.Iterations || old.Floor != now.Floor || old.EnemyBuild != now.EnemyBuild ||
			old.Pack != now.Pack || old.RampSec != now.RampSec {
			drifts = append(drifts, drift{Scenario: now.ScenarioName, Note: "scenario changed (iterations, floor, enemy build, pack or ramp) since the baseline"})
			continue
		}
		if math.Abs(now.SurvivalRate-old.SurvivalRate)*100 > tolerancePct {
			drifts = append(drifts, drift{Scenario: now.ScenarioName, Metric: "survival %", Was: old.SurvivalRate * 100, Now: now.SurvivalRate * 100})
		}
		for _, m := range []struct {
			name     string
			was, now float64
		}{
			{"dps", old.DPS, now.DPS},
			{"kills/min", old.KillsPerMinute, now.KillsPerMinute},
			{"damage taken/s", old.DamageTakenPerSec, now.DamageTakenPerSec},
			{"time alive (s)", old.AvgTimeAliveSec, now.AvgTimeAliveSec},
		} {
			if relDiffPct(m.was, m.now) > tolerancePct {
				drifts = append(drifts, drift{Scenario: now.ScenarioName, Metric: m.name, Was: m.was, Now: m.now})
			}
		}
	}
	for _, old := range base.Results {
		if wholeSet && !seen[old.ScenarioName] {
			drifts = append(drifts, drift{Scenario: old.ScenarioName, Note: "in the baseline but was not run"})
		}
	}
	return drifts
}

// relDiffPct is how far now is from was, as a percentage of was. A value that
// was zero and no longer is counts as fully changed.
func relDiffPct(was, now float64) float64 {
	if was == now {
		return 0
	}
	if was == 0 {
		return 100
	}
	return math.Abs(now-was) / math.Abs(was) * 100
}

func printDrifts(w io.Writer, drifts []drift) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SCENARIO\tMETRIC\tBASELINE\tNOW\tCHANGE")
	for _, d := range drifts {
		if d.Note != "" {
			fmt.Fprintf(tw, "%s\t%s\t\t\t\n", d.Scenario, d.Note)
			continue
		}
		change := fmt.Sprintf("%+.1f", d.Now-d.Was)
		if d.Metric != "survival %" && d.Was != 0 {
			change = fmt.Sprintf("%+.1f%%", (d.Now-d.Was)/math.Abs(d.Was)*100)
		}
		fmt.Fprintf(tw, "%s\t%s\t%.1f\t%.1f\t%s\n", d.Scenario, d.Metric, d.Was, d.Now, change)
	}
	tw.Flush()
}
