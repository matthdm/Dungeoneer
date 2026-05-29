package spells

import (
	"image/color"
	"math"
	"math/rand/v2"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Point struct {
	X, Y float64
}

// ChaosRay is an instant lightning-like beam.
type ChaosRay struct {
	Info     SpellInfo
	Path     []Point
	Duration float64
	Age      float64
	Finished bool
}

func NewChaosRay(info SpellInfo, startX, startY, endX, endY float64) *ChaosRay {
	cr := &ChaosRay{Info: info, Duration: 0.5}
	cr.Path = generateJaggedLightningPath(startX, startY, endX, endY)
	return cr
}

func generateJaggedLightningPath(x1, y1, x2, y2 float64) []Point {
	length := math.Hypot(x2-x1, y2-y1)
	segments := int(length * 2) // 2 segments per unit of distance
	if segments < 4 {
		segments = 4
	}
	if segments > 40 {
		segments = 40 // cap to avoid performance hits
	}

	points := make([]Point, segments+1)
	points[0] = Point{x1, y1}
	points[segments] = Point{x2, y2}

	dx := x2 - x1
	dy := y2 - y1
	lineLength := math.Hypot(dx, dy)
	nx := -dy / lineLength
	ny := dx / lineLength

	for i := 1; i < segments; i++ {
		t := float64(i) / float64(segments)
		// Apply jitter to t (non-uniform spacing)
		t += (rand.Float64()*2 - 1) * 0.05
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		px := x1 + dx*t
		py := y1 + dy*t

		// Random offset perpendicular to direction
		magnitude := 0.2 + 0.3*rand.Float64()
		offset := (rand.Float64()*2 - 1) * magnitude

		px += nx * offset
		py += ny * offset

		points[i] = Point{px, py}
	}

	return points
}

func (c *ChaosRay) Update(level *levels.Level, dt float64) {
	c.Age += dt
	if c.Age >= c.Duration {
		c.Finished = true
	}
}

func (c *ChaosRay) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if c.Finished {
		return
	}
	alpha := float32(1 - c.Age/c.Duration)
	col := color.NRGBA{R: 200, G: 220, B: 255, A: uint8(255 * alpha)}
	for i := 0; i < len(c.Path)-1; i++ {
		p1 := c.Path[i]
		p2 := c.Path[i+1]
		sx1, sy1 := isoToScreenFloat(p1.X, p1.Y, tileSize)
		sx2, sy2 := isoToScreenFloat(p2.X, p2.Y, tileSize)
		sx1 = (sx1-camX)*camScale + cx
		sy1 = (sy1+camY)*camScale + cy
		sx2 = (sx2-camX)*camScale + cx
		sy2 = (sy2+camY)*camScale + cy
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 3, col, true)
	}
}

func (c *ChaosRay) IsFinished() bool { return c.Finished }
