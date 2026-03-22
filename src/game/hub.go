package game

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/spells"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// loadHub loads the hand-crafted hub level from levels/hub.json.
// It scans the loaded tiles for a Portal sprite to locate the dungeon entrance,
// and finds a walkable spawn point near it for the player.
func (g *Game) loadHub() {
	l, err := leveleditor.LoadLevelFromFile("levels/hub.json")
	if err != nil {
		fmt.Printf("hub: failed to load levels/hub.json, using fallback: %v\n", err)
		l = g.generateFallbackHub()
	}

	// Scan tiles for the portal position
	portalX, portalY := -1, -1
	for y := 0; y < l.H; y++ {
		for x := 0; x < l.W; x++ {
			tile := l.Tile(x, y)
			if tile == nil {
				continue
			}
			for _, s := range tile.Sprites {
				if strings.EqualFold(s.ID, "portal") {
					portalX, portalY = x, y
				}
			}
		}
	}

	// Find player spawn: nearest walkable tile to the portal
	spawnX, spawnY := g.findHubSpawn(l, portalX, portalY)

	// Set up the world
	newWorld := levels.NewLayeredLevel(l)
	g.currentWorld = newWorld
	g.currentLevel = l
	g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
	g.editor.OnLayerChange = g.editorLayerChanged
	g.editor.OnStairPlaced = g.stairPlaced
	g.editor.Active = false
	g.UpdateSeenTiles(*l)

	// Position player
	g.player.TileX = spawnX
	g.player.TileY = spawnY
	g.player.MoveController.InterpX = float64(spawnX)
	g.player.MoveController.InterpY = float64(spawnY)
	g.player.MoveController.Path = nil
	g.player.MoveController.Stop()
	g.player.CollisionBox.X = float64(spawnX)
	g.player.CollisionBox.Y = float64(spawnY)
	g.player.IsDead = false
	g.player.HP = g.player.MaxHP
	g.player.Mana = g.player.MaxMana

	// Clear run state
	g.Monsters = []*entities.Monster{}
	g.ItemDrops = []*entities.ItemDrop{}
	g.ActiveSpells = []spells.Spell{}
	g.ExitEntity = nil
	g.RunState = nil
	g.IsInHub = true
	g.hubPortalX = portalX
	g.hubPortalY = portalY
	g.FullBright = true
	snapIsoX, snapIsoY := g.cartesianToIso(float64(spawnX), float64(spawnY))
	g.camX = snapIsoX
	g.camY = -snapIsoY
	g.cachedRays = nil
	g.RaycastWalls = fov.LevelToWalls(g.currentLevel)
	g.State = StatePlaying
}

// findHubSpawn returns a walkable tile adjacent to the portal, or falls back
// to the first walkable tile found via BFS.
func (g *Game) findHubSpawn(l *levels.Level, portalX, portalY int) (int, int) {
	if portalX >= 0 && portalY >= 0 {
		// Check cardinal neighbours first
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			nx, ny := portalX+d[0], portalY+d[1]
			if l.IsWalkable(nx, ny) {
				return nx, ny
			}
		}
		// Then diagonal
		for _, d := range [][2]int{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			nx, ny := portalX+d[0], portalY+d[1]
			if l.IsWalkable(nx, ny) {
				return nx, ny
			}
		}
	}
	// Fallback: first walkable tile
	sx, sy := levels.FindSpawnPoint(l)
	return sx, sy
}

// generateFallbackHub builds a simple procedural hub when hub.json is missing.
func (g *Game) generateFallbackHub() *levels.Level {
	ss := g.spriteSheet
	const size = 24
	l := levels.CreateNewBlankLevel(size, size, constants.DefaultTileSize, ss)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if x == 0 || y == 0 || x == size-1 || y == size-1 {
				if tile := l.Tile(x, y); tile != nil {
					tile.ClearSprites()
					tile.AddSpriteByID("DungeonWall", ss.DungeonWall)
					tile.IsWalkable = false
				}
			}
		}
	}

	px, py := size/2, size/2
	if tile := l.Tile(px, py); tile != nil {
		tile.ClearSprites()
		tile.AddSpriteByID("Floor", ss.Floor)
		tile.AddSpriteByID("Portal", ss.Portal)
		tile.IsWalkable = true
	}

	for _, spot := range [][2]int{{px - 3, py - 3}, {px + 3, py - 3}, {px - 3, py + 3}, {px + 3, py + 3}} {
		if tile := l.Tile(spot[0], spot[1]); tile != nil {
			tile.ClearSprites()
			tile.AddSpriteByID("Floor", ss.Floor)
			tile.AddSpriteByID("Campfire", ss.Campfire)
			tile.IsWalkable = false
		}
	}

	return l
}

// returnToHub resets the player and loads the hub world.
func (g *Game) returnToHub() {
	g.resetPlayerForHub()
	g.loadHub()
}

// resetPlayerForHub restores the player to a fresh state for the hub.
func (g *Game) resetPlayerForHub() {
	if g.player == nil {
		return
	}
	g.player.IsDead = false
	g.player.HP = g.player.MaxHP
	g.player.Mana = g.player.MaxMana
}

// StartRun begins a new dungeon run from the hub.
func (g *Game) StartRun() {
	g.Meta.RunCount++
	SaveMeta(g.Meta)
	g.RunState = NewRunState(DefaultRunFloors)
	g.IsInHub = false
	g.FullBright = false
	g.startFloor(1)
}

// startFloor generates and activates a new dungeon floor.
func (g *Game) startFloor(floorNum int) {
	ctx := g.RunState.BuildFloorContext(floorNum)
	g.RunState.CurrentFloor = floorNum

	// Generate the level
	lvl := levels.Generate64x64(ctx.GenParams)
	newWorld := levels.NewLayeredLevel(lvl)
	g.currentWorld = newWorld
	g.currentLevel = lvl
	g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
	g.editor.OnLayerChange = g.editorLayerChanged
	g.editor.OnStairPlaced = g.stairPlaced
	g.editor.Active = false
	g.UpdateSeenTiles(*lvl)

	// Find spawn and exit using two-pass BFS (guarantees max separation)
	spawnX, spawnY, exitX, exitY := levels.FindSpawnAndExit(lvl)
	g.player.TileX = spawnX
	g.player.TileY = spawnY
	g.player.MoveController.InterpX = float64(spawnX)
	g.player.MoveController.InterpY = float64(spawnY)
	g.player.MoveController.Path = nil
	g.player.MoveController.Stop()
	g.player.CollisionBox.X = float64(spawnX)
	g.player.CollisionBox.Y = float64(spawnY)
	g.ExitEntity = entities.NewExitEntity(exitX, exitY, g.spriteSheet.Portal, "Portal")

	// Spawn entities from level data
	g.spawnEntitiesFromLevel()

	// Spawn additional scaled monsters for the floor
	g.spawnFloorMonsters(ctx)

	// Reset camera and FOV
	snapIsoX, snapIsoY := g.cartesianToIso(float64(spawnX), float64(spawnY))
	g.camX = snapIsoX
	g.camY = -snapIsoY
	g.cachedRays = nil
	g.RaycastWalls = fov.LevelToWalls(g.currentLevel)
}

// advanceFloor moves to the next floor or triggers victory.
func (g *Game) advanceFloor() {
	g.RunState.FloorsCleared++
	if g.RunState.IsLastFloor() {
		g.endRunVictory()
		return
	}
	g.RunState.CurrentFloor++
	g.startFloor(g.RunState.CurrentFloor)
}

// endRunDeath handles the end of a run due to player death.
func (g *Game) endRunDeath() {
	g.RunState.Active = false
	g.RunState.RemnantEarned = g.RunState.CalculateRemnants()
	g.Meta.Remnants += g.RunState.RemnantEarned
	g.Meta.TotalKills += g.RunState.KillCount
	if g.RunState.FloorsCleared > g.Meta.BestFloor {
		g.Meta.BestFloor = g.RunState.FloorsCleared
	}
	SaveMeta(g.Meta)
	g.State = StateDeathScreen
}

// endRunVictory handles the end of a run due to completing the final floor.
func (g *Game) endRunVictory() {
	g.RunState.Active = false
	g.RunState.FloorsCleared = g.RunState.TotalFloors
	remnants := g.RunState.CalculateRemnants()
	g.RunState.RemnantEarned = remnants * 2 // bonus for victory
	g.Meta.Remnants += g.RunState.RemnantEarned
	g.Meta.TotalKills += g.RunState.KillCount
	if g.RunState.TotalFloors > g.Meta.BestFloor {
		g.Meta.BestFloor = g.RunState.TotalFloors
	}
	SaveMeta(g.Meta)
	g.State = StateVictoryScreen
}

// spawnFloorMonsters places procedural monsters scaled to floor difficulty.
func (g *Game) spawnFloorMonsters(ctx FloorContext) {
	ss := g.spriteSheet

	type monsterTemplate struct {
		name   string
		sprite *ebiten.Image
	}
	templates := []monsterTemplate{
		{"Grey Knight", ss.GreyKnight},
		{"Sentinel", ss.Sentinel},
		{"Chimera", ss.Chimera},
		{"Lesser Demon", ss.LesserDemon},
	}

	// Base monster count scales with floor number
	baseCount := 3 + ctx.FloorNumber*2
	count := baseCount + rand.IntN(3)

	for i := 0; i < count; i++ {
		// Find a random walkable tile
		attempts := 0
		for attempts < 50 {
			x := rand.IntN(g.currentLevel.W)
			y := rand.IntN(g.currentLevel.H)
			if !g.currentLevel.IsWalkable(x, y) {
				attempts++
				continue
			}
			if x == g.player.TileX && y == g.player.TileY {
				attempts++
				continue
			}
			if g.ExitEntity != nil && x == g.ExitEntity.TileX && y == g.ExitEntity.TileY {
				attempts++
				continue
			}
			dx := x - g.player.TileX
			dy := y - g.player.TileY
			if dx*dx+dy*dy < 25 {
				attempts++
				continue
			}

			t := templates[rand.IntN(len(templates))]
			hpScale := 1.0 + ctx.Difficulty*0.5
			dmgScale := 1.0 + ctx.Difficulty*0.3
			baseHP := 8
			baseDmg := 2
			m := &entities.Monster{
				Name:             t.name,
				TileX:            x,
				TileY:            y,
				InterpX:          float64(x),
				InterpY:          float64(y),
				Sprite:           t.sprite,
				MovementDuration: 30,
				LeftFacing:       true,
				HP:               int(float64(baseHP) * hpScale),
				MaxHP:            int(float64(baseHP) * hpScale),
				Damage:           int(float64(baseDmg) * dmgScale),
				AttackRate:       45,
				Behavior:         entities.NewRoamingWanderBehavior(5),
				Level:            ctx.FloorNumber,
			}
			g.Monsters = append(g.Monsters, m)
			break
		}
	}
}
