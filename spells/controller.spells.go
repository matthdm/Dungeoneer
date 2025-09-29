package spells

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"time"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// -----------------------------
// Core types
// -----------------------------

// AbilitySource tells us where a learned/equipped spell came from.
type AbilitySource int

const (
	SourceUnknown AbilitySource = iota
	SourceItem                  // granted by an equipped item
	SourceNative                // granted by level-up / class tree
)

// SlotCount: how many quick slots appear on the HUD.
const SlotCount = 5

// SpellDef describes a learnable spell type, how to cast it, and its UI.
type SpellDef struct {
	// Unique key (same as SpellInfo.Name)
	Name string

	// Base info (cooldown, base damage, etc.). Can be scaled by level if desired.
	Info SpellInfo

	// Icon displayed in the HUD spell bar. Provide a 32–48px square.
	Icon *ebiten.Image

	// Factory is responsible for creating a live Spell instance.
	// It is called only if cooldown is ready.
	Factory func(ctx CastContext) (Spell, error)

	// Optional: if true, spell is instant at target tile and doesn’t need start pos.
	// Mainly a hint for input UX — we won’t enforce anything here.
	Instant bool
}

// CastContext bundles everything a factory needs to create the Spell.
type CastContext struct {
	Level  *levels.Level
	Caster *Caster

	// World-space tile coordinates
	StartX, StartY   float64
	TargetX, TargetY float64

	// A copy of the (possibly level-scaled) info for this cast
	Info SpellInfo

	// Optional art/resources that the factory may need
	Icon *ebiten.Image
	// You can add more shared assets here (e.g. Fireball frames, impact sprites)
	Assets map[string]any
}

// Equipped spell slot entry.
type SpellSlot struct {
	Name   string        // refers to registry key
	Source AbilitySource // where it came from (item/native)
}

// Controller is the central orchestrator.
// - owns the spell registry
// - tracks what spells the player knows
// - manages equipped slots & cooldowns (via Caster)
// - updates/draws active spell instances
type Controller struct {
	registry map[string]*SpellDef

	// All spells granted to the player (key: SpellDef.Name)
	known map[string]AbilitySource

	// Ordered quick-slots rendered in HUD
	slots [SlotCount]SpellSlot

	// Cooldowns & timing are handled by the Caster you already wrote
	Caster *Caster

	// Active spell instances living in the world until IsFinished()
	Active []Spell

	// HUD config
	hudRect   image.Rectangle // where to draw the bar
	iconSize  int
	iconGap   int
	fontColor color.Color

	// For cooldown visuals
	now func() time.Time
}

// EffectResolver applies per-frame spell effects such as damage, healing, or
// other game-specific interactions. Game implements this interface so the
// controller can remain agnostic of monsters, HUD, etc.
type EffectResolver interface {
	ResolveSpellEffects(sp Spell, level *levels.Level)
}

// SpawnTaker is implemented by spells that can spawn child spell instances
// (e.g. lightning storms spawning strikes, fractal blooms spawning nodes).
type SpawnTaker interface {
	TakeSpawns() []Spell
}

// NewController creates a fresh controller.
func NewController() *Controller {
	return &Controller{
		registry:  make(map[string]*SpellDef),
		known:     make(map[string]AbilitySource),
		Caster:    NewCaster(), // uses your existing cooldown map :contentReference[oaicite:2]{index=2}
		iconSize:  40,
		iconGap:   8,
		fontColor: color.White,
		now:       time.Now,
	}
}

// -----------------------------
// Registration & unlock APIs
// -----------------------------

// Register adds or replaces a spell type definition.
// Call this at startup for all spell types.
func (sc *Controller) Register(def *SpellDef) error {
	if def == nil || def.Name == "" || def.Factory == nil {
		return errors.New("invalid SpellDef")
	}
	sc.registry[def.Name] = def
	return nil
}

// Grant a spell to the player from an item (on equip).
func (sc *Controller) GrantFromItem(spellName string) {
	if _, ok := sc.registry[spellName]; !ok {
		return
	}
	sc.known[spellName] = SourceItem
}

// Revoke spell granted by an item (on unequip).
func (sc *Controller) RevokeFromItem(spellName string) {
	src, ok := sc.known[spellName]
	if !ok {
		return
	}
	if src == SourceItem {
		delete(sc.known, spellName)
		if sc.Caster != nil {
			delete(sc.Caster.Cooldowns, spellName)
		}
		// remove from any slot
		for i := range sc.slots {
			if sc.slots[i].Name == spellName {
				sc.slots[i] = SpellSlot{}
			}
		}
	}
}

// Grant a native ability (learned via level-up).
func (sc *Controller) GrantNative(spellName string) {
	if _, ok := sc.registry[spellName]; !ok {
		return
	}
	sc.known[spellName] = SourceNative
}

// Known spells, sorted for UI lists.
func (sc *Controller) Known() []string {
	ks := make([]string, 0, len(sc.known))
	for k := range sc.known {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// -----------------------------
// Loadout / slot management
// -----------------------------

func (sc *Controller) Equip(slot int, spellName string) error {
	if slot < 0 || slot >= len(sc.slots) {
		return fmt.Errorf("invalid slot %d", slot)
	}
	if _, ok := sc.known[spellName]; !ok {
		return fmt.Errorf("spell %q not known", spellName)
	}
	sc.slots[slot] = SpellSlot{Name: spellName, Source: sc.known[spellName]}
	return nil
}

func (sc *Controller) Unequip(slot int) {
	if slot < 0 || slot >= len(sc.slots) {
		return
	}
	sc.slots[slot] = SpellSlot{}
}

// -----------------------------
// Update / TryCast / Draw
// -----------------------------

// Update should be called once per frame.
// - ticks global cooldowns on Caster (you already implemented) :contentReference[oaicite:3]{index=3}
// - updates active spell instances and prunes finished ones
func (sc *Controller) Update(level *levels.Level, dt float64, resolver EffectResolver) {
	// tick cooldowns
	sc.Caster.Update(dt) // your Ready/PutOnCooldown logic uses this map :contentReference[oaicite:4]{index=4}

	// update active instances
	dst := sc.Active[:0]
	var additions []Spell
	for _, s := range sc.Active {
		s.Update(level, dt)
		if resolver != nil {
			resolver.ResolveSpellEffects(s, level)
		}
		if spawner, ok := s.(SpawnTaker); ok {
			children := spawner.TakeSpawns()
			if len(children) > 0 {
				additions = append(additions, children...)
			}
		}
		if !s.IsFinished() {
			dst = append(dst, s)
		}
	}
	if len(additions) > 0 {
		dst = append(dst, additions...)
	}
	sc.Active = dst
}

// TryCast attempts to cast the spell in the given slot.
// It returns true if the cast succeeded (cooldown consumed & spell spawned).
func (sc *Controller) TryCast(slot int, ctx CastContext) bool {
	if slot < 0 || slot >= len(sc.slots) {
		return false
	}
	slotEntry := sc.slots[slot]
	if slotEntry.Name == "" {
		return false
	}
	def, ok := sc.registry[slotEntry.Name]
	if !ok {
		return false
	}
	// Respect cooldowns using your Caster
	if !sc.Caster.Ready(def.Info) { // cd <= 0 means ready :contentReference[oaicite:5]{index=5}
		return false
	}
	// Provide spell-specific info to the factory
	ctx.Info = def.Info
	ctx.Icon = def.Icon
	if ctx.Caster == nil {
		ctx.Caster = sc.Caster
	}

	inst, err := def.Factory(ctx)
	if err != nil || inst == nil {
		return false
	}

	// Start cooldown immediately
	sc.Caster.PutOnCooldown(def.Info) // set cd = Info.Cooldown :contentReference[oaicite:6]{index=6}

	// Track active instance
	sc.Active = append(sc.Active, inst)
	return true
}

// DrawWorld draws all active spells in world space.
// (Each spell has its own Draw consistent with your implementations.)
func (sc *Controller) DrawWorld(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	for _, s := range sc.Active {
		s.Draw(screen, tileSize, camX, camY, camScale, cx, cy)
	}
}

// ConfigureHUD sets placement and style of the bar.
func (sc *Controller) ConfigureHUD(rect image.Rectangle, iconSize int, gap int) {
	sc.hudRect = rect
	if iconSize > 0 {
		sc.iconSize = iconSize
	}
	if gap > 0 {
		sc.iconGap = gap
	}
}

// DrawHUD renders equipped slots with icons and a cooldown overlay.
// Simple, legible, and fast. Uses a radial wedge for remaining cooldown.
func (sc *Controller) DrawHUD(screen *ebiten.Image) {
	if sc.hudRect.Empty() {
		// Default: bottom-center bar 4 icons
		sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
		totalW := SlotCount*sc.iconSize + (SlotCount-1)*sc.iconGap
		x := (sw - totalW) / 2
		y := sh - sc.iconSize - 12
		sc.hudRect = image.Rect(x, y, x+totalW, y+sc.iconSize)
	}

	x := sc.hudRect.Min.X
	y := sc.hudRect.Min.Y

	for i := 0; i < SlotCount; i++ {
		r := image.Rect(x, y, x+sc.iconSize, y+sc.iconSize)
		sc.drawSlot(screen, r, i)
		x += sc.iconSize + sc.iconGap
	}
}

func (sc *Controller) drawSlot(screen *ebiten.Image, r image.Rectangle, slot int) {
	// background box
	bg := ebiten.NewImage(r.Dx(), r.Dy())
	bg.Fill(color.NRGBA{R: 20, G: 20, B: 28, A: 200})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(bg, op)

	entry := sc.slots[slot]
	var def *SpellDef
	if entry.Name != "" {
		def = sc.registry[entry.Name]
	}

	// icon (dim if unknown)
	if def != nil && def.Icon != nil {
		op := &ebiten.DrawImageOptions{}
		// center-crop if needed (assume icon already square)
		op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
		screen.DrawImage(def.Icon, op)
	} else {
		// placeholder frame
		b := float32(2)
		vector.StrokeRect(screen, float32(r.Min.X)+b, float32(r.Min.Y)+b, float32(r.Dx())-2*b, float32(r.Dy())-2*b, b, color.NRGBA{R: 80, G: 80, B: 90, A: 220}, false)
	}

	// cooldown overlay (radial wedge)
	if def != nil {
		cd := sc.Caster.Cooldowns[def.Name] // 0..Info.Cooldown
		if cd > 0 && def.Info.Cooldown > 0 {
			frac := float32(cd / def.Info.Cooldown) // 1.0 -> full overlay
			sc.drawCooldownPie(screen, r, frac)
		}
	}

	// slot index label (tiny white dot/line on bottom-right for quick read)
	vector.StrokeRect(screen,
		float32(r.Max.X-8), float32(r.Max.Y-8),
		6, 6, 1.4, color.NRGBA{R: 220, G: 220, B: 230, A: 220}, false)
}

// Very small triangle-fan “pie” for cooldown. Single color, no allocations.
func (sc *Controller) drawCooldownPie(screen *ebiten.Image, r image.Rectangle, frac float32) {
	if frac <= 0 {
		return
	}
	cx := float32(r.Min.X + r.Dx()/2)
	cy := float32(r.Min.Y + r.Dy()/2)
	rad := float32(r.Dx()) * 0.5

	// angle from top, clockwise
	theta := float32(2*math.Pi) * frac

	const steps = 24
	verts := make([]ebiten.Vertex, 0, steps+2)
	idx := make([]uint16, 0, steps*3)

	col := color.NRGBA{R: 10, G: 10, B: 10, A: 180}
	verts = append(verts, ebiten.Vertex{
		DstX: cx, DstY: cy,
		ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255, ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255,
	})
	segments := int(float32(steps) * frac)
	if segments < 1 {
		segments = 1
	}
	for i := 0; i <= segments; i++ {
		a := -float32(math.Pi/2) + theta*float32(i)/float32(segments)
		x := cx + rad*float32(math.Cos(float64(a)))
		y := cy + rad*float32(math.Sin(float64(a)))
		verts = append(verts, ebiten.Vertex{
			DstX: x, DstY: y,
			ColorR: float32(col.R) / 255, ColorG: float32(col.G) / 255, ColorB: float32(col.B) / 255, ColorA: float32(col.A) / 255,
		})
	}
	for i := 1; i < len(verts)-1; i++ {
		idx = append(idx, 0, uint16(i), uint16(i+1))
	}
	px := ebiten.NewImage(1, 1)
	px.Fill(color.White)
	screen.DrawTriangles(verts, idx, px, nil)
}

// SlotsForHUD returns parallel arrays of icons and cooldowns for the HUD to paint.
func (sc *Controller) SlotsForHUD() (icons [SlotCount]*ebiten.Image, remaining [SlotCount]float64, totals [SlotCount]float64, infos [SlotCount]SpellInfo) {
	for i, s := range sc.slots {
		if s.Name == "" {
			continue
		}
		def, ok := sc.registry[s.Name]
		if !ok {
			continue
		}
		icons[i] = def.Icon
		infos[i] = def.Info
		totals[i] = def.Info.Cooldown
		if sc.Caster != nil {
			remaining[i] = sc.Caster.Cooldowns[def.Name]
		}
	}
	return
}

// Def returns the SpellDef by name (or nil).
func (sc *Controller) Def(name string) *SpellDef { return sc.registry[name] }

// Slot returns the spell name in a given slot ("" if empty).
func (sc *Controller) Slot(i int) string {
	if i < 0 || i >= len(sc.slots) {
		return ""
	}
	return sc.slots[i].Name
}

// SlotIndex returns the first quick-slot index containing the given spell.
func (sc *Controller) SlotIndex(name string) int {
	for i, slot := range sc.slots {
		if slot.Name == name {
			return i
		}
	}
	return -1
}

// HasEquipped reports whether a spell is equipped in the specified slot.
func (sc *Controller) HasEquipped(slot int) bool {
	if slot < 0 || slot >= len(sc.slots) {
		return false
	}
	return sc.slots[slot].Name != ""
}

// ClearItemSpells removes all item-granted spells and clears their slots.
func (sc *Controller) ClearItemSpells() {
	names := make([]string, 0, len(sc.known))
	for name, src := range sc.known {
		if src == SourceItem {
			names = append(names, name)
		}
	}
	for _, name := range names {
		sc.RevokeFromItem(name)
	}
}
