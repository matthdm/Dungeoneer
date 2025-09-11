package game

import (
	"dungeoneer/constants"
	"dungeoneer/leveleditor"
	"dungeoneer/render"
	"dungeoneer/tiles"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func notNil(t *tiles.Tile) bool {
	return t != nil
}

func (g *Game) DebugLevelEditor() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		err := leveleditor.SaveLevelToFile(g.currentLevel, "test_level.json")
		if err != nil {
			fmt.Println("Save failed:", err)
		} else {
			fmt.Println("Level saved to test_level.json")
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
		level, err := leveleditor.LoadLevelFromFile("test_level.json")
		if err != nil {
			fmt.Println("Load failed:", err)
		} else {
			g.currentLevel = level
			g.UpdateSeenTiles(*level)
			g.spawnEntitiesFromLevel()
			g.tileRenderer = render.NewTileRenderer(level, g.spriteSheet, constants.DefaultTileSize)
			fmt.Println("Level loaded from test_level.json")

		}
	}

	// Tile painting
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) &&
		!g.editor.JustSelectedSprite && !g.editor.JustSelectedEntity {
		if g.editor.SelectedEntityID != "" {
			g.editor.PlaceSelectedEntityAt(g.hoverTileX, g.hoverTileY)
			g.spawnEntitiesFromLevel()
		} else {
			g.editor.PlaceSelectedSpriteAt(g.hoverTileX, g.hoverTileY)
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		tx, ty := g.hoverTileX, g.hoverTileY
		if g.isValidTile(tx, ty) {
			tile := g.currentLevel.Tile(tx, ty)
			if tile != nil && len(tile.Sprites) > 1 {
				tile.Sprites = tile.Sprites[:len(tile.Sprites)-1] // remove top
				// Optionally: update walkable state
				last := tile.Sprites[len(tile.Sprites)-1]
				meta := leveleditor.SpriteRegistry[last.ID]
				tile.IsWalkable = meta.IsWalkable
			}
		}
	}
}
