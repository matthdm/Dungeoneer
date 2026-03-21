package spells

import (
	"image/color"
	"math"
	"math/rand"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// LightningStorm spawns repeated lightning strikes within a radius.
type LightningStorm struct {
	Info             SpellInfo
	CenterX, CenterY float64
	Radius           int
	TickRate         float64
	Duration         float64
	Elapsed          float64
	NextStrikeTime   float64
	Caster           *Caster
	ImpactImg        *ebiten.Image
	spawned          []*LightningStrike
	Finished         bool
	validStrikeTiles []struct{ X, Y int }
}

// NewLightningStorm creates a new storm centered on the given tile.
func NewLightningStorm(info SpellInfo, cx, cy float64, radius int, tickRate, duration float64, caster *Caster, impact *ebiten.Image, level *levels.Level) *LightningStorm {
	storm := &LightningStorm{
		Info:           info,
		CenterX:        cx,
		CenterY:        cy,
		Radius:         radius,
		TickRate:       tickRate,
		Duration:       duration,
		Caster:         caster,
		ImpactImg:      impact,
		NextStrikeTime: 0,
	}

	centerX := int(math.Floor(cx))
	centerY := int(math.Floor(cy))

	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if abs(dx)+abs(dy) <= radius {
				tx := centerX + dx
				ty := centerY + dy
				if level.IsWalkable(tx, ty) {
					storm.validStrikeTiles = append(storm.validStrikeTiles, struct{ X, Y int }{tx, ty})
				}
			}
		}
	}

	return storm
}

func (ls *LightningStorm) Update(level *levels.Level, dt float64) {
	if ls.Finished {
		return
	}

	ls.Elapsed += dt
	if ls.Elapsed >= ls.Duration {
		ls.Finished = true
		return
	}

	if ls.Elapsed >= ls.NextStrikeTime {
		if len(ls.validStrikeTiles) == 0 {
			return
		}

		// Pick a random tile in the AOE zone
		idx := rand.Intn(len(ls.validStrikeTiles))
		tile := ls.validStrikeTiles[idx]

		// Ensure the tile is walkable or visually allowed
		if level.IsWalkable(tile.X, tile.Y) {
			// Impact burst (visual)
			if ls.ImpactImg != nil {
				strike := NewLightningStrike(
					SpellInfo{
						Name:     "lightning",
						Level:    1,
						Cooldown: 0.01,
						Damage:   ls.Info.Damage,
					},
					float64(tile.X), float64(tile.Y), ls.ImpactImg,
				)
				ls.spawned = append(ls.spawned, strike)
			}
		}

		ls.NextStrikeTime += ls.TickRate
	}
}
func (ls *LightningStorm) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if ls.Finished {
		return
	}

	centerTileX := int(math.Floor(ls.CenterX))
	centerTileY := int(math.Floor(ls.CenterY))

	aoeColor := color.NRGBA{R: 80, G: 140, B: 255, A: 100}

	for dx := -ls.Radius; dx <= ls.Radius; dx++ {
		for dy := -ls.Radius; dy <= ls.Radius; dy++ {
			if abs(dx)+abs(dy) <= ls.Radius {
				// We use +2 and +1 to align cursor selection to tile centers.
				tx := centerTileX + dx + 2
				ty := centerTileY + dy + 1
				drawAOETile(screen, tx, ty, tileSize, camX, camY, camScale, cx, cy, aoeColor)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (ls *LightningStorm) IsFinished() bool { return ls.Finished }

// TakeSpawns returns any LightningStrikes created since the last call.
func (ls *LightningStorm) TakeSpawns() []*LightningStrike {
	out := ls.spawned
	ls.spawned = nil
	return out
}
func drawAOETile(screen *ebiten.Image, x, y int, tileSize int, camX, camY, camScale, cx, cy float64, col color.NRGBA) {
	if camScale < 0.3 {
		return // avoid blobs at far zooms
	}
	sx, sy := isoToScreenFloat(float64(x), float64(y), tileSize)
	sx = math.Round((sx-camX)*camScale + cx)
	sy = math.Round((sy+camY)*camScale + cy)

	halfW := float64(tileSize) / 2
	quarterH := float64(tileSize) / 4

	// Draw vertices
	vertices := []ebiten.Vertex{
		{DstX: float32(sx), DstY: float32(sy - quarterH), ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255, ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255},
		{DstX: float32(sx + halfW), DstY: float32(sy), ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255, ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255},
		{DstX: float32(sx), DstY: float32(sy + quarterH), ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255, ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255},
		{DstX: float32(sx - halfW), DstY: float32(sy), ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255, ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255},
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}

	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)

	screen.DrawTriangles(vertices, indices, img, nil)
}
