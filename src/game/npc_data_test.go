package game

import (
	"dungeoneer/levels"
	"testing"
)

func TestNPCPlacementPriority(t *testing.T) {
	if placementPriority(SpawnQuest) >= placementPriority(SpawnAmbient) {
		t.Errorf("SpawnQuest should have higher priority (lower value) than SpawnAmbient")
	}
	if placementPriority(SpawnBossAdjacent) >= placementPriority(SpawnEntrance) {
		t.Errorf("SpawnBossAdjacent should have higher priority than SpawnEntrance")
	}
}

func TestFilterNPCsByBiome(t *testing.T) {
	pool := []NPCTemplate{
		{ID: "all_biomes"},
		{ID: "only_crypt", Biomes: []Biome{BiomeCrypt}},
		{ID: "only_moss", Biomes: []Biome{BiomeMoss}},
	}
	
	cryptNPCs := filterNPCsByBiome(pool, BiomeCrypt)
	if len(cryptNPCs) != 2 {
		t.Errorf("expected 2 crypt NPCs, got %d", len(cryptNPCs))
	}
	
	mossNPCs := filterNPCsByBiome(pool, BiomeMoss)
	if len(mossNPCs) != 2 {
		t.Errorf("expected 2 moss NPCs, got %d", len(mossNPCs))
	}
	
	brickNPCs := filterNPCsByBiome(pool, BiomeBrick)
	if len(brickNPCs) != 1 {
		t.Errorf("expected 1 brick NPC, got %d", len(brickNPCs))
	}
}

func TestFindCorridorTile(t *testing.T) {
	lvl := levels.NewEmptyLevel(10, 10)
	for x := 1; x < 9; x++ {
		for y := 1; y < 9; y++ {
			lvl.Tile(x, y).IsWalkable = true
		}
	}
	
	avoid := make(map[[2]int]bool)
	x, y := findCorridorTile(lvl, avoid)
	if x < 1 || x > 8 || y < 1 || y > 8 {
		t.Errorf("expected valid tile, got %d, %d", x, y)
	}
}

func TestFindDeadEndTile(t *testing.T) {
	lvl := levels.NewEmptyLevel(10, 10)
	
	// Make a dead end corridor at (2, 2) connected to (3, 2)
	lvl.Tile(2, 2).IsWalkable = true
	lvl.Tile(3, 2).IsWalkable = true
	
	avoid := make(map[[2]int]bool)
	x, y := findDeadEndTile(lvl, avoid)
	if x != 2 || y != 2 {
		t.Errorf("expected (2, 2) dead end, got %d, %d", x, y)
	}
}

func TestCreateNPCFromTemplate(t *testing.T) {
	g := &Game{}
	tmpl := minorNPCPool[0]
	
	npc := g.createNPCFromTemplate(tmpl, 5, 5)
	if npc == nil {
		t.Fatalf("expected NPC")
	}
	if npc.TileX != 5 || npc.TileY != 5 {
		t.Errorf("expected pos 5, 5")
	}
	if npc.ID != tmpl.ID {
		t.Errorf("expected ID %s", tmpl.ID)
	}
}

func TestBuildPhaseTracker(t *testing.T) {
	def := majorNPCDefs[0]
	tracker := def.BuildPhaseTracker()
	
	if tracker.NPCID != def.ID {
		t.Errorf("expected NPCID %s", def.ID)
	}
	if tracker.MaxPhase != len(def.PhaseRules) {
		t.Errorf("expected MaxPhase %d", len(def.PhaseRules))
	}
}
