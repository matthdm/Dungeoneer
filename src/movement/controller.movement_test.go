package movement

import (
	"dungeoneer/pathing"
	"math"
	"testing"
)

func TestNewMovementController(t *testing.T) {
	speed := 5.0
	mc := NewMovementController(speed)
	
	if mc.Speed != speed {
		t.Errorf("expected speed %v, got %v", speed, mc.Speed)
	}
	if mc.Duration != 15 {
		t.Errorf("expected default duration 15, got %v", mc.Duration)
	}
}

func TestSetVelocityMode(t *testing.T) {
	mc := NewMovementController(5.0)

	// Test non-zero velocity
	mc.SetVelocityMode(3, 4)
	if mc.Mode != VelocityMode {
		t.Errorf("expected mode to be VelocityMode")
	}
	if !mc.Moving {
		t.Errorf("expected Moving to be true")
	}
	// hypot(3, 4) = 5
	// dx/mag = 3/5 = 0.6
	// dy/mag = 4/5 = 0.8
	// speed = 5.0
	// vx = 0.6 * 5.0 = 3.0
	// vy = 0.8 * 5.0 = 4.0
	if math.Abs(mc.VelocityX-3.0) > 1e-9 {
		t.Errorf("expected VelocityX 3.0, got %v", mc.VelocityX)
	}
	if math.Abs(mc.VelocityY-4.0) > 1e-9 {
		t.Errorf("expected VelocityY 4.0, got %v", mc.VelocityY)
	}

	// Test zero velocity
	mc.SetVelocityMode(0, 0)
	if mc.VelocityX != 0 || mc.VelocityY != 0 {
		t.Errorf("expected velocity to be (0,0)")
	}
	if mc.Moving {
		t.Errorf("expected Moving to be false")
	}
}

func TestSetVelocityFromInput(t *testing.T) {
	mc := NewMovementController(10.0)

	// Test non-zero input
	mc.SetVelocityFromInput(1, 0)
	if mc.Mode != VelocityMode {
		t.Errorf("expected mode to be VelocityMode")
	}
	if !mc.Moving {
		t.Errorf("expected Moving to be true")
	}
	if mc.VelocityX != 10.0 || mc.VelocityY != 0 {
		t.Errorf("expected velocity to be (10.0, 0), got (%v, %v)", mc.VelocityX, mc.VelocityY)
	}

	// Test zero input
	mc.SetVelocityFromInput(0, 0)
	if mc.Moving {
		t.Errorf("expected Moving to be false")
	}
	if mc.VelocityX != 0 || mc.VelocityY != 0 {
		t.Errorf("expected velocity to be (0,0)")
	}
}

func TestSetPath(t *testing.T) {
	mc := NewMovementController(5.0)

	path := []pathing.PathNode{
		{X: 1, Y: 1},
		{X: 2, Y: 1},
	}

	// Set initial path
	mc.SetPath(path)
	if mc.Mode != PathingMode {
		t.Errorf("expected mode to be PathingMode")
	}
	if mc.Moving {
		t.Errorf("expected Moving to be false after setting a new path")
	}
	if len(mc.Path) != 2 {
		t.Errorf("expected path length 2, got %v", len(mc.Path))
	}

	// Start moving to the first node
	mc.Update(1.0) // This will pop the first node and start moving
	if !mc.Moving {
		t.Errorf("expected to be moving after update")
	}
	if mc.TargetX != 1 || mc.TargetY != 1 {
		t.Errorf("expected Target to be (1, 1), got (%v, %v)", mc.TargetX, mc.TargetY)
	}

	// Set path again with the same first node while moving
	newPath := []pathing.PathNode{
		{X: 1, Y: 1},
		{X: 1, Y: 2},
	}
	mc.SetPath(newPath)
	if !mc.Moving {
		t.Errorf("expected Moving to remain true if continuing same step")
	}
	if len(mc.Path) != 1 {
		t.Errorf("expected path length 1 (tail only), got %v", len(mc.Path))
	}
	if mc.Path[0].Y != 2 {
		t.Errorf("expected next node to be (1, 2)")
	}
}

func TestStop(t *testing.T) {
	mc := NewMovementController(5.0)
	mc.VelocityX = 5
	mc.VelocityY = 5
	mc.Moving = true

	mc.Stop()
	if mc.VelocityX != 0 || mc.VelocityY != 0 {
		t.Errorf("expected velocity to be 0")
	}
	if mc.Moving {
		t.Errorf("expected Moving to be false")
	}
}

func TestUpdateVelocityMode(t *testing.T) {
	mc := NewMovementController(5.0)
	mc.Mode = VelocityMode
	mc.InterpX = 10
	mc.InterpY = 10
	mc.VelocityX = 2
	mc.VelocityY = -3

	dt := 0.5
	mc.Update(dt)

	if mc.InterpX != 11 {
		t.Errorf("expected InterpX to be 11, got %v", mc.InterpX)
	}
	if mc.InterpY != 8.5 {
		t.Errorf("expected InterpY to be 8.5, got %v", mc.InterpY)
	}
}

func TestUpdatePathingMode(t *testing.T) {
	mc := NewMovementController(5.0)
	mc.Mode = PathingMode
	mc.Duration = 10
	mc.InterpX = 0
	mc.InterpY = 0

	var steppedX, steppedY int
	mc.OnStep = func(x, y int) {
		steppedX = x
		steppedY = y
	}

	mc.SetPath([]pathing.PathNode{{X: 1, Y: 0}})
	
	// First update should pull the node and start moving
	mc.Update(1.0)
	if mc.StartX != 0 || mc.StartY != 0 {
		t.Errorf("expected Start (0,0)")
	}
	if mc.TargetX != 1 || mc.TargetY != 0 {
		t.Errorf("expected Target (1,0)")
	}
	if !mc.Moving {
		t.Errorf("expected to be Moving")
	}

	// Update halfway through (Duration = 10)
	for i := 0; i < 5; i++ {
		mc.Update(1.0)
	}
	
	if math.Abs(mc.InterpX-0.5) > 1e-9 {
		t.Errorf("expected InterpX 0.5, got %v", mc.InterpX)
	}
	if mc.Moving != true {
		t.Errorf("expected to still be moving")
	}

	// Finish movement
	for i := 0; i < 5; i++ {
		mc.Update(1.0)
	}

	if mc.InterpX != 1.0 {
		t.Errorf("expected InterpX 1.0, got %v", mc.InterpX)
	}
	if mc.Moving {
		t.Errorf("expected Moving to be false")
	}
	if steppedX != 1 || steppedY != 0 {
		t.Errorf("expected OnStep to be called with (1, 0)")
	}
}

func TestUpdatePathingModeOvershoot(t *testing.T) {
	mc := NewMovementController(5.0)
	mc.Mode = PathingMode
	mc.Duration = 5
	mc.InterpX = 0
	mc.InterpY = 0

	mc.SetPath([]pathing.PathNode{{X: 1, Y: 0}})
	mc.Update(1.0) // Start moving

	// Force InterpTicks past Duration to simulate an overshoot or long delay
	mc.InterpTicks = 10
	mc.Update(1.0)

	if mc.InterpX != 1.0 {
		t.Errorf("expected InterpX to be clamped to Target (1.0), got %v", mc.InterpX)
	}
	if mc.Moving {
		t.Errorf("expected Moving to be false after completion")
	}
}
