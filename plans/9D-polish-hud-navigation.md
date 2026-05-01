---
plan-id: 9D-polish-hud-navigation
status: queued
owner: unassigned
branch: plan/9D-polish-hud-navigation
depends-on: [9A-polish-transitions-camera]
last-touched: 2026-04-30
---

# Plan: Phase 9D — HUD & Navigation (Minimap, Floor Indicator, Status Icons)

## Goal

Add three HUD improvements that meaningfully reduce navigation friction and combat information overload: a minimap overlay showing explored rooms and exit direction, a persistent floor/biome indicator, and status effect icons with duration. This plan can run in parallel with `9B` and `9C` since it touches only HUD rendering code.

## Scope

**In scope:**
- Minimap: small corner overlay, explored rooms as rectangles, player as a dot, exit room highlighted, toggle with M key.
- Floor indicator: "Floor 2 — Ashveil Crypts" persistent HUD element top-center.
- Status effect icons: row of active buff/debuff icons with remaining duration bar beneath each.

**Out of scope (do not change in this plan):**
- Full map screen (minimap only).
- Pathfinding hints or objective markers.
- Inventory or equipment panels (already exist).

## File envelope

**Touched:**
- `src/hud/minimap.go` *(new)* — minimap rendering, room rect generation, explored tracking
- `src/hud/hud.go` — integrate minimap, floor indicator, status icon row
- `src/game/draw.game.go` — draw HUD overlay after world render
- `src/game/game.go` — track explored rooms, toggle minimap visibility
- `src/game/handlers.game.go` — M key toggles minimap
- `design-docs/roadmap.md` — mark Phase 9 HUD rows ✅ on completion
- `CLAUDE.md` — update Phase 9 status block to "🟢 COMPLETE"
- `plans/_QUEUE.md` — move to Completed on finish

**Forbidden (do not modify in this plan):**
- `src/coords/`, `src/collision/` — offset plan
- `src/levels/generate64.go` — do not change generation; read room data from existing floor context
- `src/entities/` — no entity logic changes

## Acceptance criteria

- [ ] Minimap renders in the top-right corner (120×80px). Explored rooms appear as dark gray rectangles; current room has a lighter fill; exit room has a green border; player dot is white.
- [ ] M key toggles minimap visibility (persists within a run, resets to visible on new run).
- [ ] Rooms are marked as explored on first entry; unexplored rooms are not shown.
- [ ] Floor indicator shows "Floor N — [Biome Name]" top-center, semi-transparent background, always visible during a run.
- [ ] Status effect icons (up to 6) appear as a row below the spell bar. Each icon has a small duration bar that depletes. Hovering an icon shows the effect name (or press Tab to toggle icon labels).
- [ ] `cd src && go build ./...` passes.

## Phases

### Phase 1: Floor indicator
- [ ] 1.1 In `src/hud/hud.go`, add `DrawFloorIndicator(screen *ebiten.Image, floor int, biomeName string)` — renders "Floor N — [Biome]" centered at top with a semi-transparent dark panel behind it.
- [ ] 1.2 Call from the main HUD draw when in a dungeon floor (not hub).
- [ ] 1.3 `cd src && go build ./...` passes.

### Phase 2: Status effect icons
- [ ] 2.1 In `src/hud/hud.go`, add `DrawStatusIcons(screen *ebiten.Image, effects []ActiveEffect)` — renders up to 6 icons in a horizontal row. Each is a 24×24 colored square (or sprite if exists) with a duration progress bar below.
- [ ] 2.2 Confirm `ActiveEffect` has `ID string`, `Duration float64`, `MaxDuration float64` — add if missing from `entities/effects.go`.
- [ ] 2.3 Use simple colored rectangles per effect type if no icon sprites exist (document in "What was NOT changed").
- [ ] 2.4 `cd src && go build ./...` passes.

### Phase 3: Minimap
- [ ] 3.1 Create `src/hud/minimap.go`. `Minimap` struct: `ExploredRooms map[RoomID]bool`, `AllRooms []RoomRect`, `ExitRoomID RoomID`, `Visible bool`.
- [ ] 3.2 `RoomRect` is an axis-aligned rectangle in minimap space (scaled from world tile coords to minimap pixel coords).
- [ ] 3.3 `MarkExplored(roomID RoomID)` called on room entry.
- [ ] 3.4 `Draw(screen *ebiten.Image, playerPos coords.WorldPos)` — renders the minimap panel in the top-right corner.
- [ ] 3.5 In `src/game/game.go`, store `Minimap` on `Game`; rebuild `AllRooms` on floor generation; call `MarkExplored` on room entry.
- [ ] 3.6 In `src/game/handlers.game.go`, M key toggles `minimap.Visible`.
- [ ] 3.7 `cd src && go build ./...` passes.

### Phase 4: Cleanup
- [ ] 4.1 Mark Phase 9 HUD roadmap rows ✅ in `design-docs/roadmap.md`.
- [ ] 4.2 Update `CLAUDE.md` Phase 9 status block to "🟢 COMPLETE".
- [ ] 4.3 Move this plan to `plans/COMPLETED/9D-polish-hud-navigation.md`.
- [ ] 4.4 Update `plans/_QUEUE.md`.

## Progress log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-04-30 | — | Drafted | Plan written; not yet started |

## What was NOT changed (intentional)

_None yet._

## Open questions

- **Room ID type.** The room data structure in `levels/generate64.go` may not have a stable ID per room. If not, use the room's grid index or center tile coordinates as a map key. Confirm before starting Phase 3. Affects 3.1.
- **Effect icon sprites.** If `entities/effects.go` does not have a sprite ID per effect type, use a color-coded square (poison=green, burn=orange, slow=blue, shield=gold). Affects 2.3.
