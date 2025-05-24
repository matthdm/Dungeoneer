package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type PauseMenuOption struct {
	Text   string
	Action func(*Game)
}

type PauseMenu struct {
	selectedOption int
	options        []PauseMenuOption
	width          int
	height         int
	textPaddingY   int // Add a buffer for text spacing
	lineHeight     int
}

func NewPauseMenu() *PauseMenu {
	return &PauseMenu{
		selectedOption: 0,
		options: []PauseMenuOption{
			{Text: "Resume", Action: func(g *Game) { g.isPaused = false }},
			{Text: "Settings", Action: func(g *Game) { fmt.Println("Settings selected") /* TODO: Implement settings menu */ }},
			{Text: "Exit Game", Action: func(g *Game) { fmt.Println("Exit Game selected") /* TODO: Implement proper quit */ }},
		},
		width:        300,
		height:       300,
		textPaddingY: 60,
		lineHeight:   35,
	}
}

// Calculates and returns menu dimensions and position.
func (pm *PauseMenu) getMenuBounds(screenWidth, screenHeight int) (x, y, w, h float32) {
	menuX := (screenWidth - pm.width) / 2
	menuY := (screenHeight - pm.height) / 2
	return float32(menuX), float32(menuY), float32(pm.width), float32(pm.height)
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(g.w, g.h)
	overlay.Fill(color.RGBA{0, 0, 0, 128}) // Black with 50% transparency
	screen.DrawImage(overlay, nil)

	menuX, menuY, menuW, menuH := g.pauseMenu.getMenuBounds(g.w, g.h)

	// Draw menu background
	vector.DrawFilledRect(screen, menuX, menuY, menuW, menuH, color.RGBA{10, 10, 10, 255}, false)
	// Draw menu border
	vector.StrokeRect(screen, menuX, menuY, menuW, menuH, 3, color.RGBA{150, 20, 15, 255}, false)

	// Draw title
	titleText := "PAUSED"
	titleX := menuX + (menuW-float32(len(titleText)*8))/2
	ebitenutil.DebugPrintAt(screen, titleText, int(titleX), int(menuY+20))

	// Draw menu options
	for i, option := range g.pauseMenu.options {
		y := menuY + float32(g.pauseMenu.textPaddingY+(i*g.pauseMenu.lineHeight))
		x := menuX + 20

		// Highlight selected option
		if i == g.pauseMenu.selectedOption {
			// Draw selection background
			vector.DrawFilledRect(screen, x-10, y-5, float32(g.pauseMenu.width-20), 30, color.RGBA{80, 75, 70, 255}, false)
			ebitenutil.DebugPrintAt(screen, "> "+option.Text, int(x), int(y))
		} else {
			ebitenutil.DebugPrintAt(screen, "  "+option.Text, int(x), int(y))
		}
	}

	// Draw instructions
	instructionY := menuY + menuH - 55
	ebitenutil.DebugPrintAt(screen, "W/S Navigate", int(menuX+20), int(instructionY))
	ebitenutil.DebugPrintAt(screen, "ENTER/SPACE Select", int(menuX+20), int(instructionY+15))
	ebitenutil.DebugPrintAt(screen, "ESC Resume", int(menuX+20), int(instructionY+30))
}

func (g *Game) handlePauseMenu() {
	// Handle mouse input
	mouseX, mouseY := ebiten.CursorPosition()

	menuX, menuY, menuW, _ := g.pauseMenu.getMenuBounds(g.w, g.h)
	// Check if mouse is over any menu option
	for i := range g.pauseMenu.options {
		optionY := menuY + float32(60+i*35)
		optionX := menuX + 20
		optionWidth := menuW - 40
		optionHeight := float32(g.pauseMenu.lineHeight) - 5

		// Check if mouse is hovering over this option
		if float32(mouseX) >= optionX-10 && float32(mouseX) <= optionX+optionWidth &&
			float32(mouseY) >= optionY-5 && float32(mouseY) <= optionY+optionHeight {
			// Update selection to hovered option
			g.pauseMenu.selectedOption = i

			// Check for click
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.pauseMenu.options[g.pauseMenu.selectedOption].Action(g)
			}
			break
		}
	}

	// Handle keyboard navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.pauseMenu.selectedOption = (g.pauseMenu.selectedOption - 1 + len(g.pauseMenu.options)) % len(g.pauseMenu.options)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.pauseMenu.selectedOption = (g.pauseMenu.selectedOption + 1) % len(g.pauseMenu.options)
	}

	// Handle menu selection with Enter/Space
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.pauseMenu.options[g.pauseMenu.selectedOption].Action(g)
	}
}
