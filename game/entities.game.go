package game

import "github.com/hajimehoshi/ebiten/v2"

func (g *Game) drawEntities(target *ebiten.Image, scale, cx, cy float64) {
	tileSize := g.currentLevel.TileSize
	if g.player != nil && !g.player.IsDead {
		g.player.Draw(target, tileSize, func(x, y int) (float64, float64) {
			return g.cartesianToIso(float64(x), float64(y))
		}, g.camX, g.camY, scale, cx, cy)
	}

	for _, m := range g.Monsters {
		m.Draw(target, tileSize, g.camX, g.camY, scale, cx, cy)
	}
	g.drawHitMarkers(target, scale, cx, cy)
	g.drawDamageNumbers(target, scale, cx, cy)
}
