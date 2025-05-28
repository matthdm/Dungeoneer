package game

import "github.com/hajimehoshi/ebiten/v2"

func (g *Game) drawPlayer(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || g.player.IsDead {
		return
	}

	tileSize := g.currentLevel.TileSize
	g.player.Draw(target, tileSize, func(x, y int) (float64, float64) {
		return g.cartesianToIso(float64(x), float64(y))
	}, g.camX, g.camY, scale, cx, cy)
}

func (g *Game) drawMonsters(target *ebiten.Image, scale, cx, cy float64) {
	tileSize := g.currentLevel.TileSize
	for _, m := range g.Monsters {
		m.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
	}
}
