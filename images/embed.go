package images

import (
	_ "embed"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed spritesheet.png
	Spritesheet_png []byte

	//go:embed smoke.png
	Smoke_png []byte
)

func LoadImage(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}
