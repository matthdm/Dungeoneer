package game

import (
	"encoding/json"
	"os"
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

func TestLoadSaveMeta(t *testing.T) {
	os.Remove("meta.json")
	defer os.Remove("meta.json")
	
	// Test LoadMeta when file missing
	m := LoadMeta()
	if m == nil {
		t.Fatalf("expected non-nil meta")
	}
	if m.NPCMeta == nil || m.HubState == nil || m.Upgrades == nil {
		t.Errorf("expected maps to be initialized")
	}
	
	// Test LoadMetaSaveWithError when file missing
	_, err := LoadMetaSaveWithError()
	if err == nil {
		t.Errorf("expected error when file missing")
	}
	
	// Test SaveMeta
	m.Remnants = 100
	SaveMeta(m)
	
	// Test LoadMeta when file exists
	m2 := LoadMeta()
	if m2.Remnants != 100 {
		t.Errorf("expected 100 remnants, got %d", m2.Remnants)
	}
	
	// Test LoadMetaSaveWithError when file exists
	m3, err := LoadMetaSaveWithError()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if m3.Remnants != 100 {
		t.Errorf("expected 100 remnants")
	}
	
	// Test LoadMeta with corrupted file
	os.WriteFile("meta.json", []byte("invalid"), 0644)
	m4 := LoadMeta()
	if m4.Remnants != 0 {
		t.Errorf("expected 0 remnants after corrupt load")
	}
	
	// Test LoadMetaSaveWithError with corrupted file
	_, err2 := LoadMetaSaveWithError()
	if err2 == nil {
		t.Errorf("expected error for corrupted file")
	}
}
