package spells

import (
	"image/color"
	"math"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BlinkMaxDistance is the maximum blink range in tiles.
const BlinkMaxDistance = 4.0

// FindBlinkTarget walks from start toward the cursor direction, stepping
// through tiles and stopping at the last walkable position. The player
// can never blink into, through, or past walls.
func FindBlinkTarget(level *levels.Level, startX, startY, targetX, targetY float64) (float64, float64) {
	dx := targetX - startX
	dy := targetY - startY
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return startX, startY
	}

	// Normalize direction.
	dx /= dist
	dy /= dist

	// Clamp to max blink distance.
	if dist > BlinkMaxDistance {
		dist = BlinkMaxDistance
	}

	// Walk the path in small steps, checking walkability.
	stepSize := 0.25
	steps := int(math.Ceil(dist / stepSize))
	lastX, lastY := startX, startY

	for i := 1; i <= steps; i++ {
		t := float64(i) * stepSize
		if t > dist {
			t = dist
		}
		x := startX + dx*t
		y := startY + dy*t
		tx := int(math.Floor(x))
		ty := int(math.Floor(y))
		if !level.IsWalkable(tx, ty) {
			break
		}
		lastX = x
		lastY = y
	}

	return lastX, lastY
}

// BlinkEffect is a purely visual spell that shows the blink trail.
type BlinkEffect struct {
	StartX, StartY float64
	EndX, EndY     float64
	age            float64
	Duration       float64
	Finished       bool
}

func NewBlinkEffect(startX, startY, endX, endY float64) *BlinkEffect {
	return &BlinkEffect{
		StartX:   startX,
		StartY:   startY,
		EndX:     endX,
		EndY:     endY,
		Duration: 0.3,
	}
}

func (b *BlinkEffect) Update(level *levels.Level, dt float64) {
	b.age += dt
	if b.age >= b.Duration {
		b.Finished = true
	}
}

func (b *BlinkEffect) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if b.Finished {
		return
	}
	alpha := float32(1 - b.age/b.Duration)

	// Origin flash: shrinking circle. +1 X offset aligns with player chest.
	sx1, sy1 := isoToScreenFloat(b.StartX+1, b.StartY, tileSize)
	sx1 = (sx1-camX)*camScale + cx
	sy1 = (sy1+camY)*camScale + cy
	originR := float32(8*(1-b.age/b.Duration)) * float32(camScale)
	originCol := color.NRGBA{160, 120, 255, uint8(200 * alpha)}
	vector.DrawFilledCircle(screen, float32(sx1), float32(sy1), originR, originCol, true)

	// Destination flash: expanding circle.
	sx2, sy2 := isoToScreenFloat(b.EndX+1, b.EndY, tileSize)
	sx2 = (sx2-camX)*camScale + cx
	sy2 = (sy2+camY)*camScale + cy
	destR := float32(3+5*(b.age/b.Duration)) * float32(camScale)
	destCol := color.NRGBA{200, 180, 255, uint8(220 * alpha)}
	vector.DrawFilledCircle(screen, float32(sx2), float32(sy2), destR, destCol, true)

	// Trail line connecting start to end.
	trailCol := color.NRGBA{170, 140, 255, uint8(120 * alpha)}
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 2, trailCol, true)

	// Midpoint sparkle particles (3 static dots along the path).
	for i := 1; i <= 3; i++ {
		t := float64(i) / 4.0
		mx := b.StartX + (b.EndX-b.StartX)*t
		my := b.StartY + (b.EndY-b.StartY)*t
		smx, smy := isoToScreenFloat(mx+1, my, tileSize)
		smx = (smx-camX)*camScale + cx
		smy = (smy+camY)*camScale + cy
		sparkCol := color.NRGBA{220, 200, 255, uint8(150 * alpha)}
		vector.DrawFilledCircle(screen, float32(smx), float32(smy), 2*float32(camScale), sparkCol, true)
	}
}

func (b *BlinkEffect) IsFinished() bool { return b.Finished }
