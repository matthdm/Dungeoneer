package levels

// FindFarthestWalkable returns the walkable tile farthest from (fromX, fromY)
// using BFS. This is used to place floor exits as far from the player spawn as
// possible, maximising exploration per floor.
func FindFarthestWalkable(l *Level, fromX, fromY int) (int, int) {
	if l == nil || !l.IsWalkable(fromX, fromY) {
		return fromX, fromY
	}

	type cell struct{ x, y int }
	visited := make([][]bool, l.H)
	for y := range visited {
		visited[y] = make([]bool, l.W)
	}

	queue := []cell{{fromX, fromY}}
	visited[fromY][fromX] = true
	bestX, bestY := fromX, fromY

	dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		bestX, bestY = cur.x, cur.y

		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if nx < 0 || ny < 0 || nx >= l.W || ny >= l.H {
				continue
			}
			if visited[ny][nx] {
				continue
			}
			if !l.IsWalkable(nx, ny) {
				continue
			}
			visited[ny][nx] = true
			queue = append(queue, cell{nx, ny})
		}
	}

	return bestX, bestY
}

// FindSpawnPoint returns a walkable tile near the center of the level.
// Falls back to scanning for any walkable tile if the center is blocked.
func FindSpawnPoint(l *Level) (int, int) {
	cx, cy := l.W/2, l.H/2

	// Spiral outward from center
	for radius := 0; radius < l.W; radius++ {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx != -radius && dx != radius && dy != -radius && dy != radius {
					continue // only check perimeter of this radius
				}
				x, y := cx+dx, cy+dy
				if l.IsWalkable(x, y) {
					return x, y
				}
			}
		}
	}

	return cx, cy
}
