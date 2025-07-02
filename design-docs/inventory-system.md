📄 Design Document: Inventory System

Feature Name: Inventory System
Game: Dungeoneer
Inspiration: Minecraft, Pixel Dungeon
Status: High-priority gameplay system
🎯 Design Goals

    Allow players to collect, manage, equip, and use items in a responsive, visual way

    Support core item types (weapons, armor, consumables, keys, quest items)

    Blend Minecraft-style drag-and-drop with Pixel Dungeon's compact slot logic

    Integrate seamlessly with combat, equipment, and future progression systems

    Allow for visual rendering and input handling in Ebiten

🔩 Core Mechanics
Inventory Structure

    Fixed 5x4 grid (20 slots), with support for:

        Item stacking (e.g., potions)

        Equipment slots (weapon, armor, etc.)

    Drag-and-drop UI for managing item positions

    Items can be picked up, dropped, consumed, equipped, or discarded

Slot Types
Slot Type	Description
Inventory Grid	Main 5x4 grid for general storage
Hotbar (optional)	5-slot quick access row
Equipment Slots	Weapon, Armor, Trinket
Item Structure

type Item struct {
	ID          string // e.g. "health_potion"
	Name        string
	Icon        *ebiten.Image
	Type        ItemType // "Consumable", "Weapon", "Armor"
	Stackable   bool
	Count       int
	Stats       map[string]int
	Usable      bool
	Description string
	OnUse       func(p *Player)
}

🧱 Inventory Data Structures

type Inventory struct {
	Grid       [][]*Item // 5x4 grid
	Selected   Point     // Cursor position
	Dragging   *Item     // Currently dragged item
	Equipment  map[string]*Item // e.g. {"Weapon": item}
}

Item Placement

    Items are moved by clicking and dragging

    If dropped on an occupied slot:

        If same item type and stackable → merge

        Else → swap items

🔁 Item Types and Behavior
Type	Example	Behavior
Weapon	Sword, Bow	Equippable
Armor	Leather Armor	Equippable
Consumable	Potion	Usable → triggers OnUse()
Quest Item	Gem Key	Not usable or droppable
🎮 Controls
Input	Action
Click & drag	Move item between slots
Right-click	Use item if usable
Ctrl + click	Split stack (if stackable)
Hover	Show tooltip
ESC / I	Open/close inventory
🖼️ Visual System
Layout (Ebiten)

    Draw semi-transparent background

    Grid rendered with 64x64 slots

    Items drawn centered in each slot

    Equipment panel appears on right side

    Tooltip shows item name/stats on hover

Rendering Flow

func (inv *Inventory) Draw(screen *ebiten.Image) {
	drawInventoryGrid(screen)
	drawItems(screen)
	drawDraggingItem(screen)
	drawEquipmentPanel(screen)
	drawTooltip(screen)
}

🧪 Item Usage API

All items may optionally define an OnUse(*Player) function:

item := &Item{
	ID: "health_potion",
	Usable: true,
	OnUse: func(p *Player) {
		p.HP = min(p.MaxHP, p.HP + 20)
	},
}

When a player right-clicks the item, OnUse() is called and the item is removed or decremented.
📂 Persistence

Inventory should serialize with player save data:

type PlayerSave struct {
	InventoryData [][]ItemSave
	EquipmentData map[string]ItemSave
}

🔮 Expansion Hooks
Feature	Design Consideration
Inventory Upgrades	Expand rows or slots
Item Rarity	Add color overlays
Tooltips with flavor text	Add multiline support
Item crafting	Merge items in crafting UI
Drop on death	Items spill on ground
🧠 Codex Prompt

    You're building a 2D isometric dungeon crawler using Ebiten in Go. Implement an inventory system inspired by Minecraft and Pixel Dungeon. It should support a 5x4 item grid, equipment slots (weapon, armor, etc.), drag-and-drop mouse control, item stacking, right-click usage, and tooltips. The Item struct should include fields like ID, Name, Icon, Stackable, Usable, and OnUse(*Player). Create an Inventory struct with rendering and input handling logic. Use Ebiten’s DrawImage API to visually draw the grid and items.