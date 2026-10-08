package combat

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// idleEngine leaves the state untouched, so a test can watch what the
// simulation itself does around the engine (enemy hits, regen, mana).
type idleEngine struct{}

func (idleEngine) Tick(state CombatState, _ []Action) (CombatState, []Event) { return state, nil }

func knightScenario() Scenario {
	return Scenario{
		Name:          "test knight",
		Class:         "knight",
		Artifacts:     []string{"ironbreaker_gauntlets"},
		Stats:         BaseStats{Strength: 12, Vitality: 12},
		Floor:         3,
		Iterations:    20,
		SkillRotation: []string{"slot_1", "auto"},
	}
}

func newTestRun(s Scenario, engine CombatEngine) *simRun {
	totals := &simTotals{dmgByTag: make(map[string]int64)}
	return newSimRun(s, engine, rand.New(rand.NewPCG(1, 2)), totals, nil)
}

// ─── Validation ──────────────────────────────────────────────────────────────

func TestValidateAcceptsAGoodScenario(t *testing.T) {
	if err := knightScenario().Validate(); err != nil {
		t.Fatalf("valid scenario rejected: %v", err)
	}
}

func TestValidateReportsEachProblem(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Scenario)
		want   string
	}{
		{"empty name", func(s *Scenario) { s.Name = "" }, "name is empty"},
		{"unknown class", func(s *Scenario) { s.Class = "wizard" }, `class "wizard"`},
		{"floor zero", func(s *Scenario) { s.Floor = 0 }, "floor 0"},
		{"no iterations", func(s *Scenario) { s.Iterations = 0 }, "iterations 0"},
		{"negative distance", func(s *Scenario) { s.Distance = -1 }, "distance -1"},
		{"unknown artifact", func(s *Scenario) { s.Artifacts = []string{"nope"} }, `unknown artifact "nope"`},
		{"too many artifacts", func(s *Scenario) { s.Artifacts = make([]string, 8) }, "8 artifacts"},
		{"junk rotation entry", func(s *Scenario) { s.SkillRotation = []string{"bogus"} }, `"bogus" is not`},
		{"slot out of range", func(s *Scenario) { s.SkillRotation = []string{"slot_9"} }, `"slot_9" is not`},
		{"empty slot pressed", func(s *Scenario) { s.SkillRotation = []string{"slot_2"} }, "slot_2 is empty"},
		{"passive pressed", func(s *Scenario) {
			s.Artifacts = []string{"soul_harvest"}
			s.SkillRotation = []string{"slot_1"}
		}, "a passive with nothing to activate"},
		{"bad stat priority", func(s *Scenario) { s.StatPriority = []string{"luck"} }, `"luck" is not`},
		{"unknown enemy build", func(s *Scenario) { s.EnemyBuildID = "nope" }, `unknown enemy build "nope"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := knightScenario()
			tc.mutate(&s)
			err := s.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want an error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestValidateJoinsAllProblems(t *testing.T) {
	s := knightScenario()
	s.Class = "wizard"
	s.Floor = 0
	err := s.Validate()
	if err == nil || !strings.Contains(err.Error(), "class") || !strings.Contains(err.Error(), "floor") {
		t.Fatalf("want both problems reported, got %v", err)
	}
}

func TestLoadScenarioRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "typo.json")
	body := `{"name":"x","class":"knight","floor":1,"iterations":1,"skil_rotation":["auto"]}`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenario(path); err == nil || !strings.Contains(err.Error(), "skil_rotation") {
		t.Fatalf("want an unknown-field error naming skil_rotation, got %v", err)
	}
}

func TestShippedScenariosLoad(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "cmd", "benchmarker", "scenarios", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no shipped scenarios found (err %v)", err)
	}
	for _, f := range files {
		if _, err := LoadScenario(f); err != nil {
			t.Errorf("%v", err)
		}
	}
}

func TestParseSlotIdx(t *testing.T) {
	for entry, want := range map[string]int{
		"slot_1": 0, "slot_7": 6, "slot_0": -1, "slot_": -1, "slot_x": -1, "bogus": -1, "3": -1, "slot_2x": -1,
	} {
		if got := parseSlotIdx(entry); got != want {
			t.Errorf("parseSlotIdx(%q) = %d, want %d", entry, got, want)
		}
	}
}

// ─── Reproducibility ─────────────────────────────────────────────────────────

func TestSimulateIsReproducible(t *testing.T) {
	s := knightScenario()
	first := Simulate(s, SimOptions{})

	// Running something else in between must not disturb the result.
	other := knightScenario()
	other.Name = "another build"
	Simulate(other, SimOptions{})

	if again := Simulate(s, SimOptions{}); !reflect.DeepEqual(first, again) {
		t.Fatalf("same scenario and seed gave different results:\n%+v\n%+v", first, again)
	}
	if seeded := Simulate(s, SimOptions{Seed: 7}); seeded.DPS == first.DPS && seeded.AvgDamageTaken == first.AvgDamageTaken {
		t.Fatal("a different seed gave an identical result")
	}
}

// ─── Enemy model ─────────────────────────────────────────────────────────────

// A build that one-shots every enemy must still be attacked: a kill does not
// reset the enemy attack timer.
func TestFastKillsDoNotAvoidDamage(t *testing.T) {
	s := knightScenario()
	s.Stats.Strength = 500
	res := Simulate(s, SimOptions{DurationSec: 20})
	if res.KillsPerMinute < 60 {
		t.Fatalf("test build should kill faster than once a second, got %.1f kills/min", res.KillsPerMinute)
	}
	if res.AvgDamageTaken <= 0 {
		t.Fatal("a build that kills in under the enemy attack interval took no damage")
	}
}

func TestEnemyHitsVary(t *testing.T) {
	r := newTestRun(knightScenario(), idleEngine{})
	seen := make(map[int]bool)
	r.state.PlayerMaxHP, r.state.PlayerHP = 1_000_000, 1_000_000
	prevHP := r.state.PlayerHP
	base := enemyDmgPerAttack(r.s.Floor)
	for range 60 * 60 {
		r.step()
		if hit := prevHP - r.state.PlayerHP; hit > 0 {
			seen[hit] = true
			if lo, hi := float64(base)*(1-enemyDmgVariance)-1, float64(base)*(1+enemyDmgVariance)+1; float64(hit) < lo || float64(hit) > hi {
				t.Fatalf("enemy hit for %d, outside ±%.0f%% of %d", hit, enemyDmgVariance*100, base)
			}
		}
		prevHP = r.state.PlayerHP
	}
	if len(seen) < 2 {
		t.Fatalf("every enemy hit landed for the same damage: %v", seen)
	}
}

// 40 HP/s is under one point per 60 Hz tick; it must still heal 40 a second.
func TestEnemyRegenAccumulatesFractions(t *testing.T) {
	s := knightScenario()
	s.EnemyBuildID = "the_regenerator"
	r := newTestRun(s, idleEngine{})
	r.state.TargetHP = 10
	for range 60 {
		r.step()
	}
	if got, want := r.state.TargetHP, 10+EnemyBuilds["the_regenerator"].Ability.HPRegenPerSec; got < want-1 || got > want {
		t.Fatalf("after one second the enemy has %d HP, want about %d", got, want)
	}
}

func TestSilencerBlocksMeleeButNotRanged(t *testing.T) {
	melee := knightScenario()
	melee.EnemyBuildID = "the_silencer"
	res := Simulate(melee, SimOptions{DurationSec: 10})
	sk := res.Skills["ironbreaker_gauntlets"]
	if sk.CastsPerMin != 0 || sk.BlockedSec[BlockEnemy] <= 0 {
		t.Fatalf("knight inside the silence radius: casts/min %.1f, blocked %v", sk.CastsPerMin, sk.BlockedSec)
	}

	ranged := Scenario{
		Name: "test mage", Class: "mage", Artifacts: []string{"arcane_bolt"},
		Stats: BaseStats{Intelligence: 20, Vitality: 20}, Floor: 1, Iterations: 5,
		SkillRotation: []string{"slot_1"}, EnemyBuildID: "the_silencer",
	}
	if res := Simulate(ranged, SimOptions{DurationSec: 10}); res.Skills["arcane_bolt"].CastsPerMin <= 0 {
		t.Fatal("mage outside the silence radius never cast")
	}
}

func TestOutOfRangeBuildDoesNothing(t *testing.T) {
	s := knightScenario()
	s.Distance = 4 // beyond a knight's 1.5-tile reach
	res := Simulate(s, SimOptions{DurationSec: 10})
	if res.DPS != 0 {
		t.Fatalf("a knight 4 tiles away dealt %.1f DPS", res.DPS)
	}
	if res.Skills["ironbreaker_gauntlets"].BlockedSec[FailOutOfRange] <= 0 {
		t.Fatalf("slam should be reported out of range, got %v", res.Skills["ironbreaker_gauntlets"].BlockedSec)
	}
}

// ─── Mana ────────────────────────────────────────────────────────────────────

func TestManaStarvationIsReported(t *testing.T) {
	s := Scenario{
		Name: "dry mage", Class: "mage", Artifacts: []string{"lightning_storm"},
		Stats: BaseStats{Vitality: 200}, Floor: 1, Iterations: 3,
		SkillRotation: []string{"slot_1"},
	}
	eff := ArtifactEffects["lightning_storm"]
	if eff.ManaCost <= baseMana/2 {
		t.Skipf("lightning_storm costs %d mana; test needs a spell a 0-INT mage cannot sustain", eff.ManaCost)
	}
	res := Simulate(s, SimOptions{DurationSec: 60})
	if res.Skills["lightning_storm"].BlockedSec[BlockNoMana] <= 0 {
		t.Fatalf("a 0-INT mage spamming a %d-mana spell was never short of mana: %+v", eff.ManaCost, res.Skills["lightning_storm"])
	}
}

func TestManaRegenMatchesPlayerFormula(t *testing.T) {
	s := knightScenario()
	s.Stats.Intelligence = 10 // regen = 1 + 10/5 = 3 per second
	r := newTestRun(s, idleEngine{})
	if want := baseMana + 10*manaPerINT; r.state.PlayerMaxMana != want {
		t.Fatalf("max mana %d, want %d", r.state.PlayerMaxMana, want)
	}
	r.state.PlayerMana = 0
	for range 60 {
		r.step()
	}
	if r.state.PlayerMana < 2 || r.state.PlayerMana > 3 {
		t.Fatalf("after one second mana is %d, want about 3", r.state.PlayerMana)
	}
}

// ─── Metrics ─────────────────────────────────────────────────────────────────

func TestDPSExcludesOverkill(t *testing.T) {
	s := knightScenario()
	s.Stats.Strength = 500 // every hit is far more than the enemy's HP
	s.Iterations = 1
	res := Simulate(s, SimOptions{DurationSec: 20})
	kills := res.KillsPerMinute * res.AvgTimeAliveSec / 60
	maxEffective := (kills + 1) * float64(enemyHP(s.Floor))
	if dealt := res.DPS * res.AvgTimeAliveSec; dealt > maxEffective+1e-6 {
		t.Fatalf("effective damage %.0f exceeds the HP of every enemy fought (%.0f)", dealt, maxEffective)
	}
	if res.RawDPS <= res.DPS || res.OverkillPct <= 0 {
		t.Fatalf("overkill not reported: dps %.1f raw %.1f overkill %.2f", res.DPS, res.RawDPS, res.OverkillPct)
	}
}

func TestDeadBuildIsMeasuredOverTimeAlive(t *testing.T) {
	s := knightScenario()
	s.Stats = BaseStats{Strength: 1, Vitality: 0}
	s.Floor = 20
	res := Simulate(s, SimOptions{})
	if res.SurvivalRate != 0 {
		t.Fatalf("test build should always die, survival %.2f", res.SurvivalRate)
	}
	if res.AvgTimeAliveSec <= 0 || res.AvgTimeAliveSec >= res.DurationSec/2 {
		t.Fatalf("avg time alive %.1fs is not an early death", res.AvgTimeAliveSec)
	}
	if res.MedianDeathSec <= 0 || res.FirstDeathIteration != 0 {
		t.Fatalf("death not reported: median %.1f, first iteration %d", res.MedianDeathSec, res.FirstDeathIteration)
	}
	// 11 damage per 0.8s swing is about 14 DPS while alive; diluted over the
	// full two minutes it would be under 2.
	if res.DPS < 5 {
		t.Fatalf("DPS %.1f looks averaged over the whole run instead of time alive", res.DPS)
	}
}

func TestSurvivorReportsNoDeath(t *testing.T) {
	s := knightScenario()
	s.Stats.Vitality = 10_000
	res := Simulate(s, SimOptions{DurationSec: 10})
	if res.SurvivalRate != 1 || res.FirstDeathIteration != -1 || res.MedianDeathSec != 0 {
		t.Fatalf("survivor reported a death: %+v", res)
	}
	if res.AvgTimeAliveSec < 9.9 || res.AvgTimeAliveSec > 10.1 {
		t.Fatalf("avg time alive %.2f, want the full 10s", res.AvgTimeAliveSec)
	}
}

func TestLevelingIsOffUnlessAsked(t *testing.T) {
	s := knightScenario()
	s.Stats = BaseStats{Strength: 100, Vitality: 1000}
	if res := Simulate(s, SimOptions{DurationSec: 30}); res.AvgFinalLevel != 1 {
		t.Fatalf("leveling is off but the player reached level %.1f", res.AvgFinalLevel)
	}
	s.Leveling = true
	if res := Simulate(s, SimOptions{DurationSec: 30}); res.AvgFinalLevel <= 1 {
		t.Fatal("leveling is on but the player never leveled")
	}
}

func TestSkillStatsCountCastsNotPresses(t *testing.T) {
	s := knightScenario()
	s.Stats.Vitality = 10_000
	res := Simulate(s, SimOptions{DurationSec: 60})
	// Slam has a 3s cooldown: about 20 casts a minute, however often it is pressed.
	if got := res.Skills["ironbreaker_gauntlets"].CastsPerMin; got < 15 || got > 21 {
		t.Fatalf("slam casts/min = %.1f, want about 20", got)
	}
	total := 0.0
	for _, sk := range res.Skills {
		total += sk.DamageShare
	}
	if total < 0.999 || total > 1.001 {
		t.Fatalf("damage shares sum to %.3f, want 1", total)
	}
}

func TestPercentile(t *testing.T) {
	vals := []float64{10, 20, 30, 40, 50}
	for p, want := range map[float64]float64{0: 10, 0.5: 30, 1: 50, 0.25: 20, 0.125: 15} {
		if got := percentile(vals, p); got != want {
			t.Errorf("percentile(%v) = %v, want %v", p, got, want)
		}
	}
	if got := percentile(nil, 0.5); got != 0 {
		t.Errorf("percentile of nothing = %v, want 0", got)
	}
}

// ─── Trace ───────────────────────────────────────────────────────────────────

func TestTraceFollowsOneIteration(t *testing.T) {
	var log strings.Builder
	s := knightScenario()
	s.Iterations = 3
	Simulate(s, SimOptions{DurationSec: 5, Trace: &log, TraceIteration: 2})
	out := log.String()
	for _, want := range []string{"CAST ironbreaker_gauntlets", "auto_attack", "enemy hits for"} {
		if !strings.Contains(out, want) {
			t.Errorf("trace is missing %q:\n%s", want, out)
		}
	}
	// One iteration of 5 seconds: timestamps never restart.
	if strings.Count(out, "[   0.02s]") > 3 {
		t.Errorf("trace appears to cover more than one iteration:\n%s", out)
	}
}

// ─── Waves ───────────────────────────────────────────────────────────────────

func TestPackOfOneMatchesDefault(t *testing.T) {
	s := knightScenario()
	one := s
	one.Pack = 1
	if a, b := Simulate(s, SimOptions{}), Simulate(one, SimOptions{}); !reflect.DeepEqual(a, b) {
		t.Fatalf("pack 1 differs from the default:\n%+v\n%+v", a, b)
	}
}

func TestPackMultipliesIncomingDamage(t *testing.T) {
	s := knightScenario()
	s.Stats.Vitality = 100_000
	single := Simulate(s, SimOptions{DurationSec: 60})
	s.Pack = 3
	wave := Simulate(s, SimOptions{DurationSec: 60})
	if ratio := wave.DamageTakenPerSec / single.DamageTakenPerSec; ratio < 2.7 || ratio > 3.3 {
		t.Fatalf("three enemies dealt %.2fx the damage of one, want about 3x", ratio)
	}
	if wave.Pack != 3 || single.Pack != 1 {
		t.Fatalf("result packs = %d and %d, want 3 and 1", wave.Pack, single.Pack)
	}
}

// A root holds only the enemy being fought; shadow hides the player from all.
func TestRootHoldsOnlyTheTarget(t *testing.T) {
	damageOver3s := func(pack int, setup func(*CombatState)) int {
		s := knightScenario()
		s.Pack = pack
		r := newTestRun(s, idleEngine{})
		setup(&r.state)
		start := r.state.PlayerHP
		for range 180 {
			r.step()
		}
		return start - r.state.PlayerHP
	}
	root := func(st *CombatState) { st.TargetRooted, st.RootTimer = true, 99 }
	shadow := func(st *CombatState) { st.InShadow, st.ShadowTimer = true, 99 }

	if got := damageOver3s(1, root); got != 0 {
		t.Fatalf("a rooted lone enemy dealt %d damage", got)
	}
	if got := damageOver3s(3, root); got <= 0 {
		t.Fatal("rooting the target stopped the whole wave")
	}
	if got := damageOver3s(3, shadow); got != 0 {
		t.Fatalf("the wave dealt %d damage to a player in shadow", got)
	}
}

func TestSplashCount(t *testing.T) {
	for id, want := range map[string]int{
		"fireball":        4, // area
		"lightning_storm": 4, // field
		"lightning":       2, // chain of 3: the target and two more
		"arcane_bolt":     0, // single target
	} {
		eff, ok := ArtifactEffects[id]
		if !ok {
			t.Fatalf("artifact %q is not registered", id)
		}
		if got := splashCount(eff, 4); got != want {
			t.Errorf("splashCount(%s, 4 others) = %d, want %d", id, got, want)
		}
	}
	if got := splashCount(ArtifactEffects["lightning"], 1); got != 1 {
		t.Errorf("a chain cannot hit more enemies than exist, got %d", got)
	}
}

func TestSplashDamagesAndKillsTheWave(t *testing.T) {
	s := knightScenario()
	s.Artifacts = []string{"ironbreaker_gauntlets", "soul_harvest"}
	s.Pack = 3
	r := newTestRun(s, idleEngine{})
	r.state.PlayerHP = 10
	full := r.others[0].maxHP

	r.splash(Event{Type: EventDamageDealt, Value: 5, Tag: "ironbreaker_gauntlets"})
	if r.others[0].hp != full-5 || r.others[1].hp != full-5 || r.stats.effDmg != 10 {
		t.Fatalf("a 5-point slam left the wave at %+v with %d effective damage", r.others, r.stats.effDmg)
	}

	r.others[0].hp = 3
	r.splash(Event{Type: EventDamageDealt, Value: 5, Tag: "ironbreaker_gauntlets"})
	if r.stats.kills != 1 || r.state.KillStreak != 1 {
		t.Fatalf("splash kill not counted: kills %d, streak %d", r.stats.kills, r.state.KillStreak)
	}
	if r.others[0].hp != full {
		t.Fatalf("a fresh enemy should replace the dead one, got %d HP", r.others[0].hp)
	}
	if r.stats.effDmg != 10+3+5 || r.stats.rawDmg != 20 {
		t.Fatalf("overkill on the wave miscounted: effective %d, raw %d", r.stats.effDmg, r.stats.rawDmg)
	}
	if r.state.PlayerHP <= 10 || r.stats.healed <= 0 {
		t.Fatalf("soul_harvest did not heal on a splash kill (HP %d, healed %d)", r.state.PlayerHP, r.stats.healed)
	}

	r.splash(Event{Type: EventDamageDealt, Value: 5}) // a basic attack hits one enemy
	if r.stats.rawDmg != 20 {
		t.Fatal("an untagged basic attack splashed")
	}
}

func TestAreaBuildKillsFasterInAWave(t *testing.T) {
	s := knightScenario()
	s.Stats.Vitality = 100_000
	single := Simulate(s, SimOptions{DurationSec: 60})
	s.Pack = 4
	wave := Simulate(s, SimOptions{DurationSec: 60})
	if wave.DPS <= single.DPS*1.1 {
		t.Fatalf("slam into a wave of 4 gave %.1f DPS against %.1f into one enemy", wave.DPS, single.DPS)
	}
}

func TestNextTargetIsTheMostWounded(t *testing.T) {
	s := knightScenario()
	s.Pack = 3
	r := newTestRun(s, idleEngine{})
	r.others[0].hp, r.others[1].hp = 40, 15
	r.nextTarget()
	if r.state.TargetHP != 15 {
		t.Fatalf("player turned to an enemy with %d HP, want the one with 15", r.state.TargetHP)
	}
	if r.others[1].hp != r.others[1].maxHP || r.others[0].hp != 40 {
		t.Fatalf("wave after retargeting: %+v", r.others)
	}
}

// ─── Ramp ────────────────────────────────────────────────────────────────────

func TestRampRaisesTheFloorOverTime(t *testing.T) {
	s := knightScenario()
	r := newTestRun(s, idleEngine{})
	r.time = 95
	if got := r.floorNow(); got != s.Floor {
		t.Fatalf("without a ramp the floor moved to %d", got)
	}
	r.s.RampSec = 30
	if got := r.floorNow(); got != s.Floor+3 {
		t.Fatalf("95s into a 30s ramp the floor is %d, want %d", got, s.Floor+3)
	}
	if got, want := r.freshEnemy().hp, enemyHP(s.Floor+3); got != want {
		t.Fatalf("a ramped enemy has %d HP, want %d", got, want)
	}
}

func TestRampEventuallyKillsASustainedBuild(t *testing.T) {
	s := knightScenario()
	s.Artifacts = []string{"ironbreaker_gauntlets", "soul_harvest"}
	s.Stats = BaseStats{Strength: 100, Vitality: 50}
	s.Iterations = 5
	if res := Simulate(s, SimOptions{DurationSec: 600}); res.SurvivalRate != 1 {
		t.Fatalf("test build does not sustain a fixed floor (survival %.2f)", res.SurvivalRate)
	}
	s.RampSec = 5
	res := Simulate(s, SimOptions{DurationSec: 600})
	if res.SurvivalRate != 0 || res.AvgTimeAliveSec >= 600 {
		t.Fatalf("a floor every 5s for 10 minutes did not kill the build: %+v", res)
	}
}

func TestValidateRejectsBadPackAndRamp(t *testing.T) {
	for want, mutate := range map[string]func(*Scenario){
		"pack -1":     func(s *Scenario) { s.Pack = -1 },
		"pack 21":     func(s *Scenario) { s.Pack = 21 },
		"ramp_sec -5": func(s *Scenario) { s.RampSec = -5 },
	} {
		s := knightScenario()
		mutate(&s)
		if err := s.Validate(); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("want an error containing %q, got %v", want, err)
		}
	}
}
