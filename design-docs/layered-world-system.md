📄 Design Document: Layered World System
Feature Name: Layered World
Game: Dungeoneer
Priority: High

🧠 Concept Overview
The Layered World system introduces the ability for levels to contain multiple vertically stacked isometric layers ("floors"). These layers represent spatially or thematically distinct subzones (e.g., underground crypts, upper walkways, hidden void realms), but are part of the same logical level.

Initially, transitions between layers will be handled by stairwell sprites, and only the currently active layer is visible and updated. Future expansions will allow for item-based layer shifting to access hidden or alternate realities.

🎯 Design Goals
Support Multiple Vertical Floors Per Level, optionally with different grid dimensions.

Enable Smooth Player Traversal via interactive stairwell entities or triggers.

Render Only the Active Layer, preserving performance and visual clarity.

Allow Future Extensions: time/void shifting, ghost layers, secret floors.

🧱 System Components
🧩 1. LayeredLevel Structure
A new type that wraps multiple Level instances:

go
Copy
Edit
type LayeredLevel struct {
	Layers      []*Level          // Each floor (0 = default/main)
	ActiveIndex int               // Which floor the player is on
	Stairwells  []*LayerLink      // List of inter-layer links
}
🧩 2. Level (Existing)
Your current level structure will remain intact, but the LayeredLevel will manage which one is active.

🔄 Inter-Layer Navigation
LayerLink Struct
Represents a connection point (e.g., stairwell) between two layers.

go
Copy
Edit
type LayerLink struct {
	FromLayerIndex int
	FromTile       Point
	ToLayerIndex   int
	ToTile         Point
	TriggerSprite  string // ID or path to stairwell sprite
	IsOneWay       bool
}
When the player collides with or interacts with a LayerLink sprite, they are transferred to the corresponding layer and tile.

🧭 Layer Behavior and Transition
Action	Result
Player steps on stairwell	Switches to destination layer
New layer activated	Camera recenters, FOV resets
Only active layer updates	Improves performance, avoids drawing/logic confusion
All other layers paused	Monsters/entities are frozen until re-entry

🪜 Asymmetrical Layer Support
Each Level inside LayeredLevel can have its own width, height, tilemap, and tile entities:

Main: 64x64

Sub-layer A: 10x10 secret basement

Sub-layer B: 32x32 cursed chapel

These layers may have different:

Biomes/themes

Lighting setups

Level-specific rules

🖼️ Visual Rendering
Only the ActiveIndex layer is rendered:

go
Copy
Edit
func (ll *LayeredLevel) Draw(screen *ebiten.Image) {
	ll.Layers[ll.ActiveIndex].Draw(screen)
}
Optional:

Fade in/out or flash when switching layers

Display a floating text like “Floor B1” for orientation

Disable minimap or restrict it to current layer only

🎮 Gameplay Flow
Player starts on Floor 0 (main map).

Finds stairwell (linked via LayerLink) at (10, 5).

Walks on sprite → transition triggered → moved to Floor 1, position (3, 3).

Game swaps active layer; only new layer is drawn and updated.

Player completes objective, finds return stairwell, moves back to Floor 0.

🔐 Layer Transition API
go
Copy
Edit
func (ll *LayeredLevel) SwitchToLayer(index int, entry Point) {
	ll.ActiveIndex = index
	ll.Layers[index].Player.TileX = entry.X
	ll.Layers[index].Player.TileY = entry.Y
	// Reset camera, FOV, etc.
}
🔮 Future Expansion: Item-Based Shifting
Introduce an item (e.g., Phantom Veil, Reality Sigil) that allows player to:

Temporarily shift to another layer without stairwell

Reveal secret rooms by phasing between layers

Implementation Plan:

Add optional "phantom" layer flag

Support keybinding to trigger ShiftLayer()

Visual shader/overlay effect when shifting

🧪 Dev & Debug Notes
Add DebugLayerOverlay for dev builds to visualize all layers.

In Level Editor: add tabbing or dropdown to manage layers independently.

Ensure enemy AI is paused in inactive layers to prevent logic bugs.

🧠 Codex Prompt
You’re working on a 2D isometric dungeon crawler in Go using the Ebiten engine. Implement a LayeredLevel system that contains multiple Level instances stacked vertically (floors). Each floor may have a different size (e.g., 64x64, 10x10). The player can traverse between layers via stairwell sprites, which act as LayerLink triggers. Only the active layer should be updated and rendered. Design the LayeredLevel and LayerLink structures, and implement player transfer between layers. Later, add support for item-triggered layer shifting. Also design a save/load format (.layeredlevel.json) that serializes all floors and stairwell connections. Update the existing leveleditor to support editing layered levels by switching between layers and editing tiles/entities per layer.