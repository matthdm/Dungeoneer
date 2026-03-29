package game

import (
	"dungeoneer/dialogue"
	"dungeoneer/entities"
	"dungeoneer/items"
	"dungeoneer/levels"
	"dungeoneer/ui"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// updateNPCHints shows "[E] Talk" hints for nearby interactable NPCs.
func (g *Game) updateNPCHints() {
	if g.player == nil || g.player.IsDead {
		return
	}
	// Don't show hints while dialogue is open
	if g.DialoguePanel != nil && g.DialoguePanel.Active {
		return
	}
	for _, npc := range g.NPCs {
		if !npc.Interactable {
			continue
		}
		if !npc.IsPlayerInRangeAt(g.player.MoveController.InterpX, g.player.MoveController.InterpY) {
			continue
		}
		isoX, isoY := g.cartesianToIso(npc.InterpX, npc.InterpY)
		msg := "[E] Talk"
		// The iso anchor is the sprite's top-left. Offset to center:
		// +tileSize/2 horizontally centers on the tile diamond,
		// then subtract half the text pixel width to center the text itself.
		ts := float64(g.currentLevel.TileSize)
		sprCenterX := (isoX + ts/2 - g.camX) * g.camScale
		hx := int(sprCenterX+float64(g.w/2)) - len(msg)*3
		hy := int((isoY+g.camY)*g.camScale+float64(g.h/2)) - 16
		g.ShowHintAt(msg, hx, hy)
		break // only show one hint at a time
	}
}

// findNearbyChest returns the first unopened chest in range of the player, or nil.
func (g *Game) findNearbyChest() *entities.Chest {
	if g.player == nil {
		return nil
	}
	px, py := g.player.MoveController.InterpX, g.player.MoveController.InterpY
	for _, c := range g.Chests {
		if !c.Opened && c.IsPlayerInRange(px, py) {
			return c
		}
	}
	return nil
}

// updateChestHints shows "[E] Open" hints for nearby unopened chests.
func (g *Game) updateChestHints() {
	if g.player == nil || g.player.IsDead {
		return
	}
	px, py := g.player.MoveController.InterpX, g.player.MoveController.InterpY
	for _, c := range g.Chests {
		if c.Opened || !c.IsPlayerInRange(px, py) {
			continue
		}
		isoX, isoY := g.cartesianToIso(float64(c.TileX), float64(c.TileY))
		msg := "[E] Open"
		ts := float64(g.currentLevel.TileSize)
		sprCenterX := (isoX + ts/2 - g.camX) * g.camScale
		hx := int(sprCenterX+float64(g.w/2)) - len(msg)*3
		hy := int((isoY+g.camY)*g.camScale+float64(g.h/2)) - 16
		g.ShowHintAt(msg, hx, hy)
		break
	}
}

// findNearbyNPC returns the first interactable NPC in range of the player, or nil.
func (g *Game) findNearbyNPC() *entities.NPC {
	if g.player == nil {
		return nil
	}
	for _, npc := range g.NPCs {
		if npc.Interactable && npc.IsPlayerInRangeAt(g.player.MoveController.InterpX, g.player.MoveController.InterpY) {
			return npc
		}
	}
	return nil
}

// initDialoguePanel creates the dialogue panel and wires its callbacks.
func (g *Game) initDialoguePanel() {
	dp := ui.NewDialoguePanel(g.w, g.h)
	dp.OnClose = func() {}
	dp.EvalCondition = g.evalDialogueCondition
	dp.ExecAction = g.execDialogueAction
	g.DialoguePanel = dp
}

// openDialogue starts a dialogue with the given NPC.
func (g *Game) openDialogue(npc *entities.NPC) {
	if g.DialoguePanel == nil {
		g.initDialoguePanel()
	}

	treeID := npc.DialogueID
	if treeID == "" && npc.IsMajor {
		flags := make(map[string]int)
		if g.RunState != nil {
			flags = g.RunState.QuestFlags
		}
		treeID = dialogue.SelectTree(npc.ID, flags)
	}

	tree := dialogue.Registry[treeID]
	if tree == nil {
		return
	}

	// Mark NPC as met in persistent meta (enables hub spawning).
	if g.Meta != nil && npc.ID != "" {
		if g.Meta.NPCMeta[npc.ID] == nil {
			g.Meta.NPCMeta[npc.ID] = &NPCMetaState{}
		}
		if !g.Meta.NPCMeta[npc.ID].Met {
			g.Meta.NPCMeta[npc.ID].Met = true
			SaveMeta(g.Meta)
		}
	}

	portrait := g.resolvePortrait(npc.PortraitID)
	g.DialoguePanel.Open(tree, portrait)
}

// evalDialogueCondition checks whether a dialogue condition is met.
func (g *Game) evalDialogueCondition(c *dialogue.DialogueCondition) bool {
	if c == nil {
		return true
	}
	flags := make(map[string]int)
	if g.RunState != nil {
		flags = g.RunState.QuestFlags
	}

	switch c.Type {
	case "flag_equals":
		return flags[c.Flag] == c.Value
	case "flag_gte":
		return flags[c.Flag] >= c.Value
	case "flag_lte":
		return flags[c.Flag] <= c.Value
	case "not_flag":
		return flags[c.Flag] == 0
	case "has_item":
		// Checks inventory grid and all equipped slots.
		if g.player != nil {
			return g.player.HasItemAnywhere(c.ItemID)
		}
		return false
	case "has_ability":
		// c.Flag holds the ability ID (reuses Flag field to avoid schema change).
		if g.player != nil {
			return g.player.HasAbility(c.Flag)
		}
		return false
	case "has_gold":
		if g.player != nil {
			return g.player.Gold >= c.Value
		}
		return false
	}
	return true
}

// execDialogueAction performs a dialogue action's side effect.
func (g *Game) execDialogueAction(a dialogue.DialogueAction) {
	switch a.Type {
	case "set_flag":
		if g.RunState != nil {
			g.RunState.QuestFlags[a.Flag] = a.Value
		}
	case "add_flag":
		if g.RunState != nil {
			g.RunState.QuestFlags[a.Flag] += a.Value
		}
	case "give_exp":
		if g.player != nil && a.Amount > 0 {
			g.player.EXP += a.Amount
		}
	case "give_item":
		if g.player != nil && a.ItemID != "" {
			if _, ok := items.Registry[a.ItemID]; ok {
				it := items.NewItem(a.ItemID)
				g.player.AddToInventory(it)
				g.player.RefreshAbilities()
			}
		}
	case "take_item":
		if g.player != nil && a.ItemID != "" {
			g.player.RemoveItemByID(a.ItemID)
		}
	case "give_gold":
		if g.player != nil && a.Amount > 0 {
			g.player.Gold += a.Amount
		}
	case "take_gold":
		if g.player != nil && a.Amount > 0 {
			g.player.Gold -= a.Amount
			if g.player.Gold < 0 {
				g.player.Gold = 0
			}
		}
	}
}

// resolvePortrait maps a portrait ID to a sprite image from the sprite map.
func (g *Game) resolvePortrait(id string) *ebiten.Image {
	if id == "" {
		return nil
	}
	if img, ok := g.SpriteMap[id]; ok {
		return img
	}
	return nil
}

// spawnFloorNPCs places NPCs on a dungeon floor using tag-based placement.
func (g *Game) spawnFloorNPCs(ctx FloorContext) {
	pool := filterNPCsByBiome(minorNPCPool, ctx.Biome)
	if len(pool) == 0 {
		return
	}

	// Filter by floor range and roll spawn chance.
	var eligible []NPCTemplate
	for _, t := range pool {
		if t.SpawnMinFloor > 0 && ctx.FloorNumber < t.SpawnMinFloor {
			continue
		}
		if t.SpawnMaxFloor > 0 && ctx.FloorNumber > t.SpawnMaxFloor {
			continue
		}
		if t.SpawnChance > 0 && rand.Float64() > t.SpawnChance {
			continue
		}
		eligible = append(eligible, t)
	}
	if len(eligible) == 0 {
		return
	}

	// Sort by placement priority (quest first, hidden last).
	sort.Slice(eligible, func(i, j int) bool {
		return placementPriority(eligible[i].effectivePlacement()) <
			placementPriority(eligible[j].effectivePlacement())
	})

	// Cap at 2 NPCs per floor.
	maxNPCs := 2
	if len(eligible) > maxNPCs {
		eligible = eligible[:maxNPCs]
	}

	avoid := map[[2]int]bool{
		{g.player.TileX, g.player.TileY}: true,
	}
	if g.ExitEntity != nil {
		avoid[[2]int{g.ExitEntity.TileX, g.ExitEntity.TileY}] = true
	}

	for _, tmpl := range eligible {
		x, y := g.findNPCPlacement(tmpl.effectivePlacement(), avoid)
		if x < 0 {
			continue
		}
		avoid[[2]int{x, y}] = true
		npc := g.createNPCFromTemplate(tmpl, x, y)
		g.NPCs = append(g.NPCs, npc)
	}
}

// findNPCPlacement finds a tile for the given spawn strategy.
func (g *Game) findNPCPlacement(strategy SpawnStrategy, avoid map[[2]int]bool) (int, int) {
	lvl := g.currentLevel
	switch strategy {
	case SpawnQuest:
		// Prefer sanctuary rooms, fall back to largest common room.
		if rooms := levels.RoomsByTag(lvl.Rooms, levels.TagSanctuary); len(rooms) > 0 {
			return findWalkableInRoom(lvl, rooms[0], avoid)
		}
		if rooms := levels.RoomsByTag(lvl.Rooms, levels.TagCommon); len(rooms) > 0 {
			return findWalkableInRoom(lvl, rooms[0], avoid)
		}

	case SpawnAmbient:
		// Common or crossroads rooms.
		candidates := levels.RoomsByTag(lvl.Rooms, levels.TagCommon)
		candidates = append(candidates, levels.RoomsByTag(lvl.Rooms, levels.TagCrossroads)...)
		rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
		for _, r := range candidates {
			if r.HasTag(levels.TagCleared) || r.HasTag(levels.TagBossArena) {
				continue
			}
			x, y := findWalkableInRoom(lvl, r, avoid)
			if x >= 0 {
				return x, y
			}
		}

	case SpawnWandering:
		// Corridor tile not inside any room.
		return findCorridorTile(lvl, avoid)

	case SpawnHidden:
		// Dead-end rooms or dead-end corridor tiles.
		return findDeadEndTile(lvl, avoid)

	case SpawnEntrance:
		// Near spawn room.
		if rooms := levels.RoomsByTag(lvl.Rooms, levels.TagSpawn); len(rooms) > 0 {
			return findWalkableInRoom(lvl, rooms[0], avoid)
		}

	case SpawnExit:
		// Near exit room.
		if rooms := levels.RoomsByTag(lvl.Rooms, levels.TagExit); len(rooms) > 0 {
			return findWalkableInRoom(lvl, rooms[0], avoid)
		}

	case SpawnBossAdjacent:
		// Room nearest to boss arena.
		arenas := levels.RoomsByTag(lvl.Rooms, levels.TagBossArena)
		if len(arenas) == 0 {
			return -1, -1
		}
		nearest := nearestRoom(lvl.Rooms, arenas[0])
		if nearest != nil {
			return findWalkableInRoom(lvl, nearest, avoid)
		}
	}
	return -1, -1
}

// findWalkableInRoom finds a walkable tile in the room that isn't in the avoid set.
func findWalkableInRoom(lvl *levels.Level, r *levels.Room, avoid map[[2]int]bool) (int, int) {
	// Try center first.
	if lvl.IsWalkable(r.CenterX, r.CenterY) && !avoid[[2]int{r.CenterX, r.CenterY}] {
		return r.CenterX, r.CenterY
	}
	// BFS outward from center within the room.
	type pt struct{ x, y int }
	visited := map[pt]bool{{r.CenterX, r.CenterY}: true}
	queue := []pt{{r.CenterX, r.CenterY}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := cur.x+d[0], cur.y+d[1]
			p := pt{nx, ny}
			if visited[p] || !r.Contains(nx, ny) {
				continue
			}
			visited[p] = true
			if lvl.IsWalkable(nx, ny) && !avoid[[2]int{nx, ny}] {
				return nx, ny
			}
			queue = append(queue, p)
		}
	}
	return -1, -1
}

// nearestRoom returns the room closest to target (by center distance), excluding
// the target itself and rooms tagged boss_arena.
func nearestRoom(rooms []levels.Room, target *levels.Room) *levels.Room {
	var best *levels.Room
	bestDist := math.MaxFloat64
	for i := range rooms {
		r := &rooms[i]
		if r.Index == target.Index || r.HasTag(levels.TagBossArena) {
			continue
		}
		dx := float64(r.CenterX - target.CenterX)
		dy := float64(r.CenterY - target.CenterY)
		d := dx*dx + dy*dy
		if d < bestDist {
			bestDist = d
			best = r
		}
	}
	return best
}

// openChest marks the chest as opened and spawns its loot as item drops.
func (g *Game) openChest(c *entities.Chest) {
	if c == nil || c.Opened {
		return
	}
	c.Opened = true

	if g.FloorCtx == nil {
		return
	}
	table := items.BuildDefaultLootTable(string(g.FloorCtx.Biome))
	if g.FloorCtx.BiomeConfig != nil && g.FloorCtx.BiomeConfig.LootTable != nil {
		table.Entries = append(table.Entries, g.FloorCtx.BiomeConfig.LootTable.Entries...)
	}
	results := items.RollChestLoot(table, c.Variant, g.FloorCtx.FloorNumber)
	for _, r := range results {
		tmpl, ok := items.Registry[r.ItemID]
		if !ok {
			continue
		}
		it := &items.Item{ItemTemplate: tmpl, Count: r.Count}
		g.spawnItemDrop(it, c.TileX, c.TileY)
	}
}

// spawnFloorChests places chests in treasure rooms on the current floor.
// One chest per treasure room; variant scales with floor depth.
func (g *Game) spawnFloorChests(ctx FloorContext) {
	rooms := levels.RoomsByTag(g.currentLevel.Rooms, levels.TagTreasure)
	if len(rooms) == 0 {
		return
	}
	avoid := map[[2]int]bool{
		{g.player.TileX, g.player.TileY}: true,
	}
	if g.ExitEntity != nil {
		avoid[[2]int{g.ExitEntity.TileX, g.ExitEntity.TileY}] = true
	}
	for _, room := range rooms {
		x, y := findWalkableInRoom(g.currentLevel, room, avoid)
		if x < 0 {
			continue
		}
		avoid[[2]int{x, y}] = true
		variant := chestVariantForFloor(ctx.FloorNumber, ctx.TotalFloors)
		chest := &entities.Chest{
			TileX:   x,
			TileY:   y,
			Variant: variant,
			Sprite:  g.spriteSheet.GrandChest,
		}
		g.Chests = append(g.Chests, chest)
	}
}

// chestVariantForFloor returns a chest tier appropriate for the given floor.
func chestVariantForFloor(floor, total int) string {
	progress := float64(floor) / float64(max(1, total))
	roll := rand.Float64()
	switch {
	case progress >= 0.75:
		if roll < 0.20 {
			return entities.ChestLocked
		} else if roll < 0.60 {
			return entities.ChestGold
		}
		return entities.ChestIron
	case progress >= 0.40:
		if roll < 0.10 {
			return entities.ChestGold
		} else if roll < 0.50 {
			return entities.ChestIron
		}
		return entities.ChestWooden
	default:
		if roll < 0.30 {
			return entities.ChestIron
		}
		return entities.ChestWooden
	}
}

// spawnHubNPCs places major NPCs in the hub that the player has previously met.
func (g *Game) spawnHubNPCs() {
	if g.Meta == nil {
		return
	}
	// Hub NPC positions (predetermined slots)
	hubSlots := [][2]int{{8, 8}, {16, 8}, {8, 16}, {16, 16}}
	slotIdx := 0

	for npcID, meta := range g.Meta.NPCMeta {
		if !meta.Met || slotIdx >= len(hubSlots) {
			continue
		}
		// Look up the NPC template
		var tmpl *NPCTemplate
		for i := range minorNPCPool {
			if minorNPCPool[i].ID == npcID {
				tmpl = &minorNPCPool[i]
				break
			}
		}
		if tmpl == nil {
			continue
		}
		pos := hubSlots[slotIdx]
		x, y := pos[0], pos[1]
		if !g.currentLevel.IsWalkable(x, y) {
			continue
		}
		npc := g.createNPCFromTemplate(*tmpl, x, y)
		g.NPCs = append(g.NPCs, npc)
		slotIdx++
	}
}
