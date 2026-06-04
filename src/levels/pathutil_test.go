package levels

import (
	"dungeoneer/tiles"
	"testing"
)

func TestFindSpawnAndExit(t *testing.T) {
	l := NewEmptyLevel(20, 20)
	// Create a simple corridor
	for x := 2; x < 18; x++ {
		l.Tiles[10][x] = &tiles.Tile{IsWalkable: true}
	}
	
	sx, sy, ex, ey := FindSpawnAndExit(l)
	if sx == ex && sy == ey {
		t.Errorf("spawn and exit should not be the same")
	}
	if !l.IsWalkable(sx, sy) {
		t.Errorf("spawn should be walkable")
	}
	if !l.IsWalkable(ex, ey) {
		t.Errorf("exit should be walkable")
	}
	
	// Check the distance is reasonable for this linear corridor
	dx := ex - sx
	if dx < 0 {
		dx = -dx
	}
	if dx < 10 {
		t.Errorf("expected spawn and exit to be far apart, got dx = %d", dx)
	}
}

func TestFindSpawnAndExit_Small(t *testing.T) {
	l := NewEmptyLevel(5, 5)
	l.Tiles[2][2] = &tiles.Tile{IsWalkable: true}
	l.Tiles[2][3] = &tiles.Tile{IsWalkable: true}
	
	sx, sy, ex, ey := FindSpawnAndExit(l)
	if !l.IsWalkable(sx, sy) || !l.IsWalkable(ex, ey) {
		t.Errorf("expected walkable spawn and exit")
	}
}

func TestFindFarthestWalkable(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	l.Tiles[1][1] = &tiles.Tile{IsWalkable: true}
	l.Tiles[1][2] = &tiles.Tile{IsWalkable: true}
	l.Tiles[1][3] = &tiles.Tile{IsWalkable: true}
	
	fx, fy := FindFarthestWalkable(l, 1, 1)
	if fx != 3 || fy != 1 {
		t.Errorf("expected 3,1 farthest, got %d,%d", fx, fy)
	}
}

func TestFindSpawnPoint(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	l.Tiles[5][5] = &tiles.Tile{IsWalkable: true}
	
	sx, sy := FindSpawnPoint(l)
	if sx != 5 || sy != 5 {
		t.Errorf("expected 5,5 spawn, got %d,%d", sx, sy)
	}
}

func TestBfsDistMap(t *testing.T) {
	l := NewEmptyLevel(5, 5)
	l.Tiles[2][2] = &tiles.Tile{IsWalkable: true}
	l.Tiles[2][3] = &tiles.Tile{IsWalkable: true}
	
	distMap := bfsDistMap(l, 2, 2, l.IsPassable)
	if distMap[2][2] != 0 {
		t.Errorf("expected distance 0 at start")
	}
	if distMap[2][3] != 1 {
		t.Errorf("expected distance 1")
	}
	if distMap[1][1] != -1 {
		t.Errorf("expected -1 for unreachable")
	}
}
