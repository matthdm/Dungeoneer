package game

import (
	"dungeoneer/coords"
	"dungeoneer/entities"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) handleHitMarkers() {
	var remaining []entities.HitMarker
	for _, hm := range g.HitMarkers {
		hm.Ticks++
		if hm.Ticks < hm.MaxTicks {
			remaining = append(remaining, hm)
		}
	}
	g.HitMarkers = remaining
}

func (g *Game) drawHitMarkers(target *ebiten.Image, scale, cx, cy float64) {
	ts := g.currentLevel.TileSize
	for _, hm := range g.HitMarkers {
		// TileCenterIso gives the geometric center of the tile diamond, which
		// is the correct anchor for a hit marker regardless of tile size or zoom.
		isoX, isoY := coords.WorldPos{X: hm.X, Y: hm.Y}.TileCenterIso(ts)

		x := float32((isoX-g.camX)*scale) + float32(cx)
		y := float32((isoY+g.camY)*scale) + float32(cy)

		alpha := 1.0 - float64(hm.Ticks)/float64(hm.MaxTicks)
		a := uint8(255 * alpha)
		col := color.NRGBA{255, 0, 0, a}

		size := float32(6) * float32(scale)

		// Draw diagonal /
		vector.StrokeLine(target, x-size, y-size, x+size, y+size, 2, col, true)

		// Draw diagonal \
		vector.StrokeLine(target, x+size, y-size, x-size, y+size, 2, col, true)
	}
}
