package progression

import "testing"

func TestEXPToLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected int
	}{
		{
			name:     "Level 0",
			level:    0,
			expected: 0,
		},
		{
			name:     "Level 1",
			level:    1,
			expected: 150,
		},
		{
			name:     "Level 2",
			level:    2,
			expected: 400,
		},
		{
			name:     "Level 10",
			level:    10,
			expected: 6000,
		},
		{
			name:     "Negative Level",
			level:    -1,
			expected: -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EXPToLevel(tt.level); got != tt.expected {
				t.Errorf("EXPToLevel(%d) = %d; want %d", tt.level, got, tt.expected)
			}
		})
	}
}

func TestCalculateEXPReward(t *testing.T) {
	tests := []struct {
		name        string
		enemyLevel  int
		playerLevel int
		expected    int
	}{
		{
			name:        "Same level",
			enemyLevel:  1,
			playerLevel: 1,
			expected:    60,
		},
		{
			name:        "Higher enemy level",
			enemyLevel:  5,
			playerLevel: 1,
			expected:    180,
		},
		{
			name:        "Lower enemy level (multiplier below 0.1)",
			enemyLevel:  1,
			playerLevel: 10,
			expected:    6,
		},
		{
			name:        "Lower enemy level (multiplier exactly 0.0, clamp to 0.1)",
			enemyLevel:  0,
			playerLevel: 5,
			expected:    5,
		},
		{
			name:        "Negative enemy level",
			enemyLevel:  -1,
			playerLevel: 0,
			expected:    32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateEXPReward(tt.enemyLevel, tt.playerLevel); got != tt.expected {
				t.Errorf("CalculateEXPReward(%d, %d) = %d; want %d", tt.enemyLevel, tt.playerLevel, got, tt.expected)
			}
		})
	}
}
