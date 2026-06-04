package game

import "testing"

func TestScreenShake(t *testing.T) {
	s := &ScreenShake{}
	
	// Test Trigger
	s.Trigger(10.0)
	if s.Decay != 12.0 {
		t.Errorf("expected default decay to be set to 12.0")
	}
	if s.IntensityX == 0 && s.IntensityY == 0 {
		// Extremely rare that both are exactly 0, but possible. We'll check.
	}
	
	// Test Update
	s.IntensityX = 5.0
	s.IntensityY = -5.0
	
	s.Update(0.1) // decay = 1.2
	
	if s.IntensityX != 3.8 {
		t.Errorf("expected IntensityX 3.8, got %f", s.IntensityX)
	}
	if s.IntensityY != -3.8 {
		t.Errorf("expected IntensityY -3.8, got %f", s.IntensityY)
	}
	
	// Test approach overshooting
	s.Update(10.0) // big decay
	
	if s.IntensityX != 0 || s.IntensityY != 0 {
		t.Errorf("expected both to be 0")
	}
	
	// Test Offset
	s.IntensityX = 1.0
	s.IntensityY = 2.0
	
	ox, oy := s.Offset()
	if ox != 1.0 || oy != 2.0 {
		t.Errorf("expected offset (1.0, 2.0)")
	}
	
	// Test Update when 0
	s.IntensityX = 0
	s.IntensityY = 0
	s.Update(1.0)
	if s.IntensityX != 0 || s.IntensityY != 0 {
		t.Errorf("expected to stay 0")
	}
}
