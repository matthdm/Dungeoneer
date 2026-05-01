package entities

import (
	"dungeoneer/coords"
	"math"
)

// isoToScreenFloat converts cartesian world coordinates to isometric screen
// coordinates. Delegates to coords.ToIso — the single source of truth for
// the projection formula.
func isoToScreenFloat(x, y float64, tileSize int) (float64, float64) {
	return coords.ToIso(x, y, tileSize)
}

func IsAdjacent(x1, y1, x2, y2 int) bool {
	dx := math.Abs(float64(x1 - x2))
	dy := math.Abs(float64(y1 - y2))
	return (dx+dy == 1) // orthogonally adjacent
}

func IsAdjacentRanged(x1, y1, x2, y2 int, maxDist int) bool {
	dx := math.Abs(float64(x1 - x2))
	dy := math.Abs(float64(y1 - y2))
	return dx+dy <= float64(maxDist)
}
