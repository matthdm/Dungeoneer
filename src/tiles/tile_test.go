package tiles

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestTile_AddSpriteByID(t *testing.T) {
	tile := &Tile{}
	img := ebiten.NewImage(10, 10)

	tile.AddSpriteByID("test-sprite", img)

	if len(tile.Sprites) != 1 {
		t.Fatalf("expected 1 sprite, got %d", len(tile.Sprites))
	}
	if tile.Sprites[0].ID != "test-sprite" {
		t.Errorf("expected sprite ID 'test-sprite', got '%s'", tile.Sprites[0].ID)
	}
	if tile.Sprites[0].Image != img {
		t.Errorf("expected sprite image to match")
	}
}

func TestTile_RemoveSprite(t *testing.T) {
	// Case 1: Sprites is nil
	{
		tile := &Tile{Sprites: nil}
		tile.RemoveSprite(nil) // should not panic
	}

	// Case 2: Sprites is empty slice
	{
		tile := &Tile{Sprites: []SpriteRef{}}
		tile.RemoveSprite(nil) // should not panic
	}

	// Case 3: Sprites has items
	{
		tile := &Tile{}
		img1 := ebiten.NewImage(5, 5)
		img2 := ebiten.NewImage(5, 5)
		tile.AddSpriteByID("s1", img1)
		tile.AddSpriteByID("s2", img2)

		tile.RemoveSprite(nil) // it removes the last sprite regardless of argument

		if len(tile.Sprites) != 1 {
			t.Fatalf("expected 1 sprite left, got %d", len(tile.Sprites))
		}
		if tile.Sprites[0].ID != "s1" {
			t.Errorf("expected remaining sprite to be 's1', got '%s'", tile.Sprites[0].ID)
		}
	}
}

func TestTile_RemoveLastSprite(t *testing.T) {
	// Case 1: Empty Sprites
	{
		tile := &Tile{}
		tile.RemoveLastSprite()
		if len(tile.Sprites) != 0 {
			t.Errorf("expected 0 sprites, got %d", len(tile.Sprites))
		}
	}

	// Case 2: Sprites exist
	{
		tile := &Tile{}
		img := ebiten.NewImage(2, 2)
		tile.AddSpriteByID("s1", img)
		tile.RemoveLastSprite()
		if len(tile.Sprites) != 0 {
			t.Errorf("expected 0 sprites after removal, got %d", len(tile.Sprites))
		}
	}
}

func TestTile_ClearSprites(t *testing.T) {
	tile := &Tile{}
	img := ebiten.NewImage(2, 2)
	tile.AddSpriteByID("s1", img)
	tile.AddSpriteByID("s2", img)

	tile.ClearSprites()

	if tile.Sprites == nil || len(tile.Sprites) != 0 {
		t.Errorf("expected Sprites to be empty slice, got %v", tile.Sprites)
	}
}

func TestTile_HasSpriteID(t *testing.T) {
	tile := &Tile{}
	img := ebiten.NewImage(2, 2)
	tile.AddSpriteByID("s1", img)
	tile.AddSpriteByID("s2", img)

	tests := []struct {
		id   string
		want bool
	}{
		{"s1", true},
		{"s2", true},
		{"s3", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := tile.HasSpriteID(tt.id); got != tt.want {
				t.Errorf("HasSpriteID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestTile_Tags(t *testing.T) {
	tile := &Tile{}

	// Initial state
	if tile.HasTag(TagDashLane) {
		t.Error("expected tile not to have TagDashLane initially")
	}

	// Set tag
	tile.SetTag(TagDashLane)
	if !tile.HasTag(TagDashLane) {
		t.Error("expected tile to have TagDashLane")
	}

	// Set another tag
	tile.SetTag(TagGrappleAnchor)
	if !tile.HasTag(TagDashLane) || !tile.HasTag(TagGrappleAnchor) {
		t.Error("expected tile to have both TagDashLane and TagGrappleAnchor")
	}

	if tile.HasTag(TagDoor) {
		t.Error("expected tile not to have TagDoor")
	}
}

func TestTile_Draw(t *testing.T) {
	tile := &Tile{}
	img1 := ebiten.NewImage(4, 4)
	img2 := ebiten.NewImage(4, 4)
	tile.AddSpriteByID("s1", img1)
	tile.AddSpriteByID("s2", img2)

	screen := ebiten.NewImage(10, 10)
	opts := &ebiten.DrawImageOptions{}

	// Draw should execute screen.DrawImage for each sprite without panicking
	tile.Draw(screen, opts)
}
