package entities

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// ChestVariant constants for the chest loot tier.
const (
	ChestWooden = "wooden" // common loot
	ChestIron   = "iron"   // uncommon-weighted loot
	ChestGold   = "gold"   // rare-weighted loot
	ChestLocked = "locked" // requires a key; rare/legendary loot (legacy loot-path identifier)
)

// Floor thresholds for chest variant selection.
const (
	ChestIronMinFloor = 3  // Iron chests appear from floor 3+
	ChestGoldMinFloor = 6  // Gold chests appear from floor 6+
)

// Locked-chest probability by visual tier (0.0–1.0).
const (
	ChestLockedChanceWooden = 0.0  // Wooden chests are never locked
	ChestLockedChanceIron   = 0.4  // 40% of Iron chests are locked
	ChestLockedChanceGold   = 0.7  // 70% of Gold chests are locked
)

// Chest is a static interactable loot container placed in treasure rooms.
type Chest struct {
	TileX, TileY int
	Variant      string // ChestWooden / ChestIron / ChestGold (visual tier)
	Locked       bool   // Requires a key to open; orthogonal to Variant
	Opened       bool
	Sprite       *ebiten.Image
}

// IsPlayerInRange returns true if the player is within 1.5 tiles of this chest.
func (c *Chest) IsPlayerInRange(px, py float64) bool {
	dx := px - float64(c.TileX)
	dy := py - float64(c.TileY)
	return dx*dx+dy*dy <= 2.25 // 1.5^2
}

// ChestVariantTint returns a color tint for the given chest variant.
// Wooden chests are rendered without tint (white), while Iron and Gold
// chests receive distinctive color overlays.
func ChestVariantTint(variant string) color.RGBA {
	switch variant {
	case ChestWooden:
		return color.RGBA{255, 255, 255, 255} // no tint
	case ChestIron:
		return color.RGBA{180, 200, 180, 255} // grey-green
	case ChestGold:
		return color.RGBA{255, 215, 80, 255} // gold
	default:
		return color.RGBA{255, 255, 255, 255} // safe default
	}
}
