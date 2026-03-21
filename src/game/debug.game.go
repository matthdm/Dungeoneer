package game

import (
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"dungeoneer/tiles"
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func notNil(t *tiles.Tile) bool {
	return t != nil
}

func (g *Game) DebugLevelEditor() {
	if g.editor == nil || !g.editor.Active {
		return
	}
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
			newWorld := levels.NewLayeredLevel(level)
			g.currentWorld = newWorld
			g.currentLevel = level
			g.editor = leveleditor.NewLayeredEditor(newWorld, g.w, g.h)
			g.editor.OnLayerChange = g.editorLayerChanged
			g.editor.OnStairPlaced = g.stairPlaced
			g.UpdateSeenTiles(*level)
			g.spawnEntitiesFromLevel()
			fmt.Println("Level loaded from test_level.json")

		}
	}

	// Tile painting
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) &&
		!g.editor.JustSelectedSprite && !g.editor.JustSelectedEntity && !g.editor.IsMenuOpen() {
		if g.editor.EntityMode == leveleditor.ModeDelete {
			// Delete mode - remove entities from tiles
			g.editor.DeleteEntityAt(g.hoverTileX, g.hoverTileY)
			g.spawnEntitiesFromLevel()
		} else if g.editor.SelectedEntityID != "" && g.editor.EntityMode == leveleditor.ModeSpawner {
			// Spawner mode - place entities
			g.editor.PlaceSelectedEntityAt(g.hoverTileX, g.hoverTileY)
			g.spawnEntitiesFromLevel()
		} else if g.editor.SelectedID != "" {
			// Sprite mode - place sprites
			g.editor.PlaceSelectedSpriteAt(g.hoverTileX, g.hoverTileY)
		}
	}

	// Middle click - remove top sprite from tile (works in all modes)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		tx, ty := g.hoverTileX, g.hoverTileY
		if g.isValidTile(tx, ty) {
			tile := g.currentLevel.Tile(tx, ty)
			if tile != nil && len(tile.Sprites) > 1 {
				removed := tile.Sprites[len(tile.Sprites)-1]
				tile.Sprites = tile.Sprites[:len(tile.Sprites)-1] // remove top
				// Optionally: update walkable state
				last := tile.Sprites[len(tile.Sprites)-1]
				meta := leveleditor.SpriteRegistry[last.ID]
				tile.IsWalkable = meta.IsWalkable
				if tile.HasTag(tiles.TagDoor) && (strings.Contains(strings.ToLower(removed.ID), "door_locked") ||
					strings.Contains(strings.ToLower(removed.ID), "door_unlocked")) {
					// If no door sprites remain, clear door state.
					hasDoor := false
					for _, s := range tile.Sprites {
						if strings.Contains(strings.ToLower(s.ID), "door_locked") ||
							strings.Contains(strings.ToLower(s.ID), "door_unlocked") {
							hasDoor = true
							break
						}
					}
					if !hasDoor {
						tile.Tags &^= tiles.TagDoor
						tile.DoorState = 0
						tile.DoorSpriteID = ""
					}
				}
			}
		}
	}
}
