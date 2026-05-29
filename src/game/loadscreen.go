package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// loadScreenMeta and loadScreenRunSave hold data for the current load screen session.
var loadScreenMeta *MetaSave
var loadScreenRunSave *RunSave

// updateLoadScreen handles input on the load game screen.
func (g *Game) updateLoadScreen() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if loadScreenRunSave != nil {
			// Resume mid-run session.
			rs := loadScreenRunSave
			loadScreenRunSave = nil
			loadScreenMeta = nil
			g.resumeFromSave(rs)
		} else {
			// Load meta only — enter hub.
			if loadScreenMeta != nil {
				g.Meta = loadScreenMeta
			}
			loadScreenMeta = nil
			loadScreenRunSave = nil
			g.loadHub()
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		loadScreenMeta = nil
		loadScreenRunSave = nil
		g.State = StateMainMenu
	}
	return nil
}

// drawLoadScreen renders the load game screen.
func (g *Game) drawLoadScreen(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.NRGBA{0, 0, 0, 200}, false)

	cx := w / 2
	cy := h/2 - 80

	title := "LOAD GAME"
	ebitenutil.DebugPrintAt(screen, title, cx-len(title)*3, cy)
	cy += 30
	ebitenutil.DebugPrintAt(screen, "─────────────────────────", cx-75, cy)
	cy += 20

	if loadScreenRunSave != nil {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("Resume Run — Floor %d / %d",
				loadScreenRunSave.RunState.CurrentFloor,
				loadScreenRunSave.RunState.TotalFloors),
			cx-90, cy)
		cy += 16
	}

	if loadScreenMeta != nil {
		lines := [3]string{
			fmt.Sprintf("Total Runs:     %d", loadScreenMeta.RunCount),
			fmt.Sprintf("Best Floor:     %d", loadScreenMeta.BestFloor),
			fmt.Sprintf("Remnants:       %d", loadScreenMeta.Remnants),
		}
		for _, line := range lines {
			ebitenutil.DebugPrintAt(screen, line, cx-70, cy)
			cy += 16
		}
	}

	cy += 20
	ebitenutil.DebugPrintAt(screen, "ENTER  Continue", cx-50, cy)
	cy += 16
	ebitenutil.DebugPrintAt(screen, "ESC    Back", cx-50, cy)
}
