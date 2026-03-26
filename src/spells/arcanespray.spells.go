package spells

import (
	"image/color"
	"math"
	"math/rand/v2"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ArcaneSpray is a channeled short-range cone attack.
// While active, it continuously emits particles, drains mana, and damages
// enemies in the cone each tick.
type ArcaneSpray struct {
	Info SpellInfo

	OriginX, OriginY float64
	DirAngle         float64 // current direction toward cursor (cartesian radians)
	ConeAngle        float64 // half-angle of the cone in radians
	Range            float64 // max range in tiles

	Particles []sprayParticle

	// Channeling state.
	Channeling bool    // true while the player holds the key
	ManaDrain  float64 // mana consumed per second
	DmgTimer   float64 // time accumulator for damage ticks

	age      float64
	Finished bool
}

type sprayParticle struct {
	X, Y       float64
	DirX, DirY float64
	Speed      float64
	Life       float64
	MaxLife    float64
}

// NewArcaneSpray creates a channeled cone spray. Call UpdateChannel() each
// frame to feed it the current player/cursor positions and keep it alive.
func NewArcaneSpray(info SpellInfo, originX, originY, targetX, targetY float64) *ArcaneSpray {
	dirAngle := math.Atan2(targetY-originY, targetX-originX)
	coneHalf := 45.0 * math.Pi / 180 // 90° total cone

	return &ArcaneSpray{
		Info:       info,
		OriginX:    originX,
		OriginY:    originY,
		DirAngle:   dirAngle,
		ConeAngle:  coneHalf,
		Range:      3.0,
		Channeling: true,
		ManaDrain:  8, // mana per second
	}
}

// UpdateChannel is called each frame while the spray is active.
// It updates origin/direction and spawns new particles.
func (as *ArcaneSpray) UpdateChannel(originX, originY, targetX, targetY float64) {
	as.OriginX = originX
	as.OriginY = originY
	as.DirAngle = math.Atan2(targetY-originY, targetX-originX)
}

// StopChannel ends the spray. Existing particles will fade out.
func (as *ArcaneSpray) StopChannel() {
	as.Channeling = false
}

// IsInCone checks if a point (tile coords) is within the spray's cone.
func (as *ArcaneSpray) IsInCone(tx, ty float64) bool {
	dx := tx - as.OriginX
	dy := ty - as.OriginY
	dist := math.Hypot(dx, dy)
	if dist > as.Range {
		return false
	}
	angle := math.Atan2(dy, dx)
	diff := normalizeAngle(angle - as.DirAngle)
	return math.Abs(diff) <= as.ConeAngle
}

func (as *ArcaneSpray) Update(level *levels.Level, dt float64) {
	if as.Finished {
		return
	}
	as.age += dt

	// Spawn new particles while channeling.
	if as.Channeling {
		// Spawn 2-3 particles per frame.
		for range 2 + rand.IntN(2) {
			angle := as.DirAngle + (rand.Float64()*2-1)*as.ConeAngle
			speed := 6 + rand.Float64()*4
			as.Particles = append(as.Particles, sprayParticle{
				X:       as.OriginX,
				Y:       as.OriginY,
				DirX:    math.Cos(angle),
				DirY:    math.Sin(angle),
				Speed:   speed,
				MaxLife: 0.25 + rand.Float64()*0.15,
			})
		}
	}

	// Update existing particles.
	allDead := true
	for i := range as.Particles {
		p := &as.Particles[i]
		if p.Life >= p.MaxLife {
			continue
		}
		allDead = false
		p.Life += dt
		step := p.Speed * dt
		p.X += p.DirX * step
		p.Y += p.DirY * step

		// Stop at walls.
		tx := int(math.Floor(p.X))
		ty := int(math.Floor(p.Y))
		if !level.IsWalkable(tx, ty) {
			p.Life = p.MaxLife
		}
	}

	// Prune dead particles periodically to avoid unbounded growth.
	if len(as.Particles) > 100 {
		alive := as.Particles[:0]
		for _, p := range as.Particles {
			if p.Life < p.MaxLife {
				alive = append(alive, p)
			}
		}
		as.Particles = alive
	}

	// Finish when not channeling and all particles are dead.
	if !as.Channeling && allDead {
		as.Finished = true
	}
}

func (as *ArcaneSpray) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if as.Finished {
		return
	}

	for _, p := range as.Particles {
		if p.Life >= p.MaxLife {
			continue
		}
		fade := 1.0 - p.Life/p.MaxLife
		col := color.NRGBA{
			R: uint8(170 + 50*fade),
			G: uint8(140 + 70*fade),
			B: 255,
			A: uint8(220 * fade),
		}

		sx, sy := isoToScreenFloat(p.X+1, p.Y, tileSize)
		sx = (sx-camX)*camScale + cx
		sy = (sy+camY)*camScale + cy

		r := float32(2+fade*3) * float32(camScale)
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), r, col, true)
	}
}

func (as *ArcaneSpray) IsFinished() bool { return as.Finished }
