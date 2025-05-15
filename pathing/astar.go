package pathing

import "dungeoneer/levels"

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
	return abs(x1-x2) + abs(y1-y2)
}

func reconstructPath(end *Node) []PathNode {
	var path []PathNode
	for n := end; n != nil; n = n.Parent {
		path = append([]PathNode{{X: n.X, Y: n.Y}}, path...)
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

		// Explore neighbors
		for _, dir := range [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
			nx, ny := current.X+dir[0], current.Y+dir[1]

			if !level.IsWalkable(nx, ny) || closed[[2]int{nx, ny}] {
				continue
			}

			// Check if in open list
			inOpen := false
			for _, n := range open {
				if n.X == nx && n.Y == ny {
					inOpen = true
					break
				}
			}
			if inOpen {
				continue
			}

			n := &Node{
				X:      nx,
				Y:      ny,
				GCost:  current.GCost + 1,
				HCost:  heuristic(nx, ny, goalX, goalY),
				Parent: current,
			}
			n.FCost = n.GCost + n.HCost
			open = append(open, n)
		}
	}

	return nil // No path
}
