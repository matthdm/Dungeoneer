// fireball.go
package entities

import (
	"dungeoneer/constants"
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Fireball struct {
	X, Y            float64
	DirX, DirY      float64
	Speed           float64
	Level           int
	Damage          int
	Frame           int
	Tick            int
	Exploding       bool
	Done            bool
	ExplosionSprite *ebiten.Image
	SpriteSheet     *sprites.FireballSprite
}

func NewFireball(x, y, dirX, dirY float64, level int, sheet *sprites.FireballSprite, ss *sprites.SpriteSheet) *Fireball {
	var dmg int
	var exp *ebiten.Image
	switch level {
	case 1:
		dmg = 5
		exp = ss.FireBurst
	case 2:
		dmg = 10
		exp = ss.FireBurst2
	case 3:
		dmg = 20
		exp = ss.FireBurst3
	default:
		level = 1
		dmg = 5
		exp = ss.FireBurst
	}

	length := math.Hypot(dirX, dirY)
	if length == 0 {
		dirX = 1
		length = 1
	}

	return &Fireball{
		X: x, Y: y,
		DirX: dirX / length, DirY: dirY / length,
		Speed:           5.0,
		Level:           level,
		Damage:          dmg,
		ExplosionSprite: exp,
		SpriteSheet:     sheet,
	}
}

func (f *Fireball) Update(level *levels.Level, monsters []*Monster, hitMarkers *[]HitMarker, dmgNums *[]DamageNumber) {
	if f.Done {
		return
	}

	if f.Exploding {
		f.Tick++
		if f.Tick > 15 {
			f.Done = true
		}
		return
	}

	dx := f.DirX * f.Speed / 60
	dy := f.DirY * f.Speed / 60
	f.X += dx
	f.Y += dy

	tx := int(math.Floor(f.X))
	ty := int(math.Floor(f.Y))
	if !level.IsWalkable(tx, ty) {
		f.Exploding = true
		f.Tick = 0
		return
	}

	for _, m := range monsters {
		mx, my := m.GetPosition()
		if int(mx) == tx && int(my) == ty {
			m.TakeDamage(f.Damage, hitMarkers, dmgNums)
			f.Exploding = true
			f.Tick = 0
			return
		}
	}

	f.Tick++
	if f.Tick%6 == 0 {
		f.Frame = (f.Frame + 1) % 8
	}
}

func (f *Fireball) Draw(screen *ebiten.Image, camX, camY, camScale, cx, cy float64) {
	sx, sy := isoToScreenFloat(f.X, f.Y, constants.DefaultTileSize)
	sx = (sx-camX)*camScale + cx
	sy = (sy+camY)*camScale + cy

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx, sy)

	if f.Exploding {
		op.GeoM.Translate(-float64(f.ExplosionSprite.Bounds().Dx())/2, -float64(f.ExplosionSprite.Bounds().Dy())/2)
		op.GeoM.Scale(camScale, camScale)
		screen.DrawImage(f.ExplosionSprite, op)
		return
	}

	angle := math.Atan2(f.DirY, f.DirX)
	dirIdx := int(math.Round(((angle+math.Pi)/(2*math.Pi))*8)) % 8
	sprite := f.SpriteSheet.Frames[dirIdx*8+f.Frame]

	op.GeoM.Translate(-float64(constants.DefaultTileSize)/2, -float64(constants.DefaultTileSize)/2)
	op.GeoM.Scale(camScale, camScale)
	screen.DrawImage(sprite, op)
}
