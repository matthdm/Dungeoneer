package spells

import (
	"dungeoneer/coords"
	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// OnSpellImpact is an optional callback invoked when a spell deals damage at a
// world position. Set by the game package at startup to avoid a circular import.
// worldX, worldY are a world ground point (see package coords); spellType matches the
// spell name used in SpellParticleColor (e.g. "fireball", "lightning").
var OnSpellImpact func(worldX, worldY float64, spellType string)

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

// Every spell stores its positions as world ground points (see package
// coords): entity locations come from BodyCenter(), cursor targets from the
// game's cursor pick. Nothing here adds an offset to line a visual up. The
// only choice a Draw method makes is which of these two projections to use.

// groundIso projects a world ground point to isometric pixels. Use it for
// effects that lie on the floor: rings, cracks, ground-targeted impacts.
func groundIso(x, y float64, tileSize int) (float64, float64) {
	return coords.GroundToIso(x, y, tileSize)
}

// bodyIso projects a world ground point and lifts it to body height. Use it
// for effects that travel between or sit on bodies: projectiles, beams,
// melee arcs, impact flashes on a target.
func bodyIso(x, y float64, tileSize int) (float64, float64) {
	sx, sy := coords.GroundToIso(x, y, tileSize)
	return sx, sy - coords.BodyHeightPx(tileSize)
}
