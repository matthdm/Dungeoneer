package pathing

import (
	"dungeoneer/levels"
	"dungeoneer/tiles"
	"testing"
)

func TestAbs(t *testing.T) {
	if abs(-5) != 5 {
		t.Errorf("abs(-5) != 5")
	}
	if abs(5) != 5 {
		t.Errorf("abs(5) != 5")
	}
	if abs(0) != 0 {
		t.Errorf("abs(0) != 0")
	}
}

func TestMax(t *testing.T) {
	if max(1, 2) != 2 {
		t.Errorf("max(1, 2) != 2")
	}
	if max(3, 2) != 3 {
		t.Errorf("max(3, 2) != 3")
	}
	if max(2, 2) != 2 {
		t.Errorf("max(2, 2) != 2")
	}
}

func TestHeuristic(t *testing.T) {
	// dx = 3, dy = 4 -> max(3,4)=4, min(3,4)=3 -> 10*4 + 4*3 = 40 + 12 = 52
	if h := heuristic(0, 0, 3, 4); h != 52 {
		t.Errorf("heuristic(0,0, 3,4) = %v, want 52", h)
	}

	// dx = 5, dy = 1 -> max(5,1)=5, min(5,1)=1 -> 10*5 + 4*1 = 54
	if h := heuristic(10, 10, 5, 9); h != 54 {
		t.Errorf("heuristic(10,10, 5,9) = %v, want 54", h)
	}
}

func TestReconstructPath(t *testing.T) {
	n1 := &Node{X: 1, Y: 1}
	n2 := &Node{X: 2, Y: 1, Parent: n1}
	n3 := &Node{X: 2, Y: 2, Parent: n2}

	path := reconstructPath(n3)
	// Should drop start node n1
	if len(path) != 2 {
		t.Fatalf("expected path length 2, got %v", len(path))
	}
	if path[0].X != 2 || path[0].Y != 1 {
		t.Errorf("expected path[0] to be (2,1)")
	}
	if path[1].X != 2 || path[1].Y != 2 {
		t.Errorf("expected path[1] to be (2,2)")
	}
}

func createTestLevel(w, h int) *levels.Level {
	l := &levels.Level{
		W:     w,
		H:     h,
		Tiles: make([][]*tiles.Tile, h),
	}
	for y := 0; y < h; y++ {
		l.Tiles[y] = make([]*tiles.Tile, w)
		for x := 0; x < w; x++ {
			l.Tiles[y][x] = &tiles.Tile{IsWalkable: true}
		}
	}
	return l
}

func TestAStar_OutOfBounds(t *testing.T) {
	level := createTestLevel(5, 5)
	
	// Goal out of bounds
	if path := AStar(level, 0, 0, -1, 0); path != nil {
		t.Errorf("expected nil path for out of bounds goal")
	}
	if path := AStar(level, 0, 0, 5, 0); path != nil {
		t.Errorf("expected nil path for out of bounds goal")
	}
}

func TestAStar_StraightLine(t *testing.T) {
	level := createTestLevel(5, 5)

	path := AStar(level, 0, 0, 4, 0)
	if len(path) != 4 {
		t.Fatalf("expected path length 4, got %v", len(path))
	}
	for i := 0; i < 4; i++ {
		if path[i].X != i+1 || path[i].Y != 0 {
			t.Errorf("expected path[%d] to be (%d, 0), got (%d, %d)", i, i+1, path[i].X, path[i].Y)
		}
	}
}

func TestAStar_Blocked(t *testing.T) {
	level := createTestLevel(5, 5)
	
	// Wall off the right side
	for y := 0; y < 5; y++ {
		level.Tiles[y][2].IsWalkable = false
	}

	path := AStar(level, 0, 0, 4, 0)
	if path != nil {
		t.Errorf("expected nil path since goal is blocked")
	}
}

func TestAStar_CornerCutting(t *testing.T) {
	level := createTestLevel(5, 5)
	
	// Set up touching corners
	// . x
	// x .
	level.Tiles[0][1].IsWalkable = false
	level.Tiles[1][0].IsWalkable = false

	// From 0,0 to 1,1 diagonal should be blocked due to corner cutting
	path := AStar(level, 0, 0, 1, 1)
	if path != nil {
		t.Errorf("expected nil path, corner cutting should be prevented")
	}
}

func TestAStar_Doors(t *testing.T) {
	level := createTestLevel(5, 5)
	
	// Place a door at 1,0
	level.Tiles[0][1].SetTag(tiles.TagDoor)
	level.Tiles[0][1].DoorState = 2 // Closed

	// Closed door blocks path
	path := AStar(level, 0, 0, 2, 0)
	// It should route around since 0,1 is blocked. It will take a detour
	if len(path) == 2 && path[0].X == 1 && path[0].Y == 0 {
		t.Errorf("path should not go through closed door")
	}

	// Open the door
	level.Tiles[0][1].DoorState = 1 // Open
	path = AStar(level, 0, 0, 2, 0)
	if len(path) != 2 || path[0].X != 1 || path[0].Y != 0 {
		t.Errorf("path should go through open door")
	}
}
