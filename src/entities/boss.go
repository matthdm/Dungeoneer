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

	// NPC ascension fields
	NPCID               string    // set if this boss is an ascended NPC; used for post-defeat MetaSave updates
	PreFightDialogueID  string    // dialogue tree shown once when player enters the arena
	PreFightShown       bool      // true once the pre-fight dialogue has fired
	PostFightDialogueID string    // dialogue tree shown once when the boss is defeated; portal deferred until it closes
	OnPhaseTransition   func(int) // called with new phase index on every phase change (sprite swaps, etc.)

	// Pull-line visual: set when a chain pull fires, counts down to 0.
	PullLineTicks int
	PullLineX     int // player tile X at time of pull (chain origin)
	PullLineY     int // player tile Y at time of pull (chain origin)
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
	// Decrement pull-line visual timer each tick regardless of active state.
	if b.PullLineTicks > 0 {
		b.PullLineTicks--
	}
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
			b.TransTimer = 1.5 // 1.5 second dramatic pause
			m.FlashTicksLeft = 90
			if b.OnPhaseTransition != nil {
				b.OnPhaseTransition(b.CurrentPhase)
			}
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
				a.Timer = 0
				if dist <= float64(a.AOERadius) {
					p.TakeDamage(a.Damage)
				}
			}
		case "melee":
			// Range field overrides default adjacency check when non-zero.
			r := a.Range
			if r <= 0 {
				r = 1.5
			}
			if dist <= r {
				a.Timer = 0
				p.TakeDamage(a.Damage)
			}
		case "ranged":
			// Ranged attack: fires if player is within a.Range tiles.
			if dist <= a.Range && dist > 1 {
				a.Timer = 0
				p.TakeDamage(a.Damage)
			}
		case "projectile":
			// Projectile attack: fires a moving projectile toward the player.
			if dist <= a.Range && dist > 1 {
				a.Timer = 0
				proj := NewMonsterProjectile(
					float64(m.TileX), float64(m.TileY),
					float64(p.TileX), float64(p.TileY),
					0.12, a.Damage,
				)
				m.PendingProjectiles = append(m.PendingProjectiles, proj)
			}
		case "pull_player":
			// Chain pull: teleports player to the tile adjacent to the boss,
			// then deals damage (the follow-up chain whip lands immediately).
			if dist <= a.Range && dist > 1 {
				a.Timer = 0
				b.PullLineTicks = 14
				b.PullLineX = p.TileX
				b.PullLineY = p.TileY
				bossChainPull(m, p, level)
				p.TakeDamage(a.Damage)
			}
		}
	}

	// Chase player (basic movement).
	if !m.Moving {
		m.BasicChaseLogic(p, level)
	}
}

// bossChainPull teleports the player to the tile directly adjacent to the boss
// in the direction of the player, cancelling all in-progress movement.
// If the computed target tile is not walkable, it falls back to the nearest
// walkable tile adjacent to the boss so the player never ends up inside a wall.
func bossChainPull(m *Monster, p *Player, level *levels.Level) {
	dx := p.TileX - m.TileX
	dy := p.TileY - m.TileY
	if dx == 0 && dy == 0 {
		return
	}
	var tx, ty int
	if absi(dx) >= absi(dy) {
		if dx > 0 {
			tx = m.TileX + 1
		} else {
			tx = m.TileX - 1
		}
		ty = m.TileY
	} else {
		tx = m.TileX
		if dy > 0 {
			ty = m.TileY + 1
		} else {
			ty = m.TileY - 1
		}
	}
	// Ensure destination is walkable; spiral through adjacent tiles if not.
	if level != nil && !level.IsWalkable(tx, ty) {
		fx, fy := bossNearestWalkableAdjacent(m, level)
		if fx < 0 {
			return // no safe tile found — skip pull
		}
		tx, ty = fx, fy
	}
	p.TileX = tx
	p.TileY = ty
	p.MoveController.InterpX = float64(tx)
	p.MoveController.InterpY = float64(ty)
	p.MoveController.TargetX = float64(tx)
	p.MoveController.TargetY = float64(ty)
	p.MoveController.Path = nil
	p.MoveController.Stop()
	p.CollisionBox.X = float64(tx)
	p.CollisionBox.Y = float64(ty)
}

// bossNearestWalkableAdjacent returns the first walkable tile in the 8
// cardinal/diagonal neighbors of the boss, or (-1,-1) if none exist.
func bossNearestWalkableAdjacent(m *Monster, level *levels.Level) (int, int) {
	for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		nx, ny := m.TileX+d[0], m.TileY+d[1]
		if level.IsWalkable(nx, ny) {
			return nx, ny
		}
	}
	return -1, -1
}

func absi(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
