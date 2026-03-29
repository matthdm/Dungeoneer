package ui

import (
	"dungeoneer/constants"
	"dungeoneer/entities"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// DevMenu provides a debug UI for spawning items. Toggle with F2.
type DevMenu struct {
	palette *ItemPalette
}

// NewDevMenu creates a new dev menu centered on the screen.
func NewDevMenu(w, h int, p *entities.Player, hint func(string)) *DevMenu {
	return &DevMenu{
		palette: NewItemPalette(w, h, p, hint),
	}
}

// Resize recomputes layout after a window resize.
func (dm *DevMenu) Resize(w, h int) {
	dm.palette.Resize(w, h)
}

// SetPlayer rebinds the palette to a newly loaded player.
func (dm *DevMenu) SetPlayer(p *entities.Player) {
	if dm.palette != nil {
		dm.palette.SetPlayer(p)
	}
}

// Update handles input and menu logic.
func (dm *DevMenu) Update() {
	if !constants.DebugMode {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		dm.palette.Toggle()
	}
	dm.palette.Update()
}

// IsVisible reports whether the item palette is currently open.
func (dm *DevMenu) IsVisible() bool { return dm.palette.visible }

// TogglePalette opens the palette if closed, or closes it if open.
func (dm *DevMenu) TogglePalette() { dm.palette.Toggle() }

// Draw renders the dev menu.
func (dm *DevMenu) Draw(screen *ebiten.Image) {
	if !constants.DebugMode {
		return
	}
	dm.palette.Draw(screen)
}
