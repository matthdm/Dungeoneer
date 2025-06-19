package spells

import (
	"image/color"
	"math/rand"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type lightningPoint struct {
	X, Y float64
}

// LightningStrike represents a vertical bolt that strikes a tile from above.
type LightningStrike struct {
	Info          SpellInfo
	X, Y          float64
	points        []lightningPoint
	age           float64
	Duration      float64
	DamageApplied bool
	ImpactImg     *ebiten.Image
	Finished      bool
}

// NewLightningStrike creates a new lightning strike spell targeting the given tile.
func NewLightningStrike(info SpellInfo, targetX, targetY float64, impact *ebiten.Image) *LightningStrike {
	startY := targetY - 5 - rand.Float64()*5
	segments := 6
	pts := make([]lightningPoint, segments+1)
	pts[0] = lightningPoint{X: targetX, Y: startY}
	for i := 1; i < segments; i++ {
		t := float64(i) / float64(segments)
		py := startY + (targetY-startY)*t
		px := targetX + (rand.Float64()-0.5)*0.5
		pts[i] = lightningPoint{X: px, Y: py}
	}
	pts[segments] = lightningPoint{X: targetX, Y: targetY}
	return &LightningStrike{
		Info:      info,
		X:         targetX,
		Y:         targetY,
		points:    pts,
		Duration:  .4,
		ImpactImg: impact,
	}
}

func (l *LightningStrike) Update(level *levels.Level, dt float64) {
	if l.Finished {
		return
	}
	l.age += dt
	if l.age >= l.Duration {
		l.Finished = true
	}
}

func (l *LightningStrike) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if l.Finished {
		return
	}

	// Fade out
	alpha := float32(1 - l.age/l.Duration)
	if alpha < 0 {
		alpha = 0
	}

	if l.age <= l.Duration*0.2 {
		boltColor := color.NRGBA{R: 200, G: 220, B: 255, A: uint8(255 * alpha)}

		centerX, centerY := screenCoordsFromWorldTile(l.X, l.Y, tileSize, camX, camY, camScale, cx, cy)

		const segmentCount = 10
		const strikeHeight = 150.0
		const jitter = 10.0

		screenSegments := make([]Point, segmentCount)
		for i := 0; i < segmentCount; i++ {
			t := float64(i) / float64(segmentCount-1)
			sx := centerX + (rand.Float64()-0.5)*jitter
			sy := centerY - t*strikeHeight
			screenSegments[i] = Point{sx, sy}
		}
		// Draw lightning segments
		for i := 0; i < len(screenSegments)-1; i++ {
			a := screenSegments[i]
			b := screenSegments[i+1]
			vector.StrokeLine(screen, float32(a.X), float32(a.Y), float32(b.X), float32(b.Y), 2, boltColor, true)
		}
	} else {
		// Draw impact sprite
		if l.ImpactImg != nil {
			sx, sy := isoToScreenFloat(l.X, l.Y, tileSize)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(sx, sy)
			op.GeoM.Translate(-camX, camY)
			op.GeoM.Scale(camScale, camScale)
			op.GeoM.Translate(cx, cy)
			screen.DrawImage(l.ImpactImg, op)
		}
	}
}

func (l *LightningStrike) IsFinished() bool { return l.Finished }
func screenCoordsFromWorldTile(tileX, tileY float64, tileSize int, camX, camY, camScale, cx, cy float64) (float64, float64) {
	ix := (tileX - tileY) * float64(tileSize/2)
	iy := (tileX + tileY) * float64(tileSize/4)
	sx := (ix-camX+30)*camScale + cx
	sy := (iy+camY+30)*camScale + cy
	return sx, sy
}
