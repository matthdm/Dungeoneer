package levels

import (
	"dungeoneer/sprites"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func GenerateForgottenSanctuary(w, h int, ss *sprites.SpriteSheet, seed uint64) *Level {
	l := NewEmptyLevel(w, h)

	cx := float64(w)/2.0 - 0.5
	cy := float64(h)/2.0 - 0.5

	// 1. Blueprint the Floor (Greek Cross Shape)
	floorMap := make([][]bool, h)
	for y := 0; y < h; y++ {
		floorMap[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy

			// Screen-Vertical Arm (North-South)
			isVertArm := math.Abs(dx-dy) <= 4.5 && math.Abs(dx+dy) <= 13.5
			// Screen-Horizontal Arm (East-West)
			isHorizArm := math.Abs(dx+dy) <= 4.5 && math.Abs(dx-dy) <= 13.5
			// Central Altar Room
			isCenter := dx*dx+dy*dy <= 7.5*7.5

			if isVertArm || isHorizArm || isCenter {
				floorMap[y][x] = true
			}
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t := l.Tiles[y][x]
			t.IsWalkable = false
			t.ClearSprites()

			if floorMap[y][x] {
				t.IsWalkable = true
				t.AddSpriteByID("Floor", ss.Floor)

				dx := float64(x) - cx
				dy := float64(y) - cy

				// Holy Water Pools at the ends of the East/West Arms
				isEastWestEnd := math.Abs(dx+dy) <= 2.5 && math.Abs(dx-dy) >= 10.5 && math.Abs(dx-dy) <= 12.5
				if isEastWestEnd {
					t.IsWalkable = false
					t.ClearSprites()
					t.AddSpriteByID("EnchantedWater", ss.EnchantedWater)
				}

				// Pillars along the arms
				isArmEdge := (math.Abs(dx-dy) == 4 || math.Abs(dx+dy) == 4)
				if isArmEdge && int(math.Abs(dx)+math.Abs(dy))%3 == 0 && dx*dx+dy*dy > 7.5*7.5 {
					t.IsWalkable = false
					t.AddSpriteByID("MarbleBeam", ss.MarbleBeam)
				}

				// Central Altar
				if dx*dx+dy*dy <= 1.5*1.5 {
					t.AddSpriteByID("FloorTrap", ss.FloorTrap) // Decorative center
					if math.Abs(dx) < 0.5 && math.Abs(dy) < 0.5 {
						t.IsWalkable = false
						t.AddSpriteByID("GlyphStatue", ss.GlyphStatue)
					}
				}
				// 4 Golden Statues flanking the center in a diamond
				if (math.Abs(dx) == 3 && math.Abs(dy) == 0) || (math.Abs(dx) == 0 && math.Abs(dy) == 3) || (math.Abs(dx) == 2.5 && math.Abs(dy) == 2.5) {
					t.IsWalkable = false
					t.AddSpriteByID("GoldenStatue", ss.GoldenStatue)
				}

			} else {
				// Wall Bounding & Autotiling
				hasFloorNeighbor := false
				floorN, floorS, floorE, floorW := false, false, false, false

				if y > 0 && floorMap[y-1][x] { floorN = true; hasFloorNeighbor = true }
				if y < h-1 && floorMap[y+1][x] { floorS = true; hasFloorNeighbor = true }
				if x > 0 && floorMap[y][x-1] { floorW = true; hasFloorNeighbor = true }
				if x < w-1 && floorMap[y][x+1] { floorE = true; hasFloorNeighbor = true }

				if y > 0 && x > 0 && floorMap[y-1][x-1] { hasFloorNeighbor = true }
				if y > 0 && x < w-1 && floorMap[y-1][x+1] { hasFloorNeighbor = true }
				if y < h-1 && x > 0 && floorMap[y+1][x-1] { hasFloorNeighbor = true }
				if y < h-1 && x < w-1 && floorMap[y+1][x+1] { hasFloorNeighbor = true }

				if hasFloorNeighbor {
					isHoriz := (floorN || floorS) && !(floorE || floorW)
					isVert := (floorE || floorW) && !(floorN || floorS)

					t.AddSpriteByID("Floor", ss.Floor) // Base floor under wall
					if isHoriz {
						t.AddSpriteByID("MarbleWallNE", ss.MarbleWallNE)
					} else if isVert {
						t.AddSpriteByID("MarbleWallNW", ss.MarbleWallNW)
					} else {
						t.AddSpriteByID("MarbleWall", ss.MarbleWall)
					}
				}
			}
		}
	}

	return l
}

func GenerateWallTest(w, h int, ss *sprites.SpriteSheet) *Level {
	l := NewEmptyLevel(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t := l.Tiles[y][x]
			t.IsWalkable = true
			t.AddSpriteByID("Floor", ss.Floor)
		}
	}

	wallTypes := []struct {
		ID  string
		Img *ebiten.Image
	}{
		{"MarbleBeam", ss.MarbleBeam},
		{"MarbleBeamNW", ss.MarbleBeamNW},
		{"MarbleBeamNE", ss.MarbleBeamNE},
		{"MarbleWall", ss.MarbleWall},
		{"MarbleBeamNE2", ss.MarbleBeamNE2},
		{"MarbleWall2", ss.MarbleWall2},
		{"MarbleBeamNELong", ss.MarbleBeamNELong},
		{"MarbleWallNELong", ss.MarbleWallNELong},
		{"MarbleBeamNW2", ss.MarbleBeamNW2},
		{"MarbleBeamNWLong", ss.MarbleBeamNWLong},
		{"MarbleWall3", ss.MarbleWall3},
		{"MarbleWallNWLong", ss.MarbleWallNWLong},
		{"MarbleWall4", ss.MarbleWall4},
		{"MarbleWallNW", ss.MarbleWallNW},
		{"MarbleWallNE", ss.MarbleWallNE},
	}

	for i, wt := range wallTypes {
		l.Tiles[5+i][5].AddSpriteByID(wt.ID, wt.Img) // Single

		l.Tiles[5+i][8].AddSpriteByID(wt.ID, wt.Img) // Row X
		l.Tiles[5+i][9].AddSpriteByID(wt.ID, wt.Img)
		l.Tiles[5+i][10].AddSpriteByID(wt.ID, wt.Img)

		l.Tiles[5+i][14].AddSpriteByID(wt.ID, wt.Img) // Col Y
		l.Tiles[6+i][14].AddSpriteByID(wt.ID, wt.Img)
		l.Tiles[7+i][14].AddSpriteByID(wt.ID, wt.Img)
	}

	return l
}
