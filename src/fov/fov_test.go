package fov

import (
	"dungeoneer/levels"
	"dungeoneer/tiles"
	"math"
	"testing"
)

func TestObjectPoints(t *testing.T) {
	o := Object{
		Walls: []Line{
			{X1: 0, Y1: 0, X2: 1, Y2: 0},
			{X1: 1, Y1: 0, X2: 1, Y2: 1},
		},
	}
	pts := o.Points()
	if len(pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(pts))
	}
	if pts[0][0] != 1 || pts[0][1] != 0 {
		t.Errorf("expected pt 0 to be (1, 0), got %v", pts[0])
	}
	if pts[1][0] != 1 || pts[1][1] != 1 {
		t.Errorf("expected pt 1 to be (1, 1), got %v", pts[1])
	}
	if pts[2][0] != 0 || pts[2][1] != 0 {
		t.Errorf("expected pt 2 to be (0, 0), got %v", pts[2])
	}
}

func TestNewRay(t *testing.T) {
	ray := NewRay(0, 0, 10, 0)
	if ray.X1 != 0 || ray.Y1 != 0 {
		t.Errorf("expected origin (0,0)")
	}
	if math.Abs(ray.X2-10) > 1e-9 || math.Abs(ray.Y2-0) > 1e-9 {
		t.Errorf("expected end (10,0), got (%v, %v)", ray.X2, ray.Y2)
	}

	ray2 := NewRay(0, 0, 10, math.Pi/2)
	if math.Abs(ray2.X2-0) > 1e-9 || math.Abs(ray2.Y2-10) > 1e-9 {
		t.Errorf("expected end (0,10), got (%v, %v)", ray2.X2, ray2.Y2)
	}
}

func TestIntersection(t *testing.T) {
	// Intersection
	l1 := Line{X1: 0, Y1: 0, X2: 10, Y2: 10}
	l2 := Line{X1: 0, Y1: 10, X2: 10, Y2: 0}
	hx, hy, ok := Intersection(l1, l2)
	if !ok {
		t.Errorf("expected intersection")
	}
	if math.Abs(hx-5) > 1e-9 || math.Abs(hy-5) > 1e-9 {
		t.Errorf("expected (5, 5), got (%v, %v)", hx, hy)
	}

	// Parallel lines
	l3 := Line{X1: 0, Y1: 0, X2: 10, Y2: 10}
	l4 := Line{X1: 0, Y1: 1, X2: 10, Y2: 11}
	_, _, ok = Intersection(l3, l4)
	if ok {
		t.Errorf("expected no intersection for parallel lines")
	}

	// Segments not intersecting
	l5 := Line{X1: 0, Y1: 0, X2: 1, Y2: 1}
	l6 := Line{X1: 5, Y1: 5, X2: 6, Y2: 6}
	_, _, ok = Intersection(l5, l6)
	if ok {
		t.Errorf("expected no intersection for disjoint segments")
	}
}

func TestWorldToScreen(t *testing.T) {
	sx, sy := worldToScreen(0, 0, 0, 0, 1.0, 0, 0, 64)
	if sx != 0 || sy != 0 {
		t.Errorf("expected (0,0), got (%v, %v)", sx, sy)
	}

	sx, sy = worldToScreen(1, 0, 0, 0, 1.0, 0, 0, 64)
	if sx != 32 || sy != 16 {
		t.Errorf("expected (32, 16), got (%v, %v)", sx, sy)
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

func TestIsWall(t *testing.T) {
	level := createTestLevel(3, 3)
	
	// Open tile
	if isWall(1, 1, level) {
		t.Errorf("expected (1,1) not to be wall")
	}

	// Blocked tile
	level.Tiles[1][1].IsWalkable = false
	if !isWall(1, 1, level) {
		t.Errorf("expected (1,1) to be wall")
	}

	// Out of bounds
	if !isWall(-1, 0, level) {
		t.Errorf("expected (-1,0) to be wall")
	}
	if !isWall(3, 0, level) {
		t.Errorf("expected (3,0) to be wall")
	}
}

func TestTraceLineToTiles(t *testing.T) {
	pts := TraceLineToTiles(0.5, 0.5, 2.5, 0.5)
	if len(pts) != 3 {
		t.Fatalf("expected 3 points for horizontal line, got %v", pts)
	}
	
	pts = TraceLineToTiles(0.5, 0.5, 0.5, 2.5)
	if len(pts) != 3 {
		t.Fatalf("expected 3 points for vertical line, got %v", pts)
	}

	// Diagonal crossing should include supercover corners
	pts = TraceLineToTiles(0.5, 0.5, 1.5, 1.5)
	// (0,0), (1,0) or (0,1), (1,1) - should be 3 points
	if len(pts) < 3 {
		t.Errorf("expected at least 3 points for diagonal supercover, got %d", len(pts))
	}
}

func TestLevelToWalls(t *testing.T) {
	level := createTestLevel(3, 3)
	
	// Single wall in center
	level.Tiles[1][1].IsWalkable = false
	walls := LevelToWalls(level)
	
	// Should have 4 faces + 4 diagonal gaps (NW/SE/NE/SW) = 8 lines
	if len(walls) == 0 {
		t.Fatalf("expected walls")
	}

	// Check if isOpen works correctly
	if !isOpen(level, 1, 0) {
		t.Errorf("expected (1,0) to be open")
	}
	if isOpen(level, 1, 1) {
		t.Errorf("expected (1,1) not to be open")
	}
}

func TestRayCastingAndCache(t *testing.T) {
	level := createTestLevel(3, 3)
	level.Tiles[1][1].IsWalkable = false
	walls := LevelToWalls(level)

	InvalidateCache()

	rays1 := RayCasting(0.5, 0.5, walls, level)
	if len(rays1) == 0 {
		t.Fatalf("expected rays")
	}

	// Calling again with same position should hit cache
	rays2 := RayCasting(0.5, 0.5, walls, level)
	if len(rays1) != len(rays2) {
		t.Errorf("expected same length from cache")
	}

	// Calling with different position should recompute
	rays3 := RayCasting(2.5, 2.5, walls, level)
	if len(rays3) == 0 {
		t.Fatalf("expected rays")
	}

	// Calling inside a wall (snapping logic)
	rays4 := RayCasting(1.1, 1.1, walls, level)
	if len(rays4) == 0 {
		t.Fatalf("expected rays")
	}
}

func TestResizeShadowBuffer(t *testing.T) {
	ResizeShadowBuffer(800, 600)
	if shadowImage.Bounds().Dx() != 800 {
		t.Errorf("expected width 800")
	}

	ResizeShadowBuffer(800, 600) // Call again to hit the early return (or no-op)
}

func TestBuildShadowImage(t *testing.T) {
	ResizeShadowBuffer(100, 100)
	rays := []Line{
		{X1: 0, Y1: 0, X2: 10, Y2: 0},
		{X1: 0, Y1: 0, X2: 10, Y2: 10},
		{X1: 0, Y1: 0, X2: 0, Y2: 10},
	}
	
	img := BuildShadowImage(rays, 50, 50, 0, 0, 1.0, 50, 50, 64)
	if img == nil {
		t.Errorf("expected image")
	}

	// Test with fewer than 3 rays
	img = BuildShadowImage([]Line{}, 50, 50, 0, 0, 1.0, 50, 50, 64)
	if img == nil {
		t.Errorf("expected image")
	}
}

func TestDebugDrawRaysAndWalls(t *testing.T) {
	ResizeShadowBuffer(100, 100)
	rays := []Line{
		{X1: 0, Y1: 0, X2: 10, Y2: 0},
	}
	walls := []Line{
		{X1: 5, Y1: 5, X2: 5, Y2: 10},
	}
	
	// Just verify they don't panic
	DebugDrawRays(shadowImage, rays, 50, 50, 0, 0, 1.0, 50, 50, 64)
	DebugDrawWalls(shadowImage, walls, 0, 0, 1.0, 50, 50, 64)
}
