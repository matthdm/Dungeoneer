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
	options := mainMenuOptions(CanQuit)
	return &MainMenu{
		Options:           options,
		ContinueGameLabel: continueGameLabel,
		NewGameLabel:      newGameLabel,
		LoadGameLabel:     loadGameLabel,
		OptionsLabel:      optionsLabel,
		ExitGameLabel:     exitGameLabel,
		Background:        castleFG,
		EntryRects:        make([]image.Rectangle, len(options)),
	}, nil
}

func (m *MainMenu) Update() {
	m.FrameTick++
	if m.FrameTick >= 10 { // adjust speed
		m.FrameTick = 0
	}

}

// mainMenuOptions lists the main menu entries, leaving out "Exit Game" when
// the platform cannot quit.
func mainMenuOptions(canQuit bool) []string {
	options := []string{"Continue", "New Game", "Load Game", "Options"}
	if canQuit {
		options = append(options, "Exit Game")
	}
	return options
}

// Labels returns the label image for each entry in Options, in the same order.
func (m *MainMenu) Labels() []*ebiten.Image {
	labels := make([]*ebiten.Image, 0, len(m.Options))
	for _, option := range m.Options {
		switch option {
		case "Continue":
			labels = append(labels, m.ContinueGameLabel)
		case "New Game":
			labels = append(labels, m.NewGameLabel)
		case "Load Game":
			labels = append(labels, m.LoadGameLabel)
		case "Options":
			labels = append(labels, m.OptionsLabel)
		case "Exit Game":
			labels = append(labels, m.ExitGameLabel)
		}
	}
	return labels
}
