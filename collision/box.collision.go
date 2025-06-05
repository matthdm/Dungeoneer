// Package collision handles collision detection and resolution.
package collision

import (
	"dungeoneer/constants"
	"dungeoneer/levels"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Box struct {
	X, Y          float64
	Width, Height float64
}

type SweepDebug struct {
	Start         Box
	End           Box
	CollidedTiles [][2]int
}

// Resolve stops movement if any tile along the path is unwalkable
func Resolve(level *levels.Level, box Box, dx, dy float64, maxDist float64) (Box, bool, SweepDebug) {
	start := box
	end := box
	end.X += dx
	end.Y += dy

	// Attempt move
	testBox := end
	if CollidesWithMap(level, testBox) {
		// Attempt X only
		testBox = start
		testBox.X += dx
		if !CollidesWithMap(level, testBox) {
			end.X += dx
			end.Y = start.Y // cancel Y move
		} else {
			// Try Y only
			testBox = start
			testBox.Y += dy
			if !CollidesWithMap(level, testBox) {
				end.X = start.X // cancel X move
				end.Y += dy
			} else {
				// Fully blocked
				end = start
			}
		}
	}

	sweep := SweepDebug{
		Start: start,
		End:   end,
	}
	hit := (end.X != start.X || end.Y != start.Y) && CollidesWithMap(level, end)

	return end, hit, sweep
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

func DebugDrawSweep(screen *ebiten.Image, sweep SweepDebug, camX, camY, camScale, cx, cy float64) {
	// Align sweep with bottom-center (feet)
	startY := sweep.Start.Y - (sweep.Start.Height / 2)
	endY := sweep.End.Y - (sweep.End.Height / 2)

	x1, y1 := isoToScreenFloat(sweep.Start.X, startY, constants.DefaultTileSize)
	x2, y2 := isoToScreenFloat(sweep.End.X, endY, constants.DefaultTileSize)

	x1 = (x1-camX)*camScale + cx
	y1 = (y1+camY)*camScale + cy
	x2 = (x2-camX)*camScale + cx
	y2 = (y2+camY)*camScale + cy

	ebitenutil.DrawLine(screen, x1, y1, x2, y2, color.RGBA{255, 255, 0, 255})

	for _, tile := range sweep.CollidedTiles {
		tx, ty := isoToScreenFloat(float64(tile[0]), float64(tile[1]), constants.DefaultTileSize)
		tx = (tx-camX)*camScale + cx
		ty = (ty+camY)*camScale + cy

		size := float64(constants.DefaultTileSize) * camScale
		rect := ebiten.NewImage(int(size), int(size))
		rect.Fill(color.RGBA{255, 0, 0, 128})

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(tx, ty)
		screen.DrawImage(rect, op)
	}
}

func isoToScreenFloat(x, y float64, tileSize int) (float64, float64) {
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}
