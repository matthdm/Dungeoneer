package game

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderable represents any drawable object with tile coordinates for depth sorting.
type Renderable struct {
	Image        *ebiten.Image
	Options      *ebiten.DrawImageOptions
	TileX, TileY float64
	DepthWeight  float64
}

// depth computes the draw depth of the renderable.
func (r Renderable) depth() float64 {
	return r.TileX + r.TileY + r.DepthWeight
}

// sortRenderables orders the slice from back to front.
func sortRenderables(rs []Renderable) {
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].depth() < rs[j].depth()
	})
}
