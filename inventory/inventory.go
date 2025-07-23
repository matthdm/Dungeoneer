package inventory

import "dungeoneer/items"

const (
	Width  = 5
	Height = 4
)

// Inventory holds a fixed grid of item pointers.
type Inventory struct {
	Grid [][]*items.Item
}

// NewInventory creates an empty inventory.
func NewInventory() *Inventory {
	inv := &Inventory{Grid: make([][]*items.Item, Height)}
	for i := range inv.Grid {
		inv.Grid[i] = make([]*items.Item, Width)
	}
	return inv
}

// AddItem places an item into the first available slot or stacks when possible.
func (inv *Inventory) AddItem(it *items.Item) {
	// try stacking
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			slot := inv.Grid[y][x]
			if slot != nil && slot.ID == it.ID && slot.Stackable && slot.Count < slot.MaxStack {
				needed := slot.MaxStack - slot.Count
				if it.Count <= needed {
					slot.Count += it.Count
					return
				}
				slot.Count = slot.MaxStack
				it.Count -= needed
			}
		}
	}
	// place in empty slot
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			if inv.Grid[y][x] == nil {
				inv.Grid[y][x] = it
				return
			}
		}
	}
	// inventory full; drop item (ignored for now)
}
