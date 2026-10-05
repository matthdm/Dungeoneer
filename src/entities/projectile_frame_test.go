package entities

import (
	"fmt"
	"math"
	"testing"

	"dungeoneer/coords"
)

// A projectile aimed from one body at another must reach it from every
// direction. This pins the single-world-frame rule: the launch point, the
// aim point and the point HitsPlayer tests are all BodyCenter() values, so
// there is no direction in which the shot passes "beside" the target.
//
// Before the unification the shot flew between stored positions while the
// hit test used an offset body centre, so it only connected from some angles.
func TestMonsterProjectileHitsTargetFromEveryDirection(t *testing.T) {
	playerPos := coords.WorldPos{X: 20, Y: 20}
	const dist = 5.0
	for i := 0; i < 16; i++ {
		ang := float64(i) / 16 * 2 * math.Pi
		t.Run(fmt.Sprintf("angle_%d", i), func(t *testing.T) {
			monsterPos := coords.WorldPos{
				X: playerPos.X + math.Cos(ang)*dist,
				Y: playerPos.Y + math.Sin(ang)*dist,
			}
			from, to := monsterPos.BodyCenter(), playerPos.BodyCenter()
			p := NewMonsterProjectile(from.X, from.Y, to.X, to.Y, 0.15, 1)

			hit := false
			for tick := 0; tick < p.MaxTicks; tick++ {
				p.X += p.DirX * p.Speed
				p.Y += p.DirY * p.Speed
				if p.HitsPlayer(playerPos) {
					hit = true
					break
				}
			}
			if !hit {
				t.Fatalf("projectile from %v never hit player at %v", monsterPos, playerPos)
			}
		})
	}
}

// Melee reach must not depend on which side the attacker stands on.
func TestBodyCenterDistanceIsSymmetric(t *testing.T) {
	a := coords.WorldPos{X: 10, Y: 10}
	for _, d := range [][2]float64{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}} {
		b := coords.WorldPos{X: a.X + d[0], Y: a.Y + d[1]}
		got := a.BodyCenter().DistTo(b.BodyCenter())
		want := math.Hypot(d[0], d[1])
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("offset %v: body distance = %g, want %g", d, got, want)
		}
	}
}
