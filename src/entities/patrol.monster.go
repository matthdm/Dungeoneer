package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
)

// PatrolWaypoint is a position in a patrol route.
type PatrolWaypoint struct {
	X, Y int
}

// PatrolBehavior walks between waypoints, then chases when the player is near.
type PatrolBehavior struct {
	Waypoints     []PatrolWaypoint
	CurrentWP     int
	Triggered     bool
	TriggerRadius int
	PauseTicks    int // wait at each waypoint
	PauseCounter  int
}

// NewPatrolBehavior creates a PatrolBehavior with the given trigger radius.
// Waypoints should be set by the encounter spawner after creation.
func NewPatrolBehavior(triggerRadius int) *PatrolBehavior {
	return &PatrolBehavior{
		TriggerRadius: triggerRadius,
		PauseTicks:    30,
	}
}

func (pb *PatrolBehavior) Update(m *Monster, p *Player, level *levels.Level) {
	if m.IsDead || m.Moving {
		return
	}

	// Trigger check.
	dx := m.TileX - p.TileX
	dy := m.TileY - p.TileY
	if !pb.Triggered && dx*dx+dy*dy <= pb.TriggerRadius*pb.TriggerRadius {
		pb.Triggered = true
	}

	if pb.Triggered {
		m.BasicChaseLogic(p, level)
		return
	}

	// Patrol between waypoints.
	if len(pb.Waypoints) == 0 {
		return
	}

	wp := pb.Waypoints[pb.CurrentWP]
	if m.TileX == wp.X && m.TileY == wp.Y {
		// Arrived at waypoint, pause then advance.
		pb.PauseCounter++
		if pb.PauseCounter >= pb.PauseTicks {
			pb.PauseCounter = 0
			pb.CurrentWP = (pb.CurrentWP + 1) % len(pb.Waypoints)
		}
		return
	}

	// Walk toward current waypoint.
	if len(m.Path) == 0 || m.RecalcCooldown <= 0 {
		m.Path = pathing.AStar(level, m.TileX, m.TileY, wp.X, wp.Y)
		m.RecalcCooldown = 30
		if len(m.Path) > 0 && m.Path[0].X == m.TileX && m.Path[0].Y == m.TileY {
			m.Path = m.Path[1:]
		}
	}
	m.RecalcCooldown--

	if len(m.Path) > 0 {
		next := m.Path[0]
		if level.IsWalkable(next.X, next.Y) {
			m.MoveTo(next.X, next.Y)
			m.Path = m.Path[1:]
		} else {
			m.Path = nil
		}
	}
}
