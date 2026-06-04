package constants

import "testing"

func TestDebugMode(t *testing.T) {
	// Save the original value to restore after the test
	originalDebugMode := DebugMode
	defer func() {
		DebugMode = originalDebugMode
	}()

	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{
			name:     "Set false",
			input:    false,
			expected: false,
		},
		{
			name:     "Set true",
			input:    true,
			expected: true,
		},
	}

	// Verify default state first
	if IsDebugMode() != originalDebugMode {
		t.Errorf("expected initial debug mode to be %v, got %v", originalDebugMode, IsDebugMode())
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDebugMode(tt.input)
			if got := IsDebugMode(); got != tt.expected {
				t.Errorf("IsDebugMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}
