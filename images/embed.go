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

	//go:embed black_mage_full.png
	Black_Mage_Full_png []byte
	//go:embed black_mage.png
	Black_Mage_png []byte

	//go:embed brick.png
	Brick_png []byte
	//go:embed brick1.png
	Brick1_png []byte
	//go:embed brick2.png
	Brick2_png []byte
	//go:embed brick3.png
	Brick3_png []byte
	//go:embed brick4.png
	Brick4_png []byte
	//go:embed brick5.png
	Brick5_png []byte
	//go:embed brick6.png
	Brick6_png []byte
	//go:embed catacomb.png
	Catacomb_png []byte
	//go:embed cocutos.png
	Cocutos_png []byte
	//go:embed crypt.png
	Crypt_png []byte
	//go:embed gallery.png
	Gallery_png []byte
	//go:embed gehena.png
	Gehena_png []byte
	//go:embed hive.png
	Hive_png []byte
	//go:embed lair.png
	Lair_png []byte
	//go:embed lapis.png
	Lapis_png []byte
	//go:embed moss.png
	Moss_png []byte
	//go:embed mucus.png
	Mucus_png []byte
	//go:embed normal.png
	Normal_png []byte
	//go:embed pandem1.png
	Pandem1_png []byte
	//go:embed pandem2.png
	Pandem2_png []byte
	//go:embed pandem3.png
	Pandem3_png []byte
	//go:embed pandem4.png
	Pandem4_png []byte
	//go:embed pandem6.png
	Pandem6_png []byte
	//go:embed rock.png
	Rock_png []byte
	//go:embed tunnel.png
	Tunnel_png []byte

	//go:embed fireball_0.png
	Fireball_0_png []byte

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

	//go:embed item_subset.png
	Item_subset_png []byte

	//go:embed items_structured_effects.json
	Items_structured_effects_json []byte
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
