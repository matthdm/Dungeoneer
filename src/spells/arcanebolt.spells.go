package spells

import (
	"image/color"
	"math"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ArcaneBolt is a small, fast magic projectile with a fading trail.
// It is the mage's basic attack — low cost, rapid fire.
type ArcaneBolt struct {
	Info       SpellInfo
	X, Y       float64
	DirX, DirY float64
	Speed      float64
	Radius     float64 // hitbox radius (tiles)

	// Trail: ring buffer of recent positions.
	trail    [8]Point
	trailIdx int
	trailLen int

	Impact   bool
	Finished bool
	age      float64
}

func NewArcaneBolt(info SpellInfo, startX, startY, targetX, targetY float64) *ArcaneBolt {
	dx := targetX - startX
	dy := targetY - startY
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1
	}
	dx /= dist
	dy /= dist
	ab := &ArcaneBolt{
		Info:   info,
		X:      startX,
		Y:      startY,
		DirX:   dx,
		DirY:   dy,
		Speed:  18,
		Radius: 0.4,
	}
	// Fill trail with starting position.
	for i := range ab.trail {
		ab.trail[i] = Point{startX, startY}
	}
	return ab
}

func (ab *ArcaneBolt) Update(level *levels.Level, dt float64) {
	if ab.Finished {
		return
	}
	if ab.Impact {
		ab.age += dt
		if ab.age > 0.15 {
			ab.Finished = true
		}
		return
	}

	// Record position in trail before moving.
	ab.trail[ab.trailIdx] = Point{ab.X, ab.Y}
	ab.trailIdx = (ab.trailIdx + 1) % len(ab.trail)
	if ab.trailLen < len(ab.trail) {
		ab.trailLen++
	}

	step := ab.Speed * dt
	ab.X += ab.DirX * step
	ab.Y += ab.DirY * step

	// Wall collision.
	tx := int(math.Floor(ab.X))
	ty := int(math.Floor(ab.Y))
	if !level.IsWalkable(tx, ty) {
		ab.Impact = true
		ab.age = 0
	}

	// Max range: ~12 tiles.
	origin := ab.trail[0]
	if math.Hypot(ab.X-origin.X, ab.Y-origin.Y) > 12 {
		ab.Finished = true
	}
}

func (ab *ArcaneBolt) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if ab.Finished {
		return
	}

	boltColor := color.NRGBA{170, 140, 255, 255} // pale violet
	trailColor := color.NRGBA{120, 90, 200, 180}

	if ab.Impact {
		// Impact spark: expanding then fading circle.
		sx, sy := isoToScreenFloat(ab.X+1, ab.Y, tileSize)
		sx = (sx-camX)*camScale + cx
		sy = (sy+camY)*camScale + cy
		alpha := float32(1 - ab.age/0.15)
		sparkR := float32(4+ab.age*40) * float32(camScale)
		sparkCol := color.NRGBA{200, 180, 255, uint8(180 * alpha)}
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), sparkR, sparkCol, true)
		return
	}

	// Draw trail.
	for i := 0; i < ab.trailLen; i++ {
		idx := (ab.trailIdx - 1 - i + len(ab.trail)) % len(ab.trail)
		p := ab.trail[idx]
		sx, sy := isoToScreenFloat(p.X+1, p.Y, tileSize)
		sx = (sx-camX)*camScale + cx
		sy = (sy+camY)*camScale + cy

		fade := 1.0 - float64(i)/float64(len(ab.trail))
		col := trailColor
		col.A = uint8(float64(col.A) * fade)
		r := float32(2+fade*2) * float32(camScale)
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), r, col, true)
	}

	// Draw bolt head.
	sx, sy := isoToScreenFloat(ab.X+1, ab.Y, tileSize)
	sx = (sx-camX)*camScale + cx
	sy = (sy+camY)*camScale + cy
	r := float32(4) * float32(camScale)
	vector.DrawFilledCircle(screen, float32(sx), float32(sy), r, boltColor, true)

	// Bright center.
	centerCol := color.NRGBA{220, 210, 255, 255}
	vector.DrawFilledCircle(screen, float32(sx), float32(sy), r*0.5, centerCol, true)
}

func (ab *ArcaneBolt) IsFinished() bool { return ab.Finished }
