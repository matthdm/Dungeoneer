package game

import (
	"os"
	"testing"
	"testing/fstest"

	"dungeoneer/gamedata"
	"dungeoneer/items"
)

// These tests run with the working directory at src/game, where none of the
// data paths exist on disk — the same situation as a lone executable or a
// browser build. The repo's real data files stand in for the embedded copy.

func useRepoData(t *testing.T) {
	t.Helper()
	gamedata.Register(os.DirFS(".."))
	t.Cleanup(func() { gamedata.Register(nil) })
}

func TestLoadLoreRegistry_EmbeddedFallback(t *testing.T) {
	useRepoData(t)
	defs, err := LoadLoreRegistry("data/lore.json")
	if err != nil {
		t.Fatalf("LoadLoreRegistry: %v", err)
	}
	if len(defs) == 0 {
		t.Error("no lore entries loaded")
	}
}

func TestLoadBiomeFlavor_EmbeddedFallback(t *testing.T) {
	useRepoData(t)
	BiomeFlavors = nil
	if err := LoadBiomeFlavor("data/biome_flavor.json"); err != nil {
		t.Fatalf("LoadBiomeFlavor: %v", err)
	}
	if len(BiomeFlavors) == 0 {
		t.Error("no biome flavor loaded")
	}
}

func TestLoadEnemyFlavor_EmbeddedFallback(t *testing.T) {
	useRepoData(t)
	EnemyFlavors = nil
	if err := LoadEnemyFlavor("data/enemy_flavor.json"); err != nil {
		t.Fatalf("LoadEnemyFlavor: %v", err)
	}
	if len(EnemyFlavors) == 0 {
		t.Error("no enemy flavor loaded")
	}
}

func TestLoadEventDefs_EmbeddedFallback(t *testing.T) {
	useRepoData(t)
	EventDefs = nil
	if err := LoadEventDefs("data/events.json"); err != nil {
		t.Fatalf("LoadEventDefs: %v", err)
	}
	if len(EventDefs) == 0 {
		t.Error("no event defs loaded")
	}
}

func TestLoadItemFlavor_EmbeddedFallback(t *testing.T) {
	useRepoData(t)
	if err := items.LoadItemFlavor("data/items_flavor.json"); err != nil {
		t.Fatalf("LoadItemFlavor: %v", err)
	}
}

func TestLoadDevSettings_EmbeddedFallback(t *testing.T) {
	// The default when the file is unreadable is legacy combat ON, so an
	// embedded file that turns it off proves the embedded copy was read.
	gamedata.Register(fstest.MapFS{
		"dev_settings.json": {Data: []byte(`{"use_legacy_combat": false}`)},
	})
	t.Cleanup(func() { gamedata.Register(nil) })

	if LoadDevSettings().UseLegacyCombat {
		t.Error("UseLegacyCombat = true, want false from embedded dev_settings.json")
	}
}
