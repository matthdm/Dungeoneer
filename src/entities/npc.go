package entities

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// NPCBehavior defines how an NPC acts each tick.
type NPCBehavior interface {
	Update(npc *NPC, playerX, playerY int, dt float64)
}

// NPC represents a non-player character in the world.
type NPC struct {
	// Identity
	ID    string
	Name  string
	Title string

	// Position (mirrors Monster pattern)
	TileX, TileY     int
	InterpX, InterpY float64
	LeftFacing       bool

	// Visuals
	Sprite     *ebiten.Image
	PortraitID string
	BobOffset  float64
	TickCount  int

	// NPC state
	IsMajor       bool
	Phase         int
	DialogueID    string
	Interactable  bool
	InteractRange float64 // tile distance for interaction (default 1.5)

	// Behavior
	Behavior NPCBehavior
}

// Update ticks the NPC's behavior and animation.
func (n *NPC) Update(playerX, playerY int, dt float64) {
	if n.Behavior != nil {
		n.Behavior.Update(n, playerX, playerY, dt)
	}
}

// Draw renders the NPC on screen, following the Monster.Draw pattern.
func (n *NPC) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if n.Sprite == nil {
		return
	}

	x, y := isoToScreenFloat(n.InterpX, n.InterpY, tileSize)

	op := &ebiten.DrawImageOptions{}
	bounds := n.Sprite.Bounds()
	spriteW := float64(bounds.Dx())

	// Bob animation
	const verticalOffset = 1.0
	op.GeoM.Translate(0, -verticalOffset+n.BobOffset)

	// Flip for facing direction
	if !n.LeftFacing {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteW, 0)
	}

	// Move to screen space and apply camera
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	screen.DrawImage(n.Sprite, op)
}

// IsPlayerInRange returns true if the player is close enough to interact.
func (n *NPC) IsPlayerInRange(px, py int) bool {
	return n.IsPlayerInRangeAt(float64(px), float64(py))
}

// IsPlayerInRangeAt checks interaction radius using world-space player coordinates.
func (n *NPC) IsPlayerInRangeAt(px, py float64) bool {
	if n.InteractRange <= 0 {
		dx := px - float64(n.TileX)
		dy := py - float64(n.TileY)
		return dx*dx+dy*dy <= 4 // default ~2 tiles
	}
	dx := px - n.InterpX
	dy := py - n.InterpY
	return math.Sqrt(dx*dx+dy*dy) <= n.InteractRange
}
