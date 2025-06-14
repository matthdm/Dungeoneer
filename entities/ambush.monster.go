package entities

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
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
		/*
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
			{
				Name:             "Statue Monster2",
				TileX:            10,
				TileY:            56,
				InterpX:          10,
				InterpY:          56,
				Sprite:           ss.Statue,
				MovementDuration: 30,
				LeftFacing:       true,
				HP:               8,
				MaxHP:            8,
				Damage:           2,
				AttackRate:       45,
				Behavior:         NewAmbushBehavior(4), // Trigger when player is within 2 tiles
			},
			{
				Name:             "Statue Monste3",
				TileX:            40,
				TileY:            40,
				InterpX:          40,
				InterpY:          40,
				Sprite:           ss.Statue,
				MovementDuration: 30,
				LeftFacing:       true,
				HP:               8,
				MaxHP:            8,
				Damage:           2,
				AttackRate:       45,
				Behavior:         NewAmbushBehavior(4), // Trigger when player is within 2 tiles
			},
			{
				Name:             "Statue Monster4",
				TileX:            21,
				TileY:            35,
				InterpX:          21,
				InterpY:          35,
				Sprite:           ss.Statue,
				MovementDuration: 30,
				LeftFacing:       true,
				HP:               8,
				MaxHP:            8,
				Damage:           2,
				AttackRate:       45,
				Behavior:         NewAmbushBehavior(4), // Trigger when player is within 2 tiles
			},
			{
				Name:             "Statue Monster5",
				TileX:            62,
				TileY:            4,
				InterpX:          62,
				InterpY:          4,
				Sprite:           ss.Statue,
				MovementDuration: 30,
				LeftFacing:       true,
				HP:               8,
				MaxHP:            8,
				Damage:           2,
				AttackRate:       45,
				Behavior:         NewAmbushBehavior(4), // Trigger when player is within 2 tiles
			},*/
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
