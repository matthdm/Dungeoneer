// Package pathing implements grid-based A* pathfinding used by entities.
//
// Implementation notes:
//   - Movement supports 8 directions (4 orthogonal + 4 diagonal). Orthogonal
//     steps cost 10, diagonal steps cost 14 (≈ 10√2). The heuristic is octile
//     distance, which is admissible under these costs.
//   - The A* here uses simple slices and maps for open/closed lists. For large
//     maps or frequent queries, replacing `open` with a binary heap and
//     reusing node objects from a pool will reduce allocations and improve
//     performance (expected O(N log N) with a heap).
package pathing

import (
	"dungeoneer/levels"
	"dungeoneer/tiles"
)

type Node struct {
	X, Y   int
	GCost  int // Cost from start
	HCost  int // Heuristic to goal
	FCost  int // G + H
	Parent *Node
}

type PathNode struct {
	X, Y int
}

func heuristic(x1, y1, x2, y2 int) int {
	dx := abs(x1 - x2)
	dy := abs(y1 - y2)
	// Octile distance scaled to match movement costs (orthogonal=10, diagonal=14).
	// Formula: 10*max + 4*min  (derived from 10*(dx+dy) + (14-20)*min)
	if dx > dy {
		return 10*dx + 4*dy
	}
	return 10*dy + 4*dx
}

func reconstructPath(end *Node) []PathNode {
	var path []PathNode
	for n := end; n != nil; n = n.Parent {
		path = append([]PathNode{{X: n.X, Y: n.Y}}, path...)
	}
	// Drop the start node — callers are already at that position.
	// Including it caused the controller to interpolate back to the start
	// tile when a new path was issued mid-step.
	if len(path) > 0 {
		path = path[1:]
	}
	return path
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func AStar(level *levels.Level, startX, startY, goalX, goalY int) []PathNode {
	if goalX < 0 || goalY < 0 || goalX >= level.W || goalY >= level.H {
		return nil
	}
	// Note: `open` is a simple slice of *Node. This code scans the slice to
	// find the smallest FCost which is O(n) per extraction. Switching to a
	// priority queue (binary heap) avoids this linear scan and is recommended
	// for production workloads.
	open := []*Node{}
	closed := map[[2]int]bool{}

	start := &Node{X: startX, Y: startY}
	start.HCost = heuristic(startX, startY, goalX, goalY)
	start.FCost = start.HCost

	open = append(open, start)

	var current *Node

	for len(open) > 0 {
		// Get node with lowest F cost
		currentIndex := 0
		current = open[0]
		for i, n := range open {
			if n.FCost < current.FCost {
				current = n
				currentIndex = i
			}
		}

		// Remove from open
		open = append(open[:currentIndex], open[currentIndex+1:]...)
		closed[[2]int{current.X, current.Y}] = true

		// Reached goal
		if current.X == goalX && current.Y == goalY {
			return reconstructPath(current)
		}

		// Explore neighbors: 4 orthogonal (cost 10) + 4 diagonal (cost 14).
		// [dx, dy, cost]
		for _, dir := range [8][3]int{
			{0, 1, 10}, {1, 0, 10}, {0, -1, 10}, {-1, 0, 10},
			{1, 1, 14}, {1, -1, 14}, {-1, 1, 14}, {-1, -1, 14},
		} {
			dx, dy, moveCost := dir[0], dir[1], dir[2]
			nx, ny := current.X+dx, current.Y+dy

			if !level.IsWalkable(nx, ny) || closed[[2]int{nx, ny}] {
				continue
			}

			// Check for closed/locked doors (they block movement even if tile is walkable)
			tile := level.Tile(nx, ny)
			if tile != nil && tile.HasTag(tiles.TagDoor) && (tile.DoorState == 2 || tile.DoorState == 3) {
				continue
			}

			// Corner-cutting prevention: a diagonal step is only valid when both
			// orthogonal neighbours are clear. This stops the player squeezing
			// through the gap between two touching walls.
			if dx != 0 && dy != 0 {
				if !level.IsWalkable(current.X+dx, current.Y) || !level.IsWalkable(current.X, current.Y+dy) {
					continue
				}
			}

			newGCost := current.GCost + moveCost

			// If already in the open list, only update if we found a cheaper path.
			updated := false
			for _, existing := range open {
				if existing.X == nx && existing.Y == ny {
					if newGCost < existing.GCost {
						existing.GCost = newGCost
						existing.FCost = newGCost + existing.HCost
						existing.Parent = current
					}
					updated = true
					break
				}
			}
			if updated {
				continue
			}

			n := &Node{
				X:      nx,
				Y:      ny,
				GCost:  newGCost,
				HCost:  heuristic(nx, ny, goalX, goalY),
				Parent: current,
			}
			n.FCost = n.GCost + n.HCost
			open = append(open, n)
		}
	}

	return nil // No path
}
