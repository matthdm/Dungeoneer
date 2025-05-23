package game

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	w, h         int
	currentLevel *levels.Level
	State        GameState
	isPaused     bool

	camX, camY           float64
	minCamScale          float64
	camScale, camScaleTo float64
	mousePanX, mousePanY int

	offscreen              *ebiten.Image
	hoverTileX, hoverTileY int
	spriteSheet            *sprites.SpriteSheet
	highlightImage         *ebiten.Image
	editor                 *leveleditor.Editor
	player                 *entities.Player
	Monsters               []*entities.Monster
	HitMarkers             []entities.HitMarker
	DamageNumbers          []entities.DamageNumber

	RaycastWalls             []fov.Line
	ShowRays                 bool
	lastPlayerX, lastPlayerY float64
	cachedRays               []fov.Line
	FullBright               bool

	SeenTiles [][]bool
}

type GameState int

const (
	StatePlaying GameState = iota
	StateGameOver
	StateMenu
)

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
		currentLevel:   l, //levels.NewLevel1(),
		isPaused:       false,
		camScale:       1,
		camScaleTo:     1,
		minCamScale:    0.12,
		mousePanX:      math.MinInt32,
		mousePanY:      math.MinInt32,
		spriteSheet:    ss,
		highlightImage: ss.Cursor,
		editor:         leveleditor.NewEditor(),
		player:         entities.NewPlayer(ss),
		Monsters:       entities.NewStatueMonster(ss),
		RaycastWalls:   fov.LevelToWalls(l), //levels.NewLevel1()),
	}, nil
}

// cartesianToIso transforms cartesian coordinates into isometric coordinates.
func (g *Game) cartesianToIso(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.TileSize
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}

func (g *Game) UpdateSeenTiles(level levels.Level) {
	seen := make([][]bool, level.H)
	for y := range seen {
		seen[y] = make([]bool, level.W)
	}
	g.SeenTiles = seen
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
	g.handlePauseToggle()

	if g.isPaused {
		return nil
	}
	if g.player != nil && g.player.IsDead {
		g.State = StateGameOver
	}
	if g.State == StateGameOver {
		g.handleLevelHotkeys()
		return nil // Skip updates while in game over
	}

	g.UpdateSeenTiles(*g.currentLevel)
	g.handleZoom()
	g.handlePan()
	g.handleHoverTile()
	g.handleClicks()
	g.handleLevelHotkeys()

	//raycast update
	g.RaycastWalls = fov.LevelToWalls(g.currentLevel)
	// Raycast update — only when player moves
	if g.player != nil && (g.player.InterpX != g.lastPlayerX || g.player.InterpY != g.lastPlayerY) {
		// Use player world-space position directly
		originX := g.player.InterpX
		originY := g.player.InterpY

		g.cachedRays = fov.RayCasting(originX, originY, g.RaycastWalls, g.currentLevel)
		g.lastPlayerX = originX
		g.lastPlayerY = originY

		for _, ray := range g.cachedRays {
			tx := int(ray.X2)
			ty := int(ray.Y2)
			if tx >= 0 && ty >= 0 && ty < len(g.SeenTiles) && tx < len(g.SeenTiles[0]) {
				g.SeenTiles[ty][tx] = true
			}
		}
	}

	//Player Movement
	if g.player != nil {
		// Prevent re-calculating paths while mid-movement
		path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, g.hoverTileX, g.hoverTileY)
		g.player.PathPreview = path
	}

	// Update game objects
	g.player.Update(g.currentLevel)

	//Update Monsters
	for _, m := range g.Monsters {
		m.Update(g.player, g.currentLevel)
	}

	// Optional: Level editor
	g.DebugLevelEditor()

	//Hit markers
	g.handleHitMarkers()
	g.handleDamageNumbers()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	cx, cy := float64(g.w/2), float64(g.h/2)

	// Prepare render target
	scaleLater := g.camScale > 1
	target := screen
	scale := g.camScale
	if scaleLater {
		target = g.getOrCreateOffscreen(screen.Bounds().Size())
		target.Clear()
		scale = 1
	}

	// World drawing
	g.drawTiles(target, scale, cx, cy)
	g.drawPathPreview(target, scale, cx, cy)
	g.drawEntities(target, scale, cx, cy)
	g.drawHoverTile(target, scale, cx, cy)

	// Draw to screen (if upscaled)
	if scaleLater {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Scale(g.camScale, g.camScale)
		op.GeoM.Translate(cx, cy)
		screen.DrawImage(target, op)
	}

	// === SHADOWS / FOV ===
	if g.player != nil && len(g.cachedRays) > 0 && !g.FullBright {
		fov.DrawShadows(
			screen,
			g.cachedRays,
			g.camX, g.camY,
			g.camScale,
			cx, cy,
			g.currentLevel.TileSize,
		)
	}

	if g.ShowRays && len(g.cachedRays) > 0 {
		fov.DebugDrawRays(
			screen,
			g.cachedRays,
			g.camX, g.camY,
			g.camScale,
			cx, cy,
			g.currentLevel.TileSize,
		)
	}

	fov.DebugDrawWalls(screen, g.RaycastWalls, g.camX, g.camY, g.camScale, cx, cy, g.currentLevel.TileSize)

	if g.State == StateGameOver {
		msg := "GAME OVER - Press V to Restart"
		ebitenutil.DebugPrintAt(screen, msg, g.w/2-100, g.h/2)
	}
	if g.isPaused {
		ebitenutil.DebugPrintAt(target, "PAUSED", (g.w/2)-25, 0)
	}

	// Debug UI
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

// Layout is called when the Game's layout changes.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w, g.h = outsideWidth, outsideHeight
	fov.ResizeShadowBuffer(g.w, g.h) // 🛠️ Ensures buffer is always the correct size
	return g.w, g.h
}
