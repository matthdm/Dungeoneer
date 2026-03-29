package entities

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"math"
)

// BossAttack describes one attack in a boss's pattern set.
type BossAttack struct {
	ID         string
	Type       string // "melee", "aoe", "summon"
	Damage     int
	Range      float64 // tiles
	Cooldown   float64 // seconds between uses
	WindupTime float64 // telegraph duration
	AOERadius  int     // for area attacks (in tiles)
	Timer      float64 // current cooldown accumulator
}

// Boss extends Monster with multi-phase combat.
type Boss struct {
	*Monster
	Title        string
	CurrentPhase int
	MaxPhases    int
	PhaseHP      []float64      // HP percentage thresholds for transitions (e.g. [0.5] = transition at 50%)
	Patterns     [][]BossAttack // attack patterns per phase
	IsActive     bool           // fight has begun (arena sealed)
	InTransition bool
	TransTimer   float64
}

// BossBehavior implements MonsterBehavior for the Boss entity.
type BossBehavior struct {
	Boss *Boss
}

// NewBossBehavior creates a boss-aware behavior.
func NewBossBehavior(b *Boss) *BossBehavior {
	return &BossBehavior{Boss: b}
}

func (bb *BossBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	b := bb.Boss
	if m.IsDead || !b.IsActive {
		return
	}

	// Phase transition.
	if b.InTransition {
		b.TransTimer -= 1.0 / 60.0
		if b.TransTimer <= 0 {
			b.InTransition = false
		}
		return
	}

	// Check for phase transition.
	hpPct := float64(m.HP) / float64(m.MaxHP)
	if b.CurrentPhase < b.MaxPhases-1 && b.CurrentPhase < len(b.PhaseHP) {
		if hpPct <= b.PhaseHP[b.CurrentPhase] {
			b.CurrentPhase++
			b.InTransition = true
			b.TransTimer = 1.0 // 1 second pause
			m.FlashTicksLeft = 60
			return
		}
	}

	// Get current phase patterns.
	var attacks []BossAttack
	if b.CurrentPhase < len(b.Patterns) {
		attacks = b.Patterns[b.CurrentPhase]
	}

	// Try each attack.
	dx := float64(m.TileX - p.TileX)
	dy := float64(m.TileY - p.TileY)
	dist := math.Sqrt(dx*dx + dy*dy)

	for i := range attacks {
		a := &attacks[i]
		a.Timer += 1.0 / 60.0
		if a.Timer < a.Cooldown {
			continue
		}

		switch a.Type {
		case "aoe":
			if dist <= float64(a.AOERadius)+1 {
				// AoE slam: damage player if within radius.
				a.Timer = 0
				if dist <= float64(a.AOERadius) {
					p.TakeDamage(a.Damage)
				}
			}
		case "melee":
			if IsAdjacent(m.TileX, m.TileY, p.TileX, p.TileY) {
				a.Timer = 0
				p.TakeDamage(a.Damage)
			}
		}
	}

	// Chase player (basic movement).
	if !m.Moving {
		m.BasicChaseLogic(p, level)
	}
}

// NewDungeonGuardian creates the fallback boss when no NPC has ascended.
func NewDungeonGuardian(ss *sprites.SpriteSheet, x, y int) *Boss {
	m := &Monster{
		Name:             "The Warden",
		TileX:            x,
		TileY:            y,
		InterpX:          float64(x),
		InterpY:          float64(y),
		Sprite:           ss.Death,
		MovementDuration: 20,
		LeftFacing:       true,
		HP:               200,
		MaxHP:            200,
		Damage:           15,
		HitRadius:        DefaultMonsterHitRadius,
		AttackRate:       30,
		Level:            10,
		Role:             "boss",
	}
	boss := &Boss{
		Monster:   m,
		Title:     "Hollow Sentinel of the Deep",
		MaxPhases: 2,
		PhaseHP:   []float64{0.5}, // transition at 50%
		Patterns: [][]BossAttack{
			// Phase 1: melee swings + periodic AoE slam.
			{
				{ID: "swing", Type: "melee", Damage: 12, Cooldown: 1.0},
				{ID: "slam", Type: "aoe", Damage: 8, AOERadius: 2, Cooldown: 5.0},
			},
			// Phase 2: faster melee, bigger AoE.
			{
				{ID: "swing", Type: "melee", Damage: 18, Cooldown: 0.7},
				{ID: "slam", Type: "aoe", Damage: 12, AOERadius: 3, Cooldown: 3.0},
			},
		},
	}
	m.Behavior = NewBossBehavior(boss)
	return boss
}
