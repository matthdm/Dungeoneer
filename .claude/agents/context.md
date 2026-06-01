# Dungeoneer — Agent Context

Project: Dungeoneer — a 2D isometric dark-fantasy roguelike in Go on Ebiten v2.8.
Real-time tile combat, procedurally generated floors, item-gated abilities, NPCs that
ascend into the run's final boss. Single-player, offline.

**Source root:** `src/` (~24K lines of Go). The `dungeoneer` Go module lives there.
**Build:** `cd src && go build ./...` — must pass before every commit.
**Run on Windows:** `src/build_and_run.bat`

---

## Package map

| Package | Responsibility |
|---------|---------------|
| `game/` | Main game loop (`game.go`), rendering orchestration, entity updates, subsystem files (`boss.game.go`, `biome.go`, `boss_selection.go`, `npc.game.go`, `progression.game.go`) |
| `entities/` | Player, monsters, boss base, hit markers, inventory; `entities/bosses/` for custom boss types |
| `coords/` | `WorldPos` — the authoritative position type. `BodyCenter()` is the golden rule for all combat/hit/spell math |
| `collision/` | Box collision, sweep tests |
| `fov/` | Field-of-vision, fog-of-war, shadow rendering |
| `movement/` | Movement controller abstraction |
| `pathing/` | A* implementation |
| `levels/` | Tile map, `generate64.go` (64×64 procedural generator), `room_tagger.go` (room-tag placement) |
| `dialogue/` | Dialogue engine: loader, runner, types |
| `dialogues/` | JSON dialogue trees (`varn_phase1.json` … `varn_boss_post.json`, etc.) |
| `items/` | Item templates, `load.go`, ability-granting items, loot bias |
| `sprites/` and `images/` | Spritesheet and embedded assets |
| `ui/` | Menus (main, pause, load, options, controls) |
| `hud/` | In-run HUD: spell bar, health, mana, status icons |
| `loot/` | Loot tables, rarity weights, `ShouldDrop` |
| `progression/` | Quest flags, NPC phase tracker (`NPCPhaseTracker`), `NPCMeta` |

---

## Hard constraints

- **Tile size:** 64×64 pixels; tile grid 64×64 tiles per floor.
- **Coordinate invariant:** All combat checks, hit detection, spell origins, and effect anchoring use `coords.WorldPos.BodyCenter()`. Never raw `TileX/TileY` (lags during interpolation) and never raw `InterpX/InterpY` (skips body-center offset). See `src/coords/worldpos.go` golden-rule doc comment.
- **Tick rate:** Ebiten fixed 60 TPS — logic expects fixed-tick timing.
- **No networking** — offline single-player only.
- **No per-frame allocations** in hot paths (Update, Draw). Reuse buffers.

---

## Current phase (as of 2026-06-01)

- **Phases 1–4: complete.** Run loop, combat depth, NPC/dialogue, ability gating, equipment/stats, gold economy, loot refinement, chest variants, mid-run save/load (`RunSave` + `PlayerSnapshot`).
- **Phase 5A (NPC phase tracker): complete.** `NPCPhaseTracker`, floor-entry auto-advance, trust decay.
- **Phase 5B (Varn arc): complete.** `varn_phase3.json` ascension trigger, phase 3 spawn rule.
- **Phase 5C (Boss selection + Varn boss fight): in progress.** Boss selection engine done (`game/boss_selection.go`). Varn boss entity, arena theming, pre/post dialogue outstanding.
- **Coordinate unification refactor:** concurrent in-flight. Phases 0–2 done; 3–5 partial. See `OFFSET_UNIFICATION_PLAN.md`.
- **Phases 5D–11: queued.** See `plans/_QUEUE.md`.

---

## Where to look first

| Task | Entry point |
|------|-------------|
| Coordinate / hit detection | `OFFSET_UNIFICATION_PLAN.md`, `src/coords/worldpos.go` |
| Ability / item work | `design-docs/ability-items.md`, `src/items/load.go`, `src/entities/player.go` |
| NPC / dialogue | `design-docs/dialogue-system.md`, `src/dialogue/`, `src/dialogues/*.json` |
| Boss / Varn | `design-docs/boss-system.md`, `src/entities/varn_boss.go`, `src/game/boss.game.go`, `src/game/boss_selection.go` |
| Level gen / biomes / room tags | `design-docs/biome-system.md`, `src/levels/generate64.go`, `src/levels/room_tagger.go` |
| Pathing / FOV / movement | `src/pathing/astar.go`, `src/fov/`, `src/movement/controller.movement.go` |
| Save / load | `src/game/` (RunSave wiring), `src/ui/load_menu.go` |
| Phase planning | `design-docs/roadmap.md`, `plans/_QUEUE.md` |

---

## Plans system

Work is chunked into plan docs under `plans/`. The active plan is named in `plans/_QUEUE.md`.
Protocol: `plans/_PROTOCOL.md`. Each plan specifies a file envelope (Touched / Forbidden) — do not touch Forbidden files even when convenient.

---

## Notes for the agent

- Always favor minimal, surgical edits that preserve style and performance.
- Leave TODOs where behavior is unclear rather than guessing larger design choices.
- No tests exist yet — adding `_test.go` for pure-logic packages (`coords`, `pathing`, `levels/room_tagger`, `entities/effects`) is the highest-leverage improvement available.
- User handles all git commits. Do not run `git commit`.
- Do not use git worktrees. Work directly in `C:\Github\Dungeoneer\src` on main.
