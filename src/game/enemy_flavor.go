package game

import (
	"encoding/json"
	"os"
)

// EnemyFlavors maps "role_biome" or "role" → a first-encounter flavor line.
// Loaded once at startup via LoadEnemyFlavor; never mutated after that.
var EnemyFlavors map[string]string

// LoadEnemyFlavor reads the enemy flavor JSON from path and populates EnemyFlavors.
// Non-fatal on error — the caller decides whether to log and continue.
func LoadEnemyFlavor(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var flavors map[string]string
	if err := json.Unmarshal(data, &flavors); err != nil {
		return err
	}
	EnemyFlavors = flavors
	return nil
}

// GetEnemyFlavorLine returns the flavor line for the given role and biome.
// Lookup order: "role_biome" → "role" → "".
// Returns "" if no match is found.
func GetEnemyFlavorLine(role, biome string) string {
	if EnemyFlavors == nil {
		return ""
	}
	if line, ok := EnemyFlavors[role+"_"+biome]; ok {
		return line
	}
	if line, ok := EnemyFlavors[role]; ok {
		return line
	}
	return ""
}
