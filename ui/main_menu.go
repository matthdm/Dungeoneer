package ui

import (
	"dungeoneer/images"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type MainMenu struct {
	Options           []string
	SelectedIndex     int
	ContinueGameLabel *ebiten.Image
	NewGameLabel      *ebiten.Image
	LoadGameLabel     *ebiten.Image
	OptionsLabel      *ebiten.Image
	ExitGameLabel     *ebiten.Image
	Background        *ebiten.Image
	FrameTick         int
	EntryRects        []image.Rectangle
}

func NewMainMenu() (*MainMenu, error) {
	continueGameLabel, err := images.LoadEmbeddedImage(images.Continue_Game_png)
	if err != nil {
		return nil, err
	}
	newGameLabel, err := images.LoadEmbeddedImage(images.New_Game_png)
	if err != nil {
		return nil, err
	}
	loadGameLabel, err := images.LoadEmbeddedImage(images.Load_Game_png)
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
		Options:           []string{"Continue", "New Game", "Load Game", "Options", "Exit Game"},
		ContinueGameLabel: continueGameLabel,
		NewGameLabel:      newGameLabel,
		LoadGameLabel:     loadGameLabel,
		OptionsLabel:      optionsLabel,
		ExitGameLabel:     exitGameLabel,
		Background:        castleFG,
		EntryRects:        make([]image.Rectangle, 5),
	}, nil
}

func (m *MainMenu) Update() {
	m.FrameTick++
	if m.FrameTick >= 10 { // adjust speed
		m.FrameTick = 0
	}

}
