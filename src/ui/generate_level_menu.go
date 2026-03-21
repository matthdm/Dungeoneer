package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// GenerateLevelMenu presents options for creating a new level.
type GenerateLevelMenu struct {
	Menu *Menu
}

// NewGenerateLevelMenu creates a menu with blank and procedural generation choices.
func NewGenerateLevelMenu(w, h int, onBlank func(), onProcGen func(), onCancel func()) *GenerateLevelMenu {
	style := DefaultMenuStyles()
	mw, mh := 300, 200
	mx := (w - mw) / 2
	my := (h - mh) / 2
	rect := image.Rect(mx, my, mx+mw, my+mh)

	options := []MenuOption{
		{Text: "New Blank Level", Action: onBlank},
		{Text: "Procedural Level", Action: onProcGen},
		{Text: "Back", Action: onCancel},
	}
	menu := NewMenu(rect, "GENERATE LEVEL", options, style)
	menu.SetInstructions([]string{"W/S Navigate", "Enter/Space Select", "Esc Cancel"})
	return &GenerateLevelMenu{Menu: menu}
}

func (gm *GenerateLevelMenu) Show()                     { gm.Menu.Show() }
func (gm *GenerateLevelMenu) Hide()                     { gm.Menu.Hide() }
func (gm *GenerateLevelMenu) IsVisible() bool           { return gm.Menu.IsVisible() }
func (gm *GenerateLevelMenu) Update()                   { gm.Menu.Update() }
func (gm *GenerateLevelMenu) Draw(screen *ebiten.Image) { gm.Menu.Draw(screen) }
func (gm *GenerateLevelMenu) SetRect(r image.Rectangle) { gm.Menu.rect = r }
