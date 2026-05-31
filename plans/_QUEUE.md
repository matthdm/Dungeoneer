# Work Queue

Single source of truth for what's active, queued, blocked, and done. Update this whenever a plan changes status. Plans not in this file do not exist as far as agents are concerned.

---

## Active plan

The one plan currently being worked. There is only ever one active plan at a time.

> **Active:** `5C-boss-selection.md`

---

## Queued (priority order)

Plans ready to be picked up next. Top of list = next active. When the active plan completes, promote #1 below.

| # | Plan | Size | Depends On | Notes |
|---|------|------|------------|-------|
| 1 | `5D-hub-npc-quarter.md` | Small | `5B` | Hub NPC positions + Varn hub dialogue. |
| 2 | `5C-boss-selection.md` | Large | `5B` | Boss selection engine + Varn boss fight. |
| 3 | `5D-hub-npc-quarter.md` | Small | `5B` | Hub NPC positions + Varn hub dialogue. |
| 4 | `6A-full-meta-save.md` | Medium | `5C` | Extended MetaSave, milestone unlocks, toast system. |
| 5 | `6B-ng-plus-dialogue.md` | Medium | `6A` | Meta-flag conditions, Varn NG+ trees, betrayal memory. |
| 6 | `6C-lore-system.md` | Medium | `6B` | Lore registry, unlock action, lore library UI. |
| 7 | `7A-echoes-of-self.md` | Large | `6C` | Echo recorder, echo entities, ghost visuals, shrine UI. |
| 8 | `7B-living-dungeon-ai.md` | Large | `6C` | Behavior tracker, dungeon mood, adaptive gen params. |
| 9 | `8A-item-sets.md` | Medium | `7A`, `7B` | Set registry, set bonuses, tooltip + HUD UI. |
| 10 | `8B-hub-shop-upgrades.md` | Large | `8A` | Shop NPC, upgrade station, 8 starter upgrades. |
| 11 | `9A-polish-transitions-camera.md` | Small | `8B` | Fade transitions, screen shake, boss intro pan. |
| 12 | `9B-polish-visual-effects.md` | Medium | `9A` | Particle system, damage numbers, death fades. |
| 13 | `9C-polish-audio.md` | Medium | `9B` | Audio engine, SFX hookups, music, volume wiring. |
| 14 | `9D-polish-hud-navigation.md` | Small | `9A` | Minimap, floor indicator, status effect icons. |
| 15 | `10A-additional-npcs.md` | XL | `9D`, `9C` | Seris, Mira, Kael — full arcs + boss fights. |
| 16 | `10B-abaddon-meta-narrative.md` | Large | `10A` | Abaddon, alignment system, endgame dialogue. |
| 17 | `10C-world-expansion.md` | Large | `10B` | 4 new biomes, hub evolution, cross-NPC dialogue. |
| 18 | `11A-glyph-language.md` | Large | `10C`, `10B` | NG+ glyph cipher system — inscription stones, fragment items, codex UI, 25–30 authored shrine messages. |

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

| Date | Plan | Notes |
|------|------|-------|
| 2026-05-28 | `5B-varn-arc` | varn_phase3.json (ascension trigger) + phase 3 spawn rule. All other content was already implemented. |
| 2026-05-28 | `5A-npc-phase-tracker` | NPCPhaseTracker, floor-entry auto-advance hook, add_trust clamping, trust_decay action, unit tests. |
| 2026-05-28 | `runstate-serialization` | Full mid-run save/load via RunSave + PlayerSnapshot; save on floor transition, clear on death/victory. |
| 2026-05-28 | `options-menu` | Options screen with fullscreen toggle, volume slider, controls rebind via ControlsMenu. |
| 2026-05-28 | `load-game-menu` | Load Game screen with run summary and Resume Run detection. |
| 2026-05-27 | `4F-loot-refinement` | Chest variants, Iron Key, locked-chest interaction, loot rarity scaling, loot tests. |

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
