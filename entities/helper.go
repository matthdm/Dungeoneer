package entities

import "math"

func isoToScreenFloat(x, y float64, tileSize int) (float64, float64) {
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
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
