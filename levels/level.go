package levels

import (
	"dungeoneer/constants"
	"dungeoneer/sprites"
	"dungeoneer/tiles"

	"fmt"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// PlacedEntity represents an entity placed on the map via the editor.
// It stores the tile coordinates, the entity type, and the sprite to use.
type PlacedEntity struct {
	X, Y     int
	Type     string
	SpriteID string
}

// Level represents a Game level.
type Level struct {
	W, H int

	Tiles    [][]*tiles.Tile // (Y,X) array of tiles
	TileSize int
	Entities []PlacedEntity
	DoorDensity DoorDensityConfig
}

// NewEmptyLevel creates a Level filled entirely with non-walkable tiles.
func NewEmptyLevel(w, h int) *Level {
	l := &Level{
		W:        w,
		H:        h,
		TileSize: constants.DefaultTileSize,
		Tiles:    make([][]*tiles.Tile, h),
		Entities: []PlacedEntity{},
		DoorDensity: DefaultDoorDensityConfig(),
	}
	for y := 0; y < h; y++ {
		row := make([]*tiles.Tile, w)
		for x := 0; x < w; x++ {
			row[x] = &tiles.Tile{IsWalkable: false}
		}
		l.Tiles[y] = row
	}
	return l
}

// AddEntity appends an entity to the level.
func (l *Level) AddEntity(e PlacedEntity) {
	l.Entities = append(l.Entities, e)
}

// RemoveEntityAt removes entities matching the given tile, type, and optional ID.
func (l *Level) RemoveEntityAt(x, y int, typ, id string) {
	for i := len(l.Entities) - 1; i >= 0; i-- {
		e := l.Entities[i]
		if e.X == x && e.Y == y && e.Type == typ && (id == "" || e.SpriteID == id) {
			l.Entities = append(l.Entities[:i], l.Entities[i+1:]...)
		}
	}
}

// Tile returns the tile at the provided coordinates, or nil.
func (l *Level) Tile(x, y int) *tiles.Tile {
	if x >= 0 && y >= 0 && x < l.W && y < l.H {
		return l.Tiles[y][x]
	}
	return nil
}

// Size returns the size of the Level.
func (l *Level) Size() (width, height int) {
	return l.W, l.H
}

// NewLevel returns a new randomly generated Level.
func NewDungeonLevel() (*Level, error) {
	// Create a 64x64 Level.
	l := &Level{
		W:        64, //64,
		H:        64, //64,
		TileSize: 64, //64,
		Entities: []PlacedEntity{},
		DoorDensity: DefaultDoorDensityConfig(),
	}

	// Load embedded SpriteSheet.
	ss, err := sprites.LoadSpriteSheet(l.TileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded spritesheet: %s", err)
	}

	// Fill each tile with one or more sprites randomly.
	l.Tiles = make([][]*tiles.Tile, l.H)
	for y := 0; y < l.H; y++ {
		l.Tiles[y] = make([]*tiles.Tile, l.W)
		for x := 0; x < l.W; x++ {
			t := &tiles.Tile{}
			isBorderSpace := x == 0 || y == 0 || x == l.W-1 || y == l.H-1
			val := rand.IntN(1000)
			switch {
			case isBorderSpace || val < 275:
				t.AddSpriteByID("DungeonWall", ss.DungeonWall)
				t.IsWalkable = false
			case val < 285:
				t.AddSpriteByID("Statue", ss.Statue)
				t.IsWalkable = false
			case val < 288:
				t.AddSpriteByID("Trinket", ss.Trinket)
				t.IsWalkable = false
			case val < 289:
				t.AddSpriteByID("Floor", ss.Floor)
				t.AddSpriteByID("Well", ss.Well)
				t.IsWalkable = false
			case val < 290:
				t.AddSpriteByID("Portal", ss.Portal)
				t.IsWalkable = true
			case val < 10:
				t.AddSpriteByID("DragonStatue", ss.DragonStatue)
				t.IsWalkable = false
			case val < 3:
				t.AddSpriteByID("Campfire", ss.Campfire)
				t.IsWalkable = false
			default:
				t.AddSpriteByID("Floor", ss.Floor)
				t.IsWalkable = true
			}
			l.Tiles[y][x] = t
		}
	}

	return l, nil
}

func NewMazeLevel() (*Level, error) {
	// Load embedded SpriteSheet.
	ss, err := sprites.LoadSpriteSheet(constants.DefaultTileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded spritesheet: %s", err)
	}
	cfg := MazeConfig{
		Width:        64,
		Height:       64,
		Tessellation: "ortho",
		Routing:      "prim",
		Texture:      "elitism",
		WallFlavor:   "stone",
	}
	lvl := GenerateMaze(cfg, ss)
	return lvl, nil
}

func NewForestLevel() (*Level, error) {
	// Create a 64x64 Level.
	l := &Level{
		W:        64, //64,
		H:        64, //64,
		TileSize: 64, //64,
		Entities: []PlacedEntity{},
		DoorDensity: DefaultDoorDensityConfig(),
	}

	// Load embedded SpriteSheet.
	ss, err := sprites.LoadSpriteSheet(l.TileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded spritesheet: %s", err)
	}

	// Fill each tile with one or more sprites randomly.
	l.Tiles = make([][]*tiles.Tile, l.H)
	for y := 0; y < l.H; y++ {
		l.Tiles[y] = make([]*tiles.Tile, l.W)
		for x := 0; x < l.W; x++ {
			t := &tiles.Tile{}
			isBorderSpace := x == 0 || y == 0 || x == l.W-1 || y == l.H-1
			val := rand.IntN(1000)
			switch {
			case isBorderSpace || val < 275:
				t.AddSpriteByID("OakChunk", ss.OakChunk)
				t.IsWalkable = false
			case val < 285:
				t.AddSpriteByID("OakLog", ss.OakLog)
				t.IsWalkable = false
			case val < 288:
				t.AddSpriteByID("Trinket", ss.Trinket)
				t.IsWalkable = false
			case val < 289:
				t.AddSpriteByID("Floor", ss.Floor)
				t.AddSpriteByID("Well", ss.Well)
				t.IsWalkable = false
			case val < 290:
				t.AddSpriteByID("Portal", ss.Portal)
				t.IsWalkable = true
			default:
				t.AddSpriteByID("Floor", ss.Floor)
				t.IsWalkable = true
			}
			l.Tiles[y][x] = t
		}
	}

	return l, nil
}

func (l Level) IsWalkable(x, y int) bool {
	if x < 0 || y < 0 || x >= l.W || y >= l.H {
		return false
	}
	t := l.Tiles[y][x]
	if t == nil {
		return false
	}
	// Closed/locked doors block movement even if tile is walkable
	if t.HasTag(tiles.TagDoor) && (t.DoorState == 2 || t.DoorState == 3) {
		return false
	}
	return t.IsWalkable
}

// CreateNewBlankLevel creates a new level filled with floor tiles
func CreateNewBlankLevel(width, height, tileSize int, ss *sprites.SpriteSheet) *Level {
	l := &Level{
		W:        width,
		H:        height,
		TileSize: tileSize,
		Tiles:    make([][]*tiles.Tile, height),
		Entities: []PlacedEntity{},
		DoorDensity: DefaultDoorDensityConfig(),
	}

	for y := 0; y < height; y++ {
		row := make([]*tiles.Tile, width)
		for x := 0; x < width; x++ {
			t := &tiles.Tile{
				IsWalkable: true,
			}
			t.AddSpriteByID("Floor", ss.Floor) // You can assign a real image from the sprite sheet later
			row[x] = t
		}
		l.Tiles[y] = row
	}

	return l
}

// CreateNewBlankLevelWithFloor creates a level using the provided floor sprite.
func CreateNewBlankLevelWithFloor(width, height, tileSize int, floorID string, img *ebiten.Image) *Level {
	l := &Level{
		W:        width,
		H:        height,
		TileSize: tileSize,
		Tiles:    make([][]*tiles.Tile, height),
		Entities: []PlacedEntity{},
		DoorDensity: DefaultDoorDensityConfig(),
	}
	for y := 0; y < height; y++ {
		row := make([]*tiles.Tile, width)
		for x := 0; x < width; x++ {
			t := &tiles.Tile{IsWalkable: true}
			t.AddSpriteByID(floorID, img)
			row[x] = t
		}
		l.Tiles[y] = row
	}
	return l
}
