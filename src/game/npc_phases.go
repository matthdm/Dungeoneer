package game

// PhaseCondition defines criteria for an NPC to auto-advance to the next phase at floor entry.
// Both non-zero criteria must be satisfied. Zero values mean "no requirement".
type PhaseCondition struct {
	RequiredFlag      string // QuestFlag that must be >= RequiredFlagValue (empty = ignored)
	RequiredFlagValue int    // minimum value for RequiredFlag
	RequiredFloor     int    // minimum floor number (0 = no floor requirement)
}

// NPCPhaseTracker evaluates whether a major NPC should auto-advance to the next phase
// at floor entry, based on quest flag and floor conditions. Phases can also advance
// through the advance_phase dialogue action — this tracker is an automatic fallback.
type NPCPhaseTracker struct {
	NPCID      string
	MaxPhase   int
	Conditions []PhaseCondition // Conditions[i] = condition to advance from phase i → i+1
}

// ShouldAdvance returns true if all conditions for the NPC's current phase are met.
func (t *NPCPhaseTracker) ShouldAdvance(rs *RunState) bool {
	if rs == nil {
		return false
	}
	phase := rs.QuestFlags[t.NPCID+"_phase"]
	if phase >= t.MaxPhase || phase >= len(t.Conditions) {
		return false
	}
	c := t.Conditions[phase]
	if c.RequiredFlag != "" && rs.QuestFlags[c.RequiredFlag] < c.RequiredFlagValue {
		return false
	}
	if c.RequiredFloor > 0 && rs.CurrentFloor < c.RequiredFloor {
		return false
	}
	return true
}

// Advance increments the NPC's current phase in RunState quest flags.
func (t *NPCPhaseTracker) Advance(rs *RunState) {
	if rs == nil {
		return
	}
	phase := rs.QuestFlags[t.NPCID+"_phase"]
	if phase < t.MaxPhase {
		rs.QuestFlags[t.NPCID+"_phase"] = phase + 1
	}
}

// clampTrust restricts a trust value to [0, 100].
func clampTrust(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
