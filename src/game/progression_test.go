package game

import "testing"

func TestRollGoldDrop(t *testing.T) {
	tests := []struct {
		role  string
		floor int
		exp   int
	}{
		{"swarm", 1, (3 + 2) / 2},
		{"elite", 2, (3 + 4) * 3},
		{"boss", 3, (3 + 6) * 8},
		{"melee", 4, 3 + 8},
	}

	for _, tt := range tests {
		got := rollGoldDrop(tt.role, tt.floor)
		if got != tt.exp {
			t.Errorf("rollGoldDrop(%s, %d) = %d; want %d", tt.role, tt.floor, got, tt.exp)
		}
	}
}
