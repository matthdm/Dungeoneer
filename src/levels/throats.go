package levels

import "image"

type ThroatDebugInfo struct {
	Valid        []image.Point
	Invalid      []image.Point
	ValidThroats []ThroatInfo
	RegionIDs    [][]int
	RegionIsRoom map[int]bool
}

type ThroatInfo struct {
	X, Y      int
	Orient    string
	RegionA   int
	RegionB   int
	IsRoomA   bool
	IsRoomB   bool
	NeighborA image.Point
	NeighborB image.Point
}

func BuildThroatDebug(L *Level, floodLimit, minRegionSize int) ThroatDebugInfo {
	info := ThroatDebugInfo{
		Valid:        []image.Point{},
		Invalid:      []image.Point{},
		ValidThroats: []ThroatInfo{},
		RegionIDs:    make([][]int, L.H),
		RegionIsRoom: map[int]bool{},
	}
	for y := range info.RegionIDs {
		info.RegionIDs[y] = make([]int, L.W)
	}
	if L.W == 0 || L.H == 0 {
		return info
	}

	roomMask := make([][]bool, L.H)
	corridorMask := make([][]bool, L.H)
	for y := 0; y < L.H; y++ {
		roomMask[y] = make([]bool, L.W)
		corridorMask[y] = make([]bool, L.W)
		for x := 0; x < L.W; x++ {
			if !L.Tiles[y][x].IsWalkable {
				continue
			}
			n := walkableNeighbors(L, x, y)
			if n >= 3 {
				roomMask[y][x] = true
			}
			if n <= 2 {
				corridorMask[y][x] = true
			}
		}
	}

	roomRegionIDs, roomRegionSizes, roomRegionCount := buildMaskedRegions(L, roomMask)
	corridorRegionIDs, corridorRegionSizes, _ := buildMaskedRegions(L, corridorMask)
	offset := roomRegionCount + 1
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if roomMask[y][x] && roomRegionIDs[y][x] > 0 {
				info.RegionIDs[y][x] = roomRegionIDs[y][x]
				info.RegionIsRoom[roomRegionIDs[y][x]] = true
				continue
			}
			if corridorMask[y][x] && corridorRegionIDs[y][x] > 0 {
				info.RegionIDs[y][x] = corridorRegionIDs[y][x] + offset
				info.RegionIsRoom[corridorRegionIDs[y][x]+offset] = false
			}
		}
	}
	loopMask := corridorLoops(L, corridorMask)

	for y := 1; y < L.H-1; y++ {
		for x := 1; x < L.W-1; x++ {
			if !L.Tiles[y][x].IsWalkable {
				continue
			}
			if isBoundaryAdjacent(L, x, y) {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			up := L.Tiles[y-1][x].IsWalkable
			down := L.Tiles[y+1][x].IsWalkable
			left := L.Tiles[y][x-1].IsWalkable
			right := L.Tiles[y][x+1].IsWalkable

			var a, b image.Point
			orient := ""
			hasNS := up && down
			hasEW := left && right
			if !(hasNS || hasEW) {
				continue
			}
			if hasNS && hasEW {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			if hasNS {
				a = image.Point{X: x, Y: y - 1}
				b = image.Point{X: x, Y: y + 1}
				orient = "nw"
			} else if hasEW {
				a = image.Point{X: x - 1, Y: y}
				b = image.Point{X: x + 1, Y: y}
				orient = "ne"
			} else {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}

			if loopMask[y][x] {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}

			isRoomA := roomMask[a.Y][a.X]
			isRoomB := roomMask[b.Y][b.X]
			if !isRoomA && !corridorMask[a.Y][a.X] {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			if !isRoomB && !corridorMask[b.Y][b.X] {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}

			ra := 0
			rb := 0
			if isRoomA {
				ra = roomRegionIDs[a.Y][a.X]
			} else {
				ra = corridorRegionIDs[a.Y][a.X]
			}
			if isRoomB {
				rb = roomRegionIDs[b.Y][b.X]
			} else {
				rb = corridorRegionIDs[b.Y][b.X]
			}
			if ra == 0 || rb == 0 {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			if isRoomA {
				if roomRegionSizes[ra] < 2 {
					info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
					continue
				}
			} else {
				if corridorRegionSizes[ra] < 2 {
					info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
					continue
				}
			}
			if isRoomB {
				if roomRegionSizes[rb] < 2 {
					info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
					continue
				}
			} else {
				if corridorRegionSizes[rb] < 2 {
					info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
					continue
				}
			}
			if !isRoomA && !isRoomB {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			if isRoomA == isRoomB && ra == rb {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			if reachesWithinLimit(L, a, b, x, y, floodLimit) {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}

			info.Valid = append(info.Valid, image.Point{X: x, Y: y})
			info.ValidThroats = append(info.ValidThroats, ThroatInfo{
				X:         x,
				Y:         y,
				Orient:    orient,
				RegionA:   ra,
				RegionB:   rb,
				IsRoomA:   isRoomA,
				IsRoomB:   isRoomB,
				NeighborA: a,
				NeighborB: b,
			})
		}
	}

	return info
}

func walkableNeighbors(L *Level, x, y int) int {
	count := 0
	for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := x+d.X, y+d.Y
		if nx < 0 || ny < 0 || nx >= L.W || ny >= L.H {
			continue
		}
		if L.Tiles[ny][nx].IsWalkable {
			count++
		}
	}
	return count
}

func isBoundaryAdjacent(L *Level, x, y int) bool {
	return x <= 1 || y <= 1 || x >= L.W-2 || y >= L.H-2
}

func buildMaskedRegions(L *Level, mask [][]bool) ([][]int, map[int]int, int) {
	regionIDs := make([][]int, L.H)
	for y := range regionIDs {
		regionIDs[y] = make([]int, L.W)
	}
	regionSizes := map[int]int{}
	region := 0
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if !mask[y][x] || regionIDs[y][x] != 0 {
				continue
			}
			region++
			queue := []image.Point{{X: x, Y: y}}
			regionIDs[y][x] = region
			size := 0
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				size++
				for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := p.X+d.X, p.Y+d.Y
					if nx < 0 || ny < 0 || nx >= L.W || ny >= L.H {
						continue
					}
					if !mask[ny][nx] || regionIDs[ny][nx] != 0 {
						continue
					}
					regionIDs[ny][nx] = region
					queue = append(queue, image.Point{X: nx, Y: ny})
				}
			}
			regionSizes[region] = size
		}
	}
	return regionIDs, regionSizes, region
}

func buildRegions(L *Level, regionIDs [][]int) (map[int]int, map[int]bool) {
	regionSizes := map[int]int{}
	regionAllLinear := map[int]bool{}
	region := 0
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if !L.Tiles[y][x].IsWalkable || regionIDs[y][x] != 0 {
				continue
			}
			region++
			queue := []image.Point{{X: x, Y: y}}
			regionIDs[y][x] = region
			size := 0
			allLinear := true
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				size++
				if walkableNeighbors(L, p.X, p.Y) >= 3 {
					allLinear = false
				}
				for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := p.X+d.X, p.Y+d.Y
					if nx < 0 || ny < 0 || nx >= L.W || ny >= L.H {
						continue
					}
					if !L.Tiles[ny][nx].IsWalkable || regionIDs[ny][nx] != 0 {
						continue
					}
					regionIDs[ny][nx] = region
					queue = append(queue, image.Point{X: nx, Y: ny})
				}
			}
			regionSizes[region] = size
			regionAllLinear[region] = allLinear
		}
	}
	return regionSizes, regionAllLinear
}

func corridorLoops(L *Level, corridorMask [][]bool) [][]bool {
	loopMask := make([][]bool, L.H)
	visited := make([][]bool, L.H)
	for y := 0; y < L.H; y++ {
		loopMask[y] = make([]bool, L.W)
		visited[y] = make([]bool, L.W)
	}
	for y := 0; y < L.H; y++ {
		for x := 0; x < L.W; x++ {
			if !corridorMask[y][x] || visited[y][x] {
				continue
			}
			queue := []image.Point{{X: x, Y: y}}
			component := []image.Point{}
			endpoints := 0
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				if visited[p.Y][p.X] {
					continue
				}
				visited[p.Y][p.X] = true
				component = append(component, p)

				deg := 0
				for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := p.X+d.X, p.Y+d.Y
					if nx < 0 || ny < 0 || nx >= L.W || ny >= L.H {
						continue
					}
					if corridorMask[ny][nx] {
						deg++
						if !visited[ny][nx] {
							queue = append(queue, image.Point{X: nx, Y: ny})
						}
					}
				}
				if deg <= 1 {
					endpoints++
				}
			}
			if endpoints == 0 {
				for _, p := range component {
					loopMask[p.Y][p.X] = true
				}
			}
		}
	}
	return loopMask
}

func reachesWithinLimit(L *Level, start, target image.Point, blockX, blockY, limit int) bool {
	if limit <= 0 {
		return false
	}
	visited := make([][]bool, L.H)
	for y := range visited {
		visited[y] = make([]bool, L.W)
	}
	queue := []image.Point{start}
	visited[start.Y][start.X] = true
	steps := 0
	for len(queue) > 0 && steps < limit {
		p := queue[0]
		queue = queue[1:]
		steps++
		if p.X == target.X && p.Y == target.Y {
			return true
		}
		for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := p.X+d.X, p.Y+d.Y
			if nx == blockX && ny == blockY {
				continue
			}
			if nx < 0 || ny < 0 || nx >= L.W || ny >= L.H {
				continue
			}
			if visited[ny][nx] || !L.Tiles[ny][nx].IsWalkable {
				continue
			}
			visited[ny][nx] = true
			queue = append(queue, image.Point{X: nx, Y: ny})
		}
	}
	return false
}
