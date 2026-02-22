# Dungeoneer Copilot Agent — Context

Project: Dungeoneer — a 2D isometric dungeon crawler written in Go using the Ebiten engine.

Purpose: provide a compact, focused workspace context for a Copilot-style agent that will assist development tasks across this repository.

Key repo facts
- Language: Go (modules) — go.mod present
- Build: build_and_run.bat at repo root; standard `go build` and `go test` apply
- Main entry: main.go

Relevant packages and responsibilities
- `game/` : main game loop, rendering orchestration, entity updates
- `entities/` : player, monsters, hitmarkers, inventory helper logic
- `fov/` : field-of-vision, fog-of-war and shadow rendering
- `movement/` : movement controller abstraction
- `pathing/` : A* implementation and helpers
- `levels/` : tile map, generators (64×64 grid assumption)
- `sprites/` and `images/` : spritesheet and embedded assets
- `ui/`, `hud/` : menus and HUD components

Hard constraints & engine assumptions
- Tile grid: 64×64 tile size; spritesheet tiles 64×64 pixels
- Player sprite anchor: feet center (rendering order matters)
- Fixed game loop: Ebiten-style 60 TPS; logic expects fixed-tick timing
- No networking — offline single-player only

Design & coding principles (authoritative)
- Performance first: avoid per-frame allocations, reuse buffers
- Modularity: keep packages focused and decoupled
- Idiomatic Go: clear names, simple functions, avoid globals except constants
- Debug/Dev parity: debug tools shouldn't impact release builds

Where to look first
- Rendering: `game/draw.game.go`, `fov/render.go`
- Pathfinding: `pathing/astar.go`
- Movement: `movement/controller.movement.go`
- Menus/UI: `ui/` folder

Notes for the agent
- Always favor minimal, surgical edits that preserve style and performance
- Leave TODOs where behavior is unclear rather than guessing larger design choices
- Do not implement future expansion features (see AGENTS.md) unless explicitly requested
