package entities

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"math/rand"
)

type RoamingWanderBehavior struct {
	MoveCooldown  int
	Counter       int
	Triggered     bool
	TriggerRadius int
}

func NewRoamingWanderBehavior(radius int) *RoamingWanderBehavior {
	return &RoamingWanderBehavior{
		MoveCooldown:  10,
		Counter:       0,
		TriggerRadius: radius,
		Triggered:     false,
	}
}

func NewRoamingMonster(ss *sprites.SpriteSheet) []*Monster {
	return []*Monster{
		{
			Name:             "Blue Man Monster",
			TileX:            10,
			TileY:            10,
			InterpX:          10,
			InterpY:          10,
			Sprite:           ss.BlueMan,
			MovementDuration: 30,
			LeftFacing:       true,
			HP:               8,
			MaxHP:            8,
			Damage:           2,
			AttackRate:       45,
			Behavior:         NewRoamingWanderBehavior(4),
		},
	}
}

func (r *RoamingWanderBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	//m.TickCount++
	if m.IsDead || m.Moving {
		return
	}

	r.Counter++
	if r.Counter < r.MoveCooldown {
		return
	}
	r.Counter = 0

	// Try a random direction
	directions := []struct{ X, Y int }{
		{0, -1}, {0, 1}, {-1, 0}, {1, 0},
	}
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, d := range directions {
		newX := m.TileX + d.X
		newY := m.TileY + d.Y
		if level.IsWalkable(newX, newY) {
			m.MoveTo(newX, newY)
			break
		}
	}

	if !r.Triggered {
		// Check if player is close enough to activate
		dx := m.TileX - p.TileX
		dy := m.TileY - p.TileY
		distSq := dx*dx + dy*dy
		if distSq <= r.TriggerRadius*r.TriggerRadius {
			r.Triggered = true
		} else {
			return // stay dormant
		}
	}

	// Once triggered, behave like an aggressive chaser
	if r.Triggered {
		m.BasicChaseLogic(p, level)
	}

}
