package sprites

import (
	"bytes"
	"dungeoneer/images"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteSheet represents a collection of sprite images.
type SpriteSheet struct {
	Void           *ebiten.Image
	Floor          *ebiten.Image
	Wall           *ebiten.Image
	Statue         *ebiten.Image
	Tube           *ebiten.Image
	Crown          *ebiten.Image
	Portal         *ebiten.Image
	RedMan         *ebiten.Image
	BlueMan        *ebiten.Image
	Cursor         *ebiten.Image
	EnemyCursor    *ebiten.Image
	OakBeam        *ebiten.Image
	OakWall        *ebiten.Image
	OakWall2       *ebiten.Image
	OakChunk       *ebiten.Image
	OakWall3       *ebiten.Image
	OakChunkSlim   *ebiten.Image
	OakWall4       *ebiten.Image
	Water          *ebiten.Image
	LockedDoorWest *ebiten.Image
	EnchantedWater *ebiten.Image
	LockedDoorEast *ebiten.Image
	GreyKnight     *ebiten.Image
}

// LoadSpriteSheet loads the embedded SpriteSheet.
func LoadSpriteSheet(tileSize int) (*SpriteSheet, error) {
	img, _, err := image.Decode(bytes.NewReader(images.Spritesheet_png))
	if err != nil {
		return nil, err
	}

	sheet := ebiten.NewImageFromImage(img)

	// spriteAt returns a sprite at the provided coordinates.
	spriteAt := func(x, y int) *ebiten.Image {
		return sheet.SubImage(image.Rect(x*tileSize, (y+1)*tileSize, (x+1)*tileSize, y*tileSize)).(*ebiten.Image)
	}

	// Populate SpriteSheet.
	s := &SpriteSheet{} // <– Init map}

	s.Void = spriteAt(0, 0)
	s.Cursor = spriteAt(5, 10)
	s.EnemyCursor = spriteAt(6, 10)
	s.OakBeam = spriteAt(1, 0)
	s.LockedDoorWest = spriteAt(1, 2)
	s.LockedDoorEast = spriteAt(2, 2)
	s.OakWall = spriteAt(2, 0)
	s.OakWall2 = spriteAt(3, 0)
	s.OakChunk = spriteAt(4, 0)
	s.OakWall3 = spriteAt(5, 0)
	s.OakChunkSlim = spriteAt(6, 0)
	s.OakWall4 = spriteAt(7, 0)
	s.Water = spriteAt(8, 4)
	s.EnchantedWater = spriteAt(9, 4)
	s.GreyKnight = spriteAt(4, 7)

	//s.OakWall5 = spriteAt(8, 0)
	//s.OakChunk2 = spriteAt(9, 0)
	s.RedMan = spriteAt(0, 10)
	s.BlueMan = spriteAt(1, 10)
	s.Floor = spriteAt(10, 4)
	s.Wall = spriteAt(2, 3)
	s.Statue = spriteAt(5, 4)
	s.Tube = spriteAt(3, 4)
	s.Crown = spriteAt(8, 6)
	s.Portal = spriteAt(5, 6)

	return s, nil
}
