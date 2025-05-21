package fov

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	shadowImage   = ebiten.NewImage(640, 480)
	triangleImage = ebiten.NewImage(640, 480) // solid white for masking
)

func ResizeShadowBuffer(w, h int) {
	if shadowImage.Bounds().Dx() != w || shadowImage.Bounds().Dy() != h {
		shadowImage = ebiten.NewImage(w, h)
		triangleImage = ebiten.NewImage(w, h)
		triangleImage.Fill(color.White)
	}
}

func DrawShadows(screen *ebiten.Image, rays []Line, camX, camY, camScale float64, cx, cy float64, tileSize int) {
	var offSetX float64 = 1
	if len(rays) < 3 {
		return
	}

	// Clear the shadowImage and start fully dark
	shadowImage.Clear()
	shadowImage.Fill(color.RGBA{0, 0, 0, 200}) // semi-transparent black

	// Prepare a light mask
	lightMask := ebiten.NewImage(shadowImage.Bounds().Dx(), shadowImage.Bounds().Dy())
	lightMask.Fill(color.White)

	opt := &ebiten.DrawTrianglesOptions{}
	opt.Blend = ebiten.BlendSourceOut

	// Fan of triangles from player to ray hits
	for i := range rays {
		r1 := rays[i]
		r2 := rays[(i+1)%len(rays)]

		x0, y0 := worldToScreen(r1.X1+offSetX, r1.Y1, camX, camY, camScale, cx, cy, tileSize)
		x1, y1 := worldToScreen(r1.X2+offSetX, r1.Y2, camX, camY, camScale, cx, cy, tileSize)
		x2, y2 := worldToScreen(r2.X2+offSetX, r2.Y2, camX, camY, camScale, cx, cy, tileSize)

		verts := []ebiten.Vertex{
			{DstX: float32(x0), DstY: float32(y0), ColorA: 1},
			{DstX: float32(x1), DstY: float32(y1), ColorA: 1},
			{DstX: float32(x2), DstY: float32(y2), ColorA: 1},
		}
		shadowImage.DrawTriangles(verts, []uint16{0, 1, 2}, lightMask, opt)
	}

	// Finally, overlay the shadowImage (dimmed) onto the screen
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(shadowImage, op)
}

func DebugDrawRays(screen *ebiten.Image, rays []Line, camX, camY, camScale float64, cx, cy float64, tileSize int) {
	var offSetX float64 = 1
	for _, r := range rays {
		x1, y1 := worldToScreen(r.X1+offSetX, r.Y1, camX, camY, camScale, cx, cy, tileSize)
		x2, y2 := worldToScreen(r.X2+offSetX, r.Y2, camX, camY, camScale, cx, cy, tileSize)

		vector.StrokeLine(screen,
			float32(x1), float32(y1),
			float32(x2), float32(y2),
			1, color.RGBA{255, 255, 0, 180}, true)
	}

}

func DebugDrawWalls(screen *ebiten.Image, walls []Line, camX, camY, camScale, cx, cy float64, tileSize int) {
	var offSetX float64 = 1
	//var offsetY float64 = float64(tileSize) * 0.01 // Shift down visually in isometric space
	for _, wall := range walls {
		// Transform world coords to screen coords using isometric logic
		x1, y1 := worldToScreen(wall.X1+offSetX, wall.Y1, camX, camY, camScale, cx, cy, tileSize)
		x2, y2 := worldToScreen(wall.X2+offSetX, wall.Y2, camX, camY, camScale, cx, cy, tileSize)

		// Draw red line
		vector.StrokeLine(
			screen,
			float32(x1), float32(y1),
			float32(x2), float32(y2),
			1, color.RGBA{255, 0, 0, 255}, true,
		)
	}
}

func triangleVertices(x1, y1, x2, y2, x3, y3 float64) []ebiten.Vertex {
	return []ebiten.Vertex{
		{DstX: float32(x1), DstY: float32(y1), ColorA: 1},
		{DstX: float32(x2), DstY: float32(y2), ColorA: 1},
		{DstX: float32(x3), DstY: float32(y3), ColorA: 1},
	}
}
