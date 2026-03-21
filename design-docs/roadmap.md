# Dungeoneer Development Roadmap

This document defines the implementation order for all systems. Each phase builds on the previous. Within each phase, items are listed in dependency order.

---

## Phase 1: The Run Loop

*Goal: Player can enter the dungeon, traverse multiple floors, and die or win.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 1.1 | Floor Sequencer — `FloorContext` struct, difficulty scaling per floor, biome assignment | None | `game/run.go` | `game/game.go`, `levels/generate64.go` |
| 1.2 | Exit Entity — stairwell on each floor placed in farthest room, interaction advances floor | 1.1 | `entities/exit.go` | `game/handlers.game.go`, `levels/generate64.go` |
| 1.3 | Run State — `RunState` struct tracking floor, quest flags, kill count, meta currency earned | 1.1 | `game/runstate.go` | `game/game.go` |
| 1.4 | Death Handling — player death ends run, calculates Remnants, returns to main menu | 1.3 | — | `game/game.go`, `entities/player.go` |
| 1.5 | Victory State — defeating boss floor ends run, returns to main menu | 1.3 | — | `game/game.go` |
| 1.6 | Hub World (Minimal) — hand-crafted level, player spawns here, portal starts run | 1.4, 1.5 | `levels/hub_level.go` | `game/game.go` |
| 1.7 | Meta Save (Basic) — save/load Remnants and run count to `meta.json` | 1.4 | `saves/meta.go` | `game/game.go` |
| 1.8 | Death Screen — run summary showing floors, kills, Remnants earned | 1.4 | `ui/death_screen.go` | `game/game.go` |

**Milestone: Playable 3-floor run with death/victory loop.**

---

## Phase 2: Enemy & Combat Depth

*Goal: Combat is varied, challenging, and floor-appropriate.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 2.1 | Enemy Variety — Ranged, Patrol, Swarm behaviors | Phase 1 | `entities/ranged.monster.go`, `entities/patrol.monster.go`, `entities/swarm.monster.go` | `entities/monster.go` |
| 2.2 | Biome Config — `BiomeConfig` struct, wall/floor flavor mapping, enemy pool definitions | Phase 1 | `game/biome.go` | `levels/generate64.go` |
| 2.3 | Encounter Templates — placement patterns (solo, ambush pair, firing line, etc.) | 2.1 | `game/encounters.go` | `levels/generate64.go` |
| 2.4 | Loot Tables — floor-scaled drops, enemy death drops, chest spawning | 2.2 | `items/loot.go` | `entities/monster.go`, `entities/itemdrop.go` |
| 2.5 | Status Effects — buff/debuff system (poison, slow, burn, shield) | 2.1 | `entities/effects.go` | `entities/player.go`, `entities/monster.go` |
| 2.6 | Boss Entity (Basic) — `Boss` struct, multi-phase HP, boss health bar, fallback boss | 2.1 | `entities/boss.go`, `hud/bossbar.go` | `game/game.go` |
| 2.7 | Boss Arena — sealed room generation for boss floor | 2.6 | — | `levels/generate64.go` |

**Milestone: Diverse combat encounters, biome-themed floors, defeatable boss.**

---

## Phase 3: NPCs & Dialogue

*Goal: Player can meet and talk to NPCs with branching dialogue.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 3.1 | NPC Entity — position, sprite, interaction range, interact key (E) | Phase 1 | `entities/npc.go` | `controls/controls.go`, `game/handlers.game.go` |
| 3.2 | Dialogue Panel — dark bordered UI, text rendering, typewriter effect | 3.1 | `ui/dialogue.go` | `game/game.go`, `game/draw.game.go` |
| 3.3 | Dialogue Tree — node/response/condition/action data model | 3.2 | `dialogue/tree.go`, `dialogue/types.go` | — |
| 3.4 | Dialogue Engine — condition evaluation, action execution, tree navigation | 3.3 | `dialogue/engine.go` | `ui/dialogue.go` |
| 3.5 | Quest Flag Store — `map[string]int` in RunState, read/write from dialogue actions | 3.4 | — | `game/runstate.go` |
| 3.6 | Minor NPCs — 3-5 ambient NPCs with simple dialogue, biome-assigned | 3.2 | `dialogue/data/` | `game/run.go` |
| 3.7 | NPC Spawning — spawn rules per floor, placement in rooms | 3.6 | — | `game/run.go`, `levels/generate64.go` |
| 3.8 | Interaction Indicator — speech bubble icon when player is near NPC | 3.1 | — | `game/draw.game.go` |

**Milestone: NPCs appear in dungeon, player has branching conversations.**

---

## Phase 4: Character Ascension

*Goal: Helping an NPC transforms them into the final boss.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.1 | Varn Phase 0-1 — dialogue trees, quest flags for introduction and first task | Phase 3 | `dialogue/data/varn_phase0.json`, `varn_phase1.json` | — |
| 4.2 | NPC Phase Tracker — advance NPC phase based on quest flags and floor thresholds | 4.1 | — | `game/run.go`, `entities/npc.go` |
| 4.3 | Varn Phase 2-3 — conflict dialogue, moral choice, ascension trigger | 4.2 | `dialogue/data/varn_phase2.json`, `varn_phase3.json` | — |
| 4.4 | Boss Selection — evaluate quest flags, pick ascending NPC as boss | 4.3, 2.6 | — | `game/run.go`, `entities/boss.go` |
| 4.5 | Varn Boss Form — 3-phase boss with chain-themed attacks, unique arena | 4.4 | `entities/bosses/varn.go` | `entities/boss.go` |
| 4.6 | Pre/Post-Fight Dialogue — dialogue before and after boss fight | 4.5 | `dialogue/data/varn_boss.json` | `ui/dialogue.go` |
| 4.7 | Hub NPC Quarter — defeated/met NPCs appear in hub between runs | 4.1 | — | `levels/hub_level.go`, `game/game.go` |

**Milestone: Full NPC arc from introduction through boss fight.**

---

## Phase 5: Persistence & NG+

*Goal: The Dungeon remembers across runs.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 5.1 | Full Meta Save — NPC meta state, upgrade state, lore unlocks | Phase 4 | — | `saves/meta.go` |
| 5.2 | Upgrade Station — hub UI for purchasing permanent upgrades | 5.1 | `ui/upgrade_station.go` | `game/game.go` |
| 5.3 | Shop — hub UI for run loadout purchases | 5.1 | `ui/shop.go` | `game/game.go` |
| 5.4 | NG+ Dialogue — Varn post-defeat dialogue, meta-flag-conditioned branches | 5.1 | `dialogue/data/varn_ng_plus.json` | `dialogue/engine.go` |
| 5.5 | Echoes of Self — record player on death, spawn echoes in future runs | 5.1 | `entities/echo.go`, `game/echo_recorder.go` | `game/game.go` |
| 5.6 | Echo Shrine — hub UI for managing echoes | 5.5 | `ui/echo_shrine.go` | `game/game.go` |
| 5.7 | Living Dungeon AI — player profile tracking, dungeon mood, GenParams modification | 5.1 | `game/dungeon_ai.go` | `game/run.go`, `levels/generate64.go` |
| 5.8 | Lore Library — hub UI for reading unlocked lore entries | 5.1 | `ui/lore_library.go` | `game/game.go` |

**Milestone: Full meta-progression loop with NG+ narrative depth.**

---

## Phase 6: Polish & Feel

*Goal: The game feels complete.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 6.1 | Screen Transitions — fade-to-black between floors, hub transitions | Phase 1 | `game/transition.go` | `game/game.go` |
| 6.2 | Audio System — SFX engine, ambient loops per biome, boss music | Phase 2 | `audio/audio.go` | `game/game.go`, `entities/` |
| 6.3 | Particle Effects — spell impacts, enemy death, item pickup sparkle | Phase 2 | `game/particles.go` | `game/draw.game.go` |
| 6.4 | Screen Shake — on big hits, boss phase transitions, explosions | Phase 2 | — | `game/game.go` |
| 6.5 | Enemy Death Animations — fade, dissolve, or collapse on kill | Phase 2 | — | `entities/monster.go` |
| 6.6 | Minimap — small overlay showing explored rooms and exit direction | Phase 1 | `hud/minimap.go` | `hud/hud.go` |

**Milestone: Professional game feel.**

---

## Phase 7: Thematic Completion

*Goal: The narrative vision is fully realized.*

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 7.1 | Additional Major NPCs — 2-4 more NPCs with full ascension arcs | Phase 4 | `dialogue/data/`, `entities/bosses/` | NPC/boss systems |
| 7.2 | Abaddon Integration — meta-NPC appearing after N runs, philosophical commentary | Phase 5 | `dialogue/data/abaddon.json` | `game/game.go` |
| 7.3 | Faction Alignment — player choices map to philosophical axes, affects dialogue | Phase 4 | `game/alignment.go` | `game/runstate.go` |
| 7.4 | Endgame Narrative — post-boss philosophical dialogue per NPC archetype | Phase 4 | `dialogue/data/` | `ui/dialogue.go` |
| 7.5 | Hub Evolution — visual changes to hub based on progression milestones | Phase 5 | — | `levels/hub_level.go` |
| 7.6 | Biome Expansion — Gehenna, Pandemonium, Cocytus, Lapis unlocks | Phase 2 | — | `game/biome.go` |

**Milestone: The full Dungeoneer vision is playable.**

---

## Critical Path

The shortest path to a narratively meaningful, replayable game:

```
Phase 1 (Run Loop) ──────────────────────────┐
   │                                          │
Phase 2 (Combat)     Phase 3 (NPCs)          │
   │                    │                     │
   └────────┬───────────┘                     │
            │                                 │
      Phase 4 (Ascension)                     │
            │                                 │
      Phase 5 (NG+)                           │
            │                                 │
   Phase 6 (Polish)    Phase 7 (Narrative)    │
            │                    │            │
            └────────────────────┘            │
                                              │
                                         SHIP IT
```

Phases 2 and 3 can be developed in parallel after Phase 1.
Phases 6 and 7 can be developed in parallel after Phase 5.

---

## Estimated Scope

| Phase | New Go Files | New Data Files | Complexity |
|---|---|---|---|
| Phase 1 | 5-7 | 1-2 | Medium |
| Phase 2 | 6-8 | 2-3 | Medium-High |
| Phase 3 | 5-8 | 5-10 (dialogue JSON) | Medium |
| Phase 4 | 3-5 | 5-8 (dialogue JSON) | High |
| Phase 5 | 5-7 | 2-3 | High |
| Phase 6 | 3-5 | Audio assets | Medium |
| Phase 7 | 2-4 | 10+ (dialogue JSON) | Medium |
