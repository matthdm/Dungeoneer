package collision

import (
	"dungeoneer/levels"
	"math"
	"testing"
)

func testLevelWithWalls(walls ...[2]int) *levels.Level {
	level := levels.NewEmptyLevel(5, 5)
	for y := 0; y < level.H; y++ {
		for x := 0; x < level.W; x++ {
			level.Tiles[y][x].IsWalkable = true
		}
	}
	for _, w := range walls {
		level.Tiles[w[1]][w[0]].IsWalkable = false
	}
	return level
}

func testLevelWithWall(wallX, wallY int) *levels.Level {
	return testLevelWithWalls([2]int{wallX, wallY})
}

func TestCollidesWithMapUsesPlayerSpriteAnchor(t *testing.T) {
	level := testLevelWithWall(2, 2)
	box := Box{X: 1.6, Y: 1.15, Width: 0.55, Height: 0.8}

	if CollidesWithMapWithAnchor(level, box, SpriteAnchor{}) {
		t.Fatal("zero anchor should not move the box into the wall tile")
	}
	if !CollidesWithMap(level, box) {
		t.Fatal("player sprite anchor should move the box into the wall tile")
	}
}

func TestPredictAndClipStopsBeforeAnchoredWallCollision(t *testing.T) {
	level := testLevelWithWall(2, 2)
	box := Box{X: 1.3, Y: 1.15, Width: 0.55, Height: 0.8}

	finalBox, hitX, hitY := PredictAndClip(level, box, 1.0, 0)
	if !hitX {
		t.Fatal("PredictAndClip should report an X collision")
	}
	if hitY {
		t.Fatal("PredictAndClip should not report a Y collision for horizontal movement")
	}
	if CollidesWithMap(level, finalBox) {
		t.Fatal("final clipped box should not collide with the wall")
	}
}

func TestPredictAndClip(t *testing.T) {
	levelSingle := testLevelWithWall(2, 2)
	levelBoth := testLevelWithWalls([2]int{2, 2}, [2]int{1, 3})

	tests := []struct {
		name     string
		level    *levels.Level
		box      Box
		dx, dy   float64
		wantHitX bool
		wantHitY bool
	}{
		{
			name:     "Successful movement with no collisions (X only)",
			level:    levelSingle,
			box:      Box{X: 1.0, Y: 1.0, Width: 0.5, Height: 0.5},
			dx:       0.4, dy: 0.0,
			wantHitX: false, wantHitY: false,
		},
		{
			name:     "Successful movement with no collisions (Y only)",
			level:    levelSingle,
			box:      Box{X: 1.0, Y: 1.0, Width: 0.5, Height: 0.5},
			dx:       0.0, dy: 0.4,
			wantHitX: false, wantHitY: false,
		},
		{
			name:     "Collision on X axis moving right",
			level:    levelSingle,
			box:      Box{X: 1.3, Y: 1.15, Width: 0.55, Height: 0.8},
			dx:       1.0, dy: 0.0,
			wantHitX: true, wantHitY: false,
		},
		{
			name:     "Collision on X axis moving left",
			level:    levelSingle,
			box:      Box{X: 3.1, Y: 1.15, Width: 0.55, Height: 0.8},
			dx:       -1.0, dy: 0.0,
			wantHitX: true, wantHitY: false,
		},
		{
			name:     "Collision on Y axis moving down",
			level:    levelSingle,
			box:      Box{X: 1.79, Y: 0.5, Width: 0.55, Height: 0.8},
			dx:       0.0, dy: 1.0,
			wantHitX: false, wantHitY: true,
		},
		{
			name:     "Collision on Y axis moving up",
			level:    levelSingle,
			box:      Box{X: 1.79, Y: 3.0, Width: 0.55, Height: 0.8},
			dx:       0.0, dy: -1.5,
			wantHitX: false, wantHitY: true,
		},
		{
			name:     "Diagonal movement - only X collides",
			level:    levelSingle,
			box:      Box{X: 1.3, Y: 1.6, Width: 0.55, Height: 0.8},
			dx:       1.0, dy: 0.5,
			wantHitX: true, wantHitY: false,
		},
		{
			name:     "Diagonal movement - only Y collides",
			level:    levelSingle,
			box:      Box{X: 1.3, Y: 0.5, Width: 0.55, Height: 0.8},
			dx:       1.0, dy: 1.5,
			wantHitX: false, wantHitY: true,
		},
		{
			name:     "Diagonal movement - both collide",
			level:    levelBoth,
			box:      Box{X: 1.3, Y: 1.6, Width: 0.55, Height: 0.8},
			dx:       1.0, dy: 1.5,
			wantHitX: true, wantHitY: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBox, gotHitX, gotHitY := PredictAndClip(tt.level, tt.box, tt.dx, tt.dy)
			if gotHitX != tt.wantHitX {
				t.Errorf("PredictAndClip() gotHitX = %v, want %v", gotHitX, tt.wantHitX)
			}
			if gotHitY != tt.wantHitY {
				t.Errorf("PredictAndClip() gotHitY = %v, want %v", gotHitY, tt.wantHitY)
			}
			if tt.wantHitX {
				if CollidesWithMap(tt.level, gotBox) {
					t.Errorf("PredictAndClip() resulting box collides with map: %+v", gotBox)
				}
			} else if tt.dx != 0 {
				expectedX := tt.box.X + tt.dx
				if math.Abs(gotBox.X-expectedX) > 1e-9 {
					t.Errorf("PredictAndClip() gotBox.X = %f, want %f", gotBox.X, expectedX)
				}
			}
			if tt.wantHitY {
				if CollidesWithMap(tt.level, gotBox) {
					t.Errorf("PredictAndClip() resulting box collides with map: %+v", gotBox)
				}
			} else if tt.dy != 0 {
				expectedY := tt.box.Y + tt.dy
				if math.Abs(gotBox.Y-expectedY) > 1e-9 {
					t.Errorf("PredictAndClip() gotBox.Y = %f, want %f", gotBox.Y, expectedY)
				}
			}
		})
	}
}

func TestCollidesWithMapWithAnchor_Boundaries(t *testing.T) {
	level := levels.NewEmptyLevel(3, 3)
	for y := 0; y < level.H; y++ {
		for x := 0; x < level.W; x++ {
			level.Tiles[y][x].IsWalkable = true
		}
	}

	tests := []struct {
		name   string
		box    Box
		anchor SpriteAnchor
		want   bool
	}{
		{
			name:   "Completely inside walkable level, zero anchor",
			box:    Box{X: 1.5, Y: 1.5, Width: 0.8, Height: 0.8},
			anchor: SpriteAnchor{OffsetX: 0, OffsetY: 0},
			want:   false,
		},
		{
			name:   "Out of bounds left (x < 0)",
			box:    Box{X: 0.1, Y: 1.5, Width: 0.8, Height: 0.8},
			anchor: SpriteAnchor{OffsetX: -0.5, OffsetY: 0},
			want:   true,
		},
		{
			name:   "Out of bounds top (y < 0)",
			box:    Box{X: 1.5, Y: 0.1, Width: 0.8, Height: 0.8},
			anchor: SpriteAnchor{OffsetX: 0, OffsetY: -0.5},
			want:   true,
		},
		{
			name:   "Out of bounds right (x >= level.W)",
			box:    Box{X: 2.9, Y: 1.5, Width: 0.8, Height: 0.8},
			anchor: SpriteAnchor{OffsetX: 0.5, OffsetY: 0},
			want:   true,
		},
		{
			name:   "Out of bounds bottom (y >= level.H)",
			box:    Box{X: 1.5, Y: 2.9, Width: 0.8, Height: 0.8},
			anchor: SpriteAnchor{OffsetX: 0, OffsetY: 0.5},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollidesWithMapWithAnchor(level, tt.box, tt.anchor)
			if got != tt.want {
				t.Errorf("CollidesWithMapWithAnchor() = %v, want %v", got, tt.want)
			}
		})
	}
}
