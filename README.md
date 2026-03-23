#  Dungeoneer

**Dungeoneer** is a 2D isometric dark fantasy game built using Go and the Ebiten engine. Set in a world of crumbling castles, haunted forests, and cursed ruins, the player must navigate procedurally generated dungeons filled with monsters, secrets, and mystery.

This project is a technical and artistic love letter to games like *Diablo* and *Stardew Valleys* dungeon sections  with a focus on real-time combat, tile-based navigation, and rich pixel art.

---

##  Developer Notes: Phase 1 Nearing Completion

**Status:** Phase 1 is 95%+ complete with no active blockers. Core loop, combat, and FOV systems are production-ready.

_Remaining Phase 1 Tasks:_

- [ ] Complete full run state serialization (Remnants save working, full state TODO)
- [ ] Implement load game menu (cosmetic)
- [ ] Implement options menu (cosmetic)
- [ ] Investigate corner ray FOV edge case (visual polish)

---

##   Feature Checklist

| Feature Description                                  | Status   |
|------------------------------------------------------|----------|
| Isometric tile-based rendering engine                | ✅       |
| Smooth tile-based player movement (mouse + keyboard) | ✅       |
| Monster pathfinding & AI                             | ✅       |
| Real-time combat system                              | ✅       |
| Health bars, damage numbers, and hit markers         | ✅       |
| Monster death animations                             | ✅       |
| Field-of-view with raycasting                        | ✅       |
| Fog of war memory system                             | ✅       |
| Click-to-move A* pathfinding                         | ✅       |
| Corner-casting to eliminate visual blind spots       | ✅       |
| Full-bright debug toggle                             | ✅       |
| Interactive level editor                             | ✅       |
| Procedural forest level generator                    | ✅ (Mature) |
| Procedural dungeon generator                         | ✅ (Mature) |
| Developer test level                                 | ✅       |
| In-game monster spawning via hotkeys                 | ✅       |
| Particle and visual effects                          | ✅       |
| Save/load system (run-level)                         | ✅       |
| Audio engine (SFX + music)                           | ✅       |
| Custom UI system                                     | ✅       |
| Main menu with animated background                   | ✅       |
| Parallax and animated layers                         | ✅       |
| Game state handling (Game Over, Restart)             | ✅       |
| Inventory system                                     | ✅       |
| 6 spell types (Fireball, Chaos, Lightning ×2, Fractals ×2) | ✅ |

---

##  Known Issues / Bug Bounties

- [x] **Bug #1**: Spamming right-click causes player to stutter in place or revert to earlier tiles. ***(Fixed)*** — State machine sync in movement controller (OnStep fires at tile arrival).
- [x] **Bug #2**: Crash when player character dies without proper state handling. ***(Fixed)***
- [ ] **Bug #3**: Corner ray leaks light in some rare edge cases. ***(In Progress)*** — Needs epsilon tolerance refinement in ray-wall intersection.
- [x] **Bug #4**: UI input not responsive after Game Over screen. ***(Fixed)***
- [ ] [Performance] Corner-casting rays may impact performance with 500+ tiles.

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
| **A / ←**            | Move left                                                              |
| **D / →**            | Move right                                                             |
| **W / ↑**            | Move up                                                                |
| **S / ↓**            | Move down                                                              |
| **1**                | Cast Fireball spell                                                    |
| **2**                | Cast Chaos Ray spell                                                   |
| **3**                | Cast Lightning spell                                                   |
| **4**                | Cast Lightning Storm spell                                             |
| **5**                | Cast Fractal Bloom spell                                               |
| **6**                | Cast Fractal Canopy spell                                              |
| **Shift**            | Dash ability (short burst movement)                                    |
| **F**                | Grapple ability (interact/grab)                                        |
| **E**                | Interact with objects (doors, NPCs)                                    |
| **Tab**              | Open/close inventory                                                  |
| **H**                | Open hero panel (stats/progression)                                    |
| **ESC**              | Open pause menu                                                        |
| **F1**               | Open controls menu                                                     |
| **F10**              | Toggle HUD visibility                                                 |
| **Q**                | Generate random forest level *(debug)*                                |
| **R**                | Generate random dungeon level *(debug)*                               |
| **T**                | Load developer test level *(debug)*                                   |
| **Y**                | Toggle full-bright rendering (disables FOV) *(debug)*                 |

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


