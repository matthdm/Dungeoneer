package combat

import (
	"math"
	"sort"
)

// SimResult holds aggregate metrics from one benchmarker scenario run.
// Rates are measured over the time the player was alive, so a build that dies
// early is not diluted by the seconds it spent dead.
type SimResult struct {
	ScenarioName string  `json:"scenario"`
	Floor        int     `json:"floor"`
	EnemyBuild   string  `json:"enemy_build,omitempty"`
	Pack         int     `json:"pack"`               // enemies attacking at once
	RampSec      float64 `json:"ramp_sec,omitempty"` // seconds per enemy floor gained; 0 = fixed floor
	Iterations   int     `json:"iterations"`
	Seed         uint64  `json:"seed"`
	DurationSec  float64 `json:"duration_sec"`

	SurvivalRate    float64 `json:"survival_rate"`      // 0.0–1.0 fraction of iterations the player survived
	AvgTimeAliveSec float64 `json:"avg_time_alive_sec"` // equals DurationSec when nobody died
	MedianDeathSec  float64 `json:"median_death_sec"`   // median time of death among iterations that died; 0 if none

	// FirstDeathIteration is the first iteration that died, or -1. Replay it
	// with SimOptions.TraceIteration to see why.
	FirstDeathIteration int `json:"first_death_iteration"`

	DPS         float64 `json:"dps"`     // damage that removed enemy HP, per second alive
	DPSP10      float64 `json:"dps_p10"` // per-iteration DPS percentiles
	DPSP50      float64 `json:"dps_p50"`
	DPSP90      float64 `json:"dps_p90"`
	RawDPS      float64 `json:"raw_dps"`      // as rolled, including overkill
	OverkillPct float64 `json:"overkill_pct"` // share of rolled damage wasted on dead enemies

	KillsPerMinute  float64 `json:"kills_per_min"`
	AvgClearTimeSec float64 `json:"avg_clear_time_sec"` // seconds alive per kill; 0 if no kills

	AvgDamageTaken    float64 `json:"avg_damage_taken"` // per iteration
	DamageTakenPerSec float64 `json:"damage_taken_per_sec"`
	HealingPerSec     float64 `json:"healing_per_sec"`
	MinHPPctAvg       float64 `json:"min_hp_pct_avg"` // lowest HP reached, as a fraction of max
	MinHPPctP05       float64 `json:"min_hp_pct_p05"` // the worst 5% of iterations dipped at least this low

	StreakAvg     float64 `json:"streak_avg"`      // mean of each iteration's longest kill streak
	AvgFinalLevel float64 `json:"avg_final_level"` // 1 unless the scenario enables leveling

	// Skills is keyed by artifact ID, plus AutoAttack for basic attacks.
	Skills map[string]SkillStats `json:"skills"`
}

// SkillStats describes how one skill performed across a run.
type SkillStats struct {
	CastsPerMin float64 `json:"casts_per_min"` // casts that actually fired
	DamageShare float64 `json:"damage_share"`  // fraction of effective damage
	// BlockedSec is the average time per iteration the skill sat off cooldown
	// but could not be cast, by reason (a Fail* constant, BlockNoMana or
	// BlockEnemy). An execute waiting for a low target shows up here.
	BlockedSec map[string]float64 `json:"blocked_sec,omitempty"`
}

// percentile returns the p-th percentile (0–1) of sorted values, interpolating
// between neighbours. sorted must be ascending.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	pos := p * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(pos-float64(lo))
}

// aggregate folds per-iteration measurements into a SimResult.
func aggregate(s Scenario, seed uint64, duration float64, runs []iterStats, totals *simTotals) SimResult {
	n := float64(len(runs))
	res := SimResult{
		ScenarioName: s.Name,
		Floor:        s.Floor,
		EnemyBuild:   s.EnemyBuildID,
		Pack:         max(s.Pack, 1),
		RampSec:      s.RampSec,
		Iterations:   len(runs),
		Seed:         seed,
		DurationSec:  duration,
		Skills:       make(map[string]SkillStats),
	}

	var kills, rawDmg, effDmg, dmgTaken, healed, survived int
	var alive, streak, minHP, level float64
	dps := make([]float64, 0, len(runs))
	minHPs := make([]float64, 0, len(runs))
	var deaths []float64
	res.FirstDeathIteration = -1
	for i, r := range runs {
		if r.died && res.FirstDeathIteration < 0 {
			res.FirstDeathIteration = i
		}
		kills += r.kills
		rawDmg += r.rawDmg
		effDmg += r.effDmg
		dmgTaken += r.dmgTaken
		healed += r.healed
		alive += r.timeAlive
		streak += float64(r.maxStreak)
		minHP += r.minHPPct
		level += float64(r.level)
		if r.died {
			deaths = append(deaths, r.timeAlive)
		} else {
			survived++
		}
		if r.timeAlive > 0 {
			dps = append(dps, float64(r.effDmg)/r.timeAlive)
		}
		minHPs = append(minHPs, r.minHPPct)
	}
	sort.Float64s(dps)
	sort.Float64s(minHPs)
	sort.Float64s(deaths)

	res.SurvivalRate = float64(survived) / n
	res.AvgTimeAliveSec = alive / n
	res.MedianDeathSec = percentile(deaths, 0.5)
	res.DPSP10 = percentile(dps, 0.10)
	res.DPSP50 = percentile(dps, 0.50)
	res.DPSP90 = percentile(dps, 0.90)
	res.AvgDamageTaken = float64(dmgTaken) / n
	res.MinHPPctAvg = minHP / n
	res.MinHPPctP05 = percentile(minHPs, 0.05)
	res.StreakAvg = streak / n
	res.AvgFinalLevel = level / n
	if alive > 0 {
		res.DPS = float64(effDmg) / alive
		res.RawDPS = float64(rawDmg) / alive
		res.KillsPerMinute = float64(kills) / alive * 60
		res.DamageTakenPerSec = float64(dmgTaken) / alive
		res.HealingPerSec = float64(healed) / alive
	}
	if rawDmg > 0 {
		res.OverkillPct = float64(rawDmg-effDmg) / float64(rawDmg)
	}
	if kills > 0 {
		res.AvgClearTimeSec = alive / float64(kills)
	}

	// Per-skill figures. The rotation presses one entry per tick, so a slot
	// that appears k times in a rotation of length L is pressed every L/k
	// ticks; that converts blocked presses into blocked seconds.
	pressesPerTick := [maxArtifactSlots]float64{}
	if rotLen := len(s.SkillRotation); rotLen > 0 {
		for _, entry := range s.SkillRotation {
			if slot := parseSlotIdx(entry); slot >= 0 && slot < maxArtifactSlots {
				pressesPerTick[slot] += 1 / float64(rotLen)
			}
		}
	}
	share := func(tag string) float64 {
		if effDmg == 0 {
			return 0
		}
		return float64(totals.dmgByTag[tag]) / float64(effDmg)
	}
	for slot := range maxArtifactSlots {
		if slot >= len(s.Artifacts) || s.Artifacts[slot] == "" {
			continue
		}
		id := s.Artifacts[slot]
		if pressesPerTick[slot] == 0 && totals.dmgByTag[id] == 0 {
			continue // never pressed and dealt nothing: a pure passive
		}
		st := res.Skills[id]
		st.DamageShare = share(id)
		if alive > 0 {
			st.CastsPerMin += float64(totals.casts[slot]) / alive * 60
		}
		for ri, presses := range totals.blocked[slot] {
			if presses == 0 {
				continue
			}
			if st.BlockedSec == nil {
				st.BlockedSec = make(map[string]float64)
			}
			st.BlockedSec[blockReasons[ri]] += float64(presses) / pressesPerTick[slot] * tickDuration / n
		}
		res.Skills[id] = st
	}
	res.Skills[AutoAttack] = SkillStats{DamageShare: share(AutoAttack)}

	return res
}
