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
			Level:            1,
		},
	}
}

func (r *RoamingWanderBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead || m.Moving {
		return
	}

	if !r.Triggered {
		dx := m.TileX - p.TileX
		dy := m.TileY - p.TileY
		if dx*dx+dy*dy <= r.TriggerRadius*r.TriggerRadius {
			r.Triggered = true
			return // wait 1 frame before chasing
		}

		// Roaming logic
		r.Counter++
		if r.Counter < r.MoveCooldown {
			return
		}
		r.Counter = 0

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
		return
	}

	// Chase logic
	m.BasicChaseLogic(p, level)
}
