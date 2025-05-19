package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type MonsterBehavior interface {
	Update(m *Monster, player *Player, level *levels.Level)
}

//AggressiveBehavior
//AmbushBehavior
//PatrolBehavior
//BossBehavior

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
	Behavior       MonsterBehavior
	HP             int
	MaxHP          int
	Damage         int
	AttackRate     int
	AttackTick     int
	IsDead         bool
	FlashTick      int
	FlashTicksLeft int // total ticks remaining for the flash effect

}

func NewMonster(ss *sprites.SpriteSheet) []*Monster {
	return []*Monster{
		{
			Name:             "Statue",
			TileX:            5,
			TileY:            7,
			InterpX:          5,
			InterpY:          7,
			Sprite:           ss.Statue, // swap in any sprite
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
	if m.Behavior != nil {
		m.Behavior.Update(m, player, level)
	}
	m.UpdateFlashStatus()
	m.CombatCheck(player)
}

func (m *Monster) BasicChaseLogic(p *Player, level *levels.Level) {
	const bobAmplitude = 1.5
	const bobFrequency = 0.15

	//m.TickCount++
	m.AttackTick++
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

	// Check for path recompute, pathfinding
	needRecalc := len(m.Path) == 0 || !level.IsWalkable(m.Path[0].X, m.Path[0].Y)
	if needRecalc {
		// Find a walkable adjacent tile near the player
		adjTargets := []struct{ X, Y int }{
			{p.TileX + 1, p.TileY},
			{p.TileX - 1, p.TileY},
			{p.TileX, p.TileY + 1},
			{p.TileX, p.TileY - 1},
		}

		found := false
		for _, target := range adjTargets {
			if level.IsWalkable(target.X, target.Y) {
				m.Path = pathing.AStar(level, m.TileX, m.TileY, target.X, target.Y)
				found = true
				break
			}
		}

		if !found {
			// Nowhere to go
			m.Path = nil
		}
		m.RecalcCooldown = 30
		if len(m.Path) > 0 && m.Path[0].X == m.TileX && m.Path[0].Y == m.TileY {
			m.Path = m.Path[1:]
		}
	}
	// Move to next tile
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
}

func (m *Monster) CombatCheck(player *Player) {
	if !m.IsDead && !player.IsDead &&
		!m.Moving &&
		IsAdjacent(m.TileX, m.TileY, player.TileX, player.TileY) {
		m.AttackTick++
		if m.AttackTick >= m.AttackRate {
			player.TakeDamage(m.Damage)
			m.AttackTick = 0
		}
	}
}

func (m *Monster) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if m.Sprite == nil || m.IsDead {
		return
	}

	x, y := isoToScreenFloat(m.InterpX, m.InterpY, tileSize)

	op := &ebiten.DrawImageOptions{}
	bounds := m.Sprite.Bounds()
	spriteW := float64(bounds.Dx())

	// Then apply the bob
	const verticalOffset = 1.0 // tweak until it feels good
	op.GeoM.Translate(0, -verticalOffset+m.BobOffset)

	// Flip for facing direction
	if !m.LeftFacing {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteW, 0)
	}

	// Move to screen space and apply camera
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	if m.FlashTicksLeft > 0 {
		op.ColorScale.Scale(1, 1, 1, 0.7) // Brighter flash
	}

	screen.DrawImage(m.Sprite, op)
	m.UpdateHealthBar(screen, x, y, camX, camY, camScale, cx, cy)

}

func (m *Monster) UpdateHealthBar(screen *ebiten.Image, x, y float64, camX, camY, camScale, cx, cy float64) {
	if !m.IsDead || m.MaxHP > 0 {
		hpPercent := float64(m.HP) / float64(m.MaxHP)
		barWidth := 32.0
		barHeight := 4.0

		hpBar := ebiten.NewImage(int(barWidth), int(barHeight))
		hpBar.Fill(color.RGBA{255, 0, 0, 255})
		hpBarFilled := ebiten.NewImage(int(barWidth*hpPercent), int(barHeight))
		hpBarFilled.Fill(color.RGBA{0, 255, 0, 255})

		// Position HP bar
		barOp := &ebiten.DrawImageOptions{}
		barOp.GeoM.Translate(x-barWidth/2+35, y-10)
		barOp.GeoM.Translate(-camX, camY)
		barOp.GeoM.Scale(camScale, camScale)
		barOp.GeoM.Translate(cx, cy)

		screen.DrawImage(hpBar, barOp)
		screen.DrawImage(hpBarFilled, barOp)
	}
}

func (m *Monster) UpdateFlashStatus() {
	// Flash logic
	if m.FlashTicksLeft > 0 {
		m.FlashTicksLeft--
	}
}

func (m *Monster) TakeDamage(dmg int, markers *[]HitMarker, damageNumbers *[]DamageNumber) {
	m.HP -= dmg
	if m.HP <= 0 {
		m.IsDead = true
	} else {
		m.FlashTicksLeft = 15 // e.g. 15 ticks = flicker for 0.25s at 60fps
	}

	// Add red X marker on hit
	*markers = append(*markers, HitMarker{
		X:        m.InterpX,
		Y:        m.InterpY,
		Ticks:    0,
		MaxTicks: 30, // 0.5 seconds at 60 TPS
	})

	// Add damage number
	*damageNumbers = append(*damageNumbers, DamageNumber{
		X:        float64(m.TileX),
		Y:        float64(m.TileY),
		Value:    dmg,
		Ticks:    0,
		MaxTicks: 30,
	})

}
