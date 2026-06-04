package items

import (
	"testing"
)

func TestLoadDefaultItems(t *testing.T) {
	// Loading default items will parse JSON and construct the registry
	err := LoadDefaultItems()
	if err != nil {
		t.Fatalf("LoadDefaultItems failed: %v", err)
	}

	// Verify iron_key was registered
	if tmpl, ok := Registry["iron_key"]; !ok {
		t.Errorf("expected iron_key to be registered")
	} else if tmpl.Type != ItemConsumable {
		t.Errorf("expected iron_key to be consumable")
	}

	// Verify an ability override was applied
	if tmpl, ok := Registry["item_0_1"]; ok {
		if tmpl.GrantsAbility != "slash_combo" {
			t.Errorf("expected item_0_1 to grant slash_combo, got %v", tmpl.GrantsAbility)
		}
	} else {
		t.Errorf("expected item_0_1 to be in registry after loading")
	}
}
