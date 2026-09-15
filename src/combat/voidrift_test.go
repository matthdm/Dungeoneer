package combat

import "testing"

// TestVoidRiftCatalystRootsAndDamagesTarget verifies void_rift_catalyst fires,
// deals DamageMultiplier-scaled damage, and applies a root — mirrors the
// existing ironbreaker_gauntlets/ashbound_chain engine test coverage.
func TestVoidRiftCatalystRootsAndDamagesTarget(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.EquippedArtifacts[0] = "void_rift_catalyst"
	state.ArtifactCooldowns[0] = 0

	action := Action{Type: ActionActivateSkill, SlotIdx: 0}
	newState, events := eng.Tick(state, []Action{action})

	fired := false
	dealt := false
	for _, ev := range events {
		if ev.Type == EventSkillFired && ev.Tag == "void_rift_catalyst" {
			fired = true
		}
		if ev.Type == EventDamageDealt && ev.Tag == "void_rift_catalyst" && ev.Value > 0 {
			dealt = true
		}
	}
	if !fired {
		t.Fatal("expected EventSkillFired for void_rift_catalyst")
	}
	if !dealt {
		t.Fatal("expected EventDamageDealt with positive value for void_rift_catalyst")
	}
	if !newState.TargetRooted {
		t.Fatal("expected TargetRooted=true after void_rift_catalyst fires")
	}
	if newState.RootTimer <= 0 || newState.RootTimer > 1.5 {
		t.Fatalf("expected RootTimer in (0, 1.5], got %f", newState.RootTimer)
	}
}

// TestVoidRiftCatalystRequiresTargetInRange checks the melee-reach gate
// (shared by every DamageMultiplier>0 skill, e.g. ironbreaker_gauntlets)
// also applies to void_rift_catalyst: out of range, no fire.
func TestVoidRiftCatalystRequiresTargetInRange(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.TargetInRange = false
	state.EquippedArtifacts[0] = "void_rift_catalyst"
	state.ArtifactCooldowns[0] = 0

	action := Action{Type: ActionActivateSkill, SlotIdx: 0}
	_, events := eng.Tick(state, []Action{action})

	for _, ev := range events {
		if ev.Type == EventSkillFired {
			t.Fatal("expected void_rift_catalyst to NOT fire when target is out of range")
		}
	}
}

// TestVoidboundPendantIsRegisteredAsPassive guards against registration
// typos: voidbound_pendant must be a zero-cooldown passive with its
// skill-duration bonus set, so aggregatePassive (new_adapter.go) picks it up.
func TestVoidboundPendantIsRegisteredAsPassive(t *testing.T) {
	eff, ok := ArtifactEffects["voidbound_pendant"]
	if !ok {
		t.Fatal("voidbound_pendant not registered in ArtifactEffects")
	}
	if !eff.IsPassive {
		t.Fatal("expected voidbound_pendant.IsPassive = true")
	}
	if eff.Cooldown != 0 {
		t.Fatalf("expected voidbound_pendant.Cooldown = 0, got %f", eff.Cooldown)
	}
	if eff.SkillDurationPct <= 0 {
		t.Fatalf("expected voidbound_pendant.SkillDurationPct > 0, got %d", eff.SkillDurationPct)
	}
}
