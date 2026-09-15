package spells

import (
	"image/color"
	"math"

	"dungeoneer/coords"
	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// This file holds bespoke visual effects for the CR1 artifact skills — each
// one reads its shape from the artifact's mechanical identity (a slam cracks
// the ground, a bind wraps chains, an execute swings a scythe) rather than
// falling back to a generic colored particle burst.

// ─── Ironbreaker Gauntlets: ground slam shockwave ──────────────────────────

// GroundSlam is a ring of cracked earth radiating outward from the origin,
// with a bright iron flash at the moment of impact.
type GroundSlam struct {
	OriginX, OriginY float64
	Radius           float64
	age              float64
	Duration         float64
	Finished         bool
}

func NewGroundSlam(originX, originY, radius float64) *GroundSlam {
	return &GroundSlam{OriginX: originX, OriginY: originY, Radius: radius, Duration: 0.45}
}

func (s *GroundSlam) Update(level *levels.Level, dt float64) {
	s.age += dt
	if s.age >= s.Duration {
		s.Finished = true
	}
}

func (s *GroundSlam) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if s.Finished {
		return
	}
	progress := s.age / s.Duration
	alpha := float32(1 - progress)
	ringRadius := s.Radius * easeOutQuad(progress)

	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}

	// Fracture cracks: straight lines radiating from the origin, jagged like
	// broken stone, drawn once and fading with the ring.
	const cracks = 8
	crackCol := color.NRGBA{80, 70, 60, uint8(180 * alpha)}
	for i := 0; i < cracks; i++ {
		ang := float64(i) / cracks * 2 * math.Pi
		midR := ringRadius * 0.55
		x0, y0 := project(s.OriginX, s.OriginY)
		x1, y1 := project(s.OriginX+math.Cos(ang)*midR, s.OriginY+math.Sin(ang)*midR*0.5)
		x2, y2 := project(s.OriginX+math.Cos(ang)*ringRadius, s.OriginY+math.Sin(ang)*ringRadius*0.5)
		vector.StrokeLine(screen, x0, y0, x1, y1, 2, crackCol, true)
		vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, crackCol, true)
	}

	// Shockwave ring: steel-blue outer edge over a dust-brown inner band.
	const segs = 24
	outerCol := color.NRGBA{190, 205, 225, uint8(220 * alpha)}
	innerCol := color.NRGBA{140, 115, 90, uint8(140 * alpha)}
	for i := 0; i < segs; i++ {
		t1 := float64(i) / segs * 2 * math.Pi
		t2 := float64(i+1) / segs * 2 * math.Pi
		x1, y1 := project(s.OriginX+math.Cos(t1)*ringRadius, s.OriginY+math.Sin(t1)*ringRadius*0.5)
		x2, y2 := project(s.OriginX+math.Cos(t2)*ringRadius, s.OriginY+math.Sin(t2)*ringRadius*0.5)
		vector.StrokeLine(screen, x1, y1, x2, y2, 3, outerCol, true)
		ix1, iy1 := project(s.OriginX+math.Cos(t1)*ringRadius*0.75, s.OriginY+math.Sin(t1)*ringRadius*0.375)
		ix2, iy2 := project(s.OriginX+math.Cos(t2)*ringRadius*0.75, s.OriginY+math.Sin(t2)*ringRadius*0.375)
		vector.StrokeLine(screen, ix1, iy1, ix2, iy2, 4, innerCol, true)
	}

	// Impact flash at the origin, brightest on the first frames.
	if progress < 0.35 {
		flashA := uint8(200 * (1 - progress/0.35))
		ox, oy := project(s.OriginX, s.OriginY)
		vector.DrawFilledCircle(screen, ox, oy, float32(6+10*progress)*float32(camScale), color.NRGBA{230, 235, 245, flashA}, true)
	}
}

func (s *GroundSlam) IsFinished() bool { return s.Finished }

// ─── Shroud Cloak: shadow-strike teleport + guaranteed-crit flash ──────────

// ShadowStrike replaces the plain blink trail for shroud_cloak: violet-black
// smoke unspools from the origin, and a bright crit slash flashes at the
// destination the instant the player arrives, selling "guaranteed critical."
type ShadowStrike struct {
	StartX, StartY float64
	EndX, EndY     float64
	age            float64
	Duration       float64
	Finished       bool
}

func NewShadowStrike(startX, startY, endX, endY float64) *ShadowStrike {
	return &ShadowStrike{StartX: startX, StartY: startY, EndX: endX, EndY: endY, Duration: 0.4}
}

func (s *ShadowStrike) Update(level *levels.Level, dt float64) {
	s.age += dt
	if s.age >= s.Duration {
		s.Finished = true
	}
}

func (s *ShadowStrike) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if s.Finished {
		return
	}
	progress := s.age / s.Duration
	alpha := float32(1 - progress)

	// coords.BodyDX/BodyDY replaces the older ad hoc "+1 X only" chest-alignment
	// nudge (see the original in blink.spells.go) with the same named constant
	// every other bespoke CR1 visual in this file uses — same offset, just no
	// longer a magic number, and now also accounts for the Y component.
	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}
	ox, oy := project(s.StartX, s.StartY)
	dx, dy := project(s.EndX, s.EndY)

	// Origin: collapsing shadow puff.
	originR := float32(10*(1-progress)) * float32(camScale)
	vector.DrawFilledCircle(screen, ox, oy, originR, color.NRGBA{40, 15, 60, uint8(200 * alpha)}, true)

	// Three curling smoke tendrils drifting from origin toward the destination.
	for i := 0; i < 3; i++ {
		t := float64(i) / 3.0
		curl := math.Sin(progress*6+t*4) * 4
		mx := s.StartX + (s.EndX-s.StartX)*float64(progress)*(0.5+t*0.2)
		my := s.StartY + (s.EndY-s.StartY)*float64(progress)*(0.5+t*0.2)
		sx, sy := project(mx, my)
		vector.DrawFilledCircle(screen, sx+float32(curl), sy, float32(3)*float32(camScale), color.NRGBA{90, 40, 140, uint8(120 * alpha)}, true)
	}

	// Destination: bright crit-flash slash — an X-shaped double stroke that
	// snaps in immediately (near-instant on arrival, then fades).
	if progress > 0.25 {
		flashT := (progress - 0.25) / 0.75
		flashA := uint8(255 * (1 - flashT))
		flashR := float32(4 + 10*flashT)
		critCol := color.NRGBA{255, 235, 160, flashA}
		half := flashR * float32(camScale)
		vector.StrokeLine(screen, dx-half, dy-half*0.5, dx+half, dy+half*0.5, 3, critCol, true)
		vector.StrokeLine(screen, dx-half, dy+half*0.5, dx+half, dy-half*0.5, 3, critCol, true)
		vector.DrawFilledCircle(screen, dx, dy, float32(4+6*(1-flashT))*float32(camScale), color.NRGBA{150, 80, 220, uint8(150 * alpha)}, true)
	}
}

func (s *ShadowStrike) IsFinished() bool { return s.Finished }

// ─── Warden's Medallion: taunt aggro pulse ─────────────────────────────────

// TauntPulse is a golden aggro burst centered on the caster: expanding rings
// of light plus a rotating shield glyph, selling "all nearby enemies are now
// forced to look at me."
type TauntPulse struct {
	X, Y     float64
	Radius   float64
	age      float64
	Duration float64
	Finished bool
}

func NewTauntPulse(x, y, radius float64) *TauntPulse {
	return &TauntPulse{X: x, Y: y, Radius: radius, Duration: 0.7}
}

func (t *TauntPulse) Update(level *levels.Level, dt float64) {
	t.age += dt
	if t.age >= t.Duration {
		t.Finished = true
	}
}

func (t *TauntPulse) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if t.Finished {
		return
	}
	progress := t.age / t.Duration
	alpha := float32(1 - progress)

	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}
	ox, oy := project(t.X, t.Y)
	gold := color.NRGBA{255, 205, 90, 255}

	// Two expanding rings, staggered, so the pulse reads as a heartbeat rather
	// than a single flat burst.
	for _, phase := range [2]float64{0, 0.35} {
		p := progress - phase
		if p < 0 || p > 1 {
			continue
		}
		r := float32(t.Radius*easeOutQuad(p)) * float32(camScale) * float32(tileSize) / 4
		ringA := uint8(180 * (1 - p) * float64(alpha))
		col := gold
		col.A = ringA
		vector.StrokeCircle(screen, ox, oy, r, 2.5, col, true)
	}

	// Rotating shield glyph above the caster's head: three chevrons spinning
	// slowly, fading with the pulse.
	glyphY := oy - float32(18)*float32(camScale)
	spin := progress * 4
	for i := 0; i < 3; i++ {
		ang := spin*math.Pi + float64(i)*2*math.Pi/3
		gx := glyphY
		_ = gx
		px := ox + float32(math.Cos(ang))*6*float32(camScale)
		py := glyphY + float32(math.Sin(ang))*6*float32(camScale)
		vector.DrawFilledCircle(screen, px, py, float32(2.5)*float32(camScale), color.NRGBA{255, 225, 140, uint8(220 * alpha)}, true)
	}
}

func (t *TauntPulse) IsFinished() bool { return t.Finished }

// ─── Ashbound Chain: root bind ──────────────────────────────────────────────

// RootBind snaps three rusted chain links shut around the target, cinching
// tight, then holding taut for the bind's cast flash.
type RootBind struct {
	X, Y     float64
	age      float64
	Duration float64
	Finished bool
}

func NewRootBind(x, y float64) *RootBind {
	return &RootBind{X: x, Y: y, Duration: 0.5}
}

func (r *RootBind) Update(level *levels.Level, dt float64) {
	r.age += dt
	if r.age >= r.Duration {
		r.Finished = true
	}
}

func (r *RootBind) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if r.Finished {
		return
	}
	progress := r.age / r.Duration
	alpha := float32(1)
	if progress > 0.6 {
		alpha = float32(1 - (progress-0.6)/0.4)
	}
	cinch := math.Min(progress/0.4, 1.0) // chains snap shut over the first 40%

	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}
	cx0, cy0 := project(r.X, r.Y)
	rustCol := color.NRGBA{110, 70, 40, uint8(230 * alpha)}

	// Three chain loops closing from wide (dropping in from above) to tight
	// around the target's base.
	for i := 0; i < 3; i++ {
		startR := float32(16 - float32(i)*2)
		endR := float32(5 + float32(i)*1.5)
		r := startR + (endR-startR)*float32(cinch)
		r *= float32(camScale)
		yOff := float32(i-1) * 4 * float32(camScale)
		const links = 10
		for j := 0; j < links; j++ {
			t1 := float64(j) / links * 2 * math.Pi
			t2 := float64(j+1) / links * 2 * math.Pi
			x1 := cx0 + float32(math.Cos(t1))*r
			y1 := cy0 + yOff + float32(math.Sin(t1))*r*0.5
			x2 := cx0 + float32(math.Cos(t2))*r
			y2 := cy0 + yOff + float32(math.Sin(t2))*r*0.5
			// Draw every other segment to read as chain links, not a solid ring.
			if j%2 == 0 {
				vector.StrokeLine(screen, x1, y1, x2, y2, 2.5, rustCol, true)
			}
		}
	}

	// Snap flash at the moment the chains close.
	if progress < 0.15 {
		flashA := uint8(200 * (1 - progress/0.15))
		vector.DrawFilledCircle(screen, cx0, cy0, float32(10)*float32(camScale), color.NRGBA{160, 100, 60, flashA}, true)
	}
}

func (r *RootBind) IsFinished() bool { return r.Finished }

// ─── Grave Reaper: execute scythe ──────────────────────────────────────────

// ReaperExecute swings a dark void-edged scythe arc through the target; if
// the execute lands (Killed), a skull-shaped particle flash punctuates it.
type ReaperExecute struct {
	X, Y     float64
	Killed   bool
	age      float64
	Duration float64
	Finished bool
}

func NewReaperExecute(x, y float64, killed bool) *ReaperExecute {
	return &ReaperExecute{X: x, Y: y, Killed: killed, Duration: 0.5}
}

func (e *ReaperExecute) Update(level *levels.Level, dt float64) {
	e.age += dt
	if e.age >= e.Duration {
		e.Finished = true
	}
}

func (e *ReaperExecute) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if e.Finished {
		return
	}
	progress := e.age / e.Duration
	sweep := math.Min(progress/0.4, 1.0)
	alpha := float32(1)
	if progress > 0.5 {
		alpha = float32(1 - (progress-0.5)/0.5)
	}

	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}
	center, _ := project(e.X, e.Y)
	_ = center

	// Scythe arc: a wide dark-violet crescent sweeping across the target,
	// leading edge bright, trailing edge fading to black.
	const segs = 16
	startAngle := -math.Pi*0.75 + sweep*math.Pi*1.5
	radius := float32(14) * float32(camScale)
	for i := 0; i < segs; i++ {
		t1 := startAngle - float64(i)/segs*1.1
		t2 := startAngle - float64(i+1)/segs*1.1
		ox, oy := project(e.X, e.Y)
		x1 := ox + float32(math.Cos(t1))*radius
		y1 := oy + float32(math.Sin(t1))*radius*0.5
		x2 := ox + float32(math.Cos(t2))*radius
		y2 := oy + float32(math.Sin(t2))*radius*0.5
		fade := 1.0 - float64(i)/segs
		col := color.NRGBA{
			R: uint8(140 * fade), G: uint8(20 * fade), B: uint8(180 * fade),
			A: uint8(230 * float64(alpha) * fade),
		}
		vector.StrokeLine(screen, x1, y1, x2, y2, 4, col, true)
	}

	// Kill confirmation: a void-purple skull-shaped flash (two eye sockets +
	// a burst) at the target the instant the execute lands.
	if e.Killed && progress < 0.45 {
		ox, oy := project(e.X, e.Y)
		flashA := uint8(220 * (1 - progress/0.45))
		skullCol := color.NRGBA{220, 200, 255, flashA}
		vector.DrawFilledCircle(screen, ox, oy-4*float32(camScale), float32(6)*float32(camScale), skullCol, true)
		eyeCol := color.NRGBA{30, 10, 40, flashA}
		vector.DrawFilledCircle(screen, ox-2*float32(camScale), oy-5*float32(camScale), float32(1.5)*float32(camScale), eyeCol, true)
		vector.DrawFilledCircle(screen, ox+2*float32(camScale), oy-5*float32(camScale), float32(1.5)*float32(camScale), eyeCol, true)
	}
}

func (e *ReaperExecute) IsFinished() bool { return e.Finished }

// ─── Arcane Surge: converging nuke ─────────────────────────────────────────

// ArcaneSurgeNuke draws bolts converging from six compass directions into the
// target before detonating in a bright arcane flash — the more the player
// has on cooldown, the harder this hits, so the visual reads as "everything
// I've spent is being called back in."
type ArcaneSurgeNuke struct {
	X, Y     float64
	age      float64
	Duration float64
	Finished bool
}

func NewArcaneSurgeNuke(x, y float64) *ArcaneSurgeNuke {
	return &ArcaneSurgeNuke{X: x, Y: y, Duration: 0.55}
}

func (n *ArcaneSurgeNuke) Update(level *levels.Level, dt float64) {
	n.age += dt
	if n.age >= n.Duration {
		n.Finished = true
	}
}

func (n *ArcaneSurgeNuke) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if n.Finished {
		return
	}
	progress := n.age / n.Duration

	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}
	ox, oy := project(n.X, n.Y)
	arcaneCol := color.NRGBA{130, 170, 255, 220}

	if progress < 0.6 {
		// Converging phase: 6 bolts collapse inward from a wide ring.
		converge := progress / 0.6
		startRadius := float32(60) * float32(camScale)
		radius := startRadius * (1 - float32(easeOutQuad(converge)))
		const bolts = 6
		for i := 0; i < bolts; i++ {
			ang := float64(i) / bolts * 2 * math.Pi
			bx := ox + float32(math.Cos(ang))*radius
			by := oy + float32(math.Sin(ang))*radius*0.5
			col := arcaneCol
			col.A = uint8(230 * converge)
			vector.StrokeLine(screen, bx, by, ox, oy, 2.5, col, true)
			vector.DrawFilledCircle(screen, bx, by, float32(2.5)*float32(camScale), col, true)
		}
	} else {
		// Detonation flash: bright white-blue burst that expands and fades.
		burst := (progress - 0.6) / 0.4
		r := float32(4+22*burst) * float32(camScale)
		a := uint8(255 * (1 - burst))
		vector.DrawFilledCircle(screen, ox, oy, r, color.NRGBA{210, 225, 255, a}, true)
		vector.StrokeCircle(screen, ox, oy, r*1.3, 2, color.NRGBA{130, 170, 255, a}, true)
	}
}

func (n *ArcaneSurgeNuke) IsFinished() bool { return n.Finished }

// ─── Blood Price: sacrifice lance ──────────────────────────────────────────

// BloodPriceStrike shows the HP cost leaving the caster as a red mist that
// crosses the gap to the target and lands as a void-black lance impact —
// the price paid, made visible.
type BloodPriceStrike struct {
	CasterX, CasterY float64
	TargetX, TargetY float64
	age              float64
	Duration         float64
	Finished         bool
}

func NewBloodPriceStrike(casterX, casterY, targetX, targetY float64) *BloodPriceStrike {
	return &BloodPriceStrike{CasterX: casterX, CasterY: casterY, TargetX: targetX, TargetY: targetY, Duration: 0.45}
}

func (b *BloodPriceStrike) Update(level *levels.Level, dt float64) {
	b.age += dt
	if b.age >= b.Duration {
		b.Finished = true
	}
}

func (b *BloodPriceStrike) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if b.Finished {
		return
	}
	progress := b.age / b.Duration

	project := func(wx, wy float64) (float32, float32) {
		sx, sy := isoToScreenFloat(wx+coords.BodyDX, wy+coords.BodyDY, tileSize)
		return float32((sx-camX)*camScale + cx), float32((sy+camY)*camScale + cy)
	}
	ox, oy := project(b.CasterX, b.CasterY)
	tx, ty := project(b.TargetX, b.TargetY)

	// Blood-red pulse at the caster — the HP cost being spent.
	casterA := uint8(220 * (1 - math.Min(progress/0.3, 1.0)))
	vector.DrawFilledCircle(screen, ox, oy, float32(8-4*progress)*float32(camScale), color.NRGBA{200, 20, 30, casterA}, true)

	// Lance: a dark void bolt with a blood-red core travels from caster to
	// target over the first 70% of the effect.
	travel := math.Min(progress/0.7, 1.0)
	lx := ox + (tx-ox)*float32(travel)
	ly := oy + (ty-oy)*float32(travel)
	vector.StrokeLine(screen, ox, oy, lx, ly, 3, color.NRGBA{40, 5, 50, 200}, true)
	vector.DrawFilledCircle(screen, lx, ly, float32(4)*float32(camScale), color.NRGBA{180, 20, 40, 230}, true)

	// Impact burst once the lance lands.
	if progress > 0.6 {
		burst := (progress - 0.6) / 0.4
		r := float32(3+12*burst) * float32(camScale)
		a := uint8(220 * (1 - burst))
		vector.DrawFilledCircle(screen, tx, ty, r, color.NRGBA{60, 10, 70, a}, true)
		vector.StrokeCircle(screen, tx, ty, r*1.2, 2, color.NRGBA{180, 20, 40, a}, true)
	}
}

func (b *BloodPriceStrike) IsFinished() bool { return b.Finished }

// ─── shared helpers ─────────────────────────────────────────────────────────

func easeOutQuad(t float64) float64 {
	if t > 1 {
		t = 1
	}
	if t < 0 {
		t = 0
	}
	return 1 - (1-t)*(1-t)
}
