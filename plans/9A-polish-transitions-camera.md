---
plan-id: 9A-polish-transitions-camera
status: queued
owner: unassigned
branch: plan/9A-polish-transitions-camera
depends-on: [8B-hub-shop-upgrades]
last-touched: 2026-04-30
---

# Plan: Phase 9A — Transitions, Camera & Screen Shake

## Goal

Add professional game feel through smooth screen transitions (fade-to-black between floors), a boss intro camera pan, and screen shake on heavy hits and boss phase transitions. These are purely presentational — no gameplay logic changes.

## Scope

**In scope:**
- Fade-to-black transition overlay on floor change and hub load (configurable duration, default 0.4s).
- Screen shake system: intensity float, decay curve, applied as a draw offset in the render pass.
- Boss intro sequence: camera pans to boss room, brief pause, combat begins.

**Out of scope (do not change in this plan):**
- Particle effects (plan `9B-polish-visual-effects.md`).
- Audio (plan `9C-polish-audio.md`).
- Minimap or HUD additions (plan `9D-polish-hud-navigation.md`).

## File envelope

**Touched:**
- `src/game/transition.go` *(new)* — fade overlay, duration, state machine
- `src/game/screenshake.go` *(new)* — shake state, decay, draw offset
- `src/game/game.go` — integrate transition and shake into update/draw
- `src/game/draw.game.go` — apply shake offset to camera; draw fade overlay last
- `src/game/hub.go` — trigger fade-out on floor transition, fade-in on new floor
- `src/game/boss.game.go` — boss intro camera pan sequence
- `design-docs/roadmap.md` — mark Phase 9 transition rows ✅ on completion
- `CLAUDE.md` — update Phase 9 status block
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/entities/` — no entity logic changes
- `src/audio/` — doesn't exist yet

## Acceptance criteria

- [ ] Floor transitions fade to black over 0.4s, then fade in on the new floor.
- [ ] Hub load fades in from black.
- [ ] Screen shake triggers on: player taking heavy damage (>20% max HP in one hit), boss phase transition, explosion-type spell impact.
- [ ] Shake intensity decays smoothly over 0.3–0.5s with no abrupt cutoff.
- [ ] Boss intro: on boss room entry, camera pans to boss spawn point over 1.5s, holds for 0.5s, then combat begins.
- [ ] All effects are purely cosmetic — no gameplay events fire during transition frames.
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Fade transition
- [ ] 1.1 Create `src/game/transition.go`. `Transition` struct: `Active bool`, `Phase` (fade-out / hold / fade-in), `Progress float64`, `Duration float64`, `OnComplete func()`.
- [ ] 1.2 `Update(dt float64)` advances progress; calls `OnComplete` when fade-in reaches 1.0.
- [ ] 1.3 `Draw(*ebiten.Image)` draws a black rectangle at alpha `easeInOut(Progress)` over the full screen.
- [ ] 1.4 In `src/game/hub.go` floor transition: start fade-out, in `OnComplete` generate new floor and start fade-in.
- [ ] 1.5 `cd src && go build ./...` passes.

### Phase 2: Screen shake
- [ ] 2.1 Create `src/game/screenshake.go`. `ScreenShake` struct: `IntensityX, IntensityY float64`, `Decay float64`.
- [ ] 2.2 `Trigger(intensity float64)` sets random X/Y intensity up to the given magnitude.
- [ ] 2.3 `Update(dt float64)` decays intensity by `Decay * dt`.
- [ ] 2.4 `Offset() (float64, float64)` returns current random offset scaled by intensity.
- [ ] 2.5 In `src/game/draw.game.go`, apply shake offset to the camera translation matrix before drawing entities.
- [ ] 2.6 Wire shake triggers: player `OnHeavyHit`, boss phase transition, spell impact callback.
- [ ] 2.7 `cd src && go build ./...` passes.

### Phase 3: Boss intro camera pan
- [ ] 3.1 In `src/game/boss.game.go`, on boss room entry: freeze player input, start a `CameraPan` to boss spawn position over 1.5s, hold for 0.5s, unfreeze input.
- [ ] 3.2 `CameraPan` uses a linear or ease-in-out interpolation on the camera's target position.
- [ ] 3.3 Pre-fight dialogue fires after the pan completes (coordinate with existing dialogue trigger).
- [ ] 3.4 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark Phase 9 transition/camera roadmap rows ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 9 status block.
- [ ] 4.3 Move this plan to `plans/COMPLETED/9A-polish-transitions-camera.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Camera system existence.** Confirm whether `draw.game.go` already has a camera offset/matrix or draws entities in world-space directly. If the latter, the shake offset requires introducing a camera translation — potentially a larger change. Investigate before starting Phase 2. Affects 2.5.
