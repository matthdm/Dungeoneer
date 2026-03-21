package levels

import (
	"dungeoneer/constants"
	"dungeoneer/sprites"
	"math/rand/v2"
)

// MazeConfig defines parameters for generating a maze level.
type MazeConfig struct {
	Width, Height int
	Tessellation  string // "orthogonal" or "fractal"
	Routing       string // "prim", "eller", "braid", "unicursal"
	Texture       string // "elitism", "run", "river"
	WallFlavor    string // wall palette flavor
}

// GenerateMaze creates a new Level containing a maze using the provided sprite sheet.
func GenerateMaze(cfg MazeConfig, ss *sprites.SpriteSheet) *Level {
	if ss == nil {
		var err error
		ss, err = sprites.LoadSpriteSheet(constants.DefaultTileSize)
		if err != nil {
			return nil
		}
	}

	l := CreateNewBlankLevel(cfg.Width, cfg.Height, constants.DefaultTileSize, ss)

	// Load the requested wall palette and store it in the SpriteSheet map.
	wss, err := sprites.LoadWallSpriteSheet(cfg.WallFlavor)
	if err != nil {
		wss, _ = sprites.LoadWallSpriteSheet("normal")
	}
	if ss.Walls == nil {
		ss.Walls = make(map[string]*sprites.WallSpriteSheet)
	}
	ss.Walls[cfg.WallFlavor] = wss

	var grid [][]bool
	switch cfg.Tessellation {
	case "fractal":
		grid = fractalMaze(cfg.Width, cfg.Height, cfg.Routing)
	default:
		grid = orthogonalMaze(cfg.Width, cfg.Height, cfg.Routing)
	}

	applyTexture(cfg.Texture, grid)
	addEntranceExit(grid)
	applyGridToLevel(grid, l, wss)
	return l
}

type point struct{ x, y int }

// orthogonalMaze generates a maze on a standard orthogonal grid.
func orthogonalMaze(w, h int, routing string) [][]bool {
	switch routing {
	case "eller":
		return ellerMaze(w, h)
	case "braid":
		g := primMaze(w, h)
		braidify(g)
		return g
	case "unicursal":
		return unicursalMaze(w, h)
	default:
		return primMaze(w, h)
	}
}

// primMaze implements a simple randomized Prim's algorithm.
func primMaze(w, h int) [][]bool {
	grid := make([][]bool, h)
	visited := make([][]bool, h)
	for y := range grid {
		grid[y] = make([]bool, w)
		visited[y] = make([]bool, w)
		for x := range grid[y] {
			grid[y][x] = true
		}
	}

	start := point{1, 1}
	if start.x >= w || start.y >= h {
		start = point{0, 0}
	}
	grid[start.y][start.x] = false
	visited[start.y][start.x] = true

	type edge struct{ from, to point }
	var frontier []edge
	add := func(p point) {
		dirs := []point{{0, -2}, {0, 2}, {-2, 0}, {2, 0}}
		for _, d := range dirs {
			nx, ny := p.x+d.x, p.y+d.y
			if nx <= 0 || ny <= 0 || nx >= w || ny >= h {
				continue
			}
			if !visited[ny][nx] {
				frontier = append(frontier, edge{p, point{nx, ny}})
			}
		}
	}

	add(start)
	for len(frontier) > 0 {
		idx := rand.IntN(len(frontier))
		e := frontier[idx]
		frontier = append(frontier[:idx], frontier[idx+1:]...)
		if visited[e.to.y][e.to.x] {
			continue
		}
		mid := point{(e.from.x + e.to.x) / 2, (e.from.y + e.to.y) / 2}
		grid[e.to.y][e.to.x] = false
		grid[mid.y][mid.x] = false
		visited[e.to.y][e.to.x] = true
		add(e.to)
	}
	// surround with walls
	out := make([][]bool, h)
	for y := range out {
		out[y] = make([]bool, w)
		copy(out[y], grid[y])
	}
	return out
}

// braidify removes dead ends to create loops.
func braidify(g [][]bool) {
	h := len(g)
	if h == 0 {
		return
	}
	w := len(g[0])
	dirs := []point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			if g[y][x] {
				continue
			}
			cnt := 0
			var walls []point
			for _, d := range dirs {
				nx, ny := x+d.x, y+d.y
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				if g[ny][nx] {
					walls = append(walls, point{nx, ny})
				} else {
					cnt++
				}
			}
			if cnt == 1 && len(walls) > 0 && rand.Float64() < 0.5 {
				wpt := walls[rand.IntN(len(walls))]
				g[wpt.y][wpt.x] = false
			}
		}
	}
}

// unicursalMaze creates a simple S-shaped unicursal labyrinth.
func unicursalMaze(w, h int) [][]bool {
	grid := make([][]bool, h)
	for y := range grid {
		grid[y] = make([]bool, w)
		for x := range grid[y] {
			grid[y][x] = true
		}
	}
	if w < 2 || h < 2 {
		return grid
	}
	right := true
	y := 1
	for y < h-1 {
		if right {
			for x := 1; x < w-1; x++ {
				grid[y][x] = false
			}
			if y+1 < h-1 {
				grid[y+1][w-2] = false
			}
		} else {
			for x := w - 2; x > 0; x-- {
				grid[y][x] = false
			}
			if y+1 < h-1 {
				grid[y+1][1] = false
			}
		}
		right = !right
		y += 2
	}
	return grid
}

// simple implementation of Eller’s algorithm.
func ellerMaze(w, h int) [][]bool {
	grid := make([][]bool, h)
	for y := range grid {
		grid[y] = make([]bool, w)
		for x := range grid[y] {
			grid[y][x] = true
		}
	}

	sets := make([]int, w)
	nextSet := 1
	for y := 1; y < h; y += 2 {
		// step 1: assign sets
		for x := 1; x < w; x += 2 {
			if sets[x] == 0 {
				sets[x] = nextSet
				nextSet++
			}
		}
		// step 2: join cells horizontally
		for x := 1; x < w-2; x += 2 {
			if rand.IntN(2) == 0 && sets[x] != sets[x+2] {
				grid[y][x+1] = false
				old := sets[x+2]
				for i := range sets {
					if sets[i] == old {
						sets[i] = sets[x]
					}
				}
			}
		}
		// step 3: carve vertical connections
		newSets := make([]int, w)
		x := 1
		for x < w {
			setID := sets[x]
			var cells []int
			for x < w && sets[x] == setID {
				cells = append(cells, x)
				x += 2
			}
			count := rand.IntN(len(cells)) + 1
			rand.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
			for i := 0; i < count; i++ {
				cx := cells[i]
				if y+1 < h {
					grid[y+1][cx] = false
					newSets[cx] = setID
				}
			}
		}
		if y+2 < h {
			sets = newSets
		}
	}
	return grid
}

// fractalMaze uses recursive division (a simple fractal-like tessellation).
func fractalMaze(w, h int, routing string) [][]bool {
	grid := make([][]bool, h)
	for y := range grid {
		grid[y] = make([]bool, w)
		for x := range grid[y] {
			grid[y][x] = false
		}
	}
	// surround borders
	for x := 0; x < w; x++ {
		grid[0][x] = true
		grid[h-1][x] = true
	}
	for y := 0; y < h; y++ {
		grid[y][0] = true
		grid[y][w-1] = true
	}

	var divide func(x, y, width, height int, vertical bool)
	divide = func(x, y, width, height int, vertical bool) {
		// When the region becomes less than or equal to 2 tiles wide or
		// high, further division would result in invalid calls to
		// rand.IntN with zero or negative arguments. Stop recursing at
		// this size to avoid panics.
		if width <= 2 || height <= 2 {
			return
		}
		if vertical {
			wx := x + 1 + rand.IntN(width-2)
			passage := y + rand.IntN(height)
			for i := y; i < y+height; i++ {
				grid[i][wx] = true
			}
			grid[passage][wx] = false
			divide(x, y, wx-x, height, !vertical)
			divide(wx+1, y, x+width-wx-1, height, !vertical)
		} else {
			wy := y + 1 + rand.IntN(height-2)
			passage := x + rand.IntN(width)
			for i := x; i < x+width; i++ {
				grid[wy][i] = true
			}
			grid[wy][passage] = false
			divide(x, y, width, wy-y, !vertical)
			divide(x, wy+1, width, y+height-wy-1, !vertical)
		}
	}
	divide(1, 1, w-2, h-2, w > h)
	return grid
}

// applyTexture adds noise based on the selected texture flavor.
func applyTexture(flavor string, g [][]bool) {
	h := len(g)
	if h == 0 {
		return
	}
	w := len(g[0])
	switch flavor {
	case "elitism":
		for y := 1; y < h-1; y++ {
			for x := 1; x < w-1; x++ {
				if g[y][x] && rand.Float64() < 0.05 {
					g[y][x] = false
				}
			}
		}
	case "run":
		for y := 1; y < h-1; y += 2 {
			for x := 1; x < w-2; x += 2 {
				if g[y][x] == false && rand.Float64() < 0.4 {
					g[y][x+1] = false
				}
			}
		}
	case "river":
		for y := 1; y < h-1; y++ {
			for x := 1; x < w-1; x++ {
				if g[y][x] && rand.Float64() < 0.1 {
					g[y][x] = false
				}
			}
		}
	}
}

// addEntranceExit carves an opening at the top and bottom borders.
func addEntranceExit(g [][]bool) {
	h := len(g)
	if h == 0 {
		return
	}
	w := len(g[0])
	if w < 2 || h < 2 {
		return
	}
	g[0][1] = false
	g[h-1][w-2] = false
}

// applyGridToLevel converts the boolean grid into actual tiles with wall sprites.
func applyGridToLevel(g [][]bool, l *Level, wss *sprites.WallSpriteSheet) {
	h := len(g)
	if h == 0 {
		return
	}
	w := len(g[0])
	if wss == nil {
		return
	}
	for y := 0; y < h && y < l.H; y++ {
		for x := 0; x < w && x < l.W; x++ {
			if g[y][x] {
				l.Tiles[y][x].AddSpriteByID("Wall", wss.Wall)
				l.Tiles[y][x].IsWalkable = false
			}
		}
	}
}
