package game

import (
	"dungeoneer/entities"
	"dungeoneer/levels"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// EncounterTemplate defines a reusable enemy placement pattern.
type EncounterTemplate struct {
	ID          string
	MinFloor    int
	MaxFloor    int             // 0 = no limit
	MinRoomSize levels.RoomSize // minimum room size to fit this encounter
	Enemies     []EnemySlot
	Weight      float64 // selection probability weight
}

// EnemySlot describes one or more enemies within an encounter template.
type EnemySlot struct {
	Role     string // "melee", "ranged", "elite", "swarm", "caster", "ambush"
	Position string // "room_center", "room_back", "room_front", "room_edges", "room_scattered"
	Behavior string // override: "ambush", "roaming", "patrol", "ranged", "swarm", "" = use EnemyDef default
	Count    int    // for multi-spawn slots (default 1)
}

// encounterTemplates defines the built-in encounter patterns.
var encounterTemplates = []EncounterTemplate{
	{
		ID: "solo_guardian", MinFloor: 1, MinRoomSize: levels.RoomSmall, Weight: 3.0,
		Enemies: []EnemySlot{
			{Role: "melee", Position: "room_center", Count: 1},
		},
	},
	{
		ID: "ambush_pair", MinFloor: 1, MinRoomSize: levels.RoomSmall, Weight: 2.0,
		Enemies: []EnemySlot{
			{Role: "melee", Position: "room_edges", Behavior: "ambush", Count: 2},
		},
	},
	{
		ID: "firing_line", MinFloor: 1, MinRoomSize: levels.RoomMedium, Weight: 2.5,
		Enemies: []EnemySlot{
			{Role: "ranged", Position: "room_back", Count: 1},
			{Role: "melee", Position: "room_front", Count: 2},
		},
	},
	{
		ID: "corridor_patrol", MinFloor: 1, MinRoomSize: levels.RoomSmall, Weight: 1.5,
		Enemies: []EnemySlot{
			{Role: "melee", Position: "room_center", Behavior: "patrol", Count: 1},
		},
	},
	{
		ID: "swarm_room", MinFloor: 2, MinRoomSize: levels.RoomMedium, Weight: 2.0,
		Enemies: []EnemySlot{
			{Role: "swarm", Position: "room_scattered", Count: 5},
		},
	},
	{
		ID: "elite_and_adds", MinFloor: 3, MinRoomSize: levels.RoomLarge, Weight: 1.0,
		Enemies: []EnemySlot{
			{Role: "elite", Position: "room_center", Count: 1},
			{Role: "melee", Position: "room_edges", Count: 2},
		},
	},
	{
		ID: "caster_den", MinFloor: 2, MinRoomSize: levels.RoomMedium, Weight: 2.0,
		Enemies: []EnemySlot{
			{Role: "caster", Position: "room_back", Behavior: "caster", Count: 1},
			{Role: "melee", Position: "room_front", Count: 1},
		},
	},
	{
		ID: "ranged_nest", MinFloor: 1, MinRoomSize: levels.RoomSmall, Weight: 1.5,
		Enemies: []EnemySlot{
			{Role: "ranged", Position: "room_center", Count: 2},
		},
	},
}

// enemyBudget returns the max enemies for a floor.
func enemyBudget(floorNumber int) int {
	b := 8 + floorNumber*3
	if b > 25 {
		b = 25
	}
	return b + rand.IntN(4)
}

// slotEnemyCount returns the actual count for a slot (defaults to 1).
func slotEnemyCount(s EnemySlot) int {
	if s.Count <= 0 {
		return 1
	}
	return s.Count
}

// totalTemplateEnemies returns how many enemies a template spawns.
func totalTemplateEnemies(t EncounterTemplate) int {
	n := 0
	for _, s := range t.Enemies {
		n += slotEnemyCount(s)
	}
	return n
}

// eligibleTemplates filters templates by floor number and room size.
func eligibleTemplates(floor int, roomSize levels.RoomSize) []EncounterTemplate {
	sizeRank := map[levels.RoomSize]int{
		levels.RoomSmall: 0, levels.RoomMedium: 1, levels.RoomLarge: 2,
	}
	roomRank := sizeRank[roomSize]

	var out []EncounterTemplate
	for _, t := range encounterTemplates {
		if floor < t.MinFloor {
			continue
		}
		if t.MaxFloor > 0 && floor > t.MaxFloor {
			continue
		}
		if sizeRank[t.MinRoomSize] > roomRank {
			continue
		}
		out = append(out, t)
	}
	return out
}

// pickTemplate does weighted random selection from eligible templates.
func pickTemplate(templates []EncounterTemplate) *EncounterTemplate {
	if len(templates) == 0 {
		return nil
	}
	total := 0.0
	for _, t := range templates {
		total += t.Weight
	}
	r := rand.Float64() * total
	for i := range templates {
		r -= templates[i].Weight
		if r <= 0 {
			return &templates[i]
		}
	}
	return &templates[len(templates)-1]
}

// resolvePosition picks a walkable tile within the room for a position string.
// occupied tracks used tiles to prevent stacking.
func resolvePosition(room *levels.Room, pos string, level *levels.Level, occupied map[[2]int]bool) (int, int, bool) {
	// Helper: find a random walkable, unoccupied tile in a rect region.
	findIn := func(x0, y0, x1, y1 int) (int, int, bool) {
		candidates := [][2]int{}
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if level.IsWalkable(x, y) && !occupied[[2]int{x, y}] {
					candidates = append(candidates, [2]int{x, y})
				}
			}
		}
		if len(candidates) == 0 {
			return 0, 0, false
		}
		c := candidates[rand.IntN(len(candidates))]
		return c[0], c[1], true
	}

	cx, cy := room.CenterX, room.CenterY
	switch pos {
	case "room_center":
		// Try center, then nearby.
		if level.IsWalkable(cx, cy) && !occupied[[2]int{cx, cy}] {
			return cx, cy, true
		}
		return findIn(cx-2, cy-2, cx+3, cy+3)

	case "room_back":
		// Back = high Y within the room.
		backY := room.Y + room.H - 2
		return findIn(room.X+1, backY, room.X+room.W-1, room.Y+room.H-1)

	case "room_front":
		// Front = low Y within the room.
		return findIn(room.X+1, room.Y+1, room.X+room.W-1, room.Y+room.H/2)

	case "room_edges":
		// 1-tile border inside the room.
		candidates := [][2]int{}
		for y := room.Y; y < room.Y+room.H; y++ {
			for x := room.X; x < room.X+room.W; x++ {
				if (x == room.X || x == room.X+room.W-1 || y == room.Y || y == room.Y+room.H-1) &&
					level.IsWalkable(x, y) && !occupied[[2]int{x, y}] {
					candidates = append(candidates, [2]int{x, y})
				}
			}
		}
		if len(candidates) == 0 {
			return findIn(room.X, room.Y, room.X+room.W, room.Y+room.H)
		}
		c := candidates[rand.IntN(len(candidates))]
		return c[0], c[1], true

	case "room_scattered":
		return findIn(room.X, room.Y, room.X+room.W, room.Y+room.H)

	default:
		return findIn(room.X, room.Y, room.X+room.W, room.Y+room.H)
	}
}

// makeBehavior creates the appropriate MonsterBehavior for a behavior string.
func makeBehavior(behavior string) entities.MonsterBehavior {
	switch behavior {
	case "ambush":
		return entities.NewAmbushBehavior(4)
	case "patrol":
		return entities.NewPatrolBehavior(5)
	case "ranged":
		return entities.NewRangedBehavior(6)
	case "swarm":
		return entities.NewSwarmBehavior(4)
	case "caster":
		return entities.NewCasterBehavior(7)
	default:
		return entities.NewRoamingWanderBehavior(5)
	}
}

// spawnEncounterMonsters places monsters using the encounter template system.
func (g *Game) spawnEncounterMonsters(ctx FloorContext) {
	if ctx.BiomeConfig == nil || len(g.currentLevel.Rooms) == 0 {
		// Fallback to legacy spawner if no biome config or rooms.
		g.spawnFloorMonsters(ctx)
		return
	}

	budget := enemyBudget(ctx.FloorNumber)
	occupied := map[[2]int]bool{
		{g.player.TileX, g.player.TileY}: true,
	}
	if g.ExitEntity != nil {
		occupied[[2]int{g.ExitEntity.TileX, g.ExitEntity.TileY}] = true
	}

	// Mark tiles near player as off-limits.
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			occupied[[2]int{g.player.TileX + dx, g.player.TileY + dy}] = true
		}
	}

	// Shuffle rooms so placement varies each run.
	rooms := make([]levels.Room, len(g.currentLevel.Rooms))
	copy(rooms, g.currentLevel.Rooms)
	rand.Shuffle(len(rooms), func(i, j int) { rooms[i], rooms[j] = rooms[j], rooms[i] })

	// Skip the room containing the player spawn.
	playerRoom := g.currentLevel.RoomAt(g.player.TileX, g.player.TileY)

	for i := range rooms {
		if budget <= 0 {
			break
		}
		room := &rooms[i]
		if playerRoom != nil && room.Index == playerRoom.Index {
			continue // don't populate spawn room
		}
		if room.HasTag(levels.TagCleared) {
			continue // sanctuary / NPC rooms are monster-free
		}

		eligible := eligibleTemplates(ctx.FloorNumber, room.Size)
		tmpl := pickTemplate(eligible)
		if tmpl == nil {
			continue
		}

		cost := totalTemplateEnemies(*tmpl)
		if cost > budget {
			continue
		}

		monsters := g.spawnTemplateInRoom(tmpl, room, ctx, occupied)
		g.Monsters = append(g.Monsters, monsters...)
		budget -= len(monsters)
	}
}

// spawnTemplateInRoom places all enemies from a template into the given room.
func (g *Game) spawnTemplateInRoom(tmpl *EncounterTemplate, room *levels.Room, ctx FloorContext, occupied map[[2]int]bool) []*entities.Monster {
	var spawned []*entities.Monster
	var swarmGroup []*entities.Monster

	for _, slot := range tmpl.Enemies {
		count := slotEnemyCount(slot)
		enemyDef := ctx.BiomeConfig.EnemyByRole(slot.Role)
		if enemyDef == nil {
			// Fall back to melee if role not found.
			enemyDef = ctx.BiomeConfig.EnemyByRole("melee")
		}
		if enemyDef == nil {
			continue
		}

		sprite := g.resolveSprite(enemyDef.SpriteID)
		if sprite == nil {
			continue
		}

		for j := 0; j < count; j++ {
			x, y, ok := resolvePosition(room, slot.Position, g.currentLevel, occupied)
			if !ok {
				continue
			}
			occupied[[2]int{x, y}] = true

			behaviorStr := slot.Behavior
			if behaviorStr == "" {
				behaviorStr = enemyDef.Behavior
			}

			hpScale := 1.0 + ctx.Difficulty*0.5
			dmgScale := 1.0 + ctx.Difficulty*0.3

			m := &entities.Monster{
				Name:             enemyDef.Name,
				TileX:            x,
				TileY:            y,
				InterpX:          float64(x),
				InterpY:          float64(y),
				Sprite:           sprite,
				MovementDuration: enemyDef.BaseSpeed,
				LeftFacing:       true,
				HP:               int(float64(enemyDef.BaseHP) * hpScale),
				MaxHP:            int(float64(enemyDef.BaseHP) * hpScale),
				Damage:           int(float64(enemyDef.BaseDamage) * dmgScale),
				HitRadius:        entities.DefaultMonsterHitRadius,
				AttackRate:       enemyDef.AttackRate,
				Behavior:         makeBehavior(behaviorStr),
				Level:            ctx.FloorNumber,
				Role:             slot.Role,
			}

			// Set patrol waypoints for patrol behavior.
			if pb, ok := m.Behavior.(*entities.PatrolBehavior); ok {
				pb.Waypoints = []entities.PatrolWaypoint{
					{X: room.X + 1, Y: room.Y + 1},
					{X: room.X + room.W - 2, Y: room.Y + room.H - 2},
				}
			}

			if slot.Role == "swarm" {
				swarmGroup = append(swarmGroup, m)
			}
			spawned = append(spawned, m)
		}
	}

	// Link swarm siblings.
	for _, m := range swarmGroup {
		m.Siblings = swarmGroup
	}

	return spawned
}

// resolveSprite looks up a sprite by its SpriteID string.
func (g *Game) resolveSprite(spriteID string) *ebiten.Image {
	if g.SpriteMap != nil {
		if img, ok := g.SpriteMap[spriteID]; ok {
			return img
		}
	}
	return nil
}
