package leveleditor

import (
	"dungeoneer/levels"
	"dungeoneer/tiles"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// TileData is a serializable representation of a tile.
type TileData struct {
	Sprites []string `json:"sprites"` // sprite keys like "Floor", "Wall"
}

// LevelJSON is a serialized version of a level.
type LevelJSON struct {
	Width    int          `json:"width"`
	Height   int          `json:"height"`
	TileSize int          `json:"tileSize"`
	Tiles    [][]TileData `json:"tiles"` // 2D array [y][x]
}

func SaveToFile(level *LevelJSON, path string) error {
	data, err := json.MarshalIndent(level, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadFromFile(path string, spriteMap map[string]*ebiten.Image) (*levels.Level, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var levelData LevelJSON
	if err := json.Unmarshal(data, &levelData); err != nil {
		return nil, err
	}
	// Convert to levels.Level
	l := &levels.Level{
		W:        levelData.Width,
		H:        levelData.Height,
		TileSize: levelData.TileSize,
		Tiles:    make([][]*tiles.Tile, levelData.Height),
	}

	for y := 0; y < levelData.Height; y++ {
		l.Tiles[y] = make([]*tiles.Tile, levelData.Width)
		for x := 0; x < levelData.Width; x++ {
			tile := &tiles.Tile{}
			for _, spriteKey := range levelData.Tiles[y][x].Sprites {
				if img, ok := spriteMap[spriteKey]; ok {
					tile.AddSprite(img)
				}
			}
			l.Tiles[y][x] = tile
		}
	}

	return l, nil
}

func FromLevel(level *levels.Level, reverseMap map[*ebiten.Image]string) *LevelJSON {
	out := &LevelJSON{
		Width:    level.W,
		Height:   level.H,
		TileSize: level.TileSize,
		Tiles:    make([][]TileData, level.H),
	}
	for y := 0; y < level.H; y++ {
		out.Tiles[y] = make([]TileData, level.W)
		for x := 0; x < level.W; x++ {
			tile := level.Tiles[y][x]
			if tile == nil {
				continue
			}
			sdata := TileData{}
			for _, sprite := range tile.Sprites {
				if key, ok := reverseMap[sprite]; ok {
					sdata.Sprites = append(sdata.Sprites, key)
				} else {
					fmt.Println("Unknown sprite", sprite)
				}
			}
			out.Tiles[y][x] = sdata
		}
	}
	return out
}
