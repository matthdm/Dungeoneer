package entities

import (
	"dungeoneer/images"
	"dungeoneer/inventory"
	"dungeoneer/items"
	"dungeoneer/movement"
	"dungeoneer/spells"
)

// PlayerSave represents the serializable player state.
type PlayerSave struct {
	Name            string                    `json:"name"`
	TileX           int                       `json:"tile_x"`
	TileY           int                       `json:"tile_y"`
	Stats           BaseStats                 `json:"stats"`
	Inventory       [][]items.ItemSave        `json:"inventory"`
	Equipment       map[string]items.ItemSave `json:"equipment"`
	HP              int                       `json:"hp"`
	Mana            int                       `json:"mana"`
	Level           int                       `json:"level"`
	EXP             int                       `json:"exp"`
	AttributePoints int                       `json:"ap"`
}

// ToSaveData converts the player to a serializable form.
func (p *Player) ToSaveData() PlayerSave {
	return PlayerSave{
		Name:            p.Name,
		TileX:           p.TileX,
		TileY:           p.TileY,
		Stats:           p.Stats,
		HP:              p.HP,
		Mana:            p.Mana,
		Level:           p.Level,
		EXP:             p.EXP,
		AttributePoints: p.AttributePoints,
		Inventory:       p.Inventory.ToSaveData(),
		Equipment:       items.SerializeEquipment(p.Equipment),
	}
}

// LoadPlayer reconstructs a Player from saved data.
func LoadPlayer(data PlayerSave) *Player {
	mc := movement.NewMovementController(5)
	mc.InterpX = float64(data.TileX)
	mc.InterpY = float64(data.TileY)

	blackMage, _ := images.LoadEmbeddedImage(images.Black_Mage_Full_png)

	p := &Player{
		TileX:           data.TileX,
		TileY:           data.TileY,
		LeftFacing:      true,
		Sprite:          blackMage,
		Stats:           data.Stats,
		TempModifiers:   StatModifiers{},
		Inventory:       inventory.FromSaveData(data.Inventory),
		Equipment:       items.DeserializeEquipment(data.Equipment),
		HP:              data.HP,
		Mana:            data.Mana,
		Level:           data.Level,
		EXP:             data.EXP,
		AttributePoints: data.AttributePoints,
		Name:            data.Name,
		MoveController:  mc,
		Caster:          spells.NewCaster(),
	}

	mc.OnStep = func(x, y int) {
		p.TileX = x
		p.TileY = y
	}

	p.RecalculateStats()
	return p
}
