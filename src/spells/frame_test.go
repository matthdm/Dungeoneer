package spells

import (
	"math"
	"testing"

	"dungeoneer/coords"
)

const testTileSize = 64

// A body-height effect located at an entity's logical position (its
// BodyCenter()) must be drawn on the middle of that entity's sprite, and a
// ground effect there must be drawn on the middle of the floor diamond the
// entity stands on. Sprite cells are 64×64 and blitted at ToIso(pos); the
// floor diamond is the bottom half of the cell.
func TestEffectAtBodyCenterIsDrawnOnTheSprite(t *testing.T) {
	for _, pos := range []coords.WorldPos{{X: 5, Y: 9}, {X: 12.25, Y: 3.5}, {X: 0, Y: 0}} {
		cellX, cellY := pos.ToIso(testTileSize)
		body := pos.BodyCenter()

		bx, by := bodyIso(body.X, body.Y, testTileSize)
		if bx-cellX != 32 || by-cellY != 32 {
			t.Errorf("pos %v: body effect at cell pixel (%g,%g), want sprite centre (32,32)", pos, bx-cellX, by-cellY)
		}
		gx, gy := groundIso(body.X, body.Y, testTileSize)
		if gx-cellX != 32 || gy-cellY != 48 {
			t.Errorf("pos %v: ground effect at cell pixel (%g,%g), want diamond centre (32,48)", pos, gx-cellX, gy-cellY)
		}
	}
}

// The slash arc that is drawn (ArcPoints) is the boundary of the area that is
// tested (IsInArc): every drawn arc point lies exactly Radius from the origin
// and inside the tested wedge, in the same world frame as the targets.
func TestSlashArcDrawnShapeIsTheHitShape(t *testing.T) {
	origin := coords.WorldPos{X: 7, Y: 7}.BodyCenter()
	for _, dir := range []float64{0, math.Pi / 3, math.Pi, -math.Pi / 2, 2.5} {
		s := NewSlashArc(SpellInfo{}, origin.X, origin.Y, dir, 0)
		for i, p := range s.ArcPoints {
			if d := math.Hypot(p.X-origin.X, p.Y-origin.Y); math.Abs(d-s.Radius) > 1e-9 {
				t.Fatalf("dir %g: arc point %d is %g from origin, want radius %g", dir, i, d, s.Radius)
			}
			// Pull the point a hair inside the boundary; it must register as hit.
			in := Point{X: origin.X + (p.X-origin.X)*0.999, Y: origin.Y + (p.Y-origin.Y)*0.999}
			// The two end points sit exactly on the angular limit; skip them
			// to stay clear of floating-point ties.
			if i == 0 || i == len(s.ArcPoints)-1 {
				continue
			}
			if !s.IsInArc(in.X, in.Y) {
				t.Fatalf("dir %g: point just inside drawn arc segment %d is not in the hit arc", dir, i)
			}
		}
	}
}

// A target standing at the same distance is hit or missed identically no
// matter which side of the attacker it is on, as long as the swing is aimed
// at it.
func TestSlashReachIsTheSameInEveryDirection(t *testing.T) {
	attacker := coords.WorldPos{X: 10, Y: 10}
	origin := attacker.BodyCenter()
	radius := SlashComboHits[0].Radius
	for i := 0; i < 16; i++ {
		ang := float64(i) / 16 * 2 * math.Pi
		for _, tc := range []struct {
			dist float64
			want bool
		}{{radius * 0.9, true}, {radius * 1.1, false}} {
			targetPos := coords.WorldPos{
				X: attacker.X + math.Cos(ang)*tc.dist,
				Y: attacker.Y + math.Sin(ang)*tc.dist,
			}
			tb := targetPos.BodyCenter()
			s := NewSlashArc(SpellInfo{}, origin.X, origin.Y, math.Atan2(tb.Y-origin.Y, tb.X-origin.X), 0)
			if got := s.IsInArc(tb.X, tb.Y); got != tc.want {
				t.Errorf("angle %d dist %.2f: IsInArc = %v, want %v", i, tc.dist, got, tc.want)
			}
		}
	}
}
