package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Monster struct {
	TileX, TileY     int
	InterpX, InterpY float64
	recalcCooldown   int
	Sprite           *ebiten.Image
	Name             string
	movementDuration int // how often it moves (e.g. 30 ticks)
	TickCount        int
	LeftFacing       bool
	BobOffset        float64
	Path             []pathing.PathNode
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
			movementDuration: 45,
			LeftFacing:       true,
		},
	}
}

func (m *Monster) Update(playerX, playerY int, level *levels.Level) {
	const bobAmplitude = 1.5
	const bobFrequency = 0.15
	m.TickCount++
	m.BobOffset = math.Sin(float64(m.TickCount)*bobFrequency) * bobAmplitude
	if m.recalcCooldown > 0 {
		m.recalcCooldown--
		return
	}

	if m.TickCount < m.movementDuration {
		return
	}
	m.TickCount = 0

	// If no path or path blocked or stuck, recalc path
	needRecalc := false
	if len(m.Path) == 0 {
		needRecalc = true
	} else {
		next := m.Path[0]
		if !level.IsWalkable(next.X, next.Y) {
			needRecalc = true
		}
	}

	if needRecalc {
		m.Path = pathing.AStar(level, m.TileX, m.TileY, playerX, playerY)
		m.recalcCooldown = 30 // wait 30 ticks before recalculating again
		// drop current pos from path
		if len(m.Path) > 0 && m.Path[0].X == m.TileX && m.Path[0].Y == m.TileY {
			m.Path = m.Path[1:]
		}
	}

	// Try to move one step
	if len(m.Path) > 0 {
		next := m.Path[0]
		if level.IsWalkable(next.X, next.Y) {
			m.TileX = next.X
			m.TileY = next.Y
			m.Path = m.Path[1:]
		} else {
			// path blocked unexpectedly
			m.Path = nil
		}
	}

	const interpSpeed = 1 // how fast to approach target tile

	dx := float64(m.TileX) - m.InterpX
	dy := float64(m.TileY) - m.InterpY

	m.InterpX += dx * interpSpeed
	m.InterpY += dy * interpSpeed
}

func (m *Monster) Draw(
	screen *ebiten.Image,
	tileSize int,
	isoToScreen func(int, int) (float64, float64),
	camX, camY, camScale, cx, cy float64,
) {
	if m.Sprite == nil {
		return
	}

	//x, y := isoToScreen(int(m.InterpX), int(m.InterpY))
	x, y := isoToScreenFloat(m.InterpX, m.InterpY, 64)
	op := &ebiten.DrawImageOptions{}
	bounds := m.Sprite.Bounds()
	spriteW := float64(bounds.Dx())
	spriteH := float64(bounds.Dy())

	const verticalOffset = 0.1

	// Center the sprite
	op.GeoM.Translate(-spriteW/2, -spriteH/2)

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

	screen.DrawImage(m.Sprite, op)
}
