package game

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	echoDir          = "echoes"
	echoSnapshotRate = 2.0 // seconds between position snapshots
	echoMaxRuns      = 10  // maximum stored echo run files
)

// PositionSnapshot captures the player's state at a point in time.
type PositionSnapshot struct {
	Floor     int     `json:"floor"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	HP        int     `json:"hp"`
	Timestamp float64 `json:"t"`
}

// EchoRecord captures a full run's replay data for echo spawning.
type EchoRecord struct {
	RunIndex      int                `json:"run_index"`
	Snapshots     []PositionSnapshot `json:"snapshots"`
	EquippedItems []string           `json:"equipped_items"`
	DeathCause    string             `json:"death_cause"`
	DeathFloor    int                `json:"death_floor"`
	DeathX        float64            `json:"death_x"`
	DeathY        float64            `json:"death_y"`
}

// EchoRecorder ticks each game update and builds an EchoRecord for the current run.
type EchoRecorder struct {
	record      EchoRecord
	snapshotAcc float64 // seconds since last snapshot
	active      bool
}

// Start begins recording for a new run.
func (r *EchoRecorder) Start(runIndex int) {
	r.record = EchoRecord{RunIndex: runIndex}
	r.snapshotAcc = 0
	r.active = true
}

// Tick is called every game update. dt is delta time in seconds.
// x, y are the player's interpolated tile coordinates; floor is the current floor number.
func (r *EchoRecorder) Tick(dt float64, x, y float64, hp, floor int) {
	if !r.active {
		return
	}
	r.snapshotAcc += dt
	if r.snapshotAcc >= echoSnapshotRate {
		r.snapshotAcc = 0
		r.record.Snapshots = append(r.record.Snapshots, PositionSnapshot{
			Floor:     floor,
			X:         x,
			Y:         y,
			HP:        hp,
			Timestamp: float64(len(r.record.Snapshots)) * echoSnapshotRate,
		})
		// Cap snapshots to prevent unbounded growth (max 50 per floor × 7 floors).
		if len(r.record.Snapshots) > 350 {
			r.record.Snapshots = r.record.Snapshots[len(r.record.Snapshots)-350:]
		}
	}
}

// Finalize stops recording and writes the echo file. cause is the death/victory string.
func (r *EchoRecorder) Finalize(cause string, floor int, x, y float64, equippedIDs []string, meta *MetaSave) {
	if !r.active {
		return
	}
	r.active = false
	r.record.DeathCause = cause
	r.record.DeathFloor = floor
	r.record.DeathX = x
	r.record.DeathY = y
	r.record.EquippedItems = equippedIDs

	if err := os.MkdirAll(echoDir, 0755); err != nil {
		fmt.Printf("echo: could not create echoes dir: %v\n", err)
		return
	}
	path := fmt.Sprintf("%s/run_%d.json", echoDir, r.record.RunIndex)
	data, err := marshalEchoRecord(&r.record)
	if err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("echo: write failed: %v\n", err)
		return
	}

	// Register and evict oldest if over cap.
	meta.EchoFiles = append(meta.EchoFiles, path)
	if len(meta.EchoFiles) > echoMaxRuns {
		oldest := meta.EchoFiles[0]
		meta.EchoFiles = meta.EchoFiles[1:]
		_ = os.Remove(oldest)
	}
}

// marshalEchoRecord encodes an EchoRecord to indented JSON.
func marshalEchoRecord(rec *EchoRecord) ([]byte, error) {
	return json.MarshalIndent(rec, "", "  ")
}

// LoadEchoRecord reads a single echo record from disk.
func LoadEchoRecord(path string) (*EchoRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec EchoRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// SnapshotsForFloor filters snapshots to those on the given floor.
func (rec *EchoRecord) SnapshotsForFloor(floor int) []PositionSnapshot {
	var out []PositionSnapshot
	for _, s := range rec.Snapshots {
		if s.Floor == floor {
			out = append(out, s)
		}
	}
	return out
}
