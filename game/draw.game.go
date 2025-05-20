package game

import "github.com/hajimehoshi/ebiten/v2"

func (g *Game) drawTiles(target *ebiten.Image, scale, cx, cy float64) {
	op := &ebiten.DrawImageOptions{}
	padding := float64(g.currentLevel.TileSize) * scale

	for y := 0; y < g.currentLevel.H; y++ {
		for x := 0; x < g.currentLevel.W; x++ {
			tile := g.currentLevel.Tiles[y][x]
			if tile == nil {
				continue
			}

			xi, yi := g.cartesianToIso(float64(x), float64(y))
			drawX := ((xi - g.camX) * scale) + cx
			drawY := ((yi + g.camY) * scale) + cy

			if drawX+padding < 0 || drawY+padding < 0 || drawX > float64(g.w) || drawY > float64(g.h) {
				continue
			}

			op.GeoM.Reset()
			op.GeoM.Translate(xi, yi)
			op.GeoM.Translate(-g.camX, g.camY)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(cx, cy)

			tile.Draw(target, op)
		}
	}
}

func (g *Game) drawHoverTile(target *ebiten.Image, scale, cx, cy float64) {
	if g.hoverTileX < 0 || g.hoverTileY < 0 ||
		g.hoverTileX >= g.currentLevel.W || g.hoverTileY >= g.currentLevel.H {
		return
	}

	xi, yi := g.cartesianToIso(float64(g.hoverTileX), float64(g.hoverTileY))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(xi, yi)
	op.GeoM.Translate(-g.camX, g.camY)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx, cy)
	// If this is the last tile in the path AND contains a living monster, draw red
	if len(g.player.PathPreview) > 0 {
		finalSpot := g.player.PathPreview[len(g.player.PathPreview)-1]

		for _, m := range g.Monsters {
			if !m.IsDead && m.TileX == finalSpot.X && m.TileY == finalSpot.Y {
				op.ColorScale.Scale(1, 0, 0, 0.8) // red
				break
			}
		}
	}
	target.DrawImage(g.highlightImage, op)
}

func (g *Game) drawPathPreview(target *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil {
		return
	}

	preview := g.player.PathPreview
	for _, step := range preview {
		if step.X < 0 || step.Y < 0 || step.X >= g.currentLevel.W || step.Y >= g.currentLevel.H {
			continue
		}
		xi, yi := g.cartesianToIso(float64(step.X), float64(step.Y))

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(xi, yi)
		op.GeoM.Translate(-g.camX, g.camY)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, cy)

		// Default: translucent white cursor
		op.ColorScale.Scale(1, 1, 1, 0.4)
		img := g.spriteSheet.Cursor

		target.DrawImage(img, op)
	}
}
