package fov

import (
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

// InvalidateCache forces the next RayCasting call to recompute rays.
// Must be called whenever the wall geometry changes (e.g. on level load).
func InvalidateCache() {
	cached.Rays = nil
}

func RayCasting(cx, cy float64, walls []Line, level *levels.Level) []Line {
	if cx == cached.PlayerX && cy == cached.PlayerY && cached.Rays != nil {
		return cached.Rays
	}

	// Store original position for cache key (before any clamping/shifting).
	origCX, origCY := cx, cy

	const rayLength = 1000
	const numRays = 360
	var rays []Line

	// Offset to avoid standing "inside" a wall
	const epsilon = 0.02
	// get the tile the player is standing on
	tx := int(math.Floor(cx))
	ty := int(math.Floor(cy))

	// The collision system's visual offsets (XWallVisualOffset, box width)
	// allow InterpX+0.5 to cross into a wall tile when the player is pressed
	// against a wall on their right (cartesian east).  If the ray origin lands
	// inside a wall tile, rays aimed in that direction start past the wall face
	// and miss it entirely, lighting up the next room.
	//
	// Fix: detect the case and snap the origin back to just outside the
	// nearest walkable tile boundary.
	if isWall(tx, ty, level) {
		fracX := cx - float64(tx)
		fracY := cy - float64(ty)

		// Pull back on the axis closest to a tile boundary first.
		pulled := false
		if fracX < 0.5 && !isWall(tx-1, ty, level) {
			cx = float64(tx) - epsilon
			pulled = true
		} else if fracX >= 0.5 && !isWall(tx+1, ty, level) {
			cx = float64(tx+1) + epsilon
			pulled = true
		}
		if fracY < 0.5 && !isWall(tx, ty-1, level) {
			cy = float64(ty) - epsilon
			pulled = true
		} else if fracY >= 0.5 && !isWall(tx, ty+1, level) {
			cy = float64(ty+1) + epsilon
			pulled = true
		}
		if pulled {
			tx = int(math.Floor(cx))
			ty = int(math.Floor(cy))
		}
	}

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
			// Move the endpoint slightly into the wall so the tile
			// is registered as visible even when the ray hits a corner.
			const pen = 0.01
			hitX += pen * math.Cos(angle)
			hitY += pen * math.Sin(angle)

			path := TraceLineToTiles(cx, cy, hitX, hitY)
			rays = append(rays, Line{X1: cx, Y1: cy, X2: hitX, Y2: hitY, Path: path})
		}
	}

	cached = RayCache{PlayerX: origCX, PlayerY: origCY, Rays: rays}

	return rays
}

func isWall(x, y int, level *levels.Level) bool {
	if y < 0 || y >= len(level.Tiles) || x < 0 || x >= len(level.Tiles[y]) {
		return true // treat out of bounds as wall
	}
	t := level.Tiles[y][x]
	return t == nil || !t.IsWalkable
}

// TraceLineToTiles returns every tile the line geometrically touches,
// including tiles only grazed at a corner. Standard Bresenham picks one
// tile at each corner crossing and skips the other; this supercover variant
// emits both, eliminating the "dark corner tile" gaps in radial FOV.
func TraceLineToTiles(x1, y1, x2, y2 float64) []Point {
	var points []Point

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
				// Line crosses a tile corner: include the tile we were in
				// before the Y step so the corner tile is never skipped.
				points = append(points, Point{X: x, Y: y})
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
				// Same supercover treatment for the X-step case.
				points = append(points, Point{X: x, Y: y})
				x += stepX
				err += dy
			}
			points = append(points, Point{X: x, Y: y})
		}
	}

	return points
}
