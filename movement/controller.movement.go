package movement

import (
	"dungeoneer/pathing"
	"math"
)

type MovementMode int

const (
	PathingMode MovementMode = iota
	VelocityMode
)

type MovementController struct {
	// Shared
	Mode   MovementMode
	Speed  float64
	Moving bool

	// A* Pathing
	Path             []pathing.PathNode
	InterpTicks      int
	Duration         int // ticks per tile
	StartX, StartY   float64
	TargetX, TargetY float64
	OnStep           func(x, y int)

	// Velocity-based
	VelocityX float64
	VelocityY float64

	// Position
	InterpX float64
	InterpY float64
}

// NewMovementController creates a new controller with a given movement speed.
func NewMovementController(speed float64) *MovementController {
	return &MovementController{
		Speed:    speed,
		Duration: 15, // default tile interpolation time
	}
}

// SetVelocityMode activates velocity movement
func (c *MovementController) SetVelocityMode(dx, dy float64) {
	c.Mode = VelocityMode
	mag := math.Hypot(dx, dy)
	if mag > 0 {
		c.VelocityX = dx / mag * c.Speed
		c.VelocityY = dy / mag * c.Speed
		c.Moving = true
	} else {
		c.Stop()
	}
}

// SetVelocityFromInput sets velocity based on input vector and switches to velocity mode.
func (c *MovementController) SetVelocityFromInput(dx, dy float64) {
	if dx == 0 && dy == 0 {
		c.Stop()
		return
	}
	c.Mode = VelocityMode
	mag := math.Hypot(dx, dy)
	c.VelocityX = dx / mag * c.Speed
	c.VelocityY = dy / mag * c.Speed
	c.Moving = true
}

// SetPathingMode assigns a path and begins movement
func (c *MovementController) SetPath(path []pathing.PathNode) {
	c.Path = path
	c.Mode = PathingMode
	c.Moving = false
}

// Stop halts movement
func (c *MovementController) Stop() {
	c.VelocityX = 0
	c.VelocityY = 0
	c.Moving = false
}

// Update moves the entity based on the selected mode.
func (c *MovementController) Update(dt float64) {
	switch c.Mode {
	case VelocityMode:
		c.InterpX += c.VelocityX * dt
		c.InterpY += c.VelocityY * dt
		// Optional: clamp to bounds here or use collision system

	case PathingMode:
		if c.Moving {
			c.InterpTicks++
			t := float64(c.InterpTicks) / float64(c.Duration)
			if t > 1 {
				t = 1
			}
			c.InterpX = c.StartX + (c.TargetX-c.StartX)*t
			c.InterpY = c.StartY + (c.TargetY-c.StartY)*t

			if t >= 1 {
				c.Moving = false
				c.InterpX = c.TargetX
				c.InterpY = c.TargetY
			}
		} else if len(c.Path) > 0 {
			next := c.Path[0]
			c.Path = c.Path[1:]
			c.StartX = c.InterpX
			c.StartY = c.InterpY
			c.TargetX = float64(next.X)
			c.TargetY = float64(next.Y)
			c.InterpTicks = 0
			c.Moving = true

			if c.OnStep != nil {
				c.OnStep(next.X, next.Y)
			}
		}

	}
}
