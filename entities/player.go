package entities

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	TileX, TileY int // current grid position
	LeftFacing   bool
	Sprite       *ebiten.Image

	Path      []PathNode
	TickCount int

	// Smooth movement
	InterpX, InterpY float64
	Moving           bool
	StartX, StartY   float64
	TargetX, TargetY float64
	InterpTicks      int
	BobOffset        float64
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

		InterpX: float64(3),
		InterpY: float64(3),
	}
}

func generateShadowImage(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size/4)
	// dark gray ellipse
	img.Fill(color.RGBA{0, 0, 0, 80})
	return img
}

func isoToScreenFloat(x, y float64, tileSize int) (float64, float64) {
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, isoToScreen func(int, int) (float64, float64), camX, camY, camScale, cx, cy float64) {
	//x, y := isoToScreen(int(p.InterpX), int(p.InterpY))
	// Optional: more accurate rendering
	x, y := isoToScreenFloat(p.InterpX, p.InterpY, 64)

	op := &ebiten.DrawImageOptions{}
	const verticalOffset = 1.0 // tweak until it feels good

	// Then apply:
	op.GeoM.Translate(0, -verticalOffset+p.BobOffset)

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
	const moveDuration = 15 // ticks per tile
	var bobAmplitude = 1.5
	var bobFrequency = 0.3 // radians per tick
	p.TickCount++
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFrequency) * bobAmplitude

	const swayAmplitude = 1.5
	const swayFrequency = 0.15

	p.SwayOffset = math.Sin(float64(p.TickCount)*swayFrequency+math.Pi) * swayAmplitude
	if p.Moving {
		p.InterpTicks++
		t := float64(p.InterpTicks) / float64(moveDuration)
		if t > 1 {
			t = 1
		}

		p.InterpX = p.StartX + (p.TargetX-p.StartX)*t
		p.InterpY = p.StartY + (p.TargetY-p.StartY)*t

		p.BobOffset = math.Sin(float64(p.InterpTicks)*bobFrequency) * bobAmplitude

		if t >= 1 {
			p.Moving = false
			p.TileX = int(p.TargetX)
			p.TileY = int(p.TargetY)
			p.InterpX = p.TargetX
			p.InterpY = p.TargetY
		}
		return
	} else {
		bobFrequency = 0.1
	}
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFrequency) * bobAmplitude

	// If not moving, check for next tile in path
	if len(p.Path) > 0 {
		next := p.Path[0]

		if next.X > p.TileX {
			p.LeftFacing = false
		} else if next.X < p.TileX {
			p.LeftFacing = true
		}

		if !p.CanMoveTo(next.X, next.Y, level) {
			p.Path = nil
			return
		}

		p.StartX = p.InterpX
		p.StartY = p.InterpY
		p.TargetX = float64(next.X)
		p.TargetY = float64(next.Y)
		p.InterpTicks = 0
		p.Moving = true

		p.Path = p.Path[1:]
	}
}
