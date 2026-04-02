package dialogue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Registry is the in-memory cache of loaded dialogue trees keyed by tree ID.
var Registry = map[string]*DialogueTree{}

// LoadTree reads a single DialogueTree from a JSON file and adds it to the Registry.
func LoadTree(path string) (*DialogueTree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dialogue: read %s: %w", path, err)
	}
	var tree DialogueTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("dialogue: parse %s: %w", path, err)
	}
	// Backfill node IDs from map keys if missing.
	for k, node := range tree.Nodes {
		if node.ID == "" {
			node.ID = k
		}
	}
	Registry[tree.ID] = &tree
	return &tree, nil
}

// LoadSimple reads a SimpleDialogue JSON file and converts it to a DialogueTree.
func LoadSimple(path string) (*DialogueTree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dialogue: read %s: %w", path, err)
	}
	var sd SimpleDialogue
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("dialogue: parse %s: %w", path, err)
	}
	// Derive tree ID from filename (e.g. "hollow_monk.json" -> "hollow_monk")
	base := filepath.Base(path)
	id := strings.TrimSuffix(base, filepath.Ext(base))
	tree := sd.ToTree(id)
	Registry[tree.ID] = tree
	return tree, nil
}

// LoadAll loads every .json file in the given directory into the Registry.
// Files with a "root" field are treated as DialogueTrees; those with "lines"
// are treated as SimpleDialogues.
func LoadAll(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("dialogue: readdir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("dialogue: skip %s: %v\n", path, err)
			continue
		}
		// Peek at the JSON to decide which loader to use.
		var probe struct {
			Root  string   `json:"root"`
			Lines []string `json:"lines"`
		}
		_ = json.Unmarshal(data, &probe)

		if probe.Root != "" {
			if _, err := LoadTree(path); err != nil {
				fmt.Printf("dialogue: %v\n", err)
			}
		} else if len(probe.Lines) > 0 {
			if _, err := LoadSimple(path); err != nil {
				fmt.Printf("dialogue: %v\n", err)
			}
		}
	}
	return nil
}

// SelectTree picks the appropriate tree ID for an NPC based on quest flags.
// When the player is in NG+ (npc_ng_plus > 0), it first checks whether an
// NG+-specific variant exists (e.g. "varn_ng_phase0") before falling back to
// the standard tree ("varn_phase0"). This allows NG+ questlines to diverge
// without touching the normal trees.
func SelectTree(npcID string, flags map[string]int) string {
	phase := flags[npcID+"_phase"]
	if flags[npcID+"_ng_plus"] > 0 {
		ngKey := fmt.Sprintf("%s_ng_phase%d", npcID, phase)
		if _, ok := Registry[ngKey]; ok {
			return ngKey
		}
	}
	return fmt.Sprintf("%s_phase%d", npcID, phase)
}
