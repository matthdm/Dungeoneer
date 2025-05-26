package images

import (
	"bytes"

	_ "embed"
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed spritesheet.png
	Spritesheet_png []byte

	//go:embed castle_fg.png
	Castle_FG_png []byte

	//go:embed new_game.png
	New_Game_png []byte

	//go:embed exit_game.png
	Exit_Game_png []byte

	//go:embed options.png
	Options_png []byte

	//go:embed window_icon.png
	Window_Icon_png []byte

	//go:embed smoke.png
	Smoke_png []byte
)

// LoadEmbeddedImage loads images available through embed system, can pass name reference instead of the path
func LoadEmbeddedImage(png []byte) (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(png))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

// Potentially deprecated in favor of load embedded image
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

func SetDefaultWindowIcon() {
	windowImage, err := LoadEmbeddedImage(Window_Icon_png)
	if err != nil {
		fmt.Printf("failed to load window_icon.png: %s", err)
		return
	}
	iconImages := []image.Image{windowImage}
	ebiten.SetWindowIcon(iconImages)
}
