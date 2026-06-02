package game

import (
	"encoding/json"
	"os"
)

// LoreCategory classifies a lore entry for the library UI tabs.
type LoreCategory string

const (
	LoreCategoryCharacter LoreCategory = "character"
	LoreCategoryCosmology LoreCategory = "cosmology"
	LoreCategoryHistory   LoreCategory = "history"
	LoreCategoryFragment  LoreCategory = "fragment"
)

// LoreDef is a single lore entry.
type LoreDef struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Category LoreCategory `json:"category"`
	Body     string       `json:"body"`
}

// LoreRegistry is the loaded set of all lore definitions.
var LoreRegistry []LoreDef

// LoadLoreRegistry reads lore entries from path and caches them in LoreRegistry.
func LoadLoreRegistry(path string) ([]LoreDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var defs []LoreDef
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, err
	}
	LoreRegistry = defs
	return defs, nil
}

// IsLoreUnlocked returns true if the given lore ID is in MetaSave.LoreUnlocked.
func IsLoreUnlocked(meta *MetaSave, id string) bool {
	for _, uid := range meta.LoreUnlocked {
		if uid == id {
			return true
		}
	}
	return false
}

// UnlockLore adds a lore ID to MetaSave.LoreUnlocked (idempotent). Returns true if newly added.
func UnlockLore(meta *MetaSave, id string) bool {
	if IsLoreUnlocked(meta, id) {
		return false
	}
	meta.LoreUnlocked = append(meta.LoreUnlocked, id)
	return true
}
