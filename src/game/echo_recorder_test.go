package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEchoRecord_SerializeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	rec := &EchoRecord{
		RunIndex:      3,
		Snapshots:     []PositionSnapshot{{Floor: 2, X: 10.5, Y: 8.0, HP: 45, Timestamp: 2.0}},
		EquippedItems: []string{"item_0_26"},
		DeathCause:    "melee",
		DeathFloor:    2,
		DeathX:        10.5,
		DeathY:        8.0,
	}
	path := filepath.Join(dir, "run_3.json")
	data, err := marshalEchoRecord(rec)
	if err != nil {
		t.Fatalf("marshalEchoRecord: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadEchoRecord(path)
	if err != nil {
		t.Fatalf("LoadEchoRecord: %v", err)
	}
	if loaded.RunIndex != 3 {
		t.Errorf("RunIndex want 3, got %d", loaded.RunIndex)
	}
	if loaded.DeathCause != "melee" {
		t.Errorf("DeathCause want melee, got %s", loaded.DeathCause)
	}
	if len(loaded.Snapshots) != 1 {
		t.Errorf("Snapshots want 1, got %d", len(loaded.Snapshots))
	}
}

func TestEchoRecord_SnapshotsForFloor(t *testing.T) {
	rec := &EchoRecord{
		Snapshots: []PositionSnapshot{
			{Floor: 1, X: 5, Y: 5},
			{Floor: 2, X: 8, Y: 8},
			{Floor: 2, X: 9, Y: 9},
			{Floor: 3, X: 2, Y: 2},
		},
	}
	snaps := rec.SnapshotsForFloor(2)
	if len(snaps) != 2 {
		t.Errorf("want 2 snapshots for floor 2, got %d", len(snaps))
	}
}

func TestEchoRecorder_Tick_SnapshotsAtInterval(t *testing.T) {
	r := &EchoRecorder{}
	r.Start(0)
	r.Tick(1.0, 5.0, 5.0, 100, 1) // 1s — no snapshot yet
	if len(r.record.Snapshots) != 0 {
		t.Error("should not snapshot before 2s interval")
	}
	r.Tick(1.1, 5.0, 5.0, 100, 1) // 2.1s total — should snapshot
	if len(r.record.Snapshots) != 1 {
		t.Errorf("should have 1 snapshot at 2s, got %d", len(r.record.Snapshots))
	}
}
