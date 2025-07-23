package items

import "github.com/hajimehoshi/ebiten/v2"

// ItemType categorizes items for basic behavior.
type ItemType string

const (
	ItemWeapon     ItemType = "Weapon"
	ItemArmor      ItemType = "Armor"
	ItemConsumable ItemType = "Consumable"
	ItemKey        ItemType = "Key"
	ItemQuest      ItemType = "Quest"
	ItemMisc       ItemType = "Misc"
)

// ItemTemplate defines common data shared across item instances.
type ItemTemplate struct {
	ID          string
	Name        string
	Type        ItemType
	Description string
	Stackable   bool
	MaxStack    int
	Usable      bool
	Equippable  bool
	Stats       map[string]int
	Icon        *ebiten.Image
	OnUse       func(p interface{})
	OnEquip     func(p interface{})
	OnUnequip   func(p interface{})
}

// Item represents an inventory instance.
type Item struct {
	*ItemTemplate
	Count int
}

// ItemSave is a minimal representation for serialization.
type ItemSave struct {
	ID    string
	Count int
}

// ToSave converts an item instance to its save form.
func (i *Item) ToSave() ItemSave {
	return ItemSave{ID: i.ID, Count: i.Count}
}

// FromSave recreates an item from saved data.
func FromSave(data ItemSave) *Item {
	it := NewItem(data.ID)
	it.Count = data.Count
	return it
}

// SerializeEquipment converts an equipment map into savable data.
func SerializeEquipment(eq map[string]*Item) map[string]ItemSave {
	res := make(map[string]ItemSave)
	for slot, it := range eq {
		if it != nil {
			res[slot] = it.ToSave()
		}
	}
	return res
}

// DeserializeEquipment reconstructs an equipment map from saved data.
func DeserializeEquipment(data map[string]ItemSave) map[string]*Item {
	eq := make(map[string]*Item)
	for slot, sv := range data {
		if sv.ID != "" {
			eq[slot] = FromSave(sv)
		}
	}
	return eq
}
