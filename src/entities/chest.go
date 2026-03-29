package entities

import "github.com/hajimehoshi/ebiten/v2"

// ChestVariant constants for the chest loot tier.
const (
	ChestWooden = "wooden" // common loot
	ChestIron   = "iron"   // uncommon-weighted loot
	ChestGold   = "gold"   // rare-weighted loot
	ChestLocked = "locked" // requires a key; rare/legendary loot
)

// Chest is a static interactable loot container placed in treasure rooms.
type Chest struct {
	TileX, TileY int
	Variant      string // ChestWooden / ChestIron / ChestGold / ChestLocked
	Opened       bool
	Sprite       *ebiten.Image
}

// IsPlayerInRange returns true if the player is within 1.5 tiles of this chest.
func (c *Chest) IsPlayerInRange(px, py float64) bool {
	dx := px - float64(c.TileX)
	dy := py - float64(c.TileY)
	return dx*dx+dy*dy <= 2.25 // 1.5^2
}
