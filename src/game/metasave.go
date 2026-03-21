package game

import (
	"encoding/json"
	"os"
)

// MetaSave holds persistent cross-run progression data.
type MetaSave struct {
	Remnants   int `json:"remnants"`
	RunCount   int `json:"run_count"`
	BestFloor  int `json:"best_floor"`
	TotalKills int `json:"total_kills"`
}

const metaSavePath = "meta.json"

// LoadMeta reads the meta save file, returning defaults if it doesn't exist.
func LoadMeta() *MetaSave {
	data, err := os.ReadFile(metaSavePath)
	if err != nil {
		return &MetaSave{}
	}
	var m MetaSave
	if err := json.Unmarshal(data, &m); err != nil {
		return &MetaSave{}
	}
	return &m
}

// SaveMeta writes the meta save to disk.
func SaveMeta(m *MetaSave) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(metaSavePath, data, 0644)
}
