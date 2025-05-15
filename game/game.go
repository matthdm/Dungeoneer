package game

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	w, h         int
	currentLevel *levels.Level

	camX, camY           float64
	camScale, camScaleTo float64
	mousePanX, mousePanY int

	offscreen              *ebiten.Image
	hoverTileX, hoverTileY int
	spriteSheet            *sprites.SpriteSheet
	highlightImage         *ebiten.Image
	editor                 *leveleditor.Editor
	player                 *entities.Player
	Monsters               []*entities.Monster
}

func NewGame() (*Game, error) {
	l, err := levels.NewDungeonLevel()
	if err != nil {
		return nil, fmt.Errorf("failed to create new level: %s", err)
	}
	ss, err := sprites.LoadSpriteSheet(l.TileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load sprite sheet: %s", err)
	}

	return &Game{
		currentLevel:   levels.NewLevel1(),
		camScale:       1,
		camScaleTo:     1,
		mousePanX:      math.MinInt32,
		mousePanY:      math.MinInt32,
		spriteSheet:    ss,
		highlightImage: ss.Cursor,
		editor:         leveleditor.NewEditor(),
		player:         entities.NewPlayer(ss),
		Monsters:       entities.NewMonster(ss),
	}, nil
}

// cartesianToIso transforms cartesian coordinates into isometric coordinates.
func (g *Game) cartesianToIso(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.TileSize
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}

//This function might be useful for those who want to modify this example.

// isoToCartesian transforms isometric coordinates into cartesian coordinates.
func (g *Game) isoToCartesian(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.TileSize
	cx := (x/float64(tileSize/2) + y/float64(tileSize/4)) / 2
	cy := (y/float64(tileSize/4) - (x / float64(tileSize/2))) / 2
	return cx, cy
}

func (g *Game) Update() error {
	g.handleZoom()
	g.handlePan()
	g.updateHoverTile()

	// Level switching
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		l, err := levels.NewDungeonLevel()
		if err != nil {
			return err
		}
		g.currentLevel = l
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		l, err := levels.NewForestLevel()
		if err != nil {
			return err
		}
		g.currentLevel = l
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.currentLevel = levels.NewLevel1()
	}

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
				entities.IsAdjacent(g.player.TileX, g.player.TileY, m.TileX, m.TileY) &&
				g.player.CanAttack() {
				m.TakeDamage(g.player.Damage)
				g.player.AttackTick = 0
			}
		}
	}
	// Player path preview
	if g.player != nil {
		path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, g.hoverTileX, g.hoverTileY)
		g.player.PathPreview = path
	}

	// Update game objects
	g.player.Update(g.currentLevel)
	for _, m := range g.Monsters {
		m.Update(g.player, g.currentLevel)
	}

	// Optional: Level editor
	g.DebugLevelEditor()
	return nil
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
	g.camScaleTo = math.Max(0.01, math.Min(100, g.camScaleTo))

	div := 10.0
	if g.camScaleTo > g.camScale {
		g.camScale += (g.camScaleTo - g.camScale) / div
	} else {
		g.camScale -= (g.camScale - g.camScaleTo) / div
	}
}

func (g *Game) handlePan() {
	pan := 7.0 / g.camScale
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.camX -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.camX += pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.camY -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.camY += pan
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
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

func (g *Game) updateHoverTile() {
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
}

func (g *Game) handleClicks() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Movement
		tx, ty := g.hoverTileX, g.hoverTileY
		if g.player.CanMoveTo(tx, ty, g.currentLevel) {
			g.player.Path = pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, tx, ty)
			g.player.PathPreview = nil
		}

		// Combat
		mx, my := ebiten.CursorPosition()
		worldX := (float64(mx)-float64(g.w/2))/g.camScale + g.camX
		worldY := (float64(my)-float64(g.h/2))/g.camScale - g.camY
		fx, fy := g.isoToCartesian(worldX, worldY)
		cx := int(math.Floor(fx - 1.5))
		cy := int(math.Floor(fy - 0.5))

		for _, m := range g.Monsters {
			if !m.IsDead && m.TileX == cx && m.TileY == cy &&
				entities.IsAdjacent(g.player.TileX, g.player.TileY, m.TileX, m.TileY) &&
				g.player.CanAttack() {
				m.TakeDamage(g.player.Damage)
				g.player.AttackTick = 0
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	cx, cy := float64(g.w/2), float64(g.h/2)

	// Decide target surface and scale mode
	scaleLater := g.camScale > 1
	target := screen
	scale := g.camScale

	if scaleLater {
		target = g.getOrCreateOffscreen(screen.Bounds().Size())
		target.Clear()
		scale = 1
	}

	g.drawTiles(target, scale, cx, cy)
	g.drawHoverTile(target, scale, cx, cy)
	g.drawPathPreview(target, scale, cx, cy)
	g.drawEntities(target, scale, cx, cy)

	// Draw offscreen buffer if used
	if scaleLater {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Scale(g.camScale, g.camScale)
		op.GeoM.Translate(cx, cy)
		screen.DrawImage(target, op)
	}

	// Debug info
	ebitenutil.DebugPrint(screen, fmt.Sprintf(constants.DEBUG_TEMPLATE, ebiten.ActualFPS(), ebiten.ActualTPS(), g.camScale, g.camX, g.camY))
}

func (g *Game) getOrCreateOffscreen(size image.Point) *ebiten.Image {
	if g.offscreen != nil && g.offscreen.Bounds().Size() == size {
		return g.offscreen
	}
	if g.offscreen != nil {
		g.offscreen.Deallocate()
	}
	g.offscreen = ebiten.NewImage(size.X, size.Y)
	return g.offscreen
}

func (g *Game) drawTiles(target *ebiten.Image, scale, cx, cy float64) {
	op := &ebiten.DrawImageOptions{}
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

			op.GeoM.Reset()
			op.GeoM.Translate(xi, yi)
			op.GeoM.Translate(-g.camX, g.camY)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx, cy)

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
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(xi, yi)
	op.GeoM.Translate(-g.camX, g.camY)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx, cy)

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
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(xi, yi)
		op.GeoM.Translate(-g.camX, g.camY)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, cy)

		op.ColorScale.Scale(1, 1, 1, 0.4) // semi-transparent
		target.DrawImage(g.spriteSheet.Cursor, op)
	}
}

func (g *Game) drawEntities(target *ebiten.Image, scale, cx, cy float64) {
	tileSize := g.currentLevel.TileSize

	if g.player != nil {
		g.player.Draw(target, tileSize, func(x, y int) (float64, float64) {
			return g.cartesianToIso(float64(x), float64(y))
		}, g.camX, g.camY, scale, cx, cy)
	}

	for _, m := range g.Monsters {
		m.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
	}
}

// Layout is called when the Game's layout changes.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w, g.h = outsideWidth, outsideHeight
	return g.w, g.h
}
