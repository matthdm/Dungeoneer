package entities

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	TileX, TileY int
	LeftFacing   bool
	Sprite       *ebiten.Image
	Path         []PathNode
	TickCount    int
}

type PathNode struct {
	X, Y int
}

func NewPlayer(ss *sprites.SpriteSheet) *Player {
	return &Player{
		TileX:      3,
		TileY:      3,
		LeftFacing: true,
		Sprite:     ss.GreyKnight,
	}
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, isoToScreen func(int, int) (float64, float64), camX, camY, camScale, cx, cy float64) {
	x, y := isoToScreen(p.TileX, p.TileY)

	op := &ebiten.DrawImageOptions{}

	const verticalOffset = 1.0 // tweak until it feels good

	// Then apply:
	op.GeoM.Translate(0, -verticalOffset)

	// 2. Flip horizontally if facing right
	if !p.LeftFacing {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(tileSize), 0)
	}

	// 3. Move to isometric position
	op.GeoM.Translate(x, y)

	// 4. Apply camera transform
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	screen.DrawImage(p.Sprite, op)
}

func (p *Player) CanMoveTo(x, y int, level *levels.Level) bool {
	return x >= 0 && y >= 0 && x < level.W && y < level.H
}

func BuildPath(sx, sy, tx, ty int) []PathNode {
	var path []PathNode
	dx := tx - sx
	dy := ty - sy

	steps := max(abs(dx), abs(dy))
	for i := 1; i <= steps; i++ {
		x := sx + i*dx/steps
		y := sy + i*dy/steps
		path = append(path, PathNode{X: x, Y: y})
	}
	return path
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (p *Player) Update(level *levels.Level) {
	if len(p.Path) == 0 {
		p.TickCount = 0
		return
	}

	p.TickCount++
	if p.TickCount >= 60 {
		next := p.Path[0]
		if p.CanMoveTo(next.X, next.Y, level) {
			if next.X > p.TileX {
				p.LeftFacing = false
			} else if next.X < p.TileX {
				p.LeftFacing = true
			}
			p.TileX = next.X
			p.TileY = next.Y
			p.Path = p.Path[1:]
		} else {
			// Stop path if invalid
			p.Path = nil
		}
		p.TickCount = 0
	}
}
