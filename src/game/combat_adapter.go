package game

import (
	"encoding/json"
	"os"
)

// CombatAdapter abstracts combat logic so legacy and new engine implementations
// can be swapped at runtime via dev_settings.json without touching game code.
//
// LegacyCombatAdapter: wraps existing handlers.game.go (untouched).
// NewCombatAdapter:    routes to src/combat engine.
type CombatAdapter interface {
	// HandleTargetSelect fires when the player left-clicks on a world position.
	HandleTargetSelect(g *Game, worldX, worldY float64)
	// HandleTargetNearest selects the nearest visible enemy (C key).
	HandleTargetNearest(g *Game)
	// HandleMoveToAttack fires on Spacebar when a target is live — paths to target
	// and begins auto-attacking once in range.
	HandleMoveToAttack(g *Game)
	// HandleSkillActivation fires the artifact in slot slotIdx (0-indexed, 6 = elite).
	HandleSkillActivation(g *Game, slotIdx int)
	// ProcessTick runs per-frame combat logic: auto-attack timer, momentum, status ticks.
	ProcessTick(g *Game, dt float64)
}

// DevSettings controls runtime feature flags. Lives in dev_settings.json (dev-only, gitignored).
type DevSettings struct {
	UseLegacyCombat bool `json:"use_legacy_combat"`
}

// LoadDevSettings reads dev_settings.json next to the binary; returns safe defaults on any error.
func LoadDevSettings() DevSettings {
	defaults := DevSettings{UseLegacyCombat: true} // legacy on until new engine is wired
	data, err := os.ReadFile("dev_settings.json")
	if err != nil {
		return defaults
	}
	var s DevSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return defaults
	}
	return s
}
