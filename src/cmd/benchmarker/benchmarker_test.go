package main

import (
	"bytes"
	"dungeoneer/combat"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseFloors(t *testing.T) {
	good := map[string][]int{
		"3":      {3},
		"1-4":    {1, 2, 3, 4},
		"3,5,8":  {3, 5, 8},
		"1-2, 7": {1, 2, 7},
		"2-2":    {2},
	}
	for spec, want := range good {
		got, err := parseFloors(spec)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("parseFloors(%q) = %v, %v; want %v", spec, got, err, want)
		}
	}
	for _, spec := range []string{"", "x", "0", "5-3", "1-", "-3", "1,,2"} {
		if got, err := parseFloors(spec); err == nil {
			t.Errorf("parseFloors(%q) = %v, want an error", spec, got)
		}
	}
}

func TestSweepAndMatrixExpandScenarios(t *testing.T) {
	scenarios := []combat.Scenario{
		{Name: "build", Floor: 2},
		{Name: "matchup", Floor: 2, EnemyBuildID: "the_judge"},
	}

	swept := sweepFloors(scenarios, []int{1, 5})
	if len(swept) != 4 || swept[0].Floor != 1 || swept[1].Floor != 5 || swept[3].Name != "matchup" {
		t.Fatalf("sweepFloors gave %+v", swept)
	}
	if scenarios[0].Floor != 2 {
		t.Fatal("sweepFloors changed its input")
	}

	matrix := enemyMatrix(scenarios)
	if want := len(combat.EnemyBuilds) + 1; len(matrix) != want {
		t.Fatalf("enemyMatrix gave %d rows, want %d (one build against a plain enemy and every archetype)", len(matrix), want)
	}
	if matrix[0].EnemyBuildID != "" {
		t.Fatalf("first matrix column should be the plain enemy, got %q", matrix[0].EnemyBuildID)
	}
	for _, s := range matrix {
		if s.Name != "build" {
			t.Fatalf("a single matchup leaked into the matrix: %+v", s)
		}
	}
}

func TestLoadScenariosRejectsDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	body := `{"name":"same","class":"knight","floor":1,"iterations":1}`
	for _, f := range []string{"a.json", "b.json"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := loadScenarios(dir); err == nil || !strings.Contains(err.Error(), `"same"`) {
		t.Fatalf("want a duplicate-name error, got %v", err)
	}
}

func TestLoadScenariosSurfacesValidationErrors(t *testing.T) {
	dir := t.TempDir()
	body := `{"name":"bad","class":"wizard","floor":1,"iterations":1}`
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadScenarios(dir); err == nil || !strings.Contains(err.Error(), "wizard") {
		t.Fatalf("want a validation error, got %v", err)
	}
}

func result(name string, survival, dps float64) combat.SimResult {
	return combat.SimResult{
		ScenarioName: name, Floor: 3, Iterations: 100, Pack: 1,
		SurvivalRate: survival, DPS: dps, KillsPerMinute: 10, DamageTakenPerSec: 5, AvgTimeAliveSec: 120,
	}
}

func TestCompareBaseline(t *testing.T) {
	base := Baseline{Seed: 0, DurationSec: 120, Results: []combat.SimResult{
		result("a", 1.0, 100),
		result("b", 0.5, 50),
	}}

	t.Run("identical results pass", func(t *testing.T) {
		if d := compareBaseline(base, base.Results, 0, 120, 5, true); len(d) != 0 {
			t.Fatalf("unexpected drift: %+v", d)
		}
	})
	t.Run("small moves are within tolerance", func(t *testing.T) {
		now := []combat.SimResult{result("a", 0.97, 103), result("b", 0.5, 50)}
		if d := compareBaseline(base, now, 0, 120, 5, true); len(d) != 0 {
			t.Fatalf("unexpected drift: %+v", d)
		}
	})
	t.Run("large moves are reported per metric", func(t *testing.T) {
		now := []combat.SimResult{result("a", 0.80, 120), result("b", 0.5, 50)}
		d := compareBaseline(base, now, 0, 120, 5, true)
		if len(d) != 2 || d[0].Metric != "survival %" || d[1].Metric != "dps" || d[1].Was != 100 || d[1].Now != 120 {
			t.Fatalf("want survival and dps drift for a, got %+v", d)
		}
	})
	t.Run("new and missing scenarios are reported", func(t *testing.T) {
		now := []combat.SimResult{result("a", 1.0, 100), result("c", 1.0, 10)}
		d := compareBaseline(base, now, 0, 120, 5, true)
		if len(d) != 2 || d[0].Scenario != "c" || d[1].Scenario != "b" {
			t.Fatalf("want c (new) and b (missing), got %+v", d)
		}
	})
	t.Run("a partial run does not report the rest as missing", func(t *testing.T) {
		if d := compareBaseline(base, base.Results[:1], 0, 120, 5, false); len(d) != 0 {
			t.Fatalf("unexpected drift: %+v", d)
		}
	})
	t.Run("a changed scenario is reported", func(t *testing.T) {
		changed := result("a", 1.0, 100)
		changed.Floor = 9
		d := compareBaseline(base, []combat.SimResult{changed, base.Results[1]}, 0, 120, 5, true)
		if len(d) != 1 || !strings.Contains(d[0].Note, "scenario changed") {
			t.Fatalf("want a scenario-changed note, got %+v", d)
		}
	})
	t.Run("a changed pack is reported", func(t *testing.T) {
		changed := result("a", 1.0, 100)
		changed.Pack = 3
		d := compareBaseline(base, []combat.SimResult{changed, base.Results[1]}, 0, 120, 5, true)
		if len(d) != 1 || !strings.Contains(d[0].Note, "scenario changed") {
			t.Fatalf("want a scenario-changed note, got %+v", d)
		}
	})
	t.Run("different run settings are refused", func(t *testing.T) {
		d := compareBaseline(base, base.Results, 7, 120, 5, true)
		if len(d) != 1 || !strings.Contains(d[0].Note, "seed") {
			t.Fatalf("want one settings-mismatch note, got %+v", d)
		}
	})
}

func TestRelDiffPct(t *testing.T) {
	for _, tc := range []struct{ was, now, want float64 }{
		{100, 100, 0}, {100, 110, 10}, {100, 90, 10}, {0, 0, 0}, {0, 5, 100},
	} {
		if got := relDiffPct(tc.was, tc.now); got < tc.want-1e-9 || got > tc.want+1e-9 {
			t.Errorf("relDiffPct(%v, %v) = %v, want %v", tc.was, tc.now, got, tc.want)
		}
	}
}

func TestBaselineRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.json")
	want := Baseline{Seed: 3, DurationSec: 60, Results: []combat.SimResult{result("a", 1, 100)}}
	want.Results[0].Skills = map[string]combat.SkillStats{"x": {CastsPerMin: 2, BlockedSec: map[string]float64{"no_mana": 1.5}}}
	if err := saveBaseline(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := loadBaseline(path)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip gave %+v, %v", got, err)
	}
}

func TestOutputFormats(t *testing.T) {
	results := combat.SimulateAll([]combat.Scenario{{
		Name: "fmt build", Class: "knight", Artifacts: []string{"ironbreaker_gauntlets"},
		Stats: combat.BaseStats{Strength: 12, Vitality: 12}, Floor: 2, Iterations: 3,
		SkillRotation: []string{"slot_1", "auto"},
	}}, combat.SimOptions{DurationSec: 5})

	var buf bytes.Buffer
	if err := printJSON(&buf, results); err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil || len(decoded) != 1 {
		t.Fatalf("JSON output is not an array of one result: %v\n%s", err, buf.String())
	}
	for _, key := range []string{"scenario", "survival_rate", "dps", "kills_per_min", "skills"} {
		if _, ok := decoded[0][key]; !ok {
			t.Errorf("JSON result has no %q key", key)
		}
	}

	buf.Reset()
	if err := printCSV(&buf, results); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil || len(rows) != 2 || len(rows[0]) != len(rows[1]) || rows[1][0] != "fmt build" {
		t.Fatalf("CSV output malformed: %v %v", rows, err)
	}

	buf.Reset()
	printSummary(&buf, results)
	printDetail(&buf, results)
	for _, want := range []string{"BUILD", "fmt build", "ironbreaker_gauntlets", "auto_attack"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("text output is missing %q", want)
		}
	}
}

func TestSortResults(t *testing.T) {
	results := []combat.SimResult{result("b", 0.5, 10), result("a", 1.0, 30), result("c", 0.7, 20)}
	names := func() string {
		out := ""
		for _, r := range results {
			out += r.ScenarioName
		}
		return out
	}
	for by, want := range map[string]string{"name": "abc", "dps": "acb", "survival": "acb"} {
		if err := sortResults(results, by); err != nil || names() != want {
			t.Errorf("sort by %s gave %s (%v), want %s", by, names(), err, want)
		}
	}
	if err := sortResults(results, "nonsense"); err == nil {
		t.Error("an unknown sort column was accepted")
	}
}
