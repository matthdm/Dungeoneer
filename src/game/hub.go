package game

import (
	"dungeoneer/audio"
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/fov"
	"dungeoneer/inventory"
	"dungeoneer/items"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/spells"
	"dungeoneer/ui"
	"fmt"
	"math/rand/v2"
	"os"
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
	g.MonsterProjectiles = nil
	g.ExitEntity = nil
	g.RunState = nil
	g.FloorCtx = nil
	g.CurrentBoss = nil
	g.BossBar = nil
	g.BossRoom = nil
	g.NPCs = []*entities.NPC{}
	g.Chests = []*entities.Chest{}
	g.IsInHub = true
	g.hubPortalX = portalX
	g.hubPortalY = portalY
	g.FullBright = true
	snapIsoX, snapIsoY := g.cartesianToIso(float64(spawnX), float64(spawnY))
	g.camX = snapIsoX
	g.camY = -snapIsoY
	g.cachedRays = nil
	g.RaycastWalls = fov.LevelToWalls(g.currentLevel)
	fov.InvalidateCache()
	g.spawnHubNPCs()

	// Milestone-gated hub features.
	if g.Meta != nil {
		if g.Meta.HubState[MilestoneShop] {
			shopTmpl := NPCTemplate{
				ID: "shop_npc", Name: "The Merchant", Title: "Items for Sale",
				IsMajor: false, DialogueID: "", SpriteID: "Sorcerer", PortraitID: "",
			}
			sx, sy := 6, 8
			if !g.currentLevel.IsWalkable(sx, sy) {
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := sx+d[0], sy+d[1]
					if g.currentLevel.IsWalkable(nx, ny) {
						sx, sy = nx, ny
						break
					}
				}
			}
			if g.currentLevel.IsWalkable(sx, sy) {
				npc := g.createNPCFromTemplate(shopTmpl, sx, sy)
				npc.HintText = "[E] Shop"
				npc.OnInteract = func() { g.openShop() }
				g.NPCs = append(g.NPCs, npc)
			}
		}
		if g.Meta.HubState[MilestoneUpgrades] {
			upgTmpl := NPCTemplate{
				ID: "upgrade_npc", Name: "The Artificer", Title: "Upgrades",
				IsMajor: false, DialogueID: "", SpriteID: "Oracle", PortraitID: "",
			}
			ux, uy := 18, 8
			if !g.currentLevel.IsWalkable(ux, uy) {
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := ux+d[0], uy+d[1]
					if g.currentLevel.IsWalkable(nx, ny) {
						ux, uy = nx, ny
						break
					}
				}
			}
			if g.currentLevel.IsWalkable(ux, uy) {
				npc := g.createNPCFromTemplate(upgTmpl, ux, uy)
				npc.HintText = "[E] Upgrade"
				npc.OnInteract = func() { g.openUpgradeStation() }
				g.NPCs = append(g.NPCs, npc)
			}
		}
		if g.Meta.HubState[MilestoneEchoShrine] {
			loreTmpl := NPCTemplate{
				ID: "echo_shrine", Name: "Echo Shrine", Title: "Echoes of the Past",
				IsMajor: false, DialogueID: "", SpriteID: "TorturedSoul", PortraitID: "",
			}
			sx, sy := 8, 12
			if !g.currentLevel.IsWalkable(sx, sy) {
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := sx+d[0], sy+d[1]
					if g.currentLevel.IsWalkable(nx, ny) {
						sx, sy = nx, ny
						break
					}
				}
			}
			if g.currentLevel.IsWalkable(sx, sy) {
				npc := g.createNPCFromTemplate(loreTmpl, sx, sy)
				npc.HintText = "[E] Commune"
				npc.OnInteract = func() { g.openEchoShrine() }
				g.NPCs = append(g.NPCs, npc)
			}
		}
		if g.Meta.HubState[MilestoneLoreLibrary] {
			// Lore pedestal — opens the lore library UI.
			// entities.NPC.OnInteract and g.openLoreLibrary() added by Agent E.
			loreTmpl := NPCTemplate{
				ID:         "lore_pedestal",
				Name:       "Lore Library",
				Title:      "Ancient Records",
				IsMajor:    false,
				DialogueID: "",
				SpriteID:   "Chest",
				PortraitID: "",
			}
			px, py := 10, 12
			if !g.currentLevel.IsWalkable(px, py) {
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := px+d[0], py+d[1]
					if g.currentLevel.IsWalkable(nx, ny) {
						px, py = nx, ny
						break
					}
				}
			}
			if g.currentLevel.IsWalkable(px, py) {
				npc := g.createNPCFromTemplate(loreTmpl, px, py)
				npc.HintText = "[E] Read"
				npc.OnInteract = func() { g.openLoreLibrary() }
				g.NPCs = append(g.NPCs, npc)
			}
		}
	}

	// Fade in from black when arriving at the hub.
	if g.Transition != nil {
		g.Transition.StartFadeIn()
	}
	if g.Audio != nil {
		g.Audio.StopMusic()
		g.Audio.PlayAmbient(audio.AmbientHub)
	}
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
// Per design: full death reset — player loses all items, equipment, gold,
// stats, and levels. Only meta-progression (Remnants) survives.
func (g *Game) resetPlayerForHub() {
	if g.player == nil {
		return
	}
	g.player.IsDead = false

	// Reset level and progression.
	g.player.Level = 1
	g.player.EXP = 0
	g.player.UnspentPoints = 0
	g.player.Gold = 0

	// Reset stats to defaults.
	g.player.Stats = entities.BaseStats{
		Strength: 1, Dexterity: 1, Vitality: 1, Intelligence: 1, Luck: 1,
	}
	g.player.TempModifiers = entities.StatModifiers{}

	// Clear inventory and equipment.
	g.player.Inventory = inventory.New(inventory.Width, inventory.Height)
	for slot := range g.player.Equipment {
		g.player.Equipment[slot] = nil
	}

	// Clear status effects and abilities.
	g.player.Effects = entities.EffectHolder{}
	g.player.ClearAbilities()

	// Recalculate derived stats and re-equip class starters.
	g.player.RecalculateStats()
	g.player.EquipStarter()

	// Apply permanent meta-progression upgrades so hub stats reflect earned bonuses.
	if g.Meta != nil {
		for _, def := range UpgradeRegistry {
			level := g.Meta.Upgrades[def.ID]
			if level > 0 && def.ID != "innate_dash" {
				def.Apply(g.player, level)
			}
		}
	}

	g.player.HP = g.player.MaxHP
	g.player.Mana = g.player.MaxMana
}

// StartRun begins a new dungeon run from the hub.
func (g *Game) StartRun() {
	g.Meta.RunCount++

	// Apply meta-progression upgrades to the player before the run begins.
	if g.player != nil {
		for _, def := range UpgradeRegistry {
			level := g.Meta.Upgrades[def.ID]
			if level > 0 && def.ID != "innate_dash" {
				def.Apply(g.player, level)
			}
		}
		// Apply innate_dash separately (grants ability directly).
		if g.Meta.Upgrades["innate_dash"] > 0 {
			if g.player.Abilities == nil {
				g.player.Abilities = make(map[string]bool)
			}
			g.player.Abilities["dash"] = true
		}
		// Give items purchased in the hub shop this run.
		for _, itemID := range g.Meta.ShopPurchases {
			if _, ok := items.Registry[itemID]; ok {
				it := items.NewItem(itemID)
				g.player.AddToInventory(it)
			}
		}
		// Reset shop purchases so next run starts fresh.
		g.Meta.ShopPurchases = nil
	}

	SaveMeta(g.Meta)
	if g.EchoRecorder != nil && g.Meta != nil {
		g.EchoRecorder.Start(g.Meta.CompletedRuns)
	}
	// Compute dungeon mood from player behavior profile.
	if g.Meta != nil {
		g.Meta.CurrentProfile = BuildProfile(g.Meta.RecentBehavior)
		g.Meta.CurrentMood = int(InferMood(g.Meta.CurrentProfile))
	}
	// Start behavior tracker (enemy count unknown at run start; RecordEnemySpawned called per spawn).
	if g.BehaviorTracker != nil {
		g.BehaviorTracker.Start(0)
	}
	g.RunState = NewRunState(DefaultRunFloors)
	g.seedNPCPhaseFlags()
	g.IsInHub = false
	g.FullBright = false
	g.startFloor(1)
}

// seedNPCPhaseFlags initialises per-run QuestFlags from MetaSave at run start.
//
// Phase intentionally resets to 0 every run — the full questline must be
// completed within a single run. What carries over is relationship state:
//   - {id}_met    — player has spoken to this NPC before (greeting varies)
//   - {id}_ng_plus — player has defeated this NPC as boss (unlocks NG+ dialogue)
//
// HighestPhase is tracked in MetaSave for analytics/display but is NOT seeded
// into QuestFlags here, so it cannot skip the questline on a new run.
func (g *Game) seedNPCPhaseFlags() {
	if g.Meta == nil || g.RunState == nil {
		return
	}
	for npcID, state := range g.Meta.NPCMeta {
		if state.Met {
			g.RunState.QuestFlags[npcID+"_met"] = 1
		}
		if state.DefeatCount > 0 {
			g.RunState.QuestFlags[npcID+"_ng_plus"] = 1
			g.RunState.QuestFlags[npcID+"_defeat_count"] = state.DefeatCount
		}
		if state.Betrayed {
			g.RunState.QuestFlags[npcID+"_betrayed"] = 1
		}
	}
}

// startFloor generates and activates a new dungeon floor.
// buildRunSave snapshots the current run state for mid-run persistence.
func (g *Game) buildRunSave() *RunSave {
	if g.RunState == nil || g.FloorCtx == nil || g.player == nil {
		return nil
	}
	var seed int64
	seed = g.FloorCtx.GenParams.Seed

	monsterSnaps := make([]MonsterSnap, 0, len(g.Monsters))
	for _, m := range g.Monsters {
		if m == nil || m.HP <= 0 {
			continue // skip dead monsters
		}
		monsterSnaps = append(monsterSnaps, MonsterSnap{
			TileX: m.TileX,
			TileY: m.TileY,
			HP:    m.HP,
			MaxHP: m.MaxHP,
			Level: m.Level,
			Role:  m.Role,
		})
	}

	return &RunSave{
		RunState:  *g.RunState,
		FloorSeed: seed,
		Player:    g.player.ToSnapshot(),
		Monsters:  monsterSnaps,
	}
}

// saveRun persists the current mid-run state. Non-fatal on error.
func (g *Game) saveRun() {
	rs := g.buildRunSave()
	if rs == nil {
		return
	}
	_ = SaveRunSave(rs)
}

func (g *Game) startFloor(floorNum int) {
	ctx := g.RunState.BuildFloorContext(floorNum)
	g.startFloorWithContext(ctx)
}

// startFloorWithContext generates and activates a dungeon floor from a pre-built FloorContext.
// It is the core of startFloor; callers that need a specific seed (e.g. resume from save)
// override ctx.GenParams.Seed before calling this.
func (g *Game) startFloorWithContext(ctx FloorContext) {
	g.RunState.CurrentFloor = ctx.FloorNumber
	g.FloorCtx = &ctx
	g.MonsterProjectiles = nil

	// Apply dungeon mood modifiers to generation params.
	if g.Meta != nil {
		mood := DungeonMood(g.Meta.CurrentMood)
		delta := MoodGenModifiers(mood)
		ctx.GenParams.RoomCountMod = delta.RoomCountDelta
		ctx.GenParams.CorridorWidthMod = delta.CorridorWidthMod
		// Store bias on game state for encounter spawner to read.
		g.CurrentMoodDelta = delta
	}

	// Floor-1 dungeon whisper toast.
	if ctx.FloorNumber == 1 && g.Meta != nil {
		mood := DungeonMood(g.Meta.CurrentMood)
		whisper := PickWhisper(mood, g.Meta.CompletedRuns)
		if whisper != "" {
			g.pendingToasts = append(g.pendingToasts, whisper)
		}
	}

	// Biome entry line — fires once per biome per run.
	if g.RunState != nil {
		biomeKey := string(ctx.Biome)
		if !g.RunState.BiomesIntroduced[biomeKey] {
			g.RunState.BiomesIntroduced[biomeKey] = true
			if line := GetBiomeFlavorLine(biomeKey, ctx.FloorNumber); line != "" {
				g.pendingToasts = append(g.pendingToasts, line)
			}
		}
	}
	// Assign event room for this floor (implemented in event_room.go by another agent).
	g.assignEventRoom()

	// Start biome ambient audio.
	if g.Audio != nil {
		g.Audio.PlayAmbient(audio.BiomeAmbient(string(ctx.Biome)))
	}

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
	// Fallback: if BFS couldn't separate spawn and exit (degenerate level),
	// find any walkable tile that isn't the spawn point.
	if exitX == spawnX && exitY == spawnY || !lvl.IsWalkable(exitX, exitY) {
		for y := 0; y < lvl.H && exitX == spawnX && exitY == spawnY; y++ {
			for x := 0; x < lvl.W; x++ {
				if lvl.IsWalkable(x, y) && (x != spawnX || y != spawnY) {
					exitX, exitY = x, y
					break
				}
			}
		}
	}
	g.ExitEntity = entities.NewExitEntity(exitX, exitY, g.spriteSheet.Portal, "Portal")

	// Spawn entities from level data
	g.spawnEntitiesFromLevel()

	// Boss floor: spawn boss in the largest room instead of a normal exit.
	g.CurrentBoss = nil
	g.BossBar = nil
	g.BossRoom = nil
	isBossFloor := g.RunState.IsLastFloor() && len(lvl.Rooms) > 0
	if isBossFloor {
		g.setupBossFloor(lvl)
		g.bossFloorAnnouncement = 240 // ~4 seconds at 60 TPS
	}

	// Tag rooms for semantic placement (must run after boss setup).
	levels.TagRooms(lvl, spawnX, spawnY, exitX, exitY, isBossFloor)

	// Spawn monsters using encounter template system (falls back to legacy if needed)
	g.spawnEncounterMonsters(ctx)

	// Spawn echo entities from prior runs.
	g.spawnFloorEchoes(ctx.FloorNumber)

	// Advance NPC phases whose auto-advance conditions are now met.
	g.checkMajorNPCPhaseAdvancement()

	// Spawn NPCs: major NPCs first (get priority placement), then minor NPCs.
	g.NPCs = []*entities.NPC{}
	g.spawnMajorNPCs(ctx)
	g.spawnFloorNPCs(ctx)

	// Spawn chests in treasure rooms
	g.Chests = []*entities.Chest{}
	g.spawnFloorChests(ctx)

	// Reset camera and FOV
	snapIsoX, snapIsoY := g.cartesianToIso(float64(spawnX), float64(spawnY))
	g.camX = snapIsoX
	g.camY = -snapIsoY
	g.cachedRays = nil
	g.RaycastWalls = fov.LevelToWalls(g.currentLevel)
	fov.InvalidateCache()

	// Persist run state so the player can resume from this floor.
	// (Ebiten v2.8 has no OnClose callback; floor transitions cover the important save point.)
	if g.RunState != nil && g.RunState.Active {
		g.saveRun()
	}
}

// advanceFloor moves to the next floor or triggers victory.
// Uses a fade-out→floor load→fade-in transition when Transition is available.
func (g *Game) advanceFloor() {
	g.RunState.FloorsCleared++
	if g.RunState.IsLastFloor() {
		g.endRunVictory()
		return
	}
	nextFloor := g.RunState.CurrentFloor + 1
	if g.Transition != nil {
		g.Transition.Start(func() {
			g.startFloor(nextFloor)
		})
	} else {
		g.startFloor(nextFloor)
	}
}

// endRunDeath handles the end of a run due to player death.
func (g *Game) endRunDeath() {
	g.RunState.Active = false
	g.RunState.RemnantEarned = g.RunState.CalculateRemnants()
	g.Meta.Remnants += g.RunState.RemnantEarned
	g.Meta.TotalKills += g.RunState.KillCount
	g.Meta.TotalDeaths++
	g.Meta.TotalRemnants += g.RunState.RemnantEarned
	if g.RunState.FloorsCleared > g.Meta.BestFloor {
		g.Meta.BestFloor = g.RunState.FloorsCleared
	}
	if g.EchoRecorder != nil && g.player != nil {
		g.EchoRecorder.Finalize("death", g.RunState.CurrentFloor,
			g.player.MoveController.InterpX, g.player.MoveController.InterpY,
			nil, g.Meta)
	}
	if g.BehaviorTracker != nil && g.Meta != nil {
		rec := g.BehaviorTracker.Finalize(true)
		g.Meta.RecentBehavior = append(g.Meta.RecentBehavior, rec)
		if len(g.Meta.RecentBehavior) > 5 {
			g.Meta.RecentBehavior = g.Meta.RecentBehavior[1:]
		}
	}
	newly := CheckMilestones(g.Meta)
	g.queueMilestoneToasts(newly)
	SaveMeta(g.Meta)
	ClearRunSave()
	g.State = StateDeathScreen
}

// endRunVictory handles the end of a run due to completing the final floor.
func (g *Game) endRunVictory() {
	g.RunState.Active = false
	g.RunState.FloorsCleared = g.RunState.TotalFloors
	remnants := g.RunState.CalculateRemnants()
	g.RunState.RemnantEarned = remnants * 2
	g.Meta.Remnants += g.RunState.RemnantEarned
	g.Meta.TotalKills += g.RunState.KillCount
	g.Meta.CompletedRuns++
	g.Meta.TotalRemnants += g.RunState.RemnantEarned
	if g.RunState.TotalFloors > g.Meta.BestFloor {
		g.Meta.BestFloor = g.RunState.TotalFloors
	}
	if g.EchoRecorder != nil && g.player != nil {
		g.EchoRecorder.Finalize("victory", g.RunState.TotalFloors,
			g.player.MoveController.InterpX, g.player.MoveController.InterpY,
			nil, g.Meta)
	}
	if g.BehaviorTracker != nil && g.Meta != nil {
		rec := g.BehaviorTracker.Finalize(false)
		g.Meta.RecentBehavior = append(g.Meta.RecentBehavior, rec)
		if len(g.Meta.RecentBehavior) > 5 {
			g.Meta.RecentBehavior = g.Meta.RecentBehavior[1:]
		}
	}
	newly := CheckMilestones(g.Meta)
	g.queueMilestoneToasts(newly)
	SaveMeta(g.Meta)
	ClearRunSave()
	g.State = StateVictoryScreen
}

// queueMilestoneToasts enqueues toast messages for each newly unlocked milestone.
func (g *Game) queueMilestoneToasts(milestoneIDs []string) {
	for _, id := range milestoneIDs {
		if msg, ok := MilestoneMessages[id]; ok {
			g.pendingToasts = append(g.pendingToasts, msg)
		}
	}
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
				HitRadius:        entities.DefaultMonsterHitRadius,
				AttackRate:       45,
				Behavior:         entities.NewRoamingWanderBehavior(5),
				Level:            ctx.FloorNumber,
			}
			g.Monsters = append(g.Monsters, m)
			break
		}
	}
}

// restoreMonsters recreates the monster list from a saved snapshot.
// Call after startFloorWithContext to replace the randomly-spawned monsters.
func (g *Game) restoreMonsters(snaps []MonsterSnap) {
	ss := g.spriteSheet
	g.Monsters = make([]*entities.Monster, 0, len(snaps))
	for _, snap := range snaps {
		sprite := ss.GreyKnight // default sprite
		name := "Grey Knight"
		switch snap.Role {
		case "elite":
			sprite = ss.Sentinel
			name = "Sentinel"
		case "caster":
			sprite = ss.Chimera
			name = "Chimera"
		case "ambush", "swarm":
			sprite = ss.LesserDemon
			name = "Lesser Demon"
		}
		m := &entities.Monster{
			Name:             name,
			TileX:            snap.TileX,
			TileY:            snap.TileY,
			InterpX:          float64(snap.TileX),
			InterpY:          float64(snap.TileY),
			Sprite:           sprite,
			MovementDuration: 30,
			LeftFacing:       true,
			HP:               snap.HP,
			MaxHP:            snap.MaxHP,
			Damage:           2 + snap.Level,
			HitRadius:        entities.DefaultMonsterHitRadius,
			AttackRate:       45,
			Behavior:         entities.NewRoamingWanderBehavior(5),
			Level:            snap.Level,
			Role:             snap.Role,
		}
		g.Monsters = append(g.Monsters, m)
	}
}

// resumeFromSave restores a mid-run session from a saved RunSave.
func (g *Game) resumeFromSave(rs *RunSave) {
	// Restore run state (copy to avoid aliasing).
	runState := rs.RunState
	g.RunState = &runState

	g.IsInHub = false
	g.FullBright = false
	g.seedNPCPhaseFlags()

	// Build the floor context then override the seed so we get the exact same
	// layout the player was on when they saved.
	floorNum := rs.RunState.CurrentFloor
	ctx := g.RunState.BuildFloorContext(floorNum)
	ctx.GenParams.Seed = rs.FloorSeed

	// Generate floor with saved seed, spawn entities (NPCs, chests).
	// Monsters are restored separately from snapshot below.
	g.startFloorWithContext(ctx)

	// Restore player vitals, position, inventory, and equipment.
	g.player.ApplySnapshot(rs.Player)
	g.player.RefreshAbilities()
	g.player.RecalculateStats()

	// Replace randomly-spawned monsters with saved snapshot.
	g.restoreMonsters(rs.Monsters)

	g.State = StatePlaying
}

// spawnFloorEchoes reads stored EchoRecords and spawns up to 3 echo entities
// near historical death/visit locations matching the current floor.
func (g *Game) spawnFloorEchoes(floor int) {
	if g.Meta == nil || len(g.Meta.EchoFiles) == 0 {
		return
	}

	type candidate struct {
		rec  *EchoRecord
		snap PositionSnapshot
	}
	var candidates []candidate
	for _, path := range g.Meta.EchoFiles {
		rec, err := LoadEchoRecord(path)
		if err != nil {
			continue
		}
		snaps := rec.SnapshotsForFloor(floor)
		if len(snaps) > 0 {
			// Use the last snapshot on this floor (nearest to death or deepest progress).
			candidates = append(candidates, candidate{rec, snaps[len(snaps)-1]})
		}
	}
	if len(candidates) == 0 {
		return
	}

	// Simple shuffle.
	for i := range candidates {
		j := i + rand.IntN(len(candidates)-i)
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}

	for _, c := range candidates {
		x, y := int(c.snap.X), int(c.snap.Y)
		// Find nearest walkable tile if needed.
		if g.currentLevel != nil && !g.currentLevel.IsWalkable(x, y) {
			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if g.currentLevel.IsWalkable(nx, ny) {
					x, y = nx, ny
					break
				}
			}
		}
		if g.currentLevel == nil || !g.currentLevel.IsWalkable(x, y) {
			continue
		}
		switch entities.EchoWeightedType() {
		case 0: // Wicked
			m := entities.NewWickedEcho(g.spriteSheet, x, y, c.rec.DeathFloor)
			g.Monsters = append(g.Monsters, m)
		case 1: // Hero
			m := entities.NewHeroEcho(g.spriteSheet, x, y, c.rec.DeathFloor, &g.Monsters)
			g.Monsters = append(g.Monsters, m)
		case 2: // Memory Fragment
			npc := entities.NewMemoryFragmentNPC(x, y, c.rec.DeathCause, g.spriteSheet)
			g.NPCs = append(g.NPCs, npc)
		}
	}
}

// openShop builds the shop entry list from ShopCatalog and opens the shop UI.
func (g *Game) openShop() {
	if g.Shop == nil || g.Meta == nil {
		return
	}
	var entries []ui.ShopEntry
	for _, si := range ShopCatalog {
		// Count how many times this item was already purchased this run.
		count := 0
		for _, p := range g.Meta.ShopPurchases {
			if p == si.ItemID {
				count++
			}
		}
		entries = append(entries, ui.ShopEntry{
			ItemID:    si.ItemID,
			Name:      si.Name,
			Cost:      si.Cost,
			Purchased: count,
			StockMax:  si.StockMax,
			CanAfford: g.Meta.Remnants >= si.Cost,
		})
	}
	g.Shop.OnPurchase = func(itemID string, cost int) bool {
		if g.Meta.Remnants < cost {
			return false
		}
		g.Meta.Remnants -= cost
		g.Meta.ShopPurchases = append(g.Meta.ShopPurchases, itemID)
		SaveMeta(g.Meta)
		return true
	}
	g.Shop.Open(entries, g.Meta.Remnants)
}

// openUpgradeStation builds the upgrade entry list and opens the upgrade station UI.
func (g *Game) openUpgradeStation() {
	if g.UpgradeStation == nil || g.Meta == nil {
		return
	}
	var entries []ui.UpgradeEntry
	for _, def := range UpgradeRegistry {
		level := g.Meta.Upgrades[def.ID]
		cost := UpgradeCost(def, level)
		entries = append(entries, ui.UpgradeEntry{
			ID:           def.ID,
			Name:         def.Name,
			Description:  def.Description,
			CurrentLevel: level,
			MaxLevel:     def.MaxLevel,
			NextCost:     cost,
			CanAfford:    cost > 0 && g.Meta.Remnants >= cost,
		})
	}
	g.UpgradeStation.OnUpgrade = func(upgradeID string, cost int) bool {
		if g.Meta.Remnants < cost {
			return false
		}
		g.Meta.Remnants -= cost
		g.Meta.Upgrades[upgradeID]++
		SaveMeta(g.Meta)
		return true
	}
	g.UpgradeStation.Open(entries, g.Meta.Remnants)
}

// openEchoShrine assembles EchoEntry list from stored records and opens the UI.
func (g *Game) openEchoShrine() {
	if g.EchoShrine == nil || g.Meta == nil {
		return
	}
	var entries []ui.EchoEntry
	for _, path := range g.Meta.EchoFiles {
		rec, err := LoadEchoRecord(path)
		if err != nil {
			continue
		}
		entries = append(entries, ui.EchoEntry{
			Path:       path,
			RunIndex:   rec.RunIndex,
			DeathCause: rec.DeathCause,
			DeathFloor: rec.DeathFloor,
		})
	}
	g.EchoShrine.OnBanish = func(path string) {
		_ = os.Remove(path)
		for i, f := range g.Meta.EchoFiles {
			if f == path {
				g.Meta.EchoFiles = append(g.Meta.EchoFiles[:i], g.Meta.EchoFiles[i+1:]...)
				break
			}
		}
		SaveMeta(g.Meta)
		g.openEchoShrine() // refresh
	}
	g.EchoShrine.Open(entries)
}
