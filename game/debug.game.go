package game

import (
	"dungeoneer/tiles"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func notNil(t *tiles.Tile) bool {
	return t != nil
}

func (g *Game) DebugLevelEditor() {
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
		g.editor.SetSelectedSprite(g.spriteSheet.Wall)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorWest)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		t.IsWalkable = false
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorEast)
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
		g.editor.SetSelectedSprite(g.spriteSheet.Wall)
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
		g.editor.SetSelectedSprite(g.spriteSheet.Wall)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorWest)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorEast)
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
		g.editor.SetSelectedSprite(g.spriteSheet.Wall)
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
		g.editor.SetSelectedSprite(g.spriteSheet.Wall)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorWest)
		g.editor.PlaceTile(t)
	case inpututil.IsKeyJustPressed(ebiten.Key4):
		t := g.currentLevel.Tile(g.hoverTileX, g.hoverTileY)
		g.editor.SetSelectedSprite(g.spriteSheet.LockedDoorEast)
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
		g.editor.SetSelectedSprite(g.spriteSheet.Wall)
		g.editor.PlaceTile(t)
	}
}
