package levels

import (
	"dungeoneer/tiles"
	"testing"
)

func TestThroats(t *testing.T) {
	l := NewEmptyLevel(20, 20)
	
	// Create a corridor
	for x := 2; x < 18; x++ {
		l.Tiles[10][x] = &tiles.Tile{IsWalkable: true}
	}
	// Create a branch
	for y := 5; y < 15; y++ {
		l.Tiles[y][10] = &tiles.Tile{IsWalkable: true}
	}
	

	BuildThroatDebug(l, 10, 2)
	// We just want to execute the logic to increase coverage.
	
	// Try it with rooms as well to cover throat regions
	r1 := Room{X: 1, Y: 1, W: 5, H: 5}
	r2 := Room{X: 12, Y: 1, W: 5, H: 5}
	l.Rooms = []Room{r1, r2}
	
	for y := 1; y < 6; y++ {
		for x := 1; x < 6; x++ {
			l.Tiles[y][x] = &tiles.Tile{IsWalkable: true}
		}
	}
	for y := 1; y < 6; y++ {
		for x := 12; x < 17; x++ {
			l.Tiles[y][x] = &tiles.Tile{IsWalkable: true}
		}
	}
	
	BuildThroatDebug(l, 10, 2)
}
