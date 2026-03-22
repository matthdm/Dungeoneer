package fov

import "dungeoneer/levels"

// isOpen returns true if the tile at (x,y) is walkable or out of bounds.
// Used to decide which faces of a wall tile should become ray-blocking segments:
// only faces that border open (walkable) space block rays — interior faces shared
// between two adjacent wall tiles are omitted so rays stop at the correct surface.
func isOpen(level *levels.Level, x, y int) bool {
	if y < 0 || y >= len(level.Tiles) || x < 0 || x >= len(level.Tiles[y]) {
		return true // out-of-bounds treated as open so boundary edges are always added
	}
	t := level.Tiles[y][x]
	return t == nil || t.IsWalkable
}

func LevelToWalls(level *levels.Level) []Line {
	var walls []Line

	for y := 0; y < level.H; y++ {
		for x := 0; x < level.W; x++ {
			t := level.Tiles[y][x]
			if t == nil || t.IsWalkable {
				continue
			}

			xf := float64(x)
			yf := float64(y)

			// Only emit a face if the neighbour in that direction is open (walkable).
			// This strips interior edges shared between two adjacent wall tiles,
			// which caused rays to terminate inside thick walls or skip the near face.
			if isOpen(level, x, y-1) { // north neighbour open → north face blocks
				walls = append(walls, Line{X1: xf, Y1: yf, X2: xf + 1, Y2: yf})
			}
			if isOpen(level, x+1, y) { // east neighbour open → east face blocks
				walls = append(walls, Line{X1: xf + 1, Y1: yf, X2: xf + 1, Y2: yf + 1})
			}
			if isOpen(level, x, y+1) { // south neighbour open → south face blocks
				walls = append(walls, Line{X1: xf + 1, Y1: yf + 1, X2: xf, Y2: yf + 1})
			}
			if isOpen(level, x-1, y) { // west neighbour open → west face blocks
				walls = append(walls, Line{X1: xf, Y1: yf + 1, X2: xf, Y2: yf})
			}
		}
	}
	return walls
}
