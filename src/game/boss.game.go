package game

import (
	"dungeoneer/entities"
	"dungeoneer/hud"
	"dungeoneer/levels"
	"dungeoneer/tiles"
)

// spawnVarnBoss creates and places Varn as the boss on the final floor.
func (g *Game) spawnVarnBoss(x, y int) {
	isNGPlus := false
	if g.Meta != nil {
		if state := g.Meta.NPCMeta["varn"]; state != nil {
			isNGPlus = state.DefeatCount > 0
		}
	}
	boss := entities.NewVarnChainkeeper(g.spriteSheet, x, y, isNGPlus)
	g.CurrentBoss = boss
	g.Monsters = append(g.Monsters, boss.Monster)
	g.BossBar = &hud.BossHealthBar{
		Name:         boss.Monster.Name,
		Title:        boss.Title,
		MaxHP:        boss.Monster.MaxHP,
		CurrentHP:    boss.Monster.HP,
		PhaseMarkers: boss.PhaseHP,
		Visible:      false,
	}
}

// setupBossFloor identifies the largest room as the arena, spawns the boss,
// and removes the normal exit portal (it will be spawned on boss death).
func (g *Game) setupBossFloor(lvl *levels.Level) {
	// Find the largest room for the arena.
	var best *levels.Room
	bestArea := 0
	for i := range lvl.Rooms {
		r := &lvl.Rooms[i]
		area := r.W * r.H
		if area > bestArea {
			bestArea = area
			best = r
		}
	}
	if best == nil {
		return
	}

	// Find a walkable tile near the room center for boss placement.
	bx, by := findWalkableNear(lvl, best.CenterX, best.CenterY, best)
	if bx < 0 {
		return // room has no walkable tiles — skip boss
	}

	g.BossRoom = best

	// Use Varn as boss if his questline has reached the confrontation phase.
	if g.RunState != nil && g.RunState.QuestFlags["varn_phase"] >= 3 {
		g.spawnVarnBoss(bx, by)
	} else {
		g.spawnBoss(bx, by)
	}

	// Remove the normal exit — player must defeat the boss to proceed.
	g.ExitEntity = nil
}

// findWalkableNear returns a walkable tile near (cx, cy) within the room,
// spiraling outward. Returns (-1,-1) if none found.
func findWalkableNear(lvl *levels.Level, cx, cy int, room *levels.Room) (int, int) {
	if lvl.IsWalkable(cx, cy) {
		return cx, cy
	}
	// BFS spiral outward from center.
	type pt struct{ x, y int }
	visited := map[pt]bool{{cx, cy}: true}
	queue := []pt{{cx, cy}}
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			p := pt{nx, ny}
			if visited[p] || !room.Contains(nx, ny) {
				continue
			}
			visited[p] = true
			if lvl.IsWalkable(nx, ny) {
				return nx, ny
			}
			queue = append(queue, p)
		}
	}
	return -1, -1
}

// spawnBoss creates and places the boss on the final floor.
func (g *Game) spawnBoss(x, y int) {
	boss := entities.NewDungeonGuardian(g.spriteSheet, x, y)
	g.CurrentBoss = boss
	// The boss's embedded Monster is added to the Monsters slice so it gets
	// normal update/draw treatment via collectRenderables.
	g.Monsters = append(g.Monsters, boss.Monster)

	g.BossBar = &hud.BossHealthBar{
		Name:         boss.Monster.Name,
		Title:        boss.Title,
		MaxHP:        boss.Monster.MaxHP,
		CurrentHP:    boss.Monster.HP,
		PhaseMarkers: boss.PhaseHP,
		Visible:      false, // shown once fight activates
	}
}

// activateBoss seals the arena and starts the fight.
func (g *Game) activateBoss() {
	if g.CurrentBoss == nil || g.CurrentBoss.IsActive {
		return
	}
	g.CurrentBoss.IsActive = true
	if g.BossBar != nil {
		g.BossBar.Visible = true
	}
}

// onBossDefeated handles the boss dying: plays any post-fight dialogue, then
// unseals the arena and spawns the exit portal once the dialogue closes.
func (g *Game) onBossDefeated() {
	if g.CurrentBoss == nil {
		return
	}
	g.CurrentBoss.IsActive = false
	if g.BossBar != nil {
		g.BossBar.Visible = false
	}

	// finaliseBossDefeat performs the actions that follow once the post-fight
	// sequence (if any) has finished: save meta, unseal, spawn portal.
	bx, by := g.CurrentBoss.Monster.TileX, g.CurrentBoss.Monster.TileY
	finaliseBossDefeat := func() {
		if g.CurrentBoss != nil && g.CurrentBoss.NPCID != "" && g.Meta != nil {
			npcID := g.CurrentBoss.NPCID
			if g.Meta.NPCMeta[npcID] == nil {
				g.Meta.NPCMeta[npcID] = &NPCMetaState{}
			}
			g.Meta.NPCMeta[npcID].DefeatCount++
			SaveMeta(g.Meta)
		}
		g.unsealBossRoom()
		g.ExitEntity = entities.NewExitEntity(bx, by, g.spriteSheet.Portal, "Portal")
	}

	// If the boss has post-fight dialogue, show it before finalising.
	if g.CurrentBoss.PostFightDialogueID != "" {
		g.triggerPostFightDialogue(g.CurrentBoss.PostFightDialogueID, finaliseBossDefeat)
		return
	}

	finaliseBossDefeat()
}

// sealBossRoom locks all door tiles on the boss room perimeter.
func (g *Game) sealBossRoom() {
	if g.BossRoom == nil || g.currentLevel == nil {
		return
	}
	g.forEachBossRoomDoor(func(x, y int) {
		tile := g.currentLevel.Tile(x, y)
		if tile != nil && tile.HasTag(tiles.TagDoor) {
			tile.DoorState = 3 // locked
			tile.IsWalkable = false
		}
	})
}

// unsealBossRoom opens all door tiles on the boss room perimeter.
func (g *Game) unsealBossRoom() {
	if g.BossRoom == nil || g.currentLevel == nil {
		return
	}
	g.forEachBossRoomDoor(func(x, y int) {
		tile := g.currentLevel.Tile(x, y)
		if tile != nil && tile.HasTag(tiles.TagDoor) {
			tile.DoorState = 1 // open
			tile.IsWalkable = true
		}
	})
}

// forEachBossRoomDoor iterates all tiles on the room's perimeter border.
func (g *Game) forEachBossRoomDoor(fn func(x, y int)) {
	r := g.BossRoom
	// Top and bottom edges.
	for x := r.X; x < r.X+r.W; x++ {
		fn(x, r.Y-1)
		fn(x, r.Y+r.H)
	}
	// Left and right edges.
	for y := r.Y; y < r.Y+r.H; y++ {
		fn(r.X-1, y)
		fn(r.X+r.W, y)
	}
}
