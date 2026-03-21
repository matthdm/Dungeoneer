package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// updateDeathScreen handles input on the death summary screen.
func (g *Game) updateDeathScreen() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.returnToHub()
	}
	return nil
}

// updateVictoryScreen handles input on the victory summary screen.
func (g *Game) updateVictoryScreen() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.returnToHub()
	}
	return nil
}

// drawDeathScreen renders the run summary after the player dies.
func (g *Game) drawDeathScreen(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Dark overlay
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.NRGBA{0, 0, 0, 200}, false)

	cx := w / 2
	cy := h/2 - 80

	// Title
	title := "YOU HAVE FALLEN"
	ebitenutil.DebugPrintAt(screen, title, cx-len(title)*3, cy)
	cy += 30

	// Separator
	ebitenutil.DebugPrintAt(screen, "─────────────────────────", cx-75, cy)
	cy += 20

	// Stats
	if g.RunState != nil {
		lines := []string{
			fmt.Sprintf("Floors Cleared:   %d / %d", g.RunState.FloorsCleared, g.RunState.TotalFloors),
			fmt.Sprintf("Monsters Slain:   %d", g.RunState.KillCount),
			fmt.Sprintf("Remnants Earned:  %d", g.RunState.RemnantEarned),
			"",
			fmt.Sprintf("Total Remnants:   %d", g.Meta.Remnants),
			fmt.Sprintf("Total Runs:       %d", g.Meta.RunCount),
			fmt.Sprintf("Best Floor:       %d", g.Meta.BestFloor),
		}
		for _, line := range lines {
			ebitenutil.DebugPrintAt(screen, line, cx-90, cy)
			cy += 16
		}
	}

	cy += 20
	prompt := "Press ENTER to return"
	ebitenutil.DebugPrintAt(screen, prompt, cx-len(prompt)*3, cy)
}

// drawVictoryScreen renders the run summary after the player wins.
func (g *Game) drawVictoryScreen(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Dark overlay with slight gold tint
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.NRGBA{10, 8, 0, 200}, false)

	cx := w / 2
	cy := h/2 - 80

	// Title
	title := "VICTORY"
	ebitenutil.DebugPrintAt(screen, title, cx-len(title)*3, cy)
	cy += 30

	ebitenutil.DebugPrintAt(screen, "─────────────────────────", cx-75, cy)
	cy += 20

	if g.RunState != nil {
		lines := []string{
			fmt.Sprintf("Floors Cleared:   %d / %d", g.RunState.FloorsCleared, g.RunState.TotalFloors),
			fmt.Sprintf("Monsters Slain:   %d", g.RunState.KillCount),
			fmt.Sprintf("Remnants Earned:  %d (x2 Victory Bonus!)", g.RunState.RemnantEarned),
			"",
			fmt.Sprintf("Total Remnants:   %d", g.Meta.Remnants),
			fmt.Sprintf("Total Runs:       %d", g.Meta.RunCount),
			fmt.Sprintf("Best Floor:       %d", g.Meta.BestFloor),
		}
		for _, line := range lines {
			ebitenutil.DebugPrintAt(screen, line, cx-90, cy)
			cy += 16
		}
	}

	cy += 20
	prompt := "Press ENTER to return"
	ebitenutil.DebugPrintAt(screen, prompt, cx-len(prompt)*3, cy)
}
