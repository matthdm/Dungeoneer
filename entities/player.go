package entities

import (
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

	// Combat
	HP         int
	MaxHP      int
	Damage     int
	AttackRate int
	AttackTick int
	IsDead     bool

	// New movement controller
	MoveController *movement.MovementController
}

func NewPlayer(ss *sprites.SpriteSheet) *Player {
	mc := movement.NewMovementController(3)
	mc.InterpX = 3
	mc.InterpY = 3

	p := &Player{
		TileX:          3,
		TileY:          3,
		LeftFacing:     true,
		Sprite:         ss.GreyKnight,
		HP:             10,
		MaxHP:          10,
		Damage:         2,
		AttackRate:     60,
		IsDead:         false,
		MoveController: mc,
	}

	// Sync TileX/TileY on every new tile start
	mc.OnStep = func(x, y int) {
		p.TileX = x
		p.TileY = y
	}

	return p
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, isoToScreen func(int, int) (float64, float64), camX, camY, camScale, cx, cy float64) {
	if p.IsDead {
		return
	}
	//x, y := isoToScreen(int(p.InterpX), int(p.InterpY))
	// Optional: more accurate rendering
	op := &ebiten.DrawImageOptions{}
	x, y := isoToScreenFloat(p.MoveController.InterpX, p.MoveController.InterpY, 64)
	bounds := p.Sprite.Bounds()
	spriteW := float64(bounds.Dx())

	// Then apply the bob
	const verticalOffset = 1.0 // tweak until it feels good
	op.GeoM.Translate(0, -verticalOffset+p.BobOffset)

	// 2. Flip horizontally if facing right
	if !p.LeftFacing {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteW, 0)
	}

	// 3. Move to isometric position
	op.GeoM.Translate(x, y)

	// 4. Apply camera transform
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	//Player Draw
	screen.DrawImage(p.Sprite, op)

	p.UpdateHealthBar(screen, x, y, camX, camY, camScale, cx, cy)
}
func (p *Player) UpdateHealthBar(screen *ebiten.Image, x, y float64, camX, camY, camScale, cx, cy float64) {
	if p.MaxHP > 0 {
		hpPercent := float64(p.HP) / float64(p.MaxHP)
		barWidth := 32.0
		barHeight := 4.0

		// Ensure the width is always at least 1 to avoid panic
		width := int(barWidth * hpPercent)
		if width < 1 {
			width = 1
		}

		hpBar := ebiten.NewImage(int(barWidth), int(barHeight))
		hpBar.Fill(color.RGBA{255, 0, 0, 255})
		hpBarFilled := ebiten.NewImage(int(width), int(barHeight))
		hpBarFilled.Fill(color.RGBA{0, 255, 0, 255})

		// Position HP bar
		barOp := &ebiten.DrawImageOptions{}
		barOp.GeoM.Translate(x-barWidth/2+30, y)
		barOp.GeoM.Translate(-camX, camY)
		barOp.GeoM.Scale(camScale, camScale)
		barOp.GeoM.Translate(cx, cy)

		screen.DrawImage(hpBar, barOp)
		screen.DrawImage(hpBarFilled, barOp)
	}
}

func (p *Player) CanMoveTo(x, y int, level *levels.Level) bool {
	return x >= 0 && y >= 0 && x < level.W && y < level.H
}

func (p *Player) Update(level *levels.Level, dt float64) {
	const bobAmplitude = 1.5
	var bobFrequency = 0.3

	p.TickCount++
	p.AttackTick++
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFrequency) * bobAmplitude

	p.MoveController.Update(dt)

	if p.MoveController.Mode == movement.PathingMode && len(p.MoveController.Path) > 0 && !p.MoveController.Moving {
		next := p.MoveController.Path[0]
		if !p.CanMoveTo(next.X, next.Y, level) {
			p.MoveController.Path = nil
			return
		}

		p.LeftFacing = next.X < p.TileX
	}

	if p.MoveController.Mode == movement.VelocityMode {
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
	if p.HP < 0 {
		p.HP = 0
		p.IsDead = true
	}
	// TODO: Add visual feedback / flashing / etc
}
