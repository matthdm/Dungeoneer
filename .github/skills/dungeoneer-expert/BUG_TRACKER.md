# Dungeoneer Bug Tracker

**Last Updated**: 2026-03-22

This file tracks all active bugs, their status, attempted fixes, blockers, and code references. Updated by skill context sync and manual bug reports.

---

## Bug #1: Right-Click Movement Stutter (FIXED ✓)

**Status**: 🟢 Fixed / **Priority**: —

**Phase**: Was Phase 1 Blocker (RESOLVED)

**Description**: 
Previously: Spamming right-click (RMB) caused player stutter, jumping backwards, and position reversion. Bug is now **RESOLVED**.

**Root Cause**:
Movement state machine was not properly synchronized between input handler and pathfinding controller. Position conflicts occurred when rapid input queued new paths while old paths were still executing.

**Fix Applied**:
Refactored movement state machine in `src/movement/controller.movement.go` to properly synchronize player position between input handler and pathfinding queue. Player position is now atomically updated, preventing conflicts.

**Code References**:
- 📍 `src/entities/player_io.go` (RMB handler — FIXED)
- 📍 `src/movement/controller.movement.go` (pathfinding state machine — FIXED)
- 📍 `src/entities/player.go` (position sync — FIXED)

**Solution Summary**:
- Implemented atomic position updates in movement controller
- Added proper queue prioritization (new paths cancel pending paths)
- Validated player position consistency on each movement step

**Attempts & Resolution**:

### Attempt 1: Input Debouncing ❌
Added rate-limiting to RMB → failed, input not the issue

### Attempt 2: Pathfinding Queue Flush ❌
Cleared queue before new path → made movement jerky

### Attempt 3: State Machine Refactor ✓
Rewrote movement controller state machine with atomic updates → **SUCCESSFUL**

**Tested**: ✓ Rapid RMB spam (100+ clicks/sec) no longer causes stutter or reversion

---

## Bug #2: Crash on Player Death (FIXED ✓)

**Status**: 🟢 Fixed / **Priority**: —

**Description**: 
Crash when player character dies without proper state handling. Game Over screen didn't display correctly.

**Code**: 
📍 `src/game/deathscreen.go`

**Fix Applied**:
Added proper state transition in `game.go` and death UI handling in `deathscreen.go`. Now runs death screen → calculates Remnants → returns to main menu.

**Tested**: ✓ (Death screen works; need full retest with save/load)

---

## Bug #3: Corner Ray Light Leak

**Status**: 🔴 Open / **Priority**: MEDIUM

**Description**:
In FOV system with corner-casting, certain thin north-wall positions leak light through walls. Player can see through walls in specific geometric configurations.

**Reproduction**:
1. Load any level (Q for forest, R for dungeon, T for test)
2. Position player near thin north-wall (vertical wall segment)
3. Observe: Light bleeds through wall edges in corner-casting raycaster

**Root Cause** (suspected):
- Ray-wall intersection logic in corner-casting not handling thin wall segments correctly
- Edge case in raycasting when wall segment width is minimal

**Code References**:
- 📍 `src/fov/walls.go` (line ~85-100, ray-wall intersection)
- 📍 `src/fov/fov.go` (corner-casting raycaster)

**Previous Attempts**:
- ❌ Attempt marked "Fixed" in prior version — light leak persists

**Blocker**: 
Need detailed geometric tracing of ray-wall intersection for thin north-wall segments. May require edge case handling or epsilon tolerance adjustment.

**Next Steps**:
1. Add debug visualization of raycasts to identify exact light leak points
2. Trace ray-wall intersection math for north-wall edge cases
3. Test with various wall thickness values
4. Implement epsilon tolerance or edge-case detection

---

## Bug #4: UI Unresponsive After Game Over (FIXED ✓)

**Status**: 🟢 Fixed / **Priority**: —

**Description**:
After Game Over screen, input was not responsive; menu navigation broken.

**Fix Applied**:
Fixed hotkey binding logic in `ui/` input handler.

**Tested**: ✓

---

## Performance: Corner-Casting Ray Cost

**Status**: 🟡 Open / **Priority**: MEDIUM

**Description**:
Corner-casting raycasting rays may cause frame drops with high tile density (large dungeons or dense forests). Performance degrades as more rays are cast.

**Code**: 
📍 `src/fov/fov.go` (raycasting loop, line ~110-150)
📍 `src/fov/render.go` (rendering integration)

**Suspected Cause**:
- Ray count grows with map size / density
- No spatial indexing or ray caching
- Every frame re-casts all rays (no delta rendering)

**Possible Solutions** (ideate):
1. Spatial grid to cull ray checks
2. Cache ray results if player hasn't moved
3. Reduce ray count for lower performance targets
4. Multi-threaded ray casting

**Next Steps**:
Profiling needed. Measure frame time vs. map density at different player positions.

---

## Tracking Template

For new bugs, fill out:

```markdown
## Bug #N: [Title]

**Status**: 🔴 Open / 🟡 In Progress / 🔴 Blocked / 🟢 Fixed

**Priority**: HIGH / MEDIUM / LOW

**Description**: 
[What is broken?]

**Reproduction**:
[Step-by-step to trigger]

**Root Cause** (suspected):
[Your hypothesis]

**Code References**:
- 📍 `src/path/file.go` (function, lines X-Y)

**Attempts & Findings**:

### Attempt N: [Approach]
**What**: [What did you try?]
**Reason**: [Why did you think it would work?]
**Result**: ❌ Failed / ✓ Success
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
