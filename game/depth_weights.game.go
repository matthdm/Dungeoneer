package game

import "strings"

// depthWeightForSprite returns a default depth weight for a sprite ID.
func depthWeightForSprite(id string) float64 {
	lower := strings.ToLower(id)

	switch {
	case strings.Contains(lower, "wall") || strings.Contains(lower, "beam") ||
		strings.Contains(lower, "statue") || strings.Contains(lower, "tree") ||
		strings.Contains(lower, "log") || strings.Contains(lower, "chunk"):
		return 1.0
	case isFloorSprite(lower):
		// floors are drawn separately but return a negative weight just in case
		if lower == "floor" {
			return -1.0
		}
		return -0.9
	default:
		return 0.5
	}
}

// isFloorSprite reports whether the sprite represents a floor-like tile.
func isFloorSprite(id string) bool {
	return strings.Contains(id, "floor") || strings.Contains(id, "water") ||
		strings.Contains(id, "lava") || strings.Contains(id, "stairs") ||
		strings.Contains(id, "portal") || strings.Contains(id, "trap")
}
