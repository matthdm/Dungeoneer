# Void Rift + Voidbound Aura — VFX Pilot Design Spec

**Date:** 2026-07-26
**Status:** Approved, pending implementation plan
**Scope:** A new Kage-shader-based visual effects pipeline for Dungeoneer's spell/skill VFX, proven out on one brand-new CR1 void-domain artifact pair rather than retrofitted onto an existing spell. The resulting shader primitives become the reference model future spell/artifact visuals are built from. Legacy spell visuals (Fireball, Lightning Storm, Chaos Ray, Fractal Canopy, etc.) are explicitly untouched in this pass.

---

## Motivation

Current spell VFX (`src/spells/*.go`) is entirely CPU-side immediate-mode drawing: `vector.StrokeLine`, flat-color `DrawTriangles`, single pre-rendered `ebiten.Image` blits, linear color lerps. There is no glow, bloom, additive blending, or distortion anywhere in the pipeline. This reads as flat/uninspired even on the more structurally complex effects (Fractal Canopy's recursive branch growth still looks like thin colored lines).

Rather than retrofit an existing, already-liked spell (Fireball was considered and explicitly rejected — "I already like how it works"), this pilot builds a **new** spell + enchantment pair from scratch against the new shader pipeline, so the pipeline is proven end-to-end (mechanics → engine → visual bridge → shader render → equip/test loop) before being applied to the existing roster in a follow-up pass.

---

## Domain & Concept

**Domain:** Void. Chosen because it currently has zero unique bespoke visual identity in the codebase (existing void artifacts — `ashbound_chain`, `grave_reaper`, `blood_price`, `soul_harvest` — largely fall back to the generic domain-colored particle burst, aside from root/execute/blood-price strikes already covered in `src/spells/artifacts.spells.go`). A new void effect adds no overlap with prior bespoke work, and gravitational/distortion shaders are the strongest possible demonstration of what Kage shaders can do that the current CPU vector approach cannot — a genuine proof of value, not just a prettier version of an existing effect.

### Item 1 — `void_rift_catalyst` (active skill artifact)
Targeted skill: a projectile opens a collapsing dark rift at the target position — screen-space UV distortion pulling nearby pixels inward, dark purple/black energy core with a thin bright event-horizon rim and chromatic fringing at the edge. Mechanically: damage + brief root, expressed with the existing `ArtifactEffect.IsRoot` / `DamageFlat` fields already wired into the combat engine (CR1-C). If those fields are sufficient, **no `src/combat/` changes are needed** — this stays a visual-layer pilot, not new engine work. If a genuinely new mechanic is required, that must be flagged before implementation, not silently added.

### Item 2 — `voidbound_pendant` (passive "enchantment" artifact)
`IsPassive`, no active skill. Grants a persistent Voidbound Aura visual — a slow, subtle gravitational-lensing distortion field clinging to the caster while equipped — plus a small passive stat/mechanical effect consistent with CR1's "every item changes the character" rule.

Both items follow the mandatory-stat-bonus convention already used by all 23 existing CR1 artifacts.

---

## Visual Architecture

### Shared primitives — `src/spells/vfx_shaders.go` (new)
- Kage shader sources compiled **once** at package load via `ebiten.NewShader` (never per-frame — this is a hard requirement, see Non-Goals/Guardrails).
- `drawGlowQuad(...)` — additive-blended radial glow.
- `drawDistortionQuad(...)` — screen-space UV-warp shader, used by both the rift detonation and the aura.
- Both take the same camera-transform signature already used throughout `src/spells` (`isoToScreenFloat` + `camX, camY, camScale, cx, cy`), so they drop into the existing per-spell `Draw` methods without touching coordinate math.

### Active skill — `src/spells/voidrift.spells.go` (new)
Implements the existing `Spell` interface (`Update`/`Draw`/`IsFinished`), same shape as `FractalNode` (age/Duration lifecycle), using `drawDistortionQuad` for the collapsing-rift phase.

### Passive aura — new small struct, not a `Spell`
Voidbound Aura is persistent while equipped, not spawned/despawned like a triggered skill effect, so it does not fit the `Spell` interface's finite lifecycle. It is drawn every frame from `draw.game.go`, gated on the artifact ID being present in `combatState.PassiveArtifacts` — mirroring the existing continuous-passive-effect pattern already used for burn DoT particles (`new_adapter.go:419-429`, the `BurnActive` check), but driving shader rendering instead of a particle emit.

**Position source:** `coords.WorldPos.BodyCenter()` — per the coordinate invariant in `CLAUDE.md`. This is called out explicitly because "aura fires in the wrong position" was the original complaint that motivated this whole effort; the aura must never be positioned from raw `TileX/TileY` or raw `InterpX/InterpY`.

---

## Integration Points

| File | Change |
|------|--------|
| `src/items/load.go` | Register `void_rift_catalyst` and `voidbound_pendant` via `registerNewArtifacts()`, following existing artifact registration pattern. |
| `src/spells/vfx_shaders.go` (new) | Shared Kage shader primitives, compiled once. |
| `src/spells/voidrift.spells.go` (new) | `Spell`-interface implementation for the active rift effect. |
| `src/game/new_adapter.go` | One new `spawnSkillVisual` case for `void_rift_catalyst`, following the bespoke-visual pattern already established for the 7 CR1-F artifact skills. |
| `src/game/draw.game.go` | One new per-frame draw call for the Voidbound Aura, gated on `PassiveArtifacts` membership. |
| `src/items/artifact_icons.go` | Reuse existing procedural-icon pattern for both new items — no new art asset pipeline. |
| `src/data/items_flavor.json` | Flavor text entries for both items, following existing convention. |
| `cmd/benchmarker/scenarios/void_rift_test.json` (new) | Scenario file equipping both new artifacts, for the dev-menu test path below. |
| `src/game/devtools.game.go` | One new "Load: Void Rift Test" dev entry, calling `devLoadBuild()` on the new scenario — mirrors the existing 14 "Load: <build>" entries. |

---

## In-Game Testing Path

The existing dev overlay (F2, `devtools.game.go`) already has a proven pattern for this: "Load: <build>" entries call `devLoadBuild()` (`combat_debug.go:20`), which writes artifact IDs directly into the player's spell slots **mid-run**, bypassing the normal hub unlock/MetaSave loadout flow entirely. This pilot adds one more entry using the same mechanism, so both new artifacts can be equipped and tested immediately from the dev menu with no unlock grind and no hub trip.

---

## Verification

- `go build ./...` and `go test ./combat/...` must stay green throughout.
- No new combat-engine tests are expected unless Void Rift's mechanic can't be expressed in existing `ArtifactEffect` fields (see Open Question below).
- No automated visual-quality test exists or is proposed — not feasible for shader aesthetics. Verification is manual: launch via the `run` skill, trigger the effect via the new dev-menu entry, compare screenshots (both Claude-captured and user-provided) against the design intent, iterate shader parameters.

---

## Non-Goals / Guardrails

- Fireball and all other legacy spells (`lightningstorm.spells.go`, `chaosray.spells.go`, `fractalbloom.spells.go`, `fractalcanopy.spells.go`, etc.) are **not modified** in this pass.
- The known per-frame allocation bug in `drawAOETile` (`lightningstorm.spells.go:159`, allocates `ebiten.NewImage(1,1)` every call) is **not fixed** here — it lives in an unrelated file and is out of scope for this pilot. Tracked separately.
- No new `src/combat/` engine mechanics unless Void Rift's effect genuinely cannot be expressed with existing `ArtifactEffect` fields (`IsRoot`, `DamageFlat`, etc.) — if so, this must be flagged explicitly during implementation rather than silently expanding scope.
- Shaders are compiled once at package load, never per-frame or per-draw-call — matches the project's no-per-frame-allocation-in-hot-paths standard (`CLAUDE.md`, `.github/agents/instructions.md`).

---

## Open Questions

| # | Question | Resolution plan |
|---|----------|-----------------|
| 1 | Does Void Rift's root+damage fit existing `ArtifactEffect` fields cleanly, or does the "collapsing rift" pull effect need new engine state? | Confirm during implementation. If existing fields suffice, zero `src/combat/` changes. If not, flag before writing engine code — do not expand scope silently. |
| 2 | Exact aura visual tuning (opacity, distortion strength, radius) | Cannot be fully specified in text — resolved iteratively via screenshot review per the Verification section. |

---

## Success Criteria

- Both artifacts build clean, equip via the new dev-menu entry, and are visually and mechanically functional in-run.
- Void Rift's detonation and the Voidbound Aura both render with correct world-position anchoring at all times (moving player, moving target) — no "wrong position" artifacts.
- Shader primitives in `vfx_shaders.go` are written generically enough to be reused by at least the next follow-up spell without rewriting the compile/draw plumbing.
- `go build ./...` and `go test ./...` (excluding the 2 known pre-existing `levels` package failures) stay green.
