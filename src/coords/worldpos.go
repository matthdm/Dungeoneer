// Package coords provides the canonical coordinate types and conversions for
// Dungeoneer's cartesian/isometric world.
//
// # Coordinate System
//
// World space is a 2D cartesian grid. A WorldPos{X, Y} means X tiles east and
// Y tiles south of the origin. Positions are float64 for smooth interpolation.
//
// Isometric screen space is the 2D projection rendered on screen:
//
//	isoX = (X - Y) * (TileSize / 2)
//	isoY = (X + Y) * (TileSize / 4)
//
// Entity positions are anchored at the TOP-LEFT vertex of the isometric
// diamond (where isoToScreenFloat/cartesianToIso place them). The visible
// body center is BodyDX units east and BodyDY units south of that anchor.
//
// # Golden rule
//
// Always pass WorldPos.BodyCenter() into hit detection, spell origins, and
// effect anchoring. Never pass raw X/Y for anything combat-related.
package coords

import "math"

const (
	// BodyDX is the horizontal cartesian offset from the tile anchor to the
	// entity's visual body center. One unit east places it at the mid-point
	// of the 64 px isometric diamond.
	BodyDX = 1.0

	// BodyDY is the vertical (depth) cartesian offset from the tile anchor
	// (feet) to the entity's body center.
	BodyDY = 0.30

	// SpriteVerticalShift is the number of isometric-space pixels sprites are
	// shifted upward during rendering. Apply as a negative Y delta in every
	// entity draw transform so that sprite feet land on the tile anchor.
	SpriteVerticalShift = 1.0

	// MeleeRange is the maximum BodyCenter-to-BodyCenter distance (tile units)
	// for a melee attack to connect. 1.0 = tiles are touching; 1.5 gives
	// leeway during movement interpolation and matches the old IsAdjacent feel.
	MeleeRange = 1.5
)

// WorldPos is a position in cartesian world space measured in tile units.
// It is the single source of truth for where an entity is. Integer tile
// coordinates (TileX, TileY) are derived from it, not stored alongside it.
type WorldPos struct {
	X, Y float64
}

// TileX returns the integer tile column containing this position (floor of X).
func (w WorldPos) TileX() int { return int(math.Floor(w.X)) }

// TileY returns the integer tile row containing this position (floor of Y).
func (w WorldPos) TileY() int { return int(math.Floor(w.Y)) }

// BodyCenter returns the combat/hit body center of an entity at this position.
// Use this for all distance checks, hit detection, spell origins, health-bar
// anchoring, and effect placement — never raw X/Y for anything combat-related.
func (w WorldPos) BodyCenter() WorldPos {
	return WorldPos{X: w.X + BodyDX, Y: w.Y + BodyDY}
}

// ToIso projects the world position into isometric screen-space coordinates
// (in pixels at tileSize scale, before the camera transform is applied).
func (w WorldPos) ToIso(tileSize int) (float64, float64) {
	h := float64(tileSize / 2)
	q := float64(tileSize / 4)
	return (w.X - w.Y) * h, (w.X + w.Y) * q
}

// FromIso unprojects isometric screen-space coordinates back into world space.
// x and y must already be in world-render coordinates: camera, zoom, and screen
// centering are caller responsibilities.
func FromIso(x, y float64, tileSize int) WorldPos {
	h := float64(tileSize / 2)
	q := float64(tileSize / 4)
	return WorldPos{
		X: (x/h + y/q) / 2,
		Y: (y/q - x/h) / 2,
	}
}

// DistTo returns the euclidean distance from w to other in world space.
func (w WorldPos) DistTo(other WorldPos) float64 {
	return math.Hypot(w.X-other.X, w.Y-other.Y)
}

// TileCenterIso returns the isometric screen coordinates of the visual center
// of the tile diamond. Use this to anchor health bars, labels, hit markers,
// and other effects so they appear centered on the entity's tile.
func (w WorldPos) TileCenterIso(tileSize int) (float64, float64) {
	ix, iy := w.ToIso(tileSize)
	ts := float64(tileSize)
	return ix + ts/2, iy + ts/4
}

// RenderIso returns the isometric screen coordinates for rendering an entity
// sprite at this world position. It bakes in the canonical SpriteVerticalShift
// and the per-frame bob offset, so callers can go straight to
// GeoM.Translate(sx, sy) without a separate vertical-nudge step.
func (w WorldPos) RenderIso(tileSize int, bob float64) (float64, float64) {
	ix, iy := w.ToIso(tileSize)
	return ix, iy - SpriteVerticalShift + bob
}

// ToIso is a package-level convenience for converting raw float64 world
// coordinates to isometric screen coordinates. Prefer WorldPos.ToIso when
// you already have a WorldPos.
func ToIso(x, y float64, tileSize int) (float64, float64) {
	return WorldPos{X: x, Y: y}.ToIso(tileSize)
}
