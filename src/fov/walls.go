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
	// Diagonal corner anti-peek segments.
	//
	// When two wall tiles share only a corner point (e.g. NW+SE are walls,
	// NE+SW are open), geometric rays can slip through that zero-width gap:
	// the ray passes exactly through the shared corner without intersecting
	// any boundary face, because the faces only cover the cardinal edges.
	//
	// Fix: for each such diagonal pair, add a short segment that spans the
	// corner and lies along the wall diagonal, so gap-rays hit it instead
	// of passing through.  The segment runs from the centre of one wall tile
	// to the centre of the other (half-length = 0.5 each side of the corner)
	// which is long enough to catch any ray that enters the gap region.
	//
	//   '\' gap (NW+SE walls, NE+SW open) → segment goes NW↔SE
	//   '/' gap (NE+SW walls, NW+SE open) → segment goes NE↔SW
	const dc = 0.5
	for gy := 1; gy < level.H; gy++ {
		for gx := 1; gx < level.W; gx++ {
			// Corner at grid point (gx, gy), bordered by four tiles:
			//   NW=(gx-1,gy-1)  NE=(gx,gy-1)
			//   SW=(gx-1,gy)    SE=(gx,gy)
			nw := !isOpen(level, gx-1, gy-1)
			ne := !isOpen(level, gx, gy-1)
			sw := !isOpen(level, gx-1, gy)
			se := !isOpen(level, gx, gy)
			px, py := float64(gx), float64(gy)
			if nw && se && !ne && !sw {
				// '\' gap
				walls = append(walls, Line{X1: px - dc, Y1: py - dc, X2: px + dc, Y2: py + dc})
			}
			if ne && sw && !nw && !se {
				// '/' gap
				walls = append(walls, Line{X1: px + dc, Y1: py - dc, X2: px - dc, Y2: py + dc})
			}
		}
	}

	return walls
}
