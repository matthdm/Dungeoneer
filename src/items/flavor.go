package items

import (
	"dungeoneer/gamedata"
	"encoding/json"
)

// ItemFlavorEntry is one entry in items_flavor.json.
type ItemFlavorEntry struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Line string `json:"line"`
}

// LoadItemFlavor reads items_flavor.json from path and patches matching
// entries in the registry with FlavorText and FlavorLine.
func LoadItemFlavor(path string) error {
	data, err := gamedata.ReadFile(path)
	if err != nil {
		return err
	}
	var entries []ItemFlavorEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		if tmpl, ok := Registry[e.ID]; ok {
			tmpl.FlavorText = e.Text
			tmpl.FlavorLine = e.Line
		}
	}
	return nil
}
