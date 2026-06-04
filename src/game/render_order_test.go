package game

import (
	"testing"
)

func TestRenderOrder(t *testing.T) {
	rs := []Renderable{
		{TileX: 5, TileY: 5, DepthWeight: 0},
		{TileX: 1, TileY: 1, DepthWeight: 0},
		{TileX: 3, TileY: 3, DepthWeight: 0},
		{TileX: 3, TileY: 3, DepthWeight: 0.5},
	}

	sortRenderables(rs)

	if rs[0].TileX != 1 {
		t.Errorf("expected front to be (1,1)")
	}
	if rs[1].TileX != 3 || rs[1].DepthWeight != 0 {
		t.Errorf("expected second to be (3,3,0)")
	}
	if rs[2].TileX != 3 || rs[2].DepthWeight != 0.5 {
		t.Errorf("expected third to be (3,3,0.5)")
	}
	if rs[3].TileX != 5 {
		t.Errorf("expected back to be (5,5)")
	}
}
