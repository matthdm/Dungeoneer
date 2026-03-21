package inventory

import "dungeoneer/items"

const (
	Width  = 5
	Height = 4
)

// Inventory holds a grid of item pointers.
// Width and Height define the grid dimensions to keep
// save data stable even if the constructor allows custom sizes.
type Inventory struct {
	Grid          [][]*items.Item
	Width, Height int
}

// ToSaveData serializes the inventory grid into a 2D slice of ItemSave.
func (inv *Inventory) ToSaveData() [][]items.ItemSave {
	data := make([][]items.ItemSave, inv.Height)
	for y := 0; y < inv.Height; y++ {
		row := make([]items.ItemSave, inv.Width)
		for x := 0; x < inv.Width; x++ {
			if it := inv.Grid[y][x]; it != nil {
				row[x] = it.ToSave()
			}
		}
		data[y] = row
	}
	return data
}

// FromSaveData reconstructs an Inventory from saved data.
func FromSaveData(data [][]items.ItemSave) *Inventory {
	h := Height
	if len(data) > 0 {
		h = len(data)
	}
	w := Width
	if len(data) > 0 && len(data[0]) > 0 {
		w = len(data[0])
	}
	inv := New(w, h)
	for y := 0; y < len(data) && y < inv.Height; y++ {
		for x := 0; x < len(data[y]) && x < inv.Width; x++ {
			if data[y][x].ID != "" {
				inv.Grid[y][x] = items.FromSave(data[y][x])
			}
		}
	}
	return inv
}

// New creates an empty inventory with the given dimensions.
func New(w, h int) *Inventory {
	inv := &Inventory{
		Grid:   make([][]*items.Item, h),
		Width:  w,
		Height: h,
	}
	for y := 0; y < h; y++ {
		inv.Grid[y] = make([]*items.Item, w)
	}
	return inv
}

// AddItem places an item into the first available slot or stacks when possible.
// It returns true if the entire item stack was added, false if inventory was full.
func (inv *Inventory) AddItem(it *items.Item) bool {
	// try stacking
	if TryStack(inv, *it) {
		return true
	}
	// place in empty slot
	for y := 0; y < inv.Height; y++ {
		for x := 0; x < inv.Width; x++ {
			if inv.Grid[y][x] == nil {
				inv.Grid[y][x] = it
				return true
			}
		}
	}
	// inventory full
	return false
}

// TryStack attempts to merge the given item into an existing stack.
// It returns true only if the entire item count was merged.
func TryStack(inv *Inventory, it items.Item) bool {
	for y := 0; y < inv.Height; y++ {
		for x := 0; x < inv.Width; x++ {
			slot := inv.Grid[y][x]
			if slot != nil && slot.ID == it.ID && slot.Stackable && slot.Count < slot.MaxStack {
				space := slot.MaxStack - slot.Count
				if it.Count <= space {
					slot.Count += it.Count
					return true
				}
			}
		}
	}
	return false
}

// FirstEmpty returns the coordinates of the first empty grid cell.
func FirstEmpty(inv *Inventory) (x, y int, ok bool) {
	for y = 0; y < inv.Height; y++ {
		for x = 0; x < inv.Width; x++ {
			if inv.Grid[y][x] == nil {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}
