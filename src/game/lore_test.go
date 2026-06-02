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
