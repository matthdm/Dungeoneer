package tiles

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	TagNone          = 0
	TagDashLane      = 1 << 0
	TagGrappleAnchor = 1 << 1
	TagDoor          = 1 << 2
)

// Tile represents a space with an x,y coordinate within a Level. Any number of
// sprites may be added to a Tile.
type Tile struct {
	Sprites    []SpriteRef
	IsWalkable bool
	Tags       uint8
	// Door state: 0 = no door, 1 = open, 2 = closed, 3 = locked
	DoorState  uint8
	DoorSpriteID string // ID of door sprite (for removal/changing)
}

type SpriteRef struct {
	ID    string
	Image *ebiten.Image
}

// AddSprite adds a sprite to the Tile.
func (t *Tile) AddSpriteByID(id string, img *ebiten.Image) {
	t.Sprites = append(t.Sprites, SpriteRef{ID: id, Image: img})
}

func (t *Tile) RemoveSprite(s *ebiten.Image) {
	if t.Sprites == nil {

		fmt.Println("Slice is empty, cannot remove element")
	}
	if len(t.Sprites) > 0 {
		t.Sprites = t.Sprites[:len(t.Sprites)-1]
		fmt.Println("Slice after removing last element:", t.Sprites)
	} else {
		fmt.Println("Slice is empty, cannot remove element")
	}

}

// RemoveLastSprite removes the last sprite on the tile
func (t *Tile) RemoveLastSprite() {
	if len(t.Sprites) > 0 {
		t.Sprites = t.Sprites[:len(t.Sprites)-1]
	}
}

// RemoveSprites clears all sprites from the tile
func (t *Tile) ClearSprites() {
	t.Sprites = []SpriteRef{}

}

// Draw draws the Tile on the screen using the provided options.
func (t *Tile) Draw(screen *ebiten.Image, options *ebiten.DrawImageOptions) {
	for _, s := range t.Sprites {
		screen.DrawImage(s.Image, options)
	}
}

func (t *Tile) HasSpriteID(id string) bool {
	for _, s := range t.Sprites {
		if s.ID == id {
			return true
		}
	}
	return false
}

func (t *Tile) SetTag(tag uint8) {
	t.Tags |= tag
}

func (t *Tile) HasTag(tag uint8) bool {
	return t.Tags&tag != 0
}
