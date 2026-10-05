package coords

import (
	"math"
	"testing"
)

// The floor diamond in the tile art occupies the bottom half of the 64×64
// cell (measured from spritesheet cell (10,4)): top vertex at pixel (32,32),
// centre at (32,48), bottom vertex at (32,64), left/right vertices at (0,48)
// and (64,48). The ground projection must reproduce exactly those pixels for
// the four corners and the centre of the tile, relative to where the tile's
// sprite cell is blitted.
func TestGroundProjectionMatchesTileArt(t *testing.T) {
	const ts = 64
	for _, tile := range [][2]int{{0, 0}, {3, 7}, {12, 4}} {
		tx, ty := float64(tile[0]), float64(tile[1])
		cellX, cellY := ToIso(tx, ty, ts)
		cases := []struct {
			name       string
			wx, wy     float64
			pixX, pixY float64
		}{
			{"top vertex", tx, ty, 32, 32},
			{"right vertex", tx + 1, ty, 64, 48},
			{"bottom vertex", tx + 1, ty + 1, 32, 64},
			{"left vertex", tx, ty + 1, 0, 48},
			{"centre", tx + 0.5, ty + 0.5, 32, 48},
		}
		for _, c := range cases {
			gx, gy := GroundToIso(c.wx, c.wy, ts)
			if gx-cellX != c.pixX || gy-cellY != c.pixY {
				t.Errorf("tile %v %s: pixel in cell = (%g, %g), want (%g, %g)",
					tile, c.name, gx-cellX, gy-cellY, c.pixX, c.pixY)
			}
		}
	}
}

func TestIsoToGroundInvertsGroundToIso(t *testing.T) {
	for _, want := range []WorldPos{{0, 0}, {2, 1}, {2.5, 1.5}, {12.25, 4.75}, {-3.5, 9}} {
		ix, iy := want.GroundIso(64)
		got := IsoToGround(ix, iy, 64)
		if math.Abs(got.X-want.X) > 1e-9 || math.Abs(got.Y-want.Y) > 1e-9 {
			t.Errorf("IsoToGround(GroundToIso(%v)) = %v", want, got)
		}
	}
}

// Every pixel inside a tile's drawn floor diamond must pick that tile.
func TestCursorPickHitsTheDrawnDiamond(t *testing.T) {
	const ts = 64
	tx, ty := 5, 9
	cellX, cellY := ToIso(float64(tx), float64(ty), ts)
	// Points inside the diamond (centre (32,48), half-extents 32×16).
	for _, p := range [][2]float64{{32, 48}, {32, 34}, {32, 62}, {4, 48}, {60, 48}, {18, 42}, {46, 54}} {
		got := IsoToGround(cellX+p[0], cellY+p[1], ts)
		if got.TileX() != tx || got.TileY() != ty {
			t.Errorf("pixel %v in cell picked tile (%d,%d), want (%d,%d)", p, got.TileX(), got.TileY(), tx, ty)
		}
	}
	// Points in the empty top half of the cell belong to other tiles.
	for _, p := range [][2]float64{{32, 16}, {8, 8}, {56, 8}} {
		got := IsoToGround(cellX+p[0], cellY+p[1], ts)
		if got.TileX() == tx && got.TileY() == ty {
			t.Errorf("pixel %v is outside the diamond but picked the tile", p)
		}
	}
}

// An entity standing on a tile has its feet on the centre of that tile's
// diamond, and its body (chest) at the centre of its own sprite cell.
func TestEntityBodyGeometry(t *testing.T) {
	const ts = 64
	pos := WorldPos{X: 5, Y: 9} // stored position of an entity on tile (5,9)
	body := pos.BodyCenter()
	if body != TileCenter(5, 9) {
		t.Fatalf("BodyCenter() = %v, want tile centre %v", body, TileCenter(5, 9))
	}
	if body.TileX() != 5 || body.TileY() != 9 {
		t.Fatalf("BodyCenter() lies on tile (%d,%d), want (5,9)", body.TileX(), body.TileY())
	}
	cellX, cellY := pos.ToIso(ts)
	gx, gy := body.GroundIso(ts)
	if gx-cellX != 32 || gy-cellY != 48 {
		t.Errorf("feet pixel in sprite cell = (%g,%g), want (32,48)", gx-cellX, gy-cellY)
	}
	if cy := gy - BodyHeightPx(ts) - cellY; cy != 32 {
		t.Errorf("chest pixel Y in sprite cell = %g, want 32", cy)
	}
}

// A body-height effect aimed with AtBodyHeight is drawn exactly on the cursor.
func TestAtBodyHeightPassesThroughCursor(t *testing.T) {
	const ts = 64
	cursorX, cursorY := 173.0, 911.0
	aim := IsoToGround(cursorX, cursorY, ts).AtBodyHeight()
	gx, gy := aim.GroundIso(ts)
	if math.Abs(gx-cursorX) > 1e-9 || math.Abs(gy-BodyHeightPx(ts)-cursorY) > 1e-9 {
		t.Errorf("lifted aim point drawn at (%g,%g), want cursor (%g,%g)", gx, gy-BodyHeightPx(ts), cursorX, cursorY)
	}
}
