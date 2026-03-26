package spells

import (
	"image/color"
	"math"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SlashHit defines the properties for each combo hit.
type SlashHit struct {
	ArcDegrees float64 // arc width in degrees
	Radius     float64 // range in tiles
	DamageMult float64 // multiplier applied to base damage
	SweepTime  float64 // how long the arc animates (seconds)
	FadeTime   float64 // how long the trail lingers (seconds)
	LineWidth  float32 // stroke width of the arc
}

// Combo hit definitions.
var SlashComboHits = [3]SlashHit{
	{ArcDegrees: 120, Radius: 1.8, DamageMult: 1.0, SweepTime: 0.12, FadeTime: 0.18, LineWidth: 3},
	{ArcDegrees: 80, Radius: 1.8, DamageMult: 1.2, SweepTime: 0.10, FadeTime: 0.15, LineWidth: 3.5},
	{ArcDegrees: 360, Radius: 1.2, DamageMult: 1.8, SweepTime: 0.15, FadeTime: 0.25, LineWidth: 4},
}

// SlashArc is a melee sweep spell drawn as a procedural arc.
// It checks hits in cartesian space but renders in isometric.
type SlashArc struct {
	Info SpellInfo

	// Origin in tile/cartesian space.
	OriginX, OriginY float64

	// Direction the slash faces (radians, cartesian space).
	DirAngle float64

	// Combo hit index (0, 1, 2).
	ComboHit int

	// Derived from SlashComboHits.
	ArcRadians float64
	Radius     float64
	SweepTime  float64
	FadeTime   float64
	LineWidth  float32

	// Arc points pre-computed in cartesian space for drawing.
	ArcPoints []Point

	// Damage targets: monster tile positions that were hit.
	// Populated externally by the game after creation.
	HitTargets []Point

	age      float64
	Finished bool
}

// NewSlashArc creates a melee slash targeting a direction.
// originX/Y: player position in tile space.
// dirAngle: angle toward the cursor in cartesian radians.
// comboHit: 0, 1, or 2.
func NewSlashArc(info SpellInfo, originX, originY, dirAngle float64, comboHit int) *SlashArc {
	if comboHit < 0 || comboHit > 2 {
		comboHit = 0
	}
	hit := SlashComboHits[comboHit]
	arcRad := hit.ArcDegrees * math.Pi / 180

	// Pre-compute arc points in cartesian space (used for drawing).
	segments := 12
	points := make([]Point, segments+1)
	startAngle := dirAngle - arcRad/2
	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		angle := startAngle + t*arcRad
		points[i] = Point{
			X: originX + math.Cos(angle)*hit.Radius,
			Y: originY + math.Sin(angle)*hit.Radius,
		}
	}

	return &SlashArc{
		Info:       info,
		OriginX:    originX,
		OriginY:    originY,
		DirAngle:   dirAngle,
		ComboHit:   comboHit,
		ArcRadians: arcRad,
		Radius:     hit.Radius,
		SweepTime:  hit.SweepTime,
		FadeTime:   hit.FadeTime,
		LineWidth:  hit.LineWidth,
		ArcPoints:  points,
	}
}

// IsInArc checks whether a point (cartesian tile coords) falls within this slash's arc.
func (s *SlashArc) IsInArc(tx, ty float64) bool {
	dx := tx - s.OriginX
	dy := ty - s.OriginY
	dist := math.Hypot(dx, dy)
	if dist > s.Radius {
		return false
	}
	// For the heavy slam (hit 3), it's a full circle — skip angle check.
	if s.ComboHit == 2 {
		return true
	}
	angle := math.Atan2(dy, dx)
	diff := normalizeAngle(angle - s.DirAngle)
	return math.Abs(diff) <= s.ArcRadians/2
}

// normalizeAngle wraps an angle to [-pi, pi].
func normalizeAngle(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

func (s *SlashArc) Update(level *levels.Level, dt float64) {
	s.age += dt
	if s.age >= s.SweepTime+s.FadeTime {
		s.Finished = true
	}
}

func (s *SlashArc) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if s.Finished {
		return
	}

	segments := len(s.ArcPoints) - 1
	if segments < 1 {
		return
	}

	// Colors per combo hit.
	type slashColors struct {
		edge, trail color.NRGBA
	}
	palette := [3]slashColors{
		{color.NRGBA{240, 245, 255, 255}, color.NRGBA{140, 180, 220, 255}}, // hit 1: white → steel blue
		{color.NRGBA{255, 255, 220, 255}, color.NRGBA{200, 220, 255, 255}}, // hit 2: bright yellow-white → bright blue
		{color.NRGBA{255, 200, 100, 255}, color.NRGBA{255, 120, 40, 255}},  // hit 3: gold → orange (heavy slam)
	}
	colors := palette[s.ComboHit]

	// Global alpha fade during the fade phase.
	globalAlpha := float64(1)
	if s.age > s.SweepTime {
		globalAlpha = 1 - (s.age-s.SweepTime)/s.FadeTime
	}
	if globalAlpha < 0 {
		globalAlpha = 0
	}

	// During sweep: progressively reveal segments.
	// During fade: all segments visible but fading.
	var visibleSegments int
	if s.age < s.SweepTime {
		progress := s.age / s.SweepTime
		visibleSegments = int(math.Ceil(progress * float64(segments)))
	} else {
		visibleSegments = segments
	}

	// Heavy slam (hit 3): draw expanding ring instead of arc.
	if s.ComboHit == 2 {
		s.drawSlamRing(screen, tileSize, camX, camY, camScale, cx, cy, globalAlpha, colors)
		return
	}

	// Draw arc segments — later segments (leading edge) are brighter.
	for i := 0; i < visibleSegments; i++ {
		p1 := s.ArcPoints[i]
		p2 := s.ArcPoints[i+1]

		sx1, sy1 := isoToScreenFloat(p1.X, p1.Y, tileSize)
		sx2, sy2 := isoToScreenFloat(p2.X, p2.Y, tileSize)
		sx1 = (sx1-camX)*camScale + cx
		sy1 = (sy1+camY)*camScale + cy
		sx2 = (sx2-camX)*camScale + cx
		sy2 = (sy2+camY)*camScale + cy

		// Fade trail: segments near the start are dimmer.
		segFrac := float64(i) / float64(segments)
		col := lerpColor(colors.trail, colors.edge, segFrac)
		col.A = uint8(float64(col.A) * globalAlpha)

		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), s.LineWidth, col, true)
	}

	// Draw a connecting line from origin to the start of the arc (the "blade arm").
	if visibleSegments > 0 {
		tip := s.ArcPoints[visibleSegments]
		osx, osy := isoToScreenFloat(s.OriginX, s.OriginY, tileSize)
		osx = (osx-camX)*camScale + cx
		osy = (osy+camY)*camScale + cy
		tsx, tsy := isoToScreenFloat(tip.X, tip.Y, tileSize)
		tsx = (tsx-camX)*camScale + cx
		tsy = (tsy+camY)*camScale + cy

		armColor := colors.edge
		armColor.A = uint8(float64(armColor.A) * globalAlpha * 0.5)
		vector.StrokeLine(screen, float32(osx), float32(osy), float32(tsx), float32(tsy), 1.5, armColor, true)
	}
}

func (s *SlashArc) drawSlamRing(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64, globalAlpha float64, colors struct{ edge, trail color.NRGBA }) {
	// Expanding ring: grows from 0 to full radius during sweep, then fades.
	var ringRadius float64
	if s.age < s.SweepTime {
		ringRadius = s.Radius * (s.age / s.SweepTime)
	} else {
		ringRadius = s.Radius
	}

	ringSegments := 20
	for i := 0; i < ringSegments; i++ {
		t1 := float64(i) / float64(ringSegments) * 2 * math.Pi
		t2 := float64(i+1) / float64(ringSegments) * 2 * math.Pi
		p1x := s.OriginX + math.Cos(t1)*ringRadius
		p1y := s.OriginY + math.Sin(t1)*ringRadius
		p2x := s.OriginX + math.Cos(t2)*ringRadius
		p2y := s.OriginY + math.Sin(t2)*ringRadius

		sx1, sy1 := isoToScreenFloat(p1x, p1y, tileSize)
		sx2, sy2 := isoToScreenFloat(p2x, p2y, tileSize)
		sx1 = (sx1-camX)*camScale + cx
		sy1 = (sy1+camY)*camScale + cy
		sx2 = (sx2-camX)*camScale + cx
		sy2 = (sy2+camY)*camScale + cy

		col := lerpColor(colors.trail, colors.edge, 0.7)
		col.A = uint8(float64(col.A) * globalAlpha)
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), s.LineWidth, col, true)
	}

	// Inner flash during sweep.
	if s.age < s.SweepTime {
		flash := colors.edge
		flash.A = uint8(80 * globalAlpha)
		osx, osy := isoToScreenFloat(s.OriginX, s.OriginY, tileSize)
		osx = (osx-camX)*camScale + cx
		osy = (osy+camY)*camScale + cy
		vector.DrawFilledCircle(screen, float32(osx), float32(osy), float32(ringRadius*camScale*float64(tileSize/4)), flash, true)
	}
}

func (s *SlashArc) IsFinished() bool { return s.Finished }
