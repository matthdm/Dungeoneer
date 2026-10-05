// Package coords provides the canonical coordinate types and conversions for
// Dungeoneer's cartesian/isometric world.
//
// # The one world frame
//
// There is exactly ONE logical coordinate frame: world space, measured in
// tiles on the ground plane. Tile (x, y) covers the square [x, x+1) × [y, y+1),
// so floor(X), floor(Y) of any world point is the tile it lies on. Tile
// indices, pathing, wall checks, cursor picks, spell targets, projectile
// positions and hit tests are all expressed in this frame and can be compared
// with each other directly. Nothing in game logic ever adds an offset.
//
// # The one projection
//
// A world ground point is projected to isometric pixels by GroundToIso and
// unprojected by IsoToGround. These two functions are the ONLY place the
// relationship between the logical grid and the artwork is encoded.
//
//	isoX = (X - Y) * (TileSize / 2) + TileSize/2
//	isoY = (X + Y) * (TileSize / 4) + TileSize/2
//
// The +TileSize/2 terms exist because tile art is a TileSize×TileSize cell
// whose floor diamond occupies the BOTTOM half of the cell: the diamond's top
// vertex (the world-space corner of the tile) sits at pixel (TileSize/2,
// TileSize/2) inside the cell. See TestGroundProjectionMatchesTileArt.
//
// # Sprites
//
// A tile-sized sprite cell for something anchored at world (x, y) is blitted
// with its top-left at ToIso(x, y). ToIso is therefore a *sprite cell origin*,
// not the screen position of the point (x, y). Use it only for blitting
// TileSize×TileSize cells (tiles, characters, chests). To find where a world
// point appears on screen, use GroundToIso.
//
// # Entities
//
// An entity's stored position (InterpX/InterpY, exposed as Pos()) is the
// origin corner of the tile-sized cell it occupies: an entity standing on tile
// (3, 7) has Pos() == {3, 7}. Its feet are in the middle of that cell.
// BodyCenter() returns that point — the entity's true location on the ground —
// and is what every distance check, hit test and spell origin must use.
//
// # Height
//
// Isometric screen-Y mixes depth and height, so a single 2D point cannot be
// both "on the floor" and "at chest level". Height is therefore never folded
// into world coordinates. Effects that should appear on a body (projectiles,
// hit markers, impact flashes) are positioned on the ground in world space and
// lifted in SCREEN space by BodyHeightPx at draw time.
//
// # Golden rule
//
// Logic: world space only, entities via BodyCenter(). Drawing: GroundToIso,
// optionally minus BodyHeightPx. Never add a constant to make something line
// up — if it does not line up, one side is not in world space.
package coords

import "math"

const (
	// SpriteVerticalShift is the number of isometric-space pixels sprites are
	// shifted upward during rendering. Apply as a negative Y delta in every
	// entity draw transform so that sprite feet land on the tile anchor.
	SpriteVerticalShift = 1.0

	// MeleeRange is the maximum BodyCenter-to-BodyCenter distance (tile units)
	// for a melee attack to connect. 1.0 = tiles are touching; 1.5 gives
	// leeway during movement interpolation and matches the old IsAdjacent feel.
	MeleeRange = 1.5

	// cellHalf is the offset from an entity's stored position (the origin
	// corner of the tile cell it occupies) to the middle of that cell, where
	// its feet are. This is geometry (half a tile), not a tuning value.
	cellHalf = 0.5
)

// WorldPos is a position in world space measured in tile units.
type WorldPos struct {
	X, Y float64
}

// TileX returns the integer tile column containing this position (floor of X).
func (w WorldPos) TileX() int { return int(math.Floor(w.X)) }

// TileY returns the integer tile row containing this position (floor of Y).
func (w WorldPos) TileY() int { return int(math.Floor(w.Y)) }

// BodyCenter converts an entity's stored position (cell origin) into the world
// point its body stands on: the middle of the cell. Use the result for all
// distance checks, hit detection, spell origins and effect placement.
//
// Call it only on entity positions (Pos()). Points that are already world
// ground points — cursor picks, spell targets, projectile positions — must be
// used as they are.
func (w WorldPos) BodyCenter() WorldPos {
	return WorldPos{X: w.X + cellHalf, Y: w.Y + cellHalf}
}

// EntityPos is the inverse of BodyCenter: the position to store on an entity
// so that its body stands on world point w. Use it when game logic picks a
// destination in world space (a blink target, a knock-back point) and has to
// write it back to InterpX/InterpY.
func (w WorldPos) EntityPos() WorldPos {
	return WorldPos{X: w.X - cellHalf, Y: w.Y - cellHalf}
}

// TileCenter returns the world point at the middle of tile (tx, ty).
func TileCenter(tx, ty int) WorldPos {
	return WorldPos{X: float64(tx) + cellHalf, Y: float64(ty) + cellHalf}
}

// DistTo returns the euclidean distance from w to other in world space.
func (w WorldPos) DistTo(other WorldPos) float64 {
	return math.Hypot(w.X-other.X, w.Y-other.Y)
}

// ── Projection ──────────────────────────────────────────────────────────────

// GroundToIso projects a world ground point to isometric pixels (before the
// camera transform). This is where the point visibly is on the floor.
func GroundToIso(x, y float64, tileSize int) (float64, float64) {
	ix, iy := ToIso(x, y, tileSize)
	o := float64(tileSize / 2)
	return ix + o, iy + o
}

// IsoToGround is the exact inverse of GroundToIso: it returns the world ground
// point shown at the given isometric pixel. x and y must already be in
// world-render coordinates (camera, zoom and screen centering removed).
func IsoToGround(x, y float64, tileSize int) WorldPos {
	o := float64(tileSize / 2)
	return FromIso(x-o, y-o, tileSize)
}

// GroundIso is the method form of GroundToIso.
func (w WorldPos) GroundIso(tileSize int) (float64, float64) {
	return GroundToIso(w.X, w.Y, tileSize)
}

// BodyHeightPx is how far above the ground, in isometric pixels, the middle of
// a character's body is drawn. Subtract it from the screen Y of a ground point
// to place an effect on a body instead of on the floor under it.
func BodyHeightPx(tileSize int) float64 { return float64(tileSize / 4) }

// AtBodyHeight returns the world point directly beneath a point that is seen
// through the given ground point at body height. Use it to turn a cursor pick
// into an aim point for body-height effects: a projectile drawn BodyHeightPx
// above AtBodyHeight(cursorGround) passes exactly through the cursor.
func (w WorldPos) AtBodyHeight() WorldPos {
	// Raising a point by BodyHeightPx (tileSize/4 px) on screen is the same
	// screen displacement as moving half a tile toward the viewer on each axis.
	return WorldPos{X: w.X + cellHalf, Y: w.Y + cellHalf}
}

// ── Sprite cells ────────────────────────────────────────────────────────────

// ToIso returns the isometric pixel at which to blit the top-left corner of a
// tileSize×tileSize sprite cell anchored at this position. It is NOT the
// on-screen location of the point itself — use GroundIso for that.
func (w WorldPos) ToIso(tileSize int) (float64, float64) {
	h := float64(tileSize / 2)
	q := float64(tileSize / 4)
	return (w.X - w.Y) * h, (w.X + w.Y) * q
}

// FromIso is the inverse of ToIso (sprite cell origin → anchor position).
// For cursor picking use IsoToGround instead.
func FromIso(x, y float64, tileSize int) WorldPos {
	h := float64(tileSize / 2)
	q := float64(tileSize / 4)
	return WorldPos{
		X: (x/h + y/q) / 2,
		Y: (y/q - x/h) / 2,
	}
}

// RenderIso returns the blit position for an entity sprite cell at this
// position. It bakes in the canonical SpriteVerticalShift and the per-frame
// bob offset, so callers can go straight to GeoM.Translate(sx, sy).
func (w WorldPos) RenderIso(tileSize int, bob float64) (float64, float64) {
	ix, iy := w.ToIso(tileSize)
	return ix, iy - SpriteVerticalShift + bob
}

// ToIso is the package-level form of WorldPos.ToIso (sprite cell origin).
func ToIso(x, y float64, tileSize int) (float64, float64) {
	return WorldPos{X: x, Y: y}.ToIso(tileSize)
}
