package levels

import (
	"dungeoneer/tiles"
	"testing"
)

func TestContains(t *testing.T) {
	r := &Room{X: 0, Y: 0, W: 10, H: 10}
	if !r.Contains(0, 0) {
		t.Errorf("expected 0,0 in bounds")
	}
	if r.Contains(10, 10) {
		t.Errorf("expected 10,10 out of bounds")
	}
	if r.Contains(-1, -1) {
		t.Errorf("expected -1,-1 out of bounds")
	}
}

func TestTags(t *testing.T) {
	r := &Room{Tags: []RoomTag{TagSpawn, "test"}}
	if !r.HasTag(TagSpawn) {
		t.Errorf("expected spawn tag")
	}
	if r.HasTag(TagBossArena) {
		t.Errorf("expected no boss tag")
	}
	r.AddTag(TagBossArena)
	if !r.HasTag(TagBossArena) {
		t.Errorf("expected boss tag")
	}
	if r.PrimaryTag() != TagSpawn {
		t.Errorf("expected spawn as primary tag")
	}
}

func TestRoomsByTag(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	r1 := Room{Tags: []RoomTag{TagSpawn}}
	r2 := Room{Tags: []RoomTag{TagBossArena}}
	r3 := Room{Tags: []RoomTag{TagSpawn, TagBossArena}}
	l.Rooms = []Room{r1, r2, r3}
	
	spawns := RoomsByTag(l.Rooms, TagSpawn)
	if len(spawns) != 2 {
		t.Errorf("expected 2 spawn rooms, got %d", len(spawns))
	}
}

func TestRoomAt(t *testing.T) {
	l := NewEmptyLevel(20, 20)
	r := Room{X: 5, Y: 5, W: 5, H: 5}
	l.Rooms = []Room{r}
	
	found := l.RoomAt(6, 6)
	if found == nil || found.X != r.X {
		t.Errorf("expected room at 6,6")
	}
	if l.RoomAt(0, 0) != nil {
		t.Errorf("expected no room at 0,0")
	}
}

func TestIsWalkable(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	l.Tiles[1][1] = &tiles.Tile{IsWalkable: true}
	l.Tiles[2][2] = &tiles.Tile{IsWalkable: false}
	
	if !l.IsWalkable(1, 1) {
		t.Errorf("expected floor to be walkable")
	}
	if l.IsWalkable(2, 2) {
		t.Errorf("expected wall not walkable")
	}
	if l.IsWalkable(15, 15) {
		t.Errorf("out of bounds should not be walkable")
	}
}

func TestIsPassable(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	l.Tiles[1][1] = &tiles.Tile{IsWalkable: true}
	l.Tiles[2][2] = &tiles.Tile{IsWalkable: false}
	
	door := &tiles.Tile{IsWalkable: false, Tags: tiles.TagDoor}
	l.Tiles[3][3] = door
	
	if !l.IsPassable(1, 1) {
		t.Errorf("floor should be passable")
	}
	if l.IsPassable(2, 2) {
		t.Errorf("wall should not be passable")
	}
	// closed door passable for some pathfinding? The method says yes for doors
	if !l.IsPassable(3, 3) {
		t.Errorf("door should be passable")
	}
}

func TestEntityManagement(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	l.AddEntity(PlacedEntity{X: 1, Y: 1, Type: "player1"})
	if len(l.Entities) != 1 {
		t.Errorf("entity not added correctly")
	}
	l.RemoveEntityAt(1, 1, "player1", "")
	if len(l.Entities) != 0 {
		t.Errorf("entity should have been removed")
	}
}

func TestLevelSizes(t *testing.T) {
	l := NewEmptyLevel(10, 20)
	w, h := l.Size()
	if w != 10 || h != 20 {
		t.Errorf("size incorrect")
	}
	if l.Tile(5, 5) == nil {
		t.Errorf("empty level should have tiles")
	}
}

func TestClassifyRoomSize(t *testing.T) {
	if ClassifyRoomSize(3, 3) != RoomSmall {
		t.Errorf("expected small")
	}
	if ClassifyRoomSize(10, 10) != RoomLarge {
		t.Errorf("expected large")
	}
	if ClassifyRoomSize(15, 15) != RoomLarge {
		t.Errorf("expected large")
	}
}

func TestLevelGenerators(t *testing.T) {
	// These rely on spritesheets being loadable, which they should be if they are embedded.
	l, err := NewDungeonLevel()
	if err != nil {
		t.Fatalf("failed to create dungeon level: %v", err)
	}
	if l.W != 64 || l.H != 64 {
		t.Errorf("expected 64x64 dungeon level")
	}
	
	l, err = NewMazeLevel()
	if err != nil {
		t.Fatalf("failed to create maze level: %v", err)
	}
	if l == nil {
		t.Errorf("expected maze level")
	}
	
	l, err = NewForestLevel()
	if err != nil {
		t.Fatalf("failed to create forest level: %v", err)
	}
	if l.W != 64 || l.H != 64 {
		t.Errorf("expected 64x64 forest level")
	}
}
