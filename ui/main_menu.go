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
	Background     *ebiten.Image
	FrameIndex     int
	FrameTick      int
}

func NewMainMenu() (*MainMenu, error) {
	newGameLabel, err := images.LoadEmbeddedImage(images.New_Game_png)
	if err != nil {
		return nil, err
	}
	optionsLabel, err := images.LoadEmbeddedImage(images.Options_png)
	if err != nil {
		return nil, err
	}
	exitGameLabel, err := images.LoadEmbeddedImage(images.Exit_Game_png)
	if err != nil {
		return nil, err
	}

	castleFG, err := images.LoadEmbeddedImage(images.Castle_FG_png)
	if err != nil {
		return nil, err
	}
	return &MainMenu{
		Options:       []string{"New Game", "Options", "Exit Game"},
		SelectedIndex: 0,
		NewGameLabel:  newGameLabel,
		OptionsLabel:  optionsLabel,
		ExitGameLabel: exitGameLabel,
		Background:    castleFG,
	}, nil
}

func (m *MainMenu) Update() {
	m.FrameTick++
	if m.FrameTick >= 10 { // adjust speed
		m.FrameTick = 0
	}

}
