package entities

import (
	"dungeoneer/inventory"
	"dungeoneer/items"
)

// Equip moves an item from the inventory grid to the specified equipment slot.
func (p *Player) Equip(slot string, gx, gy int) bool {
	if p.Inventory == nil || p.Equipment == nil {
		return false
	}
	if gy < 0 || gy >= p.Inventory.Height || gx < 0 || gx >= p.Inventory.Width {
		return false
	}
	it := p.Inventory.Grid[gy][gx]
	if it == nil || !it.Equippable {
		return false
	}
	if existing := p.Equipment[slot]; existing != nil {
		if !inventory.TryStack(p.Inventory, *existing) {
			x, y, ok := inventory.FirstEmpty(p.Inventory)
			if !ok {
				return false
			}
			p.Inventory.Grid[y][x] = existing
		}
	}
	p.Equipment[slot] = it
	p.Inventory.Grid[gy][gx] = nil
	if it.OnEquip != nil {
		it.OnEquip(p)
	}
	p.RecalculateStats()
	p.RefreshAbilities()
	return true
}

// Unequip removes an item from the given slot and returns it to the inventory.
func (p *Player) Unequip(slot string) bool {
	if p.Inventory == nil || p.Equipment == nil {
		return false
	}
	it := p.Equipment[slot]
	if it == nil {
		return false
	}
	if !inventory.TryStack(p.Inventory, *it) {
		x, y, ok := inventory.FirstEmpty(p.Inventory)
		if !ok {
			return false
		}
		p.Inventory.Grid[y][x] = it
	}
	if it.OnUnequip != nil {
		it.OnUnequip(p)
	}
	p.Equipment[slot] = nil
	p.RecalculateStats()
	p.RefreshAbilities()
	return true
}

// DropEquipped removes the item from the given equipment slot and returns it.
func (p *Player) DropEquipped(slot string) *items.Item {
	if p.Equipment == nil {
		return nil
	}
	it := p.Equipment[slot]
	if it == nil {
		return nil
	}
	if it.OnUnequip != nil {
		it.OnUnequip(p)
	}
	p.Equipment[slot] = nil
	p.RecalculateStats()
	p.RefreshAbilities()
	return it
}

// DropFromInventory removes count items from the specified grid cell and returns them.
func (p *Player) DropFromInventory(gx, gy int, count int) *items.Item {
	if p.Inventory == nil {
		return nil
	}
	if gy < 0 || gy >= p.Inventory.Height || gx < 0 || gx >= p.Inventory.Width {
		return nil
	}
	it := p.Inventory.Grid[gy][gx]
	if it == nil {
		return nil
	}
	if count <= 0 || count >= it.Count {
		p.Inventory.Grid[gy][gx] = nil
		return it
	}
	it.Count -= count
	return &items.Item{ItemTemplate: it.ItemTemplate, Count: count}
}

// HasItemAnywhere returns true if the player has an item with the given ID in
// their inventory grid or any equipped slot.
func (p *Player) HasItemAnywhere(id string) bool {
	if p.Inventory != nil && p.Inventory.HasItem(id) {
		return true
	}
	for _, it := range p.Equipment {
		if it != nil && it.ID == id {
			return true
		}
	}
	return false
}

// RemoveItemByID removes one item with the given ID from the player's inventory
// or equipment. Equipment slots trigger OnUnequip and RefreshAbilities.
// Returns true if an item was found and removed.
func (p *Player) RemoveItemByID(id string) bool {
	if p.Inventory != nil {
		for y := 0; y < p.Inventory.Height; y++ {
			for x := 0; x < p.Inventory.Width; x++ {
				if it := p.Inventory.Grid[y][x]; it != nil && it.ID == id {
					if it.Count > 1 {
						it.Count--
					} else {
						p.Inventory.Grid[y][x] = nil
					}
					return true
				}
			}
		}
	}
	for slot, it := range p.Equipment {
		if it != nil && it.ID == id {
			if it.OnUnequip != nil {
				it.OnUnequip(p)
			}
			p.Equipment[slot] = nil
			p.RecalculateStats()
			p.RefreshAbilities()
			return true
		}
	}
	return false
}

// AddToInventory attempts to add an item to the player's inventory.
// Returns true if successful, false if the inventory is full.
func (p *Player) AddToInventory(it *items.Item) bool {
	if p.Inventory == nil {
		return false
	}
	if inventory.TryStack(p.Inventory, *it) {
		return true
	}
	x, y, ok := inventory.FirstEmpty(p.Inventory)
	if !ok {
		return false
	}
	p.Inventory.Grid[y][x] = it
	return true
}
