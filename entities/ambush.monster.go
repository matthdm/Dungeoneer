package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"math"
)

type AmbushBehavior struct {
	Triggered     bool
	TriggerRadius int
}

func NewAmbushBehavior(radius int) *AmbushBehavior {
	return &AmbushBehavior{
		TriggerRadius: radius,
		Triggered:     false,
	}
}

func NewStatueMonster(ss *sprites.SpriteSheet) []*Monster {
	return []*Monster{
		{
			Name:             "Statue Monster",
			TileX:            10,
			TileY:            10,
			InterpX:          10,
			InterpY:          10,
			Sprite:           ss.Statue,
			MovementDuration: 30,
			LeftFacing:       true,
			HP:               8,
			MaxHP:            8,
			Damage:           2,
			AttackRate:       45,
			Behavior:         NewAmbushBehavior(4), // Trigger when player is within 2 tiles
		},
	}
}

func (b *AmbushBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead {
		return
	}

	if !b.Triggered {
		// Check if player is close enough to activate
		dx := m.TileX - p.TileX
		dy := m.TileY - p.TileY
		distSq := dx*dx + dy*dy
		if distSq <= b.TriggerRadius*b.TriggerRadius {
			b.Triggered = true
		} else {
			return // stay dormant
		}
	}

	// Once triggered, behave like an aggressive chaser
	m.BasicChaseLogic(p, level)
}

