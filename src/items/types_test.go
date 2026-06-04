package items

import (
	"testing"
)

func TestItemRegistryAndNewItem(t *testing.T) {
	// Register a test item template
	tmpl := &ItemTemplate{
		ID:        "test_sword",
		Name:      "Test Sword",
		Type:      ItemWeapon,
		Stackable: false,
		MaxStack:  1,
	}
	RegisterItem(tmpl)

	// Verify it is registered
	if Registry["test_sword"] != tmpl {
		t.Fatalf("expected template to be registered")
	}

	// Test NewItem
	it := NewItem("test_sword")
	if it.ItemTemplate.ID != "test_sword" || it.Count != 1 {
		t.Errorf("NewItem returned incorrect item: %+v", it)
	}

	// Test NewItem panic for invalid ID
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid item ID")
		}
	}()
	NewItem("invalid_id")
}

func TestItemSaveLoad(t *testing.T) {
	tmpl := &ItemTemplate{
		ID:        "test_potion",
		Name:      "Test Potion",
		Type:      ItemConsumable,
		Stackable: true,
		MaxStack:  99,
	}
	RegisterItem(tmpl)

	it := &Item{ItemTemplate: tmpl, Count: 5}

	// Test ToSave
	sv := it.ToSave()
	if sv.ID != "test_potion" || sv.Count != 5 {
		t.Errorf("ToSave returned incorrect save data: %+v", sv)
	}

	// Test FromSave
	loaded := FromSave(sv)
	if loaded.ID != "test_potion" || loaded.Count != 5 {
		t.Errorf("FromSave returned incorrect item: %+v", loaded)
	}
}

func TestEquipmentSerialization(t *testing.T) {
	tmpl1 := &ItemTemplate{ID: "test_helm", Name: "Test Helm", Type: ItemArmor}
	tmpl2 := &ItemTemplate{ID: "test_shield", Name: "Test Shield", Type: ItemArmor}
	RegisterItem(tmpl1)
	RegisterItem(tmpl2)

	eq := map[string]*Item{
		"Head":    {ItemTemplate: tmpl1, Count: 1},
		"Offhand": {ItemTemplate: tmpl2, Count: 1},
		"Ring1":   nil,
	}

	// Test SerializeEquipment
	savedEq := SerializeEquipment(eq)
	if len(savedEq) != 2 {
		t.Errorf("expected 2 saved items, got %d", len(savedEq))
	}
	if savedEq["Head"].ID != "test_helm" || savedEq["Offhand"].ID != "test_shield" {
		t.Errorf("SerializeEquipment returned incorrect values: %+v", savedEq)
	}

	// Test DeserializeEquipment
	loadedEq := DeserializeEquipment(savedEq)
	if len(loadedEq) != 2 {
		t.Errorf("expected 2 loaded items, got %d", len(loadedEq))
	}
	if loadedEq["Head"] == nil || loadedEq["Head"].ID != "test_helm" {
		t.Errorf("failed to deserialize Head item correctly")
	}
	if loadedEq["Offhand"] == nil || loadedEq["Offhand"].ID != "test_shield" {
		t.Errorf("failed to deserialize Offhand item correctly")
	}
}

func TestItemEffectCooldown(t *testing.T) {
	eff := &ItemEffect{
		CooldownSec: 2.5,
	}

	if !eff.IsReady() {
		t.Error("expected effect to be ready initially")
	}

	eff.PutOnCooldown()
	if eff.IsReady() {
		t.Error("expected effect to be on cooldown after PutOnCooldown")
	}

	// Tick part way
	eff.TickCooldown(1.0)
	if eff.IsReady() {
		t.Error("expected effect to still be on cooldown after ticking 1s")
	}

	// Tick to completion
	eff.TickCooldown(1.5)
	if !eff.IsReady() {
		t.Error("expected effect to be ready after cooldown expires")
	}

	// Tick past 0
	eff.TickCooldown(0.5)
	if !eff.IsReady() {
		t.Error("expected effect to stay ready")
	}
}
