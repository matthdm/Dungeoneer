# Dungeoneer Design Documents

## Roadmap

- [roadmap.md](roadmap.md) — **Master development roadmap** with phased implementation plan

## Core Loop

- [run-loop.md](run-loop.md) — Run structure, floor progression, death/victory handling
- [hub-world.md](hub-world.md) — Hub layout, shops, upgrades, echo shrine, lore library
- [meta-progression.md](meta-progression.md) — Remnants currency, permanent upgrades, NG+ memory, death reset

## Narrative

- [dialogue-system.md](dialogue-system.md) — BG1/RuneScape-style dialogue UI, branching trees, quest flag conditions
- [npc-system.md](npc-system.md) — NPC entity, major vs minor NPCs, Varn character profile, spawning rules
- [quest-system.md](quest-system.md) — Character ascension quest arcs, 4-phase NPC progression
- [boss-system.md](boss-system.md) — Boss entity, combat phases, ascension binding, arenas, NG+ mutations

## World

- [biome-system.md](biome-system.md) — Biome definitions, enemy pools, loot tables, floor assignment
- [enemy-placement.md](enemy-placement.md) — Dark Souls-inspired placement philosophy, encounter templates
- [room-tagging.md](room-tagging.md) — Post-generation room role assignment (sanctuary, treasure, guard post, etc.)
- [spawn-placement.md](spawn-placement.md) — Tiered NPC/enemy placement using room tags (quest, ambient, wandering, hidden)
- [layered-world-system.md](layered-world-system.md) — Multi-floor dungeons, stairwell layer transitions

## Player

- [player-stats.md](player-stats.md) — Base stats, derived stats, equipment modifiers, save/load
- [inventory-system.md](inventory-system.md) — Grid inventory, drag-drop, equipment slots
- [items-system.md](items-system.md) — Item registry, templates, effects, serialization
- [ability-items.md](ability-items.md) — Abilities gated behind items, item sets, quest-locked rewards, mana costs

## Meta Systems

- [echoes-of-self.md](echoes-of-self.md) — Ghost recording on death, echo types (Wicked/Hero/Memory)
- [living-dungeon-ai.md](living-dungeon-ai.md) — Player behavior tracking, dungeon mood, adaptive generation

## Quality

- [test-cases.md](test-cases.md) — Manual test scenarios by phase (NPC, dialogue, spawning, etc.)
