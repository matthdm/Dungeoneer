package fov

import (
	"math"
)

type Line struct {
	X1, Y1, X2, Y2 float64
}

type RayCache struct {
	PlayerX, PlayerY float64
	Rays             []Line
}

var cached RayCache

func RayCasting(cx, cy float64, walls []Line) []Line {
	if cx == cached.PlayerX && cy == cached.PlayerY && cached.Rays != nil {
		return cached.Rays
	}

	const rayLength = 1000
	const numRays = 360
	var rays []Line

	for i := 0; i < numRays; i++ {
		angle := float64(i) * (2 * math.Pi / numRays)
		ray := NewRay(cx, cy, rayLength, angle)

		minDist := math.Inf(1)
		var hitX, hitY float64

		for _, wall := range walls {
			if x, y, ok := Intersection(ray, wall); ok {
				dist := (cx-x)*(cx-x) + (cy-y)*(cy-y)
				if dist < minDist {
					minDist = dist
					hitX = x
					hitY = y
				}
			}
		}

		if !math.IsInf(minDist, 1) {
			rays = append(rays, Line{cx, cy, hitX, hitY})
		}
	}

	cached = RayCache{PlayerX: cx, PlayerY: cy, Rays: rays}
	return rays
}
