package spells

import (
	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpellInfo holds data about a spell type.
type SpellInfo struct {
	Name     string
	Level    int
	Cooldown float64
	Damage   int
	Cost     int
}

type Spell interface {
	Update(level *levels.Level, dt float64)
	Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64)
	IsFinished() bool
}

// Caster tracks cooldowns for spells.
type Caster struct {
	Cooldowns map[string]float64
}

func NewCaster() *Caster {
	return &Caster{Cooldowns: make(map[string]float64)}
}

func (c *Caster) Update(dt float64) {
	for k, v := range c.Cooldowns {
		if v > 0 {
			v -= dt
			if v < 0 {
				v = 0
			}
			c.Cooldowns[k] = v
		}
	}
}

func (c *Caster) Ready(info SpellInfo) bool {
	if cd, ok := c.Cooldowns[info.Name]; ok {
		return cd <= 0
	}
	return true
}

func (c *Caster) PutOnCooldown(info SpellInfo) {
	c.Cooldowns[info.Name] = info.Cooldown
}
