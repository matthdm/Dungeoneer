# Void Rift VFX Pilot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship two new CR1 void-domain artifacts — `void_rift_catalyst` (active skill) and `voidbound_pendant` (passive "enchantment") — rendered with a new Kage-shader VFX pipeline (`src/spells/vfx_shaders.go`), proving out shader-based glow/rift effects as the reference model for future spell visuals, equippable in-run via a new F2 dev-menu entry for screenshot-based iteration.

**Architecture:** Two compiled-once Kage shaders (radial glow, procedural collapsing-rift) live behind small Go draw helpers in `src/spells/vfx_shaders.go`. `VoidRift` implements the existing `Spell` interface for the triggered rift effect (spawned via `spawnSkillVisual`, same pattern as the 7 existing CR1-F bespoke artifact visuals). `VoidboundAura` is deliberately *not* a `Spell` — it's a persistent per-frame draw gated on artifact equip state, mirroring the existing `BurnActive` continuous-effect pattern. Both artifacts register through the existing `ItemTemplate` + `combat.ArtifactEffects` dual-registration convention; `void_rift_catalyst`'s damage+root is expressed entirely with existing engine fields (`DamageMultiplier`, `IsRoot`, `DurationSec`), so **zero `src/combat/engine.go` changes are required** — this resolves the spec's Open Question #1.

**Tech Stack:** Go, Ebitengine v2.8, Kage shader language (`ebiten.NewShader`, `DrawTrianglesShader`).

## Global Constraints

- `go build ./...` must stay green after every task.
- `go test ./combat/...` must stay green after every task that touches `src/combat/`.
- Shaders are compiled **once**, at package load (`init()` / package-level `var`), never per-frame or per-draw-call — no-per-frame-allocation-in-hot-paths standard (`CLAUDE.md`).
- Fireball and all other legacy spells (`lightningstorm.spells.go`, `chaosray.spells.go`, `fractalbloom.spells.go`, `fractalcanopy.spells.go`, `arcanebolt.spells.go`) are **not modified**.
- The known per-frame allocation bug in `drawAOETile` (`src/spells/lightningstorm.spells.go:159`) is **not fixed** here — unrelated file, out of scope.
- No `src/combat/engine.go` mechanic changes — `void_rift_catalyst` is deliberately designed to fit existing `ArtifactEffect` fields (see Task 2).
- No commits are made by the implementing agent — the user (Matthew Morales) commits everything, per `CLAUDE.md`.
- **Technical correction from the spec:** the spec described the rift as "screen-space UV distortion pulling nearby pixels inward" (i.e. warping the actual dungeon floor behind it). True background distortion requires capturing the current frame to an offscreen buffer before drawing the effect — Ebitengine cannot sample the same image it's currently drawing to. That's new rendering-pipeline plumbing this pilot deliberately avoids. Instead, the rift shader **procedurally synthesizes** a swirling core + rim pattern (self-contained, no background sampling) — same "collapsing rift with gravitational pull" read, without the architectural cost. This is called out explicitly rather than silently simplified.

---

### Task 1: Shared Kage shader primitives

**Files:**
- Create: `src/spells/vfx_shaders.go`

**Interfaces:**
- Produces: `drawGlow(screen *ebiten.Image, sx, sy, radius float32, col [3]float32, intensity float32, t float64)` — additive radial glow centered at screen-pixel `(sx, sy)`.
- Produces: `drawRift(screen *ebiten.Image, sx, sy, radius float32, coreCol, rimCol [3]float32, t float64)` — procedural collapsing-rift quad (dark swirling core + bright rim), normal alpha blend.
- Produces (internal): `drawShaderQuad(screen *ebiten.Image, shader *ebiten.Shader, sx, sy, halfSize float32, uniforms map[string]interface{}, blend ebiten.Blend)` — shared quad-builder both wrappers use.

- [ ] **Step 1: Write the shader sources and draw helpers**

```go
package spells

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// This file holds the shared Kage-shader VFX primitives new spell/artifact
// visuals build on: a radial glow (additive light/energy) and a procedural
// collapsing-rift pattern (swirling core + bright rim). Both are compiled
// once at package load — never per-frame — and driven by small Go wrappers
// that follow the same screen-space-quad convention as every other spell
// Draw method in this package.

// glowShaderSrc renders a soft additive radial glow: bright core fading to
// transparent at Radius, with a slow pulse driven by Time.
const glowShaderSrc = `//kage:unit pixels
package main

var Center vec2
var Radius float
var Time float
var Col vec3
var Intensity float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	dist := distance(dstPos.xy, Center)
	pulse := 0.85 + 0.15*sin(Time*1.6)
	falloff := 1.0 - smoothstep(0.0, Radius*pulse, dist)
	falloff *= falloff
	a := clamp(falloff*Intensity, 0.0, 1.0)
	return vec4(Col*a, a)
}
`

// riftShaderSrc procedurally synthesizes a collapsing-rift pattern: a dark
// swirling core (angle-based sine swirl, animated by Time) inside Radius*0.55,
// with a bright rim band between Radius*0.55 and Radius*0.85. Self-contained —
// it does not sample the background, so it needs no offscreen render target.
const riftShaderSrc = `//kage:unit pixels
package main

var Center vec2
var Radius float
var Time float
var CoreCol vec3
var RimCol vec3

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	offset := dstPos.xy - Center
	dist := length(offset)
	angle := atan(offset.y, offset.x)
	mask := 1.0 - smoothstep(Radius*0.9, Radius, dist)

	swirl := sin(angle*6.0+Time*4.0-dist*0.3)*0.5 + 0.5
	core := 1.0 - smoothstep(0.0, Radius*0.55, dist)
	core = pow(core, 1.6)

	rimOuter := 1.0 - smoothstep(Radius*0.7, Radius*0.85, dist)
	rimInner := 1.0 - smoothstep(Radius*0.55, Radius*0.7, dist)
	rim := clamp(rimOuter-rimInner, 0.0, 1.0)

	col := CoreCol*core*(0.6+0.4*swirl) + RimCol*rim
	a := clamp(core+rim, 0.0, 1.0) * mask
	return vec4(col*a, a)
}
`

var (
	glowShader *ebiten.Shader
	riftShader *ebiten.Shader
)

func init() {
	var err error
	glowShader, err = ebiten.NewShader([]byte(glowShaderSrc))
	if err != nil {
		panic("spells: failed to compile glowShaderSrc: " + err.Error())
	}
	riftShader, err = ebiten.NewShader([]byte(riftShaderSrc))
	if err != nil {
		panic("spells: failed to compile riftShaderSrc: " + err.Error())
	}
}

// drawShaderQuad draws a single shader-filled quad covering a square of side
// 2*halfSize centered at (sx,sy) in screen pixels. Vertex colors are left at
// white/opaque — the shaders compute their own color and alpha, ignoring the
// vertex color input.
func drawShaderQuad(screen *ebiten.Image, shader *ebiten.Shader, sx, sy, halfSize float32, uniforms map[string]interface{}, blend ebiten.Blend) {
	x0, y0 := sx-halfSize, sy-halfSize
	x1, y1 := sx+halfSize, sy+halfSize
	vs := []ebiten.Vertex{
		{DstX: x0, DstY: y0, SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x1, DstY: y0, SrcX: 1, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x1, DstY: y1, SrcX: 1, SrcY: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: x0, DstY: y1, SrcX: 0, SrcY: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	is := []uint16{0, 1, 2, 0, 2, 3}
	op := &ebiten.DrawTrianglesShaderOptions{Uniforms: uniforms, Blend: blend}
	screen.DrawTrianglesShader(vs, is, shader, op)
}

// drawGlow renders an additive radial glow of the given screen-pixel radius
// centered at (sx,sy). col is linear RGB in [0,1]; intensity scales opacity.
func drawGlow(screen *ebiten.Image, sx, sy, radius float32, col [3]float32, intensity float32, t float64) {
	drawShaderQuad(screen, glowShader, sx, sy, radius*1.3, map[string]interface{}{
		"Center":    []float32{sx, sy},
		"Radius":    radius,
		"Time":      float32(t),
		"Col":       col[:],
		"Intensity": intensity,
	}, ebiten.BlendLighter)
}

// drawRift renders the procedural collapsing-rift pattern (see riftShaderSrc)
// of the given screen-pixel radius centered at (sx,sy).
func drawRift(screen *ebiten.Image, sx, sy, radius float32, coreCol, rimCol [3]float32, t float64) {
	drawShaderQuad(screen, riftShader, sx, sy, radius*1.1, map[string]interface{}{
		"Center":  []float32{sx, sy},
		"Radius":  radius,
		"Time":    float32(t),
		"CoreCol": coreCol[:],
		"RimCol":  rimCol[:],
	}, ebiten.BlendSourceOver)
}
```

- [ ] **Step 2: Verify it compiles and the shaders parse**

Run: `cd src && go build ./...`
Expected: no output (clean build). If either `ebiten.NewShader` call panics at program init due to a Kage syntax error, the build itself won't catch it (compilation happens at runtime) — instead run:

Run: `cd src && go vet ./spells/...`
Expected: no output. Then do a smoke run via the `run` skill (Task 7 covers full manual verification) — if the Kage source has a syntax error, the game will panic immediately on launch with the shader compile error message, which is the fastest signal.

- [ ] **Step 3: No unit test for this task**

This file is pure rendering (shader source + draw calls) — there is no meaningful assertion beyond "it compiles and doesn't panic," which Step 2 already covers. Visual correctness is verified in Task 7 via screenshot review, consistent with the spec's Verification section (no automated visual-quality test exists or is proposed).

---

### Task 2: Register the two artifacts (mechanics only, no visuals yet)

**Files:**
- Modify: `src/combat/skills.go` (add two `ArtifactEffects` map entries)
- Modify: `src/items/load.go` (add two `RegisterItem` calls inside `registerNewArtifacts()`)
- Create: `src/combat/voidrift_test.go`

**Interfaces:**
- Produces: `combat.ArtifactEffects["void_rift_catalyst"]`, `combat.ArtifactEffects["voidbound_pendant"]` — consumed by Task 3 (visual spawn), Task 5 (`spawnSkillVisual` case), Task 6 (icon/flavor lookup by item ID).
- Produces: item IDs `"void_rift_catalyst"` (ability ID `"void_rift_blast"`) and `"voidbound_pendant"` (ability ID `"voidbound_ward"`) in `items.Registry`.

- [ ] **Step 1: Write the failing engine tests**

Create `src/combat/voidrift_test.go`:

```go
package combat

import "testing"

// TestVoidRiftCatalystRootsAndDamagesTarget verifies void_rift_catalyst fires,
// deals DamageMultiplier-scaled damage, and applies a root — mirrors the
// existing ironbreaker_gauntlets/ashbound_chain engine test coverage.
func TestVoidRiftCatalystRootsAndDamagesTarget(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.EquippedArtifacts[0] = "void_rift_catalyst"
	state.ArtifactCooldowns[0] = 0

	action := Action{Type: ActionActivateSkill, SlotIdx: 0}
	newState, events := eng.Tick(state, []Action{action})

	fired := false
	dealt := false
	for _, ev := range events {
		if ev.Type == EventSkillFired && ev.Tag == "void_rift_catalyst" {
			fired = true
		}
		if ev.Type == EventDamageDealt && ev.Tag == "void_rift_catalyst" && ev.Value > 0 {
			dealt = true
		}
	}
	if !fired {
		t.Fatal("expected EventSkillFired for void_rift_catalyst")
	}
	if !dealt {
		t.Fatal("expected EventDamageDealt with positive value for void_rift_catalyst")
	}
	if !newState.TargetRooted {
		t.Fatal("expected TargetRooted=true after void_rift_catalyst fires")
	}
	if newState.RootTimer <= 0 || newState.RootTimer > 1.5 {
		t.Fatalf("expected RootTimer in (0, 1.5], got %f", newState.RootTimer)
	}
}

// TestVoidRiftCatalystRequiresTargetInRange checks the melee-reach gate
// (shared by every DamageMultiplier>0 skill, e.g. ironbreaker_gauntlets)
// also applies to void_rift_catalyst: out of range, no fire.
func TestVoidRiftCatalystRequiresTargetInRange(t *testing.T) {
	eng := newEngine()
	state := baseState()
	state.TargetInRange = false
	state.EquippedArtifacts[0] = "void_rift_catalyst"
	state.ArtifactCooldowns[0] = 0

	action := Action{Type: ActionActivateSkill, SlotIdx: 0}
	_, events := eng.Tick(state, []Action{action})

	for _, ev := range events {
		if ev.Type == EventSkillFired {
			t.Fatal("expected void_rift_catalyst to NOT fire when target is out of range")
		}
	}
}

// TestVoidboundPendantIsRegisteredAsPassive guards against registration
// typos: voidbound_pendant must be a zero-cooldown passive with its
// skill-duration bonus set, so aggregatePassive (new_adapter.go) picks it up.
func TestVoidboundPendantIsRegisteredAsPassive(t *testing.T) {
	eff, ok := ArtifactEffects["voidbound_pendant"]
	if !ok {
		t.Fatal("voidbound_pendant not registered in ArtifactEffects")
	}
	if !eff.IsPassive {
		t.Fatal("expected voidbound_pendant.IsPassive = true")
	}
	if eff.Cooldown != 0 {
		t.Fatalf("expected voidbound_pendant.Cooldown = 0, got %f", eff.Cooldown)
	}
	if eff.SkillDurationPct <= 0 {
		t.Fatalf("expected voidbound_pendant.SkillDurationPct > 0, got %d", eff.SkillDurationPct)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd src && go test ./combat/... -run TestVoidRift -v` and `go test ./combat/... -run TestVoidboundPendant -v`
Expected: FAIL — `ArtifactEffects["void_rift_catalyst"]` / `["voidbound_pendant"]` don't exist yet (test dereferences a zero-value `ArtifactEffect{}`, so `TestVoidboundPendantIsRegisteredAsPassive` fails on the `!ok` check; the other two fail because no `EventSkillFired`/`EventDamageDealt` is emitted for an unregistered artifact ID).

- [ ] **Step 3: Add the ArtifactEffects entries**

In `src/combat/skills.go`, add to the `var ArtifactEffects = map[string]ArtifactEffect{...}` map literal (anywhere inside the map body — by convention, group with the other void-domain entries near `ashbound_chain`/`grave_reaper`):

```go
	"void_rift_catalyst": {
		Cooldown:         9.0,
		Domain:           "void",
		DamageMultiplier: 1.4,
		IsRoot:           true,
		DurationSec:      1.5,
	},
	"voidbound_pendant": {
		Cooldown:         0,
		Domain:           "void",
		IsPassive:        true,
		SkillDurationPct: 15,
	},
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd src && go test ./combat/... -run "TestVoidRift|TestVoidboundPendant" -v`
Expected: `PASS` for all three tests.

- [ ] **Step 5: Register both items in `src/items/load.go`**

Inside `registerNewArtifacts()` (`src/items/load.go`), add two `RegisterItem` calls — place them after the `ember_mantle` block (end of Wave 1) to keep the void pair together:

```go
	RegisterItem(&ItemTemplate{
		ID:             "void_rift_catalyst",
		Name:           "Void Rift Catalyst",
		Type:           ItemWeapon,
		Description:    "Tear open a collapsing rift at your target: deal damage and root them for 1.5s. +6 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 6},
		GrantsAbility:  "void_rift_blast",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "void",
	})
	RegisterItem(&ItemTemplate{
		ID:             "voidbound_pendant",
		Name:           "Voidbound Pendant",
		Type:           ItemArmor,
		Description:    "Passive: a warping void aura extends all your skill durations by 15%. +6 Max Mana.",
		Stackable:      false,
		MaxStack:       1,
		Equippable:     true,
		Stats:          map[string]int{"MaxMana": 6},
		GrantsAbility:  "voidbound_ward",
		AbilitySlot:    AbilitySlotSpell,
		Quality:        RarityRare,
		IsArtifact:     true,
		ArtifactDomain: "void",
	})
```

- [ ] **Step 6: Build and run the full combat test suite**

Run: `cd src && go build ./... && go test ./combat/... ./items/... -v`
Expected: build clean; all tests `PASS`, including the pre-existing suite (no regressions).

---

### Task 3: Void Rift active-skill visual

**Files:**
- Create: `src/spells/voidrift.spells.go`

**Interfaces:**
- Consumes: `drawGlow`, `drawRift` (Task 1); `isoToScreenFloat` (existing package-level helper already used throughout `src/spells`); `easeOutQuad` (existing helper in `src/spells/artifacts.spells.go`).
- Produces: `spells.NewVoidRift(x, y float64) *VoidRift`, implementing the `Spell` interface (`Update`, `Draw`, `IsFinished`) — consumed by Task 5's `spawnSkillVisual` case.

- [ ] **Step 1: Write the visual**

```go
package spells

import (
	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
)

// VoidRift is void_rift_catalyst's active-skill visual: a shader-driven
// collapsing rift at the target's position — it expands over the first 40%
// of its lifetime, then collapses, "opens, pulls, snaps shut." Damage and
// root are already applied by the combat engine before this is spawned (see
// EventSkillFired handling in game/new_adapter.go) — this is purely visual.
type VoidRift struct {
	X, Y     float64
	age      float64
	Duration float64
	Finished bool
}

// NewVoidRift creates a rift centered at the given world (cartesian) position,
// which should already be the target's body center (callers pass lockX/lockY
// per the existing artifact-skill convention in spawnSkillVisual).
func NewVoidRift(x, y float64) *VoidRift {
	return &VoidRift{X: x, Y: y, Duration: 0.6}
}

func (v *VoidRift) Update(level *levels.Level, dt float64) {
	if v.Finished {
		return
	}
	v.age += dt
	if v.age >= v.Duration {
		v.Finished = true
	}
}

func (v *VoidRift) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if v.Finished {
		return
	}
	progress := v.age / v.Duration
	var sizeT float64
	if progress < 0.4 {
		sizeT = easeOutQuad(progress / 0.4)
	} else {
		sizeT = 1.0 - easeOutQuad((progress-0.4)/0.6)
	}
	radius := float32(6+22*sizeT) * float32(camScale)

	sx, sy := isoToScreenFloat(v.X, v.Y, tileSize)
	ssx := float32((sx-camX)*camScale + cx)
	ssy := float32((sy+camY)*camScale + cy)

	coreCol := [3]float32{0.35, 0.08, 0.5}
	rimCol := [3]float32{0.75, 0.55, 1.0}
	drawRift(screen, ssx, ssy, radius, coreCol, rimCol, v.age)

	glowA := float32(1)
	if progress > 0.7 {
		glowA = float32(1 - (progress-0.7)/0.3)
	}
	drawGlow(screen, ssx, ssy, radius*0.5, [3]float32{0.55, 0.25, 0.85}, 0.8*glowA, v.age)
}

func (v *VoidRift) IsFinished() bool { return v.Finished }
```

- [ ] **Step 2: Build**

Run: `cd src && go build ./...`
Expected: clean build. No unit test here for the same reason as Task 1 — this is a rendering-only file (age/duration lifecycle logic is trivial and structurally identical to the already-unverified-by-test `GroundSlam`/`RootBind` in `artifacts.spells.go`, which the codebase doesn't test either). Verified visually in Task 7.

---

### Task 4: Voidbound Aura passive visual

**Files:**
- Create: `src/spells/voidboundaura.spells.go`

**Interfaces:**
- Consumes: `drawGlow` (Task 1), `isoToScreenFloat`.
- Produces: `spells.VoidboundAura` struct with `Update(dt float64)` and `Draw(screen *ebiten.Image, tileSize int, worldX, worldY float64, camX, camY, camScale, cx, cy float64)`. **Deliberately does not implement the `Spell` interface** — it has no finite lifetime and is never appended to `ActiveSpells`. Consumed by Task 5's `Game.VoidboundAura` field and `drawVoidboundAura` method.

- [ ] **Step 1: Write the aura**

```go
package spells

import "github.com/hajimehoshi/ebiten/v2"

// VoidboundAura is voidbound_pendant's passive visual: a slow-pulsing violet
// glow clinging to the caster while the pendant is active. Unlike a Spell
// (spawned once, expires, appended to ActiveSpells), this is drawn every
// frame the pendant is equipped — the caller owns gating and lifetime, not
// this struct. age accumulates continuously to drive the pulse animation.
type VoidboundAura struct {
	age float64
}

// Update advances the aura's animation clock. Safe to call every frame
// regardless of whether the pendant is currently equipped — Draw is what's
// actually gated by the caller.
func (a *VoidboundAura) Update(dt float64) {
	a.age += dt
}

// Draw renders the aura centered at (worldX, worldY), which the caller must
// already have body-center-offset (e.g. via Player.BodyX()/BodyY(), which
// applies coords.WorldPos.BodyCenter() internally) — never raw tile
// coordinates, per the project's coordinate invariant.
func (a *VoidboundAura) Draw(screen *ebiten.Image, tileSize int, worldX, worldY float64, camX, camY, camScale, cx, cy float64) {
	sx, sy := isoToScreenFloat(worldX, worldY, tileSize)
	ssx := float32((sx-camX)*camScale + cx)
	ssy := float32((sy+camY)*camScale + cy)
	radius := float32(10) * float32(camScale)
	drawGlow(screen, ssx, ssy, radius, [3]float32{0.5, 0.25, 0.85}, 0.55, a.age)
}
```

- [ ] **Step 2: Build**

Run: `cd src && go build ./...`
Expected: clean build.

---

### Task 5: Wire both visuals into the game layer

**Files:**
- Modify: `src/game/game.go` (add `VoidboundAura` field to `Game` struct)
- Modify: `src/game/spell_visuals.go` (add `spawnSkillVisual` case for `void_rift_blast`)
- Modify: `src/game/new_adapter.go` (add `hasArtifactEquipped` helper)
- Modify: `src/game/draw.game.go` (add `drawVoidboundAura` method + call site)
- Modify: `src/game/spells.game.go` (advance the aura's animation clock in `updateSpells()`)

**Interfaces:**
- Consumes: `spells.NewVoidRift` (Task 3), `spells.VoidboundAura` (Task 4), `combat.ArtifactEffects` entries (Task 2).
- Produces: `Game.hasArtifactEquipped(id string) bool`, `Game.drawVoidboundAura(screen *ebiten.Image, scale, cx, cy float64)` — internal wiring, not consumed elsewhere in this plan.

- [ ] **Step 1: Add the `VoidboundAura` field to `Game`**

In `src/game/game.go`, find the line:
```go
	ActiveSpells      []spells.Spell
```
and add immediately after it:
```go
	VoidboundAura     *spells.VoidboundAura // lazily created in drawVoidboundAura; nil until first drawn
```

- [ ] **Step 2: Add the `spawnSkillVisual` case**

In `src/game/spell_visuals.go`, inside the `switch id {` block of `spawnSkillVisual` (around line 137, right after the `case "blood_price_strike":` line and before `default:`), add:

```go
	case "void_rift_blast":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewVoidRift(lockX, lockY))
```

This follows the exact convention of the other artifact-skill cases (`void_bind`, `execute`, `arcane_surge_blast`) — it acts on the locked target (`lockX, lockY`), which `handleSkillFired` in `new_adapter.go` already resolves to the target's body center before calling `spawnSkillVisual`.

- [ ] **Step 3: Add the `hasArtifactEquipped` helper**

In `src/game/new_adapter.go`, immediately after the closing brace of `buildEquipped` (after the line `}` that ends the function containing `a.combatState.PassiveArtifacts = append(...)`), add:

```go
// hasArtifactEquipped reports whether the given canonical artifact ID is
// currently active on the player — either occupying a spell-bar slot
// (EquippedArtifacts, which is how dev-menu scenario loads place even
// passive items — see devLoadBuild in combat_debug.go) or worn as
// passive-only equipment (PassiveArtifacts). Used by persistent visual
// effects that aren't tied to a single triggered EventSkillFired, such as
// the Voidbound Aura.
func (g *Game) hasArtifactEquipped(id string) bool {
	na, ok := g.CombatAdapt.(*NewCombatAdapter)
	if !ok {
		return false
	}
	for _, eq := range na.combatState.EquippedArtifacts {
		if eq == id {
			return true
		}
	}
	for _, eq := range na.combatState.PassiveArtifacts {
		if eq == id {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Add `drawVoidboundAura` and call it from `drawPlaying`**

In `src/game/draw.game.go`, add this method near `drawSpells` (after its closing brace, around line 220):

```go
// drawVoidboundAura renders the Voidbound Aura while voidbound_pendant is
// active. The aura is not a Spell and is never appended to ActiveSpells — it
// persists for as long as the pendant is equipped, gated here each frame.
func (g *Game) drawVoidboundAura(screen *ebiten.Image, scale, cx, cy float64) {
	if g.player == nil || g.currentLevel == nil {
		return
	}
	if !g.hasArtifactEquipped("voidbound_pendant") {
		return
	}
	if g.VoidboundAura == nil {
		g.VoidboundAura = &spells.VoidboundAura{}
	}
	g.VoidboundAura.Draw(screen, g.currentLevel.TileSize, g.player.BodyX(), g.player.BodyY(), g.camX, g.camY, scale, cx, cy)
}
```

Then in `drawPlaying` (same file), find:
```go
	g.drawSpells(target, scale, cx, cy)
```
and add immediately after it:
```go
	g.drawVoidboundAura(target, scale, cx, cy)
```

- [ ] **Step 5: Advance the aura's animation clock**

In `src/game/spells.game.go`, at the top of `updateSpells()` (before `var remaining []spells.Spell`), add:

```go
	if g.VoidboundAura != nil {
		g.VoidboundAura.Update(g.DeltaTime)
	}
```

- [ ] **Step 6: Build**

Run: `cd src && go build ./...`
Expected: clean build.

---

### Task 6: Icons and flavor text

**Files:**
- Modify: `src/items/artifact_icons.go` (two new icon generators)
- Modify: `src/data/items_flavor.json` (two new entries)

**Interfaces:**
- Consumes: item IDs `"void_rift_catalyst"`, `"voidbound_pendant"` (Task 2).
- No new exported symbols consumed by later tasks — this is leaf content.

- [ ] **Step 1: Add icon generators**

In `src/items/artifact_icons.go`, add two entries to the `iconGenerators` map (after the `"quicksilver_talisman"` line):

```go
	"void_rift_catalyst": {"void", drawRiftIcon},
	"voidbound_pendant":  {"void", drawPendantIcon},
```

Then add the two draw functions (place after `drawTalismanIcon`, before the `─── low-level primitives ───` section):

```go
// drawRiftIcon: a jagged dark tear ringed by a bright event-horizon —
// void_rift_catalyst.
func drawRiftIcon(img *ebiten.Image) {
	rim := color.NRGBA{190, 130, 255, 255}
	dark := color.NRGBA{35, 15, 55, 255}
	pts := [][2]float32{{16, 6}, {21, 13}, {26, 16}, {20, 19}, {16, 27}, {12, 19}, {6, 16}, {11, 13}}
	for i := 1; i < len(pts)-1; i++ {
		drawFilledTriangle(img, pts[0], pts[i], pts[i+1], dark)
	}
	for i := 0; i < len(pts); i++ {
		p1 := pts[i]
		p2 := pts[(i+1)%len(pts)]
		vector.StrokeLine(img, p1[0], p1[1], p2[0], p2[1], 1.2, rim, true)
	}
	vector.DrawFilledCircle(img, 16, 16, 2, color.NRGBA{230, 210, 255, 255}, true)
}

// drawPendantIcon: a teardrop pendant on a chain, warping the space around
// it — voidbound_pendant.
func drawPendantIcon(img *ebiten.Image) {
	chain := color.NRGBA{150, 100, 220, 200}
	vector.StrokeLine(img, 12, 5, 16, 11, 1, chain, true)
	vector.StrokeLine(img, 20, 5, 16, 11, 1, chain, true)
	gem := color.NRGBA{130, 60, 190, 255}
	drawFlameShape(img, 16, 27, 6.5, gem)
	vector.StrokeCircle(img, 16, 18, 9, 1, color.NRGBA{190, 130, 255, 140}, true)
}
```

- [ ] **Step 2: Add flavor text**

In `src/data/items_flavor.json`, add two entries to the top-level JSON array (after the opening `[` — position doesn't matter, the loader matches by `id`):

```json
  {
    "id": "void_rift_catalyst",
    "text": "It does not open a door. It opens an absence — a place the dungeon forgot to finish building — and whatever stands there falls halfway out of the world for as long as the tear holds.",
    "line": "Something briefly stops existing where you point it."
  },
  {
    "id": "voidbound_pendant",
    "text": "Worn close to the skin, the pendant keeps a fraction of a second in reserve at all times — borrowed from wherever the rifts lead, spent back into every spell you hold a moment longer than it should last.",
    "line": "Time moves slower near the pendant. Not for you."
  },
```

- [ ] **Step 3: Build**

Run: `cd src && go build ./...`
Expected: clean build. `go vet ./items/...` should also be run to catch any unused-variable issues in the new icon functions:

Run: `cd src && go vet ./items/...`
Expected: no output.

---

### Task 7: Dev-menu test path and final verification

**Files:**
- Create: `src/cmd/benchmarker/scenarios/void_rift_test.json`
- Modify: `src/game/devtools.game.go` (one new dev-menu entry)

**Interfaces:**
- Consumes: `combat.LoadScenario`, `Game.devLoadBuild` (both pre-existing, `src/combat/sim.go` and `src/game/combat_debug.go`).
- Terminal task — nothing downstream consumes this.

- [ ] **Step 1: Write the test scenario**

Create `src/cmd/benchmarker/scenarios/void_rift_test.json`:

```json
{
  "name": "Void Rift VFX Test",
  "class": "knight",
  "artifacts": ["void_rift_catalyst", "voidbound_pendant", "", "", "", "", ""],
  "stats": { "Strength": 12, "Dexterity": 10, "Vitality": 12, "Intelligence": 6, "Luck": 6 },
  "floor": 3,
  "biome": "crypt",
  "enemy_pool": ["melee"],
  "iterations": 1,
  "skill_rotation": ["slot_1", "auto"],
  "stat_priority": ["str", "vit"]
}
```

- [ ] **Step 2: Add the dev-menu entry**

In `src/game/devtools.game.go`, inside the `scenarioPairs` slice (the list of `{label, file}` structs feeding the `-- BUILDS --` section), add one entry:

```go
		{"Void Rift Test", "cmd/benchmarker/scenarios/void_rift_test.json"},
```

This reuses the existing loop that turns every `scenarioPairs` entry into a `"Load: <label>"` dev-menu button calling `devLoadBuild()` — no other code changes needed.

- [ ] **Step 3: Build**

Run: `cd src && go build ./...`
Expected: clean build.

- [ ] **Step 4: Full regression pass**

Run: `cd src && go test ./... 2>&1 | tail -20`
Expected: all packages `ok` except `dungeoneer/levels`, which has 2 pre-existing failures (`TestDoorPlacement`, a water-tile-walkability test) unrelated to this work — confirm no *new* failures appear anywhere, and specifically that `combat`, `items`, and `game` all report `ok`.

- [ ] **Step 5: Manual in-game verification**

Launch the game (via the `run` skill, or `src/build_and_run.bat`). In-run:
1. Press F2 to open the dev item spawner / F10 or the dev overlay hotkey to reach "Load: Void Rift Test" under `-- BUILDS --`.
2. Click it — player should be equipped with `void_rift_catalyst` in slot 1 and `voidbound_pendant` passively.
3. Confirm the **Voidbound Aura** (soft pulsing violet glow) is visible around the player immediately, anchored to the player's body — moving the player should keep the aura correctly centered, not lagging or offset.
4. Spawn an enemy (dev menu → "Spawn: Melee Enemy"), target it (C key or click), and fire slot 1 (`void_rift_catalyst`). Confirm the rift visual spawns at the *target's* position (not the cursor, not the player), expands then collapses over ~0.6s, and the target is visibly rooted (its movement should freeze — cross-check against the "Combat State Overlay" dev toggle, which shows a live `Root: ON` badge).
5. Take screenshots of: the idle aura, and the rift at its peak expansion (~40% through its 0.6s duration — roughly 240ms after activation). Compare against the design intent (dark swirling core, bright rim, no positional drift) and iterate on `vfx_shaders.go` / `voidrift.spells.go` parameters (radius, color, timing) if it doesn't read as intended before calling this done.

This is the actual acceptance test for this pilot — per the spec's Verification section, there is no automated substitute for visual-quality judgment.

---

## Self-Review Notes

- **Spec coverage:** all 7 integration points listed in the spec's table are covered (Tasks 2, 3, 4, 5, 6, 7). The spec's "no `src/combat/` changes unless existing fields prove insufficient" open question is resolved concretely in Task 2 — `DamageMultiplier` + `IsRoot` + `DurationSec` suffice, zero engine changes made. The spec's UV-distortion language is explicitly corrected in Global Constraints rather than silently reinterpreted.
- **Type consistency checked:** `VoidRift` (Task 3) and `VoidboundAura` (Task 4) intentionally have different method signatures — `VoidRift` implements `Spell` (`Update(level, dt)`, `Draw(screen, tileSize, camX, camY, camScale, cx, cy)`, `IsFinished()`), `VoidboundAura` does not (`Update(dt)` only, `Draw` takes explicit `worldX, worldY`). This is called out in both tasks' Interfaces blocks so the distinction isn't lost by an implementer working the tasks out of order.
- **No placeholders:** every step has complete, runnable code — no "add error handling" or "similar to Task N" shortcuts.
