package game

import (
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
	for _, hm := range g.HitMarkers {
		xi, yi := g.cartesianToIso(hm.X, hm.Y)

		x := float32(xi-g.camX+35)*float32(scale) + float32(cx) // tweak +4 as needed
		y := float32(yi+g.camY+15)*float32(scale) + float32(cy) // tweak -8 until centered

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
