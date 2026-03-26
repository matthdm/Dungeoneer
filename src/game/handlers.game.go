package game

import (
	"dungeoneer/controls"
	"dungeoneer/entities"
	"dungeoneer/items"
	"dungeoneer/menumanager"
	"dungeoneer/movement"
	"dungeoneer/pathing"
	"dungeoneer/spells"
	"dungeoneer/tiles"
	"math"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// isActionPressed returns true while the bound key is held down.
func (g *Game) isActionPressed(action controls.ActionID) bool {
	return ebiten.IsKeyPressed(g.Controls.GetBinding(action).Primary)
}

// isActionJustPressed returns true on the first frame the bound key is pressed.
func (g *Game) isActionJustPressed(action controls.ActionID) bool {
	return inpututil.IsKeyJustPressed(g.Controls.GetBinding(action).Primary)
}

func (g *Game) handleMainMenuInput() {
	if g.State != StateMainMenu || g.Menu == nil {
		return
	}

	// Mouse hover and click handling
	mx, my := ebiten.CursorPosition()
	for i, r := range g.Menu.EntryRects {
		if mx >= r.Min.X && mx <= r.Max.X && my >= r.Min.Y && my <= r.Max.Y {
			g.Menu.SelectedIndex = i
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				switch g.Menu.Options[i] {
				case "Continue":
					g.loadHub()
				case "New Game":
					g.loadHub()
				case "Load Game":
					// TODO: Show load game menu
				case "Options":
					// TODO: Show options menu
				case "Exit Game":
					os.Exit(2)
				}
				return
			}
			break
		}
	}

	// Confirm selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch g.Menu.Options[g.Menu.SelectedIndex] {
		case "Continue":
			g.loadHub()
		case "New Game":
			g.loadHub()
		case "Load Game":
			// TODO: Show load game menu
		case "Options":
			// TODO: Show options menu
		case "Exit Game":
			os.Exit(2)
		}
	}
}

func (g *Game) handlePause() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		// Close editor palettes first
		if g.editor.PaletteOpen {
			g.editor.TogglePalette()
			return
		}
		if g.editor.EntityPaletteOpen {
			g.editor.ToggleEntityPalette()
			return
		}
		menumanager.Manager().HandleEscapePress()
	}
	g.isPaused = g.State == StatePlaying && menumanager.Manager().IsMenuOpen()
}

func (g *Game) resumeGame() {
	menumanager.Manager().CloseActiveMenu()
	g.isPaused = false
}

func (g *Game) handleZoom() {
	var scrollY float64
	if ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		scrollY = -0.25
	} else if ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		scrollY = 0.25
	} else {
		_, scrollY = ebiten.Wheel()
		scrollY = math.Max(-1, math.Min(1, scrollY))
	}
	g.camScaleTo += scrollY * (g.camScaleTo / 7)
	g.camScaleTo = math.Max(g.minCamScale, math.Min(100, g.camScaleTo))

	div := 10.0
	if g.camScaleTo > g.camScale {
		g.camScale += (g.camScaleTo - g.camScale) / div
	} else {
		g.camScale -= (g.camScale - g.camScaleTo) / div
	}
}

func (g *Game) handlePan() {
	pan := 7.0 / g.camScale
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.camX -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.camX += pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.camY -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.camY += pan
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton3) {
		if g.mousePanX == math.MinInt32 {
			g.mousePanX, g.mousePanY = ebiten.CursorPosition()
		} else {
			x, y := ebiten.CursorPosition()
			dx := float64(g.mousePanX - x)
			dy := float64(g.mousePanY - y)
			g.camX -= dx * (pan / 100)
			g.camY += dy * (pan / 100)
		}
	} else {
		g.mousePanX, g.mousePanY = math.MinInt32, math.MinInt32
	}

	// Clamp camera
	worldWidth := float64(g.currentLevel.W * g.currentLevel.TileSize / 2)
	worldHeight := float64(g.currentLevel.H * g.currentLevel.TileSize / 2)
	g.camX = math.Max(-worldWidth, math.Min(worldWidth, g.camX))
	g.camY = math.Max(-worldHeight, math.Min(0, g.camY))
}

func (g *Game) handleClicks() {
	// Handle player movement (right-click)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		tx, ty := g.hoverTileX, g.hoverTileY
		if g.player.CanMoveTo(tx, ty, g.currentLevel) {
			path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, tx, ty)
			if len(path) > 0 {
				g.player.MoveController.SetPath(path)
				g.player.PathPreview = nil
			}
		}
	}

	// Handle player attacking monster or opening doors
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		worldX := (float64(mx)-float64(g.w/2))/g.camScale + g.camX
		worldY := (float64(my)-float64(g.h/2))/g.camScale - g.camY
		tx, ty := g.isoToCartesian(worldX, worldY)

		cx := int(math.Floor(tx - 1.5))
		cy := int(math.Floor(ty - 0.5))

		// Check if clicking on a door first
		if g.currentLevel != nil && g.isValidTile(cx, cy) {
			tile := g.currentLevel.Tile(cx, cy)
			if tile != nil && tile.HasTag(tiles.TagDoor) {
				if entities.IsAdjacentRanged(g.player.TileX, g.player.TileY, cx, cy, 2) {
					if tile.DoorState == 1 {
						if g.closeDoor(cx, cy) {
							return // Door closed, skip monster attack check
						}
					} else {
						if g.openDoor(cx, cy) {
							return // Door opened, skip monster attack check
						}
					}
				}
			}
		}

		// Primary attack dispatch — ability determines attack type.
		g.handlePrimaryAttack(tx, ty, cx, cy)
	}
}

func (g *Game) handleHoverTile() {
	mx, my := ebiten.CursorPosition()
	cx := (float64(mx)-float64(g.w/2))/g.camScale + g.camX
	cy := (float64(my)-float64(g.h/2))/g.camScale - g.camY
	tx, ty := g.isoToCartesian(cx, cy)
	// These offsets align the hover tile with visual center of diamond tiles
	g.hoverTileX = int(math.Floor(tx - 1.5))
	g.hoverTileY = int(math.Floor(ty - 0.5))

	// Update path preview
	if g.player != nil {
		path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, g.hoverTileX, g.hoverTileY)
		if len(path) > 0 {
			g.player.PathPreview = path
		} else {
			g.player.PathPreview = nil
		}
	}
}

func (g *Game) handleLevelHotkeys() {
	if g.isActionJustPressed(controls.ActionToggleKeybind) {
		if g.ControlsMenu != nil {
			if g.ControlsMenu.IsVisible() {
				g.ControlsMenu.Hide()
			} else {
				g.ControlsMenu.Show()
			}
		}
	}

	if g.isActionJustPressed(controls.ActionToggleKeybind) {
		controlToggle = !controlToggle
	}
	if g.isActionJustPressed(controls.ActionShowHUD) {
		g.ShowHUD = !g.ShowHUD
	}
	// F3 — toggle level editor (hardcoded dev shortcut, not in player controls)
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		if g.editor != nil {
			g.editor.Active = !g.editor.Active
		}
	}
	// F12 — open/close dev tools overlay
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		if g.DevTools != nil {
			g.DevTools.Toggle()
		}
	}
	if g.isActionJustPressed(controls.ActionHeroPanel) {
		if g.HeroPanel != nil {
			g.HeroPanel.Toggle()
		}
	}
	// Spell casting: dispatch through player's SpellSlots (ability-gated).
	spellActions := []controls.ActionID{
		controls.ActionSpell1, controls.ActionSpell2, controls.ActionSpell3,
		controls.ActionSpell4, controls.ActionSpell5, controls.ActionSpell6,
	}
	for i, action := range spellActions {
		if g.isActionJustPressed(action) && g.player != nil {
			g.castSpellSlot(i)
		}
	}
	if g.State == StateGameOver && ebiten.IsKeyPressed(ebiten.KeyV) {
		g.returnToHub()
	}
}

func (g *Game) handleInput() {
	switch g.State {
	case StateMainMenu:
		g.handleInputMainMenu()
	case StatePlaying:
		g.handleInputPlaying()
	case StateGameOver:
		g.handleInputGameOver()
	}
}

func (g *Game) handleInputMainMenu() {
	if g.Menu == nil {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.Menu.SelectedIndex--
		if g.Menu.SelectedIndex < 0 {
			g.Menu.SelectedIndex = len(g.Menu.Options) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.Menu.SelectedIndex++
		if g.Menu.SelectedIndex >= len(g.Menu.Options) {
			g.Menu.SelectedIndex = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch g.Menu.Options[g.Menu.SelectedIndex] {
		case "New Game":
			g.State = StatePlaying
		case "Options":
			// Future support
		case "Exit Game":
			os.Exit(0)
		}
	}
}

func (g *Game) handleInputPlaying() {
	g.handlePause()
	if g.isPaused {
		// Pause menu navigation handled separately
		g.PauseMenu.Update()
		return
	}
	// Dialogue panel blocks all other input when active
	if g.DialoguePanel != nil && g.DialoguePanel.Active {
		g.DialoguePanel.Update()
		return
	}

	if g.InventoryScreen != nil && g.InventoryScreen.Active {
		g.InventoryScreen.Update(g.player, g.ShowHint, func(it *items.Item) {
			g.spawnItemDrop(it, g.player.TileX, g.player.TileY)
		})
		return
	}
	if g.isActionJustPressed(controls.ActionInventory) {
		g.InventoryScreen.Open()
		return
	}

	// NPC interaction (E key) — check before other E-key handlers
	if g.isActionJustPressed(controls.ActionInteract) {
		if npc := g.findNearbyNPC(); npc != nil {
			g.openDialogue(npc)
			return
		}
	}

	g.handleZoom()
	g.handlePan()
	g.handleDash()
	g.handleGrapple()
	g.handlePlayerVelocity()
	g.handleHoverTile()
	g.handleDoorInteract()
	g.handleClicks()
	g.handleLevelHotkeys()
}

func (g *Game) handleDoorInteract() {
	if g.player == nil || g.currentLevel == nil {
		return
	}
	if !g.isValidTile(g.hoverTileX, g.hoverTileY) {
		return
	}
	tile := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
	if tile == nil || !tile.HasTag(tiles.TagDoor) {
		return
	}
	if tile.DoorState == 0 {
		lower := ""
		for _, s := range tile.Sprites {
			if isDoorSpriteID(s.ID) {
				lower = strings.ToLower(s.ID)
				break
			}
		}
		if strings.Contains(lower, "unlockeddoor") || strings.Contains(lower, "door_unlocked") {
			tile.DoorState = 1
			tile.IsWalkable = true
		} else if strings.Contains(lower, "door_locked") || strings.Contains(lower, "lockeddoor") {
			tile.DoorState = 3
			tile.IsWalkable = false
		}
	}
	if !entities.IsAdjacentRanged(g.player.TileX, g.player.TileY, g.hoverTileX, g.hoverTileY, 2) {
		return
	}

	isoX, isoY := g.cartesianToIso(float64(g.hoverTileX), float64(g.hoverTileY))
	hx := int((isoX-g.camX)*g.camScale + float64(g.w/2))
	hy := int((isoY+g.camY)*g.camScale + float64(g.h/2) - 18)

	if g.hintTimer == 0 {
		switch tile.DoorState {
		case 3:
			g.ShowHintAt("Locked door. Press Q to unlock", hx, hy)
		case 2:
			g.ShowHintAt("Closed door. Click to open", hx, hy)
		case 1:
			g.ShowHintAt("Open door. Click to close", hx, hy)
		default:
			g.ShowHintAt("Closed door. Click to open", hx, hy)
		}
	}
	if tile.DoorState == 3 && inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if g.unlockDoor(g.hoverTileX, g.hoverTileY) {
			g.ShowHintAt("Door unlocked", hx, hy)
		}
	}
}

func (g *Game) handlePlayerVelocity() {
	if g.player.IsDashing || g.player.Grapple.Active {
		return
	}
	dx, dy := 0.0, 0.0

	// In isometric view each action maps to a diagonal tile-space direction.
	if g.isActionPressed(controls.ActionMoveLeft) {
		dx -= 1
		dy += 1
	}
	if g.isActionPressed(controls.ActionMoveRight) {
		dx += 1
		dy -= 1
	}
	if g.isActionPressed(controls.ActionMoveUp) {
		dx -= 1
		dy -= 1
	}
	if g.isActionPressed(controls.ActionMoveDown) {
		dx += 1
		dy += 1
	}

	// Update facing direction based on horizontal movement
	if dx < 0 {
		g.player.LeftFacing = true
	} else if dx > 0 {
		g.player.LeftFacing = false
	}

	// Only enable velocity mode if a direction is pressed
	if dx != 0 || dy != 0 {
		g.player.MoveController.SetVelocityFromInput(dx, dy)
		g.player.MoveController.Mode = movement.VelocityMode
		mag := math.Hypot(dx, dy)
		if mag != 0 {
			g.player.LastMoveDirX = dx / mag
			g.player.LastMoveDirY = dy / mag
		}
	} else if g.player.MoveController.Mode == movement.VelocityMode {
		// Stop smoothly when keys released
		g.player.MoveController.Stop()
	}
}

func (g *Game) handleGrapple() {
	if g.player == nil {
		return
	}
	if g.isActionJustPressed(controls.ActionGrapple) {
		if g.player.Grapple.Active {
			g.player.CancelGrapple()
		} else if g.player.HasAbility("grapple") {
			g.player.StartGrapple(float64(g.hoverTileX), float64(g.hoverTileY))
		}
	}
}

func (g *Game) handleDash() {
	if g.player == nil {
		return
	}
	if g.player.Grapple.Active {
		return
	}
	if !g.isActionJustPressed(controls.ActionDash) {
		return
	}

	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY

	// Blink: mage movement — teleport toward cursor, wall-safe. 2s cooldown.
	if g.player.HasAbility("blink") {
		blinkInfo := spells.SpellInfo{Name: "blink", Cooldown: 2.0}
		if g.player.Caster.Ready(blinkInfo) {
			g.player.Caster.PutOnCooldown(blinkInfo)
			g.handleBlink(px, py, float64(g.hoverTileX), float64(g.hoverTileY))
		}
		return
	}

	// Dash: knight movement — directional burst.
	if g.player.HasAbility("dash") && g.player.DashCharges > 0 && !g.player.IsDashing {
		dirX, dirY := 0.0, 0.0
		if g.player.MoveController.Mode == movement.VelocityMode {
			dirX = g.player.MoveController.VelocityX
			dirY = g.player.MoveController.VelocityY
		} else if g.player.MoveController.Mode == movement.PathingMode {
			if g.player.MoveController.Moving {
				dirX = g.player.MoveController.TargetX - px
				dirY = g.player.MoveController.TargetY - py
			} else if len(g.player.MoveController.Path) > 0 {
				next := g.player.MoveController.Path[0]
				dirX = float64(next.X) - px
				dirY = float64(next.Y) - py
			}
		}
		if dirX == 0 && dirY == 0 {
			dirX = g.player.LastMoveDirX
			dirY = g.player.LastMoveDirY
		}
		g.player.StartDash(dirX, dirY)
	}
}

func (g *Game) handleInputGameOver() {
	g.handleLevelHotkeys()
}
