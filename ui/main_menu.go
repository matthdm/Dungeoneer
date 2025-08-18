package ui

import (
	"dungeoneer/images"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type MainMenu struct {
	Menu       *Menu
	Background *ebiten.Image
}

type MainMenuCallbacks struct {
	OnNewGame func()
	OnOptions func()
	OnExit    func()
}

func NewMainMenu(w, h int, cb MainMenuCallbacks) (*MainMenu, error) {
	castleFG, err := images.LoadEmbeddedImage(images.Castle_FG_png)
	if err != nil {
		return nil, err
	}
	menuWidth := 400
	menuHeight := 300
	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	rect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)
	options := []MenuOption{
		{Text: "New Game", Action: cb.OnNewGame},
		{Text: "Options", Action: cb.OnOptions},
		{Text: "Exit Game", Action: cb.OnExit},
	}
	menu := NewMenu(rect, "DUNGEONEER", options, DefaultMenuStyles())
	menu.SetInstructions([]string{"W/S/Arrows Navigate", "Enter/Space Select"})
	return &MainMenu{Menu: menu, Background: castleFG}, nil
}

func (m *MainMenu) Update() { m.Menu.Update() }

func (m *MainMenu) Draw(screen *ebiten.Image) {
	if m.Background != nil {
		sw, sh := m.Background.Size()
		bw, bh := screen.Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(bw)/float64(sw), float64(bh)/float64(sh))
		screen.DrawImage(m.Background, op)
	}
	m.Menu.Draw(screen)
}
