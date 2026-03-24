package entities

import (
	"dungeoneer/levels"
	"math"
	"math/rand/v2"
)

// SwarmBehavior makes monsters cluster together and attack as a group.
type SwarmBehavior struct {
	Triggered     bool
	TriggerRadius int
	WanderCounter int
}

// NewSwarmBehavior creates a SwarmBehavior with the given trigger radius.
func NewSwarmBehavior(triggerRadius int) *SwarmBehavior {
	return &SwarmBehavior{
		TriggerRadius: triggerRadius,
	}
}

func (s *SwarmBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead || m.Moving {
		return
	}

	m.TickCount++
	m.AttackTick++
	m.BobOffset = math.Sin(float64(m.TickCount)*0.15) * 1.5

	dx := m.TileX - p.TileX
	dy := m.TileY - p.TileY

	// Self-trigger on proximity.
	if !s.Triggered && dx*dx+dy*dy <= s.TriggerRadius*s.TriggerRadius {
		s.Triggered = true
	}

	// Propagate trigger to siblings.
	if !s.Triggered {
		for _, sib := range m.Siblings {
			if sib.IsDead {
				continue
			}
			if sb, ok := sib.Behavior.(*SwarmBehavior); ok && sb.Triggered {
				s.Triggered = true
				break
			}
		}
	}

	if !s.Triggered {
		// Wander slowly near group centroid.
		s.WanderCounter++
		if s.WanderCounter < 15 {
			return
		}
		s.WanderCounter = 0
		s.wanderNearGroup(m, level)
		return
	}

	// Triggered: chase player, biased toward siblings.
	s.chaseWithGroupBias(m, p, level)
	m.CombatCheck(p)
}

// wanderNearGroup moves randomly but stays close to living siblings.
func (s *SwarmBehavior) wanderNearGroup(m *Monster, level *levels.Level) {
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		nx, ny := m.TileX+d[0], m.TileY+d[1]
		if level.IsWalkable(nx, ny) {
			m.MoveTo(nx, ny)
			return
		}
	}
}

// chaseWithGroupBias picks the adjacent tile that minimises distance to the
// player while keeping the monster within a few tiles of siblings.
func (s *SwarmBehavior) chaseWithGroupBias(m *Monster, p *Player, level *levels.Level) {
	// Compute centroid of living siblings.
	cx, cy := float64(m.TileX), float64(m.TileY)
	count := 1.0
	for _, sib := range m.Siblings {
		if sib.IsDead || sib == m {
			continue
		}
		cx += float64(sib.TileX)
		cy += float64(sib.TileY)
		count++
	}
	cx /= count
	cy /= count

	best := math.MaxFloat64
	bestX, bestY := m.TileX, m.TileY
	for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := m.TileX+d[0], m.TileY+d[1]
		if !level.IsWalkable(nx, ny) {
			continue
		}
		// Weighted score: mostly player distance, slight bias toward centroid.
		pDist := math.Abs(float64(nx-p.TileX)) + math.Abs(float64(ny-p.TileY))
		gDist := math.Abs(float64(nx)-cx) + math.Abs(float64(ny)-cy)
		score := pDist + gDist*0.3
		if score < best {
			best = score
			bestX = nx
			bestY = ny
		}
	}
	if bestX != m.TileX || bestY != m.TileY {
		m.MoveTo(bestX, bestY)
	}
}
