package levels

import (
	"image"
	"sort"

	"math/rand/v2"

	"dungeoneer/constants"
	"dungeoneer/sprites"
	"dungeoneer/tiles"

	"github.com/hajimehoshi/ebiten/v2"
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
	CoverageTarget             float64 // e.g., 0.40–0.55 desired walkable ratio
	FillerRoomsMax             int     // hard cap on post-pass filler rooms
	DoorLockChance             float64 // 0.0-1.0 chance for locked doors
	// Visual/theme
	WallFlavor  string // e.g., "crypt", "moss", "normal"
	FloorFlavor string // usually same list as wall flavors
}

type rect struct{ X, Y, W, H int }

type edge struct{ A, B int }

var currentCenters []image.Point
var currentParams GenParams
var rng *rand.Rand
var roomMask [][]bool
var corridorMask [][]bool

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
		p.CorridorWidth = 1
	} else {
		p.CorridorWidth = 1
	}
	if p.DashLaneMinLen == 0 {
		p.DashLaneMinLen = 8
	}
	if p.GrappleRange == 0 {
		p.GrappleRange = 12
	}
	if p.CoverageTarget <= 0 {
		p.CoverageTarget = 0.45
	} // tune 0.40–0.55
	if p.FillerRoomsMax == 0 {
		p.FillerRoomsMax = 6
	} // a few extra pockets
	if p.DoorLockChance <= 0 {
		p.DoorLockChance = 0.35
	}
	if p.WallFlavor == "" {
		p.WallFlavor = "crypt"
	}
	if p.FloorFlavor == "" {
		p.FloorFlavor = "crypt"
	}

	currentParams = p
	rng = rand.New(rand.NewPCG(uint64(p.Seed), uint64(p.Seed^0xface)))

	l := NewEmptyLevel(p.Width, p.Height)
	ss, err := sprites.LoadSpriteSheet(constants.DefaultTileSize)
	if err != nil {
		ss = nil
	}

	depth := 3
	if p.RoomCountMin > 8 {
		depth = 4
	}
	regions := bspRegions(p.Width, p.Height, depth, rng)
	centers := poissonInRegions(regions, p, rng)
	currentCenters = centers
	roomMask = make([][]bool, p.Height)
	for y := range roomMask {
		roomMask[y] = make([]bool, p.Width)
	}
	corridorMask = make([][]bool, p.Height)
	for y := range corridorMask {
		corridorMask[y] = make([]bool, p.Width)
	}
	rooms := growRooms(l, centers, p, rng)
	edges := connectKNN(centers, 3)
	edges = mstPlusExtras(edges, centers, p.Extras, rng)
	carveCorridors(l, edges, p.CorridorWidth)
	// Corridors must remain exactly 1 tile wide; no widening passes.
	// Do not carve a perimeter loop; it creates a moat between walls and void.
	optionalPerimeterLoop(l, p.CorridorWidth, corridorHalf(p.CorridorWidth), false)
	tagDashLanes(l, p.CorridorWidth, p.DashLaneMinLen)
	placeGrappleAnchors(l, rooms, p.GrappleRange, rng)
	pruneDeadEnds(l, 3)
	ensureConnectivity(l)     // Connectivity check ignores doors (treats them as walkable)
	growToCoverage(l, p, rng) // NEW
	ensureConnectivity(l)     // Ensure all filler rooms are connected
	sealWalkableEdges(l)
	if ss != nil {
		wallMask := buildWallMask(l)
		// Try to load flavored sheets; fall back to base if unavailable
		var floorImg, wallImg = ss.Floor, ss.DungeonWall

		// These helpers are assumed in your sprites pkg; if LoadWallSpriteSheet returns (imgFloor, imgWall)
		// or a struct with .Floor / .Wall, adjust the field names accordingly.
		if wss, err := sprites.LoadWallSpriteSheet(p.WallFlavor); err == nil && wss != nil {
			if wss.Wall != nil {
				wallImg = wss.Wall
			}
		}
		if fss, err := sprites.LoadWallSpriteSheet(p.FloorFlavor); err == nil && fss != nil {
			if fss.Floor != nil {
				floorImg = fss.Floor
			}
		}

		for y := 0; y < l.H; y++ {
			for x := 0; x < l.W; x++ {
				t := l.Tiles[y][x]

				// If you have layer-aware clears, use them; otherwise this is fine.
				// t.ClearSprites()

				if t.HasTag(tiles.TagDoor) {
					// Doors should not get wall sprites; keep floor under the door.
					t.AddSpriteByID(p.FloorFlavor+"_floor", floorImg)
					continue
				}
				if t.IsWalkable {
					t.AddSpriteByID(p.FloorFlavor+"_floor", floorImg)
				} else if wallMask[y][x] {
					t.AddSpriteByID(p.WallFlavor+"_wall", wallImg)
				}
			}
		}
	}
	placeDoorsFromValidatedThroats(l, p)
	return l
}

func optionalPerimeterLoop(L *Level, inset, half int, enable bool) {
	if !enable {
		return
	}
	x0, y0 := inset, inset
	x1, y1 := L.W-inset-1, L.H-inset-1
	carveCorridorSegment(L, x0, y0, x1, y0, half)
	carveCorridorSegment(L, x1, y0, x1, y1, half)
	carveCorridorSegment(L, x1, y1, x0, y1, half)
	carveCorridorSegment(L, x0, y1, x0, y0, half)
}

// --- helpers ---

func growToCoverage(L *Level, p GenParams, rng *rand.Rand) {
	target := p.CoverageTarget
	if target <= 0 {
		return
	}
	if target < 0.10 {
		target = 0.10
	}
	if target > 0.90 {
		target = 0.90
	}

	// Estimate how many blobs we might need for the remaining delta.
	// Each blob is ~ area of a small room; tune blob area a bit from your room mins.
	cur := coverage(L)
	if cur >= target {
		return
	}

	avgBlob := max(9, (p.RoomWMin*p.RoomHMin)/4) // rough, fast, small-ish blobs
	needTiles := int((target - cur) * float64(L.W*L.H))
	estBlobs := needTiles / avgBlob
	if estBlobs < 1 {
		estBlobs = 1
	}

	// Respect FillerRoomsMax as a soft cap, but allow some headroom so coverage is achievable.
	budget := max(p.FillerRoomsMax, estBlobs)
	maxAttempts := max(50, budget*6) // generous attempt budget

	cw := corridorHalf(p.CorridorWidth)
	attempts := 0
	for coverage(L) < target && attempts < maxAttempts {
		attempts++

		sx, sy, ok := farthestEmpty(L, 4) // margin from border
		if !ok {
			break
		}

		// carve a compact blob based on min room sizes, with a little randomness
		rw := clampi(p.RoomWMin-2, 4, 9) + rng.IntN(3) // 4–11
		rh := clampi(p.RoomHMin-2, 4, 9) + rng.IntN(3)
		rx := max(1, sx-rw/2)
		ry := max(1, sy-rh/2)
		rw = min(rw, L.W-rx-1)
		rh = min(rh, L.H-ry-1)
		if rw < 3 || rh < 3 {
			continue // skip too-small rooms
		}
		carveRect(L, rx, ry, rw, rh)

		// Find best connection point on the new room's edge
		roomCenterX := rx + rw/2
		roomCenterY := ry + rh/2
		roomEdgeX, roomEdgeY := findRoomEdgeConnectionPoint(L, rx, ry, rw, rh)

		// Connect to closest existing walkable using room edge point
		if tx, ty, ok2 := nearestWalkable(L, roomEdgeX, roomEdgeY); ok2 {
			carveL(L, roomEdgeX, roomEdgeY, tx, ty, cw)
			// Verify connection was successful by checking if room is now connected
			comps := floodComponents(L)
			if len(comps) > 1 {
				// Connection failed, try direct path from room center
				if tx2, ty2, ok3 := nearestWalkable(L, roomCenterX, roomCenterY); ok3 {
					carveL(L, roomCenterX, roomCenterY, tx2, ty2, cw)
				}
			}
		}
	}
}

func coverage(L *Level) float64 {
	walk, total := 0, L.W*L.H
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if L.Tiles[y][x].IsWalkable {
				walk++
			}
		}
	}
	return float64(walk) / float64(total)
}

func farthestEmpty(L *Level, border int) (int, int, bool) {
	bestD, bx, by := -1, 0, 0
	for y := border; y < L.H-border; y++ {
		for x := border; x < L.W-border; x++ {
			if L.Tiles[y][x].IsWalkable {
				continue
			}
			// distance to nearest walkable (cheap Manhattan scan window)
			d := nearestWalkableDist(L, x, y, 12) // 12-tile local window
			if d > bestD {
				bestD, bx, by = d, x, y
			}
		}
	}
	if bestD < 0 {
		return 0, 0, false
	}
	return bx, by, true
}

func nearestWalkableDist(L *Level, x, y, window int) int {
	best := 1 << 30
	x0, x1 := max(0, x-window), min(L.W-1, x+window)
	y0, y1 := max(0, y-window), min(L.H-1, y+window)
	for yy := y0; yy <= y1; yy++ {
		for xx := x0; xx <= x1; xx++ {
			if L.Tiles[yy][xx].IsWalkable {
				d := abs(xx-x) + abs(yy-y)
				if d < best {
					best = d
				}
			}
		}
	}
	if best == 1<<30 {
		return -1
	}
	return best
}

// nearestWalkable finds the nearest walkable tile to the given point.
// It first searches locally (within a window) for efficiency, then falls back to full search.
func nearestWalkable(L *Level, x, y int) (int, int, bool) {
	// First try local window search (much faster)
	window := 20
	x0, x1 := max(0, x-window), min(L.W-1, x+window)
	y0, y1 := max(0, y-window), min(L.H-1, y+window)

	best := 1 << 30
	bx, by := 0, 0
	found := false

	for yy := y0; yy <= y1; yy++ {
		for xx := x0; xx <= x1; xx++ {
			if !L.Tiles[yy][xx].IsWalkable {
				continue
			}
			d := (xx-x)*(xx-x) + (yy-y)*(yy-y)
			if d < best {
				best, bx, by = d, xx, yy
				found = true
			}
		}
	}

	if found {
		return bx, by, true
	}

	// Fall back to full grid search if nothing found locally
	best = 1 << 30
	for yy := 0; yy < L.H; yy++ {
		for xx := 0; xx < L.W; xx++ {
			if !L.Tiles[yy][xx].IsWalkable {
				continue
			}
			d := (xx-x)*(xx-x) + (yy-y)*(yy-y)
			if d < best {
				best, bx, by = d, xx, yy
			}
		}
	}
	if best == 1<<30 {
		return 0, 0, false
	}
	return bx, by, true
}

// findRoomEdgeConnectionPoint finds the best point on a room's edge to connect from.
// It tries to find a point that faces the nearest existing walkable area.
func findRoomEdgeConnectionPoint(L *Level, rx, ry, rw, rh int) (int, int) {
	cx := rx + rw/2
	cy := ry + rh/2

	// Try to find nearest walkable to determine best edge
	if tx, ty, ok := nearestWalkableLocal(L, cx, cy, 15); ok {
		dx := tx - cx
		dy := ty - cy

		// Determine which edge to use based on direction to nearest walkable
		if abs(dx) > abs(dy) {
			// Horizontal connection
			if dx > 0 {
				// Connect from right edge
				return rx + rw - 1, clamp(ty, ry, ry+rh-1)
			} else {
				// Connect from left edge
				return rx, clamp(ty, ry, ry+rh-1)
			}
		} else {
			// Vertical connection
			if dy > 0 {
				// Connect from bottom edge
				return clamp(tx, rx, rx+rw-1), ry + rh - 1
			} else {
				// Connect from top edge
				return clamp(tx, rx, rx+rw-1), ry
			}
		}
	}

	// Fallback: use center of closest edge to level center
	levelCenterX := L.W / 2
	levelCenterY := L.H / 2
	if cx < levelCenterX {
		return rx + rw - 1, cy // right edge
	} else if cx > levelCenterX {
		return rx, cy // left edge
	} else if cy < levelCenterY {
		return cx, ry + rh - 1 // bottom edge
	} else {
		return cx, ry // top edge
	}
}

// nearestWalkableLocal searches for the nearest walkable tile within a local window.
func nearestWalkableLocal(L *Level, x, y, window int) (int, int, bool) {
	x0, x1 := max(0, x-window), min(L.W-1, x+window)
	y0, y1 := max(0, y-window), min(L.H-1, y+window)

	best := 1 << 30
	bx, by := 0, 0
	found := false

	for yy := y0; yy <= y1; yy++ {
		for xx := x0; xx <= x1; xx++ {
			if !L.Tiles[yy][xx].IsWalkable {
				continue
			}
			d := (xx-x)*(xx-x) + (yy-y)*(yy-y)
			if d < best {
				best, bx, by = d, xx, yy
				found = true
			}
		}
	}

	return bx, by, found
}

func carveRect(L *Level, x, y, w, h int) {
	x = max(1, x)
	y = max(1, y)
	for yy := y; yy < y+h && yy < L.H-1; yy++ {
		for xx := x; xx < x+w && xx < L.W-1; xx++ {
			L.Tiles[yy][xx].IsWalkable = true
		}
	}
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func bspRegions(w, h, depth int, rng *rand.Rand) []rect {
	regs := []rect{{0, 0, w, h}}
	minW := currentParams.RoomWMax + currentParams.CorridorWidth*2
	minH := currentParams.RoomHMax + currentParams.CorridorWidth*2
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
	maxRoom := max(p.RoomWMax, p.RoomHMax)
	minDist := maxRoom + 2
	minR2 := minDist * minDist
	marginX := 1 + p.RoomWMax/2
	marginY := 1 + p.RoomHMax/2
	var pts []image.Point
	for _, r := range rs {
		if rng.IntN(100) < 20 {
			continue
		} // skip some regions for asymmetry
		minX := r.X + marginX
		maxX := r.X + r.W - marginX - 1
		minY := r.Y + marginY
		maxY := r.Y + r.H - marginY - 1
		if minX > maxX || minY > maxY {
			continue
		}
		pt := image.Point{minX + rng.IntN(maxX-minX+1), minY + rng.IntN(maxY-minY+1)}
		ok := true
		for _, q := range pts {
			dx, dy := q.X-pt.X, q.Y-pt.Y
			if dx*dx+dy*dy < minR2 {
				ok = false
				break
			}
		}
		if ok {
			pts = append(pts, pt)
		}
	}
	// supplement globally
	attempts := 0
	for len(pts) < p.RoomCountMin && attempts < 1000 {
		cx := marginX + rng.IntN(p.Width-2*marginX)
		cy := marginY + rng.IntN(p.Height-2*marginY)
		pt := image.Point{cx, cy}
		ok := true
		for _, q := range pts {
			dx, dy := q.X-pt.X, q.Y-pt.Y
			if dx*dx+dy*dy < minR2 {
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

var centerIndexToRoom = map[int]rect{} // add near top (file-scope)

func growRooms(L *Level, centers []image.Point, p GenParams, rng *rand.Rand) []rect {
	rooms := make([]rect, len(centers))
	margin := 1
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
		centerIndexToRoom[i] = r // <— store mapping
		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				L.Tiles[yy][xx].IsWalkable = true
				if yy >= 0 && yy < len(roomMask) && xx >= 0 && xx < len(roomMask[yy]) {
					roomMask[yy][xx] = true
				}
			}
		}
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

func corridorHalf(w int) int {
	if w <= 1 {
		return 0
	}
	return w / 2
}

func carveCorridors(L *Level, es []edge, W int) {
	half := corridorHalf(W)
	for _, e := range es {
		aCenter := currentCenters[e.A]
		bCenter := currentCenters[e.B]
		ra := centerIndexToRoom[e.A]
		rb := centerIndexToRoom[e.B]
		sa := doorToward(ra, bCenter)
		sb := doorToward(rb, aCenter)
		// Keep a single-tile doorway at the room edge, then carve the corridor.
		if countOverlapL(L, sa.X, sa.Y, sb.X, sb.Y, true) <= countOverlapL(L, sa.X, sa.Y, sb.X, sb.Y, false) {
			carveLOrdered(L, sa.X, sa.Y, sb.X, sb.Y, half, true)
		} else {
			carveLOrdered(L, sa.X, sa.Y, sb.X, sb.Y, half, false)
		}
	}
}

func carveLOrdered(L *Level, x1, y1, x2, y2, half int, horizFirst bool) {
	if horizFirst {
		carveCorridorSegment(L, x1, y1, x2, y1, half)
		carveCorridorSegment(L, x2, y1, x2, y2, half)
		carveDisk(L, x2, y1, half)
		return
	}
	carveCorridorSegment(L, x1, y1, x1, y2, half)
	carveCorridorSegment(L, x1, y2, x2, y2, half)
	carveDisk(L, x1, y2, half)
}

func countOverlapL(L *Level, x1, y1, x2, y2 int, horizFirst bool) int {
	count := 0
	if horizFirst {
		count += countOverlapSegment(L, x1, y1, x2, y1)
		count += countOverlapSegment(L, x2, y1, x2, y2)
		return count
	}
	count += countOverlapSegment(L, x1, y1, x1, y2)
	count += countOverlapSegment(L, x1, y2, x2, y2)
	return count
}

func countOverlapSegment(L *Level, x1, y1, x2, y2 int) int {
	count := 0
	if x1 == x2 {
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for y := y1; y <= y2; y++ {
			if inBounds(L, x1, y) && len(roomMask) > 0 && !roomMask[y][x1] {
				if corridorMask[y][x1] {
					count++
				}
			}
		}
		return count
	}
	if y1 == y2 {
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		for x := x1; x <= x2; x++ {
			if inBounds(L, x, y1) && len(roomMask) > 0 && !roomMask[y1][x] {
				if corridorMask[y1][x] {
					count++
				}
			}
		}
	}
	return count
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
					if len(corridorMask) > 0 {
						corridorMask[y][nx] = true
					}
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
					if len(corridorMask) > 0 {
						corridorMask[ny][x] = true
					}
				}
			}
		}
	}
}

func corridorWidthAt(L *Level, x, y int) (wx, wy int) {
	l := 0
	for i := x; i >= 0 && L.Tiles[y][i].IsWalkable; i-- {
		l++
	}
	r := 0
	for i := x; i < L.W && L.Tiles[y][i].IsWalkable; i++ {
		r++
	}
	u := 0
	for j := y; j >= 0 && L.Tiles[j][x].IsWalkable; j-- {
		u++
	}
	d := 0
	for j := y; j < L.H && L.Tiles[j][x].IsWalkable; j++ {
		d++
	}
	return l + r - 1, u + d - 1
}

func widenPinches(L *Level, W int) {
	target := max(1, W)
	for y := 1; y < L.H-1; y++ {
		for x := 1; x < L.W-1; x++ {
			if !L.Tiles[y][x].IsWalkable {
				continue
			}
			wx, wy := corridorWidthAt(L, x, y)
			if wx < target {
				if !L.Tiles[y][x-1].IsWalkable {
					L.Tiles[y][x-1].IsWalkable = true
					continue
				}
				if !L.Tiles[y][x+1].IsWalkable {
					L.Tiles[y][x+1].IsWalkable = true
					continue
				}
			}
			if wy < target {
				if !L.Tiles[y-1][x].IsWalkable {
					L.Tiles[y-1][x].IsWalkable = true
					continue
				}
				if !L.Tiles[y+1][x].IsWalkable {
					L.Tiles[y+1][x].IsWalkable = true
					continue
				}
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

// placeDoorsByThroats places doors on "throat" tiles that separate two walkable regions.
func placeDoorsByThroats(L *Level, p GenParams, rng *rand.Rand) {
	// Try to load wall sprite sheet for door sprites
	var wss *sprites.WallSpriteSheet
	var err error
	if p.WallFlavor != "" {
		wss, err = sprites.LoadWallSpriteSheet(p.WallFlavor)
	}
	if err != nil || wss == nil {
		// Fall back to normal flavor
		wss, err = sprites.LoadWallSpriteSheet("normal")
		if err != nil {
			return // Can't place doors without door sprites
		}
	}

	limit := 18
	visited := make([][]int, L.H)
	for y := range visited {
		visited[y] = make([]int, L.W)
	}
	stamp := 0

	for y := 1; y < L.H-1; y++ {
		for x := 1; x < L.W-1; x++ {
			t := L.Tile(x, y)
			if t == nil || !t.IsWalkable || t.HasTag(tiles.TagDoor) {
				continue
			}

			up := L.Tiles[y-1][x].IsWalkable
			down := L.Tiles[y+1][x].IsWalkable
			left := L.Tiles[y][x-1].IsWalkable
			right := L.Tiles[y][x+1].IsWalkable

			walkableNeighbors := 0
			if up {
				walkableNeighbors++
			}
			if down {
				walkableNeighbors++
			}
			if left {
				walkableNeighbors++
			}
			if right {
				walkableNeighbors++
			}
			if walkableNeighbors != 2 {
				continue
			}

			var orient string
			var a, b image.Point
			if up && down && !left && !right {
				orient = "nw"
				a = image.Point{x, y - 1}
				b = image.Point{x, y + 1}
			} else if left && right && !up && !down {
				orient = "ne"
				a = image.Point{x - 1, y}
				b = image.Point{x + 1, y}
			} else {
				continue
			}

			stamp++
			if throatRegionsConnect(L, a, b, x, y, limit, visited, stamp) {
				continue
			}

			placeDoorAt(L, x, y, orient, wss, p.WallFlavor, p.DoorLockChance, rng)
		}
	}
}

func inBounds(L *Level, x, y int) bool {
	return x >= 0 && y >= 0 && x < L.W && y < L.H
}

func sealWalkableEdges(L *Level) {
	if L.W == 0 || L.H == 0 {
		return
	}
	for x := 0; x < L.W; x++ {
		L.Tiles[0][x].IsWalkable = false
		L.Tiles[L.H-1][x].IsWalkable = false
	}
	for y := 0; y < L.H; y++ {
		L.Tiles[y][0].IsWalkable = false
		L.Tiles[y][L.W-1].IsWalkable = false
	}
}

func buildWallMask(L *Level) [][]bool {
	walls := make([][]bool, L.H)
	for y := range walls {
		walls[y] = make([]bool, L.W)
	}
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if L.Tiles[y][x].IsWalkable {
				continue
			}
			if hasWalkableNeighbor(L, x, y) {
				walls[y][x] = true
			}
		}
	}
	return walls
}

func hasWalkableNeighbor(L *Level, x, y int) bool {
	for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := x+d.X, y+d.Y
		if nx < 0 || ny < 0 || nx >= L.W || ny >= L.H {
			continue
		}
		if L.Tiles[ny][nx].IsWalkable {
			return true
		}
	}
	return false
}

func placeDoorsFromValidatedThroats(L *Level, p GenParams) {
	info := BuildThroatDebug(L, 18, 12)
	if len(info.ValidThroats) == 0 {
		return
	}

	var wss *sprites.WallSpriteSheet
	var err error
	if p.WallFlavor != "" {
		wss, err = sprites.LoadWallSpriteSheet(p.WallFlavor)
	}
	if err != nil || wss == nil {
		wss, err = sprites.LoadWallSpriteSheet("normal")
		if err != nil {
			return
		}
	}

	for _, t := range info.ValidThroats {
		if (t.IsRoomA && !t.IsRoomB) || (t.IsRoomB && !t.IsRoomA) {
			placeUnlockedDoorAt(L, t.X, t.Y, t.Orient, wss, p.WallFlavor)
		}
	}
}

func placeUnlockedDoorAt(L *Level, x, y int, orientation string, wss *sprites.WallSpriteSheet, flavor string) {
	tile := L.Tile(x, y)
	if tile == nil || tile.HasTag(tiles.TagDoor) {
		return
	}

	var doorImg *ebiten.Image
	var doorID string
	if orientation == "nw" {
		doorImg = wss.UnlockedDoorNW
		doorID = flavor + "_door_unlocked_nw"
	} else {
		doorImg = wss.UnlockedDoorNE
		doorID = flavor + "_door_unlocked_ne"
	}
	if doorImg == nil {
		return
	}

	tile.AddSpriteByID(doorID, doorImg)
	tile.SetTag(tiles.TagDoor)
	tile.DoorSpriteID = doorID
	tile.DoorState = 1 // unlocked/open (metadata only; do not change walkability)
}

func throatRegionsConnect(L *Level, a, b image.Point, blockX, blockY, limit int, visited [][]int, stamp int) bool {
	queue := []image.Point{a}
	visited[a.Y][a.X] = stamp
	steps := 0
	for len(queue) > 0 && steps < limit {
		p := queue[0]
		queue = queue[1:]
		steps++
		if p.X == b.X && p.Y == b.Y {
			return true
		}
		for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := p.X+d.X, p.Y+d.Y
			if nx == blockX && ny == blockY {
				continue
			}
			if !inBounds(L, nx, ny) || visited[ny][nx] == stamp {
				continue
			}
			if !L.Tiles[ny][nx].IsWalkable {
				continue
			}
			visited[ny][nx] = stamp
			queue = append(queue, image.Point{nx, ny})
		}
	}
	return false
}


// checkDoorPosition checks if a position is valid for placing a door.
// Returns orientation and validity.
// A valid door position must:
// 1. Be in a corridor (walkable tile) or can replace a wall between walkable areas
// 2. Have walkable tiles on both sides (opposite directions)
// 3. Have at least 2 adjacent walkable tiles total
// 4. Have appropriate orientation based on the corridor/wall direction
func checkDoorPosition(L *Level, x, y int) (string, bool) {
	tile := L.Tile(x, y)
	if tile == nil {
		return "", false
	}

	// Don't place doors on tiles that already have doors
	if tile.HasTag(tiles.TagDoor) {
		return "", false
	}

	// Check all four cardinal directions to find positions with walkable tiles on both sides
	dirs := []struct {
		dx, dy int
		orient string
	}{
		{1, 0, "ne"},  // Right/East - NE orientation
		{-1, 0, "nw"}, // Left/West - NW orientation
		{0, 1, "ne"},  // Down/South - NE orientation
		{0, -1, "nw"}, // Up/North - NW orientation
	}

	for _, d := range dirs {
		// Check if there are walkable tiles on both sides of this position
		side1x, side1y := x+d.dx, y+d.dy // Side 1 (one direction)
		side2x, side2y := x-d.dx, y-d.dy // Side 2 (opposite direction)

		// Bounds check
		if side1x < 0 || side1y < 0 || side1x >= L.W || side1y >= L.H {
			continue
		}
		if side2x < 0 || side2y < 0 || side2x >= L.W || side2y >= L.H {
			continue
		}

		tile1 := L.Tile(side1x, side1y)
		tile2 := L.Tile(side2x, side2y)

		if tile1 == nil || tile2 == nil {
			continue
		}

		// Both sides must have walkable tiles (door connects two walkable areas)
		if !tile1.IsWalkable || !tile2.IsWalkable {
			continue
		}

		// Count total adjacent walkable tiles (should have at least 2)
		adjacentCount := 0
		adjDirs := []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
		for _, adj := range adjDirs {
			ax, ay := x+adj.X, y+adj.Y
			if ax >= 0 && ay >= 0 && ax < L.W && ay < L.H {
				at := L.Tile(ax, ay)
				if at != nil && at.IsWalkable && !at.HasTag(tiles.TagDoor) {
					adjacentCount++
				}
			}
		}

		if adjacentCount < 2 {
			continue // Need at least 2 adjacent walkable tiles
		}

		// Door can be placed on either:
		// 1. A walkable tile (in a corridor) - common case
		// 2. A non-walkable tile (wall) if it has walkable tiles on both sides
		// Both cases require walkable tiles on both sides (already verified above)

		// If it's a wall tile, we'll make it walkable when placing the door
		// This position is valid for a door
		return d.orient, true
	}

	return "", false
}

// placeDoorAt places a door at the specified tile position with the given orientation.
func placeDoorAt(L *Level, x, y int, orientation string, wss *sprites.WallSpriteSheet, flavor string, lockChance float64, rng *rand.Rand) {
	tile := L.Tile(x, y)
	if tile == nil {
		return // Invalid tile
	}

	if lockChance < 0 {
		lockChance = 0
	}
	if lockChance > 1 {
		lockChance = 1
	}
	locked := rng.Float64() < lockChance

	var doorImg *ebiten.Image
	var doorID string

	// Choose door sprite based on orientation and lock state
	if orientation == "nw" {
		if locked {
			doorImg = wss.LockedDoorNW
			doorID = flavor + "_door_locked_nw"
			tile.DoorState = 3 // locked
		} else {
			doorImg = wss.LockedDoorNW
			doorID = flavor + "_door_locked_nw"
			tile.DoorState = 2 // closed but unlocked
		}
	} else { // orientation == "ne"
		if locked {
			doorImg = wss.LockedDoorNE
			doorID = flavor + "_door_locked_ne"
			tile.DoorState = 3 // locked
		} else {
			doorImg = wss.LockedDoorNE
			doorID = flavor + "_door_locked_ne"
			tile.DoorState = 2 // closed but unlocked
		}
	}

	// Place door sprite and mark tile
	tile.AddSpriteByID(doorID, doorImg)
	tile.SetTag(tiles.TagDoor)
	tile.DoorSpriteID = doorID

	// Closed/locked doors block movement.
	tile.IsWalkable = false
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
	first := true
	for {
		if !first && !L.Tiles[y0][x0].IsWalkable {
			return false
		}
		first = false
		if x0 == x1 && y0 == y1 {
			break
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
	// For connectivity check, treat closed doors as walkable so we can verify
	// all areas are reachable (doors will be opened by player)
	for comps := floodComponentsIgnoreDoors(L); len(comps) > 1; comps = floodComponentsIgnoreDoors(L) {
		a, b := comps[0], comps[1]
		best := 1 << 30
		var pa, pb image.Point
		for _, p := range a {
			for _, q := range b {
				d := (p.X-q.X)*(p.X-q.X) + (p.Y-q.Y)*(p.Y-q.Y)
				if d < best {
					best, pa, pb = d, p, q
				}
			}
		}
		half := max(1, currentParams.CorridorWidth/2)
		carveL(L, pa.X, pa.Y, pb.X, pb.Y, half)
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
			// Check actual walkability (doors block movement)
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

// floodComponentsIgnoreDoors treats closed doors as walkable for connectivity checking.
// This ensures all rooms are connected even if doors block movement initially.
func floodComponentsIgnoreDoors(L *Level) [][]image.Point {
	visited := make([][]bool, L.H)
	for i := range visited {
		visited[i] = make([]bool, L.W)
	}
	dirs := []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	var comps [][]image.Point
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			tile := L.Tiles[y][x]
			// Treat closed doors as walkable for connectivity check
			isReachable := tile.IsWalkable || (tile.HasTag(tiles.TagDoor) && tile.DoorState >= 2)
			if visited[y][x] || !isReachable {
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
					if nx >= 0 && ny >= 0 && nx < L.W && ny < L.H && !visited[ny][nx] {
						nt := L.Tiles[ny][nx]
						isReachable = nt.IsWalkable || (nt.HasTag(tiles.TagDoor) && nt.DoorState >= 2)
						if isReachable {
							visited[ny][nx] = true
							queue = append(queue, image.Point{nx, ny})
						}
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

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}

func doorToward(r rect, target image.Point) image.Point {
	cx, cy := r.X+r.W/2, r.Y+r.H/2
	dx, dy := target.X-cx, target.Y-cy
	if abs(dx) > abs(dy) {
		if dx > 0 {
			return image.Point{r.X + r.W - 1, clamp(target.Y, r.Y, r.Y+r.H-1)}
		}
		return image.Point{r.X, clamp(target.Y, r.Y, r.Y+r.H-1)}
	}
	if dy > 0 {
		return image.Point{clamp(target.X, r.X, r.X+r.W-1), r.Y + r.H - 1}
	}
	return image.Point{clamp(target.X, r.X, r.X+r.W-1), r.Y}
}

func carveL(L *Level, x1, y1, x2, y2, half int) {
	// choose shorter-first axis
	if abs(x2-x1) < abs(y2-y1) {
		carveCorridorSegment(L, x1, y1, x1, y2, half)
		carveCorridorSegment(L, x1, y2, x2, y2, half)
		// small corner fillet
		carveDisk(L, x1, y2, half)
	} else {
		carveCorridorSegment(L, x1, y1, x2, y1, half)
		carveCorridorSegment(L, x2, y1, x2, y2, half)
		carveDisk(L, x2, y1, half)
	}
}

func carveDisk(L *Level, cx, cy, r int) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if x >= 0 && y >= 0 && x < L.W && y < L.H && dx*dx+dy*dy <= r*r {
				L.Tiles[y][x].IsWalkable = true
				if len(corridorMask) > 0 {
					corridorMask[y][x] = true
				}
			}
		}
	}
}
