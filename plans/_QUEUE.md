# Work Queue

Single source of truth for what's active, queued, blocked, and done. Update this whenever a plan changes status. Plans not in this file do not exist as far as agents are concerned.

---

## Active plan

The one plan currently being worked. There is only ever one active plan at a time.

> **Active:** `8A-item-sets.md`
>
> Set registry, set bonuses (2-piece and 3-piece), tooltip + HUD UI. 3 starter sets: Stormcaller, Fractalist, Chainbreaker.
>
> ⚠️ Phase 6 (6A/6B/6C) is implemented but **untested**. Manual test plan T1–T9 remains deferred. See CLAUDE.md for test scenarios.

---

## Queued (priority order)

Plans ready to be picked up next. Top of list = next active. When the active plan completes, promote #1 below.

| # | Plan | Size | Depends On | Notes |
|---|------|------|------------|-------|
| 3 | `8B-hub-shop-upgrades.md` | Large | `8A` | Shop NPC, upgrade station, 8 starter upgrades. |
| 6 | `8B-hub-shop-upgrades.md` | Large | `8A` | Shop NPC, upgrade station, 8 starter upgrades. |
| 7 | `9A-polish-transitions-camera.md` | Small | `8B` | Fade transitions, screen shake, boss intro pan. |
| 8 | `9B-polish-visual-effects.md` | Medium | `9A` | Particle system, damage numbers, death fades. |
| 9 | `9C-polish-audio.md` | Medium | `9B` | Audio engine, SFX hookups, music, volume wiring. |
| 10 | `9D-polish-hud-navigation.md` | Small | `9A` | Minimap, floor indicator, status effect icons. |
| 11 | `10A-additional-npcs.md` | XL | `9D`, `9C` | Seris, Mira, Kael — full arcs + boss fights. |
| 12 | `10B-abaddon-meta-narrative.md` | Large | `10A` | Abaddon, alignment system, endgame dialogue. |
| 13 | `10C-world-expansion.md` | Large | `10B` | 4 new biomes, hub evolution, cross-NPC dialogue. |
| 14 | `11A-glyph-language.md` | Large | `10C`, `10B` | NG+ glyph cipher system — inscription stones, fragment items, codex UI, 25–30 authored shrine messages. |

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
| 2026-07-05 | `7A-echoes-of-self.md` | Echo recorder, 3 entity types (Wicked/Hero/Memory), ghost tint, spawner, Echo Shrine UI, Remnant award on WickedEcho death. |
| 2026-07-04 | `7B-living-dungeon-ai.md` | BehaviorTracker, PlayerProfile, DungeonMood inference, GenParamsDelta, dungeon whispers. All wired; build and tests pass. |
| 2026-06-29 | `CR1-combat-overhaul.md` | GW1-style artifact combat: 23 items, 14 builds, decoupled engine, benchmarker. All 14 builds verified ≥90% survival. |
| 2026-06-02 | `6C-lore-system` | Lore registry, 15 entries, unlock_lore action, lore library UI, hub pedestal. ⚠️ Untested. |
| 2026-06-02 | `6B-ng-plus-dialogue` | meta_flag conditions, SelectTree NG+ branching, varn_ng1/2/3+betrayed trees. ⚠️ Untested. |
| 2026-06-02 | `6A-full-meta-save` | MetaSave v1, milestones (4), toast UI, run-end stat tracking, hub guards. ⚠️ Untested. |
| 2026-06-01 | `5C-boss-selection` | Varn boss entity, pre/post dialogue, arena chain tint, boss selection engine |
| 2026-06-01 | `5D-hub-npc-quarter` | Hub NPC spawning, varn_hub.json wired, npc_positions in hub.json |
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
