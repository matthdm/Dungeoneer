package game

import (
	"dungeoneer/constants"
	"dungeoneer/fov"
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var menuStart = time.Now()

func (g *Game) Draw(screen *ebiten.Image) {
	cx, cy := float64(g.w/2), float64(g.h/2)

	switch g.State {
	case StateMainMenu:
		g.drawMainMenu(screen, cx, cy)
	case StateGameOver:
		g.drawGameOver(screen)
	case StatePlaying:
		g.drawPlaying(screen, cx, cy)
	}

	if g.isPaused {
		g.pauseMenu.Draw(screen)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(constants.DEBUG_TEMPLATE, ebiten.ActualFPS(), ebiten.ActualTPS(), g.camScale, g.camX, g.camY))
}

func (g *Game) drawTiles(target *ebiten.Image, scale, cx, cy float64) {
	padding := float64(g.currentLevel.TileSize) * scale

	for y := 0; y < g.currentLevel.H; y++ {
		for x := 0; x < g.currentLevel.W; x++ {
			tile := g.currentLevel.Tiles[y][x]
			if tile == nil {
				continue
			}
			xi, yi := g.cartesianToIso(float64(x), float64(y))

			drawX := ((xi - g.camX) * scale) + cx
			drawY := ((yi + g.camY) * scale) + cy

			if drawX+padding < 0 || drawY+padding < 0 || drawX > float64(g.w) || drawY > float64(g.h) {
				continue
			}

			op := g.getDrawOp(xi, yi, scale, cx, cy)
			tile.Draw(target, op)
		}
	}
}

func (g *Game) drawHoverTile(target *ebiten.Image, scale, cx, cy float64) {
	if g.hoverTileX < 0 || g.hoverTileY < 0 ||
		g.hoverTileX >= g.currentLevel.W || g.hoverTileY >= g.currentLevel.H {
		return
	}

	xi, yi := g.cartesianToIso(float64(g.hoverTileX), float64(g.hoverTileY))
	op := g.getDrawOp(xi, yi, scale, cx, cy)

	// If this is the last tile in the path AND contains a living monster, draw red
	if len(g.player.PathPreview) > 0 {
		finalSpot := g.player.PathPreview[len(g.player.PathPreview)-1]
		for _, m := range g.Monsters {
			if !m.IsDead && m.TileX == finalSpot.X && m.TileY == finalSpot.Y {
				op.ColorScale.Scale(1, 0, 0, 0.8) // red
				break
			}
		}
	}

	target.DrawImage(g.highlightImage, op)
}

func (g *Game) drawPathPreview(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil {
		return
	}

	for _, step := range g.player.PathPreview {
		if step.X < 0 || step.Y < 0 || step.X >= g.currentLevel.W || step.Y >= g.currentLevel.H {
			continue
		}

		xi, yi := g.cartesianToIso(float64(step.X), float64(step.Y))
		op := g.getDrawOp(xi, yi, scale, cx, cy)
		op.ColorScale.Scale(1, 1, 1, 0.4)
		target.DrawImage(g.spriteSheet.Cursor, op)
	}
}

func (g *Game) drawMainMenuLabels(screen *ebiten.Image, cx, cy float64) {
	const magicNum = 195
	var scale = 0.25
	var spacing = magicNum*scale + 10

	if g.Menu.Background != nil {
		sw, sh := g.Menu.Background.Size()
		scaleX := float64(g.w) / float64(sw)
		scaleY := float64(g.h) / float64(sh)
		bgOp := &ebiten.DrawImageOptions{}
		bgOp.GeoM.Scale(scaleX, scaleY)
		screen.DrawImage(g.Menu.Background, bgOp)
	}
	labels := []*ebiten.Image{
		g.Menu.NewGameLabel,
		g.Menu.OptionsLabel,
		g.Menu.ExitGameLabel,
	}

	totalHeight := spacing * float64(len(labels)-1)
	startY := cy - totalHeight/2

	for i, img := range labels {
		x := 35.0
		y := startY + float64(i)*spacing

		op := &ebiten.DrawImageOptions{}

		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)

		// Only apply glow to the selected label
		if i == g.Menu.SelectedIndex {
			elapsed := time.Since(menuStart).Seconds()
			pulse := 0.3 + 0.7*math.Abs(math.Sin(elapsed*math.Pi))

			// Slight glow tint + pulse
			op.ColorScale.Scale(1.2, 1.1, 1.3, float32(pulse))
		}

		screen.DrawImage(img, op)
	}
}

func (g *Game) drawMainMenu(screen *ebiten.Image, cx, cy float64) {
	g.drawMainMenuLabels(screen, cx, cy)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	msg := "GAME OVER - Press V to Restart"
	ebitenutil.DebugPrintAt(screen, msg, g.w/2-100, g.h/2)
}

func (g *Game) drawPlaying(screen *ebiten.Image, cx, cy float64) {
	scaleLater := g.camScale > 1
	target := screen
	scale := g.camScale
	if scaleLater {
		target = g.getOrCreateOffscreen(screen.Bounds().Size())
		target.Clear()
		scale = 1
	}

	g.drawTiles(target, scale, cx, cy)
	g.drawPathPreview(target, scale, cx, cy)
	g.drawEntities(target, scale, cx, cy)
	g.drawHoverTile(target, scale, cx, cy)

	if scaleLater {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Scale(g.camScale, g.camScale)
		op.GeoM.Translate(cx, cy)
		screen.DrawImage(target, op)
	}

	if g.player != nil && len(g.cachedRays) > 0 && !g.FullBright {
		fov.DrawShadows(screen, g.cachedRays, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)
	}
	if g.ShowRays && len(g.cachedRays) > 0 {
		fov.DebugDrawRays(screen, g.cachedRays, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)
	}
	fov.DebugDrawWalls(screen, g.RaycastWalls, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)
}

// Converts world coordinates to screen-space DrawImageOptions.
func (g *Game) getDrawOp(worldX, worldY, scale, cx, cy float64) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(worldX, worldY)
	op.GeoM.Translate(-g.camX, g.camY)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx, cy)
	return op
}
