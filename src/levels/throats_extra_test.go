package levels

import (
	"dungeoneer/tiles"
	"testing"
)

func TestBuildRegions(t *testing.T) {
	l := NewEmptyLevel(10, 10)
	
	// Region 1: linear corridor
	l.Tiles[1][1] = &tiles.Tile{IsWalkable: true}
	l.Tiles[1][2] = &tiles.Tile{IsWalkable: true}
	l.Tiles[1][3] = &tiles.Tile{IsWalkable: true}
	
	// Region 2: non-linear (cross shape)
	l.Tiles[5][5] = &tiles.Tile{IsWalkable: true} // center
	l.Tiles[4][5] = &tiles.Tile{IsWalkable: true} // up
	l.Tiles[6][5] = &tiles.Tile{IsWalkable: true} // down
	l.Tiles[5][4] = &tiles.Tile{IsWalkable: true} // left
	l.Tiles[5][6] = &tiles.Tile{IsWalkable: true} // right
	
	regionIDs := make([][]int, l.H)
	for y := range regionIDs {
		regionIDs[y] = make([]int, l.W)
	}
	
	sizes, allLinear := buildRegions(l, regionIDs)
	
	if len(sizes) != 2 {
		t.Errorf("expected 2 regions")
	}
	
	r1 := regionIDs[1][1]
	if sizes[r1] != 3 {
		t.Errorf("expected region 1 size 3")
	}
	if !allLinear[r1] {
		t.Errorf("expected region 1 to be linear")
	}
	
	r2 := regionIDs[5][5]
	if sizes[r2] != 5 {
		t.Errorf("expected region 2 size 5")
	}
	if allLinear[r2] {
		t.Errorf("expected region 2 to be non-linear")
	}
}
