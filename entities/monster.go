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
			movementDuration: 60,
			LeftFacing:       true,
		},
	}
}

func (m *Monster) Update(playerX, playerY int, level *levels.Level) {
	const bobAmplitude = 1.5
	const bobFrequency = 0.15
	m.TickCount++
	m.BobOffset = math.Sin(float64(m.TickCount)*bobFrequency) * bobAmplitude
	// Move only every N ticks
	if m.TickCount < m.movementDuration {
		return
	}
	m.TickCount = 0

	if len(m.Path) > 0 {
		next := m.Path[0]
		if level.IsWalkable(next.X, next.Y) {
			if next.X >= 0 && next.Y >= 0 && next.X < level.W && next.Y < level.H {
				m.TileX = next.X
				m.TileY = next.Y
				m.Path = m.Path[1:]

				if next.X > m.TileX {
					m.LeftFacing = false
				} else if next.X < m.TileX {
					m.LeftFacing = true
				}
			} else {
				// Out of bounds path — clear it
				m.Path = nil
			}
		}
		return
	}

	// Compute path to player
	m.Path = pathing.AStar(level, m.TileX, m.TileY, playerX, playerY)

	// Drop the first tile if it's current position
	if len(m.Path) > 0 && m.Path[0].X == m.TileX && m.Path[0].Y == m.TileY {
		m.Path = m.Path[1:]
	}
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

	x, y := isoToScreen(m.TileX, m.TileY)

	op := &ebiten.DrawImageOptions{}
	bounds := m.Sprite.Bounds()
	spriteW := float64(bounds.Dx())
	spriteH := float64(bounds.Dy())

	const verticalOffset = 0.5

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
