package ui

import (
	"dungeoneer/controls"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// PauseMenuState tracks which submenu is active
type PauseMenuState int

const (
	PauseStateMain PauseMenuState = iota
	PauseStateSettings
	PauseStateControls
)

// PauseMenu manages the game's pause state and owns the UI menus

type PauseMenu struct {
	currentState PauseMenuState
	MainMenu     *Menu
	SettingsMenu *Menu
	ControlsMenu *ControlsMenu
}

type PauseMenuCallbacks struct {
	OnResume           func()
	OnLoadLevel        func()
	OnLoadPlayer       func()
	OnSavePlayer       func()
	OnSaveLevel        func()
	OnGenerate         func()
	OnExit             func()
	OnShowSettings     func()
	OnBackFromSettings func()
}

func NewPauseMenu(w, h int, ctrl *controls.Controls, cb PauseMenuCallbacks) *PauseMenu {
	pm := &PauseMenu{
		currentState: PauseStateMain,
	}

	menuStyle := DefaultMenuStyles()
	menuWidth := 300
	menuHeight := 400
	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	menuRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)
	menuInstructions := []string{"W/S/Arrows Navigate", "Enter/Space Select", "Esc Resume"}

	mainOptions := []MenuOption{
		{Text: "Resume", Action: cb.OnResume},
		{Text: "Load Level", Action: cb.OnLoadLevel},
		{Text: "Load Player", Action: cb.OnLoadPlayer},
		{Text: "Save Player", Action: cb.OnSavePlayer},
		{Text: "Save Level", Action: cb.OnSaveLevel},
		{Text: "Generate Level", Action: cb.OnGenerate},
		{Text: "Settings", Action: func() {
			pm.SwitchToSettings()
			if cb.OnShowSettings != nil {
				cb.OnShowSettings()
			}
		}},
		{Text: "Exit Game", Action: cb.OnExit},
	}
	pm.MainMenu = NewMenu(menuRect, "PAUSED", mainOptions, menuStyle)
	pm.MainMenu.SetInstructions(menuInstructions)

	settingsOptions := []MenuOption{
		{Text: "Controls", Action: func() {
			pm.SwitchToControls()
		}},
		{Text: "Back", Action: func() {
			pm.SwitchToMain()
			if cb.OnBackFromSettings != nil {
				cb.OnBackFromSettings()
			}
		}},
	}
	pm.SettingsMenu = NewMenu(menuRect, "SETTINGS", settingsOptions, menuStyle)
	pm.SettingsMenu.SetInstructions(menuInstructions)

	// Create controls menu with callback to return to settings
	pm.ControlsMenu = NewControlsMenu(w, h, ctrl, func() {
		pm.SwitchToSettings()
	})

	return pm
}

// Render the currently active menu
func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	switch pm.currentState {
	case PauseStateMain:
		if pm.MainMenu.IsVisible() {
			pm.MainMenu.Draw(screen)
		}
	case PauseStateSettings:
		if pm.SettingsMenu.IsVisible() {
			pm.SettingsMenu.Draw(screen)
		}
	case PauseStateControls:
		if pm.ControlsMenu.IsVisible() {
			pm.ControlsMenu.Draw(screen)
		}
	}
}

// Update handles input for the currently active menu and global pause actions
func (pm *PauseMenu) Update() {
	// Update the currently active menu
	switch pm.currentState {
	case PauseStateMain:
		if pm.MainMenu.IsVisible() {
			pm.MainMenu.Update()
		}
	case PauseStateSettings:
		if pm.SettingsMenu.IsVisible() {
			pm.SettingsMenu.Update()
		}
	case PauseStateControls:
		if pm.ControlsMenu.IsVisible() {
			pm.ControlsMenu.Update()
		}
	}
}

// Show shows the main pause menu
func (pm *PauseMenu) Show() {
	pm.currentState = PauseStateMain // Default to main pause menu
	pm.SwitchToMain()
}

func (pm *PauseMenu) SwitchToMain() {
	pm.currentState = PauseStateMain
	pm.MainMenu.Show()
	pm.MainMenu.SetSelectedIndex(0) // Reset selected menu option to 1st index
	pm.SettingsMenu.Hide()
	pm.ControlsMenu.Hide()
}

func (pm *PauseMenu) SwitchToSettings() {
	pm.currentState = PauseStateSettings
	pm.SettingsMenu.Show()
	pm.SettingsMenu.SetSelectedIndex(0) // Reset selected menu option to 1st index
	pm.MainMenu.Hide()
	pm.ControlsMenu.Hide()
}

func (pm *PauseMenu) SwitchToControls() {
	pm.currentState = PauseStateControls
	pm.ControlsMenu.Show()
	pm.MainMenu.Hide()
	pm.SettingsMenu.Hide()
}

func (pm *PauseMenu) Hide() {
	pm.MainMenu.Hide()
	pm.SettingsMenu.Hide()
	pm.ControlsMenu.Hide()
	pm.currentState = PauseStateMain
}

func (pm *PauseMenu) IsVisible() bool {
	return pm.MainMenu.IsVisible() || pm.SettingsMenu.IsVisible() || pm.ControlsMenu.IsVisible()
}
