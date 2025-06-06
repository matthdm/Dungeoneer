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

func PredictAndClip(level *levels.Level, box Box, dx, dy float64) (Box, bool) {
	const stepSize = 0.1 // tile units
	steps := int(math.Ceil(math.Max(math.Abs(dx), math.Abs(dy)) / stepSize))

	if steps == 0 {
		steps = 1
	}

	stepX := dx / float64(steps)
	stepY := dy / float64(steps)

	curBox := box
	collided := false

	for i := 0; i < steps; i++ {
		nextBox := curBox
		// Step X axis first
		nextBox.X += stepX
		if CollidesWithMap(level, nextBox) {
			stepX = 0
			collided = true
		} else {
			curBox.X = nextBox.X
		}

		// Step Y axis next
		nextBox = curBox
		nextBox.Y += stepY
		if CollidesWithMap(level, nextBox) {
			stepY = 0
			collided = true
		} else {
			curBox.Y = nextBox.Y
		}
	}

	return curBox, collided
}

// CollidesWithMap checks if the given box overlaps with unwalkable tiles.
func CollidesWithMap(level *levels.Level, box Box) bool {
	tileLeft := int(math.Floor(box.X - box.Width/2))
	tileTop := int(math.Floor(box.Y - box.Height/2))
	tileRight := int(math.Floor(box.X + box.Width/2))
	tileBottom := int(math.Floor(box.Y + box.Height/2))

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
