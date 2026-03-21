package entities

import (
	"encoding/json"
	"os"
)

// SavePlayerToFile saves the player state to a JSON file.
func SavePlayerToFile(p *Player, path string) error {
	data := p.ToSaveData()
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0644)
}

// LoadPlayerFromFile loads a player from a JSON file.
func LoadPlayerFromFile(path string) (*Player, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data PlayerSave
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return LoadPlayer(data), nil
}
