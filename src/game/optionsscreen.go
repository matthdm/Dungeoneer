package game

import (
	"dungeoneer/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

// optionsScreen is the singleton options panel, created lazily.
var optionsScreen *ui.OptionsScreen

// openOptions transitions to StateOptions, remembering the caller's state.
func (g *Game) openOptions() {
	g.previousState = g.State
	optionsScreen = ui.NewOptionsScreen(g.w, g.h,
		&g.Options.Fullscreen,
		&g.Options.MasterVolume,
		func() { // onSave
			g.Options.Apply()
			g.Options.Save()
		},
		func() { // onControls — show the existing ControlsMenu overlay
			g.ControlsMenu.Show()
		},
		func() { // onBack
			g.State = g.previousState
			optionsScreen = nil
		},
	)
	g.State = StateOptions
}

func (g *Game) updateOptionsScreen() error {
	// If ControlsMenu is open on top, let it handle input.
	if g.ControlsMenu != nil && g.ControlsMenu.IsVisible() {
		g.ControlsMenu.Update()
		return nil
	}
	if optionsScreen != nil {
		optionsScreen.Update()
	}
	return nil
}

func (g *Game) drawOptionsScreen(screen *ebiten.Image) {
	// Draw the previous state as background (dimmed by overlay).
	switch g.previousState {
	case StateMainMenu:
		cx, cy := float64(g.w/2), float64(g.h/2)
		g.drawMainMenu(screen, cx, cy)
	case StatePlaying:
		cx, cy := float64(g.w/2), float64(g.h/2)
		g.drawPlaying(screen, cx, cy)
	}

	if optionsScreen != nil {
		optionsScreen.Resize(g.w, g.h)
		optionsScreen.Draw(screen)
	}

	// ControlsMenu draws on top if open.
	if g.ControlsMenu != nil && g.ControlsMenu.IsVisible() {
		g.ControlsMenu.Draw(screen)
	}
}
