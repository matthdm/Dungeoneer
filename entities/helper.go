package entities

import (
	"dungeoneer/levels"
	"math"
)

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

// Check if the center point (x, y) would collide with any solid tile
func CollidesWithMap(level *levels.Level, feetX, feetY, width, height float64) bool {
	epsilon := 0.5

	left := feetX - width/2 + .6
	right := feetX + width/2 - epsilon
	bottom := feetY - epsilon
	top := feetY - height + epsilon
	centerX := feetX
	centerY := feetY - height/2

	samplePoints := [][2]float64{
		{left, top},
		{centerX, top},
		{right, top},
		{left, centerY},
		{centerX, centerY},
		{right, centerY},
		{left, bottom},
		{centerX, bottom},
		{right, bottom},
		{(left + right) / 2, top}, // <– new
	}

	for _, pt := range samplePoints {
		tx := int(math.Floor(pt[0]))
		ty := int(math.Floor(pt[1]))
		if !level.IsWalkable(tx, ty) {
			return true
		}
	}
	return false
}
