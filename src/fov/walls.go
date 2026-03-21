package fov

import "dungeoneer/levels"

func LevelToWalls(level *levels.Level) []Line {
	var walls []Line

	for y := 0; y < level.H; y++ {
		for x := 0; x < level.W; x++ {
			t := level.Tiles[y][x]
			if t == nil || t.IsWalkable {
				continue
			}

			xf := float64(x)
			yf := float64(y)
			w := 1.0
			//add 4 bound edges to the walls
			walls = append(walls, []Line{
				{X1: xf, Y1: yf, X2: xf + w, Y2: yf},
				{X1: xf + w, Y1: yf, X2: xf + w, Y2: yf + w},
				{X1: xf + w, Y1: yf + w, X2: xf, Y2: yf + w},
				{X1: xf, Y1: yf + w, X2: xf, Y2: yf},
			}...)
		}
	}
	return walls
}
