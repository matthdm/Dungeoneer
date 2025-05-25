package ui

import (
	"dungeoneer/images"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type MainMenu struct {
	Options        []string
	SelectedIndex  int
	HighlightImage *ebiten.Image
	Font           font.Face // Or bitmap font if using a sprite sheet
	NewGameLabel   *ebiten.Image
	OptionsLabel   *ebiten.Image
	ExitGameLabel  *ebiten.Image
}

func NewMainMenu(highLightImgPath string) (*MainMenu, error) {
	newGameLabel, err := images.LoadImage("images/new_game.png")
	if err != nil {
		return nil, err
	}
	optionsLabel, err := images.LoadImage("images/options.png")
	if err != nil {
		return nil, err
	}
	exitGameLabel, err := images.LoadImage("images/exit_game.png")
	if err != nil {
		return nil, err
	}
	return &MainMenu{
		Options:       []string{"New Game", "Options", "Exit Game"},
		SelectedIndex: 0,
		NewGameLabel:  newGameLabel,
		OptionsLabel:  optionsLabel,
		ExitGameLabel: exitGameLabel,
	}, nil
}
