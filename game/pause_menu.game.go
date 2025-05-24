package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type PauseMenu struct {
	selectedOption int
	options        []string
	width          int
	height         int
}

func NewPauseMenu() *PauseMenu {
	return &PauseMenu{
		selectedOption: 0,
		options:        []string{"Resume", "Settings", "Exit Game"},
		width:          300,
		height:         300,
	}
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(g.w, g.h)
	overlay.Fill(color.RGBA{0, 0, 0, 128}) // Black with 50% transparency
	screen.DrawImage(overlay, nil)

	// Define menu dimensions
	menuX := (g.w - g.pauseMenu.width) / 2
	menuY := (g.h - g.pauseMenu.height) / 2

	// Draw menu background
	vector.DrawFilledRect(screen, float32(menuX), float32(menuY), float32(g.pauseMenu.width), float32(g.pauseMenu.height), color.RGBA{10, 10, 10, 255}, false)

	// Draw menu border
	vector.StrokeRect(screen, float32(menuX), float32(menuY), float32(g.pauseMenu.width), float32(g.pauseMenu.height), 3, color.RGBA{150, 20, 15, 255}, false)

	// Draw title
	titleText := "DUNGEONEER - PAUSED"
	titleX := menuX + (g.pauseMenu.width-len(titleText)*8)/2 // Rough centering for debug font
	ebitenutil.DebugPrintAt(screen, titleText, titleX, menuY+20)

	// Draw menu options
	for i, option := range g.pauseMenu.options {
		y := menuY + 60 + i*35
		x := menuX + 20

		// Highlight selected option
		if i == g.pauseMenu.selectedOption {
			// Draw selection background
			vector.DrawFilledRect(screen, float32(x-10), float32(y-5), float32(g.pauseMenu.width-20), 30, color.RGBA{80, 75, 70, 255}, false)
			ebitenutil.DebugPrintAt(screen, "> "+option, x, y)
		} else {
			ebitenutil.DebugPrintAt(screen, "  "+option, x, y)
		}
	}

	// Draw instructions
	instructionY := menuY + g.pauseMenu.height - 55
	ebitenutil.DebugPrintAt(screen, "W/S Navigate", menuX+20, instructionY)
	ebitenutil.DebugPrintAt(screen, "ENTER/SPACE Select", menuX+20, instructionY+15)
	ebitenutil.DebugPrintAt(screen, "ESC Resume", menuX+20, instructionY+30)
}

func (g *Game) handlePauseMenu() {
	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()
	g.handlePauseMenuMouse(mouseX, mouseY)

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.pauseMenu.selectedOption--
		if g.pauseMenu.selectedOption < 0 {
			g.pauseMenu.selectedOption = len(g.pauseMenu.options) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.pauseMenu.selectedOption++
		if g.pauseMenu.selectedOption >= len(g.pauseMenu.options) {
			g.pauseMenu.selectedOption = 0
		}
	}

	// Handle menu selection with Enter or Space
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.executePauseMenuAction()
	}
}

func (g *Game) handlePauseMenuMouse(mouseX, mouseY int) {
	// Calculate menu dimensions (same as in drawPauseMenu)
	menuStartX := (g.w - g.pauseMenu.width) / 2
	menuStartY := (g.h - g.pauseMenu.height) / 2

	// Check if mouse is over any menu option
	for i := range g.pauseMenu.options {
		optionY := menuStartY + 60 + i*35
		optionX := menuStartX + 20
		optionWidth := g.pauseMenu.width - 40
		optionHeight := 30

		// Check if mouse is hovering over this option
		if mouseX >= optionX-10 && mouseX <= optionX+optionWidth &&
			mouseY >= optionY-5 && mouseY <= optionY+optionHeight {
			// Update selection to hovered option
			g.pauseMenu.selectedOption = i

			// Check for click
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.executePauseMenuAction()
			}
			break
		}
	}
}

func (g *Game) executePauseMenuAction() {
	switch g.pauseMenu.selectedOption {
	case 0: // Resume
		g.isPaused = false
	case 1: // Settings
		// TODO: Implement settings menu
		fmt.Println("Settings selected")
	case 2: // Exit Game
		// TODO: Implement proper quit
		fmt.Println("Exit Game selected")
	}
}
