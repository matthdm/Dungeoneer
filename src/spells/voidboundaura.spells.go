package spells

import (
	"dungeoneer/coords"

	"github.com/hajimehoshi/ebiten/v2"
)

// VoidboundAura is voidbound_pendant's passive visual: a slow-pulsing violet
// glow clinging to the caster while the pendant is active. Unlike a Spell
// (spawned once, expires, appended to ActiveSpells), this is drawn every
// frame the pendant is equipped — the caller owns gating and lifetime, not
// this struct. age accumulates continuously to drive the pulse animation.
type VoidboundAura struct {
	age float64
}

// Update advances the aura's animation clock. Safe to call every frame
// regardless of whether the pendant is currently equipped — Draw is what's
// actually gated by the caller.
func (a *VoidboundAura) Update(dt float64) {
	a.age += dt
}

// Draw renders the aura centered at (worldX, worldY), which the caller must
// already have body-center-offset (e.g. via Player.BodyX()/BodyY(), which
// applies coords.WorldPos.BodyCenter() internally) — never raw tile
// coordinates, per the project's coordinate invariant.
func (a *VoidboundAura) Draw(screen *ebiten.Image, tileSize int, worldX, worldY float64, camX, camY, camScale, cx, cy float64) {
	// +coords.BodyDX/BodyDY re-applies the same chest-alignment nudge every
	// other bespoke CR1 visual applies on top of an already-body-centered
	// input (GroundSlam does the identical thing with player.BodyX()/BodyY());
	// worldX/worldY here is already BodyCenter() (see Draw's own doc comment),
	// so without it the aura clings to the anchor point rather than the chest.
	sx, sy := isoToScreenFloat(worldX+coords.BodyDX, worldY+coords.BodyDY, tileSize)
	ssx := float32((sx-camX)*camScale + cx)
	ssy := float32((sy+camY)*camScale + cy)
	radius := float32(10) * float32(camScale)
	drawGlow(screen, ssx, ssy, radius, [3]float32{0.5, 0.25, 0.85}, 0.55, a.age)
}
