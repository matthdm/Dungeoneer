package levels

import (
	"dungeoneer/sprites"
	"testing"
)

func TestGenerateForgottenSanctuary(t *testing.T) {
	// Create an empty SpriteSheet with dummy values to prevent nil pointer crashes during testing
	ss := &sprites.SpriteSheet{}
	
	w, h := 32, 32
	level := GenerateForgottenSanctuary(w, h, ss, 42)

	if level == nil {
		t.Fatal("GenerateForgottenSanctuary returned nil")
	}

	if level.W != w || level.H != h {
		t.Errorf("Expected dimensions %dx%d, got %dx%d", w, h, level.W, level.H)
	}

	// Verify perimeter walls
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			tile := level.Tiles[y][x]
			isPerimeter := x == 0 || x == w-1 || y == 0 || y == h-1
			if isPerimeter && tile.IsWalkable {
				t.Errorf("Tile at (%d, %d) is on perimeter but is walkable", x, y)
			}
		}
	}

	// Verify diagonal water path blocks walkability (except for the bridge)
	for i := 2; i < w-2; i++ {
		isBridge := (i >= w/2-2 && i <= w/2+1)
		tile := level.Tiles[i][i]
		
		if !isBridge && tile.IsWalkable {
			t.Errorf("Water tile at (%d, %d) should not be walkable", i, i)
		}
	}
}
