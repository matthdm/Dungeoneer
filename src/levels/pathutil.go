package levels

type bfsCell struct{ x, y int }

// bfsFarthest runs BFS from (fromX, fromY) and returns the walkable tile
// farthest from the start, plus its BFS distance.
func bfsFarthest(l *Level, fromX, fromY int) (bestX, bestY, dist int) {
	visited := make([][]int, l.H)
	for y := range visited {
		visited[y] = make([]int, l.W)
		for x := range visited[y] {
			visited[y][x] = -1
		}
	}

	dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	queue := []bfsCell{{fromX, fromY}}
	visited[fromY][fromX] = 0
	bestX, bestY, dist = fromX, fromY, 0

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		d := visited[cur.y][cur.x]
		if d > dist {
			dist = d
			bestX, bestY = cur.x, cur.y
		}
		for _, dir := range dirs {
			nx, ny := cur.x+dir[0], cur.y+dir[1]
			if nx < 0 || ny < 0 || nx >= l.W || ny >= l.H {
				continue
			}
			if visited[ny][nx] >= 0 || !l.IsWalkable(nx, ny) {
				continue
			}
			visited[ny][nx] = d + 1
			queue = append(queue, bfsCell{nx, ny})
		}
	}
	return
}

// bfsDistMap runs BFS from (fromX, fromY) and returns the full distance map.
// Unreachable or non-walkable tiles have value -1.
func bfsDistMap(l *Level, fromX, fromY int) [][]int {
	dist := make([][]int, l.H)
	for y := range dist {
		dist[y] = make([]int, l.W)
		for x := range dist[y] {
			dist[y][x] = -1
		}
	}
	dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	queue := []bfsCell{{fromX, fromY}}
	dist[fromY][fromX] = 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		d := dist[cur.y][cur.x]
		for _, dir := range dirs {
			nx, ny := cur.x+dir[0], cur.y+dir[1]
			if nx < 0 || ny < 0 || nx >= l.W || ny >= l.H {
				continue
			}
			if dist[ny][nx] >= 0 || !l.IsWalkable(nx, ny) {
				continue
			}
			dist[ny][nx] = d + 1
			queue = append(queue, bfsCell{nx, ny})
		}
	}
	return dist
}

// anyWalkable returns the first walkable tile found, scanning from the centre.
func anyWalkable(l *Level) (int, int) {
	cx, cy := l.W/2, l.H/2
	for radius := 0; radius < l.W; radius++ {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx != -radius && dx != radius && dy != -radius && dy != radius {
					continue
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

// FindSpawnAndExit places the spawn and exit at maximum separation using a
// combined score of BFS path length + Euclidean distance.
//
// Pure BFS-diameter can produce tiles that are corridor-far but Euclidean-
// close (e.g. around the corner), which the player sees immediately. The
// combined score ensures both graph distance and straight-line distance are
// large, so the exit is out of sight from the spawn point.
//
//  Pass 1: BFS from any tile → find spawn (farthest BFS tile from centre).
//  Pass 2: BFS from spawn → score every reachable tile as
//          bfsDist + euclidean(spawn, tile) and pick the highest.
func FindSpawnAndExit(l *Level) (spawnX, spawnY, exitX, exitY int) {
	sx, sy := anyWalkable(l)
	// Pass 1: establish the spawn at one extreme of the dungeon.
	ax, ay, _ := bfsFarthest(l, sx, sy)

	// Pass 2: full distance map from spawn, pick exit via combined score.
	distMap := bfsDistMap(l, ax, ay)

	// Require the exit to be at least this many tiles from spawn (Euclidean).
	// Prevents the exit from landing in a corridor that winds back near spawn.
	const minEuclidSq = 20 * 20

	scorePass := func(minDist int) (int, int, int) {
		best := -1
		bx, by := ax, ay
		for y := 0; y < l.H; y++ {
			for x := 0; x < l.W; x++ {
				d := distMap[y][x]
				if d < 0 {
					continue // unreachable
				}
				dx, dy := x-ax, y-ay
				euclid := dx*dx + dy*dy
				if euclid < minDist {
					continue
				}
				score := d*100 + euclid
				if score > best {
					best = score
					bx, by = x, y
				}
			}
		}
		return bx, by, best
	}

	// First pass: enforce minimum Euclidean separation.
	exitX, exitY, bestScore := scorePass(minEuclidSq)
	if bestScore < 0 {
		// Dungeon too small to satisfy minimum — take any reachable tile.
		exitX, exitY, _ = scorePass(0)
	}
	return ax, ay, exitX, exitY
}

// FindFarthestWalkable returns the walkable tile farthest from (fromX, fromY).
// Kept for hub portal placement; prefer FindSpawnAndExit for dungeon floors.
func FindFarthestWalkable(l *Level, fromX, fromY int) (int, int) {
	x, y, _ := bfsFarthest(l, fromX, fromY)
	return x, y
}

// FindSpawnPoint returns a walkable tile near the centre of the level.
// Prefer FindSpawnAndExit for dungeon floors.
func FindSpawnPoint(l *Level) (int, int) {
	return anyWalkable(l)
}
