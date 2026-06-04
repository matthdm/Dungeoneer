package controls

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("expected New() to not return nil")
	}
	if len(c.bindings) != len(defaultBindings) {
		t.Errorf("expected %d bindings, got %d", len(defaultBindings), len(c.bindings))
	}
	for action, expectedBinding := range defaultBindings {
		gotBinding, ok := c.bindings[action]
		if !ok {
			t.Errorf("missing default binding for action %s", action)
			continue
		}
		if gotBinding.Primary != expectedBinding.Primary || gotBinding.Secondary != expectedBinding.Secondary {
			t.Errorf("for action %s, expected binding %+v, got %+v", action, expectedBinding, gotBinding)
		}
	}
}

func TestGetBinding(t *testing.T) {
	c := New()
	// Test existing action
	binding := c.GetBinding(ActionMoveLeft)
	expected := defaultBindings[ActionMoveLeft]
	if binding.Primary != expected.Primary || binding.Secondary != expected.Secondary {
		t.Errorf("expected %+v, got %+v", expected, binding)
	}

	// Test non-existent action
	badAction := ActionID("non_existent_action")
	binding = c.GetBinding(badAction)
	if binding.Primary != ebiten.Key(0) || binding.Secondary != ebiten.Key(0) {
		t.Errorf("expected empty binding for non-existent action, got %+v", binding)
	}
}

func TestSetBinding(t *testing.T) {
	c := New()
	action := ActionMoveLeft

	// Set primary key
	c.SetBinding(action, ebiten.KeySpace)
	binding := c.GetBinding(action)
	if binding.Primary != ebiten.KeySpace {
		t.Errorf("expected primary key to be KeySpace, got %s", GetKeyName(binding.Primary))
	}

	// Try to set non-existent action (should do nothing/no panic)
	badAction := ActionID("invalid_action")
	c.SetBinding(badAction, ebiten.KeySpace)
	binding = c.GetBinding(badAction)
	if binding.Primary != ebiten.Key(0) {
		t.Errorf("expected empty binding for non-existent action after set, got %+v", binding)
	}
}

func TestSetSecondaryBinding(t *testing.T) {
	c := New()
	action := ActionMoveLeft

	// Set secondary key
	c.SetSecondaryBinding(action, ebiten.KeyEnter)
	binding := c.GetBinding(action)
	if binding.Secondary != ebiten.KeyEnter {
		t.Errorf("expected secondary key to be KeyEnter, got %s", GetKeyName(binding.Secondary))
	}

	// Try to set non-existent action (should do nothing/no panic)
	badAction := ActionID("invalid_action")
	c.SetSecondaryBinding(badAction, ebiten.KeyEnter)
	binding = c.GetBinding(badAction)
	if binding.Secondary != ebiten.Key(0) {
		t.Errorf("expected empty binding for non-existent action after set secondary, got %+v", binding)
	}
}

func TestResetAndResetBinding(t *testing.T) {
	c := New()
	action := ActionMoveLeft

	// Change binding and reset single binding
	originalBinding := defaultBindings[action]
	c.SetBinding(action, ebiten.KeySpace)
	c.SetSecondaryBinding(action, ebiten.KeyEnter)

	c.ResetBinding(action)
	binding := c.GetBinding(action)
	if binding.Primary != originalBinding.Primary || binding.Secondary != originalBinding.Secondary {
		t.Errorf("expected reset binding to match default %+v, got %+v", originalBinding, binding)
	}

	// Reset invalid binding should do nothing
	c.ResetBinding(ActionID("invalid_action"))

	// Change multiple bindings and reset all
	c.SetBinding(ActionMoveLeft, ebiten.KeySpace)
	c.SetBinding(ActionMoveRight, ebiten.KeySpace)
	c.Reset()
	for act, expected := range defaultBindings {
		got := c.GetBinding(act)
		if got.Primary != expected.Primary || got.Secondary != expected.Secondary {
			t.Errorf("after reset, expected %+v for %s, got %+v", expected, act, got)
		}
	}
}

func TestGetAllActionIDs(t *testing.T) {
	ids := GetAllActionIDs()
	if len(ids) == 0 {
		t.Error("expected non-empty slice of ActionIDs")
	}

	// Ensure there are no duplicate IDs
	seen := make(map[ActionID]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate ActionID found: %s", id)
		}
		seen[id] = true
	}
}

func TestGetActionLabel(t *testing.T) {
	tests := []struct {
		action   ActionID
		expected string
	}{
		{ActionMoveLeft, "Move Left"},
		{ActionMoveRight, "Move Right"},
		{ActionMoveUp, "Move Up"},
		{ActionMoveDown, "Move Down"},
		{ActionDash, "Dash"},
		{ActionGrapple, "Grapple"},
		{ActionMenuUp, "Menu Up"},
		{ActionMenuDown, "Menu Down"},
		{ActionMenuConfirm, "Menu Confirm"},
		{ActionMenuCancel, "Menu Cancel"},
		{ActionSpell1, "Spell 1 - Fireball"},
		{ActionSpell2, "Spell 2 - Chaos Ray"},
		{ActionSpell3, "Spell 3 - Lightning Strike"},
		{ActionSpell4, "Spell 4 - Lightning Storm"},
		{ActionSpell5, "Spell 5 - Fractal Bloom"},
		{ActionSpell6, "Spell 6 - Fractal Canopy"},
		{ActionInteract, "Interact"},
		{ActionInventory, "Open Inventory"},
		{ActionHeroPanel, "Toggle Hero Panel"},
		{ActionTogglePause, "Pause Game"},
		{ActionShowHUD, "Toggle HUD"},
		{ActionToggleKeybind, "Toggle Controls Info"},
		{ActionID("custom_action"), "custom_action"}, // fallback
	}

	for _, tc := range tests {
		t.Run(string(tc.action), func(t *testing.T) {
			got := GetActionLabel(tc.action)
			if got != tc.expected {
				t.Errorf("expected label %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestGetKeyName(t *testing.T) {
	tests := []struct {
		key      ebiten.Key
		expected string
	}{
		{ebiten.KeyArrowLeft, "←"},
		{ebiten.KeyArrowRight, "→"},
		{ebiten.KeyArrowUp, "↑"},
		{ebiten.KeyArrowDown, "↓"},
		{ebiten.KeyEscape, "Esc"},
		{ebiten.KeyControl, "Ctrl"},
		{ebiten.KeyA, "A"},
		{ebiten.KeySpace, "Space"},
		{ebiten.Key(9999), "?"}, // invalid/unrecognized key
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			got := GetKeyName(tc.key)
			if got != tc.expected {
				t.Errorf("expected key name %q for key %d, got %q", tc.expected, tc.key, got)
			}
		})
	}
}
