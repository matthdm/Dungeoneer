package sprites

import (
	"dungeoneer/constants"
	"testing"
)

func TestLoadSpriteSheet(t *testing.T) {
	// The images package loads the spritesheet PNG from embedded fs
	ss, err := LoadSpriteSheet(constants.DefaultTileSize)
	if err != nil {
		t.Fatalf("failed to load spritesheet: %v", err)
	}
	
	if ss == nil {
		t.Fatalf("expected non-nil SpriteSheet")
	}
	
	if ss.Void == nil {
		t.Errorf("expected non-nil Void sprite")
	}
	
	if ss.StairsDecending == nil {
		t.Errorf("expected non-nil StairsDecending sprite")
	}
}

func TestLoadWallSpriteSheet(t *testing.T) {
	for _, flavor := range WallFlavors {
		wss, err := LoadWallSpriteSheet(flavor)
		if err != nil {
			t.Errorf("failed to load wall sprite sheet for flavor %s: %v", flavor, err)
		}
		
		if wss == nil {
			t.Errorf("expected non-nil WallSpriteSheet for flavor %s", flavor)
		} else if wss.Wall == nil {
			t.Errorf("expected non-nil Wall sprite for flavor %s", flavor)
		}
	}
	
	// Test default flavor
	wss, err := LoadWallSpriteSheet("unknown_flavor_should_default")
	if err != nil {
		t.Errorf("failed to load fallback wall sprite sheet: %v", err)
	}
	
	if wss == nil {
		t.Errorf("expected non-nil fallback WallSpriteSheet")
	}
}
