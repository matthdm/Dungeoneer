package game

import (
	"dungeoneer/coords"

	"github.com/hajimehoshi/ebiten/v2"
)

// This file is the game package's only bridge between world space and the
// screen for points (as opposed to sprite cells). See package coords for the
// model: one world frame, one ground projection, height applied in screen
// space. Nothing outside this file and coords should add a constant to make a
// position "line up".

// cursorGround returns the world ground point under the mouse cursor.
func (g *Game) cursorGround() coords.WorldPos {
	mx, my := ebiten.CursorPosition()
	return g.screenToGround(float64(mx), float64(my))
}

// cursorAim returns the world point to aim body-height effects at so that
// they pass through the cursor: projectiles, beams, melee arcs, and "which
// monster did I click" tests (the player clicks a body, not the floor
// under it).
func (g *Game) cursorAim() coords.WorldPos {
	return g.cursorGround().AtBodyHeight()
}

// screenToGround converts a window pixel to the world ground point drawn there.
func (g *Game) screenToGround(sx, sy float64) coords.WorldPos {
	isoX := (sx-float64(g.w/2))/g.camScale + g.camX
	isoY := (sy-float64(g.h/2))/g.camScale - g.camY
	return coords.IsoToGround(isoX, isoY, g.currentLevel.TileSize)
}

// groundToScreen returns where a world ground point is drawn, using the
// camera transform parameters of the current draw pass.
func (g *Game) groundToScreen(x, y, scale, cx, cy float64) (float64, float64) {
	isoX, isoY := coords.GroundToIso(x, y, g.currentLevel.TileSize)
	return (isoX-g.camX)*scale + cx, (isoY+g.camY)*scale + cy
}

// bodyToScreen is groundToScreen lifted to body height: where the middle of a
// body standing on world point (x, y) is drawn.
func (g *Game) bodyToScreen(x, y, scale, cx, cy float64) (float64, float64) {
	sx, sy := g.groundToScreen(x, y, scale, cx, cy)
	return sx, sy - coords.BodyHeightPx(g.currentLevel.TileSize)*scale
}

// bodyToWindow is bodyToScreen for code that runs outside a draw pass
// (particle emitters fired from Update): it uses the live camera zoom and the
// window centre.
func (g *Game) bodyToWindow(x, y float64) (float64, float64) {
	return g.bodyToScreen(x, y, g.camScale, float64(g.w/2), float64(g.h/2))
}
