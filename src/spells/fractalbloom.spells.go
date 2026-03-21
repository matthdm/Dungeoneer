package spells

import (
	"math"
	"math/rand"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// FractalNode is a single explosion node within a FractalBloom
// spell. Each node is added to ActiveSpells and handles its own
// lifetime and damage application.
type FractalNode struct {
	Info          SpellInfo
	X, Y          float64
	Radius        int
	Damage        int
	ImpactImg     *ebiten.Image
	Duration      float64
	age           float64
	DamageApplied bool
	Finished      bool
	SpawnTime     float64 // relative time until this node spawns
}

// Update advances the explosion animation.
func (fn *FractalNode) Update(level *levels.Level, dt float64) {
	if fn.Finished {
		return
	}
	fn.age += dt
	if fn.age >= fn.Duration {
		fn.Finished = true
	}
}

// Draw renders the explosion at its tile location.
func (fn *FractalNode) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if fn.Finished {
		return
	}
	if fn.ImpactImg == nil {
		return
	}
	sx, sy := isoToScreenFloat(fn.X, fn.Y, tileSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx, sy)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	screen.DrawImage(fn.ImpactImg, op)
}

// IsFinished implements the Spell interface.
func (fn *FractalNode) IsFinished() bool { return fn.Finished }

// FractalBloom orchestrates a recursive series of explosion nodes.
type FractalBloom struct {
	Info       SpellInfo
	Caster     *Caster
	ImpactImg  *ebiten.Image
	spawnQueue []*FractalNode
	spawned    []*FractalNode
	elapsed    float64
	Finished   bool
}

// NewFractalBloom builds all explosion nodes upfront. The root
// explosion starts at (cx,cy) and recursively creates child nodes up to
// maxDepth. Damage and radius drop off by dropoff each depth.
func NewFractalBloom(info SpellInfo, cx, cy float64, caster *Caster, impact *ebiten.Image, level *levels.Level, maxDepth int, dropoff float64, delay float64) *FractalBloom {
	fb := &FractalBloom{
		Info:      info,
		Caster:    caster,
		ImpactImg: impact,
	}
	fb.buildNodes(cx, cy, info.Damage, 2, 0, maxDepth, dropoff, 0, delay, level)
	// sort by spawn time to ensure deterministic ordering
	for i := 0; i < len(fb.spawnQueue)-1; i++ {
		for j := i + 1; j < len(fb.spawnQueue); j++ {
			if fb.spawnQueue[j].SpawnTime < fb.spawnQueue[i].SpawnTime {
				fb.spawnQueue[i], fb.spawnQueue[j] = fb.spawnQueue[j], fb.spawnQueue[i]
			}
		}
	}
	return fb
}

func (fb *FractalBloom) buildNodes(x, y float64, dmg int, radius int, depth int, maxDepth int, dropoff float64, t float64, delay float64, level *levels.Level) {
	node := &FractalNode{
		Info:      fb.Info,
		X:         x,
		Y:         y,
		Radius:    radius,
		Damage:    dmg,
		Duration:  0.5,
		ImpactImg: fb.ImpactImg,
		SpawnTime: t,
	}
	fb.spawnQueue = append(fb.spawnQueue, node)
	if depth >= maxDepth {
		return
	}
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	childRadius := radius - 1
	childDamage := int(float64(dmg) * dropoff)
	for _, d := range dirs {
		if rand.Float64() > 0.7 {
			continue
		}
		tx := int(math.Round(x)) + d[0]
		ty := int(math.Round(y)) + d[1]
		if !level.IsWalkable(tx, ty) {
			continue
		}
		fb.buildNodes(float64(tx), float64(ty), childDamage, childRadius, depth+1, maxDepth, dropoff, t+delay, delay, level)
	}
}

// Update spawns nodes from the queue based on elapsed time.
func (fb *FractalBloom) Update(level *levels.Level, dt float64) {
	if fb.Finished {
		return
	}
	fb.elapsed += dt
	for len(fb.spawnQueue) > 0 {
		n := fb.spawnQueue[0]
		if fb.elapsed < n.SpawnTime {
			break
		}
		fb.spawnQueue = fb.spawnQueue[1:]
		fb.spawned = append(fb.spawned, n)
	}
	if len(fb.spawnQueue) == 0 {
		fb.Finished = true
	}
}

func (fb *FractalBloom) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
}

func (fb *FractalBloom) IsFinished() bool { return fb.Finished }

// TakeSpawns returns any FractalNodes spawned since last call.
func (fb *FractalBloom) TakeSpawns() []*FractalNode {
	out := fb.spawned
	fb.spawned = nil
	return out
}
