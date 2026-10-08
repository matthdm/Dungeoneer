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
	"dungeoneer/coords"
	"dungeoneer/combat"
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/hud"
	"dungeoneer/levels"
	"dungeoneer/menumanager"
	"dungeoneer/spells"
	"dungeoneer/tiles"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
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

	// Ensure background is fully opaque black for screenshots
	screen.Fill(color.Black)

	switch g.State {
	case StateMainMenu:
		g.drawMainMenu(screen, cx, cy)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateDeathScreen:
		g.drawDeathScreen(screen)
	case StateVictoryScreen:
		g.drawVictoryScreen(screen)
	case StateLoadGame:
		g.drawLoadScreen(screen)
	case StateOptions:
		g.drawOptionsScreen(screen)
	case StatePlaying:
		g.drawPlaying(screen, cx, cy)
	}

	if g.State != StateOptions && menumanager.Manager().IsMenuOpen() {
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
	if g.State != StateOptions && g.ControlsMenu != nil && g.ControlsMenu.IsVisible() {
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

	if g.LoreLibrary != nil {
		g.LoreLibrary.Draw(screen)
	}
	if g.ArtifactLibrary != nil && g.ArtifactLibrary.Visible {
		g.ArtifactLibrary.Draw(screen)
	}
	if g.ArtifactLoadout != nil && g.ArtifactLoadout.Visible {
		g.ArtifactLoadout.Draw(screen)
	}
	if g.Shop != nil {
		g.Shop.Draw(screen)
	}
	if g.UpgradeStation != nil {
		g.UpgradeStation.Draw(screen)
	}
	if g.EchoShrine != nil {
		g.EchoShrine.Draw(screen)
	}
	if eventPanelOpen && g.ActiveEvent != nil {
		g.drawEventPanel(screen)
	}
	if g.ActiveToast != nil {
		g.ActiveToast.Draw(screen, g.w, g.h)
	}
	// Fade overlay drawn last so it covers everything.
	if g.Transition != nil {
		g.Transition.Draw(screen)
	}

	if g.ScreenshotFile != "" {
		f, err := os.Create(g.ScreenshotFile)
		if err == nil {
			png.Encode(f, screen)
			f.Close()
		}
		os.Exit(0)
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
	labels := g.Menu.Labels()
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

	if g.noSaveTimer > 0 {
		msg := "No saved game found"
		ebitenutil.DebugPrintAt(screen, msg, int(cx)-len(msg)*3, int(cy)+20)
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

// drawVoidboundAura renders the Voidbound Aura while voidbound_pendant is
// active. The aura is not a Spell and is never appended to ActiveSpells — it
// persists for as long as the pendant is equipped, gated here each frame.
func (g *Game) drawVoidboundAura(screen *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || g.currentLevel == nil {
		return
	}
	if !g.hasArtifactEquipped("voidbound_pendant") {
		return
	}
	if g.VoidboundAura == nil {
		g.VoidboundAura = &spells.VoidboundAura{}
	}
	g.VoidboundAura.Draw(screen, g.currentLevel.TileSize, g.player.BodyX(), g.player.BodyY(), g.camX, g.camY, scale, cx, cy)
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
	g.drawVoidboundAura(target, scale, cx, cy)
	g.drawMonsterProjectiles(target, scale, cx, cy)
	g.drawTargetRing(target, scale, cx, cy)

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
		// Populate 9D fields before drawing.
		if g.RunState != nil {
			g.HUD.CurrentFloor = g.RunState.CurrentFloor
			g.HUD.TotalFloors = g.RunState.TotalFloors
		}
		if g.FloorCtx != nil {
			b := string(g.FloorCtx.Biome)
			if len(b) > 0 {
				g.HUD.BiomeName = strings.ToUpper(b[:1]) + b[1:]
			}
		}
		// Rebuild status effect display list without allocating a new slice.
		g.HUD.StatusEffects = g.HUD.StatusEffects[:0]
		if g.player != nil {
			for _, eff := range g.player.Effects.Effects {
				if eff == nil {
					continue
				}
				g.HUD.StatusEffects = append(g.HUD.StatusEffects, buildStatusEffectDisplay(eff))
			}
		}
		// Inject new-engine condition states into the HUD status display.
		if na, ok := g.CombatAdapt.(*NewCombatAdapter); ok {
			cs := na.combatState
			if cs.InShadow {
				g.HUD.StatusEffects = append(g.HUD.StatusEffects, hud.StatusEffectDisplay{
					TypeChar: "S", Color: color.NRGBA{120, 60, 200, 255}, Duration: cs.ShadowTimer,
				})
			}
			if cs.TauntTimer > 0 {
				g.HUD.StatusEffects = append(g.HUD.StatusEffects, hud.StatusEffectDisplay{
					TypeChar: "T", Color: color.NRGBA{200, 80, 40, 255}, Duration: cs.TauntTimer,
				})
			}
			if cs.TargetRooted {
				g.HUD.StatusEffects = append(g.HUD.StatusEffects, hud.StatusEffectDisplay{
					TypeChar: "R", Color: color.NRGBA{80, 180, 255, 255}, Duration: cs.RootTimer,
				})
			}
			if cs.BurnActive {
				g.HUD.StatusEffects = append(g.HUD.StatusEffects, hud.StatusEffectDisplay{
					TypeChar: "F", Color: color.NRGBA{255, 140, 0, 255}, Duration: cs.BurnTimer,
				})
			}
			// Sync skill cooldowns from engine (authoritative for new-engine path).
			// SyncSkillCooldown detects the >0→0 transition and triggers the ready-flash.
			for i := 0; i < 6; i++ {
				g.HUD.SyncSkillCooldown(i, cs.ArtifactCooldowns[i])
			}

			// Duration bars: show active buff windows inside the matching skill slot.
			for i := 0; i < 6; i++ {
				id := cs.EquippedArtifacts[i]
				if id == "" {
					continue
				}
				eff, hasEff := combat.ArtifactEffects[id]
				switch id {
				case "shroud_cloak":
					if cs.InShadow {
						g.HUD.SkillSlots[i].DurationTimer = cs.ShadowTimer
						if hasEff {
							g.HUD.SkillSlots[i].MaxDuration = eff.DurationSec
						}
						g.HUD.SkillSlots[i].DurationColor = color.NRGBA{120, 60, 200, 255}
					} else {
						g.HUD.SkillSlots[i].DurationTimer = 0
					}
				case "wardens_medallion":
					if cs.TauntTimer > 0 {
						g.HUD.SkillSlots[i].DurationTimer = cs.TauntTimer
						if hasEff {
							g.HUD.SkillSlots[i].MaxDuration = eff.DurationSec
						}
						g.HUD.SkillSlots[i].DurationColor = color.NRGBA{200, 80, 40, 255}
					} else {
						g.HUD.SkillSlots[i].DurationTimer = 0
					}
				case "ashbound_chain":
					if cs.TargetRooted {
						g.HUD.SkillSlots[i].DurationTimer = cs.RootTimer
						if hasEff {
							g.HUD.SkillSlots[i].MaxDuration = eff.DurationSec
						}
						g.HUD.SkillSlots[i].DurationColor = color.NRGBA{80, 180, 255, 255}
					} else {
						g.HUD.SkillSlots[i].DurationTimer = 0
					}
				case "ember_mantle":
					if cs.BurnActive {
						g.HUD.SkillSlots[i].DurationTimer = cs.BurnTimer
						if hasEff {
							g.HUD.SkillSlots[i].MaxDuration = eff.BurnDurationSec
						}
						g.HUD.SkillSlots[i].DurationColor = color.NRGBA{255, 140, 0, 255}
					} else {
						g.HUD.SkillSlots[i].DurationTimer = 0
					}
				}
			}

			// Kill streak sync: trigger pulse on new kills.
			if cs.KillStreak != g.HUD.KillStreak {
				if cs.KillStreak > g.HUD.KillStreak {
					g.HUD.StreakPulse = 0.3
				}
				g.HUD.KillStreak = cs.KillStreak
			}
		}
		g.HUD.Draw(screen, g.w, g.h)
	}
	// Draw minimap (only during dungeon runs, not in hub).
	if g.HUD != nil && g.HUD.Minimap != nil && g.currentLevel != nil && !g.IsInHub {
		exitTX, exitTY := -1, -1
		if g.ExitEntity != nil {
			exitTX = g.ExitEntity.TileX
			exitTY = g.ExitEntity.TileY
		}
		playerTX, playerTY := 0, 0
		if g.player != nil {
			playerTX = g.player.TileX
			playerTY = g.player.TileY
		}
		g.HUD.Minimap.Draw(
			screen,
			g.currentLevel.Rooms,
			g.SeenTiles,
			playerTX, playerTY,
			exitTX, exitTY,
			g.w, g.h,
		)
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
	g.drawCombatDebugOverlay(screen)
	// Target panel: monster name + health bar, top-center of screen.
	g.drawTargetPanel(screen)

	// Draw particles on top of world, below UI overlays.
	if g.Particles != nil {
		g.Particles.Draw(screen)
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
		// Player body hitbox.
		if g.player != nil {
			g.drawActorHitEllipse(target, g.player.BodyX(), g.player.BodyY(), 0.35, 0.82, scale, cx, cy, color.NRGBA{R: 0, G: 230, B: 255, A: 230}, 1.5)
		}

		// Enemy combat hit volumes: the ground circle is exactly what hit
		// tests measure against (BodyCenter() ± radius in world space); the
		// tall ellipse is the same circle lifted onto the body.
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
			g.drawWorldCircle(target, ab.X, ab.Y, ab.Radius, scale, cx, cy, color.NRGBA{R: 255, G: 0, B: 255, A: 230}, 1.5, true)
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
			nb := n.Pos().BodyCenter()
			g.drawWorldCircle(target, nb.X, nb.Y, r, scale, cx, cy, color.NRGBA{R: 255, G: 240, B: 80, A: 170}, 1.2, false)
		}

		// Floor exit interaction radius.
		if g.ExitEntity != nil && g.RunState != nil && g.RunState.Active {
			// The check compares stored positions; shifting both ends to
			// body centres gives the same circle on the ground.
			exitX, exitY := g.interactionCenterForTile(g.ExitEntity.TileX, g.ExitEntity.TileY)
			exitC := coords.WorldPos{X: exitX, Y: exitY}.BodyCenter()
			g.drawWorldCircle(
				target,
				exitC.X,
				exitC.Y,
				3.0,
				scale, cx, cy,
				color.NRGBA{R: 80, G: 255, B: 120, A: 170},
				1.2,
				false,
			)
		}

		// Hub portal interaction radius.
		if g.IsInHub && g.hubPortalX >= 0 && g.hubPortalY >= 0 {
			portalX, portalY := g.interactionCenterForTile(g.hubPortalX, g.hubPortalY)
			portalC := coords.WorldPos{X: portalX, Y: portalY}.BodyCenter()
			g.drawWorldCircle(
				target,
				portalC.X,
				portalC.Y,
				hubPortalInteractRadius,
				scale, cx, cy,
				color.NRGBA{R: 80, G: 220, B: 255, A: 170},
				1.2,
				false,
			)
		}
	}
}


// drawActorHitEllipse draws an actor's hit volume: (bodyX, bodyY) is its
// BodyCenter() in world space and radiusTiles the radius hit tests use.
func (g *Game) drawActorHitEllipse(target *ebiten.Image, bodyX, bodyY, radiusTiles, heightTiles, scale, cx, cy float64, c color.NRGBA, strokeWidth float32) {
	if radiusTiles <= 0 {
		return
	}
	// Exact logical footprint on the floor.
	g.drawWorldCircle(target, bodyX, bodyY, radiusTiles, scale, cx, cy, c, strokeWidth, false)

	// The same radius as a tall ellipse around the body, to read as volume.
	bsx, bsy := g.bodyToScreen(bodyX, bodyY, scale, cx, cy)
	// Screen half-width of a world circle: its widest points lie along the
	// (+1,-1) world diagonal.
	d := radiusTiles / math.Sqrt2
	ex, _ := g.bodyToScreen(bodyX+d, bodyY-d, scale, cx, cy)
	rx := float32(ex - bsx)
	if rx < 2 {
		rx = 2
	}
	ts := float64(g.currentLevel.TileSize)
	ry := float32(heightTiles * ts * scale * 0.3)
	faded := c
	faded.A /= 2
	g.drawScreenEllipse(target, float32(bsx), float32(bsy), rx, ry, faded, strokeWidth)
}

// drawWallDebugOverlay outlines the floor diamond of every unwalkable tile:
// the projected world-space corners of the tile, i.e. exactly the area that
// walkability, projectile wall hits and cursor picks treat as that tile.
func (g *Game) drawWallDebugOverlay(target *ebiten.Image, scale, cx, cy float64) {
	if !g.ShowWalls || g.currentLevel == nil {
		return
	}
	col := color.NRGBA{R: 255, G: 0, B: 0, A: 230}
	corners := [4][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}

	for y := 0; y < g.currentLevel.H; y++ {
		for x := 0; x < g.currentLevel.W; x++ {
			tile := g.currentLevel.Tiles[y][x]
			if tile == nil || tile.IsWalkable {
				continue
			}
			var px, py [4]float32
			for i, c := range corners {
				sx, sy := g.groundToScreen(float64(x)+c[0], float64(y)+c[1], scale, cx, cy)
				px[i], py[i] = float32(sx), float32(sy)
			}
			for i := 0; i < 4; i++ {
				j := (i + 1) % 4
				vector.StrokeLine(target, px[i], py[i], px[j], py[j], 1, col, false)
			}
		}
	}
}

// drawWorldCircle draws a circle of radius r (tiles) around a world ground
// point. With atBodyHeight the whole circle is lifted onto bodies standing
// there — use it for things that fly at body height.
func (g *Game) drawWorldCircle(target *ebiten.Image, centerX, centerY, r, scale, cx, cy float64, c color.NRGBA, strokeWidth float32, atBodyHeight bool) {
	if r <= 0 {
		return
	}
	const segments = 36
	var prevX, prevY float32
	for i := 0; i <= segments; i++ {
		t := 2 * math.Pi * float64(i) / float64(segments)
		px := centerX + math.Cos(t)*r
		py := centerY + math.Sin(t)*r
		var fx, fy float64
		if atBodyHeight {
			fx, fy = g.bodyToScreen(px, py, scale, cx, cy)
		} else {
			fx, fy = g.groundToScreen(px, py, scale, cx, cy)
		}
		sx, sy := float32(fx), float32(fy)
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

func (g *Game) drawGrapple(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || !g.player.Grapple.Active {
		return
	}
	// Rope from the player's body to the hook. HookPos is tracked like an
	// entity position (cell origin), so BodyCenter() gives its world point.
	hook := coords.WorldPos{X: g.player.Grapple.HookPos.X, Y: g.player.Grapple.HookPos.Y}.BodyCenter()
	sx1, sy1 := g.bodyToScreen(g.player.BodyX(), g.player.BodyY(), scale, cx, cy)
	sx2, sy2 := g.bodyToScreen(hook.X, hook.Y, scale, cx, cy)
	vector.StrokeLine(target, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 2, color.White, false)
}

// drawBossChainPull renders the chain line from Varn to the player for a brief
// flash after a pull_player attack fires.
func (g *Game) drawBossChainPull(target *ebiten.Image, scale, cx, cy float64) {
	if g.CurrentBoss == nil || g.CurrentBoss.PullLineTicks <= 0 || g.player == nil {
		return
	}
	// Body to body.
	bx1, by1 := g.bodyToScreen(g.CurrentBoss.Monster.BodyX(), g.CurrentBoss.Monster.BodyY(), scale, cx, cy)
	bx2, by2 := g.bodyToScreen(g.player.BodyX(), g.player.BodyY(), scale, cx, cy)
	x1, y1, x2, y2 := float32(bx1), float32(by1), float32(bx2), float32(by2)

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
	op.GeoM.Translate(-(g.camX+g.shakeOffsetX), g.camY-g.shakeOffsetY)
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

// buildStatusEffectDisplay converts a StatusEffect into the HUD display model.
func buildStatusEffectDisplay(eff *entities.StatusEffect) hud.StatusEffectDisplay {
	type effectInfo struct {
		char string
		col  color.NRGBA
	}
	m := map[entities.EffectType]effectInfo{
		entities.EffectPoison: {"P", color.NRGBA{100, 200, 50, 255}},
		entities.EffectBurn:   {"B", color.NRGBA{255, 120, 20, 255}},
		entities.EffectSlow:   {"S", color.NRGBA{80, 80, 200, 255}},
		entities.EffectShield: {"+", color.NRGBA{200, 200, 255, 255}},
		entities.EffectWeaken: {"W", color.NRGBA{180, 50, 180, 255}},
		entities.EffectHaste:  {"H", color.NRGBA{255, 240, 80, 255}},
	}
	if info, ok := m[eff.Type]; ok {
		return hud.StatusEffectDisplay{TypeChar: info.char, Color: info.col, Duration: eff.Duration}
	}
	return hud.StatusEffectDisplay{TypeChar: "?", Color: color.NRGBA{200, 200, 200, 255}, Duration: eff.Duration}
}
