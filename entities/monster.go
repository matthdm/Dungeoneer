package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Monster struct {
	Name       string
	Sprite     *ebiten.Image
	LeftFacing bool
	BobOffset  float64

	// Grid position
	TileX, TileY int

	// Interpolation
	InterpX, InterpY float64
	StartX, StartY   float64
	TargetX, TargetY float64
	InterpTicks      int
	Moving           bool

	// Movement logic
	Path             []pathing.PathNode
	TickCount        int
	MovementDuration int // ticks per tile
	RecalcCooldown   int

	// Combat
	HP         int
	MaxHP      int
	Damage     int
	AttackRate int
	AttackTick int
	IsDead     bool
}

func NewMonster(ss *sprites.SpriteSheet) []*Monster {
	return []*Monster{
		{
			Name:             "Skeleton",
			TileX:            5,
			TileY:            7,
			InterpX:          5,
			InterpY:          7,
			Sprite:           ss.BlueMan, // swap in any sprite
			MovementDuration: 45,
			LeftFacing:       true,
			HP:               10,
			MaxHP:            10,
			Damage:           1,
			AttackRate:       30,
			IsDead:           false,
		},
	}
}

func (m *Monster) Update(player *Player, level *levels.Level) {
	const bobAmplitude = 1.5
	const bobFrequency = 0.15

	m.TickCount++
	m.BobOffset = math.Sin(float64(m.TickCount)*bobFrequency) * bobAmplitude

	// Smooth interpolation update
	if m.Moving {
		m.InterpTicks++
		t := float64(m.InterpTicks) / float64(m.MovementDuration)
		if t > 1 {
			t = 1
		}
		m.InterpX = m.StartX + (m.TargetX-m.StartX)*t
		m.InterpY = m.StartY + (m.TargetY-m.StartY)*t

		if t >= 1 {
			m.Moving = false
			m.TileX = int(m.TargetX)
			m.TileY = int(m.TargetY)
			m.InterpX = m.TargetX
			m.InterpY = m.TargetY
		}
		return
	}

	if m.RecalcCooldown > 0 {
		m.RecalcCooldown--
		return
	}

	// Check for path recompute
	needRecalc := len(m.Path) == 0 || !level.IsWalkable(m.Path[0].X, m.Path[0].Y)
	if needRecalc {
		m.Path = pathing.AStar(level, m.TileX, m.TileY, player.TileX, player.TileY)
		m.RecalcCooldown = 30
		if len(m.Path) > 0 && m.Path[0].X == m.TileX && m.Path[0].Y == m.TileY {
			m.Path = m.Path[1:]
		}
	}

	if len(m.Path) > 0 {
		next := m.Path[0]
		if !level.IsWalkable(next.X, next.Y) {
			m.Path = nil
			return
		}

		if next.X > m.TileX {
			m.LeftFacing = false
		} else if next.X < m.TileX {
			m.LeftFacing = true
		}

		m.StartX = m.InterpX
		m.StartY = m.InterpY
		m.TargetX = float64(next.X)
		m.TargetY = float64(next.Y)
		m.InterpTicks = 0
		m.Moving = true
		m.Path = m.Path[1:]
	}
	if !m.IsDead && IsAdjacent(m.TileX, m.TileY, player.TileX, player.TileY) {
		m.AttackTick++
		if m.AttackTick >= m.AttackRate {
			player.TakeDamage(m.Damage)
			m.AttackTick = 0
		}
	} else {
		m.AttackTick = 0 // Reset if not in range
	}
}
func (m *Monster) Draw(
	screen *ebiten.Image,
	tileSize int,
	isoToScreen func(int, int) (float64, float64),
	camX, camY, camScale, cx, cy float64,
) {
	if m.Sprite == nil || m.IsDead {
		return
	}

	//x, y := isoToScreen(int(m.InterpX), int(m.InterpY))
	x, y := isoToScreenFloat(m.InterpX, m.InterpY, 64)
	op := &ebiten.DrawImageOptions{}
	bounds := m.Sprite.Bounds()
	spriteW := float64(bounds.Dx())
	spriteH := float64(bounds.Dy())

	const verticalOffset = 0.1

	// Apply bob offset
	op.GeoM.Translate(0, -verticalOffset+m.BobOffset)

	// Flip horizontally if not facing left
	if !m.LeftFacing {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteW, 0)
	}

	// Move to tile screen position
	op.GeoM.Translate(x, y)

	// Camera transform
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	//Monster
	screen.DrawImage(m.Sprite, op)

	// Health bar
	if !m.IsDead && m.MaxHP > 0 {
		hpPercent := float64(m.HP) / float64(m.MaxHP)
		barW, barH := 32.0, 4.0

		// Red bar background
		hpBG := ebiten.NewImage(int(barW), int(barH))
		hpBG.Fill(color.RGBA{100, 0, 0, 255})

		// Green bar
		hpFG := ebiten.NewImage(int(barW*hpPercent), int(barH))
		hpFG.Fill(color.RGBA{0, 255, 0, 255})

		// Position
		hpOp := &ebiten.DrawImageOptions{}
		hpOp.GeoM.Translate(x-barW/2, y-spriteH-1) // slightly above sprite
		hpOp.GeoM.Translate(-camX, camY)
		hpOp.GeoM.Scale(camScale, camScale)
		hpOp.GeoM.Translate(cx, cy)

		screen.DrawImage(hpBG, hpOp)

		hpOp = &ebiten.DrawImageOptions{}
		hpOp.GeoM.Translate(x-barW/2, y-spriteH-1)
		hpOp.GeoM.Translate(-camX, camY)
		hpOp.GeoM.Scale(camScale, camScale)
		hpOp.GeoM.Translate(cx, cy)

		screen.DrawImage(hpFG, hpOp)
	}
}

func (m *Monster) TakeDamage(dmg int) {
	m.HP -= dmg
	if m.HP <= 0 {
		m.IsDead = true
	}
}
