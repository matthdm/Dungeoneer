package game

import (
	"dungeoneer/entities"
	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpawnStrategy controls where an NPC is placed on a floor.
type SpawnStrategy string

const (
	SpawnQuest        SpawnStrategy = "quest"
	SpawnAmbient      SpawnStrategy = "ambient"
	SpawnWandering    SpawnStrategy = "wandering"
	SpawnHidden       SpawnStrategy = "hidden"
	SpawnEntrance     SpawnStrategy = "entrance"
	SpawnExit         SpawnStrategy = "exit"
	SpawnBossAdjacent SpawnStrategy = "boss_adjacent"
)

// NPCTemplate defines an NPC that can be spawned in the world.
type NPCTemplate struct {
	ID            string
	Name          string
	Title         string
	SpriteID      string        // key into SpriteMap
	PortraitID    string        // key into SpriteMap for dialogue portrait
	IsMajor       bool
	DialogueID    string        // dialogue tree ID or SimpleDialogue ID
	Biomes        []Biome       // which biomes this NPC appears in (empty = all)
	Placement     SpawnStrategy // placement tier (default = ambient)
	SpawnChance   float64       // 0.0-1.0, probability per eligible floor (0 = always)
	SpawnMinFloor int           // earliest floor (0 = any)
	SpawnMaxFloor int           // latest floor (0 = no limit)
}

// effectivePlacement returns the NPC's placement strategy, defaulting to ambient.
func (t *NPCTemplate) effectivePlacement() SpawnStrategy {
	if t.Placement == "" {
		return SpawnAmbient
	}
	return t.Placement
}

// placementPriority returns a sort key so higher-priority tiers are placed first.
func placementPriority(s SpawnStrategy) int {
	switch s {
	case SpawnQuest:
		return 0
	case SpawnBossAdjacent:
		return 1
	case SpawnEntrance:
		return 2
	case SpawnExit:
		return 3
	case SpawnAmbient:
		return 4
	case SpawnWandering:
		return 5
	case SpawnHidden:
		return 6
	default:
		return 7
	}
}

// findCorridorTile returns a walkable tile not inside any room.
// Prefers tiles with 3+ walkable neighbors (junctions).
func findCorridorTile(lvl *levels.Level, avoid map[[2]int]bool) (int, int) {
	bestX, bestY := -1, -1
	bestNeighbors := -1
	for y := 1; y < lvl.H-1; y++ {
		for x := 1; x < lvl.W-1; x++ {
			if !lvl.IsWalkable(x, y) {
				continue
			}
			if avoid[[2]int{x, y}] {
				continue
			}
			if lvl.RoomAt(x, y) != nil {
				continue
			}
			n := countWalkableNeighbors(lvl, x, y)
			if n > bestNeighbors {
				bestNeighbors = n
				bestX, bestY = x, y
			}
		}
	}
	return bestX, bestY
}

// findDeadEndTile returns a walkable tile with exactly 1 walkable cardinal
// neighbor, or falls back to a tile in a dead-end room.
func findDeadEndTile(lvl *levels.Level, avoid map[[2]int]bool) (int, int) {
	// First try true dead-end corridor tiles.
	for y := 1; y < lvl.H-1; y++ {
		for x := 1; x < lvl.W-1; x++ {
			if !lvl.IsWalkable(x, y) || avoid[[2]int{x, y}] {
				continue
			}
			if countWalkableNeighbors(lvl, x, y) == 1 {
				return x, y
			}
		}
	}
	// Fallback: center of a dead-end room.
	deadEnds := levels.RoomsByTag(lvl.Rooms, levels.TagDeadEnd)
	for _, r := range deadEnds {
		x, y := r.CenterX, r.CenterY
		if lvl.IsWalkable(x, y) && !avoid[[2]int{x, y}] {
			return x, y
		}
	}
	return -1, -1
}

func countWalkableNeighbors(lvl *levels.Level, x, y int) int {
	n := 0
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		if lvl.IsWalkable(x+d[0], y+d[1]) {
			n++
		}
	}
	return n
}

// MajorNPCPhaseRule controls when and how a major NPC spawns during one phase.
type MajorNPCPhaseRule struct {
	Phase      int    // exact QuestFlag phase value this rule applies to
	MinFloor   int    // earliest floor to spawn on (inclusive)
	MaxFloor   int    // latest floor to spawn on (0 = any)
	SpriteID   string // in-world sprite for this phase
	PortraitID string // dialogue portrait
}

// MajorNPCDef defines a major NPC with cross-run phase-aware spawning.
// DialogueID is left empty so SelectTree picks the phase tree automatically.
type MajorNPCDef struct {
	ID         string
	Name       string
	Title      string
	Placement  SpawnStrategy
	PhaseRules []MajorNPCPhaseRule // one entry per spawnable phase; absent phase = boss/no-spawn
}

// majorNPCDefs lists all major NPCs with their per-phase spawn rules.
// Phase 3+ entries are intentionally omitted — those phases become boss fights.
var majorNPCDefs = []MajorNPCDef{
	{
		ID: "varn", Name: "Varn", Title: "The Chainkeeper",
		Placement: SpawnQuest,
		PhaseRules: []MajorNPCPhaseRule{
			// Phase 0 & 1: GreyKnight — he is restrained, constrained, holding back.
			// Phase 2: Sentinel  — he is visibly different; the transformation is showing.
			// NG+ (DefeatCount >= 1): overridden to TorturedSoul at spawn time for phases 0-1.
			{Phase: 0, MinFloor: 1, MaxFloor: 1, SpriteID: "GreyKnight", PortraitID: "GreyKnight"},                          // floor 1 only — intro, clear the floor
			{Phase: 1, MinFloor: 2, MaxFloor: 5, SpriteID: "GreyKnight", PortraitID: "GreyKnight"},                          // floors 2-5 — Grips quest; item injected into loot
			{Phase: 2, MinFloor: 3, MaxFloor: 6, SpriteID: "Sentinel", PortraitID: "Sentinel"},                               // floors 3-6 — Chaos Emblem quest; item injected into loot
		},
	},
}

// minorNPCPool defines the set of minor NPCs that can appear on dungeon floors.
var minorNPCPool = []NPCTemplate{
	{
		ID: "hollow_monk", Name: "Hollow Monk", SpriteID: "Sorcerer", PortraitID: "Sorcerer",
		DialogueID: "hollow_monk", Biomes: []Biome{BiomeCrypt, BiomeCatacomb},
		Placement: SpawnAmbient,
	},
	{
		ID: "scavenger", Name: "Scavenger", SpriteID: "Caveman", PortraitID: "Caveman",
		DialogueID: "scavenger", Biomes: nil, // all biomes
		Placement: SpawnWandering,
	},
	{
		ID: "forgotten_soldier", Name: "Forgotten Soldier", SpriteID: "GreyKnight", PortraitID: "GreyKnight",
		DialogueID: "forgotten_soldier", Biomes: []Biome{BiomeBrick, BiomeMoss},
		Placement: SpawnAmbient,
	},
	{
		ID: "mad_scholar", Name: "Mad Scholar", SpriteID: "Oracle", PortraitID: "Oracle",
		DialogueID: "mad_scholar", Biomes: []Biome{BiomeGallery},
		Placement: SpawnHidden,
	},
	{
		ID: "weeping_shade", Name: "Weeping Shade", SpriteID: "Apparition", PortraitID: "Apparition",
		DialogueID: "weeping_shade", Biomes: []Biome{BiomeCrypt},
		Placement: SpawnHidden, SpawnChance: 0.5,
	},
}

// filterNPCsByBiome returns templates valid for the given biome.
func filterNPCsByBiome(pool []NPCTemplate, biome Biome) []NPCTemplate {
	var out []NPCTemplate
	for _, t := range pool {
		if len(t.Biomes) == 0 {
			out = append(out, t)
			continue
		}
		for _, b := range t.Biomes {
			if b == biome {
				out = append(out, t)
				break
			}
		}
	}
	return out
}

// createNPCFromTemplate instantiates an NPC entity from a template at the given position.
func (g *Game) createNPCFromTemplate(tmpl NPCTemplate, x, y int) *entities.NPC {
	var sprite *ebiten.Image
	if img, ok := g.SpriteMap[tmpl.SpriteID]; ok {
		sprite = img
	}
	return &entities.NPC{
		ID:            tmpl.ID,
		Name:          tmpl.Name,
		Title:         tmpl.Title,
		TileX:         x,
		TileY:         y,
		InterpX:       float64(x),
		InterpY:       float64(y),
		LeftFacing:    true,
		Sprite:        sprite,
		PortraitID:    tmpl.PortraitID,
		IsMajor:       tmpl.IsMajor,
		DialogueID:    tmpl.DialogueID,
		Interactable:  true,
		InteractRange: 2.0,
		Behavior:      entities.NewIdleBehavior(6),
	}
}
