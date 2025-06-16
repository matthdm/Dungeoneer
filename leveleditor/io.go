package leveleditor

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"dungeoneer/tiles"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type SpriteMetadata struct {
	Image      *ebiten.Image
	IsWalkable bool
}
type TileData struct {
	SpriteIndexes []int `json:"sprite_indexes"`
	IsWalkable    bool  `json:"is_walkable"`
}

type LevelData struct {
	Width         int          `json:"width"`
	Height        int          `json:"height"`
	TileSize      int          `json:"tile_size"`
	Tiles         [][]TileData `json:"tiles"`
	SpritePalette []string     `json:"sprite_palette"`
}

var SpriteRegistry = map[string]SpriteMetadata{}
var ReverseSpriteRegistry = map[*ebiten.Image]string{}

// RegisterWallSprites adds all sprites from the provided WallSpriteSheet using
// a flavor-prefixed identifier. For example a flavor of "crypt" will register
// IDs such as "crypt_wall" and "crypt_floor".
func RegisterWallSprites(wss *sprites.WallSpriteSheet) {
	flavor := wss.Flavor
	add := func(name string, img *ebiten.Image, walkable bool) {
		id := flavor + "_" + name
		SpriteRegistry[id] = SpriteMetadata{Image: img, IsWalkable: walkable}
		ReverseSpriteRegistry[img] = id
	}

	add("beam", wss.Beam, false)
	add("beam_nw", wss.BeamNW, false)
	add("beam_ne", wss.BeamNE, false)
	add("log", wss.Log, false)
	add("beam_sw", wss.BeamSW, false)
	add("log_slim", wss.LogSlim, false)
	add("beam_nesw", wss.BeamNESW, false)
	add("log_ne", wss.LogNE, false)
	add("beam_se", wss.BeamSE, false)
	add("log_nwse", wss.LogNWSE, false)
	add("log_slim2", wss.LogSlim2, false)
	add("wall_nesw", wss.WallNESW, false)
	add("wall", wss.Wall, false)
	add("wall_nwse", wss.WallNWSE, false)
	add("log_sw", wss.LogSW, false)
	add("chunk", wss.Chunk, false)
	add("door_locked_nw", wss.LockedDoorNW, false)
	add("door_locked_ne", wss.LockedDoorNE, false)
	add("door_unlocked_nw", wss.UnlockedDoorNW, false)
	add("door_unlocked_ne", wss.UnlockedDoorNE, false)
	add("floor", wss.Floor, true)
}

func RegisterSprites(ss *sprites.SpriteSheet) {
	SpriteRegistry = map[string]SpriteMetadata{
		"Void": {
			Image:      ss.Void,
			IsWalkable: false,
		},
		"OakBeam": {
			Image:      ss.OakBeam,
			IsWalkable: false,
		},
		"OakBeamNW": {
			Image:      ss.OakBeamNW,
			IsWalkable: false,
		},
		"OakBeamNE": {
			Image:      ss.OakBeamNE,
			IsWalkable: false,
		},
		"OakLog": {
			Image:      ss.OakLog,
			IsWalkable: false,
		},
		"OakBeamSW": {
			Image:      ss.OakBeamSW,
			IsWalkable: false,
		},
		"OakLogSlim": {
			Image:      ss.OakLogSlim,
			IsWalkable: false,
		},
		"OakBeamNESW": {
			Image:      ss.OakBeamNESW,
			IsWalkable: false,
		},
		"OakLogNE": {
			Image:      ss.OakLogNE,
			IsWalkable: false,
		},
		"OakBeamSE": {
			Image:      ss.OakBeamSE,
			IsWalkable: false,
		},
		"OakLogNWSE": {
			Image:      ss.OakLogNWSE,
			IsWalkable: false,
		},
		"OakLogSlim2": {
			Image:      ss.OakLogSlim2,
			IsWalkable: false,
		},
		"OakWallNESW": {
			Image:      ss.OakWallNESW,
			IsWalkable: false,
		},
		"OakWall": {
			Image:      ss.OakWall,
			IsWalkable: false,
		},
		"OakWallNWSE": {
			Image:      ss.OakWallNWSE,
			IsWalkable: false,
		},
		"OakLogSW": {
			Image:      ss.OakLogSW,
			IsWalkable: false,
		},
		"OakChunk": {
			Image:      ss.OakChunk,
			IsWalkable: false,
		},
		"MarbleBeam": {
			Image:      ss.MarbleBeam,
			IsWalkable: false,
		},
		"MarbleBeamNW": {
			Image:      ss.MarbleBeamNW,
			IsWalkable: false,
		},
		"MarbleBeamNE": {
			Image:      ss.MarbleBeamNE,
			IsWalkable: false,
		},
		"MarbleWall": {
			Image:      ss.MarbleWall,
			IsWalkable: false,
		},
		"MarbleBeamNE2": {
			Image:      ss.MarbleBeamNE2,
			IsWalkable: false,
		},
		"MarbleWall2": {
			Image:      ss.MarbleWall2,
			IsWalkable: false,
		},
		"MarbleBeamNELong": {
			Image:      ss.MarbleBeamNELong,
			IsWalkable: false,
		},
		"MarbleWallNELong": {
			Image:      ss.MarbleWallNELong,
			IsWalkable: false,
		},
		"MarbleBeamNW2": {
			Image:      ss.MarbleBeamNW2,
			IsWalkable: false,
		},
		"MarbleBeamNWLong": {
			Image:      ss.MarbleBeamNWLong,
			IsWalkable: false,
		},
		"MarbleWall3": {
			Image:      ss.MarbleWall3,
			IsWalkable: false,
		},
		"MarbleWallNWLong": {
			Image:      ss.MarbleWallNWLong,
			IsWalkable: false,
		},
		"MarbleWall4": {
			Image:      ss.MarbleWall4,
			IsWalkable: false,
		},
		"MarbleWallNW": {
			Image:      ss.MarbleWallNW,
			IsWalkable: false,
		},
		"MarbleWallNE": {
			Image:      ss.MarbleWallNE,
			IsWalkable: false,
		},
		"MarbleChunk": {
			Image:      ss.MarbleChunk,
			IsWalkable: false,
		},
		"LockedDoorNW": {
			Image:      ss.LockedDoorNW,
			IsWalkable: false,
		},
		"LockedDoorNE": {
			Image:      ss.LockedDoorNE,
			IsWalkable: false,
		},
		"DungeonBeam": {
			Image:      ss.DungeonBeam,
			IsWalkable: false,
		},
		"DungeonBeamNW": {
			Image:      ss.DungeonBeamNW,
			IsWalkable: false,
		},
		"DungeonBeamNE": {
			Image:      ss.DungeonBeamNE,
			IsWalkable: false,
		},
		"DungeonWall": {
			Image:      ss.DungeonWall,
			IsWalkable: false,
		},
		"DungeonBeamNE2": {
			Image:      ss.DungeonBeamNE2,
			IsWalkable: false,
		},
		"DungeonWall2": {
			Image:      ss.DungeonWall2,
			IsWalkable: false,
		},
		"DungeonBeamNELong": {
			Image:      ss.DungeonBeamNELong,
			IsWalkable: false,
		},
		"DungeonWallNELong": {
			Image:      ss.DungeonWallNELong,
			IsWalkable: false,
		},
		"DungeonBeamNW2": {
			Image:      ss.DungeonBeamNW2,
			IsWalkable: false,
		},
		"DungeonBeamNWLong": {
			Image:      ss.DungeonBeamNWLong,
			IsWalkable: false,
		},
		"DungeonWall3": {
			Image:      ss.DungeonWall3,
			IsWalkable: false,
		},
		"DungeonWallNWLong": {
			Image:      ss.DungeonWallNWLong,
			IsWalkable: false,
		},
		"DungeonWall4": {
			Image:      ss.DungeonWall4,
			IsWalkable: false,
		},
		"DungeonWallNWLong2": {
			Image:      ss.DungeonWallNWLong2,
			IsWalkable: false,
		},
		"DungeonWallNELong2": {
			Image:      ss.DungeonWallNELong2,
			IsWalkable: false,
		},
		"DungeonChunk": {
			Image:      ss.DungeonChunk,
			IsWalkable: false,
		},
		"EmeraldBeam": {
			Image:      ss.EmeraldBeam,
			IsWalkable: false,
		},
		"EmeraldBeamNW": {
			Image:      ss.EmeraldBeamNW,
			IsWalkable: false,
		},
		"EmeraldBeamNE": {
			Image:      ss.EmeraldBeamNE,
			IsWalkable: false,
		},
		"EmeraldWall": {
			Image:      ss.EmeraldWall,
			IsWalkable: false,
		},
		"EmeraldBeamNE2": {
			Image:      ss.EmeraldBeamNE2,
			IsWalkable: false,
		},
		"EmeraldWall2": {
			Image:      ss.EmeraldWall2,
			IsWalkable: false,
		},
		"EmeraldBeamNELong": {
			Image:      ss.EmeraldBeamNELong,
			IsWalkable: false,
		},
		"EmeraldWallNELong": {
			Image:      ss.EmeraldWallNELong,
			IsWalkable: false,
		},
		"EmeraldBeamNW2": {
			Image:      ss.EmeraldBeamNW2,
			IsWalkable: false,
		},
		"EmeraldBeamNWLong": {
			Image:      ss.EmeraldBeamNWLong,
			IsWalkable: false,
		},
		"EmeraldWall3": {
			Image:      ss.EmeraldWall3,
			IsWalkable: false,
		},
		"EmeraldWallNWLong": {
			Image:      ss.EmeraldWallNWLong,
			IsWalkable: false,
		},
		"EmeraldWall4": {
			Image:      ss.EmeraldWall4,
			IsWalkable: false,
		},
		"EmerladWallNWLong2": {
			Image:      ss.EmerladWallNWLong2,
			IsWalkable: false,
		},
		"EmerladWallNELong2": {
			Image:      ss.EmerladWallNELong2,
			IsWalkable: false,
		},
		"EmerladChunk": {
			Image:      ss.EmerladChunk,
			IsWalkable: false,
		},
		"Well": {
			Image:      ss.Well,
			IsWalkable: false,
		},
		"DragonStatue": {
			Image:      ss.DragonStatue,
			IsWalkable: false,
		},
		"Statue": {
			Image:      ss.Statue,
			IsWalkable: false,
		},
		"ChaosStatue": {
			Image:      ss.ChaosStatue,
			IsWalkable: false,
		},
		"Lava": {
			Image:      ss.Lava,
			IsWalkable: false,
		},
		"Water": {
			Image:      ss.Water,
			IsWalkable: false,
		},
		"EnchantedWater": {
			Image:      ss.EnchantedWater,
			IsWalkable: false,
		},
		"Floor": {
			Image:      ss.Floor,
			IsWalkable: true,
		},
		"UnlockedDoorNW": {
			Image:      ss.UnlockedDoorNW,
			IsWalkable: false,
		},
		"UnlockedDoorNE": {
			Image:      ss.UnlockedDoorNE,
			IsWalkable: false,
		},
		"StairsDecending": {
			Image:      ss.StairsDecending,
			IsWalkable: false,
		},
		"FloorTrap": {
			Image:      ss.FloorTrap,
			IsWalkable: true,
		},
		"Pentagram": {
			Image:      ss.Pentagram,
			IsWalkable: true,
		},
		"SkullHex": {
			Image:      ss.SkullHex,
			IsWalkable: true,
		},
		"StairsDecending2": {
			Image:      ss.StairsDecending2,
			IsWalkable: true,
		},
		"StairsDecending3": {
			Image:      ss.StairsDecending3,
			IsWalkable: true,
		},
		"StairsDecending4": {
			Image:      ss.StairsDecending4,
			IsWalkable: true,
		},
		"StairsDecending5": {
			Image:      ss.StairsDecending5,
			IsWalkable: true,
		},
		"StairsAscending": {
			Image:      ss.StairsAscending,
			IsWalkable: true,
		},
		"StairsAscending2": {
			Image:      ss.StairsAscending2,
			IsWalkable: true,
		},
		"StairsAscending3": {
			Image:      ss.StairsAscending3,
			IsWalkable: true,
		},
		"StairsAscending4": {
			Image:      ss.StairsAscending4,
			IsWalkable: true,
		},
		"StairsDescendingYellow": {
			Image:      ss.StairsDescendingYellow,
			IsWalkable: true,
		},
		"StairsAscendingYellow": {
			Image:      ss.StairsAscendingYellow,
			IsWalkable: true,
		},
		"ArchGlitch": {
			Image:      ss.ArchGlitch,
			IsWalkable: true,
		},
		"ArchBlack": {
			Image:      ss.ArchBlack,
			IsWalkable: true,
		},
		"ArchRed": {
			Image:      ss.ArchRed,
			IsWalkable: true,
		},
		"ArchBlue": {
			Image:      ss.ArchBlue,
			IsWalkable: true,
		},
		"ArchGreen": {
			Image:      ss.ArchGreen,
			IsWalkable: true,
		},
		"ArchDeath": {
			Image:      ss.ArchDeath,
			IsWalkable: true,
		},
		"ArchStar": {
			Image:      ss.ArchStar,
			IsWalkable: true,
		},
		"ArchEmpty": {
			Image:      ss.ArchEmpty,
			IsWalkable: true,
		},
		"ArchBoss": {
			Image:      ss.ArchBoss,
			IsWalkable: true,
		},
		"ArchStar2": {
			Image:      ss.ArchStar2,
			IsWalkable: true,
		},
		"Portal": {
			Image:      ss.Portal,
			IsWalkable: true,
		},
		"ArchDark": {
			Image:      ss.ArchDark,
			IsWalkable: true,
		},
		"HandOfGod": {
			Image:      ss.HandOfGod,
			IsWalkable: true,
		},
		"GoldenStatue": {
			Image:      ss.GoldenStatue,
			IsWalkable: false,
		},
		"DeadPile": {
			Image:      ss.DeadPile,
			IsWalkable: false,
		},
		"DeadHanging": {
			Image:      ss.DeadHanging,
			IsWalkable: false,
		},
		"SansStatue": {
			Image:      ss.SansStatue,
			IsWalkable: false,
		},
		"Glyph": {
			Image:      ss.Glyph,
			IsWalkable: false,
		},
		"DeathMarker": {
			Image:      ss.DeathMarker,
			IsWalkable: false,
		},
		"Campfire": {
			Image:      ss.Campfire,
			IsWalkable: false,
		},
		"GlyphStatue": {
			Image:      ss.GlyphStatue,
			IsWalkable: false,
		},
		"DecapitatedHead": {
			Image:      ss.DecapitatedHead,
			IsWalkable: false,
		},
		"GrandChest": {
			Image:      ss.GrandChest,
			IsWalkable: true,
		},
		"Chalice": {
			Image:      ss.Chalice,
			IsWalkable: false,
		},
		"Trinket": {
			Image:      ss.Trinket,
			IsWalkable: false,
		},
		"GreyKnight": {
			Image:      ss.GreyKnight,
			IsWalkable: false,
		},
		"Chimera": {
			Image:      ss.Chimera,
			IsWalkable: false,
		},
		"Sentinel": {
			Image:      ss.Sentinel,
			IsWalkable: false,
		},
		"Sorcerer": {
			Image:      ss.Sorcerer,
			IsWalkable: false,
		},
		"Duchess": {
			Image:      ss.Duchess,
			IsWalkable: false,
		},
		"Absolem": {
			Image:      ss.Absolem,
			IsWalkable: false,
		},
		"Death": {
			Image:      ss.Death,
			IsWalkable: false,
		},
		"Oracle": {
			Image:      ss.Oracle,
			IsWalkable: false,
		},
		"Jester": {
			Image:      ss.Jester,
			IsWalkable: false,
		},
		"GreaterDemon": {
			Image:      ss.GreaterDemon,
			IsWalkable: false,
		},
		"DemonKnight": {
			Image:      ss.DemonKnight,
			IsWalkable: false,
		},
		"Abomination": {
			Image:      ss.Abomination,
			IsWalkable: false,
		},
		"QueenOfDarkness": {
			Image:      ss.QueenOfDarkness,
			IsWalkable: false,
		},
		"LesserDemon": {
			Image:      ss.LesserDemon,
			IsWalkable: false,
		},
		"TheTerror": {
			Image:      ss.TheTerror,
			IsWalkable: false,
		},
		"Celestial": {
			Image:      ss.Celestial,
			IsWalkable: false,
		},
		"Demon": {
			Image:      ss.Demon,
			IsWalkable: false,
		},
		"Apparition": {
			Image:      ss.Apparition,
			IsWalkable: false,
		},
		"Griffon": {
			Image:      ss.Griffon,
			IsWalkable: false,
		},
		"Manticore": {
			Image:      ss.Manticore,
			IsWalkable: false,
		},
		"Minotaur": {
			Image:      ss.Minotaur,
			IsWalkable: false,
		},
		"TorturedSoul": {
			Image:      ss.TorturedSoul,
			IsWalkable: false,
		},
		"ChaosTotem": {
			Image:      ss.ChaosTotem,
			IsWalkable: false,
		},
		"HydraBase": {
			Image:      ss.HydraBase,
			IsWalkable: false,
		},
		"Hydra2": {
			Image:      ss.Hydra2,
			IsWalkable: false,
		},
		"Hydra3": {
			Image:      ss.Hydra3,
			IsWalkable: false,
		},
		"Hydra4": {
			Image:      ss.Hydra4,
			IsWalkable: false,
		},
		"Hydra5": {
			Image:      ss.Hydra5,
			IsWalkable: false,
		},
		"Hydra6": {
			Image:      ss.Hydra6,
			IsWalkable: false,
		},
		"HydraFinal": {
			Image:      ss.HydraFinal,
			IsWalkable: false,
		},
		"LesserDragon": {
			Image:      ss.LesserDragon,
			IsWalkable: false,
		},
		"GoldenDragon": {
			Image:      ss.GoldenDragon,
			IsWalkable: false,
		},
		"Wyvern": {
			Image:      ss.Wyvern,
			IsWalkable: false,
		},
		"LesserCaveDragon": {
			Image:      ss.LesserCaveDragon,
			IsWalkable: false,
		},
		"GhostWyvern": {
			Image:      ss.GhostWyvern,
			IsWalkable: false,
		},
		"AncientDragonRed": {
			Image:      ss.AncientDragonRed,
			IsWalkable: false,
		},
		"AncientDragonBlack": {
			Image:      ss.AncientDragonBlack,
			IsWalkable: false,
		},
		"AncientDragon": {
			Image:      ss.AncientDragon,
			IsWalkable: false,
		},
		"PetrifiedDragon": {
			Image:      ss.PetrifiedDragon,
			IsWalkable: false,
		},
		"GreaterDragon": {
			Image:      ss.GreaterDragon,
			IsWalkable: false,
		},
		"GreaterRedDragon": {
			Image:      ss.GreaterRedDragon,
			IsWalkable: false,
		},
		"BlueMan": {
			Image:      ss.BlueMan,
			IsWalkable: false,
		},
		"Cyclops": {
			Image:      ss.Cyclops,
			IsWalkable: false,
		},
		"TwoHeadedOgre": {
			Image:      ss.TwoHeadedOgre,
			IsWalkable: false,
		},
		"RedChampion": {
			Image:      ss.RedChampion,
			IsWalkable: false,
		},
		"BlueChampion": {
			Image:      ss.BlueChampion,
			IsWalkable: false,
		},
		"Caveman": {
			Image:      ss.Caveman,
			IsWalkable: false,
		},
		"RockCollector": {
			Image:      ss.RockCollector,
			IsWalkable: false,
		},
		"RedMan": {
			Image:      ss.RedMan,
			IsWalkable: false,
		},
		"Cursor": {
			Image:      ss.Cursor,
			IsWalkable: true,
		},
		"EnemyCursor": {
			Image:      ss.EnemyCursor,
			IsWalkable: true,
		},
		"NullSprite": {
			Image:      ss.NullSprite,
			IsWalkable: true,
		},
		"FireBurst": {
			Image:      ss.FireBurst,
			IsWalkable: true,
		},
		"FireBurst2": {
			Image:      ss.FireBurst2,
			IsWalkable: true,
		},
		"FireBurst3": {
			Image:      ss.FireBurst3,
			IsWalkable: true,
		},
		"ArcaneBurst": {
			Image:      ss.ArcaneBurst,
			IsWalkable: true,
		},
		"ArcaneBurst2": {
			Image:      ss.ArcaneBurst2,
			IsWalkable: true,
		},
		"ArcaneBurst3": {
			Image:      ss.ArcaneBurst3,
			IsWalkable: true,
		},
		"PosionBurst": {
			Image:      ss.PosionBurst,
			IsWalkable: true,
		},
		"PosionBurst2": {
			Image:      ss.PosionBurst2,
			IsWalkable: true,
		},
		"PosionBurst3": {
			Image:      ss.PosionBurst3,
			IsWalkable: true,
		},
		"Smoke": {
			Image:      ss.Smoke,
			IsWalkable: true,
		},
		"BlueSmoke": {
			Image:      ss.BlueSmoke,
			IsWalkable: true,
		},
		"BigSmoke": {
			Image:      ss.BigSmoke,
			IsWalkable: true,
		},
		"Smoke2": {
			Image:      ss.Smoke2,
			IsWalkable: true,
		},
		"PurpleSmoke": {
			Image:      ss.PurpleSmoke,
			IsWalkable: true,
		},
		"Trap": {
			Image:      ss.Trap,
			IsWalkable: true,
		},
	}

	// Build reverse lookup
	for id, meta := range SpriteRegistry {
		ReverseSpriteRegistry[meta.Image] = id
	}
}
func ConvertToLevelData(level *levels.Level) *LevelData {
	palette := []string{}
	paletteMap := map[string]int{} // string -> index

	getIndex := func(spriteID string) int {
		if idx, ok := paletteMap[spriteID]; ok {
			return idx
		}
		idx := len(palette)
		palette = append(palette, spriteID)
		paletteMap[spriteID] = idx
		return idx
	}

	data := &LevelData{
		Width:         level.W,
		Height:        level.H,
		TileSize:      level.TileSize,
		Tiles:         make([][]TileData, level.H),
		SpritePalette: palette,
	}

	for y := 0; y < level.H; y++ {
		data.Tiles[y] = make([]TileData, level.W)
		for x := 0; x < level.W; x++ {
			t := level.Tiles[y][x]
			indexes := []int{}

			for _, sprite := range t.Sprites {
				index := getIndex(sprite.ID)
				indexes = append(indexes, index)
			}

			data.Tiles[y][x] = TileData{
				SpriteIndexes: indexes,
				IsWalkable:    t.IsWalkable,
			}
		}
	}

	// now that palette was built in getIndex:
	data.SpritePalette = palette
	return data
}

func ConvertToLevel(data *LevelData) *levels.Level {
	level := &levels.Level{
		W:        data.Width,
		H:        data.Height,
		TileSize: data.TileSize,
		Tiles:    make([][]*tiles.Tile, data.Height),
	}

	for y := 0; y < data.Height; y++ {
		level.Tiles[y] = make([]*tiles.Tile, data.Width)
		for x := 0; x < data.Width; x++ {
			td := data.Tiles[y][x]
			t := &tiles.Tile{
				IsWalkable: td.IsWalkable,
			}

			for _, index := range td.SpriteIndexes {
				if index < 0 || index >= len(data.SpritePalette) {
					fmt.Printf("Warning: sprite index %d out of range\n", index)
					continue
				}
				id := data.SpritePalette[index]
				if meta, ok := SpriteRegistry[id]; ok {
					t.AddSpriteByID(id, meta.Image)
				} else {
					fmt.Printf("Warning: sprite ID '%s' not found in registry\n", id)
				}
			}

			level.Tiles[y][x] = t
		}
	}
	return level
}

func SaveLevelToFile(level *levels.Level, path string) error {
	data := ConvertToLevelData(level)
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0644)
}

func LoadLevelFromFile(path string) (*levels.Level, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data LevelData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return ConvertToLevel(&data), nil
}
