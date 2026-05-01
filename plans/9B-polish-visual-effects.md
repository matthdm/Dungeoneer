---
plan-id: 9B-polish-visual-effects
status: queued
owner: unassigned
branch: plan/9B-polish-visual-effects
depends-on: [9A-polish-transitions-camera]
last-touched: 2026-04-30
---

# Plan: Phase 9B — Visual Effects (Particles & Damage Numbers)

## Goal

Add a pooled particle system and floating damage numbers. Spell impacts burst with color-coded particles, enemy deaths dissolve with a fade effect, item pickups sparkle, and every hit shows a floating damage number. These systems dramatically improve perceived game feel without changing any gameplay logic.

## Scope

**In scope:**
- Lightweight pooled particle emitter/particle model (no per-frame heap allocations in hot path).
- Spell impact particles: color per damage type (fire=orange, lightning=blue/white, nature=green).
- Enemy death fade/dissolve effect.
- Item pickup sparkle burst.
- Floating damage numbers: color-coded by type (physical=white, spell=color of spell type, crit=yellow+larger).

**Out of scope (do not change in this plan):**
- Audio (plan `9C-polish-audio.md`).
- Screen shake (plan `9A`).
- Echo entity visual style (already handled in `7A`).

## File envelope

**Touched:**
- `src/game/particles.go` *(new)* — particle pool, emitter, update/draw
- `src/game/damage_numbers.go` *(new)* — floating number pool, update/draw
- `src/game/draw.game.go` — draw particles and damage numbers after entities
- `src/game/game.go` — store particle system and damage number pool; update each tick
- `src/spells/` — emit particles on spell hit
- `src/entities/monster.go` — emit death fade; emit pickup sparkle on item drop
- `src/game/handlers.game.go` — emit sparkle on item pickup
- `design-docs/roadmap.md` — mark Phase 9 visual effect rows ✅ on completion
- `CLAUDE.md` — update Phase 9 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/audio/` — audio plan scope
- Any gameplay logic files (`runstate.go`, `metasave.go`, etc.)

## Acceptance criteria

- [ ] Particle pool pre-allocates 512 particles; no heap allocation during particle emission in steady state.
- [ ] Spell impact at target position emits 8–12 particles in the spell's color, each with 0.3–0.5s lifetime and velocity spread.
- [ ] Enemy death triggers a 0.5s alpha fade on the enemy sprite before removal.
- [ ] Item pickup emits a small sparkle burst (6–8 particles, gold/white, 0.4s lifetime).
- [ ] Every damage event spawns a floating number at the hit position. Numbers float upward 20px and fade over 0.8s. Crits are yellow and 1.3× font size.
- [ ] Damage number pool pre-allocates 64 entries; no heap allocation in steady state.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Particle system
- [ ] 1.1 Create `src/game/particles.go`. `Particle` struct: position, velocity, color (RGBA), lifetime, maxLifetime. `ParticlePool` pre-allocates `[512]Particle`.
- [ ] 1.2 `Emit(x, y, count int, color color.RGBA, speed float64)` activates N idle particles from pool with randomized velocity.
- [ ] 1.3 `UpdateAll(dt float64)` advances all active particles; deactivates expired.
- [ ] 1.4 `DrawAll(screen *ebiten.Image)` draws active particles as 2×2 pixel squares with alpha proportional to remaining lifetime.
- [ ] 1.5 Store `ParticlePool` on `Game`, call Update/Draw each frame.
- [ ] 1.6 `cd src && go build ./...` passes.

### Phase 2: Spell and death effects
- [ ] 2.1 In `src/spells/` hit callback, call `particles.Emit()` with the spell's color constant (define `SpellColors map[string]color.RGBA` in particles.go).
- [ ] 2.2 In `src/entities/monster.go` death path, set `monster.Dying = true` and `monster.DyingAlpha = 1.0`. In draw, render dying monsters at `DyingAlpha`; update decrements it by `2.0 * dt`; remove when 0.
- [ ] 2.3 `cd src && go build ./...` passes.

### Phase 3: Pickup sparkle and damage numbers
- [ ] 3.1 In `src/game/handlers.game.go` item pickup path, call `particles.Emit()` with gold color.
- [ ] 3.2 Create `src/game/damage_numbers.go`. `DamageNumber` struct: position, value, color, lifetime. `DamageNumberPool` pre-allocates `[64]DamageNumber`.
- [ ] 3.3 `SpawnNumber(x, y float64, value int, isCrit bool, damageType string)` — activates a pool entry with appropriate color and size multiplier.
- [ ] 3.4 `UpdateAll(dt float64)` floats numbers upward (20px/s), fades alpha.
- [ ] 3.5 `DrawAll(screen *ebiten.Image)` draws numbers with Ebiten text at appropriate position and alpha.
- [ ] 3.6 Wire `SpawnNumber()` into combat hit paths in `entities/monster.go` and `entities/player.go`.
- [ ] 3.7 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark Phase 9 visual effect roadmap rows ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 9 status block.
- [ ] 4.3 Move this plan to `plans/COMPLETED/9B-polish-visual-effects.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Ebiten text for damage numbers.** Confirm the text rendering approach used elsewhere in the HUD (font face, size). Use the same font to avoid introducing a new asset dependency. Affects 3.5.
- **Particle draw layer.** Particles should render above entities but below the HUD. Confirm the draw order in `draw.game.go` and insert the particle draw call at the correct layer. Affects 1.4.
