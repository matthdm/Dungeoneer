📄 Design Document: items Package

Feature Name: Item Management System
Game: Dungeoneer
Related Systems: Inventory, Equipment, Loot, Player Interactions
🎯 Design Goals

    Define a modular, extensible system for in-game items

    Support item types (consumables, weapons, armor, keys, etc.)

    Enable item usage, stacking, equipping, and serialization

    Allow custom behavior via hooks (e.g. OnUse)

    Integrate with the inventory system's 5×4 grid and equipment slots

    Use centralized item definitions and loading for consistency

📦 Package Overview

Path: /items

This package contains:

    Item type definitions

    Global item registry

    Item behavior hooks

    Utilities for use, copy, stack, and render logic

🔩 Core Data Structures
ItemType

type ItemType string

const (
	ItemWeapon     ItemType = "Weapon"
	ItemArmor      ItemType = "Armor"
	ItemConsumable ItemType = "Consumable"
	ItemQuest      ItemType = "Quest"
	ItemKey        ItemType = "Key"
	ItemMisc       ItemType = "Misc"
)

Item

type Item struct {
	ID          string            // Unique ID (e.g. "health_potion")
	Name        string            // Human-readable name
	Description string            // Tooltip flavor text
	Icon        *ebiten.Image     // 64x64 sprite
	Type        ItemType
	Stackable   bool
	MaxStack    int
	Count       int
	Usable      bool
	Equippable  bool
	Stats       map[string]int    // e.g. {"Damage": 5, "Armor": 3}
	OnUse       func(p *entities.Player) // Optional behavior
}

Notes:

    Items are passed by pointer in inventory

    If Stackable, merging logic uses MaxStack

    OnUse is only executed if Usable == true

ItemTemplate

Used internally for instantiating items from global definitions:

type ItemTemplate struct {
	ID          string
	Type        ItemType
	Name        string
	Description string
	Stackable   bool
	MaxStack    int
	Usable      bool
	Equippable  bool
	Stats       map[string]int
	IconPath    string // Path to icon image
	OnUse       func(p *entities.Player)
}

📋 Item Registry

All items are defined at load time using a global registry:

var Registry = map[string]*ItemTemplate{}

func RegisterItem(template *ItemTemplate) {
	Registry[template.ID] = template
}

Example usage:

items.RegisterItem(&items.ItemTemplate{
	ID:        "health_potion",
	Name:      "Health Potion",
	Type:      items.ItemConsumable,
	Stackable: true,
	MaxStack:  5,
	Usable:    true,
	Description: "Restores 20 HP.",
	OnUse: func(p *entities.Player) {
		p.HP = min(p.MaxHP, p.HP+20)
	},
})

Create an Instance

func NewItem(id string) *Item {
	tmpl, ok := Registry[id]
	if !ok {
		panic("invalid item ID: " + id)
	}
	return &Item{
		ID:         tmpl.ID,
		Name:       tmpl.Name,
		Description: tmpl.Description,
		Icon:       LoadIcon(tmpl.IconPath),
		Type:       tmpl.Type,
		Stackable:  tmpl.Stackable,
		MaxStack:   tmpl.MaxStack,
		Count:      1,
		Usable:     tmpl.Usable,
		Equippable: tmpl.Equippable,
		Stats:      tmpl.Stats,
		OnUse:      tmpl.OnUse,
	}
}

🧪 Integration Points
System	Integration
Inventory	Items populate [][]*Item grid
UI	Tooltip renders Name, Stats, etc.
Player	Item usage applies OnUse() effects
Combat	Equipped weapon stats modify damage
Levels	Dropped items on ground (future)
📂 Serialization

When saving an item:

type ItemSave struct {
	ID    string
	Count int
}

When loading:

func FromSave(s ItemSave) *Item {
	item := NewItem(s.ID)
	item.Count = s.Count
	return item
}

🔮 Future Extensions
Feature	Integration Notes
Item rarity	Add Rarity string and color overlays
Modifiers/enchants	Add dynamic Modifiers []Modifier
Crafting	Use items.Combine(a, b) or RecipeMap
Item cooldowns	Track item LastUsed with time.Time
Drop on death	Convert item to entity in world map
📜 Codex Prompt

    You are working on a 2D isometric dungeon crawler using Ebiten in Go. Implement a modular items package that integrates with a 5x4 grid inventory system. Items can be weapons, armor, consumables, quest items, etc. Each item has an ID, name, type, stats, an optional OnUse(*Player) hook, and may be stackable. Include a global item registry that loads item templates and supports instantiating live Item objects. Support item serialization and usage from the inventory system.