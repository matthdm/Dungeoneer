package sprites

import (
	"dungeoneer/constants"
	"dungeoneer/images"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// WallFlavors lists all available wall sprite sheet flavors. The names
// correspond to the embedded PNG files as well as the cases handled in
// getWallPaletteFlavor.
var WallFlavors = []string{
	"brick", "brick1", "brick2", "brick3", "brick4", "brick5", "brick6",
	"catacomb", "cocutos", "crypt", "gallery", "gehena", "hive", "lair",
	"lapis", "moss", "mucus", "normal", "pandem1", "pandem2", "pandem3",
	"pandem4", "pandem6", "rock", "tunnel",
}

type WallSpriteSheet struct {
	Flavor         string
	Beam           *ebiten.Image // (0, 0)
	BeamNW         *ebiten.Image // (1, 0)
	BeamNE         *ebiten.Image // (2, 0)
	Log            *ebiten.Image // (3, 0)
	BeamSW         *ebiten.Image // (4, 0)
	LogSlim        *ebiten.Image // (5, 0)
	BeamNESW       *ebiten.Image // (6, 0)
	LogNE          *ebiten.Image // (7, 0)
	BeamSE         *ebiten.Image // (0, 1)
	LogNWSE        *ebiten.Image // (1, 1)
	LogSlim2       *ebiten.Image // (2, 1)
	WallNESW       *ebiten.Image // (3, 1)
	Wall           *ebiten.Image // (4, 1)
	WallNWSE       *ebiten.Image // (5, 1)
	LogSW          *ebiten.Image // (6, 1)
	Chunk          *ebiten.Image // (7, 1)
	LockedDoorNW   *ebiten.Image // (0, 2)
	LockedDoorNE   *ebiten.Image // (1, 2)
	UnlockedDoorNW *ebiten.Image // (2, 2)
	UnlockedDoorNE *ebiten.Image // (3, 2)
	Floor          *ebiten.Image // (4, 2)
}

func getWallPaletteFlavor(flavor string) []byte {
	switch flavor {
	case "brick":
		return images.Brick_png
	case "brick1":
		return images.Brick1_png
	case "brick2":
		return images.Brick2_png
	case "brick3":
		return images.Brick3_png
	case "brick4":
		return images.Brick4_png
	case "brick5":
		return images.Brick5_png
	case "brick6":
		return images.Brick6_png
	case "catacomb":
		return images.Catacomb_png
	case "cocutos":
		return images.Cocutos_png
	case "crypt":
		return images.Crypt_png
	case "gallery":
		return images.Gallery_png
	case "gehena":
		return images.Gehena_png
	case "hive":
		return images.Hive_png
	case "lair":
		return images.Lair_png
	case "lapis":
		return images.Lapis_png
	case "moss":
		return images.Moss_png
	case "mucus":
		return images.Mucus_png
	case "normal":
		return images.Normal_png
	case "pandem1":
		return images.Pandem1_png
	case "pandem2":
		return images.Pandem2_png
	case "pandem3":
		return images.Pandem3_png
	case "pandem4":
		return images.Pandem4_png
	case "pandem6":
		return images.Pandem6_png
	case "rock":
		return images.Rock_png
	case "tunnel":
		return images.Tunnel_png
	default:
		return images.Normal_png
	}

}

// LoadSpriteSheet loads the embedded SpriteSheet.
func LoadWallSpriteSheet(flavor string) (*WallSpriteSheet, error) {
	sheet, err := images.LoadEmbeddedImage(getWallPaletteFlavor(flavor))
	if err != nil {
		return nil, err
	}

	// spriteAt returns a sprite at the provided coordinates.
	spriteAt := func(x, y int) *ebiten.Image {
		return sheet.SubImage(image.Rect(x*constants.DefaultTileSize, (y+1)*constants.DefaultTileSize, (x+1)*constants.DefaultTileSize, y*constants.DefaultTileSize)).(*ebiten.Image)
	}

	// Populate SpriteSheet.
	wss := &WallSpriteSheet{Flavor: flavor}

	wss.Beam = spriteAt(0, 0)
	wss.BeamNW = spriteAt(1, 0)
	wss.BeamNE = spriteAt(2, 0)
	wss.Log = spriteAt(3, 0)
	wss.BeamSW = spriteAt(4, 0)
	wss.LogSlim = spriteAt(5, 0)
	wss.BeamNESW = spriteAt(6, 0)
	wss.LogNE = spriteAt(7, 0)
	wss.BeamSE = spriteAt(0, 1)
	wss.LogNWSE = spriteAt(1, 1)
	wss.LogSlim2 = spriteAt(2, 1)
	wss.WallNESW = spriteAt(3, 1)
	wss.Wall = spriteAt(4, 1)
	wss.WallNWSE = spriteAt(5, 1)
	wss.LogSW = spriteAt(6, 1)
	wss.Chunk = spriteAt(7, 1)
	wss.LockedDoorNW = spriteAt(0, 2)
	wss.LockedDoorNE = spriteAt(1, 2)
	wss.UnlockedDoorNW = spriteAt(2, 2)
	wss.UnlockedDoorNE = spriteAt(3, 2)
	wss.Floor = spriteAt(4, 2)

	return wss, nil
}
