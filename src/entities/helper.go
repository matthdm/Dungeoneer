package entities

import (
	"dungeoneer/coords"
	"math"
)

// isoToScreenFloat returns the blit origin (top-left) of a tile-sized sprite
// cell anchored at (x, y). It is for drawing sprites only — it is not the
// screen position of the world point (x, y); see coords.GroundToIso for that.
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
