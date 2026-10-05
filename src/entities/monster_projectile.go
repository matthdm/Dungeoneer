package entities

import (
	"dungeoneer/coords"
	"dungeoneer/levels"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MonsterProjectile is a ranged attack fired by a monster toward the player.
type MonsterProjectile struct {
	X, Y       float64 // world ground point the projectile is above
	DirX, DirY float64 // normalised direction
	Speed      float64 // tiles per tick
	Damage     int
	Radius     float64 // hit radius in tiles
	Finished   bool
	TicksLived int
	MaxTicks   int // auto-expire
}

// NewMonsterProjectile creates a projectile from (sx,sy) aimed at (tx,ty),
// both world ground points (pass BodyCenter() of shooter and target).
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

// HitsPlayer returns true if the projectile overlaps the player's body.
// playerPos is the player's stored position (Pos()); the test is made against
// its BodyCenter() in world space — the same frame the projectile flies in.
func (p *MonsterProjectile) HitsPlayer(playerPos coords.WorldPos) bool {
	bc := playerPos.BodyCenter()
	dx := p.X - bc.X
	dy := p.Y - bc.Y
	return dx*dx+dy*dy <= p.Radius*p.Radius
}

// Draw renders the projectile as a small coloured dot flying at body height.
func (p *MonsterProjectile) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if p.Finished {
		return
	}
	ix, iy := coords.GroundToIso(p.X, p.Y, tileSize)
	iy -= coords.BodyHeightPx(tileSize)
	sx := float32((ix-camX)*camScale + cx)
	sy := float32((iy+camY)*camScale + cy)
	vector.DrawFilledCircle(screen, sx, sy, float32(2.5*camScale), color.RGBA{255, 80, 40, 255}, true)
}
