package game

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"dungeoneer/ui"
	"fmt"
	"image"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	w, h          int
	currentLevel  *levels.Level
	State         GameState
	Menu          *ui.MainMenu
	PauseMenu     *ui.PauseMenu
	LoadLevelMenu *ui.LoadLevelMenu
	SaveLevelMenu *ui.SaveLevelMenu
	SavePrompt    *ui.TextInputMenu
	isPaused      bool

	camX, camY           float64
	minCamScale          float64
	camScale, camScaleTo float64
	mousePanX, mousePanY int
	DeltaTime            float64

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

	// Visibility tracking
	VisibleTiles [][]bool // true if currently visible
	SeenTiles    [][]bool // true if ever seen
	camSmooth    float64
}

type GameState int

const (
	StateMainMenu GameState = iota
	StatePlaying
	StateGameOver
)

func NewGame() (*Game, error) {
	ss, err := sprites.LoadSpriteSheet(constants.DefaultTileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load sprite sheet: %s", err)
	}
	l := levels.CreateNewBlankLevel(64, 64, 64, ss)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create new level: %s", err)
	//}
	//This is needed for save/loading levels
	leveleditor.RegisterSprites(ss)

	g := &Game{
		currentLevel:   l,
		isPaused:       false,
		camScale:       1,
		camScaleTo:     1,
		minCamScale:    0.12,
		mousePanX:      math.MinInt32,
		mousePanY:      math.MinInt32,
		spriteSheet:    ss,
		highlightImage: ss.Cursor,
		editor:         leveleditor.NewEditor(l, 640, 480),
		FullBright:     true,
		player:         entities.NewPlayer(ss),
		Monsters:       entities.NewStatueMonster(ss),
		RaycastWalls:   fov.LevelToWalls(l),
		State:          StateMainMenu,
		DeltaTime:      1.0 / 60.0,
		camSmooth:      0.1,
	}
	mm, err := ui.NewMainMenu()
	if err != nil {
		return nil, fmt.Errorf("failed create new main menu: %s", err)
	}
	// Main Menu
	g.Menu = mm
	// Load Level Menu
	g.LoadLevelMenu = ui.NewLoadLevelMenu(g.w, g.h,
		func(loaded *levels.Level) {
			g.currentLevel = loaded
			g.isPaused = false
			g.UpdateSeenTiles(*loaded)
		},
		func() {
			g.LoadLevelMenu.Hide()
			g.PauseMenu.Show() // <-- resume the previous menu
		},
	)
	// Pause Menu
	pm := ui.NewPauseMenu(l.W, l.H, ui.PauseMenuCallbacks{
		OnResume:    func() { g.resumeGame() },
		OnExit:      func() { os.Exit(0) },
		OnLoadLevel: func() { g.LoadLevelMenu.Show() },
		OnSaveLevel: func() {
			menuRect := image.Rect(g.w/2-200, g.h/2-100, g.w/2+200, g.h/2+100)
			g.SavePrompt = ui.NewTextInputMenu(
				menuRect,
				"Save Level",
				"Enter filename (with .json):",
				func(filename string) {
					path := "levels/" + filename
					err := leveleditor.SaveLevelToFile(g.currentLevel, path)
					if err != nil {
						fmt.Println("Error saving level:", err)
					} else {
						fmt.Println("Saved to:", path)
						// Show confirmation popup for 2 seconds
						g.SavePrompt = ui.NewTextInputMenu(
							menuRect,
							"Success",
							"Saved level to: "+filename,
							nil,
							nil,
						)
						g.SavePrompt.Instructions = []string{"Press Esc to close"}
						g.SavePrompt.Show()
					}
				},
				func() {
					fmt.Println("Canceled saving.")
				},
			)
			g.SavePrompt.Show()
		},
		OnNewBlank: func() {
			newLevel := levels.CreateNewBlankLevel(64, 64, g.currentLevel.TileSize, ss) // TODO: Prompt for dimensions later
			g.currentLevel = newLevel
			g.editor = leveleditor.NewEditor(newLevel, g.w, g.h)
			g.UpdateSeenTiles(*newLevel)
		},
	})
	g.PauseMenu = pm

	g.VisibleTiles = make([][]bool, g.currentLevel.H)
	g.SeenTiles = make([][]bool, g.currentLevel.H)
	for y := range g.VisibleTiles {
		g.VisibleTiles[y] = make([]bool, g.currentLevel.W)
		g.SeenTiles[y] = make([]bool, g.currentLevel.W)
	}

	g.editor.Active = true // or toggle with key

	return g, nil
}

// cartesianToIso transforms cartesian coordinates into isometric coordinates.
func (g *Game) cartesianToIso(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.TileSize
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}

func (g *Game) screenToTile() (int, int) {
	cx, cy := ebiten.CursorPosition()
	worldX := float64(cx)/g.camScale + g.camX
	worldY := float64(cy)/g.camScale + g.camY
	tileX := int(worldX) / g.currentLevel.TileSize
	tileY := int(worldY) / g.currentLevel.TileSize
	return tileX, tileY
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
	switch g.State {
	case StateMainMenu:
		return g.updateMainMenu()
	case StateGameOver:
		return g.updateGameOver()
	case StatePlaying:
		if g.player != nil && g.player.IsDead {
			g.State = StateGameOver
			return nil
		}
		return g.updatePlaying()
	default:
		return nil
	}
}

// updateCameraFollow centers the camera on the player's interpolated position
// while the player is moving. When the player stops, the camera stays put.
func (g *Game) updateCameraFollow() {
	if g.player == nil {
		return
	}
	mc := g.player.MoveController
	moving := mc.Moving || len(mc.Path) > 0 || mc.VelocityX != 0 || mc.VelocityY != 0
	if !moving {
		return
	}

	isoX, isoY := g.cartesianToIso(mc.InterpX, mc.InterpY)
	targetX := isoX
	targetY := -isoY

	g.camX += (targetX - g.camX) * g.camSmooth
	g.camY += (targetY - g.camY) * g.camSmooth

	// Clamp camera within world bounds
	worldWidth := float64(g.currentLevel.W*g.currentLevel.TileSize) / 2
	worldHeight := float64(g.currentLevel.H*g.currentLevel.TileSize) / 2
	if g.camX < -worldWidth {
		g.camX = -worldWidth
	} else if g.camX > worldWidth {
		g.camX = worldWidth
	}
	if g.camY < -worldHeight {
		g.camY = -worldHeight
	} else if g.camY > 0 {
		g.camY = 0
	}
}

func (g *Game) updateMainMenu() error {
	g.Menu.Update()
	g.handleMainMenuInput()
	return nil
}

func (g *Game) updateGameOver() error {
	g.handleLevelHotkeys()
	return nil
}

func (g *Game) updatePlaying() error {
	// Pause handling first
	g.handlePause()
	if g.isPaused {
		if g.SavePrompt != nil && g.SavePrompt.IsVisible() {
			g.SavePrompt.Update()
		} else if g.LoadLevelMenu != nil && g.LoadLevelMenu.Menu.IsVisible() {
			g.LoadLevelMenu.Update()
		} else {
			g.PauseMenu.Update()
		}
		return nil
	}

	if g.editor.Active {
		g.editor.Update(g.screenToTile)
	}

	// Handle controls (mouse, keys, etc.)
	g.handleInput()

	// Update seen/visible memory arrays based on current level size
	//g.UpdateSeenTiles(*g.currentLevel)

	// Rebuild raycast walls (in case the level changed)
	g.RaycastWalls = fov.LevelToWalls(g.currentLevel)

	// Determine if we should update rays:
	shouldRecast := len(g.cachedRays) == 0 ||
		g.player.MoveController.InterpX != g.lastPlayerX ||
		g.player.MoveController.InterpY != g.lastPlayerY

	if g.player != nil && shouldRecast {
		originX := g.player.MoveController.InterpX
		originY := g.player.MoveController.InterpY

		g.cachedRays = fov.RayCasting(originX, originY, g.RaycastWalls, g.currentLevel)
		g.lastPlayerX = originX
		g.lastPlayerY = originY

		// Clear visibility map
		for y := range g.VisibleTiles {
			for x := range g.VisibleTiles[y] {
				g.VisibleTiles[y][x] = false
			}
		}

		// Update visibility from ray paths
		for _, ray := range g.cachedRays {
			for _, pt := range ray.Path {
				if g.isValidTile(pt.X, pt.Y) {
					g.VisibleTiles[pt.Y][pt.X] = true
					g.SeenTiles[pt.Y][pt.X] = true
				}
			}

			// Final ray endpoint tile
			tx := int(ray.X2)
			ty := int(ray.Y2)
			if g.isValidTile(tx, ty) {
				g.VisibleTiles[ty][tx] = true
				g.SeenTiles[ty][tx] = true
			}
		}
	}

	// Path preview update (mouse hover A*)
	if g.player != nil {
		path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, g.hoverTileX, g.hoverTileY)
		g.player.PathPreview = path
		g.player.Update(g.currentLevel, g.DeltaTime)
		g.updateCameraFollow()
	}

	// Monsters
	for _, m := range g.Monsters {
		m.Update(g.player, g.currentLevel)
	}

	// Debugging / editor / effects
	g.DebugLevelEditor()
	g.handleHitMarkers()
	g.handleDamageNumbers()
	return nil
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
	fov.ResizeShadowBuffer(g.w, g.h) // Ensures buffer is always the correct size
	menuWidth := max(300, g.w/2)
	menuHeight := max(300, g.h/2)
	menuX := (g.w - menuWidth) / 2
	menuY := (g.h - menuHeight) / 2
	newRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)
	// Update menu size to grow with screen size
	if g.PauseMenu != nil {

		if g.PauseMenu.MainMenu != nil {
			g.PauseMenu.MainMenu.SetRect(newRect)
		}
		if g.PauseMenu.SettingsMenu != nil {
			g.PauseMenu.SettingsMenu.SetRect(newRect)
		}
	}

	if g.LoadLevelMenu != nil {
		g.LoadLevelMenu.SetRect(newRect)
	}

	if g.SaveLevelMenu != nil {
		g.SaveLevelMenu.SetRect(newRect)
	}

	if g.editor == nil {
		g.editor = leveleditor.NewEditor(g.currentLevel, g.w, g.h)
	}

	return g.w, g.h
}
func (g *Game) isValidTile(x, y int) bool {
	return x >= 0 && x < g.currentLevel.W && y >= 0 && y < g.currentLevel.H
}
