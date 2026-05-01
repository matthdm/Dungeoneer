# Work Queue

Single source of truth for what's active, queued, blocked, and done. Update this whenever a plan changes status. Plans not in this file do not exist as far as agents are concerned.

---

## Active plan

The one plan currently being worked. There is only ever one active plan at a time.

> **Active:** `4F-loot-refinement.md`

---

## Queued (priority order)

Plans ready to be picked up next. Top of list = next active. When the active plan completes, promote #1 below.

| # | Plan | Size | Depends On | Notes |
|---|------|------|------------|-------|
| 1 | `load-game-menu.md` | Small | — | Wires the stubbed Load Game menu. Independent. |
| 2 | `options-menu.md` | Small | — | Wires the stubbed Options menu. Independent. |
| 3 | `runstate-serialization.md` | Medium | `load-game-menu` | Full mid-run save/load. Extends load screen. |
| 4 | `5A-npc-phase-tracker.md` | Medium | `4F-loot-refinement` | NPCPhaseTracker + trust system infrastructure. |
| 5 | `5B-varn-arc.md` | Medium | `5A` | Varn's 4-phase dialogue arc (mostly JSON content). |
| 6 | `5C-boss-selection.md` | Large | `5B` | Boss selection engine + Varn boss fight. |
| 7 | `5D-hub-npc-quarter.md` | Small | `5B` | Hub NPC positions + Varn hub dialogue. |
| 8 | `6A-full-meta-save.md` | Medium | `5C` | Extended MetaSave, milestone unlocks, toast system. |
| 9 | `6B-ng-plus-dialogue.md` | Medium | `6A` | Meta-flag conditions, Varn NG+ trees, betrayal memory. |
| 10 | `6C-lore-system.md` | Medium | `6B` | Lore registry, unlock action, lore library UI. |
| 11 | `7A-echoes-of-self.md` | Large | `6C` | Echo recorder, echo entities, ghost visuals, shrine UI. |
| 12 | `7B-living-dungeon-ai.md` | Large | `6C` | Behavior tracker, dungeon mood, adaptive gen params. |
| 13 | `8A-item-sets.md` | Medium | `7A`, `7B` | Set registry, set bonuses, tooltip + HUD UI. |
| 14 | `8B-hub-shop-upgrades.md` | Large | `8A` | Shop NPC, upgrade station, 8 starter upgrades. |
| 15 | `9A-polish-transitions-camera.md` | Small | `8B` | Fade transitions, screen shake, boss intro pan. |
| 16 | `9B-polish-visual-effects.md` | Medium | `9A` | Particle system, damage numbers, death fades. |
| 17 | `9C-polish-audio.md` | Medium | `9B` | Audio engine, SFX hookups, music, volume wiring. |
| 18 | `9D-polish-hud-navigation.md` | Small | `9A` | Minimap, floor indicator, status effect icons. |
| 19 | `10A-additional-npcs.md` | XL | `9D`, `9C` | Seris, Mira, Kael — full arcs + boss fights. |
| 20 | `10B-abaddon-meta-narrative.md` | Large | `10A` | Abaddon, alignment system, endgame dialogue. |
| 21 | `10C-world-expansion.md` | Large | `10B` | 4 new biomes, hub evolution, cross-NPC dialogue. |

### Backlog (schedule when triggered)

These are independent and can be slotted into any available position. See the roadmap's Backlog section for trigger conditions.

| ID | Plan | Trigger |
|----|------|---------|
| B.1 | `backlog-fov-dynamic-ray.md` | Before any map size increase beyond 64×64 |
| B.2 | `backlog-astar-heap.md` | When monster count or map size causes frame drops |

---

## Blocked

Plans that cannot start until a dependency clears. Each entry must name what it's waiting on.

_None._

---

## In-flight legacy plans

Plans that predate this system but follow the same conventions.

- [`../OFFSET_UNIFICATION_PLAN.md`](../OFFSET_UNIFICATION_PLAN.md) — coordinate system stabilization. Phases 0–2 done; 3–5 open. Treated as concurrent with the active plan because it touches a different file envelope (`src/coords/`, `src/collision/`, render code).

---

## Completed

Most recent first.

_None yet._

---

## How to update this file

- **Promote a queued plan to active:** edit the plan's frontmatter `status: active`, move the entry from Queued to Active, set its `last-touched` date.
- **Complete a plan:** move the plan file to `plans/COMPLETED/<plan-id>.md`, append to Completed section here, pick #1 from Queued and promote it.
- **Block a plan:** move from Queued to Blocked, name the dependency.
- **Add a new plan:** copy `_TEMPLATE.md` to `plans/<plan-id>.md`, fill it in, append to Queued at the appropriate priority.
- **Slot a backlog item:** when the trigger condition is met, move the backlog entry to Queued at the appropriate priority position.

## Queue health rules

- Queued list should never exceed ~25 plans. If it grows beyond that, consolidate or defer.
- Every plan in Queued must have its `depends-on` list satisfied by plans above it (or by Completed plans).
- If two plans have no dependency on each other, they can technically run in parallel on separate branches. The agent protocol still works one at a time; parallelism requires human coordination.
