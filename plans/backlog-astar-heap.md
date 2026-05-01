---
plan-id: backlog-astar-heap
status: queued
owner: unassigned
branch: plan/backlog-astar-heap
depends-on: []
last-touched: 2026-04-30
---

# Plan: Backlog B.2 — A* Heap Upgrade

## Goal

Replace the slice-based open list in `src/pathing/astar.go:77-100` with a `container/heap` min-heap. The current O(n) extraction becomes O(log n), eliminating a performance cliff when monster count or map size increases. This is a pure algorithmic improvement with no observable behavior change.

This plan is independent and can run in any available slot. The trigger is frame drops when monster count or map size increases; proactively doing it now is cheap insurance.

## Scope

**In scope:**
- Replace the open list with a `container/heap` implementation.
- Add A* unit tests covering: finds optimal path, handles no-path case, handles single-tile path, large open area performance test (128×128 map, 50 pathing calls).

**Out of scope (do not change in this plan):**
- A* heuristic changes.
- Map size scaling (B.6).
- Dynamic FOV (B.1).
- Any entity or game logic.

## File envelope

**Touched:**
- `src/pathing/astar.go` — replace open list implementation
- `src/pathing/astar_test.go` *(new)* — unit tests
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/fov/` — separate backlog item
- `src/coords/`, `src/collision/` — offset plan
- Everything else

## Acceptance criteria

- [ ] A* open list uses `container/heap` with O(log n) extraction.
- [ ] All existing paths that were valid before are still valid after (behavior unchanged).
- [ ] Unit tests cover: shortest path on a small grid (5×5), no-path case (blocked room), single adjacent tile, diagonal path. Tests use a test grid, not game state.
- [ ] Performance test: 50 A* calls on a 128×128 grid with 20% wall density completes in under 5ms on a reasonable laptop. (Benchmark test, not a hard assertion — use `testing.B`.)
- [ ] `cd src && go build ./...` passes.
- [ ] `cd src && go test ./...` passes.

## Phases

### Phase 1: Heap implementation
- [ ] 1.1 Define `openHeap` type implementing `heap.Interface` (`Len`, `Less`, `Swap`, `Push`, `Pop`). Each entry is a `*Node` (or index+priority pair).
- [ ] 1.2 Replace the existing open list slice in `astar.go` with `openHeap`. Replace manual min-search with `heap.Pop()`. Replace append with `heap.Push()`.
- [ ] 1.3 `cd src && go build ./...` passes.

### Phase 2: Tests
- [ ] 2.1 Create `src/pathing/astar_test.go`. Build a helper `makeGrid(w, h int, walls []Point) Grid` for test grids.
- [ ] 2.2 Write correctness tests: 5×5 open grid (optimal path length = Manhattan distance), blocked grid (nil path), single step.
- [ ] 2.3 Write a benchmark `BenchmarkAStarLargeGrid` — 128×128 grid, 20% random walls, path from corner to corner, 50 iterations.
- [ ] 2.4 `cd src && go test ./...` passes; `go test -bench=. ./pathing/...` runs without panic.

### Phase 3: Cleanup
- [ ] 3.1 Move this plan to `plans/COMPLETED/backlog-astar-heap.md`.
- [ ] 3.2 Update `plans/_QUEUE.md`. Note B.2 resolved in the roadmap backlog table.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Node struct shape.** Confirm what fields the current `Node` or open-list entry has in `astar.go` before writing the heap wrapper. The `Less` comparator needs the f-score field name. Affects 1.1.
