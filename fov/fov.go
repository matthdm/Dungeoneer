package fov

import (
	"dungeoneer/constants"
	"dungeoneer/levels"
	"math"
)

type Line struct {
	X1, Y1, X2, Y2 float64
	Path           []Point
}

type Point struct {
	X int
	Y int
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
	numRays := constants.MaxFOVRays
	// Reduce the number of rays on very large maps to keep performance
	if level.W > 80 || level.H > 80 {
		numRays /= 3
	} else if level.W > 50 || level.H > 50 {
		numRays /= 2
	}

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
		angle := float64(i) * float64(2.0*math.Pi/float64(numRays)) // evenly spaced angles around the player

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
			// Move the endpoint slightly into the wall so the tile
			// is registered as visible even when the ray hits a corner.
			const pen = 0.01
			hitX += pen * math.Cos(angle)
			hitY += pen * math.Sin(angle)

			path := TraceLineToTiles(cx, cy, hitX, hitY)
			rays = append(rays, Line{X1: cx, Y1: cy, X2: hitX, Y2: hitY, Path: path})
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

// TraceLineToTiles returns a list of tile coordinates the ray crosses
func TraceLineToTiles(x1, y1, x2, y2 float64) []Point {
	var points []Point

	// Convert float coordinates to tile grid
	startX := int(math.Floor(x1))
	startY := int(math.Floor(y1))
	endX := int(math.Floor(x2))
	endY := int(math.Floor(y2))

	dx := math.Abs(float64(endX - startX))
	dy := math.Abs(float64(endY - startY))

	stepX := 1
	if endX < startX {
		stepX = -1
	}
	stepY := 1
	if endY < startY {
		stepY = -1
	}

	x := startX
	y := startY
	points = append(points, Point{X: x, Y: y})

	if dx > dy {
		err := dx / 2.0
		for x != endX {
			x += stepX
			err -= dy
			if err < 0 {
				y += stepY
				err += dx
			}
			points = append(points, Point{X: x, Y: y})
		}
	} else {
		err := dy / 2.0
		for y != endY {
			y += stepY
			err -= dx
			if err < 0 {
				x += stepX
				err += dy
			}
			points = append(points, Point{X: x, Y: y})
		}
	}

	return points
}
