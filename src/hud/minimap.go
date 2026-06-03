package hud

import (
	"image/color"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	minimapTileSize = 4 // pixels per tile on minimap
	minimapPadding  = 8 // screen edge padding
)

// Minimap renders a small room-level overview in a screen corner.
type Minimap struct {
	Visible bool
}

// ToggleOnKeyPress checks for M key and toggles visibility.
func (m *Minimap) ToggleOnKeyPress() {
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		m.Visible = !m.Visible
	}
}

// Draw renders the minimap if visible.
// rooms: all rooms on the current level.
// seenTiles: [y][x] bool grid from Game.SeenTiles.
// playerTX, playerTY: player tile position.
// exitTX, exitTY: exit portal tile position (-1,-1 if unknown).
// screenW, screenH: screen dimensions for corner placement.
func (m *Minimap) Draw(
	screen *ebiten.Image,
	rooms []levels.Room,
	seenTiles [][]bool,
	playerTX, playerTY int,
	exitTX, exitTY int,
	screenW, screenH int,
) {
	if !m.Visible || len(rooms) == 0 {
		return
	}

	// Find level bounds from rooms.
	minX, minY, maxX, maxY := 9999, 9999, 0, 0
	for _, r := range rooms {
		if r.X < minX {
			minX = r.X
		}
		if r.Y < minY {
			minY = r.Y
		}
		if r.X+r.W > maxX {
			maxX = r.X + r.W
		}
		if r.Y+r.H > maxY {
			maxY = r.Y + r.H
		}
	}

	mapW := (maxX - minX) * minimapTileSize
	mapH := (maxY - minY) * minimapTileSize

	// Clamp to reasonable size.
	const maxMapPx = 160
	scaleX, scaleY := 1.0, 1.0
	if mapW > maxMapPx {
		scaleX = float64(maxMapPx) / float64(mapW)
	}
	if mapH > maxMapPx {
		scaleY = float64(maxMapPx) / float64(mapH)
	}
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	tilePixels := int(float64(minimapTileSize) * scale)
	if tilePixels < 1 {
		tilePixels = 1
	}

	renderW := (maxX - minX) * tilePixels
	renderH := (maxY - minY) * tilePixels

	// Position: bottom-right corner.
	originX := screenW - renderW - minimapPadding
	originY := screenH - renderH - minimapPadding

	// Background.
	bg := ebiten.NewImage(renderW+4, renderH+4)
	bg.Fill(color.NRGBA{0, 0, 0, 160})
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(float64(originX-2), float64(originY-2))
	screen.DrawImage(bg, bgOp)

	// Draw explored rooms.
	for _, r := range rooms {
		// Check if any tile in this room has been seen.
		seen := false
		for ty := r.Y; ty < r.Y+r.H && !seen; ty++ {
			for tx := r.X; tx < r.X+r.W && !seen; tx++ {
				if ty < len(seenTiles) && tx < len(seenTiles[ty]) && seenTiles[ty][tx] {
					seen = true
				}
			}
		}
		if !seen {
			continue
		}
		rx := originX + (r.X-minX)*tilePixels
		ry := originY + (r.Y-minY)*tilePixels
		rw := r.W * tilePixels
		rh := r.H * tilePixels
		if rw < 1 {
			rw = 1
		}
		if rh < 1 {
			rh = 1
		}
		roomImg := ebiten.NewImage(rw, rh)
		roomImg.Fill(color.NRGBA{120, 120, 140, 220})
		rop := &ebiten.DrawImageOptions{}
		rop.GeoM.Translate(float64(rx), float64(ry))
		screen.DrawImage(roomImg, rop)
	}

	// Exit marker (yellow dot).
	if exitTX >= 0 && exitTY >= 0 {
		ex := originX + (exitTX-minX)*tilePixels
		ey := originY + (exitTY-minY)*tilePixels
		dot := ebiten.NewImage(tilePixels+2, tilePixels+2)
		dot.Fill(color.NRGBA{255, 220, 0, 255})
		dop := &ebiten.DrawImageOptions{}
		dop.GeoM.Translate(float64(ex-1), float64(ey-1))
		screen.DrawImage(dot, dop)
	}

	// Player dot (white).
	px := originX + (playerTX-minX)*tilePixels
	py := originY + (playerTY-minY)*tilePixels
	playerDot := ebiten.NewImage(tilePixels+2, tilePixels+2)
	playerDot.Fill(color.NRGBA{255, 255, 255, 255})
	pop := &ebiten.DrawImageOptions{}
	pop.GeoM.Translate(float64(px-1), float64(py-1))
	screen.DrawImage(playerDot, pop)
}
