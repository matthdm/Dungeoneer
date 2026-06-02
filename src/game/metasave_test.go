package game

import (
	"encoding/json"
	"testing"
)

func TestLoadMeta_OldSaveGetsDefaultFields(t *testing.T) {
	// Simulate loading a v0 save (no new fields).
	old := `{"remnants":50,"run_count":2,"best_floor":3,"total_kills":10}`
	var m MetaSave
	if err := json.Unmarshal([]byte(old), &m); err != nil {
		t.Fatal(err)
	}
	migrateMetaSave(&m)

	if m.HubState == nil {
		t.Error("HubState should be initialized on old save")
	}
	if m.Upgrades == nil {
		t.Error("Upgrades should be initialized on old save")
	}
	if m.Version != 1 {
		t.Errorf("Version should be 1 after migration, got %d", m.Version)
	}
	if m.Remnants != 50 {
		t.Error("existing fields must survive migration")
	}
}
