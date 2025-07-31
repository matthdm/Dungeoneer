package game

import (
	"dungeoneer/constants"
	"dungeoneer/fov"
	"dungeoneer/spells"
	"dungeoneer/ui"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var menuStart = time.Now()
var controlToggle bool = false

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
		if g.LoadLevelMenu != nil && g.LoadLevelMenu.Menu.IsVisible() {
			g.LoadLevelMenu.Draw(screen)
		} else if g.LoadPlayerMenu != nil && g.LoadPlayerMenu.Menu.IsVisible() {
			g.LoadPlayerMenu.Draw(screen)
		} else {
			g.PauseMenu.Draw(screen)
			if g.SavePrompt != nil && g.SavePrompt.IsVisible() {
				g.SavePrompt.Draw(screen)
			}
		}
	}

	if g.editor.Active && g.editor.PaletteOpen {
		g.editor.Palette.Draw(screen)
	}
	if g.editor.Active && g.editor.EntityPaletteOpen {
		g.editor.EntitiesPalette.Draw(screen)
	}
	if g.LinkPrompt != nil && g.LinkPrompt.IsVisible() {
		g.LinkPrompt.Draw(screen)
	}
	if !controlToggle {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(constants.DEBUG_TEMPLATE, ebiten.ActualFPS(), ebiten.ActualTPS(), g.camScale, g.camX, g.camY))
	} else {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(constants.DEBUG_BINDS_TEMPLATE))
	}

}

func (g *Game) drawMainMenu(screen *ebiten.Image, cx, cy float64) {
	g.drawMainMenuLabels(screen, cx, cy)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	msg := "GAME OVER - Press V to Restart"
	ebitenutil.DebugPrintAt(screen, msg, g.w/2-100, g.h/2)
}

func (g *Game) drawSpells(target *ebiten.Image, scale, cx, cy float64) {
	for _, sp := range g.ActiveSpells {
		sp.Draw(target, g.currentLevel.TileSize, g.camX, g.camY, scale, cx, cy)
		if g.SpellDebug {
			if fb, ok := sp.(*spells.Fireball); ok {
				fb.DebugDraw(target, g.currentLevel.TileSize, g.camX, g.camY, scale, cx, cy)
			}
		}
	}
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
	g.drawFloorTiles(target, scale, cx, cy)
	g.drawPathPreview(target, scale, cx, cy)
	renderables := g.collectRenderables(scale, cx, cy)
	for _, r := range renderables {
		target.DrawImage(r.Image, r.Options)
	}
	g.drawSpells(target, scale, cx, cy)

	//g.drawTiles(target, scale, cx, cy)
	//g.drawPathPreview(target, scale, cx, cy)
	//g.drawPlayer(target, scale, cx, cy)
	//g.drawMonsters(target, scale, cx, cy)
	//g.drawSpells(target, scale, cx, cy)
	g.drawHitMarkers(target, scale, cx, cy)
	g.drawDamageNumbers(target, scale, cx, cy)
	g.drawHealNumbers(target, scale, cx, cy)
	g.drawGrapple(target, scale, cx, cy)

	if scaleLater {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Scale(g.camScale, g.camScale)
		op.GeoM.Translate(cx, cy)
		screen.DrawImage(target, op)
	}

	if g.ShowRays && len(g.cachedRays) > 0 {
		fov.DebugDrawRays(screen, g.cachedRays, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)
	}
	if g.HUD != nil && g.ShowHUD {
		g.HUD.Draw(screen, g.w, g.h)
	}
	if g.HeroPanel != nil && g.HeroPanel.IsVisible() {
		g.HeroPanel.Draw(screen)
	}
	ui.DrawItemPalette(screen)
	if g.DevMenu != nil {
		g.DevMenu.Draw(screen)
	}
	//fov.DebugDrawWalls(screen, g.RaycastWalls, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)
}

func (g *Game) drawGrapple(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || !g.player.Grapple.Active {
		return
	}
	// Draw the rope from the player's current position so it follows them
	startX, startY := g.cartesianToIso(g.player.MoveController.InterpX, g.player.MoveController.InterpY)
	endX, endY := g.cartesianToIso(g.player.Grapple.HookPos.X, g.player.Grapple.HookPos.Y)
	sx1 := (startX-g.camX+30)*scale + cx
	sy1 := (startY+g.camY+25)*scale + cy
	sx2 := (endX-g.camX+30)*scale + cx
	sy2 := (endY+g.camY+25)*scale + cy
	vector.StrokeLine(target, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 2, color.White, false)
}

func (g *Game) drawFloorTiles(target *ebiten.Image, scale, cx, cy float64) {
	padding := float64(g.currentLevel.TileSize) * scale

	// Precompute screen-space bounds
	screenLeft := -padding
	screenTop := -padding
	screenRight := float64(g.w)
	screenBottom := float64(g.h)

	for y := 0; y < g.currentLevel.H; y++ {
		for x := 0; x < g.currentLevel.W; x++ {
			tile := g.currentLevel.Tiles[y][x]
			if tile == nil {
				continue
			}

			xi, yi := g.cartesianToIso(float64(x), float64(y))
			drawX := ((xi - g.camX) * scale) + cx
			drawY := ((yi + g.camY) * scale) + cy

			// Skip tiles that fall outside the screen bounds
			if drawX < screenLeft || drawY < screenTop || drawX > screenRight || drawY > screenBottom {
				continue
			}

			inFOV := g.isTileVisible(x, y)
			wasSeen := g.SeenTiles[y][x]

			// Fully hidden — skip
			if !inFOV && !wasSeen {
				continue
			}

			for _, s := range tile.Sprites {
				if !isFloorSprite(strings.ToLower(s.ID)) {
					continue
				}
				op := g.getDrawOp(xi, yi, scale, cx, cy)
				if !inFOV && wasSeen {
					op.ColorScale.Scale(0.4, 0.4, 0.4, 1.0)
				}
				target.DrawImage(s.Image, op)
			}
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
				op.ColorScale.Scale(1, 0, 0, constants.HostileTargetAlpha) // red
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
		op.ColorScale.Scale(1, 1, 1, constants.PathPreviewAlpha)
		target.DrawImage(g.spriteSheet.Cursor, op)
	}
}

func (g *Game) drawMainMenuLabels(screen *ebiten.Image, cx, cy float64) {

	var spacing = constants.MenuLabelHeightPixels*constants.MainMenuLabelScale + constants.MenuLabelVerticalPadding

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
	if len(g.Menu.EntryRects) != len(labels) {
		g.Menu.EntryRects = make([]image.Rectangle, len(labels))
	}

	totalHeight := spacing * float64(len(labels)-1)
	startY := cy - totalHeight/2

	for i, img := range labels {
		x := constants.MenuLabelOffsetX
		y := startY + float64(i)*spacing

		op := &ebiten.DrawImageOptions{}

		op.GeoM.Scale(constants.MainMenuLabelScale, constants.MainMenuLabelScale)
		op.GeoM.Translate(x, y)

		w, h := img.Size()
		rect := image.Rect(int(x), int(y), int(x+float64(w)*constants.MainMenuLabelScale), int(y+float64(h)*constants.MainMenuLabelScale))
		g.Menu.EntryRects[i] = rect

		// Only apply glow to the selected label
		if i == g.Menu.SelectedIndex {
			elapsed := time.Since(menuStart).Seconds()
			pulse := constants.GlowAlphaMin + constants.GlowAlphaRange*math.Abs(math.Sin(elapsed*math.Pi))

			// Slight glow tint + pulse
			op.ColorScale.Scale(1.2, 1.1, 1.3, float32(pulse))
		}

		screen.DrawImage(img, op)
	}
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

func (g *Game) isTileVisible(x, y int) bool {
	if y < 0 || y >= len(g.VisibleTiles) || x < 0 || x >= len(g.VisibleTiles[y]) {
		return false
	}
	return g.FullBright || g.VisibleTiles[y][x]
}
