package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// PauseMenu manages the game's pause state and owns the UI menus

type PauseMenu struct {
	ShowSettings bool
	MainMenu     *Menu
	SettingsMenu *Menu
}

type PauseMenuCallbacks struct {
	OnResume           func()
	OnLoadLevel        func()
	OnLoadPlayer       func()
	OnSavePlayer       func()
	OnSaveLevel        func()
	OnNewBlank         func()
	OnExit             func()
	OnShowSettings     func()
	OnBackFromSettings func()
}

func NewPauseMenu(w, h int, cb PauseMenuCallbacks) *PauseMenu {
	pm := &PauseMenu{}

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
		{Text: "New Blank Level", Action: cb.OnNewBlank},
		{Text: "Settings", Action: func() {
			pm.ShowSettings = true
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
		{Text: "Back", Action: func() {
			pm.ShowSettings = false
			pm.SwitchToMain()
			if cb.OnBackFromSettings != nil {
				cb.OnBackFromSettings()
			}
		}},
	}
	pm.SettingsMenu = NewMenu(menuRect, "SETTINGS", settingsOptions, menuStyle)
	pm.SettingsMenu.SetInstructions(menuInstructions)

	return pm
}

// Render the currently active menu
func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	if pm.ShowSettings {
		if pm.SettingsMenu.IsVisible() {
			pm.SettingsMenu.Draw(screen)
		}
	} else {
		if pm.MainMenu.IsVisible() {
			pm.MainMenu.Draw(screen)
		}
	}
}

// Update handles input for the currently active menu and global pause actions
func (pm *PauseMenu) Update() {
	// Update the currently active menu
	if pm.ShowSettings {
		if pm.SettingsMenu.IsVisible() {
			pm.SettingsMenu.Update()
		}
	} else {
		if pm.MainMenu.IsVisible() {
			pm.MainMenu.Update()
		}
	}
}

// Show shows the main pause menu
func (pm *PauseMenu) Show() {
	pm.ShowSettings = false // Default to main pause menu
	pm.SwitchToMain()
}

func (pm *PauseMenu) SwitchToMain() {
	pm.MainMenu.Show()
	pm.MainMenu.SetSelectedIndex(0) // Reset selected menu option to 1st index
	pm.SettingsMenu.Hide()
	pm.ShowSettings = false
}

func (pm *PauseMenu) SwitchToSettings() {
	pm.SettingsMenu.Show()
	pm.SettingsMenu.SetSelectedIndex(0) // Reset selected menu option to 1st index
	pm.MainMenu.Hide()
}

func (pm *PauseMenu) Hide() {
	pm.MainMenu.Hide()
	pm.SettingsMenu.Hide()
	pm.ShowSettings = false
}

func (pm *PauseMenu) IsVisible() bool {
	return pm.MainMenu.IsVisible() || pm.SettingsMenu.IsVisible()
}
