package combat

import (
	"testing"
)

// baseState returns a minimal CombatState ready for testing.
// Target is in range, auto-attack is active, timer starts at zero so the
// first tick fires immediately.
func baseState() CombatState {
	return CombatState{
		PlayerHP:             100,
		PlayerMaxHP:          100,
		PlayerDamage:         10,
		PlayerAttackRange:    1.5,
		PlayerAttackInterval: 0.8,
		PlayerClass:          "knight",

		HasTarget:       true,
		TargetHP:        200,
		TargetMaxHP:     200,
		TargetInRange:   true,
		IsAutoAttacking: true,

		AutoAttackTimer: 0, // fires immediately on first tick

		DeltaTime: 1.0 / 60.0,
	}
}

func newEngine() *DefaultCombatEngine {
	return NewDefaultCombatEngine(42)
}

// ─── Auto-attack ─────────────────────────────────────────────────────────────

// TestAutoAttackFiresAfterInterval verifies that the first auto-attack fires
// within one tick when AutoAttackTimer == 0, and that subsequent attacks wait
// for the full interval.
func TestAutoAttackFiresAfterInterval(t *testing.T) {
	eng := newEngine()
	state := baseState()

	// Tick 1: timer is 0, should fire immediately.
	state, events := eng.Tick(state, nil)
	hasDmg := false
	for _, ev := range events {
		if ev.Type == EventDamageDealt {
			hasDmg = true
		}
	}
	if !hasDmg {
		t.Fatal("expected EventDamageDealt on tick 1 (timer was 0)")
	}

	// Tick 2: timer just reset to ~0.8 s; should NOT fire.
	state, events = eng.Tick(state, nil)
	for _, ev := range events {
		if ev.Type == EventDamageDealt {
			t.Fatal("EventDamageDealt fired too early on tick 2")
		}
	}

	// Advance time past the attack interval.
	state.AutoAttackTimer = 0
	state, events = eng.Tick(state, nil)
	hasDmg = false
	for _, ev := range events {
		if ev.Type == EventDamageDealt {
			hasDmg = true
		}
	}
	if !hasDmg {
		t.Fatal("expected EventDamageDealt after forcing timer to 0 again")
	}
}

// TestAutoAttackNotFiredWhenNotInRange ensures no attack happens when the
// target is out of range.
func TestAutoAttackNotFiredWhenNotInRange(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.TargetInRange = false

	_, events := eng.Tick(state, nil)
	for _, ev := range events {
		if ev.Type == EventDamageDealt {
			t.Fatal("EventDamageDealt fired while target is out of range")
		}
	}
}

// ─── Kill streak ─────────────────────────────────────────────────────────────

// TestKillStreakIncrementsOnTargetDied checks that KillStreak goes up by 1
// each time EventTargetDied is emitted.
func TestKillStreakIncrementsOnTargetDied(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.TargetHP = 1 // dies on first hit

	state, events := eng.Tick(state, nil)
	hasDied := false
	for _, ev := range events {
		if ev.Type == EventTargetDied {
			hasDied = true
		}
	}
	if !hasDied {
		t.Fatal("expected EventTargetDied when target HP drops to 0")
	}
	if state.KillStreak != 1 {
		t.Fatalf("expected KillStreak=1 after first kill, got %d", state.KillStreak)
	}

	// Simulate a second kill without the streak timer expiring.
	state.HasTarget = true
	state.TargetHP = 1
	state.TargetIsDead = false
	state.TargetInRange = true
	state.IsAutoAttacking = true
	state.AutoAttackTimer = 0

	state, events = eng.Tick(state, nil)
	hasDied = false
	for _, ev := range events {
		if ev.Type == EventTargetDied {
			hasDied = true
		}
	}
	if !hasDied {
		t.Fatal("expected EventTargetDied on second kill")
	}
	if state.KillStreak != 2 {
		t.Fatalf("expected KillStreak=2 after second kill, got %d", state.KillStreak)
	}
}

// TestKillStreakResetsAfterTimer verifies that once StreakTimer counts down to
// zero the KillStreak is reset to 0 and EventStreakChange(0) is emitted.
func TestKillStreakResetsAfterTimer(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.TargetHP = 1
	state.AutoAttackTimer = 0

	// First kill: establishes a streak.
	state, _ = eng.Tick(state, nil)
	if state.KillStreak != 1 {
		t.Fatalf("precondition: expected KillStreak=1, got %d", state.KillStreak)
	}

	// Fast-forward the streak timer past expiry.
	state.StreakTimer = 0.001
	// One more tick with dt > remaining timer.
	state.DeltaTime = 0.1
	state.HasTarget = false // no new target, so no new kills
	state, events := eng.Tick(state, nil)

	if state.KillStreak != 0 {
		t.Fatalf("expected KillStreak=0 after StreakTimer expiry, got %d", state.KillStreak)
	}
	hasReset := false
	for _, ev := range events {
		if ev.Type == EventStreakChange && ev.Value == 0 {
			hasReset = true
		}
	}
	if !hasReset {
		t.Fatal("expected EventStreakChange(0) when streak timer expires")
	}
}

// ─── Kill momentum ────────────────────────────────────────────────────────────

// TestKillMomentumReducesInterval checks that a positive KillStreak causes the
// effective auto-attack interval to shrink (more attacks land in the same
// simulated time).
func TestKillMomentumReducesInterval(t *testing.T) {
	// With a 5-kill streak the reduction should be 25% (capped).
	eng := newEngine()
	state := baseState()
	state.KillStreak = 5
	state.StreakTimer = 10.0
	state.TargetHP = 9999 // target won't die; we're timing attack frequency
	state.AutoAttackTimer = 0

	// Count attacks in 10 ticks (no-streak engine for comparison).
	engBase := newEngine()
	stateBase := baseState()
	stateBase.TargetHP = 9999
	stateBase.AutoAttackTimer = 0

	const ticks = 300 // 5 seconds of sim
	dt := 1.0 / 60.0

	streakDmg := 0
	baseDmg := 0
	for i := 0; i < ticks; i++ {
		state.DeltaTime = dt
		state, _ = eng.Tick(state, nil)
		// Refresh streak so it doesn't time out.
		state.KillStreak = 5
		state.StreakTimer = 10.0

		stateBase.DeltaTime = dt
		stateBase, _ = engBase.Tick(stateBase, nil)
	}
	// Damage proxy: read from state (direct field tracking isn't exposed, so
	// use total TargetHP reduction as a proxy).
	streakDmg = 9999 - state.TargetHP
	baseDmg = 9999 - stateBase.TargetHP

	if streakDmg <= baseDmg {
		t.Fatalf("expected streak engine to deal more damage (%d streak vs %d base)", streakDmg, baseDmg)
	}
}

// TestKillMomentumCappedAt25Pct confirms the reduction never exceeds 25% even
// at very high streaks.
func TestKillMomentumCappedAt25Pct(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.KillStreak = 100 // way over the cap
	state.StreakTimer = 999
	state.TargetHP = 9999
	state.AutoAttackTimer = 0

	engBase := newEngine()
	stateBase := baseState()
	stateBase.KillStreak = 5 // exactly at cap
	stateBase.StreakTimer = 999
	stateBase.TargetHP = 9999
	stateBase.AutoAttackTimer = 0

	const ticks = 300
	dt := 1.0 / 60.0
	for i := 0; i < ticks; i++ {
		state.DeltaTime = dt
		state, _ = eng.Tick(state, nil)
		state.KillStreak = 100
		state.StreakTimer = 999

		stateBase.DeltaTime = dt
		stateBase, _ = engBase.Tick(stateBase, nil)
		stateBase.KillStreak = 5
		stateBase.StreakTimer = 999
	}

	dmgHigh := 9999 - state.TargetHP
	dmgCap := 9999 - stateBase.TargetHP

	// At 100 kills (500%) and at 5 kills (25%) the effective reduction should
	// be equal because both hit the cap.
	diff := dmgHigh - dmgCap
	if diff < 0 {
		diff = -diff
	}
	// Allow a ±1 damage tolerance from integer rounding.
	if diff > 1 {
		t.Fatalf("damage at streak=100 (%d) should equal streak=5 (%d) (both capped at 25%%)", dmgHigh, dmgCap)
	}
}

// ─── Crit ────────────────────────────────────────────────────────────────────

// TestCritEventsHaveCorrectMultiplier verifies that when a crit event is
// emitted its Value is 1.5× the base damage (within integer rounding).
func TestCritEventsHaveCorrectMultiplier(t *testing.T) {
	// Force the engine to always crit by seeding an RNG that happens to roll
	// under 0.05 on the first call. We instead override calcDamage indirectly
	// by running many ticks until we collect at least one crit event and then
	// confirm the value is >= baseDamage * 1.5.
	baseDmg := 20
	eng := NewDefaultCombatEngine(7) // arbitrary seed; crits are probabilistic
	state := baseState()
	state.PlayerDamage = baseDmg
	state.TargetHP = 99999

	got := false
	for i := 0; i < 10000 && !got; i++ {
		state.AutoAttackTimer = 0
		state.DeltaTime = 1.0 / 60.0
		_, events := eng.Tick(state, nil)
		for _, ev := range events {
			if ev.Type == EventDamageDealt && ev.IsCrit {
				expected := int(float64(baseDmg) * critMultiplier)
				if ev.Value != expected {
					t.Fatalf("crit damage: got %d, want %d (base %d × %.1f)", ev.Value, expected, baseDmg, critMultiplier)
				}
				got = true
			}
		}
	}
	if !got {
		t.Fatal("no crit event produced in 10 000 ticks (crit chance 5% — should be near-certain)")
	}
}

// ─── Skill activation ────────────────────────────────────────────────────────

// TestSkillActivationRespectsCooldown checks that a skill cannot be fired
// while its cooldown is still counting down.
func TestSkillActivationRespectsCooldown(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.EquippedArtifacts[0] = "ironbreaker_gauntlets"
	state.ArtifactCooldowns[0] = 0

	action := Action{Type: ActionActivateSkill, SlotIdx: 0}

	// First activation: should fire.
	_, events := eng.Tick(state, []Action{action})
	fired := false
	for _, ev := range events {
		if ev.Type == EventSkillFired && ev.Tag == "ironbreaker_gauntlets" {
			fired = true
		}
	}
	if !fired {
		t.Fatal("expected EventSkillFired on first skill activation")
	}

	// Re-read state from the engine output so cooldown is set.
	state, _ = eng.Tick(state, []Action{action}) // this tick sets the cooldown
	state.ArtifactCooldowns[0] = 2.9             // still on cooldown

	// Second activation while on cooldown: should NOT fire.
	_, events = eng.Tick(state, []Action{action})
	for _, ev := range events {
		if ev.Type == EventSkillFired {
			t.Fatal("EventSkillFired emitted while skill was on cooldown")
		}
	}
}

// TestSkillActivationFiresAfterCooldownExpires verifies that once the cooldown
// reaches zero the skill can be activated again.
func TestSkillActivationFiresAfterCooldownExpires(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.EquippedArtifacts[0] = "ironbreaker_gauntlets"
	state.ArtifactCooldowns[0] = 0.001 // almost expired

	// One tick that expires the cooldown.
	state.DeltaTime = 0.1
	state, _ = eng.Tick(state, nil)
	if state.ArtifactCooldowns[0] != 0 {
		t.Fatalf("expected cooldown to reach 0, got %f", state.ArtifactCooldowns[0])
	}

	// Now activate: should succeed.
	action := Action{Type: ActionActivateSkill, SlotIdx: 0}
	_, events := eng.Tick(state, []Action{action})
	fired := false
	for _, ev := range events {
		if ev.Type == EventSkillFired {
			fired = true
		}
	}
	if !fired {
		t.Fatal("expected EventSkillFired after cooldown expired")
	}
}

// ─── Simulation ──────────────────────────────────────────────────────────────

// TestRunSimulationProducesNonZeroKillsPerMinute is the smoke test: a
// well-configured scenario should produce a positive KPM.
// Stats are tuned so the player can kill the floor-1 enemy before dying:
//   - Vitality 40 → PlayerHP = 100 + 40*5 = 300
//   - Enemy HP on floor 1 = 50 + 25 = 75
//   - Enemy dmg/tick on floor 1 = 5 + 2 = 7 → 300/7 ≈ 43 ticks to die
//   - Player base dmg = 10 + Strength 50 = 60 → kills in ceil(75/60)=2 hits
//   - First hit fires on tick 1 (AutoAttackTimer=0); second after 48 ticks
//   → player kills in ~49 ticks, well under the 43-tick death ceiling.
//   Wait — 43 < 49.  Bump Strength higher so first hit one-shots (dmg ≥ 75):
//   Strength = 65 → PlayerDamage = 75.  First tick: 75 dmg → target dies.
func TestRunSimulationProducesNonZeroKillsPerMinute(t *testing.T) {
	scenario := Scenario{
		Name:      "smoke-test",
		Class:     "knight",
		Artifacts: []string{"ironbreaker_gauntlets"},
		Stats:     BaseStats{Strength: 65, Vitality: 40},
		Floor:     1,
		Biome:     "crypt",
		EnemyPool: []string{"melee"},
		Iterations: 50,
		SkillRotation: []string{"slot_1", "auto", "auto"},
	}

	eng := NewDefaultCombatEngine(99)
	result := RunSimulation(eng, scenario)

	if result.KillsPerMinute <= 0 {
		t.Fatalf("expected KillsPerMinute > 0, got %f (AvgClearTimeSec=%f)",
			result.KillsPerMinute, result.AvgClearTimeSec)
	}
	if result.Iterations != 50 {
		t.Fatalf("expected Iterations=50, got %d", result.Iterations)
	}
	if result.SurvivalRate < 0 || result.SurvivalRate > 1 {
		t.Fatalf("SurvivalRate out of range: %f", result.SurvivalRate)
	}
}
