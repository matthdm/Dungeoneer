#  Dungeoneer

**Dungeoneer** is a 2D isometric dark fantasy game built using Go and the Ebiten engine. Set in a world of crumbling castles, haunted forests, and cursed ruins, the player must navigate procedurally generated dungeons filled with monsters, secrets, and mystery.

This project is a technical and artistic love letter to games like *Diablo* and *Stardew Valleys* dungeon sections  with a focus on real-time combat, tile-based navigation, and rich pixel art.

---

##  Developer Notes: Work In Progress

_Current Feature Focus:_

- [ ] Implementing dynamic UI system with animated menu
- [~] Finalizing Main Menu visual polish and parallax layers
- [ ] Monster AI improvements (patrolling, boss behaviors)
- [ ] Save/Load system prototype
- [~] Memory and FOV persistence enhancements

---

##   Feature Checklist

| Feature Description                                  | Status   |
|------------------------------------------------------|----------|
| Isometric tile-based rendering engine                |         |
| Smooth tile-based player movement (mouse + keyboard) |         |
| Monster pathfinding & AI                             |         |
| Real-time combat system                              |         |
| Health bars, damage numbers, and hit markers         |         |
| Monster death animations                             |         |
| Field-of-view with raycasting                        |         |
| Fog of war memory system                             |         |
| Click-to-move A* pathfinding                         |         |
| Corner-casting to eliminate visual blind spots       |         |
| Full-bright debug toggle                             |         |
| Interactive level editor                             |         |
| Procedural forest level generator                    |  (Prototype) |
| Procedural dungeon generator                         |  (Prototype) |
| Developer test level                                 |         |
| In-game monster spawning via hotkeys                 |         |
| Particle and visual effects                          |         |
| Save/load system                                     |         |
| Audio engine (SFX + music)                           |         |
| Custom UI system                                     |         |
| Main menu with animated background                   |         |
| Parallax and animated layers                         |         |
| Game state handling (Game Over, Restart)             |         |
| Boss mechanics                                       |         |
| Ranged monsters and projectiles                      |         |
| Inventory system                                     |         |

---

##  Known Issues / Bug Bounties

- [ ] **Bug #1**: Spamming right-click causes player to stutter in place or revert to earlier tiles.
- [x] **Bug #2**: Crash when player character dies without proper state handling (fixed, needs retest). ***(Fixed)***
- [x] **Bug #3**: Corner ray leaks light in some north-wall positions if tile geometry is thin. ***(Fixed)***
- [x] **Bug #4**: UI input not responsive after Game Over screen (fixed via hotkey binding logic). ***(Fixed)***
- [ ] [Performance] Corner-casting rays may tank performance with high tile counts.

---

##  Art Showcase

| Layer / Asset        | Description                            |
|----------------------|----------------------------------------|
| `galaxy_bg.png`      | 12-frame animated pixel art galaxy background |
| `castle_fg.png`      | 12-frame animated dark fantasy castle with glowing towers |
| `fog_layer.png`      | 8-frame animated fog drifting layer     |
| `highlight_anim.png` | 6-frame selection highlight shimmer     |
| `new_game.png`       | Pixel font menu label  New Game        |
| `options.png`        | Pixel font menu label  Options         |
| `exit_game.png`      | Pixel font menu label  Exit Game       |

>  Art and UI are designed in pixel-perfect resolution and support dynamic scaling.

---

##  Hotkeys and Controls

| Key / Button         | Action Description                                                   |
|----------------------|------------------------------------------------------------------------|
| **Q**                | Generates a randomly generated forest level *(prototype)*             |
| **R**                | Generates a randomly generated dungeon level *(prototype)*            |
| **T**                | Loads developer test level                                             |
| **Y**                | Toggles full-bright map rendering (disables shadows & fog)            |
| **G**                | Toggles raycasting debug mode (draws rays on screen)                  |
| **WASD**             | Pans the isometric camera                                              |
| **Mouse 1 (LMB)**    | Attacks monsters within range                                          |
| **Mouse 2 (RMB)**    | Moves player toward selected tile using A* pathfinding                |
| **Middle Mouse**     | Deletes selected tile in editor mode                                  |
| **1**                | Spawns a floor tile in level editor                                   |
| **2**                | Spawns a wall tile in level editor                                    |
| **ESC**              | (Future) Open pause menu                                               |
| **Enter**            | (Menu) Select current menu item                                       |
| **Arrow Keys / W/S** | (Menu) Navigate menu options                                          |
| **R**           | Restart the game after death                                          |

---

##  Get Involved

This game is being built with love, madness, and pixelated shadows. If you're interested in contributing to the art, design, music, or code reach out via the issues!

---

## Development Setup

Prerequisites:
- Go 1.20+ (module-aware)
- Git

Quick start (Windows):

```powershell
git clone <repo-url>
cd Dungeoneer
.\build_and_run.bat
```

Or build manually:

```powershell
go build ./...
.
```

Run tests and linters:

```powershell
go test ./...
# (optional) golangci-lint run
```

Where to look first
- Rendering and game loop: `game/` (e.g., `game/draw.game.go`)
- Field-of-view & fog: `fov/`
- Pathfinding: `pathing/astar.go`
- Movement controller: `movement/controller.movement.go`
- Menus & UI: `ui/`



Contributing
- Open an issue describing proposed changes before large refactors.
- Keep commits small and focused. Add tests for logic-heavy changes (pathing, movement, FOV).


