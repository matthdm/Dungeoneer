// Package collision handles collision detection and resolution.
package collision

import (
	"dungeoneer/levels"
	"math"
)

type Box struct {
	X, Y          float64
	Width, Height float64
}

// SpriteAnchor describes the offset from a logical collision box center to the
// sprite's visual feet. Collision checks use this to keep wall contact aligned
// with rendered character sprites without scattering magic offsets.
type SpriteAnchor struct {
	OffsetX float64
	OffsetY float64
}

func (a SpriteAnchor) CollisionOffset() (x, y float64) {
	return a.OffsetX, a.OffsetY
}

// PlayerSpriteAnchor preserves the current player wall-contact feel while
// making the visual collision anchor explicit.
var PlayerSpriteAnchor = SpriteAnchor{OffsetX: 0.21, OffsetY: 0.75}

// PredictAndClip moves the given Box by dx/dy while preventing it from entering
// any unwalkable tile. Movement is stepped in small increments to avoid
// tunneling. The returned Box is the final clipped position. The bools indicate
// whether movement along each axis was blocked.
func PredictAndClip(level *levels.Level, box Box, dx, dy float64) (Box, bool, bool) {
	const stepSize = 0.05 // tile units per sweep

	moveX := dx
	moveY := dy
	collidedX := false
	collidedY := false

	// Resolve X axis first
	for math.Abs(moveX) > 0 {
		step := math.Copysign(math.Min(math.Abs(moveX), stepSize), moveX)
		next := box
		next.X += step
		if CollidesWithMap(level, next) {
			collidedX = true
			break
		}
		box = next
		moveX -= step
	}

	// Then resolve Y axis
	for math.Abs(moveY) > 0 {
		step := math.Copysign(math.Min(math.Abs(moveY), stepSize), moveY)
		next := box
		next.Y += step
		if CollidesWithMap(level, next) {
			collidedY = true
			break
		}
		box = next
		moveY -= step
	}

	return box, collidedX, collidedY
}

func CollidesWithMap(level *levels.Level, box Box) bool {
	return CollidesWithMapWithAnchor(level, box, PlayerSpriteAnchor)
}

// CollidesWithMapWithAnchor checks if the given box overlaps with unwalkable
// tiles after applying the supplied sprite anchor.
func CollidesWithMapWithAnchor(level *levels.Level, box Box, anchor SpriteAnchor) bool {
	offsetBox := box
	offsetX, offsetY := anchor.CollisionOffset()
	offsetBox.X += offsetX
	offsetBox.Y += offsetY

	tileLeft := int(math.Floor(offsetBox.X - offsetBox.Width/2))
	tileTop := int(math.Floor(offsetBox.Y - offsetBox.Height/2))
	tileRight := int(math.Floor(offsetBox.X + offsetBox.Width/2))
	tileBottom := int(math.Floor(offsetBox.Y + offsetBox.Height/2))

	for y := tileTop; y <= tileBottom; y++ {
		for x := tileLeft; x <= tileRight; x++ {
			if x < 0 || y < 0 || x >= level.W || y >= level.H {
				return true
			}
			if !level.IsWalkable(x, y) {
				return true
			}
		}
	}
	return false
}
