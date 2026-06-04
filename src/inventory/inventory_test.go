package inventory

import (
	"dungeoneer/items"
	"testing"
)

func TestNew(t *testing.T) {
	inv := New(3, 4)
	if inv.Width != 3 || inv.Height != 4 {
		t.Errorf("expected size 3x4, got %dx%d", inv.Width, inv.Height)
	}
	if len(inv.Grid) != 4 {
		t.Errorf("expected 4 rows, got %d", len(inv.Grid))
	}
	if len(inv.Grid[0]) != 3 {
		t.Errorf("expected 3 columns, got %d", len(inv.Grid[0]))
	}
}

func TestAddItemAndFirstEmpty(t *testing.T) {
	inv := New(2, 2)
	
	x, y, ok := FirstEmpty(inv)
	if !ok || x != 0 || y != 0 {
		t.Errorf("expected (0,0) empty, got %v, %v, %v", x, y, ok)
	}

	it1 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword1"}, Count: 1}
	added := inv.AddItem(it1)
	if !added {
		t.Errorf("expected item to be added")
	}
	if inv.Grid[0][0] != it1 {
		t.Errorf("expected item at (0,0)")
	}

	x, y, ok = FirstEmpty(inv)
	if !ok || x != 1 || y != 0 {
		t.Errorf("expected (1,0) empty, got %v, %v, %v", x, y, ok)
	}

	// Fill the rest
	inv.AddItem(&items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword2"}, Count: 1})
	inv.AddItem(&items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword3"}, Count: 1})
	inv.AddItem(&items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword4"}, Count: 1})

	_, _, ok = FirstEmpty(inv)
	if ok {
		t.Errorf("expected no empty slots")
	}

	// Try adding to full inventory
	added = inv.AddItem(&items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword5"}, Count: 1})
	if added {
		t.Errorf("expected false when adding to full inventory")
	}
}

func TestStacking(t *testing.T) {
	inv := New(2, 2)
	
	// Add stackable item
	it1 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "potion", Stackable: true, MaxStack: 5}, Count: 2}
	inv.AddItem(it1)

	// Try stacking
	it2 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "potion", Stackable: true, MaxStack: 5}, Count: 2}
	added := inv.AddItem(it2)
	if !added {
		t.Errorf("expected stacking to succeed")
	}
	if inv.Grid[0][0].Count != 4 {
		t.Errorf("expected count 4, got %v", inv.Grid[0][0].Count)
	}

	// Try stacking over capacity
	it3 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "potion", Stackable: true, MaxStack: 5}, Count: 2}
	// TryStack only returns true if ENTIRE count is merged.
	// Since 4 + 2 > 5, it should fail to stack and place it in a new slot.
	added = inv.AddItem(it3)
	if !added {
		t.Errorf("expected item to be added to next slot")
	}
	if inv.Grid[0][0].Count != 4 {
		t.Errorf("expected first stack to remain 4")
	}
	if inv.Grid[0][1] != it3 {
		t.Errorf("expected second stack to be created")
	}

	// Not stackable
	it4 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword", Stackable: false}, Count: 1}
	inv.AddItem(it4)
	it5 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "sword", Stackable: false}, Count: 1}
	inv.AddItem(it5)
	
	if inv.Grid[1][0] != it4 || inv.Grid[1][1] != it5 {
		t.Errorf("expected non-stackable items to occupy separate slots")
	}
}

func TestHasItem(t *testing.T) {
	inv := New(2, 2)
	inv.AddItem(&items.Item{ItemTemplate: &items.ItemTemplate{ID: "potion"}, Count: 1})

	if !inv.HasItem("potion") {
		t.Errorf("expected true for existing item")
	}
	if inv.HasItem("sword") {
		t.Errorf("expected false for missing item")
	}
}

func TestSaveLoad(t *testing.T) {
	items.RegisterItem(&items.ItemTemplate{ID: "potion"})
	inv := New(2, 2)
	it1 := &items.Item{ItemTemplate: &items.ItemTemplate{ID: "potion"}, Count: 3}
	inv.Grid[1][1] = it1

	saveData := inv.ToSaveData()
	if len(saveData) != 2 || len(saveData[0]) != 2 {
		t.Fatalf("expected 2x2 save data")
	}
	if saveData[1][1].ID != "potion" {
		t.Errorf("expected saved item to be potion, got %v", saveData[1][1].ID)
	}
	if saveData[0][0].ID != "" {
		t.Errorf("expected empty slot to be saved as empty string")
	}

	loadedInv := FromSaveData(saveData)
	if loadedInv.Width != 2 || loadedInv.Height != 2 {
		t.Errorf("expected 2x2 loaded inventory")
	}
	if loadedInv.Grid[1][1] == nil || loadedInv.Grid[1][1].ID != "potion" {
		t.Errorf("expected loaded item to be potion")
	}
	if loadedInv.Grid[1][1].Count != 3 {
		t.Errorf("expected loaded item count to be 3")
	}
	if loadedInv.Grid[0][0] != nil {
		t.Errorf("expected empty slot to remain nil")
	}
}

func TestFromSaveDataDefaults(t *testing.T) {
	// Test loading empty array
	loaded := FromSaveData([][]items.ItemSave{})
	if loaded.Width != Width || loaded.Height != Height {
		t.Errorf("expected default size for empty save data")
	}
}
