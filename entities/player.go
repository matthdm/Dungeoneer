package entities

import (
	"dungeoneer/collision"
	"dungeoneer/levels"
	"dungeoneer/movement"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)


type Player struct {
	TileX, TileY int

	Sprite     *ebiten.Image
	LeftFacing bool

	PathPreview []pathing.PathNode
	TickCount   int
	BobOffset   float64

	HP, MaxHP  int
	Damage     int
	AttackRate int
	AttackTick int
	IsDead     bool

	MoveController *movement.MovementController

	CollisionBox collision.Box
}

func NewPlayer(ss *sprites.SpriteSheet) *Player {
	mc := movement.NewMovementController(5) // Speed = 3 tiles/sec
	mc.InterpX = 3
	mc.InterpY = 3

	p := &Player{
		TileX:          3,
		TileY:          3,
		LeftFacing:     true,
		Sprite:         ss.GreyKnight,
		HP:             100,
		MaxHP:          100,
		Damage:         8,
		AttackRate:     60,
		IsDead:         false,
		MoveController: mc,
		CollisionBox:   collision.Box{X: 3, Y: 3, Width: 0.55, Height: 0.8},
	}

	// Whenever InterpX/InterpY crosses into a new tile, update TileX/TileY
	mc.OnStep = func(x, y int) {
		p.TileX = x
		p.TileY = y
	}

	return p
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if p.IsDead {
		return
	}

	// 1) Compute screen coords from continuous (InterpX/InterpY)
	sx, sy := isoToScreenFloat(p.MoveController.InterpX, p.MoveController.InterpY, tileSize)

	// 2) Vertically bob up/down
	const verticalOffset = 1.0
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, -verticalOffset+p.BobOffset)

	// 3) Flip if facing right
	b := p.Sprite.Bounds()
	if !p.LeftFacing {
		w := float64(b.Dx())
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(w, 0)
	}

	// 4) Position at feet, then apply camera
	op.GeoM.Translate(sx, sy)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	screen.DrawImage(p.Sprite, op)

	// Draw health bar above sprite
	p.drawHealthBar(screen, sx, sy, camX, camY, camScale, cx, cy)
}

func (p *Player) drawHealthBar(screen *ebiten.Image, sx, sy, camX, camY, camScale, cx, cy float64) {
	if p.MaxHP <= 0 {
		return
	}
	frac := float64(p.HP) / float64(p.MaxHP)
	barW := 32.0
	barH := 4.0
	filled := int(barW * frac)
	if filled < 1 {
		filled = 1
	}

	full := ebiten.NewImage(int(barW), int(barH))
	full.Fill(color.RGBA{255, 0, 0, 255})
	part := ebiten.NewImage(filled, int(barH))
	part.Fill(color.RGBA{0, 255, 0, 255})

	barOp := &ebiten.DrawImageOptions{}
	barOp.GeoM.Translate(sx-barW/2+30, sy)
	barOp.GeoM.Translate(-camX, camY)
	barOp.GeoM.Scale(camScale, camScale)
	barOp.GeoM.Translate(cx, cy)

	screen.DrawImage(full, barOp)
	screen.DrawImage(part, barOp)
}

func (p *Player) CanMoveTo(x, y int, level *levels.Level) bool {
	return x >= 0 && y >= 0 && x < level.W && y < level.H
}

func (p *Player) Update(level *levels.Level, dt float64) {
	// Bobbing
	const bobFreq = 0.3
	const bobAmp = 1.5
	p.TickCount++
	p.AttackTick++
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFreq) * bobAmp

	// For pathing, let the controller interpolate positions
	if p.MoveController.Mode == movement.PathingMode {
		p.MoveController.Update(dt)
	}

	// If PathingMode, check validity of next node and flip sprite direction
	if p.MoveController.Mode == movement.PathingMode && len(p.MoveController.Path) > 0 && !p.MoveController.Moving {
		next := p.MoveController.Path[0]
		if !p.CanMoveTo(next.X, next.Y, level) {
			p.MoveController.Path = nil
			return
		}
		p.LeftFacing = next.X < p.TileX
	}
	// If pure VelocityMode we will update tile coords after resolving movement

	// Update collision box so its center is at “feet” (InterpX,InterpY – half Height)
	p.CollisionBox.X = p.MoveController.InterpX
	p.CollisionBox.Y = p.MoveController.InterpY - (p.CollisionBox.Height / 2)

	// If moving by velocity, resolve collisions each frame
	if p.MoveController.Mode == movement.VelocityMode {
		vx := p.MoveController.VelocityX * dt
		vy := p.MoveController.VelocityY * dt

		// Clamp displacement to avoid tunneling
		maxStep := 0.25
		if math.Abs(vx) > maxStep {
			vx = math.Copysign(maxStep, vx)
		}
		if math.Abs(vy) > maxStep {
			vy = math.Copysign(maxStep, vy)
		}

		// Sweep box and clip movement against the tile map
		finalBox, hitX, hitY := collision.PredictAndClip(level, p.CollisionBox, vx, vy)

		// Stop velocity on the axes we collided with to prevent wall clipping
		if hitX {
			p.MoveController.VelocityX = 0
		}
		if hitY {
			p.MoveController.VelocityY = 0
		}

		// Apply resolved position back to movement controller
		p.MoveController.InterpX = finalBox.X
		p.MoveController.InterpY = finalBox.Y + (p.CollisionBox.Height / 2)

		// Sync collision box with new position
		p.CollisionBox = finalBox

		// Update tile coordinates from final position
		p.TileX = int(p.MoveController.InterpX)
		p.TileY = int(p.MoveController.InterpY)
	}
}

func (p *Player) SetPath(path []pathing.PathNode) {
	p.MoveController.SetPath(path)
}

func (p *Player) CanAttack() bool {
	return p.AttackTick >= p.AttackRate
}

func (p *Player) TakeDamage(dmg int) {
	p.HP -= dmg
	if p.HP <= 0 {
		p.HP = 0
		p.IsDead = true
	}
}
