// Package collision handles collision detection and resolution.
package collision

import (
	"dungeoneer/constants"
	"dungeoneer/levels"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Box struct {
	X, Y          float64
	Width, Height float64
}

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

// CollidesWithMap checks if the given box overlaps with unwalkable tiles.
const YWallVisualOffset = .8  // shift player's collision box
const XWallVisualOffset = .21 // shift player's collision box

func CollidesWithMap(level *levels.Level, box Box) bool {
	// Apply Y-offset to collision check (helps align visual sprite base to tile logic)
	offsetBox := box
	offsetBox.Y += YWallVisualOffset
	offsetBox.X += XWallVisualOffset

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

func DebugDrawAABB(screen *ebiten.Image, box Box, camX, camY, camScale, cx, cy float64) {
	if box.Width <= 0 || box.Height <= 0 {
		return
	}

	// Convert box center to screen-space
	centerX, centerY := isoToScreenFloat(box.X, box.Y, constants.DefaultTileSize)

	// Convert size to pixels
	pixelW := box.Width * float64(constants.DefaultTileSize)
	pixelH := box.Height * float64(constants.DefaultTileSize)

	// Draw box centered on InterpX/Y
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(centerX, centerY)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	img := ebiten.NewImage(int(pixelW), int(pixelH))
	img.Fill(color.RGBA{255, 0, 0, 128})
	screen.DrawImage(img, op)
}

func isoToScreenFloat(x, y float64, tileSize int) (float64, float64) {
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}
