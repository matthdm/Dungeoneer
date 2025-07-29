# AGENTS.md

## Project Overview: Dungeoneer

Dungeoneer is a 2D isometric dungeon crawler game written in Go using the [Ebiten](https://ebiten.org/) engine. It is a dark fantasy homage to games like **Diablo**, **Darkstone**, and classic **dungeon crawlers**. The project emphasizes:

- Tight real-time combat
- Fog-of-war and lighting effects
- Procedural level generation
- Modular AI and spellcasting systems
- A nostalgic atmosphere with tactical gameplay and narrative depth

The core of the game is tile-based, with layered levels (multi-floor dungeons), real-time pathfinding, a field-of-vision system, and a reactive UI.

---

## Core Engineering Goals

> The code must remain **modular**, **idiomatic**, and **performance-conscious** — above all, it must feel like a handcrafted dungeon crawler.

### Coding Principles
- **Performance matters**. Avoid unnecessary allocations or repeated expensive operations during the game loop.
- **Modularity wins**. Code should be organized into clearly defined packages with minimal interdependence.
- **Idiomatic Go**. Favor simplicity, clarity, and readability over clever hacks.
- **Debuggability is key**. Include in-code comments where needed, especially for math, rendering, or ECS logic.
- **Functionality before polish**. All features must be testable and integrate cleanly before expanding scope.

---

## Game Architecture

### Packages and Responsibilities

| Package        | Responsibility                                                                 |
|----------------|---------------------------------------------------------------------------------|
| `game`         | Main game loop, player/monster updates, rendering orchestrator                  |
| `entities`     | Player, Monster, Spell, HitMarker, and shared logic for in-world entities       |
| `movement`     | `MovementController` abstraction supporting pathing and velocity-based movement |
| `pathing`      | Grid-based A* implementation and helpers                                        |
| `fov`          | Field-of-vision and fog-of-war raycasting                                       |
| `levels`       | Tile grids, procedural level generators (maze, forest, dungeon), layering       |
| `collision`    | Basic box/collision prediction and resolution                                   |
| `ui`           | Menu system, HUD components, pause/options menus                                |
| `items`        | Inventory, equipment, item registry, and item effects (WIP)                     |
| `editor`       | In-game level editor with sprite palette, brush tools, and save/load            |
| `sprites`      | SpriteSheet loader and sprite indexing helpers                                  |
| `assets`       | Embedded asset loader for audio, images, levels                                 |

---

## Current Goals for Codex Agents

### Priority: Optimization and Refactor

Focus areas:
1. **Rendering**
   - Cache FOV shadow data; avoid redundant draw calls.
   - Consolidate sprite rendering logic (e.g., sprite flipping, vertical bobbing).
   - Avoid per-frame creation of small images (e.g. health bars).
2. **Pathfinding**
   - Refactor A* to use a binary heap or priority queue.
   - Reduce hover path preview calculations by caching results.
3. **Movement**
   - Use shared `MovementController` logic for both player and monster entities.
   - Prevent movement stutter when path is reissued via right-click.

### Refactor Tasks

- Eliminate duplicate logic between main menu and pause menu. Unify under `ui.Menu`.
- Centralize control mappings and keybindings for easier modification and expansion.
- Ensure debug tools do not trigger unnecessary work in release builds (e.g., rendering FOV rays).

---

## 🎮 Gameplay Design Values

> Codex should not just “make it work,” but help maintain the soul of the game.

- **Tactical combat**: Positioning, timing, and visibility matter.
- **Isometric immersion**: Everything is viewed through a skewed grid; rendering order matters.
- **Exploration through obscurity**: Use fog-of-war, secrets, and dark visuals to evoke mystery.
- **Real-time flow**: Combat, movement, and AI should feel fluid, even when tile-based under the hood.
- **Narrative consequence** *(planned)*: The Dungeon is a purgatory. Player decisions ripple into outcomes.

Do not sacrifice these principles for convenience or oversimplification.

---

## Best Practices for Codex Submissions

- **Always run and test**: Code changes must compile, run, and integrate with the rest of the project.
- **Avoid globals unless they are constants**: Favor passing in dependencies or using struct receivers.
- **Be mindful of ticks vs delta time**: Ebiten uses a fixed 60 TPS game loop; logic should match.
- **Use `defer` smartly**: Especially when managing resources (e.g. offscreen buffers).
- **Name things clearly**: Especially in visual code. Example: `shadowOverlay` is better than `tmpImage`.
- **Add TODOs for uncertain logic**: Leave breadcrumbs for humans to investigate or optimize further.

---

## Known Constraints

- Game logic assumes 64×64 tile grid unless otherwise stated.
- Spritesheets use a 64x64 pixel grid (17×12 tiles per sheet).
- Player sprite anchor is aligned to **feet center**, not sprite center.
- All logic must function offline — there is no server component.

---

## Future Expansion Plans (Do Not Implement Unless Told)

These are active design goals but **should not be implemented by agents unless explicitly requested**:

- Save/load system using JSON serialization
- Dynamic NPC interactions and branching quests
- Status effects and consumable item effects
- Non-player light sources and light decay
- Boss AIs with multiple states and patterns
- Shader integration (post-processing or palette-based lighting)
- Multiplayer or netcode

---

## 🧾 Project Philosophy

Dungeoneer is not meant to be modern or hyper-polished. It is a **dark, crunchy, lo-fi experience**. Codex should prioritize **functionality**, **readability**, and **modularity** over feature bloat or premature optimization. Each function should feel handcrafted and self-contained. Like the dungeons themselves, the code should reveal its secrets slowly, but clearly.

---

