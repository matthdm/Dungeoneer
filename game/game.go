package game

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/hud"
	"dungeoneer/items"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/menumanager"
	"dungeoneer/pathing"
	"dungeoneer/spells"
	"dungeoneer/sprites"
	"dungeoneer/ui"
	"fmt"
	"image"
	"math"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	w, h           int
	currentWorld   *levels.LayeredLevel
	currentLevel   *levels.Level
	State          GameState
	Menu           *ui.MainMenu
	PauseMenu      *ui.PauseMenu
	LoadLevelMenu  *ui.LoadLevelMenu
	LoadPlayerMenu *ui.LoadPlayerMenu
	SaveLevelMenu  *ui.SaveLevelMenu
	SavePrompt     *ui.TextInputMenu
	LinkPrompt     *ui.LayerPrompt
	isPaused       bool

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
	HealNumbers            []entities.DamageNumber

	DevMenu *ui.DevMenu
	HUD     *hud.HUD

	ActiveSpells    []spells.Spell
	fireballSprites [][]*ebiten.Image

	SpellDebug bool

	RaycastWalls             []fov.Line
	ShowRays                 bool
	lastPlayerX, lastPlayerY float64
	cachedRays               []fov.Line
	FullBright               bool

	// Visibility tracking
	VisibleTiles [][]bool // true if currently visible
	SeenTiles    [][]bool // true if ever seen
	camSmooth    float64

	// cooldown timer to prevent immediate re-triggering of stair links
	layerSwitchCooldown float64
}

// editorLayerChanged updates the game when the level editor switches layers.
func (g *Game) editorLayerChanged(l *levels.Level) {
	if l == nil || g.currentWorld == nil {
		return
	}

	idx := g.currentWorld.ActiveIndex
	entry := levels.Point{}
	if g.player != nil {
		entry = levels.Point{X: g.player.TileX, Y: g.player.TileY}
	}
	g.switchLayer(idx, entry)
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
	fbSprites, err := spells.LoadFireballSprites()
	if err != nil {
		return nil, err
	}
	l := levels.CreateNewBlankLevel(64, 64, 64, ss)
	world := levels.NewLayeredLevel(l)

	//This is needed for save/loading levels
	leveleditor.RegisterSprites(ss)
	// Load wall sprite sheets for all available flavors
	for _, fl := range sprites.WallFlavors {
		if wss, err := sprites.LoadWallSpriteSheet(fl); err == nil {
			leveleditor.RegisterWallSprites(wss)
		}
	}

	if err := items.LoadDefaultItems(); err != nil {
		return nil, err
	}

	g := &Game{
		currentWorld:    world,
		currentLevel:    l,
		isPaused:        false,
		camScale:        1,
		camScaleTo:      1,
		minCamScale:     0.12,
		mousePanX:       math.MinInt32,
		mousePanY:       math.MinInt32,
		spriteSheet:     ss,
		highlightImage:  ss.Cursor,
		editor:          leveleditor.NewLayeredEditor(world, 640, 480),
		FullBright:      true,
		player:          entities.NewPlayer(ss),
		Monsters:        []*entities.Monster{},
		fireballSprites: fbSprites,
		ActiveSpells:    []spells.Spell{},
		RaycastWalls:    fov.LevelToWalls(l),
		State:           StateMainMenu,
		DeltaTime:       1.0 / 60.0,
		camSmooth:       0.1,
		SpellDebug:      true,
	}
	g.DevMenu = ui.NewDevMenu(640, 480, g.player)
	g.editor.OnLayerChange = g.editorLayerChanged
	g.editor.OnStairPlaced = g.stairPlaced
	g.spawnEntitiesFromLevel()
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
			g.UpdateSeenTiles(*loaded)
			g.spawnEntitiesFromLevel()
		},
		func() {
			menumanager.Manager().CloseActiveMenu()
		},
	)
	g.LoadPlayerMenu = ui.NewLoadPlayerMenu(g.w, g.h,
		func(ply *entities.Player) {
			g.player = ply
		},
		func() {
			menumanager.Manager().CloseActiveMenu()
		},
	)
	// Pause Menu
	pm := ui.NewPauseMenu(l.W, l.H, ui.PauseMenuCallbacks{
		OnResume:     func() { g.resumeGame() },
		OnExit:       func() { os.Exit(0) },
		OnLoadLevel:  func() { menumanager.Manager().Open(g.LoadLevelMenu) },
		OnLoadPlayer: func() { menumanager.Manager().Open(g.LoadPlayerMenu) },
		OnSavePlayer: func() {
			menuRect := image.Rect(g.w/2-200, g.h/2-100, g.w/2+200, g.h/2+100)
			g.SavePrompt = ui.NewTextInputMenu(
				menuRect,
				"Save Player",
				"Enter filename (with .json):",
				func(filename string) {
					path := "players/" + filename
					err := entities.SavePlayerToFile(g.player, path)
					if err != nil {
						fmt.Println("Error saving player:", err)
					} else {
						fmt.Println("Saved player to:", path)
						g.SavePrompt = ui.NewTextInputMenu(
							menuRect,
							"Success",
							"Saved player to: "+filename,
							nil,
							nil,
						)
						g.SavePrompt.Instructions = []string{"Press Esc to close"}
						menumanager.Manager().Open(g.SavePrompt)
					}
				},
				func() {
					fmt.Println("Canceled saving player.")
					menumanager.Manager().CloseActiveMenu()
				},
			)
			menumanager.Manager().Open(g.SavePrompt)
		},
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
						menumanager.Manager().Open(g.SavePrompt)
					}
				},
				func() {
					fmt.Println("Canceled saving.")
					menumanager.Manager().CloseActiveMenu()
				},
			)
			menumanager.Manager().Open(g.SavePrompt)
		},
		OnNewBlank: func() {
			newLevel := levels.CreateNewBlankLevel(64, 64, g.currentLevel.TileSize, ss) // TODO: Prompt for dimensions later
			newWorld := levels.NewLayeredLevel(newLevel)
			g.currentWorld = newWorld
			g.currentLevel = newLevel
			g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
			g.editor.OnLayerChange = g.editorLayerChanged
			g.editor.OnStairPlaced = g.stairPlaced
			g.UpdateSeenTiles(*newLevel)
		},
	})
	g.PauseMenu = pm
	menumanager.Init(pm)

	g.VisibleTiles = make([][]bool, g.currentLevel.H)
	g.SeenTiles = make([][]bool, g.currentLevel.H)
	for y := range g.VisibleTiles {
		g.VisibleTiles[y] = make([]bool, g.currentLevel.W)
		g.SeenTiles[y] = make([]bool, g.currentLevel.W)
	}

	g.editor.Active = true // or toggle with key

	g.HUD = hud.New()
	tomeNames := []string{"Red Tome", "Teal Tome", "Blue Tome", "Verdant Tome", "Crypt Tome"}
	for i, name := range tomeNames {
		for _, tmpl := range items.Registry {
			if tmpl.Name == name {
				if i < len(g.HUD.SkillSlots) {
					g.HUD.SkillSlots[i].Icon = tmpl.Icon
				}
				break
			}
		}
	}

	return g, nil
}

// cartesianToIso transforms cartesian coordinates into isometric coordinates.
func (g *Game) cartesianToIso(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.TileSize
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}

// switchLayer activates the given layer and moves the player to the entry tile.
func (g *Game) switchLayer(index int, entry levels.Point) {
	orig := entry
	g.currentWorld.SwitchToLayer(index, entry)
	g.currentLevel = g.currentWorld.ActiveLayer()
	if g.currentLevel == nil {
		return
	}

	// ensure entry is within bounds of new level
	if entry.X < 0 || entry.Y < 0 || entry.X >= g.currentLevel.W || entry.Y >= g.currentLevel.H {
		entry.X = g.currentLevel.W / 2
		entry.Y = g.currentLevel.H / 2
		// update any links that targeted the out-of-bounds tile so the
		// descending stair in the new layer correctly links back
		for _, link := range g.currentWorld.Stairwells {
			if link.ToLayerIndex == index && link.ToTile == orig {
				link.ToTile = entry
			}
			if link.FromLayerIndex == index && link.FromTile == orig {
				link.FromTile = entry
			}
		}
	}
	if g.player != nil {
		g.player.TileX = entry.X
		g.player.TileY = entry.Y
		g.player.MoveController.InterpX = float64(entry.X)
		g.player.MoveController.InterpY = float64(entry.Y)
		g.player.MoveController.Path = nil
		g.player.MoveController.Stop()
	}

	if g.editor != nil {
		// Update the editor's active layer without triggering its
		// callback to avoid a recursive loop back into this method.
		g.editor.SetActiveLayerSilently(index)
	}
	g.layerSwitchCooldown = 1.0
	g.UpdateSeenTiles(*g.currentLevel)
	g.spawnEntitiesFromLevel()
	g.cachedRays = nil
	g.camX, g.camY = 0, 0
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
	vis := make([][]bool, level.H)
	for y := range seen {
		seen[y] = make([]bool, level.W)
		vis[y] = make([]bool, level.W)
	}
	g.SeenTiles = seen
	g.VisibleTiles = vis
}

// spawnEntitiesFromLevel creates monsters based on the placed entities in the
// currently loaded level.
func (g *Game) spawnEntitiesFromLevel() {
	g.Monsters = []*entities.Monster{}
	for _, ent := range g.currentLevel.Entities {
		// skip invalid coordinates to avoid crashes
		if ent.X < 0 || ent.Y < 0 || ent.X >= g.currentLevel.W || ent.Y >= g.currentLevel.H {
			continue
		}
		switch ent.Type {
		case "AmbushMonster":
			meta, ok := leveleditor.SpriteRegistry[ent.SpriteID]
			if !ok {
				continue
			}
			m := entities.CreateAmbushMonster(meta.Image, ent.X, ent.Y)
			g.Monsters = append(g.Monsters, m)
		}
	}
}

// stairPlaced is triggered by the editor when a stairwell sprite is placed.
func (g *Game) stairPlaced(x, y int, spriteID string) {
	if g.LinkPrompt != nil && g.LinkPrompt.IsVisible() {
		return
	}
	g.LinkPrompt = ui.NewLayerPrompt(g.w, g.h, sprites.WallFlavors,
		func(wd, ht int, flavor string) {
			wss, err := sprites.LoadWallSpriteSheet(flavor)
			if err != nil {
				g.LinkPrompt = nil
				return
			}
			newL := levels.CreateNewBlankLevelWithFloor(wd, ht, g.currentLevel.TileSize, flavor+"_floor", wss.Floor)
			g.currentWorld.AddLayer(newL)
			newIdx := len(g.currentWorld.Layers) - 1
			counterID := "StairsAscending"
			if strings.Contains(strings.ToLower(spriteID), "ascending") {
				counterID = "StairsDecending"
			}
			destX, destY := x, y
			if x < 0 || y < 0 || x >= newL.W || y >= newL.H {
				destX = newL.W / 2
				destY = newL.H / 2
			}
			if meta, ok := leveleditor.SpriteRegistry[counterID]; ok {
				if t := newL.Tile(destX, destY); t != nil {
					t.AddSpriteByID(counterID, meta.Image)
					t.IsWalkable = meta.IsWalkable
				}
			}
			g.currentWorld.Stairwells = append(g.currentWorld.Stairwells, &levels.LayerLink{
				FromLayerIndex: g.currentWorld.ActiveIndex,
				FromTile:       levels.Point{X: x, Y: y},
				ToLayerIndex:   newIdx,
				ToTile:         levels.Point{X: destX, Y: destY},
				TriggerSprite:  spriteID,
			})
			g.currentWorld.Stairwells = append(g.currentWorld.Stairwells, &levels.LayerLink{
				FromLayerIndex: newIdx,
				FromTile:       levels.Point{X: x, Y: y},
				ToLayerIndex:   g.currentWorld.ActiveIndex,
				ToTile:         levels.Point{X: destX, Y: destY},
				TriggerSprite:  counterID,
			})
			g.LinkPrompt = nil
		}, func() { g.LinkPrompt = nil })
	g.LinkPrompt.Show()
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
		} else if g.LoadPlayerMenu != nil && g.LoadPlayerMenu.Menu.IsVisible() {
			g.LoadPlayerMenu.Update()
		} else {
			g.PauseMenu.Update()
		}
		return nil
	}

	if g.LinkPrompt != nil && g.LinkPrompt.IsVisible() {
		g.LinkPrompt.Update()
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

	// Reduce stair switch cooldown
	if g.layerSwitchCooldown > 0 {
		g.layerSwitchCooldown -= g.DeltaTime
	}

	// Path preview update (mouse hover A*)
	if g.player != nil {
		path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, g.hoverTileX, g.hoverTileY)
		g.player.PathPreview = path
		g.player.Update(g.currentLevel, g.DeltaTime)
		g.updateCameraFollow()

		if g.HUD != nil {
			g.HUD.HealthPercent = float64(g.player.HP) / float64(g.player.MaxHP)
			g.HUD.ManaPercent = float64(g.player.Mana) / float64(g.player.MaxMana)
			g.HUD.DashCharges = g.player.DashCharges
			maxCD := 0.0
			for _, cd := range g.player.DashCooldowns {
				if cd > maxCD {
					maxCD = cd
				}
			}
			g.HUD.DashCooldown = maxCD
		}
		// Check layer links
		if g.layerSwitchCooldown <= 0 {
			for _, link := range g.currentWorld.Stairwells {
				if link.FromLayerIndex == g.currentWorld.ActiveIndex &&
					link.FromTile.X == g.player.TileX && link.FromTile.Y == g.player.TileY {
					g.switchLayer(link.ToLayerIndex, link.ToTile)
					break
				}
			}
		}
	}

	// Monsters
	for _, m := range g.Monsters {
		m.Update(g.player, g.currentLevel)
	}

	g.updateSpells()

	if g.DevMenu != nil {
		g.DevMenu.Update()
	}

	// Debugging / editor / effects
	g.DebugLevelEditor()
	g.handleHitMarkers()
	g.handleDamageNumbers()
	g.handleHealNumbers()
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

	if g.LoadPlayerMenu != nil {
		g.LoadPlayerMenu.SetRect(newRect)
	}

	if g.SaveLevelMenu != nil {
		g.SaveLevelMenu.SetRect(newRect)
	}

	if g.editor == nil {
		g.editor = leveleditor.NewLayeredEditor(g.currentWorld, g.w, g.h)
		g.editor.OnLayerChange = g.editorLayerChanged
		g.editor.OnStairPlaced = g.stairPlaced
	}

	return g.w, g.h
}
func (g *Game) isValidTile(x, y int) bool {
	return x >= 0 && x < g.currentLevel.W && y >= 0 && y < g.currentLevel.H
}
