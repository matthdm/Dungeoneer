package game

import (
	"dungeoneer/ui"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PauseMenu manages the game's pause state and owns the UI menus
type PauseMenu struct {
	game         *Game // Reference to the game instance for actions
	mainMenu     *ui.Menu
	settingsMenu *ui.Menu
}

func NewPauseMenu(g *Game) *PauseMenu {
	pm := &PauseMenu{game: g}

	menuStyle := ui.DefaultMenuStyles() // Use default styles or customize

	menuWidth := 300
	menuHeight := 300 // Can be adjusted to fit instructions

	menuX := (g.w - menuWidth) / 2
	menuY := (g.h - menuHeight) / 2
	menuRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)
	menuInstructions := []string{"W/S/Arrows Navigate", "Enter/Space Select", "Esc Resume"}

	// Main Pause Menu Options
	mainOptions := []ui.MenuOption{
		{Text: "Resume", Action: func() { pm.game.resumeGame() }},
		{Text: "Settings", Action: func() { pm.game.showSettings = true; pm.switchToSettings() }},
		{Text: "Exit Game", Action: func() { fmt.Println("Exit Game selected") /* Replace with actual quit */ }},
	}
	pm.mainMenu = ui.NewMenu(menuRect, "PAUSED", mainOptions, menuStyle)
	pm.mainMenu.SetInstructions(menuInstructions)

	// Settings Menu Options
	settingsOptions := []ui.MenuOption{
		{Text: "Back", Action: func() { pm.game.showSettings = false; pm.switchToMain() }},
	}
	pm.settingsMenu = ui.NewMenu(menuRect, "SETTINGS", settingsOptions, menuStyle)
	pm.settingsMenu.SetInstructions(menuInstructions)
	return pm
}

// Render the currently active menu
func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	if pm.game.showSettings {
		if pm.settingsMenu.IsVisible() {
			pm.settingsMenu.Draw(screen)
		}
	} else {
		if pm.mainMenu.IsVisible() {
			pm.mainMenu.Draw(screen)
		}
	}
}

// Update handles input for the currently active menu and global pause actions
func (pm *PauseMenu) Update() {
	// Global ESC key behavior for the pause system
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		pm.game.resumeGame() // From main pause menu, ESC resumes game
		return               // ESC action handled
	}

	// Update the currently active menu
	if pm.game.showSettings {
		if pm.settingsMenu.IsVisible() {
			pm.settingsMenu.Update()
		}
	} else {
		if pm.mainMenu.IsVisible() {
			pm.mainMenu.Update()
		}
	}
}

// Show shows the main pause menu
func (pm *PauseMenu) Show() {
	pm.game.isPaused = true
	pm.game.showSettings = false // Default to main pause menu
	pm.switchToMain()
}

func (pm *PauseMenu) switchToMain() {
	pm.mainMenu.Show()
	pm.mainMenu.SetSelectedIndex(0) // Reset selected menu option to 1st index
	pm.settingsMenu.Hide()
}

func (pm *PauseMenu) switchToSettings() {
	pm.settingsMenu.Show()
	pm.settingsMenu.SetSelectedIndex(0) // Reset selected menu option to 1st index
	pm.mainMenu.Hide()
}
