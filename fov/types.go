package fov

type Object struct {
	Walls []Line
}

func (o Object) Points() [][2]float64 {
	var points [][2]float64
	for _, wall := range o.Walls {
		points = append(points, [2]float64{wall.X2, wall.Y2})
	}
	p := [2]float64{o.Walls[0].X1, o.Walls[0].Y1}
	if p[0] != points[len(points)-1][0] || p[1] != points[len(points)-1][1] {
		points = append(points, p)
	}
	return points
}
