package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	TileX, TileY     int
	InterpX, InterpY float64
	StartX, StartY   float64
	TargetX, TargetY float64
	InterpTicks      int
	Moving           bool

	Sprite      *ebiten.Image
	LeftFacing  bool
	Path        []pathing.PathNode
	PathPreview []pathing.PathNode
	TickCount   int
	BobOffset   float64

	MovementDuration int // ticks per tile

	// Combat
	HP         int
	MaxHP      int
	Damage     int
	AttackRate int
	AttackTick int
	IsDead     bool
}

func NewPlayer(ss *sprites.SpriteSheet) *Player {
	return &Player{
		TileX:            3,
		TileY:            3,
		LeftFacing:       true,
		Sprite:           ss.GreyKnight,
		MovementDuration: 15,
		InterpX:          float64(3),
		InterpY:          float64(3),
		HP:               10,
		MaxHP:            10,
		Damage:           2,
		AttackRate:       60,
		IsDead:           false,
	}
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, isoToScreen func(int, int) (float64, float64), camX, camY, camScale, cx, cy float64) {
	//x, y := isoToScreen(int(p.InterpX), int(p.InterpY))
	// Optional: more accurate rendering
	op := &ebiten.DrawImageOptions{}
	x, y := isoToScreenFloat(p.InterpX, p.InterpY, 64)
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

		hpBar := ebiten.NewImage(int(barWidth), int(barHeight))
		hpBar.Fill(color.RGBA{255, 0, 0, 255})
		hpBarFilled := ebiten.NewImage(int(barWidth*hpPercent), int(barHeight))
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

func (p *Player) Update(level *levels.Level) {
	const bobAmplitude = 1.5
	var bobFrequency = 0.3

	p.TickCount++
	p.AttackTick++
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFrequency) * bobAmplitude

	if p.Moving {
		p.InterpTicks++
		t := float64(p.InterpTicks) / float64(p.MovementDuration)
		if t > 1 {
			t = 1
		}

		p.InterpX = p.StartX + (p.TargetX-p.StartX)*t
		p.InterpY = p.StartY + (p.TargetY-p.StartY)*t
		p.BobOffset = math.Sin(float64(p.InterpTicks)*bobFrequency) * bobAmplitude

		if t >= 1 {
			p.Moving = false
			p.TileX = int(p.TargetX)
			p.TileY = int(p.TargetY)
			p.InterpX = p.TargetX
			p.InterpY = p.TargetY
		}
		return
	} else {
		bobFrequency = 0.1
		p.BobOffset = math.Sin(float64(p.TickCount)*bobFrequency) * bobAmplitude
	}

	if len(p.Path) > 0 {
		next := p.Path[0]

		if !p.CanMoveTo(next.X, next.Y, level) {
			p.Path = nil
			return
		}

		p.LeftFacing = next.X < p.TileX

		p.StartX = p.InterpX
		p.StartY = p.InterpY
		p.TargetX = float64(next.X)
		p.TargetY = float64(next.Y)
		p.InterpTicks = 0
		p.Moving = true

		p.Path = p.Path[1:]
	}
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
