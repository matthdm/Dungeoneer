package game

import "testing"

func TestTransition(t *testing.T) {
	trans := NewTransition(100, 100)
	
	if trans.Active() {
		t.Errorf("expected transition to be inactive initially")
	}
	
	midpointCalled := false
	trans.Start(func() {
		midpointCalled = true
	})
	
	if !trans.Active() {
		t.Errorf("expected transition to be active after start")
	}
	
	// FadeOut
	trans.Update(0.1)
	if midpointCalled {
		t.Errorf("midpoint shouldn't be called yet")
	}
	
	trans.Update(0.3) // reaches 0.4 (fadeOutDur)
	if !midpointCalled {
		t.Errorf("midpoint should have been called")
	}
	if trans.state != transStateHold {
		t.Errorf("expected hold state, got %d", trans.state)
	}
	
	// Hold
	trans.Update(0.06) // reaches 0.05 (holdDur)
	if trans.state != transStateFadeIn {
		t.Errorf("expected fade in state, got %d", trans.state)
	}
	
	// FadeIn
	trans.Update(0.4) // reaches 0.4 (fadeInDur)
	if trans.state != transStateIdle {
		t.Errorf("expected idle state, got %d", trans.state)
	}
	if trans.Active() {
		t.Errorf("expected transition to be inactive after completion")
	}
	
	// StartFadeIn
	trans.StartFadeIn()
	if trans.state != transStateFadeIn {
		t.Errorf("expected fade in state, got %d", trans.state)
	}
	
	trans.Resize(200, 200)
	if trans.overlay.Bounds().Dx() != 200 {
		t.Errorf("expected overlay to be resized")
	}
	
	// Test update when idle
	trans.state = transStateIdle
	trans.Update(1.0)
	if trans.state != transStateIdle {
		t.Errorf("expected to stay idle")
	}
}
