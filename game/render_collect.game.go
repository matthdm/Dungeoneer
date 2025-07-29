package game

import (
	"dungeoneer/fov"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// collectRenderables gathers renderables for the current frame.
func (g *Game) collectRenderables(scale, cx, cy float64) []Renderable {
	var r []Renderable
	r = append(r, g.collectTileRenderables(scale, cx, cy)...)
	if g.player != nil && len(g.cachedRays) > 0 && !g.FullBright {
		r = append(r, g.collectShadowRenderables(scale, cx, cy)...)
	}
	r = append(r, g.collectMonsterRenderables(scale, cx, cy)...)
	if g.player != nil {
		r = append(r, g.collectPlayerRenderables(scale, cx, cy)...)
	}
	sortRenderables(r)
	return r
}

func (g *Game) collectTileRenderables(scale, cx, cy float64) []Renderable {
	var out []Renderable
	padding := float64(g.currentLevel.TileSize) * scale
	screenLeft := -padding
	screenTop := -padding
	screenRight := float64(g.w)
	screenBottom := float64(g.h)
	for y := 0; y < g.currentLevel.H; y++ {
		for x := 0; x < g.currentLevel.W; x++ {
			tile := g.currentLevel.Tiles[y][x]
			if tile == nil {
				continue
			}
			xi, yi := g.cartesianToIso(float64(x), float64(y))
			drawX := ((xi - g.camX) * scale) + cx
			drawY := ((yi + g.camY) * scale) + cy
			if drawX < screenLeft || drawY < screenTop || drawX > screenRight || drawY > screenBottom {
				continue
			}
			inFOV := g.isTileVisible(x, y)
			wasSeen := g.SeenTiles[y][x]
			if !inFOV && !wasSeen {
				continue
			}
			for _, s := range tile.Sprites {
				if isFloorSprite(strings.ToLower(s.ID)) {
					continue
				}
				op := g.getDrawOp(xi, yi, scale, cx, cy)
				if !inFOV && wasSeen {
					op.ColorScale.Scale(0.4, 0.4, 0.4, 1.0)
				}
				weight := depthWeightForSprite(s.ID)
				out = append(out, Renderable{
					Image:       s.Image,
					Options:     op,
					TileX:       float64(x),
					TileY:       float64(y),
					DepthWeight: weight,
				})
			}
		}
	}
	return out
}

func (g *Game) collectPlayerRenderables(scale, cx, cy float64) []Renderable {
	if g.player == nil || g.player.IsDead {
		return nil
	}
	sx, sy := g.cartesianToIso(g.player.MoveController.InterpX, g.player.MoveController.InterpY)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, -1.0+g.player.BobOffset)
	if g.player.IsDashing {
		op.ColorScale.Scale(1.3, 1.3, 1.3, 1)
	}
	b := g.player.Sprite.Bounds()
	if !g.player.LeftFacing {
		w := float64(b.Dx())
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(w, 0)
	}
	op.GeoM.Translate(sx, sy)
	op.GeoM.Translate(-g.camX, g.camY)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx, cy)
	return []Renderable{
		{
			Image:       g.player.Sprite,
			Options:     op,
			TileX:       g.player.MoveController.InterpX,
			TileY:       g.player.MoveController.InterpY,
			DepthWeight: 0.5,
		},
	}
}

func (g *Game) collectMonsterRenderables(scale, cx, cy float64) []Renderable {
	var out []Renderable
	for _, m := range g.Monsters {
		if m.IsDead || m.TileX < 0 || m.TileY < 0 || m.TileX >= g.currentLevel.W || m.TileY >= g.currentLevel.H {
			continue
		}
		if !g.isTileVisible(m.TileX, m.TileY) && !g.SeenTiles[m.TileY][m.TileX] {
			continue
		}
		x, y := g.cartesianToIso(m.InterpX, m.InterpY)
		op := &ebiten.DrawImageOptions{}
		const verticalOffset = 1.0
		op.GeoM.Translate(0, -verticalOffset+m.BobOffset)
		if !m.LeftFacing {
			op.GeoM.Scale(-1, 1)
			w := float64(m.Sprite.Bounds().Dx())
			op.GeoM.Translate(w, 0)
		}
		op.GeoM.Translate(x, y)
		op.GeoM.Translate(-g.camX, g.camY)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, cy)
		if m.FlashTicksLeft > 0 {
			op.ColorScale.Scale(1, 1, 1, 0.7)
		}
		out = append(out, Renderable{
			Image:       m.Sprite,
			Options:     op,
			TileX:       m.InterpX,
			TileY:       m.InterpY,
			DepthWeight: 0.5,
		})
	}
	return out
}

func (g *Game) collectShadowRenderables(scale, cx, cy float64) []Renderable {
	img := fov.BuildShadowImage(g.cachedRays, g.camX, g.camY, scale, cx, cy, g.currentLevel.TileSize)
	op := &ebiten.DrawImageOptions{}
	tx := g.player.MoveController.InterpX
	ty := g.player.MoveController.InterpY
	return []Renderable{
		{
			Image:       img,
			Options:     op,
			TileX:       tx,
			TileY:       ty,
			DepthWeight: 0.25,
		},
	}
}
