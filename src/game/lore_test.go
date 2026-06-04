package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLoreRegistry_LoadsAllCategories(t *testing.T) {
	content := `[
		{"id":"t1","title":"T1","category":"character","body":"body1"},
		{"id":"t2","title":"T2","category":"cosmology","body":"body2"},
		{"id":"t3","title":"T3","category":"history","body":"body3"},
		{"id":"t4","title":"T4","category":"fragment","body":"body4"}
	]`
	tmp := filepath.Join(t.TempDir(), "lore.json")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	reg, err := LoadLoreRegistry(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(reg))
	}
}

func TestIsLoreUnlocked(t *testing.T) {
	meta := &MetaSave{LoreUnlocked: []string{"entry_a", "entry_b"}}
	if !IsLoreUnlocked(meta, "entry_a") {
		t.Error("entry_a should be unlocked")
	}
	if IsLoreUnlocked(meta, "entry_c") {
		t.Error("entry_c should not be unlocked")
	}
}

func TestUnlockLore(t *testing.T) {
	meta := &MetaSave{}
	
	if !UnlockLore(meta, "lore1") {
		t.Errorf("expected true when unlocking new lore")
	}
	if !IsLoreUnlocked(meta, "lore1") {
		t.Errorf("expected lore1 to be unlocked")
	}
	
	if UnlockLore(meta, "lore1") {
		t.Errorf("expected false when unlocking already unlocked lore")
	}
	
	// test LoadLoreRegistry errors
	_, err := LoadLoreRegistry("non_existent_file.json")
	if err == nil {
		t.Errorf("expected error for missing file")
	}
	
	tmp := t.TempDir() + "/bad.json"
	os.WriteFile(tmp, []byte("bad json"), 0644)
	_, err = LoadLoreRegistry(tmp)
	if err == nil {
		t.Errorf("expected error for bad json")
	}
}
