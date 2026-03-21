package entities

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ExitEntity represents a floor exit (portal/stairwell) that the player
// interacts with to advance to the next dungeon floor.
type ExitEntity struct {
	TileX, TileY int
	Sprite       *ebiten.Image
	SpriteID     string
	BobOffset    float64
	TickCount    int
}

// NewExitEntity creates an exit entity at the given tile position.
func NewExitEntity(x, y int, sprite *ebiten.Image, spriteID string) *ExitEntity {
	return &ExitEntity{
		TileX:    x,
		TileY:    y,
		Sprite:   sprite,
		SpriteID: spriteID,
	}
}

// Update advances the exit entity's animation state.
func (e *ExitEntity) Update() {
	e.TickCount++
	e.BobOffset = math.Sin(float64(e.TickCount)/30.0) * 2.0
}

// IsPlayerNear returns true if the player is within interaction range.
func (e *ExitEntity) IsPlayerNear(px, py int) bool {
	dx := e.TileX - px
	dy := e.TileY - py
	return dx*dx+dy*dy <= 2 // within ~1.4 tiles
}

// PulseAlpha returns a 0.6–1.0 pulsing alpha value for rendering.
func (e *ExitEntity) PulseAlpha() float32 {
	return 0.8 + 0.2*float32(math.Sin(float64(e.TickCount)/20.0))
}
