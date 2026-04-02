// Package game orchestrates the main game loop, rendering, and high-level
// world updates.
//
// Rendering notes and conventions:
//   - The project uses an isometric/cartesian conversion (`cartesianToIso`) and
//     draws sprites anchored by the character's "feet center". This anchoring
//     affects rendering order and overlap; keep sprite offsets consistent.
//   - Offscreen targets are reused for scaled rendering to avoid allocating
//     temporary images each frame. getOrCreateOffscreen controls the lifetime of
//     these buffers; prefer reuse rather than creating new images in hot paths.
package game

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/levels"
	"dungeoneer/menumanager"
	"dungeoneer/spells"
	"dungeoneer/tiles"
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
	case StateDeathScreen:
		g.drawDeathScreen(screen)
	case StateVictoryScreen:
		g.drawVictoryScreen(screen)
	case StatePlaying:
		g.drawPlaying(screen, cx, cy)
	}

	if menumanager.Manager().IsMenuOpen() {
		if g.LoadLevelMenu != nil && g.LoadLevelMenu.Menu.IsVisible() {
			g.LoadLevelMenu.Draw(screen)
		} else if g.LoadPlayerMenu != nil && g.LoadPlayerMenu.Menu.IsVisible() {
			g.LoadPlayerMenu.Draw(screen)
		} else if g.GenerateMenu != nil && g.GenerateMenu.Menu.IsVisible() {
			g.GenerateMenu.Draw(screen)
		} else if g.ProcGenMenu != nil && g.ProcGenMenu.IsVisible() {
			g.ProcGenMenu.Draw(screen)
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
	// Draw mode buttons whenever entity palette is active
	if g.editor.Active && (g.editor.PaletteOpen || g.editor.EntityPaletteOpen) {
		g.editor.DrawModeButtons(screen)
	}
	if g.LinkPrompt != nil && g.LinkPrompt.IsVisible() {
		g.LinkPrompt.Draw(screen)
	}
	if g.ControlsMenu != nil && g.ControlsMenu.IsVisible() {
		g.ControlsMenu.Draw(screen)
	}
	if !controlToggle {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(constants.DEBUG_TEMPLATE, ebiten.ActualFPS(), ebiten.ActualTPS(), g.camScale, g.camX, g.camY))
	} else {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(constants.DEBUG_BINDS_TEMPLATE))
	}

	if g.hintTimer > 0 {
		ebitenutil.DebugPrintAt(screen, g.hint, g.hintX, g.hintY)
	}

}

func (g *Game) drawMainMenu(screen *ebiten.Image, cx, cy float64) {
	g.drawMainMenuLabels(screen, cx, cy)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	msg := "GAME OVER - Press V to Restart"
	ebitenutil.DebugPrintAt(screen, msg, g.w/2-100, g.h/2)
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
		g.Menu.ContinueGameLabel,
		g.Menu.NewGameLabel,
		g.Menu.LoadGameLabel,
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
	g.drawMonsterProjectiles(target, scale, cx, cy)

	//g.drawTiles(target, scale, cx, cy)
	//g.drawPathPreview(target, scale, cx, cy)
	//g.drawPlayer(target, scale, cx, cy)
	//g.drawMonsters(target, scale, cx, cy)
	//g.drawSpells(target, scale, cx, cy)
	g.drawHitMarkers(target, scale, cx, cy)
	g.drawDamageNumbers(target, scale, cx, cy)
	g.drawHealNumbers(target, scale, cx, cy)
	g.drawGrapple(target, scale, cx, cy)
	g.drawBossChainPull(target, scale, cx, cy)
	g.drawThroatDebug(target, scale, cx, cy)
	g.drawCombatDebugOverlays(target, scale, cx, cy)
	g.drawWallDebugOverlay(target, scale, cx, cy)

	if scaleLater {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Scale(g.camScale, g.camScale)
		op.GeoM.Translate(cx, cy)
		screen.DrawImage(target, op)
	}

	if g.ShowRays && len(g.cachedRays) > 0 {
		mc := g.player.MoveController
		isoX, isoY := g.cartesianToIso(mc.InterpX, mc.InterpY)
		ts := float64(g.currentLevel.TileSize)
		apexX := (isoX+ts/2-g.camX)*g.camScale + cx
		apexY := (isoY+ts/4+g.camY)*g.camScale + cy
		fov.DebugDrawRays(screen, g.cachedRays, apexX, apexY, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)
	}
	if g.HUD != nil && g.ShowHUD {
		g.HUD.Draw(screen, g.w, g.h)
	}
	if g.BossBar != nil {
		g.BossBar.Draw(screen, g.w)
	}
	g.drawBossFloorAnnouncement(screen)
	if g.HeroPanel != nil && g.HeroPanel.IsVisible() {
		g.HeroPanel.Draw(screen)
	}
	if g.InventoryScreen != nil && g.InventoryScreen.Active {
		g.InventoryScreen.Draw(screen, g.player)
	}
	if g.DialoguePanel != nil && g.DialoguePanel.Active {
		g.DialoguePanel.Draw(screen)
	}
	if g.DevMenu != nil {
		g.DevMenu.Draw(screen)
	}
	if g.DevTools != nil {
		g.DevTools.Draw(screen)
	}
}

func (g *Game) drawCombatDebugOverlays(target *ebiten.Image, scale, cx, cy float64) {
	if g.currentLevel == nil {
		return
	}
	if !g.ShowHitboxes && !g.ShowInteractionRadii {
		return
	}

	if g.ShowHitboxes {
		// Player body hitbox — use canonical body center.
		if g.player != nil {
			g.drawActorHitEllipse(target, g.player.BodyX(), g.player.BodyY(), 0.35, 0.82, scale, cx, cy, color.NRGBA{R: 0, G: 230, B: 255, A: 230}, 1.5)
		}

		// Enemy combat hit volumes — use BodyX/BodyY so visualization matches
		// detection (both in body-center space, not feet-anchor space).
		for _, m := range g.Monsters {
			if m == nil || m.IsDead {
				continue
			}
			r := m.HitRadius
			if r <= 0 {
				r = entities.DefaultMonsterHitRadius
			}
			g.drawActorHitEllipse(target, m.BodyX(), m.BodyY(), r*0.85, 0.66, scale, cx, cy, color.NRGBA{R: 255, G: 70, B: 70, A: 230}, 1.5)
		}

		// Projectile hit circles: player and monster projectiles.
		for _, sp := range g.ActiveSpells {
			ab, ok := sp.(*spells.ArcaneBolt)
			if !ok || ab == nil || ab.IsFinished() {
				continue
			}
			g.drawWorldCircle(target, ab.X, ab.Y, ab.Radius, scale, cx, cy, color.NRGBA{R: 255, G: 0, B: 255, A: 230}, 1.5, false)
		}
		for _, p := range g.MonsterProjectiles {
			if p == nil || p.Finished {
				continue
			}
			g.drawWorldCircle(target, p.X, p.Y, p.Radius, scale, cx, cy, color.NRGBA{R: 255, G: 140, B: 60, A: 230}, 1.5, true)
		}
	}

	if g.ShowInteractionRadii {
		// NPC interaction range.
		for _, n := range g.NPCs {
			if n == nil || !n.Interactable {
				continue
			}
			r := n.InteractRange
			if r <= 0 {
				r = 2
			}
			g.drawWorldCircle(target, n.InterpX, n.InterpY, r, scale, cx, cy, color.NRGBA{R: 255, G: 240, B: 80, A: 170}, 1.2, true)
		}

		// Floor exit interaction radius.
		if g.ExitEntity != nil && g.RunState != nil && g.RunState.Active {
			exitX, exitY := g.interactionCenterForTile(g.ExitEntity.TileX, g.ExitEntity.TileY)
			g.drawWorldCircle(
				target,
				exitX,
				exitY,
				3.0,
				scale, cx, cy,
				color.NRGBA{R: 80, G: 255, B: 120, A: 170},
				1.2,
				true,
			)
		}

		// Hub portal interaction radius.
		if g.IsInHub && g.hubPortalX >= 0 && g.hubPortalY >= 0 {
			portalX, portalY := g.interactionCenterForTile(g.hubPortalX, g.hubPortalY)
			g.drawWorldCircle(
				target,
				portalX,
				portalY,
				hubPortalInteractRadius,
				scale, cx, cy,
				color.NRGBA{R: 80, G: 220, B: 255, A: 170},
				1.2,
				true,
			)
		}
	}
}

func (g *Game) drawWorldRectFromCenter(target *ebiten.Image, centerX, centerY, w, h, scale, cx, cy float64, c color.NRGBA, strokeWidth float32, tileCentered bool) {
	halfW := w / 2
	halfH := h / 2
	x1, y1 := g.worldToScreenPoint(centerX-halfW, centerY-halfH, scale, cx, cy, tileCentered)
	x2, y2 := g.worldToScreenPoint(centerX+halfW, centerY-halfH, scale, cx, cy, tileCentered)
	x3, y3 := g.worldToScreenPoint(centerX+halfW, centerY+halfH, scale, cx, cy, tileCentered)
	x4, y4 := g.worldToScreenPoint(centerX-halfW, centerY+halfH, scale, cx, cy, tileCentered)

	vector.StrokeLine(target, x1, y1, x2, y2, strokeWidth, c, false)
	vector.StrokeLine(target, x2, y2, x3, y3, strokeWidth, c, false)
	vector.StrokeLine(target, x3, y3, x4, y4, strokeWidth, c, false)
	vector.StrokeLine(target, x4, y4, x1, y1, strokeWidth, c, false)
}

func (g *Game) drawActorHitEllipse(target *ebiten.Image, feetX, feetY, radiusTiles, heightTiles, scale, cx, cy float64, c color.NRGBA, strokeWidth float32) {
	if radiusTiles <= 0 {
		return
	}
	// tileCentered=false: callers pass the actual body center in cartesian space.
	sx, sy := g.worldToScreenPoint(feetX, feetY, scale, cx, cy, false)

	// Measure screen-space horizontal radius from the world axis that maps to iso X.
	sx2, sy2 := g.worldToScreenPoint(feetX+radiusTiles, feetY-radiusTiles, scale, cx, cy, false)
	rx := float32(math.Hypot(float64(sx2-sx), float64(sy2-sy)))
	if rx < 2 {
		rx = 2
	}

	// Vertical radius is intentionally taller than floor circles to read as body volume.
	ts := float64(g.currentLevel.TileSize)
	ry := float32(heightTiles * ts * scale * 0.6)
	if ry < rx*1.35 {
		ry = rx * 1.35
	}
	centerY := sy + ry*0.14

	g.drawScreenEllipse(target, sx, centerY, rx, ry, c, strokeWidth)
}

func (g *Game) drawWallDebugOverlay(target *ebiten.Image, scale, cx, cy float64) {
	if !g.ShowWalls || g.currentLevel == nil {
		return
	}
	ts := float64(g.currentLevel.TileSize) * scale
	halfW := ts / 2
	quarterH := ts / 4
	halfH := ts / 2
	col := color.NRGBA{R: 255, G: 0, B: 0, A: 230}

	for y := 0; y < g.currentLevel.H; y++ {
		for x := 0; x < g.currentLevel.W; x++ {
			tile := g.currentLevel.Tiles[y][x]
			if tile == nil || tile.IsWalkable {
				continue
			}
			ix, iy := g.cartesianToIso(float64(x), float64(y))
			sx := (ix-g.camX)*scale + cx
			sy := (iy+g.camY)*scale + cy

			topX, topY := float32(sx+halfW), float32(sy)
			rightX, rightY := float32(sx+ts), float32(sy+quarterH)
			botX, botY := float32(sx+halfW), float32(sy+halfH)
			leftX, leftY := float32(sx), float32(sy+quarterH)

			vector.StrokeLine(target, topX, topY, rightX, rightY, 1, col, false)
			vector.StrokeLine(target, rightX, rightY, botX, botY, 1, col, false)
			vector.StrokeLine(target, botX, botY, leftX, leftY, 1, col, false)
			vector.StrokeLine(target, leftX, leftY, topX, topY, 1, col, false)
		}
	}
}

func (g *Game) drawWorldCircle(target *ebiten.Image, centerX, centerY, r, scale, cx, cy float64, c color.NRGBA, strokeWidth float32, tileCentered bool) {
	if r <= 0 {
		return
	}
	const segments = 36
	var prevX, prevY float32
	for i := 0; i <= segments; i++ {
		t := 2 * math.Pi * float64(i) / float64(segments)
		px := centerX + math.Cos(t)*r
		py := centerY + math.Sin(t)*r
		sx, sy := g.worldToScreenPoint(px, py, scale, cx, cy, tileCentered)
		if i > 0 {
			vector.StrokeLine(target, prevX, prevY, sx, sy, strokeWidth, c, false)
		}
		prevX, prevY = sx, sy
	}
}

func (g *Game) drawScreenEllipse(target *ebiten.Image, centerX, centerY, rx, ry float32, c color.NRGBA, strokeWidth float32) {
	const segments = 40
	var prevX, prevY float32
	for i := 0; i <= segments; i++ {
		t := 2 * math.Pi * float64(i) / float64(segments)
		x := centerX + rx*float32(math.Cos(t))
		y := centerY + ry*float32(math.Sin(t))
		if i > 0 {
			vector.StrokeLine(target, prevX, prevY, x, y, strokeWidth, c, false)
		}
		prevX, prevY = x, y
	}
}

func (g *Game) worldToScreenPoint(x, y, scale, cx, cy float64, tileCentered bool) (float32, float32) {
	isoX, isoY := g.cartesianToIso(x, y)
	if tileCentered {
		ts := float64(g.currentLevel.TileSize)
		isoX += ts / 2
		isoY += ts / 4
	}
	sx := (isoX-g.camX)*scale + cx
	sy := (isoY+g.camY)*scale + cy
	return float32(sx), float32(sy)
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

// drawBossChainPull renders the chain line from Varn to the player for a brief
// flash after a pull_player attack fires.
func (g *Game) drawBossChainPull(target *ebiten.Image, scale, cx, cy float64) {
	if g.CurrentBoss == nil || g.CurrentBoss.PullLineTicks <= 0 || g.player == nil {
		return
	}
	bossX := float64(g.CurrentBoss.Monster.TileX)
	bossY := float64(g.CurrentBoss.Monster.TileY)
	playerX := g.player.MoveController.InterpX
	playerY := g.player.MoveController.InterpY

	bsx, bsy := g.cartesianToIso(bossX, bossY)
	psx, psy := g.cartesianToIso(playerX, playerY)

	x1 := float32((bsx-g.camX+30)*scale + cx)
	y1 := float32((bsy+g.camY+25)*scale + cy)
	x2 := float32((psx-g.camX+30)*scale + cx)
	y2 := float32((psy+g.camY+25)*scale + cy)

	// Fade out over the 14-tick lifetime.
	alpha := uint8(float32(g.CurrentBoss.PullLineTicks) / 14.0 * 200)
	chainColor := color.RGBA{R: 180, G: 180, B: 200, A: alpha}
	vector.StrokeLine(target, x1, y1, x2, y2, 2, chainColor, false)
}

// drawBossFloorAnnouncement renders a centered warning overlay for the first
// ~4 seconds after entering a boss floor.
func (g *Game) drawBossFloorAnnouncement(screen *ebiten.Image) {
	if g.bossFloorAnnouncement <= 0 {
		return
	}
	// Fade out over the last 60 ticks.
	alpha := 1.0
	if g.bossFloorAnnouncement < 60 {
		alpha = float64(g.bossFloorAnnouncement) / 60.0
	}
	msg := "A great evil stirs ahead..."
	charW := 7  // basicfont character width
	scale := 2.0
	textW := len(msg) * charW * int(scale)
	x := g.w/2 - textW/2
	y := g.h/3

	// Dark backdrop strip.
	stripH := int(13*scale) + 16
	stripImg := ebiten.NewImage(g.w, stripH)
	stripImg.Fill(color.RGBA{0, 0, 0, uint8(alpha * 140)})
	stripOp := &ebiten.DrawImageOptions{}
	stripOp.GeoM.Translate(0, float64(y-8))
	screen.DrawImage(stripImg, stripOp)

	// Text — fade by drawing at reduced alpha via a temporary image.
	textImg := ebiten.NewImage(textW+4, int(13*scale))
	ebitenutil.DebugPrintAt(textImg, msg, 0, 0)
	textOp := &ebiten.DrawImageOptions{}
	textOp.GeoM.Scale(scale, scale)
	textOp.GeoM.Translate(float64(x), float64(y))
	textOp.ColorScale.ScaleAlpha(float32(alpha))
	screen.DrawImage(textImg, textOp)
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
					op.ColorScale.Scale(0.2, 0.2, 0.2, 1.0)
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

func (g *Game) drawThroatDebug(target *ebiten.Image, scale, cx, cy float64) {
	if g.currentLevel == nil {
		return
	}
	if !g.ShowThroatValid && !g.ShowThroatInvalid && !g.ShowRegionDebug {
		return
	}

	info := levels.BuildThroatDebug(g.currentLevel, 18, 12)
	if g.ShowRegionDebug {
		for y := 0; y < g.currentLevel.H; y++ {
			for x := 0; x < g.currentLevel.W; x++ {
				id := info.RegionIDs[y][x]
				if id <= 0 {
					continue
				}
				isRoom := info.RegionIsRoom[id]
				baseR, baseG, baseB := uint8(70), uint8(140), uint8(220)
				if isRoom {
					baseR, baseG, baseB = 80, 200, 120
				}
				r := uint8((int(baseR) + (id*31)%120) % 255)
				gc := uint8((int(baseG) + (id*47)%120) % 255)
				b := uint8((int(baseB) + (id*59)%120) % 255)
				xi, yi := g.cartesianToIso(float64(x), float64(y))
				op := g.getDrawOp(xi, yi, scale, cx, cy)
				op.ColorScale.Scale(float32(r)/255.0, float32(gc)/255.0, float32(b)/255.0, 0.35)
				target.DrawImage(g.highlightImage, op)
			}
		}
	}

	if g.ShowThroatInvalid {
		for _, p := range info.Invalid {
			xi, yi := g.cartesianToIso(float64(p.X), float64(p.Y))
			op := g.getDrawOp(xi, yi, scale, cx, cy)
			op.ColorScale.Scale(1.0, 0.2, 0.2, 0.6)
			target.DrawImage(g.highlightImage, op)
		}
	}
	if g.ShowThroatValid {
		for _, p := range info.Valid {
			xi, yi := g.cartesianToIso(float64(p.X), float64(p.Y))
			op := g.getDrawOp(xi, yi, scale, cx, cy)
			op.ColorScale.Scale(1.0, 0.9, 0.1, 0.7)
			target.DrawImage(g.highlightImage, op)
		}
	}

	if g.ShowDoorDebug {
		for y := 0; y < g.currentLevel.H; y++ {
			for x := 0; x < g.currentLevel.W; x++ {
				tile := g.currentLevel.Tiles[y][x]
				if tile == nil || !tile.HasTag(tiles.TagDoor) {
					continue
				}
				xi, yi := g.cartesianToIso(float64(x), float64(y))
				op := g.getDrawOp(xi, yi, scale, cx, cy)
				if tile.DoorState == 3 {
					op.ColorScale.Scale(1.0, 0.2, 0.2, 0.7)
				} else {
					op.ColorScale.Scale(0.2, 0.9, 1.0, 0.6)
				}
				target.DrawImage(g.highlightImage, op)
			}
		}
	}
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

// Converts world coordinates to screen-space DrawImageOptions.
func (g *Game) getDrawOp(worldX, worldY, scale, cx, cy float64) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	// GeoM transform ordering is important: translate to world position,
	// apply camera transform, scale, then translate to screen center. This
	// ordering preserves the isometric anchor semantics (feet-centered sprites).
	op.GeoM.Translate(worldX, worldY)
	op.GeoM.Translate(-g.camX, g.camY)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx, cy)
	return op
}

// fovDecayFrames is how many frames a tile stays "visible" after the last
// ray hit it. This window absorbs the per-frame jitter in TraceLineToTiles
// caused by the sub-tile origin shift, eliminating boundary flicker without
// any perceptible lag (3 frames ≈ 50 ms at 60 TPS).
const fovDecayFrames = 3

func (g *Game) isTileVisible(x, y int) bool {
	if y < 0 || y >= len(g.visibleTick) || x < 0 || x >= len(g.visibleTick[y]) {
		return false
	}
	return g.FullBright || g.gameTick-g.visibleTick[y][x] <= fovDecayFrames
}
