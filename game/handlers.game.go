package game

import (
	"dungeoneer/entities"
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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
				g.player.Path = path
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
		g.currentLevel = levels.NewLevel1()
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
