package game

import (
	"encoding/json"
	"os"

	"dungeoneer/entities"
)

const runSavePath = "runsave.json"

// MonsterSnap captures the state of a living monster at save time.
type MonsterSnap struct {
	TileX  int    `json:"tile_x"`
	TileY  int    `json:"tile_y"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Level  int    `json:"level"`
	Role   string `json:"role"`
}

// RunSave holds everything needed to resume a mid-run session.
type RunSave struct {
	RunState  RunState                `json:"run_state"`
	FloorSeed int64                   `json:"floor_seed"`
	Player    entities.PlayerSnapshot `json:"player"`
	Monsters  []MonsterSnap           `json:"monsters"`
}

// SaveRunSave writes the current run to runsave.json.
func SaveRunSave(rs *RunSave) error {
	data, err := json.MarshalIndent(rs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(runSavePath, data, 0644)
}

// LoadRunSave reads runsave.json. Returns nil, nil if the file is absent.
// Returns nil, err if the file is present but unreadable or corrupt.
func LoadRunSave() (*RunSave, error) {
	data, err := os.ReadFile(runSavePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rs RunSave
	if err := json.Unmarshal(data, &rs); err != nil {
		return nil, err
	}
	if rs.RunState.QuestFlags == nil {
		rs.RunState.QuestFlags = make(map[string]int)
	}
	return &rs, nil
}

// ClearRunSave deletes runsave.json. Ignores "not found" errors.
func ClearRunSave() {
	_ = os.Remove(runSavePath)
}
