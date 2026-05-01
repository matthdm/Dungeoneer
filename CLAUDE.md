# CLAUDE.md — Agent Entry Point

> Read this first. It is intentionally short and points you at the right doc for the task at hand instead of the whole codebase. Update it when the project's current focus changes.

## What this is

Dungeoneer is a 2D isometric dark-fantasy roguelike in Go on Ebiten v2.8. Real-time tile combat, procedurally generated floors, item-gated abilities, and NPCs that ascend into the run's final boss. Single-player, offline. Source lives in `src/` (~24K lines of Go); the `dungeoneer` Go module sits there too.

## Current state — last updated 2026-04-27

- **Phases 1, 2, 3: complete.** Run loop, combat depth (6 enemy roles, status effects, multi-phase boss), NPCs & dialogue (room-tag placement, branching JSON trees, Varn arc through phase 2 + boss).
- **Phase 4: in stabilization.**
  - 4A (ability gating): shipped. `ItemTemplate.GrantsAbility/AbilitySlot`, `Player.RefreshAbilities`, mana costs, dynamic spell bar, dash/grapple gated.
  - 4B (ability item templates): shipped. 13 ability-granting items wired in `items/load.go`, biome-themed loot bias, Varn quest items injected by `progression.game.go`.
  - 4C (equipment & stats): mostly shipped. Stat modifiers, gold economy, item quality tiers, `give_item` dialogue action.
  - 4F (loot refinement, chest variants, treasure-room loot): open.
  - **Active stabilization work: the coordinate unification refactor.** See `OFFSET_UNIFICATION_PLAN.md` — phases 0–2 done, 3–5 partial.
- **Phases 5–10: not started.** Trust system, NPC phase tracker, hub shop/upgrades, full meta save, NG+ memory, echoes, living dungeon AI, additional major NPCs, Abaddon.

## Where to look (by task type)

| Task type | Read this first |
|---|---|
| Coordinate, hit detection, render offset | `OFFSET_UNIFICATION_PLAN.md`, `src/coords/worldpos.go` |
| Add or modify an ability/item | `design-docs/ability-items.md`, `src/items/load.go`, `src/entities/player.go` (`RefreshAbilities`, `EquipStarter`) |
| NPC or dialogue work | `design-docs/dialogue-system.md`, `src/dialogue/`, `src/dialogues/*.json`, `src/game/npc.game.go`, `src/game/npc_data.go` |
| Boss / Varn arc | `design-docs/boss-system.md`, `src/entities/varn_boss.go`, `src/game/boss.game.go` |
| Level generation / biomes / room tags | `design-docs/biome-system.md`, `design-docs/room-tagging.md`, `src/levels/generate64.go`, `src/levels/room_tagger.go`, `src/game/biome.go` |
| Pathing / FOV / movement | `src/pathing/astar.go`, `src/fov/`, `src/movement/controller.movement.go`, `src/collision/box.collision.go` |
| Phase planning, what's next | `design-docs/roadmap.md` |
| Test scenarios / acceptance | `design-docs/test-cases.md` |

## How agents pick up work

Work is chunked into **plans** under `plans/`. One plan = one branch = one bounded scope. An agent never needs to load the whole repo — the plan tells them what to read, what to touch, and what to leave alone.

To start a session: read this file, then `plans/_QUEUE.md`, then the **Active plan** named there. Follow `plans/_PROTOCOL.md` for the start-and-end ritual. Build must pass on every commit.

When the active plan is empty, promote the top-priority queued plan. When a session's work is incomplete, leave the unchecked boxes and append to the plan's progress log so the next agent resumes cleanly.

## Working contract

- **Engineering principles:** `.github/agents/instructions.md`. Performance-first, surgical edits, no per-frame allocations in hot paths, idiomatic Go.
- **Bound your scope.** Each plan has a **File envelope: Touched / Forbidden**. Do not modify forbidden files even if it would be convenient. Out-of-envelope changes go in **Open questions**, not the diff.
- **`OFFSET_UNIFICATION_PLAN.md` is the prototype** for the plan format and predates the `plans/` directory. It is followed under the same conventions and treated as a concurrent in-flight plan.
- **End every session by updating the active plan doc.** The next agent reads it cold; if your reasoning isn't in the progress log or "What was NOT changed (intentional)" section, the context is gone.
- **No tests exist yet.** Adding `_test.go` for pure-logic packages (`coords`, `pathing`, `levels/room_tagger`, `entities/effects`) is the highest-leverage thing you can do for AI maintainability — it's the only feedback loop available without launching the game.

## Coordinate invariant (do not violate)

Combat checks, hit detection, spell origins, and effect anchoring use `coords.WorldPos.BodyCenter()`. Never raw `TileX/TileY` (lags during interpolation) and never raw `InterpX/InterpY` (skips body-center offset). The "Golden rule" doc comment at the top of `src/coords/worldpos.go` is authoritative; the offset plan exists because parts of the codebase predate it and are still being migrated.

## Build & run

- `cd src && go build ./...` must pass before any commit.
- `src/build_and_run.bat` builds and launches on Windows.
- Tests: none yet. Add them as you change pure-logic code.

## Files to keep current

- This file (`CLAUDE.md`) — current phase status and the "where to look" table.
- `README.md` — public-facing status, feature checklist, bug bounties.
- `design-docs/roadmap.md` — phase planning and `Last updated` date.
- `plans/_QUEUE.md` — active and queued plans. Always reflects reality.
- `OFFSET_UNIFICATION_PLAN.md` — progress log for the in-flight refactor.

If those agree, an agent starting cold can be productive in one session.
