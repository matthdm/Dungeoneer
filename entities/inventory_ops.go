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
