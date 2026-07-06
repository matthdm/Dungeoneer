package combat

// Action is an intent applied to CombatState for one tick.
// Actions come from player input, AI, or simulation scripts — making the
// engine testable and simulation-friendly without real input.
type Action struct {
	Type    ActionType
	SlotIdx int     // ActionActivateSkill: which artifact slot (0-indexed)
	TargetX float64 // ActionSelectTarget: world X of click
	TargetY float64 // ActionSelectTarget: world Y of click
}

// ActionType classifies an action.
type ActionType string

const (
	ActionSelectTarget  ActionType = "select_target"
	ActionClearTarget   ActionType = "clear_target"
	ActionTargetNearest ActionType = "target_nearest"
	ActionMoveToAttack  ActionType = "move_to_attack"
	ActionActivateSkill ActionType = "activate_skill"
)
