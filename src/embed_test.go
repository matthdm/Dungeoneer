package main

import (
	"io/fs"
	"testing"
)

// Every path the game reads through gamedata must be in the embedded copy,
// or a lone executable / browser build silently loses that content.
func TestDataFS_ContainsGameData(t *testing.T) {
	files := []string{
		"data/lore.json",
		"data/biome_flavor.json",
		"data/enemy_flavor.json",
		"data/items_flavor.json",
		"data/events.json",
		"levels/hub.json",
		"dev_settings.json",
		"dialogues/varn_phase0.json",
	}
	for _, name := range files {
		if _, err := fs.Stat(dataFS, name); err != nil {
			t.Errorf("embedded data missing %s: %v", name, err)
		}
	}

	entries, err := fs.ReadDir(dataFS, "dialogues")
	if err != nil {
		t.Fatalf("read embedded dialogues: %v", err)
	}
	if len(entries) < 18 {
		t.Errorf("embedded dialogues has %d files, want at least 18", len(entries))
	}
}
