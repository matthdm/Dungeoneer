package game

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

const particlePoolSize = 512

// Particle is a single visual particle (pooled, no heap alloc after init).
type Particle struct {
	Active  bool
	X, Y    float64 // screen position
	VX, VY  float64 // velocity pixels/s
	Life    float64 // remaining lifetime (seconds)
	MaxLife float64
	R, G, B float32 // color (0-1)
	Size    float64 // radius in pixels
}

// ParticlePool is a fixed-size pool of particles — never allocates after creation.
type ParticlePool struct {
	particles [particlePoolSize]Particle
	next      int // next free slot hint
}

// NewParticlePool creates a pre-allocated particle pool.
func NewParticlePool() *ParticlePool {
	return &ParticlePool{}
}

// Emit activates up to count particles at screen position (sx, sy) with the given color.
func (p *ParticlePool) Emit(sx, sy float64, count int, r, g, b float32) {
	emitted := 0
	for i := 0; i < particlePoolSize && emitted < count; i++ {
		idx := (p.next + i) % particlePoolSize
		if p.particles[idx].Active {
			continue
		}
		angle := rand.Float64() * 2 * math.Pi
		speed := 30 + rand.Float64()*60
		life := 0.3 + rand.Float64()*0.3
		p.particles[idx] = Particle{
			Active:  true,
			X:       sx,
			Y:       sy,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    life,
			MaxLife: life,
			R:       r,
			G:       g,
			B:       b,
			Size:    2 + rand.Float64()*2,
		}
		emitted++
		p.next = (idx + 1) % particlePoolSize
	}
}

// Update advances all active particles by dt seconds.
func (p *ParticlePool) Update(dt float64) {
	for i := range p.particles {
		pt := &p.particles[i]
		if !pt.Active {
			continue
		}
		pt.X += pt.VX * dt
		pt.Y += pt.VY * dt
		pt.VY += 40 * dt // slight gravity
		pt.Life -= dt
		if pt.Life <= 0 {
			pt.Active = false
		}
	}
}

// Draw renders all active particles to screen.
// NOTE: ebiten.NewImage per particle is a per-frame allocation. This is acceptable
// for MVP. A proper fix would pre-allocate small fixed-size image slabs.
func (p *ParticlePool) Draw(screen *ebiten.Image) {
	for i := range p.particles {
		pt := &p.particles[i]
		if !pt.Active {
			continue
		}
		alpha := float32(pt.Life / pt.MaxLife)
		size := int(pt.Size)
		if size < 1 {
			size = 1
		}
		dot := ebiten.NewImage(size, size)
		dot.Fill(color.NRGBA{
			R: uint8(pt.R * 255),
			G: uint8(pt.G * 255),
			B: uint8(pt.B * 255),
			A: uint8(alpha * 200),
		})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pt.X-float64(size)/2, pt.Y-float64(size)/2)
		screen.DrawImage(dot, op)
	}
}

// SpellParticleColor returns the particle color for a given spell/damage type.
func SpellParticleColor(dmgType string) (r, g, b float32) {
	switch dmgType {
	case "fire", "fireball":
		return 1.0, 0.4, 0.1
	case "lightning", "lightning_storm", "lightning_aura":
		return 0.5, 0.8, 1.0
	case "nature", "fractal_bloom", "fractal_canopy", "fractal_echo":
		return 0.3, 1.0, 0.3
	case "chaos", "chaos_ray":
		return 0.8, 0.2, 1.0
	case "arcane", "arcane_bolt":
		return 0.9, 0.9, 1.0
	default:
		return 1.0, 1.0, 1.0
	}
}

// DamageNumberColor returns color for a damage number based on type and crit status.
func DamageNumberColor(dmgType string, isCrit bool) color.NRGBA {
	if isCrit {
		return color.NRGBA{255, 220, 0, 255} // yellow for crits
	}
	switch dmgType {
	case "fire", "fireball":
		return color.NRGBA{255, 130, 30, 255}
	case "lightning", "lightning_storm":
		return color.NRGBA{130, 200, 255, 255}
	case "nature", "fractal_bloom", "fractal_canopy":
		return color.NRGBA{80, 220, 80, 255}
	case "chaos", "chaos_ray":
		return color.NRGBA{200, 80, 255, 255}
	default:
		return color.NRGBA{255, 255, 255, 255}
	}
}
