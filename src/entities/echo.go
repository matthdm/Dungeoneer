package entities

import (
	"math"
	"math/rand/v2"

	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
)

// NewWickedEcho creates a hostile echo entity — the player's past self turned enemy.
// It wraps a Monster so it participates in existing combat/draw pipelines.
func NewWickedEcho(ss *sprites.SpriteSheet, x, y int, sourceFloor int) *Monster {
	hp := 40 + sourceFloor*10
	m := &Monster{
		Name:             "Wicked Echo",
		TileX:            x,
		TileY:            y,
		InterpX:          float64(x),
		InterpY:          float64(y),
		Sprite:           ss.TorturedSoul,
		MovementDuration: 16,
		LeftFacing:       true,
		HP:               hp,
		MaxHP:            hp,
		Damage:           8 + sourceFloor*2,
		HitRadius:        DefaultMonsterHitRadius,
		AttackRate:       30,
		Level:            sourceFloor,
		Role:             "melee",
		IsEcho:           true,
	}
	m.Behavior = &WickedEchoBehavior{}
	return m
}

// WickedEchoBehavior chases and attacks the player.
type WickedEchoBehavior struct{}

func (b *WickedEchoBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead {
		return
	}
	dx := float64(m.TileX - p.TileX)
	dy := float64(m.TileY - p.TileY)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist <= 1.5 {
		m.AttackTick++
		if m.AttackTick >= m.AttackRate {
			m.AttackTick = 0
			p.TakeDamage(m.Damage)
		}
	} else {
		m.AttackTick = 0
	}
	if !m.Moving {
		m.BasicChaseLogic(p, level)
	}
}

// NewHeroEcho creates an allied echo entity that fights for the player for 60 seconds.
func NewHeroEcho(ss *sprites.SpriteSheet, x, y int, sourceFloor int, monsters *[]*Monster) *Monster {
	hp := 60 + sourceFloor*5
	m := &Monster{
		Name:             "Hero Echo",
		TileX:            x,
		TileY:            y,
		InterpX:          float64(x),
		InterpY:          float64(y),
		Sprite:           ss.TorturedSoul,
		MovementDuration: 14,
		LeftFacing:       false,
		HP:               hp,
		MaxHP:            hp,
		Damage:           6 + sourceFloor,
		HitRadius:        DefaultMonsterHitRadius,
		AttackRate:       35,
		Level:            sourceFloor,
		Role:             "melee",
		IsEcho:           true,
		EchoLifetime:     60.0,
	}
	m.Behavior = &HeroEchoBehavior{Monsters: monsters}
	return m
}

// HeroEchoBehavior follows the player and attacks the nearest hostile monster.
type HeroEchoBehavior struct {
	Monsters *[]*Monster
}

func (b *HeroEchoBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead {
		return
	}
	// Lifetime countdown is ticked by the game loop via EchoLifetime on the monster.
	// When EchoLifetime <= 0, the entity is marked dead by the game loop.

	// Find nearest non-echo, non-dead monster to attack.
	var target *Monster
	bestDist := 5.0 // attack range
	if b.Monsters != nil {
		for _, other := range *b.Monsters {
			if other == nil || other.IsDead || other.IsEcho {
				continue
			}
			dx := float64(m.TileX - other.TileX)
			dy := float64(m.TileY - other.TileY)
			d := math.Sqrt(dx*dx + dy*dy)
			if d < bestDist {
				bestDist = d
				target = other
			}
		}
	}

	if target != nil {
		dx := float64(m.TileX - target.TileX)
		dy := float64(m.TileY - target.TileY)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= 1.5 {
			m.AttackTick++
			if m.AttackTick >= m.AttackRate {
				m.AttackTick = 0
				target.HP -= m.Damage
				if target.HP <= 0 {
					target.IsDead = true
				}
			}
		} else if !m.Moving {
			// Move toward target using AStar.
			path := pathing.AStar(level, m.TileX, m.TileY, target.TileX, target.TileY)
			if len(path) > 1 {
				next := path[1]
				m.MoveTo(next.X, next.Y)
			}
		}
	} else {
		// No enemies — follow player if too far away.
		dx := float64(m.TileX - p.TileX)
		dy := float64(m.TileY - p.TileY)
		if math.Sqrt(dx*dx+dy*dy) > 2.0 && !m.Moving {
			m.BasicChaseLogic(p, level)
		}
	}
}

// NewMemoryFragmentNPC creates a static ghost NPC with a contextual message.
// deathCause is used to generate the message text.
func NewMemoryFragmentNPC(x, y int, deathCause string, ss *sprites.SpriteSheet) *NPC {
	msg := "I fell here. Watch the shadows."
	if deathCause != "" {
		msg = "I was lost to " + deathCause + ". Be wary."
	}
	npc := &NPC{
		Name:          "Memory Fragment",
		TileX:         x,
		TileY:         y,
		InterpX:       float64(x),
		InterpY:       float64(y),
		Interactable:  true,
		IsEcho:        true,
		Sprite:        ss.TorturedSoul,
		HintText:      "[E] Remember",
		DialogueLines: []string{msg},
	}
	return npc
}

// EchoWeightedType picks a random echo type: 0=Wicked, 1=Hero, 2=Memory.
// Weights: 50% Wicked, 30% Hero, 20% Memory.
func EchoWeightedType() int {
	r := rand.Float64()
	if r < 0.50 {
		return 0
	} else if r < 0.80 {
		return 1
	}
	return 2
}
