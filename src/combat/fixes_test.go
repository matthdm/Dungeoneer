package combat

import (
	"math"
	"testing"
)

func countEvents(events []Event, typ EventType) int {
	n := 0
	for _, ev := range events {
		if ev.Type == typ {
			n++
		}
	}
	return n
}

func failReason(events []Event) string {
	for _, ev := range events {
		if ev.Type == EventSkillFailed {
			return ev.Reason
		}
	}
	return ""
}

func slot0(id string) []Action {
	_ = id
	return []Action{{Type: ActionActivateSkill, SlotIdx: 0}}
}

// ─── 1. Burn deals its damage ────────────────────────────────────────────────

// A 5 DPS burn lasting 3 seconds must take 15 HP off the target at 60 Hz.
func TestBurnDealsFullDamageAt60Hz(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerDamage = 0 // isolate the burn
	state.PlayerAttackInterval = 1000
	state.TargetHP, state.TargetMaxHP = 10000, 10000
	state.PassiveArtifacts = []string{"ember_mantle"}

	burnDmg := 0
	for i := 0; i < 60*5; i++ {
		var events []Event
		state, events = eng.Tick(state, nil)
		for _, ev := range events {
			if ev.Type == EventDamageDealt && ev.Tag == "ember_mantle" {
				burnDmg += ev.Value
			}
		}
	}
	if burnDmg != 15 {
		t.Fatalf("burn dealt %d over its 3s duration, want 15", burnDmg)
	}
	if got := 10000 - state.TargetHP; got != 15 {
		t.Fatalf("target lost %d HP, want 15", got)
	}
	if state.BurnActive {
		t.Fatal("burn still active after its duration")
	}
}

// Burn damage arrives in a few readable ticks, not one event per frame.
func TestBurnTicksAreChunked(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerDamage = 0
	state.PlayerAttackInterval = 1000
	state.TargetHP, state.TargetMaxHP = 10000, 10000
	state.PassiveArtifacts = []string{"ember_mantle"}
	n := 0
	for i := 0; i < 60*5; i++ {
		var events []Event
		state, events = eng.Tick(state, nil)
		for _, ev := range events {
			if ev.Type == EventDamageDealt && ev.Tag == "ember_mantle" {
				n++
			}
		}
	}
	if n < 3 || n > 8 {
		t.Fatalf("burn emitted %d damage events over 3s, want a handful (about 2 per second)", n)
	}
}

// ─── 2. Range and line of sight ──────────────────────────────────────────────

func TestSpellBlockedOutOfRange(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerIntelligence = 10
	state.IsAutoAttacking = false // isolate the skill
	state.EquippedArtifacts[0] = "fireball"
	state.TargetDist = ArtifactEffects["fireball"].CastRange + 0.5

	out, events := eng.Tick(state, slot0("fireball"))
	if countEvents(events, EventSkillFired) != 0 {
		t.Fatal("fireball fired at a target beyond its cast range")
	}
	if failReason(events) != FailOutOfRange {
		t.Fatalf("fail reason = %q, want %q", failReason(events), FailOutOfRange)
	}
	if out.ArtifactCooldowns[0] != 0 || out.TargetHP != state.TargetHP {
		t.Fatal("blocked cast changed cooldown or target HP")
	}

	state.TargetDist = ArtifactEffects["fireball"].CastRange - 0.5
	out, events = eng.Tick(state, slot0("fireball"))
	if countEvents(events, EventSkillFired) != 1 || out.TargetHP >= state.TargetHP {
		t.Fatal("fireball did not fire at a target inside its cast range")
	}
}

func TestSkillsBlockedWithoutLineOfSight(t *testing.T) {
	for _, id := range []string{"fireball", "lightning", "chaos_ray", "shroud_cloak", "blood_price", "ashbound_chain", "ironbreaker_gauntlets"} {
		eng := newEngine()
		state := baseState()
		state.PlayerIntelligence = 10
		state.EquippedArtifacts[0] = id
		state.TargetLOSBlocked = true
		out, events := eng.Tick(state, slot0(id))
		// Auto-attack still runs in baseState; only look at the skill.
		if countEvents(events, EventSkillFired) != 0 {
			t.Errorf("%s fired through a wall", id)
		}
		if failReason(events) != FailNoLineOfSight {
			t.Errorf("%s: fail reason = %q, want %q", id, failReason(events), FailNoLineOfSight)
		}
		if out.ArtifactCooldowns[0] != 0 {
			t.Errorf("%s: blocked cast started its cooldown", id)
		}
		if out.PlayerHP != state.PlayerHP {
			t.Errorf("%s: blocked cast cost HP", id)
		}
	}
}

// The blink strike is a gap-closer: it works beyond melee reach, but not from
// across the map.
func TestBlinkStrikeHasItsOwnRange(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.TargetInRange = false // outside melee reach
	state.EquippedArtifacts[0] = "shroud_cloak"

	state.TargetDist = 3
	_, events := eng.Tick(state, slot0("shroud_cloak"))
	if countEvents(events, EventSkillFired) != 1 {
		t.Fatal("blink strike blocked at 3 tiles")
	}

	state.TargetDist = 30
	_, events = eng.Tick(state, slot0("shroud_cloak"))
	if countEvents(events, EventSkillFired) != 0 {
		t.Fatal("blink strike fired at 30 tiles")
	}
}

// ─── 3. Wasted casts cost nothing ────────────────────────────────────────────

func TestExecuteAboveThresholdDoesNotFire(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.EquippedArtifacts[0] = "grave_reaper"
	state.IsAutoAttacking = false

	out, events := eng.Tick(state, slot0("grave_reaper")) // target at full HP
	if countEvents(events, EventSkillFired) != 0 || out.ArtifactCooldowns[0] != 0 {
		t.Fatal("execute on a healthy target fired or started its cooldown")
	}
	if failReason(events) != FailTargetNotLow {
		t.Fatalf("fail reason = %q, want %q", failReason(events), FailTargetNotLow)
	}

	state.TargetHP = state.TargetMaxHP * 15 / 100
	out, events = eng.Tick(state, slot0("grave_reaper"))
	if countEvents(events, EventTargetDied) != 1 || out.ArtifactCooldowns[0] == 0 {
		t.Fatal("execute below threshold did not kill and go on cooldown")
	}
}

func TestArcaneSurgeWithNothingOnCooldownDoesNotFire(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.IsAutoAttacking = false
	state.EquippedArtifacts[0] = "arcane_surge"
	state.EquippedArtifacts[1] = "ironbreaker_gauntlets"

	out, events := eng.Tick(state, slot0("arcane_surge"))
	if countEvents(events, EventSkillFired) != 0 || out.ArtifactCooldowns[0] != 0 {
		t.Fatal("arcane_surge fired for zero damage")
	}
	if failReason(events) != FailNothingOnCooldown {
		t.Fatalf("fail reason = %q, want %q", failReason(events), FailNothingOnCooldown)
	}

	state.ArtifactCooldowns[1] = 2
	out, events = eng.Tick(state, slot0("arcane_surge"))
	if countEvents(events, EventSkillFired) != 1 || out.TargetHP >= state.TargetHP {
		t.Fatal("arcane_surge did not fire with another artifact on cooldown")
	}
}

// A damaging spell with no target is refused and costs neither mana nor
// cooldown.
func TestTargetlessSpellCostsNothing(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.HasTarget = false
	state.TargetInRange = false
	state.IsAutoAttacking = false
	state.TargetHP = 0 // stale
	state.PlayerMana, state.PlayerMaxMana = 50, 50
	state.EquippedArtifacts[0] = "fireball"

	out, events := eng.Tick(state, slot0("fireball"))
	if countEvents(events, EventSkillFired) != 0 || countEvents(events, EventDamageDealt) != 0 || countEvents(events, EventTargetDied) != 0 {
		t.Fatal("targetless fireball fired, dealt damage or ghost-killed")
	}
	if failReason(events) != FailNoTarget {
		t.Fatalf("fail reason = %q, want %q", failReason(events), FailNoTarget)
	}
	if out.ArtifactCooldowns[0] != 0 || out.PlayerMana != 50 {
		t.Fatalf("targetless cast cost cooldown %.2f / mana %d", out.ArtifactCooldowns[0], out.PlayerMana)
	}
}

// Blood Price must not take HP when there is nothing to hit.
func TestBloodPriceWithoutTargetCostsNoHP(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.HasTarget = false
	state.IsAutoAttacking = false
	state.EquippedArtifacts[0] = "blood_price"
	out, events := eng.Tick(state, slot0("blood_price"))
	if out.PlayerHP != state.PlayerHP || countEvents(events, EventHPSpent) != 0 {
		t.Fatal("blood_price spent HP with no target")
	}
}

// Ground and self casts still work with no target: the healing canopy is
// placed by the cursor and the taunt is a self buff.
func TestGroundAndSelfCastsNeedNoTarget(t *testing.T) {
	for _, id := range []string{"fractal_canopy", "wardens_medallion"} {
		eng := newEngine()
		state := baseState()
		state.HasTarget = false
		state.IsAutoAttacking = false
		state.EquippedArtifacts[0] = id
		out, events := eng.Tick(state, slot0(id))
		if countEvents(events, EventSkillFired) != 1 || out.ArtifactCooldowns[0] == 0 {
			t.Errorf("%s did not fire without a target", id)
		}
		if countEvents(events, EventDamageDealt) != 0 {
			t.Errorf("%s dealt damage with no target", id)
		}
	}
}

// ─── 4. One damage rule for every skill ──────────────────────────────────────

// Arcane Surge and Blood Price get the same passive damage bonus as spells.
func TestSurgeAndBloodPriceGetPassiveDamageBonus(t *testing.T) {
	run := func(id string, withRing bool) int {
		// Crits are random; take the most common (non-crit) value over many seeds.
		counts := map[int]int{}
		for seed := uint64(1); seed <= 60; seed++ {
			eng := NewDefaultCombatEngine(seed)
			state := baseState()
			state.IsAutoAttacking = false
			state.TargetHP, state.TargetMaxHP = 100000, 100000
			state.EquippedArtifacts[0] = id
			state.EquippedArtifacts[1] = "ironbreaker_gauntlets"
			state.ArtifactCooldowns[1] = 2
			if withRing {
				state.PassiveArtifacts = []string{"marrow_ring"} // +12% damage
			}
			_, events := eng.Tick(state, slot0(id))
			for _, ev := range events {
				if ev.Type == EventDamageDealt {
					counts[ev.Value]++
				}
			}
		}
		best, bestN := 0, 0
		for v, n := range counts {
			if n > bestN {
				best, bestN = v, n
			}
		}
		return best
	}
	for _, id := range []string{"arcane_surge", "blood_price"} {
		base, boosted := run(id, false), run(id, true)
		want := int(float64(base) * 1.12)
		if boosted != want {
			t.Errorf("%s: %d without marrow_ring, %d with; want %d", id, base, boosted, want)
		}
	}
}

// They can also crit, like every other damaging skill.
func TestSurgeAndBloodPriceCanCrit(t *testing.T) {
	for _, id := range []string{"arcane_surge", "blood_price"} {
		crits := 0
		for seed := uint64(1); seed <= 400; seed++ {
			eng := NewDefaultCombatEngine(seed)
			state := baseState()
			state.IsAutoAttacking = false
			state.TargetHP, state.TargetMaxHP = 100000, 100000
			state.EquippedArtifacts[0] = id
			state.EquippedArtifacts[1] = "ironbreaker_gauntlets"
			state.ArtifactCooldowns[1] = 2
			_, events := eng.Tick(state, slot0(id))
			for _, ev := range events {
				if ev.Type == EventDamageDealt && ev.IsCrit {
					crits++
				}
			}
		}
		if crits == 0 {
			t.Errorf("%s never crit in 400 casts", id)
		}
	}
}

// ─── 5. Cooldown reduction means what it says ────────────────────────────────

func TestCooldownReductionShortensCooldownByItsPercentage(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.IsAutoAttacking = false
	state.EquippedArtifacts[0] = "wardens_medallion" // 10s base
	state.CooldownReductionPct = 28

	state, _ = eng.Tick(state, slot0("wardens_medallion"))
	ticks := 1
	for state.ArtifactCooldowns[0] > 0 {
		state, _ = eng.Tick(state, nil)
		ticks++
	}
	got := float64(ticks) / 60.0
	if math.Abs(got-7.2) > 0.05 {
		t.Fatalf("10s cooldown with 28%% CDR took %.2fs, want 7.20s", got)
	}
}

func TestCooldownReductionIsCapped(t *testing.T) {
	if got := EffectiveCooldown(10, 500); math.Abs(got-10*(1-MaxCooldownReductionPct/100.0)) > 1e-9 {
		t.Fatalf("EffectiveCooldown(10, 500) = %g, want the capped value", got)
	}
	if got := EffectiveCooldown(10, 0); got != 10 {
		t.Fatalf("EffectiveCooldown(10, 0) = %g, want 10", got)
	}
}

// ─── 6. Resonance Crystal stacks ─────────────────────────────────────────────

// Burn plus two lingering fields is three active DoTs: +24% with one crystal.
func TestResonanceCrystalStacksPerActiveDoT(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerDamage = 0
	state.PlayerIntelligence = 10
	state.TargetHP, state.TargetMaxHP = 1000000, 1000000
	state.EquippedArtifacts[0] = "lightning_storm"
	state.EquippedArtifacts[1] = "fractal_canopy"
	state.PassiveArtifacts = []string{"ember_mantle", "resonance_crystal"}

	if got := passiveDmgBonusMult(&state); got != 1.0 {
		t.Fatalf("bonus with no DoTs = %g, want 1", got)
	}
	state, _ = eng.Tick(state, nil) // auto-attack applies the burn
	if state.ActiveDoTCount != 1 {
		t.Fatalf("after burn: ActiveDoTCount = %d, want 1", state.ActiveDoTCount)
	}
	state, _ = eng.Tick(state, []Action{{Type: ActionActivateSkill, SlotIdx: 0}})
	if state.ActiveDoTCount != 2 {
		t.Fatalf("after storm: ActiveDoTCount = %d, want 2", state.ActiveDoTCount)
	}
	state, _ = eng.Tick(state, []Action{{Type: ActionActivateSkill, SlotIdx: 1}})
	if state.ActiveDoTCount != 3 {
		t.Fatalf("after canopy: ActiveDoTCount = %d, want 3", state.ActiveDoTCount)
	}
	if got := passiveDmgBonusMult(&state); math.Abs(got-1.24) > 1e-9 {
		t.Fatalf("bonus with 3 DoTs = %g, want 1.24", got)
	}

	// Recasting the same field refreshes it; it does not add a stack.
	state.ArtifactCooldowns[0] = 0
	state, _ = eng.Tick(state, []Action{{Type: ActionActivateSkill, SlotIdx: 0}})
	if state.ActiveDoTCount != 3 {
		t.Fatalf("recast changed ActiveDoTCount to %d, want 3", state.ActiveDoTCount)
	}

	// Stacks fall off as each effect expires.
	for i := 0; i < 60*6; i++ {
		state.PlayerAttackInterval = 1000 // no fresh burn
		state, _ = eng.Tick(state, nil)
	}
	if state.ActiveDoTCount != 0 {
		t.Fatalf("after everything expired: ActiveDoTCount = %d, want 0", state.ActiveDoTCount)
	}
}

// ─── 7. Taunt ────────────────────────────────────────────────────────────────

// The engine reports the taunt so the game layer can force nearby enemies
// onto the player; it carries the radius and the (duration-scaled) time.
func TestTauntEmitsTauntEvent(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.IsAutoAttacking = false
	state.SkillDurationPct = 50
	state.EquippedArtifacts[0] = "wardens_medallion"

	out, events := eng.Tick(state, slot0("wardens_medallion"))
	var taunt *Event
	for i := range events {
		if events[i].Type == EventTaunt {
			taunt = &events[i]
		}
	}
	if taunt == nil {
		t.Fatal("no EventTaunt emitted")
	}
	eff := ArtifactEffects["wardens_medallion"]
	if taunt.Radius != eff.TauntRadius || eff.TauntRadius <= 0 {
		t.Fatalf("taunt radius = %g, registry = %g", taunt.Radius, eff.TauntRadius)
	}
	if math.Abs(taunt.Duration-eff.DurationSec*1.5) > 1e-9 {
		t.Fatalf("taunt duration = %g, want %g", taunt.Duration, eff.DurationSec*1.5)
	}
	if out.DamageReductionPct != eff.DamageReductionPct {
		t.Fatal("taunt no longer grants its damage reduction")
	}
}
