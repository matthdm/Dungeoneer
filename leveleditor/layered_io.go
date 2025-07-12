package leveleditor

import (
	"dungeoneer/levels"
	"encoding/json"
	"os"
)

// LayerLinkData serializes a link between layers.
type LayerLinkData struct {
	FromLayerIndex int          `json:"from_layer"`
	FromTile       levels.Point `json:"from_tile"`
	ToLayerIndex   int          `json:"to_layer"`
	ToTile         levels.Point `json:"to_tile"`
	TriggerSprite  string       `json:"trigger_sprite"`
	IsOneWay       bool         `json:"is_one_way"`
}

// LayeredLevelData is the serializable representation of a layered level.
type LayeredLevelData struct {
	Layers      []*LevelData    `json:"layers"`
	ActiveIndex int             `json:"active_index"`
	Stairwells  []LayerLinkData `json:"stairwells"`
}

// ConvertToLayeredLevelData converts a layered level to its serializable form.
func ConvertToLayeredLevelData(ll *levels.LayeredLevel) *LayeredLevelData {
	layers := make([]*LevelData, len(ll.Layers))
	for i, l := range ll.Layers {
		layers[i] = ConvertToLevelData(l)
	}
	links := make([]LayerLinkData, len(ll.Stairwells))
	for i, link := range ll.Stairwells {
		links[i] = LayerLinkData{
			FromLayerIndex: link.FromLayerIndex,
			FromTile:       link.FromTile,
			ToLayerIndex:   link.ToLayerIndex,
			ToTile:         link.ToTile,
			TriggerSprite:  link.TriggerSprite,
			IsOneWay:       link.IsOneWay,
		}
	}
	return &LayeredLevelData{Layers: layers, ActiveIndex: ll.ActiveIndex, Stairwells: links}
}

// ConvertToLayeredLevel converts serialized data into a layered level.
func ConvertToLayeredLevel(data *LayeredLevelData) *levels.LayeredLevel {
	ll := &levels.LayeredLevel{ActiveIndex: data.ActiveIndex}
	ll.Layers = make([]*levels.Level, len(data.Layers))
	for i, ld := range data.Layers {
		ll.Layers[i] = ConvertToLevel(ld)
	}
	ll.Stairwells = make([]*levels.LayerLink, len(data.Stairwells))
	for i, l := range data.Stairwells {
		ll.Stairwells[i] = &levels.LayerLink{
			FromLayerIndex: l.FromLayerIndex,
			FromTile:       l.FromTile,
			ToLayerIndex:   l.ToLayerIndex,
			ToTile:         l.ToTile,
			TriggerSprite:  l.TriggerSprite,
			IsOneWay:       l.IsOneWay,
		}
	}
	return ll
}

// SaveLayeredLevelToFile writes a layered level to disk as JSON.
func SaveLayeredLevelToFile(ll *levels.LayeredLevel, path string) error {
	data := ConvertToLayeredLevelData(ll)
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0644)
}

// LoadLayeredLevelFromFile loads a layered level from disk.
func LoadLayeredLevelFromFile(path string) (*levels.LayeredLevel, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data LayeredLevelData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return ConvertToLayeredLevel(&data), nil
}
