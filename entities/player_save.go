package entities

import (
	"dungeoneer/inventory"
	"dungeoneer/items"
	"dungeoneer/movement"
)

// PlayerSave represents the serializable player state.
type PlayerSave struct {
	Name      string                    `json:"name"`
	TileX     int                       `json:"tile_x"`
	TileY     int                       `json:"tile_y"`
	Stats     BaseStats                 `json:"stats"`
	Inventory [][]items.ItemSave        `json:"inventory"`
	Equipment map[string]items.ItemSave `json:"equipment"`
	HP        int                       `json:"hp"`
	Mana      int                       `json:"mana"`
}

// ToSaveData converts the player to a serializable form.
func (p *Player) ToSaveData() PlayerSave {
	return PlayerSave{
		Name:      p.Name,
		TileX:     p.TileX,
		TileY:     p.TileY,
		Stats:     p.Stats,
		HP:        p.HP,
		Mana:      p.Mana,
		Inventory: p.Inventory.ToSaveData(),
		Equipment: items.SerializeEquipment(p.Equipment),
	}
}

// LoadPlayer reconstructs a Player from saved data.
func LoadPlayer(data PlayerSave) *Player {
	p := &Player{
		TileX:          data.TileX,
		TileY:          data.TileY,
		Stats:          data.Stats,
		TempModifiers:  StatModifiers{},
		Inventory:      inventory.FromSaveData(data.Inventory),
		Equipment:      items.DeserializeEquipment(data.Equipment),
		HP:             data.HP,
		Mana:           data.Mana,
		Name:           data.Name,
		MoveController: movement.NewMovementController(5),
	}
	p.RecalculateStats()
	return p
}
