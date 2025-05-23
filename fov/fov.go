package fov

import (
	"dungeoneer/levels"
	"math"
)

type Line struct {
	X1, Y1, X2, Y2 float64
}

type RayCache struct {
	PlayerX, PlayerY float64
	Rays             []Line
}

var cached RayCache

func RayCasting(cx, cy float64, walls []Line, level *levels.Level) []Line {
	if cx == cached.PlayerX && cy == cached.PlayerY && cached.Rays != nil {
		return cached.Rays
	}

	const rayLength = 1000
	const numRays = 360
	var rays []Line

	// Offset to avoid standing "inside" a wall
	const epsilon = 0.01
	// get the tile the player is standing on
	tx := int(math.Floor(cx))
	ty := int(math.Floor(cy))

	// check neighbors and shift slightly
	if isWall(tx, ty-1, level) { // wall to the north
		cy += epsilon
	}
	if isWall(tx, ty+1, level) { // south
		cy -= epsilon
	}
	if isWall(tx-1, ty, level) { // west
		cx += epsilon
	}
	if isWall(tx+1, ty, level) { // east
		cx -= epsilon
	}

	for i := 0; i < numRays; i++ {
		angle := float64(i) * (2 * math.Pi / numRays)

		// Slightly offset the origin along the ray's direction
		ox := cx + epsilon*math.Cos(angle)
		oy := cy + epsilon*math.Sin(angle)

		ray := NewRay(ox, oy, rayLength, angle)

		minDist := math.Inf(1)
		var hitX, hitY float64

		for _, wall := range walls {
			if x, y, ok := Intersection(ray, wall); ok {
				dist := (ox-x)*(ox-x) + (oy-y)*(oy-y)
				if dist < minDist {
					minDist = dist
					hitX = x
					hitY = y
				}
			}
		}

		if !math.IsInf(minDist, 1) {
			rays = append(rays, Line{cx, cy, hitX, hitY})
		}
	}

	cached = RayCache{PlayerX: cx, PlayerY: cy, Rays: rays}

	return rays
}

func isWall(x, y int, level *levels.Level) bool {
	if y < 0 || y >= len(level.Tiles) || x < 0 || x >= len(level.Tiles[y]) {
		return true // treat out of bounds as wall
	}
	t := level.Tiles[y][x]
	return t == nil || !t.IsWalkable
}
