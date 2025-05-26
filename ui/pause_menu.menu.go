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

func NewPauseMenu(w, h int, onResume, onExit func()) *PauseMenu {
	pm := &PauseMenu{}

	menuStyle := DefaultMenuStyles() // Use default styles or customize

	menuWidth := 300
	menuHeight := 300 // Can be adjusted to fit instructions

	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	menuRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)
	menuInstructions := []string{"W/S/Arrows Navigate", "Enter/Space Select", "Esc Resume"}

	// Main Pause Menu Options
	mainOptions := []MenuOption{
		{Text: "Resume", Action: onResume},
		{Text: "Settings", Action: func() { pm.ShowSettings = true; pm.SwitchToSettings() }},
		{Text: "Exit Game", Action: onExit},
	}
	pm.MainMenu = NewMenu(menuRect, "PAUSED", mainOptions, menuStyle)
	pm.MainMenu.SetInstructions(menuInstructions)

	// Settings Menu Options
	settingsOptions := []MenuOption{
		{Text: "Back", Action: func() { pm.ShowSettings = false; pm.SwitchToMain() }},
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
