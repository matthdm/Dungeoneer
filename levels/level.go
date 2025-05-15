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
	// Create a 108x108 Level.
	l := &Level{
		W:        64, //108,
		H:        64, //108,
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
				t.AddSprite(ss.Wall)
				t.IsWalkable = false
			case val < 285:
				t.AddSprite(ss.Statue)
				t.IsWalkable = false
			case val < 288:
				t.AddSprite(ss.Crown)
				t.IsWalkable = false
			case val < 289:
				t.AddSprite(ss.Floor)
				t.AddSprite(ss.Tube)
				t.IsWalkable = false
			case val < 290:
				t.AddSprite(ss.Portal)
				t.IsWalkable = true
			case val < 10:
				t.AddSprite(ss.BlueMan)
				t.IsWalkable = false
			case val < 3:
				t.AddSprite(ss.RedMan)
				t.IsWalkable = false
			default:
				t.AddSprite(ss.Floor)
				t.IsWalkable = true
			}
			l.Tiles[y][x] = t
		}
	}

	return l, nil
}

func NewLevel1() *Level {
	return LEVEL_ONE
}

func NewForestLevel() (*Level, error) {
	// Create a 108x108 Level.
	l := &Level{
		W:        64, //108,
		H:        64, //108,
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
				t.AddSprite(ss.OakChunk)
				t.IsWalkable = false
			case val < 285:
				t.AddSprite(ss.OakChunkSlim)
				t.IsWalkable = false
			case val < 288:
				t.AddSprite(ss.Crown)
				t.IsWalkable = false
			case val < 289:
				t.AddSprite(ss.Floor)
				t.AddSprite(ss.Tube)
				t.IsWalkable = false
			case val < 290:
				t.AddSprite(ss.Portal)
				t.IsWalkable = true
			default:
				t.AddSprite(ss.Floor)
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
