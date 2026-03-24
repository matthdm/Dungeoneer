package game

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"math/rand/v2"
	"time"
)

// Biome identifies a visual/gameplay theme for a dungeon floor.
type Biome string

const (
	BiomeCrypt    Biome = "crypt"
	BiomeMoss     Biome = "moss"
	BiomeGallery  Biome = "gallery"
	BiomeBrick    Biome = "brick"
	BiomeCatacomb Biome = "catacomb"
)

// availableBiomes lists the biomes we can currently generate with existing
// tilesets. This maps to sprites.WallFlavors entries that have both wall and
// floor sprites.
var availableBiomes = []Biome{BiomeCrypt, BiomeMoss, BiomeGallery, BiomeBrick}

// FloorContext holds the generation parameters and metadata for a single floor.
type FloorContext struct {
	FloorNumber int
	TotalFloors int
	Biome       Biome
	Difficulty  float64 // 0.0–1.0
	GenParams   levels.GenParams
	BiomeConfig *BiomeConfig
}

// RunState tracks all state for a single dungeon run.
type RunState struct {
	Active        bool
	CurrentFloor  int
	TotalFloors   int
	Biomes        []Biome
	KillCount     int
	FloorsCleared int
	RemnantEarned int
	StartTime     time.Time
}

// DefaultRunFloors is the starting number of floors for a new run.
const DefaultRunFloors = 3

// NewRunState initialises a new run with the given floor count.
func NewRunState(totalFloors int) *RunState {
	biomes := assignBiomes(totalFloors)
	return &RunState{
		Active:      true,
		CurrentFloor: 1,
		TotalFloors: totalFloors,
		Biomes:      biomes,
		StartTime:   time.Now(),
	}
}

// assignBiomes creates a biome sequence with no immediate repeats.
func assignBiomes(count int) []Biome {
	biomes := make([]Biome, count)
	prev := Biome("")
	for i := range biomes {
		for {
			b := availableBiomes[rand.IntN(len(availableBiomes))]
			if b != prev || len(availableBiomes) == 1 {
				biomes[i] = b
				prev = b
				break
			}
		}
	}
	return biomes
}

// BuildFloorContext creates generation parameters for the given floor number.
func (rs *RunState) BuildFloorContext(floorNum int) FloorContext {
	biome := BiomeCrypt
	if floorNum-1 < len(rs.Biomes) {
		biome = rs.Biomes[floorNum-1]
	}

	// Difficulty ramps from 0 to 1 over the course of the run.
	difficulty := 0.0
	if rs.TotalFloors > 1 {
		difficulty = float64(floorNum-1) / float64(rs.TotalFloors-1)
	}

	// Scale generation parameters with difficulty.
	roomMin := 6 + int(difficulty*4)  // 6 → 10
	roomMax := 10 + int(difficulty*6) // 10 → 16
	lockChance := 0.15 + difficulty*0.25

	flavor := string(biome)
	// Verify the flavor has wall sprites, fall back to "crypt".
	if _, err := sprites.LoadWallSpriteSheet(flavor); err != nil {
		flavor = "crypt"
	}

	ctx := FloorContext{
		FloorNumber: floorNum,
		TotalFloors: rs.TotalFloors,
		Biome:       biome,
		Difficulty:  difficulty,
		BiomeConfig: BiomeConfigs[biome],
		GenParams: levels.GenParams{
			Seed:           rand.Int64(),
			Width:          64,
			Height:         64,
			RoomCountMin:   roomMin,
			RoomCountMax:   roomMax,
			RoomWMin:       6,
			RoomWMax:       12,
			RoomHMin:       6,
			RoomHMax:       12,
			CorridorWidth:  1,
			DashLaneMinLen: 7,
			GrappleRange:   10,
			Extras:         1 + int(difficulty*2),
			CoverageTarget: 0.40 + difficulty*0.10,
			FillerRoomsMax: 4 + int(difficulty*2),
			DoorLockChance: lockChance,
			WallFlavor:     flavor,
			FloorFlavor:    flavor,
		},
	}

	// Apply biome-specific generation overrides if defined.
	if ctx.BiomeConfig != nil && ctx.BiomeConfig.GenOverrides != nil {
		o := ctx.BiomeConfig.GenOverrides
		if o.RoomCountMin != nil {
			ctx.GenParams.RoomCountMin = *o.RoomCountMin
		}
		if o.RoomCountMax != nil {
			ctx.GenParams.RoomCountMax = *o.RoomCountMax
		}
		if o.RoomWMin != nil {
			ctx.GenParams.RoomWMin = *o.RoomWMin
		}
		if o.RoomWMax != nil {
			ctx.GenParams.RoomWMax = *o.RoomWMax
		}
		if o.CorridorWidth != nil {
			ctx.GenParams.CorridorWidth = *o.CorridorWidth
		}
		if o.DoorLockChance != nil {
			ctx.GenParams.DoorLockChance = *o.DoorLockChance
		}
		if o.CoverageTarget != nil {
			ctx.GenParams.CoverageTarget = *o.CoverageTarget
		}
	}

	return ctx
}

// IsLastFloor returns true if the current floor is the final floor.
func (rs *RunState) IsLastFloor() bool {
	return rs.CurrentFloor >= rs.TotalFloors
}

// CalculateRemnants computes the meta currency earned for this run.
func (rs *RunState) CalculateRemnants() int {
	return rs.FloorsCleared*10 + rs.KillCount*2
}
