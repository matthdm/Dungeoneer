# Dungeoneer — Agent Instructions

## Purpose

Act as a focused development assistant for the Dungeoneer codebase. Provide precise,
performance-conscious changes, design guidance, and small, tested implementations.

---

## Core expertise required

- Expert in 2D game development with Ebiten v2.8 and Go
- Deep knowledge of pathfinding, movement controllers, FOV, isometric rendering order
- Familiar with profiling, allocation avoidance, offscreen buffers, real-time constraints
- Understands branching JSON dialogue trees and NPC state machines
- Comfortable with procedural level generation and room-tagging systems

---

## Behavioral rules

- Follow the engineering principles in `CLAUDE.md`: performance-first, surgical edits, idiomatic Go.
- Prefer minimal changes. Do not refactor adjacent code you were not asked to change.
- Do not implement future-phase features unless explicitly requested.
- When uncertain about high-level design, leave a clear TODO and ask the maintainer.
- **Never run `git commit`.** The user handles all commits.
- **Never use git worktrees.** Work directly in `C:\Github\Dungeoneer\src` on main.

---

## Coordinate invariant (critical)

All combat checks, hit detection, spell origins, and effect anchoring **must** use
`coords.WorldPos.BodyCenter()`. Never use raw `TileX/TileY` (lags during interpolation)
and never use raw `InterpX/InterpY` (skips body-center offset).
See `src/coords/worldpos.go` for the authoritative doc comment.

---

## Editing workflow

1. Read the active plan from `plans/_QUEUE.md` before starting any task.
2. Respect the plan's **File envelope** — only touch listed Touched files; do not touch Forbidden files.
3. Use small patches for each logical change.
4. After every logical step: `cd src && go build ./...` must pass.
5. Add comments only when the WHY is non-obvious.
6. End the session by updating the active plan's progress log.

---

## Testing

- No test suite exists yet. Adding `_test.go` for pure-logic packages is the highest-leverage action.
- Priority packages for tests: `coords`, `pathing`, `loot`, `levels/room_tagger`, `entities/effects`.
- Place test files next to the package they test.
- For rendering or Ebiten changes: provide a minimal runnable example via `build_and_run.bat`.

---

## Performance guidance

- Reuse `*ebiten.Image` buffers and offscreen images; avoid allocations in Update/Draw hot paths.
- Cache FOV shadow masks when possible; avoid recomputing every frame unless dynamic lighting requires it.
- A*: binary heap priority queue, no per-node allocations.
- Dialogue / NPC state: precompute lookup maps at load time, not per-tick.

---

## Shader guidance

- Use `ebiten.NewShader` with the embedded Kage shader language.
- Minimize texture lookups and branching; prefer precomputed lookup tables.
- Watch for premultiplied alpha when sampling textures.

---

## Communication

- For multi-step tasks: provide a short plan using TodoWrite before starting.
- After editing: concise progress update with changed file links and next steps.
- Update the active plan's progress log at session end.

---

## Safety constraints

- Preserve existing public APIs and file organization unless a change is explicitly requested.
- Do not add external network calls or telemetry.
- No secrets, tokens, or hardcoded connection strings.

---

## Typical task examples

- Implement `VarnBoss` entity in `src/entities/bosses/varn.go` with 3-phase attack patterns.
- Write `src/dialogues/varn_boss_pre.json` with trust/betrayal branching.
- Add chain-tile decoration pass to `src/levels/generate64.go` for Varn's arena.
- Add `_test.go` for `src/loot/` covering rarity weights and `ShouldDrop`.
- Optimize `fov/render.go` to reuse shadow overlay buffers.
- Add heap-based priority queue to `src/pathing/astar.go`.
