package tiles

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tile represents a space with an x,y coordinate within a Level. Any number of
// sprites may be added to a Tile.
type Tile struct {
	Sprites []*ebiten.Image
}

// AddSprite adds a sprite to the Tile.
func (t *Tile) AddSprite(s *ebiten.Image) {
	t.Sprites = append(t.Sprites, s)
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
func (t *Tile) RemoveSprites() {
	//t.Sprites = nil
	t.Sprites = []*ebiten.Image{}

}

// Draw draws the Tile on the screen using the provided options.
func (t *Tile) Draw(screen *ebiten.Image, options *ebiten.DrawImageOptions) {
	for _, s := range t.Sprites {
		screen.DrawImage(s, options)
	}
}
