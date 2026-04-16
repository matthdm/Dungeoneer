# Dungeoneer: Offset Unification Plan

> **Purpose**: This document captures the full analysis of visual/logical offset
> inconsistencies in the Dungeoneer codebase, and a phased plan to unify them.
> If a conversation is interrupted mid-effort, any future session can pick up
> from wherever the checkboxes left off.

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [How Coordinates Work Today](#how-coordinates-work-today)
3. [Complete Offset Inventory](#complete-offset-inventory)
4. [The Five Root Divergences](#the-five-root-divergences)
5. [Phased Fix Plan](#phased-fix-plan)
6. [File-by-File Change Map](#file-by-file-change-map)
7. [Testing Checklist](#testing-checklist)

---

## Problem Statement

The game has two coordinate representations that evolved independently:

| Representation | Used By | Type |
|----------------|---------|------|
| `TileX, TileY` (int) | Melee combat, AI range checks, projectile hits, pathing | Discrete grid |
| `InterpX, InterpY` (float64) | Rendering, spell hit detection, collision, smooth movement | Continuous world |

These are kept in sync by callbacks and floor-casts, but **combat subsystems
pick whichever representation was convenient at the time they were written**.
The result: a slash spell can hit a monster that a melee swing cannot (and vice
versa), projectiles snap to tile centers while the player is visually between
tiles, and hardcoded pixel offsets litter the rendering code to paper over the
mismatch.

---

## How Coordinates Work Today

### World Space (Cartesian)

Entities live in a cartesian grid. A position like `(3.5, 7.2)` means "halfway
across tile column 3, 20% into tile row 7."

- **Player**: `MoveController.InterpX/InterpY` (float64), `TileX/TileY` (int)
- **Monster**: `InterpX/InterpY` (float64), `TileX/TileY` (int)
- **NPC**: same pattern as Monster

`TileX/TileY` is updated when movement completes (arrival callback), so during
interpolation it lags behind `InterpX/InterpY`.

### Isometric Projection

```
isoX = (cartX - cartY) * (TileSize / 2)   // TileSize = 64 -> * 32
isoY = (cartX + cartY) * (TileSize / 4)   // TileSize = 64 -> * 16
```

Defined in two places (should be one):
- `entities/helper.go:7-11` — `isoToScreenFloat()`
- `game/game.go:480-485` — `cartesianToIso()`

### Camera Transform

Standard chain applied to every drawable:
```
GeoM.Translate(isoX, isoY)       // world position
GeoM.Translate(-camX, camY)      // camera offset (note: Y is NOT negated)
GeoM.Scale(camScale, camScale)   // zoom
GeoM.Translate(cx, cy)           // screen center
```

Camera tracks player via lerp in `game/game.go:683-703`.

### Body Center Convention

Entity positions anchor at the **top-left vertex of the isometric diamond**.
The visual body center is offset by:
- `constants.IsoBodyDX = 1.0` (X, horizontal)
- `MonsterHitCenterYOffset = 0.30` (Y, vertical depth)

`BodyX()/BodyY()` methods exist on Player and Monster but are only used by
spell detection code, not melee or projectile code.

---

## Complete Offset Inventory

### Collision Offsets
| Constant | Value | File | Line | Purpose |
|----------|-------|------|------|---------|
| `YWallVisualOffset` | 0.75 | `collision/box.collision.go` | 56 | Shift collision box down to align visual sprite base with tile walkability |
| `XWallVisualOffset` | 0.21 | `collision/box.collision.go` | 57 | Shift collision box right for same reason |
| `CollisionBox.Width` | 0.55 | `entities/player.go` | 144 | Player collision box width |
| `CollisionBox.Height` | 0.80 | `entities/player.go` | 144 | Player collision box height |

### Body Center Offsets
| Constant | Value | File | Line | Purpose |
|----------|-------|------|------|---------|
| `IsoBodyDX` | 1.0 | `constants/constants.go` | 64 | X offset from tile anchor to body center |
| `MonsterHitCenterYOffset` | 0.30 | `entities/monster.go` | 76 | Y offset from feet to body center |

### Rendering Offsets
| Offset | Value | File | Line | Purpose |
|--------|-------|------|------|---------|
| Player vertical | `-1.0 + BobOffset` | `entities/player.go` | 199 | Shift sprite up 1 unit + breathing bob |
| Monster vertical | `-1.0 + BobOffset` | `entities/monster.go` | 284 | Same for monsters |
| NPC vertical | `-1.0 + BobOffset` | (render_collect) | ~325 | Same for NPCs |
| Exit bob | `sin(tick/30) * 2.0` | `entities/exit_entity.go` | 32 | Floating exit portal |
| Player health bar | `sx - barW/2 + 30, sy` | `entities/player.go` | 242 | Hardcoded +30px horizontal nudge |
| Monster health bar | `x - barW/2 + 35, y - 10` | `entities/monster.go` | 320 | Hardcoded +35px, -10px nudge |
| Hit marker | `+35, +15` | `game/hitmarker.game.go` | 26-27 | Hardcoded pixel offsets |
| Projectile draw | `+tileSize/2, +tileSize/4` | `entities/monster_projectile.go` | 84 | Centers dot on tile diamond |
| Item drop scale | `0.5x, offset -w/4, ts/4-h/4` | `game/render_collect.game.go` | ~111 | Centers shrunken icon on tile |
| Item sparkle | `-2, ts/4-h/4-4` | `game/render_collect.game.go` | ~132 | Sparkle slightly left and up from item |

### Combat Range Values
| Check | Coordinates Used | File | Line |
|-------|-----------------|------|------|
| Melee `IsAdjacent` | `TileX, TileY` (int) | `entities/monster.go` | 256 |
| Ranged distance | `TileX, TileY` (int, cast to float for sqrt) | `entities/ranged.monster.go` | ~29-72 |
| Caster distance | `TileX, TileY` (int, cast to float for sqrt) | `entities/caster.monster.go` | ~39-83 |
| Slash arc `IsInArc` | `InterpX, InterpY` (float) | `spells/slash.spells.go` | ~109-124 |
| Fireball radius | `InterpX, InterpY` (float) | `game/spells.game.go` | ~115-132 |
| Projectile hit | `p.X/Y` (float) vs `px, py` (int) | `entities/monster_projectile.go` | 66-70 |
| NPC interaction | `TileX/TileY` or `InterpX/InterpY` depending on config | `entities/npc.go` | ~82-96 |
| Collision sweep | `CollisionBox.X/Y` + `YWallVisualOffset/XWallVisualOffset` | `collision/box.collision.go` | 59-63 |

---

## The Five Root Divergences

### 1. Collision box uses hardcoded visual fudge factors

`collision/box.collision.go:56-63` applies `YWallVisualOffset = 0.75` and
`XWallVisualOffset = 0.21` to shift the collision box so the sprite *looks*
like it stops at walls. These aren't derived from sprite dimensions or anchor
points — they're eyeballed. Changing sprites, zoom, or anchor conventions will
silently break collision alignment.

### 2. Melee uses int tiles while spells use floats

`IsAdjacent(m.TileX, m.TileY, player.TileX, player.TileY)` is a discrete
check: either you're on a neighboring tile or you're not. But
`slash.IsInArc(m.InterpX, m.InterpY)` is continuous: the arc can clip an
entity that's 0.3 tiles into the neighboring cell. During movement
interpolation, `TileX/TileY` lags behind `InterpX/InterpY`, so a monster
mid-step between tiles (2, 3) and (3, 3) has `TileX=2` but `InterpX=2.7`.
Melee sees it at tile 2; spells see it at 2.7.

### 3. Projectile hit detection mixes int and float

`HitsPlayer(px, py int)` in `monster_projectile.go:66-70` takes the player's
**integer tile** but compares against the projectile's **float position**. The
player could visually be at `InterpX = 4.8` (nearly tile 5) but the hit check
uses `TileX = 4`. The projectile's float position would need to be within 0.6
tiles of `(4, playerTileY)` to register, even though the player's sprite is
rendering almost a full tile away.

### 4. Body center offset is partially adopted

`BodyX()/BodyY()` adds `IsoBodyDX` and `MonsterHitCenterYOffset` to get the
"real" center of the entity. But only spell code uses it. Melee, projectile,
and AI range checks ignore body center entirely — they use raw
`TileX/TileY`. The health bar offsets (`+30`, `+35`, `-10`) are independent
eyeballed attempts to center on the sprite without using `BodyX/BodyY`.

### 5. Vertical render shift (-1.0) has no logical counterpart

Every entity's draw code applies `Translate(0, -1.0 + BobOffset)` to shift
the sprite up by 1 unit. This is purely visual — collision and combat checks
operate on the un-shifted `InterpY`. But the hit markers, health bars, and
damage numbers each independently try to compensate with their own hardcoded
pixel offsets (`+35`, `+15`, `+30`, `-10`), and they don't all agree.

---

## Phased Fix Plan

### Phase 0: Preparation (no behavior changes)
- [ ] **0.1** Create `src/coords/worldpos.go` with a `WorldPos` type
- [ ] **0.2** Add methods: `TileX() int`, `TileY() int`, `BodyCenter() WorldPos`, `ToIso(tileSize int) (float64, float64)`, `DistTo(other WorldPos) float64`
- [ ] **0.3** Add `RenderOffset() (dx, dy float64)` that returns the canonical `-1.0 + bob` shift
- [ ] **0.4** Add `HealthBarScreenPos(tileSize int, camX, camY, scale, cx, cy float64) (float64, float64)` that derives bar position from body center
- [ ] **0.5** Write unit tests for `WorldPos` methods (iso conversion, tile derivation, distance)

```go
// src/coords/worldpos.go
package coords

import "math"

const (
    BodyDX     = 1.0  // horizontal offset to body center
    BodyDY     = 0.30 // vertical offset to body center
)

type WorldPos struct {
    X, Y float64
}

func (w WorldPos) TileX() int { return int(math.Floor(w.X)) }
func (w WorldPos) TileY() int { return int(math.Floor(w.Y)) }

func (w WorldPos) BodyCenter() WorldPos {
    return WorldPos{X: w.X + BodyDX, Y: w.Y + BodyDY}
}

func (w WorldPos) ToIso(tileSize int) (float64, float64) {
    half := float64(tileSize / 2)
    quarter := float64(tileSize / 4)
    return (w.X - w.Y) * half, (w.X + w.Y) * quarter
}

func (w WorldPos) DistTo(other WorldPos) float64 {
    return math.Hypot(w.X-other.X, w.Y-other.Y)
}

// TileCenterIso returns the screen-space center of the tile diamond.
// Useful for anchoring health bars, labels, and effects.
func (w WorldPos) TileCenterIso(tileSize int) (float64, float64) {
    ix, iy := w.ToIso(tileSize)
    return ix + float64(tileSize)/2, iy + float64(tileSize)/4
}
```

### Phase 1: Adopt WorldPos in entities (minimal behavior change)
- [ ] **1.1** Add `Pos() WorldPos` method to Player (returns `WorldPos{MoveController.InterpX, MoveController.InterpY}`)
- [ ] **1.2** Add `Pos() WorldPos` method to Monster (returns `WorldPos{InterpX, InterpY}`)
- [ ] **1.3** Add `Pos() WorldPos` method to NPC
- [ ] **1.4** Refactor `BodyX()/BodyY()` on Player and Monster to delegate to `Pos().BodyCenter()`
- [ ] **1.5** Remove duplicate `isoToScreenFloat()` from `entities/helper.go`, use `WorldPos.ToIso()` instead
- [ ] **1.6** Remove duplicate `cartesianToIso()` from `game/game.go`, use `WorldPos.ToIso()` or a standalone function in `coords`
- [ ] **1.7** Consolidate `IsoBodyDX` and `MonsterHitCenterYOffset` into `coords.BodyDX` and `coords.BodyDY`

### Phase 2: Unify combat checks (behavior-changing, needs testing)
- [ ] **2.1** Change `MonsterProjectile.HitsPlayer(px, py int)` to `HitsPlayer(pos WorldPos)` — pass player's continuous `Pos()` instead of integer tile
- [ ] **2.2** Change `IsAdjacent()` melee check in `monster.CombatCheck()` to use `m.Pos().BodyCenter().DistTo(player.Pos().BodyCenter()) <= meleeRange` with a tunable `meleeRange` constant (start at ~1.5 to preserve current feel)
- [ ] **2.3** Change ranged/caster distance checks to use `m.Pos().DistTo(player.Pos())` instead of `float64(m.TileX - p.TileX)` math
- [ ] **2.4** Standardize all spell hit checks to use `target.Pos().BodyCenter()` (most already close)
- [ ] **2.5** Update `entities/helper.go` — `IsAdjacent` / `IsAdjacentRanged` either removed or rewritten to use `WorldPos`
- [ ] **2.6** Playtest: verify melee feels the same, ranged aggro distances unchanged, spell hit registration unchanged

### Phase 3: Derive collision offsets from sprite anchor (behavior-changing)
- [ ] **3.1** Define a `SpriteAnchor` struct: `{OffsetX, OffsetY float64}` per entity type, representing the offset from the WorldPos to the sprite's visual feet
- [ ] **3.2** Compute collision visual offsets from `SpriteAnchor` instead of hardcoded `0.75`/`0.21`
- [ ] **3.3** Remove `YWallVisualOffset` and `XWallVisualOffset` constants
- [ ] **3.4** Playtest: walk along all wall types, verify no clipping regressions

### Phase 4: Centralize render offsets (visual-only changes)
- [ ] **4.1** Create a `RenderPos(pos WorldPos, tileSize int, bob float64) (isoX, isoY float64)` function that applies the `-1.0 + bob` shift in iso space
- [ ] **4.2** Replace the scattered `Translate(0, -verticalOffset+m.BobOffset)` calls in Player.Draw, Monster.Draw, and render_collect with calls to `RenderPos`
- [ ] **4.3** Derive health bar position from `WorldPos.TileCenterIso()` — remove the hardcoded `+30`, `+35`, `-10` pixel nudges
- [ ] **4.4** Derive hit marker position from `WorldPos.TileCenterIso()` — remove `+35`, `+15` from `hitmarker.game.go`
- [ ] **4.5** Derive damage number position from `WorldPos` — stop using `float64(m.TileX)` in `TakeDamage()`
- [ ] **4.6** Derive projectile draw offset from `WorldPos.TileCenterIso()` — remove `+tileSize/2, +tileSize/4` from `monster_projectile.go:84`
- [ ] **4.7** Visual regression test: all entities, effects, bars, and numbers should appear in the same positions as before

### Phase 5: Clean up and document
- [ ] **5.1** Remove all orphaned offset constants
- [ ] **5.2** Add a comment block at the top of `coords/worldpos.go` explaining the coordinate system for future contributors
- [ ] **5.3** Update any level editor code that uses raw offsets
- [ ] **5.4** Final full playtest

---

## File-by-File Change Map

| File | Phase | What Changes |
|------|-------|-------------|
| `src/coords/worldpos.go` | 0 | **NEW** — WorldPos type, conversion methods |
| `src/coords/worldpos_test.go` | 0 | **NEW** — unit tests |
| `src/constants/constants.go` | 1 | Remove `IsoBodyDX`, point to `coords.BodyDX` |
| `src/entities/helper.go` | 1-2 | Remove `isoToScreenFloat`, rewrite `IsAdjacent` |
| `src/entities/player.go` | 1 | Add `Pos()`, refactor `BodyX/BodyY`, refactor `Draw` |
| `src/entities/monster.go` | 1-2 | Add `Pos()`, refactor `BodyX/BodyY`, refactor `CombatCheck`, refactor `Draw` |
| `src/entities/npc.go` | 1 | Add `Pos()` |
| `src/entities/monster_projectile.go` | 2, 4 | Change `HitsPlayer` signature, fix draw offset |
| `src/entities/ranged.monster.go` | 2 | Use `WorldPos.DistTo` for range checks |
| `src/entities/caster.monster.go` | 2 | Use `WorldPos.DistTo` for range checks |
| `src/collision/box.collision.go` | 3 | Replace hardcoded offsets with SpriteAnchor-derived values |
| `src/game/game.go` | 1 | Remove `cartesianToIso`, use `coords` package |
| `src/game/draw.game.go` | 4 | Update tile drawing to use coords functions |
| `src/game/render_collect.game.go` | 4 | Centralize vertical offset, use `RenderPos` |
| `src/game/hitmarker.game.go` | 4 | Derive positions from WorldPos, remove magic numbers |
| `src/game/spells.game.go` | 2 | Use `BodyCenter()` consistently |
| `src/game/npc.game.go` | 4 | Use `TileCenterIso` for label positioning |
| `src/spells/slash.spells.go` | 2 | Already uses body offset, verify consistency |
| `src/spells/blink.spells.go` | 2 | Replace raw `+1` with `BodyCenter()` |
| `src/leveleditor/*.go` | 5 | Audit for raw offset usage |

---

## Testing Checklist

After each phase, verify:

- [ ] Player walks smoothly, stops at walls correctly
- [ ] Melee attacks connect when visually adjacent
- [ ] Slash arc hits monsters that are visually inside the arc
- [ ] Fireball hits monsters it visually touches
- [ ] Projectiles hit the player when they visually overlap
- [ ] Health bars are centered above sprites
- [ ] Hit markers (red X) appear on the monster that was hit
- [ ] Damage numbers float up from the correct position
- [ ] Item drops are centered on their tile
- [ ] NPC interaction triggers at the expected distance
- [ ] Boss grapple chain renders correctly
- [ ] Dash doesn't clip through walls
- [ ] Grapple hook attaches at the correct wall tile
- [ ] Level editor entity placement still works
- [ ] No visual jitter at any zoom level

---

## Progress Log

_Update this section as work proceeds so the next session knows where you left off._

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-16 | Analysis | Complete | Full offset inventory and plan created |
| | Phase 0 | Not started | |
| | Phase 1 | Not started | |
| | Phase 2 | Not started | |
| | Phase 3 | Not started | |
| | Phase 4 | Not started | |
| | Phase 5 | Not started | |
