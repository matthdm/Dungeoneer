package levels

import (
	"dungeoneer/sprites"
	"testing"
)

func TestGenerate64x64(t *testing.T) {
	ss, err := sprites.LoadSpriteSheet(64)
	if err != nil {
		t.Skip("spritesheet missing")
	}
	// Generate a level with depth 1
	p := GenParams{Seed: 1}
	l := Generate64x64(p)
	if l == nil {
		t.Fatalf("expected generated level")
	}
	if l.W != 64 || l.H != 64 {
		t.Errorf("expected 64x64 level")
	}
	if len(l.Rooms) == 0 {
		t.Errorf("expected generated rooms")
	}
	
	// Test depth scaling
	p2 := GenParams{Seed: 10, RoomCountMin: 20}
	l2 := Generate64x64(p2)
	if l2 == nil {
		t.Fatalf("expected generated level")
	}
	
	// Test door locked
	p3 := GenParams{Seed: 100, DoorLockChance: 1.0, CoverageTarget: 0.1, FillerRoomsMax: 0}
	l5 := Generate64x64(p3)
	if l5 == nil {
		t.Fatalf("expected generated level")
	}
	
	// Test Blank level creation
	l3 := CreateNewBlankLevel(20, 20, 64, ss)
	if l3.W != 20 {
		t.Errorf("expected 20")
	}
	
	l4 := CreateNewBlankLevelWithFloor(20, 20, 64, "Floor", ss.Floor)
	if l4.W != 20 {
		t.Errorf("expected 20")
	}
}
