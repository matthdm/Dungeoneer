---
plan-id: backlog-fov-dynamic-ray
status: queued
owner: unassigned
branch: plan/backlog-fov-dynamic-ray
depends-on: []
last-touched: 2026-04-30
---

# Plan: Backlog B.1 — Dynamic FOV Ray Length

## Goal

Replace the hardcoded `rayLength=1000` in `src/fov/fov.go:39` with a value derived from the map diagonal (`sqrt(W²+H²) * 1.5`). This is a prerequisite for any map size beyond 64×64. For current 64×64 maps the effective behavior is unchanged; this just removes a hidden landmine.

This plan is independent of all other work and can be dropped into any available slot before map scaling is attempted.

## Scope

**In scope:**
- Replace the hardcoded constant with a computed value passed in or derived from map dimensions.
- Add a unit test confirming the ray length is correct for 64×64, 128×128, and 256×256 maps.

**Out of scope (do not change in this plan):**
- Map size scaling itself (Backlog B.6).
- A* heap upgrade (Backlog B.2).
- Any other FOV logic changes.

## File envelope

**Touched:**
- `src/fov/fov.go` — replace hardcoded `rayLength`
- `src/fov/fov_test.go` *(new)* — unit test for ray length formula
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/pathing/` — separate backlog item
- `src/coords/`, `src/collision/` — offset plan
- Everything else

## Acceptance criteria

- [ ] `rayLength` in `src/fov/fov.go` is computed as `math.Sqrt(float64(w*w+h*h)) * 1.5` where `w` and `h` are the map width and height in tiles.
- [ ] The value is passed in or available at FOV computation time (not a package-level var).
- [ ] `fov_test.go` confirms the computed ray length for a 64×64 map is approximately 135.8 (√8192 × 1.5) and for 128×128 is approximately 271.5.
- [ ] `cd src && go build ./...` passes.
- [ ] `cd src && go test ./...` passes.

## Phases

### Phase 1: Fix and test
- [ ] 1.1 In `src/fov/fov.go`, locate the `rayLength=1000` constant. Determine how `w` and `h` are available at the call site (map struct, params, or global). Update to `math.Sqrt(float64(w*w+h*h)) * 1.5`.
- [ ] 1.2 Create `src/fov/fov_test.go`. Write a test that calls the ray-length calculation function (extract it to a helper if needed) and asserts expected values for 64×64 and 128×128.
- [ ] 1.3 `cd src && go build ./... && go test ./...` passes.

### Phase 2: Cleanup
- [ ] 2.1 Move this plan to `plans/COMPLETED/backlog-fov-dynamic-ray.md`.
- [ ] 2.2 Update `plans/_QUEUE.md`. Note B.1 resolved in the roadmap backlog table.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **How `w`/`h` are currently passed to FOV.** Check `fov.go` to see if map dimensions are already available in the function signature or if they need to be threaded through. Affects 1.1.
