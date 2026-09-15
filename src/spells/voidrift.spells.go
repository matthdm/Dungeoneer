package spells

import (
	"dungeoneer/coords"
	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// VoidRift is void_rift_catalyst's active-skill visual: a shader-driven
// collapsing rift at the target's position — it expands over the first 40%
// of its lifetime, then collapses, "opens, pulls, snaps shut." Damage and
// root are already applied by the combat engine before this is spawned (see
// EventSkillFired handling in game/new_adapter.go) — this is purely visual.
type VoidRift struct {
	X, Y     float64
	age      float64
	Duration float64
	Finished bool
}

// NewVoidRift creates a rift centered at the given world (cartesian) position,
// which should already be the target's body center (callers pass lockX/lockY
// per the existing artifact-skill convention in spawnSkillVisual).
func NewVoidRift(x, y float64) *VoidRift {
	return &VoidRift{X: x, Y: y, Duration: 0.6}
}

func (v *VoidRift) Update(level *levels.Level, dt float64) {
	if v.Finished {
		return
	}
	v.age += dt
	if v.age >= v.Duration {
		v.Finished = true
	}
}

func (v *VoidRift) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if v.Finished {
		return
	}
	progress := v.age / v.Duration
	var sizeT float64
	if progress < 0.4 {
		sizeT = easeOutQuad(progress / 0.4)
	} else {
		sizeT = 1.0 - easeOutQuad((progress-0.4)/0.6)
	}
	radius := float32(6+22*sizeT) * float32(camScale)

	// +coords.BodyDX/BodyDY aligns the rift on the target's visual chest
	// rather than its raw anchor point — the same chest-alignment nudge every
	// other bespoke CR1 artifact visual applies in its own project() closure
	// (GroundSlam, RootBind, ReaperExecute, TauntPulse, ArcaneSurgeNuke,
	// BloodPriceStrike in artifacts.spells.go; ShadowStrike and the original
	// BlinkEffect via an older ad hoc "+1"). v.X/v.Y is already the target's
	// BodyCenter() position (see NewVoidRift), so without this the rift reads
	// as spawning behind/above the sprite instead of on it.
	sx, sy := isoToScreenFloat(v.X+coords.BodyDX, v.Y+coords.BodyDY, tileSize)
	ssx := float32((sx-camX)*camScale + cx)
	ssy := float32((sy+camY)*camScale + cy)

	coreCol := [3]float32{0.35, 0.08, 0.5}
	rimCol := [3]float32{0.75, 0.55, 1.0}
	drawRift(screen, ssx, ssy, radius, coreCol, rimCol, v.age)

	glowA := float32(1)
	if progress > 0.7 {
		glowA = float32(1 - (progress-0.7)/0.3)
	}
	drawGlow(screen, ssx, ssy, radius*0.5, [3]float32{0.55, 0.25, 0.85}, 0.8*glowA, v.age)
}

func (v *VoidRift) IsFinished() bool { return v.Finished }
