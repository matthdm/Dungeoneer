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

// RoomSize classifies a room by area.
type RoomSize string

const (
	RoomSmall  RoomSize = "small"  // area < 49
	RoomMedium RoomSize = "medium" // area 49–99
	RoomLarge  RoomSize = "large"  // area >= 100
)

// RoomTag represents a semantic role or modifier assigned to a room.
type RoomTag string

const (
	// Primary roles (mutually exclusive).
	TagSpawn      RoomTag = "spawn"
	TagExit       RoomTag = "exit"
	TagBossArena  RoomTag = "boss_arena"
	TagSanctuary  RoomTag = "sanctuary"
	TagTreasure   RoomTag = "treasure"
	TagGuardPost  RoomTag = "guard_post"
	TagBarracks   RoomTag = "barracks"
	TagAmbush     RoomTag = "ambush"
	TagCrossroads RoomTag = "crossroads"
	TagDeadEnd    RoomTag = "dead_end"
	TagCommon     RoomTag = "common"

	// Modifier tags (stackable).
	TagDecorated RoomTag = "decorated"
	TagLoot      RoomTag = "loot"
	TagCleared   RoomTag = "cleared"
	TagDark      RoomTag = "dark"
	TagOptional  RoomTag = "optional"
)

// Room holds metadata for a single room in a generated level.
type Room struct {
	X, Y, W, H       int
	CenterX, CenterY int
	Size              RoomSize
	Index             int
	Tags              []RoomTag
}

// Contains reports whether tile (tx, ty) falls inside the room.
func (r *Room) Contains(tx, ty int) bool {
	return tx >= r.X && tx < r.X+r.W && ty >= r.Y && ty < r.Y+r.H
}

// HasTag returns true if the room has the given tag.
func (r *Room) HasTag(tag RoomTag) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag appends a tag if not already present.
func (r *Room) AddTag(tag RoomTag) {
	if !r.HasTag(tag) {
		r.Tags = append(r.Tags, tag)
	}
}

// PrimaryTag returns the first non-modifier tag, or TagCommon if none found.
func (r *Room) PrimaryTag() RoomTag {
	for _, t := range r.Tags {
		switch t {
		case TagDecorated, TagLoot, TagCleared, TagDark, TagOptional:
			continue
		default:
			return t
		}
	}
	return TagCommon
}

// RoomsByTag returns pointers to all rooms with the given tag.
func RoomsByTag(rooms []Room, tag RoomTag) []*Room {
	var out []*Room
	for i := range rooms {
		if rooms[i].HasTag(tag) {
			out = append(out, &rooms[i])
		}
	}
	return out
}

// ClassifyRoomSize returns a size category based on room area.
func ClassifyRoomSize(w, h int) RoomSize {
	area := w * h
	if area < 49 {
		return RoomSmall
	}
	if area < 100 {
		return RoomMedium
	}
	return RoomLarge
}

// Level represents a Game level.
type Level struct {
	W, H int

	Tiles       [][]*tiles.Tile // (Y,X) array of tiles
	TileSize    int
	Entities    []PlacedEntity
	DoorDensity DoorDensityConfig
	Rooms       []Room
}

// RoomAt returns the room containing tile (x, y), or nil if not in any room.
func (l *Level) RoomAt(x, y int) *Room {
	for i := range l.Rooms {
		if l.Rooms[i].Contains(x, y) {
			return &l.Rooms[i]
		}
	}
	return nil
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

// IsPassable returns true if a tile can be traversed, treating closed and
// locked doors as passable. Used by spawn/exit placement so the BFS can
// path through doors and place objectives behind them.
func (l Level) IsPassable(x, y int) bool {
	if x < 0 || y < 0 || x >= l.W || y >= l.H {
		return false
	}
	t := l.Tiles[y][x]
	if t == nil {
		return false
	}
	if t.HasTag(tiles.TagDoor) {
		return true // doors are passable for placement purposes
	}
	return t.IsWalkable
}

func blankLevelCenterFloor(currentFloorID string, currentFloor *ebiten.Image) (string, *ebiten.Image) {
	centerFlavor := "gallery"
	if currentFloorID == "gallery_floor" {
		centerFlavor = "crypt"
	}

	wss, err := sprites.LoadWallSpriteSheet(centerFlavor)
	if err != nil || wss == nil || wss.Floor == nil {
		return currentFloorID, currentFloor
	}

	return centerFlavor + "_floor", wss.Floor
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
	centerFloorID, centerFloor := blankLevelCenterFloor("Floor", ss.Floor)

	for y := 0; y < height; y++ {
		row := make([]*tiles.Tile, width)
		for x := 0; x < width; x++ {
			t := &tiles.Tile{
				IsWalkable: true,
			}
			if x == width/2 && y == height/2 {
				t.AddSpriteByID(centerFloorID, centerFloor)
				t.AddSpriteByID("Glyph", ss.Glyph)
			} else {
				t.AddSpriteByID("Floor", ss.Floor)
			}
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
	centerFloorID, centerFloor := blankLevelCenterFloor(floorID, img)
	for y := 0; y < height; y++ {
		row := make([]*tiles.Tile, width)
		for x := 0; x < width; x++ {
			t := &tiles.Tile{IsWalkable: true}
			if x == width/2 && y == height/2 {
				t.AddSpriteByID(centerFloorID, centerFloor)
			} else {
				t.AddSpriteByID(floorID, img)
			}
			row[x] = t
		}
		l.Tiles[y] = row
	}
	return l
}
