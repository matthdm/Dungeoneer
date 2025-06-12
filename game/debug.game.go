package game

import (
	"dungeoneer/leveleditor"
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
			fmt.Println("Level loaded from test_level.json")
		}
	}
	switch {
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		if notNil(t) {
			t.RemoveSprite(g.editor.GetSelectedSprite())
		}
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = true
		g.editor.SetSelectedSprite(g.spriteSheet.Floor)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.DungeonWall)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorNW)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorNE)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key5):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.Water)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key6):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.EnchantedWater)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key7):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.EnemyCursor)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key8):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = true
		g.editor.SetSelectedSprite(g.spriteSheet.Portal)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key9):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.Trinket)
		g.editor.PlaceTile(t)
	}
}

func (g *Game) DungeonLevelEditor() {
	switch {
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		if notNil(t) {
			t.RemoveSprite(g.editor.GetSelectedSprite())
		}
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.Floor)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.DungeonWall)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorNW)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorNE)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key5):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.Water)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key6):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.EnchantedWater)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key7):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.EnemyCursor)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key8):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.Portal)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key9):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.DungeonWall)
		g.editor.PlaceTile(t)
	}
}

func (g *Game) ForestLevelEditor() {
	switch {
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		if notNil(t) {
			t.RemoveSprite(g.editor.GetSelectedSprite())
		}
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.Floor)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.OakBeam)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.OakBeamNESW)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.OakLogNWSE)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key5):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.Campfire)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key6):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.EnchantedWater)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key7):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.EnemyCursor)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key8):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.Portal)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key9):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.OakWall)
		g.editor.PlaceTile(t)
	}
}
