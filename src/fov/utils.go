package fov

import "math"

func NewRay(x, y, length, angle float64) Line {
	return Line{
		X1: x,
		Y1: y,
		X2: x + length*math.Cos(angle),
		Y2: y + length*math.Sin(angle),
	}
}

func Intersection(l1, l2 Line) (float64, float64, bool) {
	denom := (l1.X1-l1.X2)*(l2.Y1-l2.Y2) - (l1.Y1-l1.Y2)*(l2.X1-l2.X2)
	if denom == 0 {
		return 0, 0, false
	}

	t := ((l1.X1-l2.X1)*(l2.Y1-l2.Y2) - (l1.Y1-l2.Y1)*(l2.X1-l2.X2)) / denom
	u := -((l1.X1-l1.X2)*(l1.Y1-l2.Y1) - (l1.Y1-l1.Y2)*(l1.X1-l2.X1)) / denom

	if t < 0 || t > 1 || u < 0 || u > 1 {
		return 0, 0, false
	}

	x := l1.X1 + t*(l1.X2-l1.X1)
	y := l1.Y1 + t*(l1.Y2-l1.Y1)
	return x, y, true
}

func worldToScreen(x, y float64, camX, camY, camScale float64, cx, cy float64, tileSize int) (float64, float64) {
	// Convert cartesian to isometric
	isoX := (x - y) * float64(tileSize/2)
	isoY := (x + y) * float64(tileSize/4)

	screenX := (isoX-camX)*camScale + cx
	screenY := (isoY+camY)*camScale + cy
	return screenX, screenY
}
