package hud

import "testing"

func TestMinimap_StartsHidden(t *testing.T) {
	m := &Minimap{}
	if m.Visible {
		t.Error("minimap should start hidden")
	}
}

func TestMinimap_NoRoomsDrawsNothing(t *testing.T) {
	m := &Minimap{Visible: true}
	// Draw with empty rooms — should not panic.
	m.Draw(nil, nil, nil, 0, 0, -1, -1, 640, 480)
}
