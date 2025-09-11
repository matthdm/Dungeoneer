package render

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"github.com/hajimehoshi/ebiten/v2"
)

const chunkTiles = 32 // 32x32 tiles per chunk

type Chunk struct {
	Img           *ebiten.Image
	Dirty         bool
	X0, Y0        int
	Width, Height int
}

type TileRenderer struct {
	Chunks [][]*Chunk
	TilePx int
	SS     *sprites.SpriteSheet
}

func NewTileRenderer(l *levels.Level, ss *sprites.SpriteSheet, tilePx int) *TileRenderer {
	nx := (l.W + chunkTiles - 1) / chunkTiles
	ny := (l.H + chunkTiles - 1) / chunkTiles
	tr := &TileRenderer{Chunks: make([][]*Chunk, ny), TilePx: tilePx, SS: ss}
	for j := 0; j < ny; j++ {
		tr.Chunks[j] = make([]*Chunk, nx)
		for i := 0; i < nx; i++ {
			w := chunkTiles
			if (i+1)*chunkTiles > l.W {
				w = l.W - i*chunkTiles
			}
			h := chunkTiles
			if (j+1)*chunkTiles > l.H {
				h = l.H - j*chunkTiles
			}
			img := ebiten.NewImage(w*tilePx, h*tilePx)
			tr.Chunks[j][i] = &Chunk{Img: img, Dirty: true, X0: i * chunkTiles, Y0: j * chunkTiles, Width: w, Height: h}
		}
	}
	tr.RebuildAll(l)
	return tr
}

func (tr *TileRenderer) RebuildAll(l *levels.Level) {
	for j := range tr.Chunks {
		for i := range tr.Chunks[j] {
			tr.RebuildChunk(l, tr.Chunks[j][i])
		}
	}
}

func (tr *TileRenderer) RebuildChunk(l *levels.Level, c *Chunk) {
	c.Img.Clear()
	op := &ebiten.DrawImageOptions{}
	for ty := 0; ty < c.Height; ty++ {
		for tx := 0; tx < c.Width; tx++ {
			t := l.Tiles[c.Y0+ty][c.X0+tx]
			op.GeoM.Reset()
			op.GeoM.Translate(float64(tx*tr.TilePx), float64(ty*tr.TilePx))
			c.Img.DrawImage(tr.SS.Floor, op)
			if !t.IsWalkable {
				edge := false
				for dy := -1; dy <= 1 && !edge; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						nx, ny := c.X0+tx+dx, c.Y0+ty+dy
						if nx < 0 || ny < 0 || nx >= l.W || ny >= l.H || l.Tiles[ny][nx].IsWalkable {
							edge = true
							break
						}
					}
				}
				if edge {
					c.Img.DrawImage(tr.SS.DungeonWall, op)
				}
			}
		}
	}
	c.Dirty = false
}

func (tr *TileRenderer) MarkDirty(tx, ty int) {
	j := ty / chunkTiles
	i := tx / chunkTiles
	if j >= 0 && j < len(tr.Chunks) && i >= 0 && i < len(tr.Chunks[j]) {
		tr.Chunks[j][i].Dirty = true
	}
}

func (tr *TileRenderer) Draw(screen *ebiten.Image, l *levels.Level, camX, camY, viewW, viewH int) {
	tileW := tr.TilePx
	x0 := max(0, camX/tileW) - 1
	y0 := max(0, camY/tileW) - 1
	x1 := min(l.W-1, (camX+viewW)/tileW) + 1
	y1 := min(l.H-1, (camY+viewH)/tileW) + 1
	ci0 := x0 / chunkTiles
	cj0 := y0 / chunkTiles
	ci1 := x1 / chunkTiles
	cj1 := y1 / chunkTiles

	for cj := cj0; cj <= cj1; cj++ {
		if cj < 0 || cj >= len(tr.Chunks) {
			continue
		}
		for ci := ci0; ci <= ci1; ci++ {
			if ci < 0 || ci >= len(tr.Chunks[cj]) {
				continue
			}
			c := tr.Chunks[cj][ci]
			if c.Dirty {
				tr.RebuildChunk(l, c)
			}
			op := &ebiten.DrawImageOptions{}
			px := c.X0*tileW - camX
			py := c.Y0*tileW - camY
			op.GeoM.Translate(float64(px), float64(py))
			screen.DrawImage(c.Img, op)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
