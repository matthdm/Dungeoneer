package game

import (
	"dungeoneer/entities"
	"dungeoneer/levels"
	"dungeoneer/movement"
	"dungeoneer/pathing"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) handleMainMenuInput() {
	if g.State != StateMainMenu || g.Menu == nil {
		return
	}

	// Navigate up
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.Menu.SelectedIndex--
		if g.Menu.SelectedIndex < 0 {
			g.Menu.SelectedIndex = len(g.Menu.Options) - 1
		}
	}

	// Navigate down
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.Menu.SelectedIndex++
		if g.Menu.SelectedIndex >= len(g.Menu.Options) {
			g.Menu.SelectedIndex = 0
		}
	}

	// Confirm selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch g.Menu.Options[g.Menu.SelectedIndex] {
		case "New Game":
			//newGame, _ := NewGame()
			//*g = *newGame

			//TODO to add character creation sequence?
			//TODO to add intro cinematic
			//Game object already created
			g.State = StatePlaying
		case "Options":
			// Implement later
		case "Exit Game":
			os.Exit(2)
		}
	}
}

func (g *Game) handlePause() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.isPaused {
			// Only resume if on main pause menu
			if !g.pauseMenu.ShowSettings {
				g.resumeGame()
			} else {
				// go back to main pause menu instead
				g.pauseMenu.SwitchToMain()
			}
		} else {
			g.isPaused = true
			g.pauseMenu.Show()
		}
	}
}

func (g *Game) resumeGame() {
	g.isPaused = false
	if g.pauseMenu != nil { // Ensure pauseMenu exists
		g.pauseMenu.MainMenu.Hide()
		g.pauseMenu.SettingsMenu.Hide()
	}
}

func (g *Game) handleZoom() {
	var scrollY float64
	if ebiten.IsKeyPressed(ebiten.KeyC) || ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		scrollY = -0.25
	} else if ebiten.IsKeyPressed(ebiten.KeyE) || ebiten.IsKeyPressed(ebiten.KeyPageUp) {
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
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camX += pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camY -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
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

	// Handle player attacking monster
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		worldX := (float64(mx)-float64(g.w/2))/g.camScale + g.camX
		worldY := (float64(my)-float64(g.h/2))/g.camScale - g.camY
		tx, ty := g.isoToCartesian(worldX, worldY)

		cx := int(math.Floor(tx - 1.5))
		cy := int(math.Floor(ty - 0.5))

		for _, m := range g.Monsters {
			if m.IsDead {
				continue
			}
			if m.TileX == cx && m.TileY == cy &&
				entities.IsAdjacentRanged(g.player.TileX, g.player.TileY, m.TileX, m.TileY, 2) &&
				g.player.CanAttack() {
				m.TakeDamage(g.player.Damage, &g.HitMarkers, &g.DamageNumbers)
				g.player.AttackTick = 0
			}
		}
	}
}

func (g *Game) handleHoverTile() {
	mx, my := ebiten.CursorPosition()
	cx := (float64(mx)-float64(g.w/2))/g.camScale + g.camX
	cy := (float64(my)-float64(g.h/2))/g.camScale - g.camY
	tx, ty := g.isoToCartesian(cx, cy)
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
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if l, err := levels.NewDungeonLevel(); err == nil {
			g.currentLevel = l
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if l, err := levels.NewForestLevel(); err == nil {
			g.currentLevel = l
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		//g.currentLevel = levels.NewLevel1()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.ShowRays = !g.ShowRays
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		g.FullBright = !g.FullBright
	}
	if g.State == StateGameOver && ebiten.IsKeyPressed(ebiten.KeyV) {
		newGame, _ := NewGame()
		*g = *newGame
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
		g.pauseMenu.Update()
		return
	}

	g.handleZoom()
	g.handlePan()
	g.handlePlayerVelocity()
	g.handleHoverTile()
	g.handleClicks()
	g.handleLevelHotkeys()
}

func (g *Game) handlePlayerVelocity() {
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1
		g.player.LeftFacing = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1
		g.player.LeftFacing = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		dy -= 1
		g.player.LeftFacing = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		dy += 1
		g.player.LeftFacing = true
	}

	// Only enable velocity mode if a direction is pressed
	if dx != 0 || dy != 0 {
		g.player.MoveController.SetVelocityFromInput(dx, dy)
		g.player.MoveController.Mode = movement.VelocityMode
	} else if g.player.MoveController.Mode == movement.VelocityMode {
		// Stop smoothly when keys released
		g.player.MoveController.Stop()
	}
}

func (g *Game) handleInputGameOver() {
	g.handleLevelHotkeys()
}
