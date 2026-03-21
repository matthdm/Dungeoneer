package spells

import (
	"dungeoneer/levels"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Branch represents a single line segment in the fractal canopy.
// Coordinates are in tile space.
type Branch struct {
	X1, Y1 float64
	X2, Y2 float64
	Age    float64 // time since spawned
	Depth  int
}

type FractalCanopyVisual struct {
	Branches []*Branch
	MaxDepth int
	lastGrow float64
	depth    int
}

func NewFractalCanopyVisual(x, y float64, maxDepth int) *FractalCanopyVisual {
	trunkLength := 3.0

	// True isometric "up" is (-1, -1) direction
	dx := -trunkLength * math.Cos(math.Pi/4) // cos(45°)
	dy := -trunkLength * math.Sin(math.Pi/4) // sin(45°)

	trunk := &Branch{
		X1:    x,
		Y1:    y,
		X2:    x + dx,
		Y2:    y + dy,
		Depth: 0,
	}

	return &FractalCanopyVisual{
		Branches: []*Branch{trunk},
		MaxDepth: maxDepth,
	}
}

func (fv *FractalCanopyVisual) Update(dt float64) {
	fv.lastGrow += dt
	for _, b := range fv.Branches {
		b.Age += dt
	}
	if fv.depth >= fv.MaxDepth {
		return
	}
	if fv.lastGrow >= 1.0 {
		fv.lastGrow = 0
		fv.depth++
		fv.Grow()
	}
}

func (fv *FractalCanopyVisual) Grow() {
	const angle = math.Pi / 5 // 36 degrees
	const baseLength = 2.0
	const lengthDecay = 0.8

	var parents []*Branch
	for _, b := range fv.Branches {
		if b.Depth == fv.depth-1 {
			parents = append(parents, b)
		}
	}

	for _, p := range parents {
		baseAngle := math.Atan2(p.Y2-p.Y1, p.X2-p.X1)
		for _, dir := range []float64{angle, -angle} {
			childAngle := baseAngle + dir
			length := baseLength * math.Pow(lengthDecay, float64(fv.depth))

			dx := math.Cos(childAngle) * length
			dy := math.Sin(childAngle) * length

			child := &Branch{
				X1:    p.X2,
				Y1:    p.Y2,
				X2:    p.X2 + dx,
				Y2:    p.Y2 + dy,
				Depth: fv.depth,
			}
			fv.Branches = append(fv.Branches, child)
		}
	}
}

func lerpColor(a, b color.NRGBA, t float64) color.NRGBA {
	if t > 1 {
		t = 1
	}
	return color.NRGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

// FractalCanopy is a healing-over-time spell that grows outward like
// branching roots. The caster is healed if standing within the
// current radius.
type FractalCanopy struct {
	Age         float64
	MaxGrowTime float64
	MaxDuration float64
	MaxRadius   float64

	HealingMin float64
	HealingMax float64

	X, Y   float64
	Radius float64

	Visual *FractalCanopyVisual

	Finished bool

	HealingTickTimer float64
}

func (fc *FractalCanopy) Update(level *levels.Level, dt float64) {
	if fc.Finished {
		return
	}

	fc.Age += dt

	if fc.Visual != nil {
		fc.Visual.Update(dt)
	}

	if fc.Age < fc.MaxGrowTime {
		growFrac := fc.Age / fc.MaxGrowTime
		fc.Radius = fc.MaxRadius * growFrac
	} else {
		fc.Radius = fc.MaxRadius
	}

	if fc.Age >= fc.MaxGrowTime+fc.MaxDuration {
		fc.Finished = true
	}
}

// Draw renders the canopy base glow and animated branches.
func (fc *FractalCanopy) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if fc.Finished {
		return
	}
	// base glow
	sx, sy := isoToScreenFloat(fc.X, fc.Y, tileSize)
	baseCol := color.NRGBA{R: 100, G: 255, B: 120, A: 200}
	glow := ebiten.NewImage(4, 4)
	glow.Fill(baseCol)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx-2, sy-2)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	screen.DrawImage(glow, op)

	if fc.Visual == nil {
		return
	}

	if fc.Age > fc.MaxGrowTime {
		remain := fc.MaxGrowTime + fc.MaxDuration - fc.Age
		if remain < 0 {
			remain = 0
		}
	}
	centerTileX := int(math.Floor(fc.X))
	centerTileY := int(math.Floor(fc.Y))
	aoeColor := color.NRGBA{R: 80, G: 255, B: 120, A: 80}

	radius := int(fc.Radius)
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if abs(dx)+abs(dy) <= radius {
				tx := centerTileX + dx
				ty := centerTileY + dy
				drawAOETile(screen, tx+2, ty+1, tileSize, camX, camY, camScale, cx, cy, aoeColor)
			}
		}
	}

	for _, b := range fc.Visual.Branches {
		progress := b.Age
		if progress > 1 {
			progress = 1
		}
		x1 := b.X1
		y1 := b.Y1
		x2 := b.X1 + (b.X2-b.X1)*progress
		y2 := b.Y1 + (b.Y2-b.Y1)*progress

		sx1, sy1 := screenCoordsFromWorldTile(x1, y1, tileSize, camX, camY, camScale, cx, cy)
		sx2, sy2 := screenCoordsFromWorldTile(x2, y2, tileSize, camX, camY, camScale, cx, cy)

		// Color fade: green (early) to orange (late)
		startCol := color.NRGBA{R: 50, G: 255, B: 80, A: 255}
		endCol := color.NRGBA{R: 255, G: 120, B: 0, A: 255}
		ageFrac := math.Min(b.Age/10.0, 1.0)
		col := lerpColor(startCol, endCol, ageFrac)

		width := float32(0.5 + float64(fc.Visual.MaxDepth-b.Depth)*0.3)
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), width, col, true)
	}
}

// IsFinished returns true when the spell duration has ended.
func (fc *FractalCanopy) IsFinished() bool { return fc.Finished }
