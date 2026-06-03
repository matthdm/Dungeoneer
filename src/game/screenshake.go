package game

import "math/rand/v2"

// ScreenShake applies a decaying random camera offset for impact feedback.
type ScreenShake struct {
	IntensityX float64
	IntensityY float64
	Decay      float64 // intensity lost per second (e.g. 8.0 = gone in ~0.5s at intensity 4)
}

// Trigger starts a new shake. intensity is max pixel offset.
func (s *ScreenShake) Trigger(intensity float64) {
	s.IntensityX = (rand.Float64()*2 - 1) * intensity
	s.IntensityY = (rand.Float64()*2 - 1) * intensity
	if s.Decay == 0 {
		s.Decay = 12.0 // default: fade in ~0.3s from intensity 4
	}
}

// Update decays the shake intensity over time.
func (s *ScreenShake) Update(dt float64) {
	if s.IntensityX == 0 && s.IntensityY == 0 {
		return
	}
	decay := s.Decay * dt
	s.IntensityX = approach(s.IntensityX, 0, decay)
	s.IntensityY = approach(s.IntensityY, 0, decay)
}

// Offset returns the current shake offset to apply to camera translation.
func (s *ScreenShake) Offset() (float64, float64) {
	return s.IntensityX, s.IntensityY
}

// approach moves value toward target by at most delta (handles sign correctly).
func approach(value, target, delta float64) float64 {
	if value > target {
		value -= delta
		if value < target {
			value = target
		}
	} else if value < target {
		value += delta
		if value > target {
			value = target
		}
	}
	return value
}
