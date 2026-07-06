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
	Betrayed     bool `json:"betrayed,omitempty"`
}

// MetaSave holds persistent cross-run progression data.
type MetaSave struct {
	// v0 fields
	Remnants   int                      `json:"remnants"`
	RunCount   int                      `json:"run_count"`
	BestFloor  int                      `json:"best_floor"`
	TotalKills int                      `json:"total_kills"`
	NPCMeta    map[string]*NPCMetaState `json:"npc_meta,omitempty"`

	// v1 fields — all zero-safe on old saves
	Version       int             `json:"version"`
	CompletedRuns int             `json:"completed_runs"`
	TotalDeaths   int             `json:"total_deaths"`
	TotalRemnants int             `json:"total_remnants"`
	LoreUnlocked  []string        `json:"lore_unlocked,omitempty"`
	HubState      map[string]bool `json:"hub_state,omitempty"`
	Upgrades       map[string]int  `json:"upgrades,omitempty"`
	ShopPurchases  []string        `json:"shop_purchases,omitempty"` // items purchased in hub shop this run; reset on run start
	EchoFiles      []string        `json:"echo_files,omitempty"`

	// v2 fields — Phase 7B Living Dungeon AI
	RecentBehavior []BehaviorRecord `json:"recent_behavior,omitempty"`
	CurrentProfile PlayerProfile    `json:"current_profile,omitempty"`
	CurrentMood    int              `json:"current_mood,omitempty"` // DungeonMood as int

	// v3 fields — Artifact library (combat redesign)
	ArtifactCollection []string `json:"artifact_collection,omitempty"` // all artifact IDs owned across runs
	ArtifactLoadout    [7]string `json:"artifact_loadout,omitempty"`   // equipped loadout: indices 0-5 regular, 6 = elite slot
}

const metaSavePath = "meta.json"

// migrateMetaSave initialises nil maps and advances the version to 1.
// Safe to call on both new and old saves.
func migrateMetaSave(m *MetaSave) {
	if m.NPCMeta == nil {
		m.NPCMeta = make(map[string]*NPCMetaState)
	}
	if m.HubState == nil {
		m.HubState = make(map[string]bool)
	}
	if m.Upgrades == nil {
		m.Upgrades = make(map[string]int)
	}
	if m.Version < 1 {
		m.Version = 1
	}
}

// LoadMeta reads the meta save file, returning defaults if it doesn't exist.
func LoadMeta() *MetaSave {
	data, err := os.ReadFile(metaSavePath)
	if err != nil {
		m := &MetaSave{}
		migrateMetaSave(m)
		return m
	}
	var m MetaSave
	if err := json.Unmarshal(data, &m); err != nil {
		m2 := &MetaSave{}
		migrateMetaSave(m2)
		return m2
	}
	migrateMetaSave(&m)
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
	migrateMetaSave(&m)
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
