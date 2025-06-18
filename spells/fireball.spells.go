package spells

import (
	"dungeoneer/images"
	"dungeoneer/levels"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// FireballSprites holds animations for all 8 directions.
// Indexed by [direction][frame]
var FireballSprites [][]*ebiten.Image

func LoadFireballSprites() ([][]*ebiten.Image, error) {
	sheet, err := images.LoadEmbeddedImage(images.Fireball_0_png)
	if err != nil {
		return nil, err
	}
	const dirs = 8
	const framesPerDir = 8
	const size = 64
	sprites := make([][]*ebiten.Image, dirs)
	spriteAt := func(x, y int) *ebiten.Image {
		rect := image.Rect(x*size, y*size, (x+1)*size, (y+1)*size)
		return sheet.SubImage(rect).(*ebiten.Image)
	}
	for d := 0; d < dirs; d++ {
		sprites[d] = make([]*ebiten.Image, framesPerDir)
		for f := 0; f < framesPerDir; f++ {
			sprites[d][f] = spriteAt(f, d)
		}
	}
	return sprites, nil
}

// Fireball projectile
type Fireball struct {
	Info       SpellInfo
	X, Y       float64
	DirX, DirY float64
	Speed      float64

	Angle float64

	dirIndex   int
	frame      int
	tick       int
	Impact     bool
	impactTick int
	ImpactImg  *ebiten.Image
	Finished   bool
}

func NewFireball(info SpellInfo, startX, startY, targetX, targetY float64, sprites [][]*ebiten.Image, impact *ebiten.Image) *Fireball {
	dx := targetX - startX
	dy := targetY - startY
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1
	}
	dx /= dist
	dy /= dist
	// Compute angle in world space (0 rad = east, counter-clockwise)
	// Use negative Y because screen Y increases downward
	angle := math.Atan2(-dy, dx)
	fb := &Fireball{
		Info:      info,
		X:         startX,
		Y:         startY,
		DirX:      dx,
		DirY:      dy,
		Speed:     8.0,
		Angle:     angle,
		ImpactImg: impact,
	}
	fb.dirIndex = angleToDir(angle)
	FireballSprites = sprites
	return fb
}

func angleToDir(a float64) int {
	if a < 0 {
		a += 2 * math.Pi
	}
	seg := math.Pi / 4 // 45 degrees per segment
	idx := int(math.Round(a/seg)) % 8
	mapping := []int{5, 4, 3, 2, 1, 0, 7, 6}
	return mapping[idx]
}

func (f *Fireball) Update(level *levels.Level, dt float64) {
	if f.Finished {
		return
	}
	if f.Impact {
		f.impactTick++
		if f.impactTick > 15 {
			f.Finished = true
		}
		return
	}

	step := f.Speed * dt
	f.X += f.DirX * step
	f.Y += f.DirY * step

	f.tick++
	if f.tick > 3 {
		f.tick = 0
		f.frame = (f.frame + 1) % len(FireballSprites[f.dirIndex])
	}

	tx := int(math.Floor(f.X))
	ty := int(math.Floor(f.Y))
	if !level.IsWalkable(tx, ty) {
		f.Impact = true
		f.X = float64(tx)
		f.Y = float64(ty)
	}
}

func (f *Fireball) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if f.Finished {
		return
	}
	sx, sy := isoToScreenFloat(f.X, f.Y, tileSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx, sy)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	if f.Impact {
		if f.ImpactImg != nil {
			screen.DrawImage(f.ImpactImg, op)
		}
		return
	}

	img := FireballSprites[f.dirIndex][f.frame]
	screen.DrawImage(img, op)
}

// DebugDraw visualizes the angle and animation row
func (f *Fireball) DebugDraw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	sx, sy := isoToScreenFloat(f.X, f.Y, tileSize)
	sx = (sx-camX)*camScale + cx
	sy = (sy+camY)*camScale + cy
	msg := fmt.Sprintf("%.1f° row %d", f.Angle*180/math.Pi, f.dirIndex)
	ebitenutil.DebugPrintAt(screen, msg, int(sx), int(sy)-10)
}

func (f *Fireball) IsFinished() bool { return f.Finished }

func isoToScreenFloat(x, y float64, tileSize int) (float64, float64) {
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}
