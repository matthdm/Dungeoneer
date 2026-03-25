package entities

import "math"

// IdleBehavior makes an NPC stand still and face the player when nearby.
type IdleBehavior struct {
	FacePlayerRadius float64
}

// NewIdleBehavior creates an IdleBehavior that faces the player within radius tiles.
func NewIdleBehavior(faceRadius float64) *IdleBehavior {
	return &IdleBehavior{FacePlayerRadius: faceRadius}
}

func (b *IdleBehavior) Update(npc *NPC, playerX, playerY int, dt float64) {
	npc.TickCount++
	npc.BobOffset = math.Sin(float64(npc.TickCount)*0.1) * 1.0

	dx := playerX - npc.TileX
	dy := playerY - npc.TileY
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	if dist <= b.FacePlayerRadius {
		if dx < 0 {
			npc.LeftFacing = true
		} else if dx > 0 {
			npc.LeftFacing = false
		}
	}
}
