package levels

import (
	"dungeoneer/tiles"
	"testing"
)

func countDashRuns(l *Level, minLen int) (int, int) {
	horiz, vert := 0, 0
	for y := 0; y < l.H; y++ {
		run := 0
		for x := 0; x < l.W; x++ {
			if l.Tiles[y][x].HasTag(tiles.TagDashLane) {
				run++
			} else {
				if run >= minLen {
					horiz++
				}
				run = 0
			}
		}
		if run >= minLen {
			horiz++
		}
	}
	for x := 0; x < l.W; x++ {
		run := 0
		for y := 0; y < l.H; y++ {
			if l.Tiles[y][x].HasTag(tiles.TagDashLane) {
				run++
			} else {
				if run >= minLen {
					vert++
				}
				run = 0
			}
		}
		if run >= minLen {
			vert++
		}
	}
	return horiz, vert
}

func TestGenerate64x64Sanity(t *testing.T) {
	for seed := 1; seed <= 5; seed++ {
		p := GenParams{Seed: int64(seed)}
		l := Generate64x64(p)
		cov := coverage(l)
		if cov < 0.42 || cov > 0.55 {
			t.Fatalf("coverage out of range: %f", cov)
		}
		h, v := countDashRuns(l, p.DashLaneMinLen)
		if h < 3 || v < 3 {
			t.Fatalf("insufficient dash lanes: h=%d v=%d", h, v)
		}
		if p.UseMainPath && lastMainPathLen < 3 {
			t.Fatalf("main path too short: %d", lastMainPathLen)
		}
	}
}
