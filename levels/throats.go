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

	regionSizes, regionAllLinear := buildRegions(L, info.RegionIDs)
	regionIsRoom := map[int]bool{}
	for id, size := range regionSizes {
		allLinear := regionAllLinear[id]
		// Corridor regions remain corridors even if large, if they are purely linear.
		if allLinear {
			regionIsRoom[id] = false
			continue
		}
		regionIsRoom[id] = size >= minRegionSize || !allLinear
	}
	info.RegionIsRoom = regionIsRoom
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
			if !corridorMask[y][x] {
				continue
			}

			up := L.Tiles[y-1][x].IsWalkable
			down := L.Tiles[y+1][x].IsWalkable
			left := L.Tiles[y][x-1].IsWalkable
			right := L.Tiles[y][x+1].IsWalkable

			walkableCount := 0
			if up {
				walkableCount++
			}
			if down {
				walkableCount++
			}
			if left {
				walkableCount++
			}
			if right {
				walkableCount++
			}
			if walkableCount != 2 {
				continue
			}

			var a, b image.Point
			orient := ""
			if up && down && !left && !right {
				a = image.Point{X: x, Y: y - 1}
				b = image.Point{X: x, Y: y + 1}
				orient = "nw"
			} else if left && right && !up && !down {
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

			ra := info.RegionIDs[a.Y][a.X]
			rb := info.RegionIDs[b.Y][b.X]
			if ra == 0 || rb == 0 || ra == rb {
				info.Invalid = append(info.Invalid, image.Point{X: x, Y: y})
				continue
			}
			if !(regionIsRoom[ra] || regionSizes[ra] >= minRegionSize) || !(regionIsRoom[rb] || regionSizes[rb] >= minRegionSize) {
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
				IsRoomA:   regionIsRoom[ra],
				IsRoomB:   regionIsRoom[rb],
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
