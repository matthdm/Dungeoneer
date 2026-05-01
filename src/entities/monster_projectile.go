package entities

import (
	"dungeoneer/coords"
	"dungeoneer/levels"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// MonsterProjectile is a ranged attack fired by a monster toward the player.
type MonsterProjectile struct {
	X, Y       float64 // world position (tile coords, fractional)
	DirX, DirY float64 // normalised direction
	Speed      float64 // tiles per tick
	Damage     int
	Radius     float64 // hit radius in tiles
	Finished   bool
	TicksLived int
	MaxTicks   int // auto-expire
}

// NewMonsterProjectile creates a projectile from (sx,sy) aimed at (tx,ty).
func NewMonsterProjectile(sx, sy float64, tx, ty float64, speed float64, damage int) *MonsterProjectile {
	dx := tx - sx
	dy := ty - sy
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.001 {
		dist = 1
	}
	return &MonsterProjectile{
		X:        sx,
		Y:        sy,
		DirX:     dx / dist,
		DirY:     dy / dist,
		Speed:    speed,
		Damage:   damage,
		Radius:   0.6,
		MaxTicks: 180, // 3 seconds at 60 TPS
	}
}

// Update advances the projectile one tick and checks wall collision.
func (p *MonsterProjectile) Update(level *levels.Level) {
	if p.Finished {
		return
	}
	p.X += p.DirX * p.Speed
	p.Y += p.DirY * p.Speed
	p.TicksLived++

	// Expire on max age.
	if p.TicksLived >= p.MaxTicks {
		p.Finished = true
		return
	}
	// Expire on wall hit.
	tileX := int(math.Floor(p.X))
	tileY := int(math.Floor(p.Y))
	if !level.IsWalkable(tileX, tileY) {
		p.Finished = true
	}
}

// HitsPlayer returns true if the projectile overlaps the player's body center.
// Uses the player's continuous WorldPos rather than integer tile coords so the
// check matches the player's visual screen position at all times.
func (p *MonsterProjectile) HitsPlayer(playerPos coords.WorldPos) bool {
	bc := playerPos.BodyCenter()
	dx := p.X - bc.X
	dy := p.Y - bc.Y
	return dx*dx+dy*dy <= p.Radius*p.Radius
}

// Draw renders the projectile as a small coloured circle.
func (p *MonsterProjectile) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if p.Finished {
		return
	}
	// Simple 4x4 red dot at projectile position.
	size := 4
	img := ebiten.NewImage(size, size)
	img.Fill(color.RGBA{255, 80, 40, 255})

	// TileCenterIso gives the visual center of the tile diamond — the correct
	// anchor for a projectile dot so it tracks the projectile's world position.
	cx_, cy_ := coords.WorldPos{X: p.X, Y: p.Y}.TileCenterIso(tileSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(cx_, cy_)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	screen.DrawImage(img, op)
}
