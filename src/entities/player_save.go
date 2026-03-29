package entities

import (
	"dungeoneer/collision"
	"dungeoneer/constants"
	"dungeoneer/images"
	"dungeoneer/inventory"
	"dungeoneer/items"
	"dungeoneer/movement"
	"dungeoneer/spells"
)

// PlayerSave represents the serializable player state.
type PlayerSave struct {
	Name      string                    `json:"name"`
	TileX     int                       `json:"tile_x"`
	TileY     int                       `json:"tile_y"`
	Class     PlayerClass               `json:"class,omitempty"`
	Stats     BaseStats                 `json:"stats"`
	Inventory [][]items.ItemSave        `json:"inventory"`
	Equipment map[string]items.ItemSave `json:"equipment"`
	HP        int                       `json:"hp"`
	Mana      int                       `json:"mana"`
	Level     int                       `json:"level"`
	EXP       int                       `json:"exp"`
	Points    int                       `json:"points"`
	Gold      int                       `json:"gold"`
}

// ToSaveData converts the player to a serializable form.
func (p *Player) ToSaveData() PlayerSave {
	return PlayerSave{
		Name:      p.Name,
		TileX:     p.TileX,
		TileY:     p.TileY,
		Class:     p.Class,
		Stats:     p.Stats,
		HP:        p.HP,
		Mana:      p.Mana,
		Level:     p.Level,
		EXP:       p.EXP,
		Points:    p.UnspentPoints,
		Gold:      p.Gold,
		Inventory: p.Inventory.ToSaveData(),
		Equipment: items.SerializeEquipment(p.Equipment),
	}
}

// LoadPlayer reconstructs a Player from saved data.
func LoadPlayer(data PlayerSave) *Player {
	mc := movement.NewMovementController(5)
	mc.InterpX = float64(data.TileX)
	mc.InterpY = float64(data.TileY)

	blackMage, _ := images.LoadEmbeddedImage(images.Black_Mage_Full_png)

	equipment := NewEquipmentSlots()
	for slot, it := range items.DeserializeEquipment(data.Equipment) {
		equipment[slot] = it
	}
	p := &Player{
		TileX:         data.TileX,
		TileY:         data.TileY,
		LeftFacing:    true,
		Sprite:        blackMage,
		Stats:         data.Stats,
		TempModifiers: StatModifiers{},
		CollisionBox:  collision.Box{X: float64(data.TileX), Y: float64(data.TileY) - 0.4, Width: 0.55, Height: 0.8},
		DashCharges:   constants.MaxDashCharges,
		Grapple: Grapple{
			MaxDistance: constants.GrappleMaxDistance,
			Speed:       constants.GrappleSpeed,
			Delay:       constants.GrappleDelay,
		},
		Inventory:      inventory.FromSaveData(data.Inventory),
		Equipment:      equipment,
		HP:             data.HP,
		Mana:           data.Mana,
		Level:          data.Level,
		EXP:            data.EXP,
		UnspentPoints:  data.Points,
		Gold:           data.Gold,
		Name:           data.Name,
		Class:          normalizeSavedClass(data.Class, equipment),
		LastMoveDirX:   -1,
		LastMoveDirY:   0,
		MoveController: mc,
		Caster:         spells.NewCaster(),
	}

	mc.OnStep = func(x, y int) {
		p.TileX = x
		p.TileY = y
	}

	p.RecalculateStats()
	p.RefreshAbilities()
	return p
}

func normalizeSavedClass(saved PlayerClass, equipment map[string]*items.Item) PlayerClass {
	if saved == ClassKnight || saved == ClassMage {
		return saved
	}
	for _, it := range equipment {
		if it == nil {
			continue
		}
		switch it.GrantsAbility {
		case "slash_combo":
			return ClassKnight
		case "arcane_bolt":
			return ClassMage
		}
	}
	return ClassMage
}
