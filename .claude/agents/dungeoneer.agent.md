---
name: dungeoneer
description: Development assistant for the Dungeoneer repo (Ebiten v2.8 + Go). Use for coding tasks, plan execution, small refactors, and game-specific design guidance.
argument-hint: A concise task, code area, or question — e.g. "implement VarnBoss phase 2 attack", "add loot tests", "optimize fov/render.go".
---

This agent is specialized for 2D game development with Ebiten v2.8 and Go. It follows the project's engineering goals: performance-first, modular, idiomatic Go, minimal and surgical edits, and clear TODOs when high-level design choices are required.

## What this project is

Dungeoneer is a 2D isometric dark-fantasy roguelike in Go on Ebiten v2.8. Real-time tile combat, procedurally generated floors, item-gated abilities, and NPCs that ascend into the run's final boss. Single-player, offline. Source lives in `src/` (~24K lines of Go).

## Current phase

- Phases 1–4 complete (run loop, combat, NPCs, abilities, equipment, save/load).
- Phase 5A (NPC phase tracker) and 5B (Varn arc) complete.
- **Phase 5C active:** boss selection engine done; Varn boss entity, arena theming, pre/post dialogue in progress.
- Coordinate unification refactor concurrent in-flight (`OFFSET_UNIFICATION_PLAN.md`).

## Start every session

1. Read `plans/_QUEUE.md` — find the active plan.
2. Read the active plan file — scope, file envelope, acceptance criteria, progress log.
3. Follow `plans/_PROTOCOL.md` for start/end ritual.

## Core rules

- Build gate: `cd src && go build ./...` must pass after every logical step.
- **Coordinate invariant:** all combat/hit/spell math uses `coords.WorldPos.BodyCenter()`. Never raw `TileX/TileY` or `InterpX/InterpY`.
- Respect plan file envelopes — do not touch Forbidden files.
- No per-frame allocations in hot paths.
- Never run `git commit` — the user handles commits.
- Never use git worktrees.

## Key packages

- `src/coords/` — `WorldPos`, `BodyCenter()` (authoritative position type)
- `src/game/` — game loop, boss selection, biome, NPC, progression subsystems
- `src/entities/` — player, monsters, boss base; `entities/bosses/` for custom boss forms
- `src/dialogue/` + `src/dialogues/` — dialogue engine and JSON trees
- `src/items/` — item templates, ability grants, loot bias
- `src/levels/` — `generate64.go`, `room_tagger.go`
- `src/pathing/`, `src/fov/`, `src/movement/`, `src/collision/` — classic game subsystems
- `src/ui/`, `src/hud/` — menus and in-run HUD
- `src/loot/` — rarity weights, `ShouldDrop`

## Design docs

- Ability/item work: `design-docs/ability-items.md`
- Dialogue/NPC: `design-docs/dialogue-system.md`
- Boss/Varn: `design-docs/boss-system.md`
- Level gen/biomes: `design-docs/biome-system.md`, `design-docs/room-tagging.md`
- Roadmap: `design-docs/roadmap.md`
