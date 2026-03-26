package game

import (
	"dungeoneer/constants"
	"dungeoneer/controls"
	"dungeoneer/dialogue"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/hud"
	"dungeoneer/items"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/menumanager"
	"dungeoneer/pathing"
	"dungeoneer/progression"
	"dungeoneer/spells"
	"dungeoneer/sprites"
	"dungeoneer/tiles"
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
	GenerateMenu   *ui.GenerateLevelMenu
	ProcGenMenu    *ui.ProcGenMenu
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
	ItemDrops              []*entities.ItemDrop
	HitMarkers             []entities.HitMarker
	DamageNumbers          []entities.DamageNumber
	HealNumbers            []entities.DamageNumber
	ExitEntity             *entities.ExitEntity

	// Run loop
	RunState    *RunState
	Meta        *MetaSave
	IsInHub     bool
	hubPortalX  int
	hubPortalY  int

	DevMenu         *ui.DevMenu
	DevTools        *ui.DevOverlay
	HUD             *hud.HUD
	ShowHUD         bool
	HeroPanel       *ui.HeroPanel
	InventoryScreen *ui.InventoryScreen
	ControlsMenu    *ui.ControlsMenu

	Controls *controls.Controls

	ActiveSpells    []spells.Spell
	ActiveSpray     *spells.ArcaneSpray // currently channeled spray (nil if none)
	fireballSprites [][]*ebiten.Image

	SpellDebug bool
	GodMode    bool // infinite HP
	InfMana    bool // infinite mana

	RaycastWalls             []fov.Line
	ShowRays                 bool
	ShowWalls                bool
	lastPlayerX, lastPlayerY float64
	cachedRays               []fov.Line
	FullBright               bool
	ShowThroatValid          bool
	ShowThroatInvalid        bool
	ShowRegionDebug          bool
	ShowDoorDebug            bool

	// Visibility tracking
	// visibleTick[y][x] stores the gameTick when a tile was last hit by a ray.
	// isTileVisible compares against (gameTick - fovDecayFrames) so edge tiles
	// stay lit for a few frames after the ray set shifts, eliminating the
	// clear-then-rebuild flicker from sub-tile origin movement.
	visibleTick [][]int
	gameTick    int
	SeenTiles   [][]bool // true if ever seen
	camSmooth   float64

	// cooldown timer to prevent immediate re-triggering of stair links
	layerSwitchCooldown float64

	hintTimer int
	hint      string
	hintX     int
	hintY     int

	lastPlayerTileX int
	lastPlayerTileY int

	// Phase 2
	MonsterProjectiles []*entities.MonsterProjectile
	SpriteMap          map[string]*ebiten.Image
	FloorCtx           *FloorContext // current floor context for loot/biome access
	CurrentBoss        *entities.Boss
	BossBar            *hud.BossHealthBar
	BossRoom           *levels.Room // arena room on boss floor

	// Phase 3
	NPCs          []*entities.NPC
	DialoguePanel *ui.DialoguePanel
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
	StateDeathScreen
	StateVictoryScreen
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

	// Load dialogue trees from JSON files (non-fatal if directory missing).
	_ = dialogue.LoadAll("dialogues")

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
		NPCs:            []*entities.NPC{},
		fireballSprites: fbSprites,
		ActiveSpells:    []spells.Spell{},
		RaycastWalls:    fov.LevelToWalls(l),
		State:           StateMainMenu,
		DeltaTime:       1.0 / 60.0,
		camSmooth:       0.1,
		SpellDebug:      true,
		Controls:        controls.New(),
	}
	// Load saved control bindings if they exist
	if err := g.Controls.LoadBindings(); err == nil {
		fmt.Println("Loaded saved control bindings")
	}
	g.DevMenu = ui.NewDevMenu(640, 480, g.player, g.ShowHint)
	g.DevTools = ui.NewDevOverlay(640, 480, g.buildDevEntries())
	g.ControlsMenu = ui.NewControlsMenu(640, 480, g.Controls, func() {})
	g.editor.OnLayerChange = g.editorLayerChanged
	g.editor.OnStairPlaced = g.stairPlaced
	g.SpriteMap = BuildSpriteMap(ss)
	g.spawnEntitiesFromLevel()
	g.lastPlayerTileX, g.lastPlayerTileY = g.player.TileX, g.player.TileY
	mm, err := ui.NewMainMenu()
	if err != nil {
		return nil, fmt.Errorf("failed create new main menu: %s", err)
	}
	// Main Menu
	g.Menu = mm
	// Load Level Menu
	g.LoadLevelMenu = ui.NewLoadLevelMenu(g.w, g.h,
		func(loaded *levels.Level) {
			newWorld := levels.NewLayeredLevel(loaded)
			g.currentWorld = newWorld
			g.currentLevel = loaded
			g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
			g.editor.OnLayerChange = g.editorLayerChanged
			g.editor.OnStairPlaced = g.stairPlaced
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
	defaultParams := levels.GenParams{
		Seed:           1,
		Width:          64,
		Height:         64,
		RoomCountMin:   9,
		RoomCountMax:   14,
		RoomWMin:       8,
		RoomWMax:       14,
		RoomHMin:       8,
		RoomHMax:       14,
		CorridorWidth:  1,
		DashLaneMinLen: 7,
		GrappleRange:   10,
		Extras:         2,
		CoverageTarget: 0.42,
		FillerRoomsMax: 5,
		DoorLockChance: 0.35,
	}
	g.ProcGenMenu = ui.NewProcGenMenu(g.w, g.h, defaultParams, func(p levels.GenParams) {
		lvl := levels.Generate64x64(p)
		newWorld := levels.NewLayeredLevel(lvl)
		g.currentWorld = newWorld
		g.currentLevel = lvl
		g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
		g.editor.OnLayerChange = g.editorLayerChanged
		g.editor.OnStairPlaced = g.stairPlaced
		g.UpdateSeenTiles(*lvl)
		menumanager.Manager().CloseActiveMenu()
	}, func() {
		menumanager.Manager().Open(g.GenerateMenu)
	})
	g.GenerateMenu = ui.NewGenerateLevelMenu(g.w, g.h,
		func() {
			newLevel := levels.CreateNewBlankLevel(64, 64, g.currentLevel.TileSize, ss)
			newWorld := levels.NewLayeredLevel(newLevel)
			g.currentWorld = newWorld
			g.currentLevel = newLevel
			g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
			g.editor.OnLayerChange = g.editorLayerChanged
			g.editor.OnStairPlaced = g.stairPlaced
			g.UpdateSeenTiles(*newLevel)
			menumanager.Manager().CloseActiveMenu()
		},
		func() { menumanager.Manager().Open(g.ProcGenMenu) },
		func() { menumanager.Manager().CloseActiveMenu() },
	)
	// Pause Menu — use 640x480 as the initial size; Layout() will resize correctly
	// on the first tick. We must not use l.W/l.H here (those are tile counts, not pixels).
	pm := ui.NewPauseMenu(640, 480, g.Controls, ui.PauseMenuCallbacks{
		OnResume:     func() { g.resumeGame() },
		OnExit:       func() { os.Exit(0) },
		OnLoadLevel:  func() { menumanager.Manager().Open(g.LoadLevelMenu) },
		OnLoadPlayer: func() { menumanager.Manager().Open(g.LoadPlayerMenu) },
		OnGenerate:   func() { menumanager.Manager().Open(g.GenerateMenu) },
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
						menumanager.Manager().CloseActiveMenu()
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
						menumanager.Manager().CloseActiveMenu()
					}
				},
				func() {
					fmt.Println("Canceled saving.")
					menumanager.Manager().CloseActiveMenu()
				},
			)
			menumanager.Manager().Open(g.SavePrompt)
		},
	})
	g.PauseMenu = pm
	menumanager.Init(pm)

	g.visibleTick = make([][]int, g.currentLevel.H)
	g.SeenTiles = make([][]bool, g.currentLevel.H)
	for y := range g.visibleTick {
		g.visibleTick[y] = make([]int, g.currentLevel.W)
		g.SeenTiles[y] = make([]bool, g.currentLevel.W)
	}

	// Load meta progression
	g.Meta = LoadMeta()

	g.editor.Active = true // or toggle with key

	g.HUD = hud.New()
	g.ShowHUD = true
	panelRect := image.Rect(g.w/2-150, g.h/2-150, g.w/2+150, g.h/2+150)
	g.HeroPanel = ui.NewHeroPanel(panelRect, g.player)
	g.InventoryScreen = ui.NewInventoryScreen()

	return g, nil
}

// ShowHint displays a temporary on-screen message.
func (g *Game) ShowHint(msg string) {
	g.hint = msg
	g.hintTimer = 60
	g.hintX = g.w/2 - 50
	g.hintY = g.h - 20
}

// ShowHintAt displays a temporary on-screen message at a screen position.
func (g *Game) ShowHintAt(msg string, x, y int) {
	g.hint = msg
	g.hintTimer = 60
	g.hintX = x
	g.hintY = y
}

// AddItemToPlayer tries to add an item to the player's inventory.
// If the inventory is full, a hint is shown.
func (g *Game) AddItemToPlayer(it *items.Item) bool {
	if g.player == nil {
		return false
	}
	if g.player.Inventory.AddItem(it) {
		return true
	}
	g.ShowHint("Inventory full")
	return false
}

// spawnItemDrop places an item into the world at the given tile.
func (g *Game) spawnItemDrop(it *items.Item, x, y int) {
	if it == nil {
		return
	}
	drop := &entities.ItemDrop{TileX: x, TileY: y, Item: *it}
	g.ItemDrops = append(g.ItemDrops, drop)
	if g.currentLevel != nil {
		g.currentLevel.AddEntity(levels.PlacedEntity{
			X: x, Y: y, Type: "ItemDrop", SpriteID: it.ID,
		})
	}
}

// pickupItemsAt transfers any item drops at the given tile into the player's inventory.
func (g *Game) pickupItemsAt(x, y int) {
	if g.player == nil || g.player.Inventory == nil {
		return
	}
	for i := len(g.ItemDrops) - 1; i >= 0; i-- {
		d := g.ItemDrops[i]
		if d.TileX != x || d.TileY != y {
			continue
		}
		it := d.Item
		if g.AddItemToPlayer(&it) {
			g.ItemDrops = append(g.ItemDrops[:i], g.ItemDrops[i+1:]...)
			if g.currentLevel != nil {
				g.currentLevel.RemoveEntityAt(x, y, "ItemDrop", d.Item.ID)
			}
		}
	}
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
	tick := make([][]int, level.H)
	for y := range seen {
		seen[y] = make([]bool, level.W)
		tick[y] = make([]int, level.W)
	}
	g.SeenTiles = seen
	g.visibleTick = tick
}

// spawnEntitiesFromLevel creates monsters based on the placed entities in the
// currently loaded level.
func (g *Game) spawnEntitiesFromLevel() {
	g.Monsters = []*entities.Monster{}
	g.ItemDrops = []*entities.ItemDrop{}
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
		case "ItemDrop":
			if tmpl, ok := items.Registry[ent.SpriteID]; ok {
				it := &items.Item{ItemTemplate: tmpl, Count: 1}
				g.ItemDrops = append(g.ItemDrops, &entities.ItemDrop{TileX: ent.X, TileY: ent.Y, Item: *it})
			}
		}
	}
}

// stairPlaced is triggered by the editor when a stairwell sprite is placed.
func (g *Game) stairPlaced(x, y int, spriteID string) {
	if !g.editor.Active {
		return
	}
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
	if g.hintTimer > 0 {
		g.hintTimer--
	}
	switch g.State {
	case StateMainMenu:
		return g.updateMainMenu()
	case StateGameOver:
		return g.updateGameOver()
	case StateDeathScreen:
		return g.updateDeathScreen()
	case StateVictoryScreen:
		return g.updateVictoryScreen()
	case StatePlaying:
		if g.player != nil && g.player.IsDead {
			if g.RunState != nil && g.RunState.Active {
				g.endRunDeath()
			} else {
				g.State = StateGameOver
			}
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
		} else if g.GenerateMenu != nil && g.GenerateMenu.Menu.IsVisible() {
			g.GenerateMenu.Update()
		} else if g.ProcGenMenu != nil && g.ProcGenMenu.IsVisible() {
			g.ProcGenMenu.Update()
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

	// Recast whenever the player's interpolated position changes (sub-tile
	// continuous updates for smooth shadow movement).
	shouldRecast := len(g.cachedRays) == 0 ||
		g.player.MoveController.InterpX != g.lastPlayerX ||
		g.player.MoveController.InterpY != g.lastPlayerY

	if g.player != nil && shouldRecast {
		// Offset by +0.5 to centre the ray origin on the player's tile rather
		// than at its top-left iso corner.  In iso space this shifts the shadow
		// apex +16 px in Y (diamond centre), aligning it with the character sprite.
		originX := g.player.MoveController.InterpX + 0.5
		originY := g.player.MoveController.InterpY + 0.5

		g.cachedRays = fov.RayCasting(originX, originY, g.RaycastWalls, g.currentLevel)
		g.lastPlayerX = g.player.MoveController.InterpX
		g.lastPlayerY = g.player.MoveController.InterpY

		// Advance the recast counter only when the visible set actually changes.
		// gameTick is a "recast counter", not a wall-clock frame counter.
		// Keeping it still when the player is stationary means visibleTick[y][x]
		// stays within the fovDecayFrames window and tiles never go gray at rest.
		g.gameTick++

		// Stamp the current tick on each tile hit by a ray.
		// No full-array clear needed — tiles decay naturally once gameTick
		// advances fovDecayFrames beyond their last stamp.
		for _, ray := range g.cachedRays {
			for _, pt := range ray.Path {
				if g.isValidTile(pt.X, pt.Y) {
					g.visibleTick[pt.Y][pt.X] = g.gameTick
					g.SeenTiles[pt.Y][pt.X] = true
				}
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
		prevX, prevY := g.lastPlayerTileX, g.lastPlayerTileY
		g.player.Update(g.currentLevel, g.DeltaTime)
		if g.player.TileX != prevX || g.player.TileY != prevY {
			g.pickupItemsAt(g.player.TileX, g.player.TileY)
			g.lastPlayerTileX, g.lastPlayerTileY = g.player.TileX, g.player.TileY
		}
		g.updateCameraFollow()

		// Dev cheats.
		if g.GodMode {
			g.player.HP = g.player.MaxHP
		}
		if g.InfMana {
			g.player.Mana = g.player.MaxMana
		}

		if g.HUD != nil {
			g.HUD.HealthPercent = float64(g.player.HP) / float64(g.player.MaxHP)
			g.HUD.ManaPercent = float64(g.player.Mana) / float64(g.player.MaxMana)
			g.HUD.PlayerMana = g.player.Mana
			g.HUD.DashCharges = g.player.DashCharges
			g.HUD.DashEnabled = g.player.HasAbility("dash")
			g.HUD.GrappleEnabled = g.player.HasAbility("grapple")
			maxCD := 0.0
			for _, cd := range g.player.DashCooldowns {
				if cd > maxCD {
					maxCD = cd
				}
			}
			g.HUD.DashCooldown = maxCD
			g.HUD.ExpCurrent = g.player.EXP
			g.HUD.ExpNeeded = progression.EXPToLevel(g.player.Level)
			g.syncHUDSpellSlots()
		}
		if g.HeroPanel != nil {
			g.HeroPanel.Update()
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

	// Monster projectiles
	g.updateMonsterProjectiles()

	// NPCs
	if g.player != nil {
		for _, npc := range g.NPCs {
			npc.Update(g.player.TileX, g.player.TileY, g.DeltaTime)
		}
	}

	// Boss arena activation: seal doors when player enters the boss room.
	if g.CurrentBoss != nil && !g.CurrentBoss.IsActive && g.BossRoom != nil {
		if g.BossRoom.Contains(g.player.TileX, g.player.TileY) {
			g.activateBoss()
			g.sealBossRoom()
		}
	}

	// Boss health bar sync.
	if g.CurrentBoss != nil && g.BossBar != nil && g.CurrentBoss.IsActive {
		g.BossBar.CurrentHP = g.CurrentBoss.Monster.HP
		g.BossBar.Visible = true
		if g.CurrentBoss.Monster.IsDead {
			g.onBossDefeated()
		}
	}

	g.updateSpells()

	if g.DevMenu != nil {
		g.DevMenu.Update()
	}
	if g.DevTools != nil {
		g.DevTools.Update()
	}
	if g.ControlsMenu != nil && g.ControlsMenu.IsVisible() {
		g.ControlsMenu.Update()
	}

	// NPC interaction hints
	g.updateNPCHints()

	// Exit entity interaction (run loop)
	if g.ExitEntity != nil && g.RunState != nil && g.RunState.Active {
		g.ExitEntity.Update()
		if g.ExitEntity.IsPlayerNear(g.player.TileX, g.player.TileY) {
			isoX, isoY := g.cartesianToIso(float64(g.ExitEntity.TileX), float64(g.ExitEntity.TileY))
			hx := int((isoX-g.camX)*g.camScale + float64(g.w/2))
			hy := int((isoY+g.camY)*g.camScale + float64(g.h/2) - 32)
			g.ShowHintAt("[E] Descend", hx, hy)
			if g.isActionJustPressed(controls.ActionInteract) {
				g.advanceFloor()
			}
		}
	}

	// Hub portal interaction
	if g.IsInHub && g.hubPortalX >= 0 && g.hubPortalY >= 0 {
		dx := g.player.TileX - g.hubPortalX
		dy := g.player.TileY - g.hubPortalY
		if dx*dx+dy*dy <= 9 {
			isoX, isoY := g.cartesianToIso(float64(g.hubPortalX), float64(g.hubPortalY))
			hx := int((isoX-g.camX)*g.camScale + float64(g.w/2))
			hy := int((isoY+g.camY)*g.camScale + float64(g.h/2) - 32)
			g.ShowHintAt("[E] Enter the Dungeon", hx, hy)
			if g.isActionJustPressed(controls.ActionInteract) {
				g.StartRun()
			}
		}
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

	// Compact fixed rect for pause/settings menus
	pauseW, pauseH := 300, 400
	pauseX := (g.w - pauseW) / 2
	pauseY := (g.h - pauseH) / 2
	pauseRect := image.Rect(pauseX, pauseY, pauseX+pauseW, pauseY+pauseH)
	if g.PauseMenu != nil {
		if g.PauseMenu.MainMenu != nil {
			g.PauseMenu.MainMenu.SetRect(pauseRect)
		}
		if g.PauseMenu.SettingsMenu != nil {
			g.PauseMenu.SettingsMenu.SetRect(pauseRect)
		}
	}

	// Larger rect for content menus (load/save/generate)
	menuWidth := max(300, g.w/2)
	menuHeight := max(300, g.h/2)
	menuX := (g.w - menuWidth) / 2
	menuY := (g.h - menuHeight) / 2
	newRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)

	if g.LoadLevelMenu != nil {
		g.LoadLevelMenu.SetRect(newRect)
	}

	if g.LoadPlayerMenu != nil {
		g.LoadPlayerMenu.SetRect(newRect)
	}

	if g.SaveLevelMenu != nil {
		g.SaveLevelMenu.SetRect(newRect)
	}

	if g.GenerateMenu != nil {
		g.GenerateMenu.SetRect(newRect)
	}
	if g.ProcGenMenu != nil {
		g.ProcGenMenu.SetRect(newRect)
	}

	if g.HeroPanel != nil {
		panel := image.Rect(g.w/2-150, g.h/2-150, g.w/2+150, g.h/2+150)
		g.HeroPanel.SetRect(panel)
	}

	// Keep both ControlsMenu instances sized to the actual screen
	if g.ControlsMenu != nil {
		g.ControlsMenu.Resize(g.w, g.h)
	}
	if g.DevTools != nil {
		g.DevTools.Resize(g.w, g.h)
	}
	if g.PauseMenu != nil && g.PauseMenu.ControlsMenu != nil {
		g.PauseMenu.ControlsMenu.Resize(g.w, g.h)
	}
	if g.DialoguePanel != nil {
		g.DialoguePanel.Resize(g.w, g.h)
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

// openDoor attempts to open a door at the specified tile coordinates.
// Returns true if the door was opened, false if it couldn't be opened (locked or already open).
func (g *Game) openDoor(x, y int) bool {
	if g.currentLevel == nil {
		return false
	}
	tile := g.currentLevel.Tile(x, y)
	if tile == nil || !tile.HasTag(tiles.TagDoor) {
		return false
	}

	// Locked doors require keys (future: check player inventory for key)
	if tile.DoorState == 0 {
		id := tile.DoorSpriteID
		if id == "" {
			for _, s := range tile.Sprites {
				if isDoorSpriteID(s.ID) {
					id = s.ID
					break
				}
			}
		}
		lower := strings.ToLower(id)
		if strings.Contains(lower, "unlockeddoor") || strings.Contains(lower, "door_unlocked") {
			tile.DoorState = 1
			tile.IsWalkable = true
			g.setDoorSprite(tile, false)
		} else {
			tile.DoorState = 2
			tile.IsWalkable = false
			g.setDoorSprite(tile, true)
		}
	}
	if tile.DoorState == 3 {
		g.ShowHint("Door is locked")
		return false
	}

	// Already open
	if tile.DoorState == 1 {
		return false
	}

	// Open the door (only unlocked doors can be opened)
	if tile.DoorState == 2 {
		tile.DoorState = 1
		tile.IsWalkable = true
		g.setDoorSprite(tile, false)

		g.ShowHint("Door opened")
		return true
	}

	return false
}

// unlockDoor attempts to unlock a locked door and open it.
// Returns true if the door was unlocked/opened.
func (g *Game) unlockDoor(x, y int) bool {
	if g.currentLevel == nil {
		return false
	}
	tile := g.currentLevel.Tile(x, y)
	if tile == nil || !tile.HasTag(tiles.TagDoor) {
		return false
	}
	if tile.DoorState != 3 {
		return false
	}
	tile.DoorState = 2 // unlock (closed)
	tile.IsWalkable = false
	g.setDoorSprite(tile, true)
	return g.openDoor(x, y)
}

// closeDoor closes an open door and restores collision.
func (g *Game) closeDoor(x, y int) bool {
	if g.currentLevel == nil {
		return false
	}
	tile := g.currentLevel.Tile(x, y)
	if tile == nil || !tile.HasTag(tiles.TagDoor) {
		return false
	}
	if tile.DoorState != 1 {
		return false
	}
	tile.DoorState = 2
	tile.IsWalkable = false
	g.setDoorSprite(tile, true)
	g.ShowHint("Door closed")
	return true
}

func (g *Game) setDoorSprite(tile *tiles.Tile, closed bool) {
	if tile == nil {
		return
	}
	id := tile.DoorSpriteID
	if id == "" {
		for _, s := range tile.Sprites {
			if isDoorSpriteID(s.ID) {
				id = s.ID
				break
			}
		}
		if id == "" {
			return
		}
	}
	if closed {
		id = swapDoorSpriteID(id, true)
	} else {
		id = swapDoorSpriteID(id, false)
	}
	meta, ok := leveleditor.SpriteRegistry[id]
	if !ok {
		return
	}
	updated := false
	for i, s := range tile.Sprites {
		if s.ID == tile.DoorSpriteID {
			tile.Sprites[i] = tiles.SpriteRef{ID: id, Image: meta.Image}
			updated = true
			break
		}
	}
	if !updated {
		tile.Sprites = append(tile.Sprites, tiles.SpriteRef{ID: id, Image: meta.Image})
	}
	tile.DoorSpriteID = id
}

func isDoorSpriteID(id string) bool {
	lower := strings.ToLower(id)
	return strings.Contains(lower, "door_locked") || strings.Contains(lower, "door_unlocked") ||
		strings.Contains(lower, "lockeddoor") || strings.Contains(lower, "unlockeddoor")
}

// swapDoorSpriteID maps locked/unlocked variants for both flavored and base door IDs.
// closed=true selects the locked/closed sprite; closed=false selects unlocked/open sprite.
func swapDoorSpriteID(id string, closed bool) string {
	if strings.Contains(id, "_door_locked_") || strings.Contains(id, "_door_unlocked_") {
		if closed {
			return strings.Replace(id, "_door_unlocked_", "_door_locked_", 1)
		}
		return strings.Replace(id, "_door_locked_", "_door_unlocked_", 1)
	}
	if strings.Contains(id, "LockedDoorNW") || strings.Contains(id, "UnlockedDoorNW") {
		if closed {
			return strings.Replace(id, "UnlockedDoorNW", "LockedDoorNW", 1)
		}
		return strings.Replace(id, "LockedDoorNW", "UnlockedDoorNW", 1)
	}
	if strings.Contains(id, "LockedDoorNE") || strings.Contains(id, "UnlockedDoorNE") {
		if closed {
			return strings.Replace(id, "UnlockedDoorNE", "LockedDoorNE", 1)
		}
		return strings.Replace(id, "LockedDoorNE", "UnlockedDoorNE", 1)
	}
	return id
}
