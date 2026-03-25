package levels

import "math"

// TagRooms performs a post-generation pass that assigns semantic tags to rooms.
// spawnX/Y and exitX/Y are the player spawn and exit tile coordinates.
// bossFloor indicates whether this is the final floor with a boss arena.
func TagRooms(l *Level, spawnX, spawnY, exitX, exitY int, bossFloor bool) {
	if len(l.Rooms) == 0 {
		return
	}

	// Reset all tags.
	for i := range l.Rooms {
		l.Rooms[i].Tags = nil
	}

	// Step 1: Tag spawn and exit rooms.
	spawnRoom := l.RoomAt(spawnX, spawnY)
	exitRoom := l.RoomAt(exitX, exitY)
	if spawnRoom != nil {
		spawnRoom.AddTag(TagSpawn)
		spawnRoom.AddTag(TagCleared)
	}
	if exitRoom != nil {
		exitRoom.AddTag(TagExit)
	}

	// Step 2: Classify room connectivity (count walkable exits per room).
	exits := countRoomExits(l)

	// Step 3: Tag dead-ends and crossroads from connectivity.
	for i := range l.Rooms {
		r := &l.Rooms[i]
		if r.HasTag(TagSpawn) || r.HasTag(TagExit) {
			continue
		}
		e := exits[r.Index]
		if e <= 1 {
			r.AddTag(TagDeadEnd)
			r.AddTag(TagOptional)
		} else if e >= 3 {
			r.AddTag(TagCrossroads)
		}
	}

	// Step 4: Boss arena on boss floors (largest room).
	if bossFloor {
		best := largestUntaggedRoom(l, spawnRoom, exitRoom)
		if best != nil {
			best.Tags = nil // clear any connectivity tags
			best.AddTag(TagBossArena)
			best.AddTag(TagDecorated)
		}
	}

	// Step 5: Pick sanctuary — 1 per floor, Medium+ room, farthest from spawn.
	// Prefer optional/dead-end rooms so the sanctuary is off the beaten path.
	pickSanctuary(l, spawnX, spawnY)

	// Step 6: Pick treasure rooms — 1-2 per floor from dead-ends / small optionals.
	pickTreasure(l, spawnRoom, exitRoom)

	// Step 7: All remaining untagged rooms → common.
	for i := range l.Rooms {
		r := &l.Rooms[i]
		if r.PrimaryTag() == TagCommon && !r.HasTag(TagDeadEnd) && !r.HasTag(TagCrossroads) {
			r.AddTag(TagCommon)
		}
	}
}

// countRoomExits counts how many walkable border tiles connect each room to
// tiles outside the room (corridors or other rooms). This approximates the
// number of doorways/passages into the room.
func countRoomExits(l *Level) map[int]int {
	exits := make(map[int]int)
	for i := range l.Rooms {
		r := &l.Rooms[i]
		count := 0
		// Scan the room's border tiles.
		for x := r.X; x < r.X+r.W; x++ {
			for y := r.Y; y < r.Y+r.H; y++ {
				if x != r.X && x != r.X+r.W-1 && y != r.Y && y != r.Y+r.H-1 {
					continue // interior tile, skip
				}
				if !l.IsWalkable(x, y) && !l.IsPassable(x, y) {
					continue
				}
				// Check cardinal neighbors outside the room.
				for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					nx, ny := x+d[0], y+d[1]
					if r.Contains(nx, ny) {
						continue
					}
					if nx < 0 || ny < 0 || nx >= l.W || ny >= l.H {
						continue
					}
					if l.IsWalkable(nx, ny) || l.IsPassable(nx, ny) {
						count++
					}
				}
			}
		}
		exits[r.Index] = count
	}
	return exits
}

// largestUntaggedRoom returns the largest room that isn't spawn/exit.
func largestUntaggedRoom(l *Level, skip ...*Room) *Room {
	skipIdx := make(map[int]bool)
	for _, r := range skip {
		if r != nil {
			skipIdx[r.Index] = true
		}
	}
	var best *Room
	bestArea := 0
	for i := range l.Rooms {
		r := &l.Rooms[i]
		if skipIdx[r.Index] {
			continue
		}
		area := r.W * r.H
		if area > bestArea {
			bestArea = area
			best = r
		}
	}
	return best
}

// pickSanctuary selects one Medium+ room farthest from spawn and marks it
// as a sanctuary. Prefers rooms already tagged optional/dead_end.
func pickSanctuary(l *Level, spawnX, spawnY int) {
	var best *Room
	bestDist := -1.0
	bestOptional := false

	for i := range l.Rooms {
		r := &l.Rooms[i]
		// Skip rooms that already have a primary role.
		pt := r.PrimaryTag()
		if pt == TagSpawn || pt == TagExit || pt == TagBossArena {
			continue
		}
		// Sanctuary needs Medium+ room for space.
		if r.Size == RoomSmall {
			continue
		}
		dx := float64(r.CenterX - spawnX)
		dy := float64(r.CenterY - spawnY)
		dist := math.Sqrt(dx*dx + dy*dy)
		isOpt := r.HasTag(TagOptional) || r.HasTag(TagDeadEnd)

		// Prefer optional rooms; among same category, prefer farthest.
		if best == nil ||
			(isOpt && !bestOptional) ||
			(isOpt == bestOptional && dist > bestDist) {
			best = r
			bestDist = dist
			bestOptional = isOpt
		}
	}

	if best != nil {
		// Clear connectivity tags and set sanctuary.
		best.Tags = filterModifiers(best.Tags)
		best.AddTag(TagSanctuary)
		best.AddTag(TagCleared)
		best.AddTag(TagDecorated)
	}
}

// pickTreasure selects 1-2 dead-end or optional rooms as treasure rooms.
func pickTreasure(l *Level, skip ...*Room) {
	skipIdx := make(map[int]bool)
	for _, r := range skip {
		if r != nil {
			skipIdx[r.Index] = true
		}
	}

	count := 0
	maxTreasure := 2

	// First pass: dead-end rooms make the best treasure rooms.
	for i := range l.Rooms {
		if count >= maxTreasure {
			break
		}
		r := &l.Rooms[i]
		if skipIdx[r.Index] {
			continue
		}
		pt := r.PrimaryTag()
		if pt != TagCommon && pt != TagDeadEnd {
			continue
		}
		if !r.HasTag(TagDeadEnd) {
			continue
		}
		r.Tags = filterModifiers(r.Tags)
		r.AddTag(TagTreasure)
		r.AddTag(TagLoot)
		r.AddTag(TagDecorated)
		r.AddTag(TagOptional)
		count++
	}

	// Second pass: any small optional room.
	for i := range l.Rooms {
		if count >= maxTreasure {
			break
		}
		r := &l.Rooms[i]
		if skipIdx[r.Index] {
			continue
		}
		pt := r.PrimaryTag()
		if pt != TagCommon {
			continue
		}
		if r.Size != RoomSmall {
			continue
		}
		r.Tags = filterModifiers(r.Tags)
		r.AddTag(TagTreasure)
		r.AddTag(TagLoot)
		r.AddTag(TagDecorated)
		r.AddTag(TagOptional)
		count++
	}
}

// filterModifiers returns only the modifier tags, stripping primary roles.
func filterModifiers(tags []RoomTag) []RoomTag {
	var out []RoomTag
	for _, t := range tags {
		switch t {
		case TagDecorated, TagLoot, TagCleared, TagDark, TagOptional:
			out = append(out, t)
		}
	}
	return out
}
