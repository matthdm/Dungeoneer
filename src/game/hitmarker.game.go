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
		// hm.X/hm.Y is the struck entity's BodyCenter(); mark its body.
		bx, by := g.bodyToScreen(hm.X, hm.Y, scale, cx, cy)
		x, y := float32(bx), float32(by)

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
