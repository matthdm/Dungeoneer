package game

import "github.com/hajimehoshi/ebiten/v2"

func (g *Game) drawPlayer(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || g.player.IsDead {
		return
	}

	tileSize := g.currentLevel.TileSize
	g.player.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
}

func (g *Game) drawMonsters(target *ebiten.Image, scale, cx, cy float64) {
	tileSize := g.currentLevel.TileSize
	for _, m := range g.Monsters {
		if g.isTileVisible(m.TileX, m.TileY) {
			m.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
		} else if g.SeenTiles[m.TileY][m.TileX] {
			// optionally show faded sprite or placeholder
		} else {
			continue
		}

	}
}
