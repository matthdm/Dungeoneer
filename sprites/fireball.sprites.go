package sprites

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type FireballSprite struct {
	Frames []*ebiten.Image
}

func LoadFireballSprite(sheet *ebiten.Image) *FireballSprite {
	var frames []*ebiten.Image
	frameSize := 32
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			sub := sheet.SubImage(image.Rect(x*frameSize, y*frameSize, (x+1)*frameSize, (y+1)*frameSize)).(*ebiten.Image)
			frames = append(frames, sub)
		}
	}
	return &FireballSprite{Frames: frames}
}
