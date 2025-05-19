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
	HitMarkers             []entities.HitMarker
	DamageNumbers          []entities.DamageNumber
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
		Monsters:       entities.NewStatueMonster(ss),
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
	g.handleHoverTile()
	g.handleClicks()
	g.handleLevelHotkeys()

	// Player path preview
	if g.player != nil {
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
	g.drawPathPreview(target, scale, cx, cy)
	g.drawEntities(target, scale, cx, cy)
	g.drawHoverTile(target, scale, cx, cy)

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

// Layout is called when the Game's layout changes.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w, g.h = outsideWidth, outsideHeight
	return g.w, g.h
}
