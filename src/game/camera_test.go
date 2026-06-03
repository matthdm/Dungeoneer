package game

import "testing"

func TestScreenShake_TriggerSetsIntensity(t *testing.T) {
	s := &ScreenShake{}
	s.Trigger(4.0)
	if s.IntensityX == 0 && s.IntensityY == 0 {
		t.Error("Trigger should set non-zero intensity")
	}
}

func TestScreenShake_DecaysToZero(t *testing.T) {
	s := &ScreenShake{IntensityX: 4.0, IntensityY: -3.0, Decay: 12.0}
	for i := 0; i < 120; i++ {
		s.Update(1.0 / 60.0)
	}
	if s.IntensityX != 0 || s.IntensityY != 0 {
		t.Errorf("shake should decay to zero, got X=%.3f Y=%.3f", s.IntensityX, s.IntensityY)
	}
}

func TestTransition_CompletesAfterDuration(t *testing.T) {
	tr := &Transition{
		overlay:    nil, // skip image creation in test
		fadeOutDur: 0.4,
		holdDur:    0.05,
		fadeInDur:  0.4,
	}
	midpointFired := false
	tr.Start(func() { midpointFired = true })

	if !tr.Active() {
		t.Error("transition should be active after Start")
	}
	// Advance past full cycle (0.4 + 0.05 + 0.4 = 0.85s).
	for i := 0; i < 60; i++ {
		tr.Update(1.0 / 60.0)
	}
	if !midpointFired {
		t.Error("OnMidpoint callback should have fired")
	}
	if tr.Active() {
		t.Error("transition should be idle after full cycle")
	}
}

func TestTransition_NotActiveBeforeStart(t *testing.T) {
	tr := &Transition{fadeOutDur: 0.4, holdDur: 0.05, fadeInDur: 0.4}
	if tr.Active() {
		t.Error("transition should not be active before Start")
	}
}
