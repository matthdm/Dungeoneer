# Dungeoneer Bug Tracker

**Last Updated**: 2026-03-23

This file tracks all active bugs, their status, attempted fixes, blockers, and code references. Updated by skill context sync and manual bug reports.

---

## Bug #1: Right-Click Movement Stutter (FIXED)

**Status**: Fixed / **Priority**: —

**Phase**: Was Phase 1 Blocker (RESOLVED)

**Description**:
Spamming right-click (RMB) caused player stutter, jumping backwards, and position reversion.

**Root Cause**:
Movement state machine was not properly synchronized between input handler and pathfinding controller. Position conflicts occurred when rapid input queued new paths while old paths were still executing.

**Fix Applied**:
Refactored movement state machine in `src/movement/controller.movement.go` to properly synchronize player position. OnStep fires at tile arrival, preventing conflicts.

**Code References**:
- `src/entities/player_io.go` (RMB handler)
- `src/movement/controller.movement.go` (pathfinding state machine)
- `src/entities/player.go` (position sync)

---

## Bug #2: Crash on Player Death (FIXED)

**Status**: Fixed / **Priority**: —

**Description**:
Crash when player character dies without proper state handling.

**Fix Applied**:
Added proper state transition in `game.go` and death UI handling in `deathscreen.go`.

---

## Bug #3: Corner Ray Light Leak

**Status**: Open / **Priority**: MEDIUM

**Description**:
In FOV system with corner-casting, certain thin north-wall positions leak light through walls. Player can see through walls in specific geometric configurations.

**Reproduction**:
1. Load any level (Q for forest, R for dungeon, T for test)
2. Position player near thin north-wall (vertical wall segment)
3. Observe: Light bleeds through wall edges in corner-casting raycaster

**Root Cause** (suspected):
- Ray-wall intersection logic in corner-casting not handling thin wall segments correctly
- Needs epsilon tolerance refinement in ray-wall intersection

**Code References**:
- `src/fov/walls.go` (line ~85-100, ray-wall intersection)
- `src/fov/fov.go` (corner-casting raycaster)

**Next Steps**:
1. Add debug visualization of raycasts to identify exact light leak points
2. Trace ray-wall intersection math for north-wall edge cases
3. Implement epsilon tolerance or edge-case detection

---

## Bug #4: UI Unresponsive After Game Over (FIXED)

**Status**: Fixed / **Priority**: —

**Description**:
After Game Over screen, input was not responsive; menu navigation broken.

**Fix Applied**:
Fixed hotkey binding logic in `ui/` input handler.

---

## Bug #5: Monsters Visible Through Walls (FIXED)

**Status**: Fixed / **Priority**: —

**Phase**: Phase 2

**Description**:
Monsters on previously-seen (fog-of-war) tiles were still rendered even when behind walls, allowing wall peeking.

**Root Cause**:
`render_collect.game.go` rendered monsters on any tile in `SeenTiles`, not just actively visible tiles.

**Fix Applied**:
Changed monster rendering to only show monsters in active FOV (`isTileVisible`), not fog-of-war tiles.

**Code References**:
- `src/game/render_collect.game.go` (monster visibility check)

---

## Bug #6: Boss Could Spawn Inside a Wall (FIXED)

**Status**: Fixed / **Priority**: —

**Phase**: Phase 2

**Description**:
Boss entity spawned at room center, which wasn't guaranteed walkable after dead-end pruning in level generation.

**Root Cause**:
Room center tile could be converted to a wall during post-generation cleanup passes.

**Fix Applied**:
Added BFS walkability check (`findWalkableNear`) that spirals outward from room center to find a guaranteed walkable tile for boss placement.

**Code References**:
- `src/game/boss.game.go` (`setupBossFloor`, `findWalkableNear`)

---

## Bug #7: Player Items Persisted After Run Death/Victory (FIXED)

**Status**: Fixed / **Priority**: —

**Phase**: Phase 2

**Description**:
Player retained all items, equipment, stats, gold, and levels after a run ended (death or victory). Should be fully reset per design.

**Root Cause**:
`resetPlayerForHub()` only reset HP and Mana, not inventory, equipment, stats, or gold.

**Fix Applied**:
Expanded `resetPlayerForHub()` to perform full death reset: Level, EXP, UnspentPoints, Gold, Stats, TempModifiers, Inventory, Equipment, Effects all reset to defaults. `RecalculateStats()` called after reset.

**Code References**:
- `src/game/hub.go` (`resetPlayerForHub`)

---

## Bug #8: Monster AI Followed Breadcrumbs to Old Position (FIXED)

**Status**: Fixed / **Priority**: —

**Phase**: Phase 2

**Description**:
Monsters would pathfind to the player's position when they were first spotted, then continue walking there even after the player moved. Only recalculated path when the current path was empty or blocked.

**Root Cause**:
`BasicChaseLogic` in `monster.go` only recalculated paths when `len(m.Path) == 0` or the next tile was unwalkable. No check for whether the player had moved significantly.

**Fix Applied**:
Added `PathTargetX/PathTargetY` tracking to Monster. Path recalculation now triggers when player moves more than 2 tiles from the path target. Recalc cooldown reduced from 30 to 15 ticks.

**Code References**:
- `src/entities/monster.go` (`BasicChaseLogic`, `PathTargetX/PathTargetY`)

---

## Bug #9: Exit Always Reachable Without Opening Doors (FIXED)

**Status**: Fixed / **Priority**: —

**Phase**: Phase 2

**Description**:
The exit portal always spawned in the same walkable region as the player spawn, meaning the player never had to open any doors to reach the exit. This skipped most encounters and loot behind doors.

**Root Cause**:
BFS in `FindSpawnAndExit` used `IsWalkable`, which treats closed/locked doors as walls. The exit was always placed in the door-free region reachable from spawn.

**Fix Applied**:
- Created `IsPassable()` method on Level that treats doors as traversable
- Refactored `bfsFarthest` and `bfsDistMap` to accept a passable function parameter
- `FindSpawnAndExit` now uses `IsPassable` for BFS (crosses doors) but filters exit candidates with `IsWalkable` (so neither spawn nor exit lands on a door tile)
- Added spawn walkability safety check via `nearestWalkable`

**Code References**:
- `src/levels/level.go` (`IsPassable`)
- `src/levels/pathutil.go` (`FindSpawnAndExit`, `bfsFarthest`, `bfsDistMap`)

---

## Performance: Corner-Casting Ray Cost

**Status**: Open / **Priority**: MEDIUM

**Description**:
Corner-casting raycasting rays may cause frame drops with high tile density (large dungeons or dense forests). Performance degrades as more rays are cast.

**Code**:
- `src/fov/fov.go` (raycasting loop, line ~110-150)
- `src/fov/render.go` (rendering integration)

**Next Steps**:
Profiling needed. Measure frame time vs. map density at different player positions.

---

## Tracking Template

For new bugs, fill out:

```markdown
## Bug #N: [Title]

**Status**: Open / In Progress / Blocked / Fixed

**Priority**: HIGH / MEDIUM / LOW

**Description**:
[What is broken?]

**Reproduction**:
[Step-by-step to trigger]

**Root Cause** (suspected):
[Your hypothesis]

**Code References**:
- `src/path/file.go` (function, lines X-Y)

**Attempts & Findings**:

### Attempt N: [Approach]
**What**: [What did you try?]
**Reason**: [Why did you think it would work?]
**Result**: Failed / Success
**Lesson**: [What did you learn?]

**Blocker**:
[What's stopping progress? Test environment? Unknown root cause?]

**Next Steps**:
1. [Action]
2. [Action]
```

---

## Sync Notes

This file is auto-updated by `/dungeoneer sync`:
- Status changes reflected in README.md Known Issues
- Bug count in PROJECT_SNAPSHOT.json
- Blocker analysis informs roadmap phase readiness
