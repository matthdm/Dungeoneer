package combat

import "testing"

// activate returns an ActionActivateSkill for slot 0.
func activate() []Action {
	return []Action{{Type: ActionActivateSkill, SlotIdx: 0}}
}

// ─── Mana enforcement ────────────────────────────────────────────────────────

// TestManaDeductedOnSpellCast verifies a spell costs its ManaCost when the
// state models mana, and that EventManaChanged reports the spend.
func TestManaDeductedOnSpellCast(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerMana = 50
	state.PlayerMaxMana = 50
	state.EquippedArtifacts[0] = "fireball" // ManaCost 8

	state, events := eng.Tick(state, activate())

	if state.PlayerMana != 42 {
		t.Fatalf("expected 42 mana after fireball (50-8), got %d", state.PlayerMana)
	}
	spent := 0
	for _, ev := range events {
		if ev.Type == EventManaChanged {
			spent = ev.Value
		}
	}
	if spent != 8 {
		t.Fatalf("expected EventManaChanged with Value 8, got %d", spent)
	}
}

// TestInsufficientManaBlocksSkill verifies the skill neither fires nor goes on
// cooldown when the player can't afford it.
func TestInsufficientManaBlocksSkill(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerMana = 3
	state.PlayerMaxMana = 50
	state.EquippedArtifacts[0] = "fireball" // ManaCost 8

	state, events := eng.Tick(state, activate())

	for _, ev := range events {
		if ev.Type == EventSkillFired {
			t.Fatal("skill fired despite insufficient mana")
		}
	}
	if state.ArtifactCooldowns[0] != 0 {
		t.Fatal("cooldown consumed despite skill not firing")
	}
	if state.PlayerMana != 3 {
		t.Fatalf("mana changed despite blocked cast: %d", state.PlayerMana)
	}
}

// TestManaNotEnforcedWithoutManaModel keeps benchmarker scenarios (which don't
// populate mana) behaving exactly as before mana was wired.
func TestManaNotEnforcedWithoutManaModel(t *testing.T) {
	eng := newEngine()
	state := baseState()
	// PlayerMaxMana stays 0 — sim-style state.
	state.EquippedArtifacts[0] = "fireball"

	_, events := eng.Tick(state, activate())

	fired := false
	for _, ev := range events {
		if ev.Type == EventSkillFired {
			fired = true
		}
		if ev.Type == EventManaChanged {
			t.Fatal("mana event emitted for state with no mana model")
		}
	}
	if !fired {
		t.Fatal("skill blocked despite state having no mana model")
	}
}

// TestManaCostReduction verifies ManaCostReductionPct scales the cost (100% =
// free casting, used by the InfMana dev flag).
func TestManaCostReduction(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerMana = 50
	state.PlayerMaxMana = 50
	state.ManaCostReductionPct = 100
	state.EquippedArtifacts[0] = "fireball"

	state, events := eng.Tick(state, activate())

	if state.PlayerMana != 50 {
		t.Fatalf("expected free cast at 100%% reduction, mana is %d", state.PlayerMana)
	}
	for _, ev := range events {
		if ev.Type == EventManaChanged {
			t.Fatal("EventManaChanged emitted for a free cast")
		}
	}
	fired := false
	for _, ev := range events {
		if ev.Type == EventSkillFired {
			fired = true
		}
	}
	if !fired {
		t.Fatal("skill did not fire")
	}
}

// ─── Targetless activation ───────────────────────────────────────────────────

// TestTargetlessSpellBurnsCooldownWithoutDamage verifies casting into empty
// ground consumes cooldown and mana and fires the visual event, but deals no
// damage and cannot ghost-kill a stale target.
func TestTargetlessSpellBurnsCooldownWithoutDamage(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.HasTarget = false
	state.TargetInRange = false
	state.IsAutoAttacking = false
	state.TargetHP = 0 // stale
	state.PlayerMana = 50
	state.PlayerMaxMana = 50
	state.EquippedArtifacts[0] = "fireball"

	state, events := eng.Tick(state, activate())

	fired := false
	for _, ev := range events {
		switch ev.Type {
		case EventSkillFired:
			fired = true
		case EventDamageDealt:
			t.Fatal("damage dealt with no target")
		case EventTargetDied:
			t.Fatal("ghost kill recorded with no target")
		}
	}
	if !fired {
		t.Fatal("targetless cast did not fire (visuals need the event)")
	}
	if state.ArtifactCooldowns[0] == 0 {
		t.Fatal("cooldown not consumed on targetless cast")
	}
	if state.PlayerMana != 42 {
		t.Fatalf("mana not spent on targetless cast: %d", state.PlayerMana)
	}
}

// TestMeleeSkillRequiresTargetInRange verifies weapon-multiplier skills
// (ironbreaker_gauntlets) do not fire out of range now that the engine ticks
// every frame, while blink strikes (gap-closers) still do.
func TestMeleeSkillRequiresTargetInRange(t *testing.T) {
	eng := newEngine()

	state := baseState()
	state.TargetInRange = false
	state.EquippedArtifacts[0] = "ironbreaker_gauntlets"
	state, events := eng.Tick(state, activate())
	for _, ev := range events {
		if ev.Type == EventSkillFired {
			t.Fatal("melee skill fired out of range")
		}
	}
	if state.ArtifactCooldowns[0] != 0 {
		t.Fatal("melee skill consumed cooldown out of range")
	}

	state = baseState()
	state.TargetInRange = false
	state.EquippedArtifacts[0] = "shroud_cloak" // blink strike: usable from range
	_, events = eng.Tick(state, activate())
	fired := false
	for _, ev := range events {
		if ev.Type == EventSkillFired {
			fired = true
		}
	}
	if !fired {
		t.Fatal("blink strike blocked out of range (should gap-close)")
	}
}

// ─── PassiveArtifacts (worn equipment, no activation slot) ───────────────────

// TestPassiveArtifactBurnFromEquipment verifies ember_mantle applies its burn
// DoT when carried in PassiveArtifacts (worn item) rather than an active slot.
func TestPassiveArtifactBurnFromEquipment(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PassiveArtifacts = []string{"ember_mantle"}

	// First tick fires an auto-attack (timer 0), which should apply the burn.
	state, _ = eng.Tick(state, nil)

	if !state.BurnActive {
		t.Fatal("ember_mantle in PassiveArtifacts did not apply burn on auto-attack")
	}
	if state.BurnDPS != 5 {
		t.Fatalf("expected burn DPS 5, got %d", state.BurnDPS)
	}
}

// TestPassiveArtifactHealOnKill verifies soul_harvest heals from
// PassiveArtifacts when the target dies.
func TestPassiveArtifactHealOnKill(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.PlayerHP = 50
	state.TargetHP = 1 // next auto-attack kills
	state.PassiveArtifacts = []string{"soul_harvest"}

	state, _ = eng.Tick(state, nil)

	if !state.TargetIsDead {
		t.Fatal("target should have died")
	}
	// soul_harvest: heal 20% of 100 max HP = 20.
	if state.PlayerHP != 70 {
		t.Fatalf("expected 70 HP after soul_harvest heal (50+20), got %d", state.PlayerHP)
	}
}

// TestBurnDroppedWhenTargetGone verifies an active burn cannot transfer to a
// future target after the current one is cleared.
func TestBurnDroppedWhenTargetGone(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.BurnActive = true
	state.BurnDPS = 5
	state.BurnTimer = 3.0
	state.ActiveDoTCount = 1
	state.HasTarget = false
	state.IsAutoAttacking = false

	state, _ = eng.Tick(state, nil)

	if state.BurnActive {
		t.Fatal("burn survived target clear")
	}
	if state.ActiveDoTCount != 0 {
		t.Fatalf("ActiveDoTCount not decremented: %d", state.ActiveDoTCount)
	}
}
