package game

import (
	"testing"
)

func TestAbilitySpellName(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"fireball", "fireball"},
		{"chaos_ray", "chaosray"},
		{"lightning", "lightning"},
		{"lightning_storm", "lightningstorm"},
		{"fractal_bloom", "fractalbloom"},
		{"fractal_canopy", "fractalcanopy"},
		{"arcane_spray", "arcane_spray"},
		{"arcane_bolt", "arcane_bolt"},
		{"unknown_spell", "unknown_spell"},
	}
	for _, tt := range tests {
		if got := abilitySpellName(tt.in); got != tt.out {
			t.Errorf("abilitySpellName(%q) = %q; want %q", tt.in, got, tt.out)
		}
	}
}

func TestPointSegmentDistance(t *testing.T) {
	// Point (0, 0), line from (1, 0) to (1, 10)
	// Expected distance: 1
	d := pointSegmentDistance(0, 0, 1, 0, 1, 10)
	if d != 1.0 {
		t.Errorf("expected 1.0, got %v", d)
	}

	// Point (0, -5), line from (1, 0) to (1, 10)
	// Expected distance to closest point (1,0) is hypot(1, 5)
	d2 := pointSegmentDistance(0, -5, 1, 0, 1, 10)
	if d2 < 5.0 || d2 > 5.1 {
		t.Errorf("expected ~5.099, got %v", d2)
	}

	// Line is a point
	d3 := pointSegmentDistance(3, 4, 0, 0, 0, 0)
	if d3 != 5.0 {
		t.Errorf("expected 5.0, got %v", d3)
	}
}
