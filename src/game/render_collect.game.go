package game

import (
	"dungeoneer/entities"
	"dungeoneer/fov"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var itemSparkle *ebiten.Image

func init() {
	itemSparkle = ebiten.NewImage(5, 5)
	white := color.NRGBA{255, 255, 255, 255}
	itemSparkle.Set(2, 0, white)
	itemSparkle.Set(2, 1, white)
	itemSparkle.Set(2, 2, white)
	itemSparkle.Set(2, 3, white)
	itemSparkle.Set(2, 4, white)
	itemSparkle.Set(0, 2, white)
	itemSparkle.Set(1, 2, white)
	itemSparkle.Set(3, 2, white)
	itemSparkle.Set(4, 2, white)
}

// collectRenderables gathers renderables for the current frame.
func (g *Game) collectRenderables(scale, cx, cy float64) []Renderable {
	var r []Renderable
	r = append(r, g.collectTileRenderables(scale, cx, cy)...)
	r = append(r, g.collectChestRenderables(scale, cx, cy)...)
	r = append(r, g.collectItemDropRenderables(scale, cx, cy)...)
	r = append(r, g.collectExitEntityRenderables(scale, cx, cy)...)

	r = append(r, g.collectMonsterRenderables(scale, cx, cy)...)
	r = append(r, g.collectNPCRenderables(scale, cx, cy)...)
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
					op.ColorScale.Scale(0.2, 0.2, 0.2, 1.0)
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

func (g *Game) collectItemDropRenderables(scale, cx, cy float64) []Renderable {
	var out []Renderable
	for _, d := range g.ItemDrops {
		if d.TileX < 0 || d.TileY < 0 || d.TileX >= g.currentLevel.W || d.TileY >= g.currentLevel.H {
			continue
		}
		inFOV := g.isTileVisible(d.TileX, d.TileY)
		wasSeen := g.SeenTiles[d.TileY][d.TileX]
		if !inFOV && !wasSeen {
			continue
		}
		xi, yi := g.cartesianToIso(float64(d.TileX), float64(d.TileY))
		b := d.Item.Icon.Bounds()
		w := float64(b.Dx())
		h := float64(b.Dy())
		ts := float64(g.currentLevel.TileSize)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(0.5, 0.5)
		op.GeoM.Translate(-w/4, ts/4-h/4)
		op.GeoM.Translate(xi, yi)
		op.GeoM.Translate(-g.camX, g.camY)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, cy)
		phase := 0.8 + 0.2*math.Sin(float64(time.Now().UnixNano()%1e9)/1e9*2*math.Pi)
		if inFOV {
			p := float32(phase)
			op.ColorScale.Scale(p, p, p, 1)
		} else if wasSeen {
			op.ColorScale.Scale(0.2, 0.2, 0.2, 1)
		}
		out = append(out, Renderable{
			Image:       d.Item.Icon,
			Options:     op,
			TileX:       float64(d.TileX),
			TileY:       float64(d.TileY),
			DepthWeight: 0.3,
		})
		if inFOV {
			spark := &ebiten.DrawImageOptions{}
			spark.GeoM.Translate(-2, ts/4-h/4-4)
			spark.GeoM.Translate(xi, yi)
			spark.GeoM.Translate(-g.camX, g.camY)
			spark.GeoM.Scale(scale, scale)
			spark.GeoM.Translate(cx, cy)
			spark.ColorScale.Scale(1, 1, 1, float32(0.6+0.4*phase))
			out = append(out, Renderable{
				Image:       itemSparkle,
				Options:     spark,
				TileX:       float64(d.TileX),
				TileY:       float64(d.TileY),
				DepthWeight: 0.31,
			})
		}
	}
	return out
}

func (g *Game) collectExitEntityRenderables(scale, cx, cy float64) []Renderable {
	if g.ExitEntity == nil || g.ExitEntity.Sprite == nil {
		return nil
	}
	e := g.ExitEntity
	if e.TileX < 0 || e.TileY < 0 || e.TileX >= g.currentLevel.W || e.TileY >= g.currentLevel.H {
		return nil
	}
	inFOV := g.isTileVisible(e.TileX, e.TileY)
	wasSeen := g.SeenTiles[e.TileY][e.TileX]
	if !inFOV && !wasSeen {
		return nil
	}
	xi, yi := g.cartesianToIso(float64(e.TileX), float64(e.TileY))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, e.BobOffset)
	op.GeoM.Translate(xi, yi)
	op.GeoM.Translate(-g.camX, g.camY)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx, cy)
	if inFOV {
		p := e.PulseAlpha()
		op.ColorScale.Scale(p, p, p, 1)
	} else {
		op.ColorScale.Scale(0.2, 0.2, 0.2, 1)
	}
	return []Renderable{
		{
			Image:       e.Sprite,
			Options:     op,
			TileX:       float64(e.TileX),
			TileY:       float64(e.TileY),
			DepthWeight: 0.3,
		},
	}
}

func (g *Game) collectChestRenderables(scale, cx, cy float64) []Renderable {
	var out []Renderable
	for _, c := range g.Chests {
		if c.Sprite == nil {
			continue
		}
		if c.TileX < 0 || c.TileY < 0 || c.TileX >= g.currentLevel.W || c.TileY >= g.currentLevel.H {
			continue
		}
		inFOV := g.isTileVisible(c.TileX, c.TileY)
		wasSeen := g.SeenTiles[c.TileY][c.TileX]
		if !inFOV && !wasSeen {
			continue
		}
		xi, yi := g.cartesianToIso(float64(c.TileX), float64(c.TileY))
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(xi, yi)
		op.GeoM.Translate(-g.camX, g.camY)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, cy)

		// Tint by variant when in FOV; dim when only seen.
		if inFOV {
			r, gv, b := chestTint(c.Variant, c.Opened)
			op.ColorScale.Scale(r, gv, b, 1)
		} else {
			op.ColorScale.Scale(0.2, 0.2, 0.2, 1)
		}
		out = append(out, Renderable{
			Image:       c.Sprite,
			Options:     op,
			TileX:       float64(c.TileX),
			TileY:       float64(c.TileY),
			DepthWeight: 0.3,
		})
	}
	return out
}

// chestTint returns RGB multipliers for a chest variant. Opened chests are dimmed.
func chestTint(variant string, opened bool) (float32, float32, float32) {
	if opened {
		return 0.45, 0.35, 0.25
	}
	switch variant {
	case entities.ChestGold:
		return 1.0, 0.85, 0.3
	case entities.ChestIron:
		return 0.75, 0.85, 1.0
	case entities.ChestLocked:
		return 0.9, 0.4, 0.9
	default: // wooden
		return 1.0, 0.75, 0.5
	}
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
		if !g.isTileVisible(m.TileX, m.TileY) {
			continue // monsters only render when actively in FOV, not fog-of-war
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

func (g *Game) collectNPCRenderables(scale, cx, cy float64) []Renderable {
	var out []Renderable
	for _, n := range g.NPCs {
		if n.Sprite == nil {
			continue
		}
		if n.TileX < 0 || n.TileY < 0 || n.TileX >= g.currentLevel.W || n.TileY >= g.currentLevel.H {
			continue
		}
		if !g.isTileVisible(n.TileX, n.TileY) {
			continue
		}
		x, y := g.cartesianToIso(n.InterpX, n.InterpY)
		op := &ebiten.DrawImageOptions{}
		const verticalOffset = 1.0
		op.GeoM.Translate(0, -verticalOffset+n.BobOffset)
		if !n.LeftFacing {
			op.GeoM.Scale(-1, 1)
			w := float64(n.Sprite.Bounds().Dx())
			op.GeoM.Translate(w, 0)
		}
		op.GeoM.Translate(x, y)
		op.GeoM.Translate(-g.camX, g.camY)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, cy)
		out = append(out, Renderable{
			Image:       n.Sprite,
			Options:     op,
			TileX:       n.InterpX,
			TileY:       n.InterpY,
			DepthWeight: 0.5,
		})
	}
	return out
}

func (g *Game) collectShadowRenderables(scale, cx, cy float64) []Renderable {
	// Place the shadow apex at the player's head — sprite pixel (tileSize/2, tileSize/4)
	// from the sprite's iso anchor, which aligns with the character's visual position.
	mc := g.player.MoveController
	isoX, isoY := g.cartesianToIso(mc.InterpX, mc.InterpY)
	ts := float64(g.currentLevel.TileSize)
	apexX := (isoX + ts/2 - g.camX) * scale + cx
	apexY := (isoY + ts/4 + g.camY) * scale + cy

	img := fov.BuildShadowImage(g.cachedRays, apexX, apexY, g.camX, g.camY, scale, cx, cy, g.currentLevel.TileSize)
	op := &ebiten.DrawImageOptions{}
	tx := mc.InterpX
	ty := mc.InterpY
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
