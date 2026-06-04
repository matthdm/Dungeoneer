package game

import "testing"

func TestDepthWeightForSprite(t *testing.T) {
	tests := []struct {
		id       string
		expected float64
	}{
		{"crypt_wall", 1.0},
		{"moss_beam", 1.0},
		{"gallery_statue", 1.0},
		{"tree_green", 1.0},
		{"oak_log", 1.0},
		{"dirt_chunk", 1.0},
		{"floor", -1.0},
		{"water_shallow", -0.9},
		{"lava_pool", -0.9},
		{"stairs_up", -0.9},
		{"portal_blue", -0.9},
		{"trap_spike", -0.9},
		{"goblin", 0.5},
		{"chest", 0.5},
	}

	for _, tt := range tests {
		got := depthWeightForSprite(tt.id)
		if got != tt.expected {
			t.Errorf("depthWeightForSprite(%q) = %v; want %v", tt.id, got, tt.expected)
		}
	}
}
