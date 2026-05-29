package entities

import (
	"dungeoneer/inventory"
	"dungeoneer/items"
)

// PlayerSnapshot is a serializable capture of player state for run saves.
type PlayerSnapshot struct {
	TileX     int                       `json:"tile_x"`
	TileY     int                       `json:"tile_y"`
	InterpX   float64                   `json:"interp_x"`
	InterpY   float64                   `json:"interp_y"`
	HP        int                       `json:"hp"`
	MaxHP     int                       `json:"max_hp"`
	Mana      int                       `json:"mana"`
	MaxMana   int                       `json:"max_mana"`
	Inventory [][]items.ItemSave        `json:"inventory"`
	Equipment map[string]items.ItemSave `json:"equipment"`
}

// ToSnapshot captures the player's current state as a serializable snapshot.
func (p *Player) ToSnapshot() PlayerSnapshot {
	var invData [][]items.ItemSave
	if p.Inventory != nil {
		invData = p.Inventory.ToSaveData()
	}
	return PlayerSnapshot{
		TileX:     p.TileX,
		TileY:     p.TileY,
		InterpX:   p.MoveController.InterpX,
		InterpY:   p.MoveController.InterpY,
		HP:        p.HP,
		MaxHP:     p.MaxHP,
		Mana:      p.Mana,
		MaxMana:   p.MaxMana,
		Inventory: invData,
		Equipment: items.SerializeEquipment(p.Equipment),
	}
}

// ApplySnapshot restores player state from a snapshot.
// Call RefreshAbilities() after this to rebuild the spell bar.
func (p *Player) ApplySnapshot(snap PlayerSnapshot) {
	p.TileX = snap.TileX
	p.TileY = snap.TileY
	p.MoveController.InterpX = snap.InterpX
	p.MoveController.InterpY = snap.InterpY
	p.MoveController.TargetX = snap.InterpX
	p.MoveController.TargetY = snap.InterpY
	p.HP = snap.HP
	p.MaxHP = snap.MaxHP
	p.Mana = snap.Mana
	p.MaxMana = snap.MaxMana
	if snap.Inventory != nil {
		p.Inventory = inventory.FromSaveData(snap.Inventory)
	}
	if snap.Equipment != nil {
		p.Equipment = items.DeserializeEquipment(snap.Equipment)
	}
}
