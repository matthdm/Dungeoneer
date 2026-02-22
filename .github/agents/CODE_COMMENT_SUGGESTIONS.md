# CODE_COMMENT_SUGGESTIONS.md

This document lists suggested inline comment and doc improvements across the codebase. These are recommendations only — no source files are modified.

Purpose
- Help future maintainers understand hot paths, performance assumptions, and rendering math without changing code now.

Suggested places to add comments (examples)

- `fov/` (e.g., `fov.go`, `render.go`, `walls.go`)
  - Explain the raycasting approach used, including any corner-casting assumptions.
  - Note why caching FOV masks is safe/when it needs invalidation (dynamic lights, player movement).
  - Add math comment for angle/step calculations and integer-to-world coordinate conversions.

- `pathing/astar.go`
  - Document node memory layout and heuristics used (Manhattan vs Euclidean).
  - Mark which allocations are performed per-query and suggest reuse points for node pools.
  - Add a simple complexity note (expected O(N log N) with heap) and constraints (grid size).

- `movement/controller.movement.go`
  - Clarify the decision to use fixed-tick movement vs velocity integration (Ebiten 60 TPS assumptions).
  - Document race conditions that may occur when path is reissued rapidly (right-click stutter bug).

- `game/draw.game.go` and `game/render_collect.game.go`
  - Note rendering order assumptions for isometric layering and why feet-center anchoring matters.
  - Document any offscreen buffer reuse and the lifetime of images to avoid GC churn.

- `sprites/spritesheet.go`
  - Add comment about tile grid size (64×64) and sprite anchor conventions.

- `ui/` (menus)
  - Document shared menu state vs per-menu logic; point to `ui/menu.go` as canonical.

- General: add package-level README comments
  - At the top of each package add a short comment describing responsibilities and invariants (e.g., `// Package pathing implements grid-based A* pathfinding used by entities.`)

Example comment snippets

```
// Cache FOV results per player position + light state. Invalidate when dynamic lights
// change or when the player teleports. This avoids expensive per-frame raycasts.
```

```
// A* uses Manhattan distance on the 64×64 grid for speed and determinism. If diagonal
// movement cost changes, update the heuristic accordingly to keep admissibility.
```

Next steps
- If you'd like, I can open PRs that add these comments into targeted files (small patches), or produce a branch with the changes for review.
