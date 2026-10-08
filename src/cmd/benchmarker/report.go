package main

import (
	"dungeoneer/combat"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"
)

// enemyLabel names the enemy a result was fought against.
func enemyLabel(id string) string {
	if id == "" {
		return "-"
	}
	return id
}

// printSummary prints one row per result.
func printSummary(w io.Writer, results []combat.SimResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "BUILD\tFLOOR\tENEMY\tPACK\tSURVIVAL\tALIVE\tDPS\tP10-P90\tKILLS/MIN\tDMG IN/S\tHEAL/S\tLOW HP\tOVERKILL")
	ramped := false
	for _, r := range results {
		floor := strconv.Itoa(r.Floor)
		if r.RampSec > 0 {
			ramped = true
			floor += "+"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%.1f%%\t%.0fs\t%.1f\t%.0f-%.0f\t%.1f\t%.1f\t%.1f\t%.0f%%\t%.0f%%\n",
			r.ScenarioName, floor, enemyLabel(r.EnemyBuild), r.Pack,
			r.SurvivalRate*100, r.AvgTimeAliveSec,
			r.DPS, r.DPSP10, r.DPSP90,
			r.KillsPerMinute, r.DamageTakenPerSec, r.HealingPerSec,
			r.MinHPPctAvg*100, r.OverkillPct*100)
	}
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "ALIVE = average seconds survived. DPS excludes overkill and is measured over time alive.")
	fmt.Fprintln(w, "LOW HP = average lowest HP reached, as a share of max HP.")
	if ramped {
		fmt.Fprintln(w, "FLOOR n+ = enemies start at floor n and gain a floor at the scenario's ramp interval.")
	}
}

// printDetail prints every metric for each result, including per-skill figures.
func printDetail(w io.Writer, results []combat.SimResult) {
	for _, r := range results {
		fmt.Fprintf(w, "Scenario: %s\n", r.ScenarioName)
		fmt.Fprintf(w, "  Floor %d, enemy %s, pack of %d, %d iterations of %.0fs, seed %d\n",
			r.Floor, enemyLabel(r.EnemyBuild), r.Pack, r.Iterations, r.DurationSec, r.Seed)
		if r.RampSec > 0 {
			fmt.Fprintf(w, "  Enemies gain a floor every %.0fs (floor %d at the average time of death or end)\n",
				r.RampSec, r.Floor+int(r.AvgTimeAliveSec/r.RampSec))
		}
		fmt.Fprintf(w, "  Survival:        %.1f%%  (avg %.1fs alive)\n", r.SurvivalRate*100, r.AvgTimeAliveSec)
		if r.FirstDeathIteration >= 0 {
			fmt.Fprintf(w, "  Deaths:          median at %.1fs; first in iteration %d (replay with --trace --trace-iter %d)\n",
				r.MedianDeathSec, r.FirstDeathIteration, r.FirstDeathIteration)
		}
		fmt.Fprintf(w, "  DPS:             %.1f  (p10 %.1f, p50 %.1f, p90 %.1f)\n", r.DPS, r.DPSP10, r.DPSP50, r.DPSP90)
		fmt.Fprintf(w, "  Raw DPS:         %.1f  (%.0f%% overkill)\n", r.RawDPS, r.OverkillPct*100)
		fmt.Fprintf(w, "  Kills/min:       %.1f  (%.1fs per kill)\n", r.KillsPerMinute, r.AvgClearTimeSec)
		fmt.Fprintf(w, "  Damage taken:    %.1f/s  (%.0f per run)\n", r.DamageTakenPerSec, r.AvgDamageTaken)
		fmt.Fprintf(w, "  Healing:         %.1f/s\n", r.HealingPerSec)
		fmt.Fprintf(w, "  Lowest HP:       avg %.0f%%, worst 5%% of runs %.0f%%\n", r.MinHPPctAvg*100, r.MinHPPctP05*100)
		fmt.Fprintf(w, "  Longest streak:  %.1f\n", r.StreakAvg)
		if r.AvgFinalLevel > 1 {
			fmt.Fprintf(w, "  Final level:     %.1f\n", r.AvgFinalLevel)
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
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  SKILL\tDAMAGE\tCASTS/MIN\tREADY BUT NOT CAST (s per run)")
		for _, id := range ids {
			sk := r.Skills[id]
			casts := "-"
			if id != combat.AutoAttack {
				casts = fmt.Sprintf("%.1f", sk.CastsPerMin)
			}
			fmt.Fprintf(tw, "  %s\t%.1f%%\t%s\t%s\n", id, sk.DamageShare*100, casts, blockedSummary(sk.BlockedSec))
		}
		tw.Flush()
		fmt.Fprintln(w)
	}
}

// blockedSummary renders a skill's blocked time as "reason 12.3, reason 4.0".
func blockedSummary(blocked map[string]float64) string {
	if len(blocked) == 0 {
		return "-"
	}
	reasons := make([]string, 0, len(blocked))
	for reason := range blocked {
		reasons = append(reasons, reason)
	}
	sort.Slice(reasons, func(i, j int) bool { return blocked[reasons[i]] > blocked[reasons[j]] })
	out := ""
	for i, reason := range reasons {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%s %.1f", reason, blocked[reason])
	}
	return out
}

// printPivot prints a build-by-column grid of "survival% / DPS" cells. column
// returns the column label of a result; columns fixes their order.
func printPivot(w io.Writer, results []combat.SimResult, columns []string, column func(combat.SimResult) string) {
	cells := make(map[string]map[string]string)
	var builds []string
	for _, r := range results {
		row, ok := cells[r.ScenarioName]
		if !ok {
			row = make(map[string]string)
			cells[r.ScenarioName] = row
			builds = append(builds, r.ScenarioName)
		}
		row[column(r)] = fmt.Sprintf("%.0f%% / %.0f", r.SurvivalRate*100, r.DPS)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprint(tw, "BUILD")
	for _, c := range columns {
		fmt.Fprintf(tw, "\t%s", c)
	}
	fmt.Fprintln(tw)
	for _, b := range builds {
		fmt.Fprint(tw, b)
		for _, c := range columns {
			fmt.Fprintf(tw, "\t%s", cells[b][c])
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Each cell is survival rate / DPS.")
}

// printJSON prints all results as one indented JSON array.
func printJSON(w io.Writer, results []combat.SimResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

// printCSV prints one row per result. Per-skill figures are left to JSON.
func printCSV(w io.Writer, results []combat.SimResult) error {
	cw := csv.NewWriter(w)
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 4, 64) }
	cw.Write([]string{
		"scenario", "floor", "enemy_build", "pack", "ramp_sec", "iterations", "seed", "duration_sec",
		"survival_rate", "avg_time_alive_sec", "median_death_sec",
		"dps", "dps_p10", "dps_p50", "dps_p90", "raw_dps", "overkill_pct",
		"kills_per_min", "avg_clear_time_sec",
		"avg_damage_taken", "damage_taken_per_sec", "healing_per_sec",
		"min_hp_pct_avg", "min_hp_pct_p05", "streak_avg", "avg_final_level",
	})
	for _, r := range results {
		cw.Write([]string{
			r.ScenarioName, strconv.Itoa(r.Floor), r.EnemyBuild, strconv.Itoa(r.Pack), f(r.RampSec), strconv.Itoa(r.Iterations),
			strconv.FormatUint(r.Seed, 10), f(r.DurationSec),
			f(r.SurvivalRate), f(r.AvgTimeAliveSec), f(r.MedianDeathSec),
			f(r.DPS), f(r.DPSP10), f(r.DPSP50), f(r.DPSP90), f(r.RawDPS), f(r.OverkillPct),
			f(r.KillsPerMinute), f(r.AvgClearTimeSec),
			f(r.AvgDamageTaken), f(r.DamageTakenPerSec), f(r.HealingPerSec),
			f(r.MinHPPctAvg), f(r.MinHPPctP05), f(r.StreakAvg), f(r.AvgFinalLevel),
		})
	}
	cw.Flush()
	return cw.Error()
}
