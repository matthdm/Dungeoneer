package collision

import (
	"dungeoneer/levels"
	"testing"
)

func testLevelWithWall(wallX, wallY int) *levels.Level {
	level := levels.NewEmptyLevel(5, 5)
	for y := 0; y < level.H; y++ {
		for x := 0; x < level.W; x++ {
			level.Tiles[y][x].IsWalkable = true
		}
	}
	level.Tiles[wallY][wallX].IsWalkable = false
	return level
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
