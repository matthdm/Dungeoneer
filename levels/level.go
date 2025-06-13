package levels

import (
	"dungeoneer/sprites"
	"dungeoneer/tiles"

	"fmt"
	"math/rand/v2"
)

// Level represents a Game level.
type Level struct {
	W, H int

	Tiles    [][]*tiles.Tile // (Y,X) array of tiles
	TileSize int
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

func NewForestLevel() (*Level, error) {
	// Create a 64x64 Level.
	l := &Level{
		W:        64, //64,
		H:        64, //64,
		TileSize: 64, //64,
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
	return t.IsWalkable
}

// CreateNewBlankLevel creates a new level filled with floor tiles
func CreateNewBlankLevel(width, height, tileSize int, ss *sprites.SpriteSheet) *Level {
	l := &Level{
		W:        width,
		H:        height,
		TileSize: tileSize,
		Tiles:    make([][]*tiles.Tile, height),
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
