package game

import (
	"encoding/json"
	"os"
)

// NPCMetaState tracks persistent cross-run state for a single NPC.
type NPCMetaState struct {
	Met          bool `json:"met"`
	DefeatCount  int  `json:"defeat_count"`
	HighestPhase int  `json:"highest_phase"`
	TotalTrust   int  `json:"total_trust"`
}

// MetaSave holds persistent cross-run progression data.
type MetaSave struct {
	Remnants   int                       `json:"remnants"`
	RunCount   int                       `json:"run_count"`
	BestFloor  int                       `json:"best_floor"`
	TotalKills int                       `json:"total_kills"`
	NPCMeta    map[string]*NPCMetaState  `json:"npc_meta,omitempty"`
}

const metaSavePath = "meta.json"

// LoadMeta reads the meta save file, returning defaults if it doesn't exist.
func LoadMeta() *MetaSave {
	data, err := os.ReadFile(metaSavePath)
	if err != nil {
		return &MetaSave{NPCMeta: make(map[string]*NPCMetaState)}
	}
	var m MetaSave
	if err := json.Unmarshal(data, &m); err != nil {
		return &MetaSave{NPCMeta: make(map[string]*NPCMetaState)}
	}
	if m.NPCMeta == nil {
		m.NPCMeta = make(map[string]*NPCMetaState)
	}
	return &m
}

// LoadMetaSaveWithError reads the meta save file. Returns nil, os.ErrNotExist if the
// file is absent. Returns nil, err if the file is present but unreadable or corrupt.
func LoadMetaSaveWithError() (*MetaSave, error) {
	data, err := os.ReadFile(metaSavePath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var m MetaSave
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.NPCMeta == nil {
		m.NPCMeta = make(map[string]*NPCMetaState)
	}
	return &m, nil
}

// SaveMeta writes the meta save to disk.
func SaveMeta(m *MetaSave) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(metaSavePath, data, 0644)
}
