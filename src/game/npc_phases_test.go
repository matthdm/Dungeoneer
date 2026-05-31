package game

import "testing"

func makeRS(flags map[string]int, floor int) *RunState {
	return &RunState{
		QuestFlags:   flags,
		CurrentFloor: floor,
	}
}

func TestShouldAdvance_ConditionsMet(t *testing.T) {
	tracker := NPCPhaseTracker{
		NPCID:    "varn",
		MaxPhase: 3,
		Conditions: []PhaseCondition{
			{RequiredFlag: "varn_quest1_done", RequiredFlagValue: 1, RequiredFloor: 2},
		},
	}
	rs := makeRS(map[string]int{"varn_phase": 0, "varn_quest1_done": 1}, 2)
	if !tracker.ShouldAdvance(rs) {
		t.Fatal("expected ShouldAdvance true when all conditions met")
	}
}

func TestShouldAdvance_FlagBelowThreshold(t *testing.T) {
	tracker := NPCPhaseTracker{
		NPCID:    "varn",
		MaxPhase: 3,
		Conditions: []PhaseCondition{
			{RequiredFlag: "varn_quest1_done", RequiredFlagValue: 1, RequiredFloor: 2},
		},
	}
	rs := makeRS(map[string]int{"varn_phase": 0, "varn_quest1_done": 0}, 2)
	if tracker.ShouldAdvance(rs) {
		t.Fatal("expected ShouldAdvance false when flag below threshold")
	}
}

func TestShouldAdvance_FloorBelowThreshold(t *testing.T) {
	tracker := NPCPhaseTracker{
		NPCID:    "varn",
		MaxPhase: 3,
		Conditions: []PhaseCondition{
			{RequiredFlag: "varn_quest1_done", RequiredFlagValue: 1, RequiredFloor: 3},
		},
	}
	rs := makeRS(map[string]int{"varn_phase": 0, "varn_quest1_done": 1}, 2)
	if tracker.ShouldAdvance(rs) {
		t.Fatal("expected ShouldAdvance false when floor below threshold")
	}
}

func TestShouldAdvance_AtMaxPhase(t *testing.T) {
	tracker := NPCPhaseTracker{
		NPCID:      "varn",
		MaxPhase:   2,
		Conditions: []PhaseCondition{{}, {}},
	}
	rs := makeRS(map[string]int{"varn_phase": 2}, 5)
	if tracker.ShouldAdvance(rs) {
		t.Fatal("expected ShouldAdvance false when already at MaxPhase")
	}
}

func TestShouldAdvance_NoConditionsEntry(t *testing.T) {
	// Empty Conditions slice = dialogue-driven only, never auto-advances.
	tracker := NPCPhaseTracker{
		NPCID:      "varn",
		MaxPhase:   3,
		Conditions: nil,
	}
	rs := makeRS(map[string]int{"varn_phase": 0}, 5)
	if tracker.ShouldAdvance(rs) {
		t.Fatal("expected ShouldAdvance false when no conditions defined")
	}
}

func TestAdvance_IncrementsPhase(t *testing.T) {
	tracker := NPCPhaseTracker{NPCID: "varn", MaxPhase: 3}
	rs := makeRS(map[string]int{"varn_phase": 0}, 2)
	tracker.Advance(rs)
	if rs.QuestFlags["varn_phase"] != 1 {
		t.Fatalf("expected phase 1 after Advance, got %d", rs.QuestFlags["varn_phase"])
	}
}

func TestAdvance_DoesNotExceedMaxPhase(t *testing.T) {
	tracker := NPCPhaseTracker{NPCID: "varn", MaxPhase: 2}
	rs := makeRS(map[string]int{"varn_phase": 2}, 5)
	tracker.Advance(rs)
	if rs.QuestFlags["varn_phase"] != 2 {
		t.Fatalf("expected phase to stay at MaxPhase, got %d", rs.QuestFlags["varn_phase"])
	}
}

func TestClampTrust(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, 0},
		{50, 50},
		{100, 100},
		{150, 100},
		{-10, 0},
	}
	for _, c := range cases {
		got := clampTrust(c.in)
		if got != c.want {
			t.Errorf("clampTrust(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
