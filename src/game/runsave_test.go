package game

import (
	"os"
	"testing"
)

func TestRunSave(t *testing.T) {
	// Clean up any existing runsave.json
	os.Remove("runsave.json")
	defer os.Remove("runsave.json")

	// Test LoadRunSave with no file
	rs, err := LoadRunSave()
	if err != nil {
		t.Fatalf("expected no error when file is missing: %v", err)
	}
	if rs != nil {
		t.Fatalf("expected nil rs when file is missing")
	}

	// Test SaveRunSave
	save := &RunSave{
		FloorSeed: 12345,
	}
	
	err = SaveRunSave(save)
	if err != nil {
		t.Fatalf("failed to save run: %v", err)
	}

	// Test LoadRunSave
	loaded, err := LoadRunSave()
	if err != nil {
		t.Fatalf("failed to load run: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected loaded run save to not be nil")
	}
	if loaded.FloorSeed != 12345 {
		t.Errorf("expected floor seed 12345, got %d", loaded.FloorSeed)
	}
	if loaded.RunState.QuestFlags == nil {
		t.Errorf("expected QuestFlags to be initialized")
	}

	// Test corrupted file
	os.WriteFile("runsave.json", []byte("invalid json"), 0644)
	rs2, err2 := LoadRunSave()
	if err2 == nil {
		t.Fatalf("expected error on corrupted json")
	}
	if rs2 != nil {
		t.Fatalf("expected nil on corrupted json")
	}

	// Test ClearRunSave
	ClearRunSave()
	_, err3 := os.Stat("runsave.json")
	if !os.IsNotExist(err3) {
		t.Fatalf("expected runsave.json to be deleted")
	}
}
