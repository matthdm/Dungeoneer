package levels

import (
	"dungeoneer/tiles"
	"testing"
)

func TestTagRooms(t *testing.T) {
	l := NewEmptyLevel(20, 20)
	
	// Room 0: Spawn room (0,0, 5,5)
	r0 := Room{X: 1, Y: 1, W: 3, H: 3, Size: RoomSmall, Index: 0}
	
	// Room 1: Exit room (10,10, 5,5)
	r1 := Room{X: 10, Y: 10, W: 3, H: 3, Size: RoomSmall, Index: 1}
	
	// Room 2: Dead end room (1, 10, 5, 5)
	r2 := Room{X: 1, Y: 10, W: 3, H: 3, Size: RoomSmall, Index: 2}
	
	// Room 3: Crossroads (10, 1, 5, 5)
	r3 := Room{X: 10, Y: 1, W: 3, H: 3, Size: RoomSmall, Index: 3}
	
	// Room 4: Boss arena
	r4 := Room{X: 15, Y: 15, W: 4, H: 4, Size: RoomLarge, Index: 4}
	
	l.Rooms = []Room{r0, r1, r2, r3, r4}
	
	// make all room tiles walkable
	for _, r := range l.Rooms {
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				l.Tiles[y][x] = &tiles.Tile{IsWalkable: true}
			}
		}
	}
	
	// make some borders walkable to simulate exits
	// room 2 has 1 exit
	l.Tiles[10][4] = &tiles.Tile{IsWalkable: true} // right wall exit
	// room 3 has 3 exits
	l.Tiles[4][11] = &tiles.Tile{IsWalkable: true} // bottom
	l.Tiles[2][9] = &tiles.Tile{IsWalkable: true} // left
	l.Tiles[2][13] = &tiles.Tile{IsWalkable: true} // right
	
	TagRooms(l, 2, 2, 11, 11, true)
	
	if !l.Rooms[0].HasTag(TagSpawn) {
		t.Errorf("expected spawn tag")
	}
	if !l.Rooms[1].HasTag(TagExit) {
		t.Errorf("expected exit tag")
	}
	if !l.Rooms[2].HasTag(TagTreasure) {
		// dead end usually gets treasure
		t.Errorf("expected treasure tag for r2")
	}
	if !l.Rooms[3].HasTag(TagCrossroads) {
		t.Errorf("expected crossroads tag for r3")
	}
	if !l.Rooms[4].HasTag(TagBossArena) {
		t.Errorf("expected boss arena tag for r4")
	}
	
	// Test without boss floor
	TagRooms(l, 2, 2, 11, 11, false)
	if l.Rooms[4].HasTag(TagBossArena) {
		t.Errorf("expected no boss arena tag")
	}
}
