package game

import (
	"encoding/json"
	"os"
)

// BiomeFlavors maps biome key → slice of flavor lines (3 per biome).
// Loaded once at startup via LoadBiomeFlavor; never mutated after that.
var BiomeFlavors map[string][]string

// LoadBiomeFlavor reads the biome flavor JSON from path and populates BiomeFlavors.
// Non-fatal on error — the caller decides whether to log and continue.
func LoadBiomeFlavor(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var flavors map[string][]string
	if err := json.Unmarshal(data, &flavors); err != nil {
		return err
	}
	BiomeFlavors = flavors
	return nil
}

// GetBiomeFlavorLine returns the flavor line for biome at the given floor number.
// Cycles through the 3 available lines using floorNum % 3.
// Returns "" if the biome is not found or has no lines.
func GetBiomeFlavorLine(biome string, floorNum int) string {
	lines, ok := BiomeFlavors[biome]
	if !ok || len(lines) == 0 {
		return ""
	}
	idx := floorNum % len(lines)
	if idx < 0 {
		idx = 0
	}
	return lines[idx]
}
