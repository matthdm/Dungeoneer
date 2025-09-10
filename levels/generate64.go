package levels

import (
	"image"
	"sort"

	"math/rand/v2"

	"dungeoneer/tiles"
)

type GenParams struct {
	Seed                       int64
	Width, Height              int
	RoomCountMin, RoomCountMax int
	RoomWMin, RoomWMax         int
	RoomHMin, RoomHMax         int
	CorridorWidth              int
	DashLaneMinLen             int
	GrappleRange               int
	Extras                     int
}

type rect struct{ X, Y, W, H int }

type edge struct{ A, B int }

func Generate64x64(p GenParams) *Level {
	if p.Width == 0 {
		p.Width = 64
	}
	if p.Height == 0 {
		p.Height = 64
	}
	if p.RoomCountMin == 0 {
		p.RoomCountMin = 8
	}
	if p.RoomCountMax == 0 {
		p.RoomCountMax = 14
	}
	if p.RoomCountMax < p.RoomCountMin {
		p.RoomCountMax = p.RoomCountMin
	}
	if p.RoomWMin == 0 {
		p.RoomWMin = 6
	}
	if p.RoomWMax == 0 {
		p.RoomWMax = 12
	}
	if p.RoomWMax < p.RoomWMin {
		p.RoomWMax = p.RoomWMin
	}
	if p.RoomHMin == 0 {
		p.RoomHMin = 6
	}
	if p.RoomHMax == 0 {
		p.RoomHMax = 12
	}
	if p.RoomHMax < p.RoomHMin {
		p.RoomHMax = p.RoomHMin
	}
	if p.CorridorWidth == 0 {
		p.CorridorWidth = 3
	}
	if p.DashLaneMinLen == 0 {
		p.DashLaneMinLen = 8
	}
	if p.GrappleRange == 0 {
		p.GrappleRange = 12
	}
	rng := rand.New(rand.NewPCG(uint64(p.Seed), uint64(p.Seed^0xface)))

	l := NewEmptyLevel(p.Width, p.Height)

	depth := 3
	if p.RoomCountMin > 8 {
		depth = 4
	}
	regions := bspRegions(p.Width, p.Height, depth, p, rng)
	centers := poissonInRegions(regions, p, rng)
	rooms := growRooms(l, centers, p, rng)
	edges := connectKNN(centers, 3)
	edges = mstPlusExtras(edges, centers, p.Extras, rng)
	carveCorridors(l, centers, edges, p.CorridorWidth, rng)
	widenPinches(l, p.CorridorWidth)
	tagDashLanes(l, p.CorridorWidth, p.DashLaneMinLen)
	placeGrappleAnchors(l, rooms, p.GrappleRange, rng)
	pruneDeadEnds(l, 3)
	ensureConnectivity(l)
	return l
}

// --- helpers ---

func bspRegions(w, h, depth int, p GenParams, rng *rand.Rand) []rect {
	regs := []rect{{0, 0, w, h}}
	minW := p.RoomWMax + p.CorridorWidth*2
	minH := p.RoomHMax + p.CorridorWidth*2
	for d := 0; d < depth; d++ {
		var next []rect
		for _, r := range regs {
			if r.W <= minW*2 && r.H <= minH*2 {
				next = append(next, r)
				continue
			}
			splitVert := r.W >= r.H
			if splitVert {
				if r.W < minW*2 {
					next = append(next, r)
					continue
				}
				min := int(float64(r.W) * 0.4)
				max := int(float64(r.W) * 0.6)
				if max-min <= 1 {
					next = append(next, r)
					continue
				}
				s := rng.IntN(max-min) + min
				left := rect{r.X, r.Y, s, r.H}
				right := rect{r.X + s, r.Y, r.W - s, r.H}
				next = append(next, left, right)
			} else {
				if r.H < minH*2 {
					next = append(next, r)
					continue
				}
				min := int(float64(r.H) * 0.4)
				max := int(float64(r.H) * 0.6)
				if max-min <= 1 {
					next = append(next, r)
					continue
				}
				s := rng.IntN(max-min) + min
				top := rect{r.X, r.Y, r.W, s}
				bot := rect{r.X, r.Y + s, r.W, r.H - s}
				next = append(next, top, bot)
			}
		}
		regs = next
	}
	return regs
}

func poissonInRegions(rs []rect, p GenParams, rng *rand.Rand) []image.Point {
	marginX := p.CorridorWidth + p.RoomWMax/2
	marginY := p.CorridorWidth + p.RoomHMax/2
	var pts []image.Point
	// one candidate per region
	for _, r := range rs {
		minX := r.X + marginX
		maxX := r.X + r.W - marginX - 1
		minY := r.Y + marginY
		maxY := r.Y + r.H - marginY - 1
		if minX > maxX || minY > maxY {
			continue
		}
		cx := minX + rng.IntN(maxX-minX+1)
		cy := minY + rng.IntN(maxY-minY+1)
		pt := image.Point{cx, cy}
		ok := true
		for _, q := range pts {
			dx := q.X - pt.X
			dy := q.Y - pt.Y
			if dx*dx+dy*dy < 64 {
				ok = false
				break
			}
		}
		if ok {
			pts = append(pts, pt)
		}
	}
	// supplement until reaching RoomCountMin
	attempts := 0
	for len(pts) < p.RoomCountMin && attempts < 1000 {
		cx := marginX + rng.IntN(p.Width-2*marginX)
		cy := marginY + rng.IntN(p.Height-2*marginY)
		pt := image.Point{cx, cy}
		ok := true
		for _, q := range pts {
			dx := q.X - pt.X
			dy := q.Y - pt.Y
			if dx*dx+dy*dy < 64 {
				ok = false
				break
			}
		}
		if ok {
			pts = append(pts, pt)
		}
		attempts++
	}
	if len(pts) > p.RoomCountMax {
		rng.Shuffle(len(pts), func(i, j int) { pts[i], pts[j] = pts[j], pts[i] })
		pts = pts[:p.RoomCountMax]
	}
	return pts
}

func growRooms(L *Level, centers []image.Point, p GenParams, rng *rand.Rand) []rect {
	rooms := make([]rect, len(centers))
	margin := p.CorridorWidth
	for i, c := range centers {
		w := p.RoomWMin + rng.IntN(p.RoomWMax-p.RoomWMin+1)
		h := p.RoomHMin + rng.IntN(p.RoomHMax-p.RoomHMin+1)
		x := c.X - w/2
		y := c.Y - h/2
		if x < margin {
			x = margin
		}
		if y < margin {
			y = margin
		}
		if x+w >= L.W-margin {
			x = L.W - margin - w
		}
		if y+h >= L.H-margin {
			y = L.H - margin - h
		}
		r := rect{x, y, w, h}
		rooms[i] = r
		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				L.Tiles[yy][xx].IsWalkable = true
			}
		}
		// shave corners
		L.Tiles[y][x].IsWalkable = false
		L.Tiles[y][x+w-1].IsWalkable = false
		L.Tiles[y+h-1][x].IsWalkable = false
		L.Tiles[y+h-1][x+w-1].IsWalkable = false
	}
	return rooms
}

func connectKNN(pts []image.Point, k int) []edge {
	n := len(pts)
	m := map[edge]struct{}{}
	for i := 0; i < n; i++ {
		type pair struct{ j, d int }
		arr := make([]pair, 0, n-1)
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			dx := pts[i].X - pts[j].X
			dy := pts[i].Y - pts[j].Y
			arr = append(arr, pair{j, dx*dx + dy*dy})
		}
		sort.Slice(arr, func(a, b int) bool { return arr[a].d < arr[b].d })
		for c := 0; c < k && c < len(arr); c++ {
			a, b := i, arr[c].j
			if a > b {
				a, b = b, a
			}
			m[edge{a, b}] = struct{}{}
		}
	}
	es := make([]edge, 0, len(m))
	for e := range m {
		es = append(es, e)
	}
	return es
}

func mstPlusExtras(edges []edge, pts []image.Point, extras int, rng *rand.Rand) []edge {
	n := len(pts)
	if n == 0 {
		return nil
	}
	in := make([]bool, n)
	in[0] = true
	var result []edge
	for count := 1; count < n; count++ {
		bestA, bestB, bestD := 0, 0, 1<<30
		for i := 0; i < n; i++ {
			if !in[i] {
				continue
			}
			for j := 0; j < n; j++ {
				if in[j] {
					continue
				}
				dx := pts[i].X - pts[j].X
				dy := pts[i].Y - pts[j].Y
				d := dx*dx + dy*dy
				if d < bestD {
					bestA, bestB, bestD = i, j, d
				}
			}
		}
		result = append(result, edge{bestA, bestB})
		in[bestB] = true
	}
	existing := map[edge]struct{}{}
	for _, e := range result {
		a, b := e.A, e.B
		if a > b {
			a, b = b, a
		}
		existing[edge{a, b}] = struct{}{}
	}
	var cand []edge
	for _, e := range edges {
		a, b := e.A, e.B
		if a > b {
			a, b = b, a
		}
		if _, ok := existing[edge{a, b}]; !ok {
			cand = append(cand, edge{a, b})
		}
	}
	for extras > 0 && len(cand) > 0 {
		idx := rng.IntN(len(cand))
		result = append(result, cand[idx])
		cand = append(cand[:idx], cand[idx+1:]...)
		extras--
	}
	return result
}

func carveCorridors(L *Level, centers []image.Point, es []edge, W int, rng *rand.Rand) {
	half := W / 2
	for _, e := range es {
		a := centers[e.A]
		b := centers[e.B]
		x1, y1 := a.X, a.Y
		x2, y2 := b.X, b.Y
		horizFirst := rng.IntN(2) == 0
		if horizFirst {
			carveCorridorSegment(L, x1, y1, x1, y2, half)
			carveCorridorSegment(L, x1, y2, x2, y2, half)
		} else {
			carveCorridorSegment(L, x1, y1, x2, y1, half)
			carveCorridorSegment(L, x2, y1, x2, y2, half)
		}
	}
}

func carveCorridorSegment(L *Level, x1, y1, x2, y2, half int) {
	if x1 == x2 {
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for y := y1; y <= y2; y++ {
			for dx := -half; dx <= half; dx++ {
				nx := x1 + dx
				if nx >= 0 && nx < L.W && y >= 0 && y < L.H {
					L.Tiles[y][nx].IsWalkable = true
				}
			}
		}
	} else if y1 == y2 {
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		for x := x1; x <= x2; x++ {
			for dy := -half; dy <= half; dy++ {
				ny := y1 + dy
				if x >= 0 && x < L.W && ny >= 0 && ny < L.H {
					L.Tiles[ny][x].IsWalkable = true
				}
			}
		}
	}
}

func widenPinches(L *Level, W int) {
	for y := 1; y < L.H-1; y++ {
		for x := 1; x < L.W-1; x++ {
			if L.Tiles[y][x].IsWalkable {
				continue
			}
			cnt := 0
			if L.Tiles[y-1][x].IsWalkable {
				cnt++
			}
			if L.Tiles[y+1][x].IsWalkable {
				cnt++
			}
			if L.Tiles[y][x-1].IsWalkable {
				cnt++
			}
			if L.Tiles[y][x+1].IsWalkable {
				cnt++
			}
			if cnt >= 3 {
				L.Tiles[y][x].IsWalkable = true
			}
		}
	}
}

func tagDashLanes(L *Level, W, minLen int) {
	half := W / 2
	// horizontal
	for y := half; y < L.H-half; y++ {
		run := 0
		for x := 0; x < L.W; x++ {
			if isWideFloor(L, x, y, half, true) {
				run++
			} else {
				if run >= minLen {
					for xx := x - run; xx < x; xx++ {
						for off := -half; off <= half; off++ {
							L.Tiles[y+off][xx].SetTag(tiles.TagDashLane)
						}
					}
				}
				run = 0
			}
		}
		if run >= minLen {
			for xx := L.W - run; xx < L.W; xx++ {
				for off := -half; off <= half; off++ {
					L.Tiles[y+off][xx].SetTag(tiles.TagDashLane)
				}
			}
		}
	}
	// vertical
	for x := half; x < L.W-half; x++ {
		run := 0
		for y := 0; y < L.H; y++ {
			if isWideFloor(L, x, y, half, false) {
				run++
			} else {
				if run >= minLen {
					for yy := y - run; yy < y; yy++ {
						for off := -half; off <= half; off++ {
							L.Tiles[yy][x+off].SetTag(tiles.TagDashLane)
						}
					}
				}
				run = 0
			}
		}
		if run >= minLen {
			for yy := L.H - run; yy < L.H; yy++ {
				for off := -half; off <= half; off++ {
					L.Tiles[yy][x+off].SetTag(tiles.TagDashLane)
				}
			}
		}
	}
}

func isWideFloor(L *Level, x, y, half int, horiz bool) bool {
	if x < 0 || y < 0 || x >= L.W || y >= L.H {
		return false
	}
	if !L.Tiles[y][x].IsWalkable {
		return false
	}
	if horiz {
		for off := -half; off <= half; off++ {
			ny := y + off
			if ny < 0 || ny >= L.H || !L.Tiles[ny][x].IsWalkable {
				return false
			}
		}
	} else {
		for off := -half; off <= half; off++ {
			nx := x + off
			if nx < 0 || nx >= L.W || !L.Tiles[y][nx].IsWalkable {
				return false
			}
		}
	}
	return true
}

func placeGrappleAnchors(L *Level, rooms []rect, rangeTiles int, rng *rand.Rand) {
	dirs := []image.Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, r := range rooms {
		attempts := 0
		placed := 0
		for placed < 1 && attempts < 10 {
			attempts++
			side := rng.IntN(4)
			var x, y int
			switch side {
			case 0:
				x = r.X
				y = r.Y + r.H/2
			case 1:
				x = r.X + r.W - 1
				y = r.Y + r.H/2
			case 2:
				x = r.X + r.W/2
				y = r.Y
			case 3:
				x = r.X + r.W/2
				y = r.Y + r.H - 1
			}
			d := dirs[side]
			tx, ty := x, y
			found := false
			for step := 0; step < rangeTiles; step++ {
				tx += d.X
				ty += d.Y
				if tx < 0 || ty < 0 || tx >= L.W || ty >= L.H {
					break
				}
				if L.Tiles[ty][tx].IsWalkable {
					if bresenhamClear(L, x, y, tx, ty) {
						L.Tiles[y][x].SetTag(tiles.TagGrappleAnchor)
						placed++
						found = true
					}
					break
				}
			}
			if found {
				break
			}
		}
		if placed == 0 {
			x := r.X + r.W/2
			y := r.Y + r.H/2
			L.Tiles[y][x].SetTag(tiles.TagGrappleAnchor)
		}
	}
}

func bresenhamClear(L *Level, x0, y0, x1, y1 int) bool {
	dx := abs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -abs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		if x0 == x1 && y0 == y1 {
			break
		}
		if !(x0 == x0 && y0 == y0) { // ignore start
			if !L.Tiles[y0][x0].IsWalkable {
				return false
			}
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
	return true
}

func pruneDeadEnds(L *Level, minLen int) {
	changed := true
	dirs := []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for changed {
		changed = false
		for y := 1; y < L.H-1; y++ {
			for x := 1; x < L.W-1; x++ {
				t := L.Tiles[y][x]
				if !t.IsWalkable || t.HasTag(tiles.TagGrappleAnchor) {
					continue
				}
				cnt := 0
				for _, d := range dirs {
					if L.Tiles[y+d.Y][x+d.X].IsWalkable {
						cnt++
					}
				}
				if cnt <= 1 {
					t.IsWalkable = false
					changed = true
				}
			}
		}
	}
}

func ensureConnectivity(L *Level) {
	comps := floodComponents(L)
	for len(comps) > 1 {
		a := comps[0]
		b := comps[1]
		best := 1 << 30
		var pa, pb image.Point
		for _, p := range a {
			for _, q := range b {
				d := (p.X-q.X)*(p.X-q.X) + (p.Y-q.Y)*(p.Y-q.Y)
				if d < best {
					best = d
					pa = p
					pb = q
				}
			}
		}
		carveCorridorSegment(L, pa.X, pa.Y, pb.X, pa.Y, 1)
		carveCorridorSegment(L, pb.X, pa.Y, pb.X, pb.Y, 1)
		comps = floodComponents(L)
	}
}

func floodComponents(L *Level) [][]image.Point {
	visited := make([][]bool, L.H)
	for i := range visited {
		visited[i] = make([]bool, L.W)
	}
	dirs := []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	var comps [][]image.Point
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if visited[y][x] || !L.Tiles[y][x].IsWalkable {
				continue
			}
			queue := []image.Point{{x, y}}
			visited[y][x] = true
			var comp []image.Point
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				comp = append(comp, p)
				for _, d := range dirs {
					nx, ny := p.X+d.X, p.Y+d.Y
					if nx >= 0 && ny >= 0 && nx < L.W && ny < L.H && !visited[ny][nx] && L.Tiles[ny][nx].IsWalkable {
						visited[ny][nx] = true
						queue = append(queue, image.Point{nx, ny})
					}
				}
			}
			comps = append(comps, comp)
		}
	}
	return comps
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
