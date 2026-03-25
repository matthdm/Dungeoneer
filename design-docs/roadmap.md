# Dungeoneer Development Roadmap

This document defines the implementation order for all systems. Each phase builds on the previous. Within each phase, items are listed in dependency order.

**Last updated:** 2026-03-24

---

## Phase 1: The Run Loop ✅ COMPLETE

*Goal: Player can enter the dungeon, traverse multiple floors, and die or win.*

| # | Task | Status |
|---|---|---|
| 1.1 | Floor Sequencer — `FloorContext` struct, difficulty scaling, biome assignment | ✅ |
| 1.2 | Exit Entity — stairwell placed in farthest room, interaction advances floor | ✅ |
| 1.3 | Run State — `RunState` struct tracking floor, quest flags, kill count, Remnants | ✅ |
| 1.4 | Death Handling — player death ends run, calculates Remnants, returns to hub | ✅ |
| 1.5 | Victory State — completing final floor ends run, returns to hub | ✅ |
| 1.6 | Hub World — hand-crafted level from `hub.json`, portal starts run, fallback generator | ✅ |
| 1.7 | Meta Save — save/load Remnants, run count, best floor to `meta.json` | ✅ |
| 1.8 | Death Screen — run summary showing floors cleared, kills, Remnants earned | ✅ |

**Milestone: Playable multi-floor run with death/victory loop.** ✅

Key files: `game/hub.go`, `game/runstate.go`, `game/metasave.go`, `entities/exit.go`, `ui/death_screen.go`

---

## Phase 2: Enemy & Combat Depth ✅ COMPLETE

*Goal: Combat is varied, challenging, and floor-appropriate.*

| # | Task | Status |
|---|---|---|
| 2.1 | Enemy Variety — Roaming, Ranged, Patrol, Swarm, Ambush, Caster behaviors | ✅ |
| 2.2 | Biome Config — `BiomeConfig` struct, wall/floor flavoring, enemy pool per biome | ✅ |
| 2.3 | Encounter Templates — 8 placement patterns (solo, ambush pair, firing line, swarm, etc.) | ✅ |
| 2.4 | Loot Tables — floor-scaled drops, enemy death drops, chest spawning | ✅ |
| 2.5 | Status Effects — buff/debuff system (poison, slow, burn, shield) with `EffectHolder` | ✅ |
| 2.6 | Boss Entity — `Boss` struct, multi-phase HP, boss health bar, fallback boss | ✅ |
| 2.7 | Boss Arena — sealed room generation, boss room tagging | ✅ |

**Milestone: Diverse combat encounters, biome-themed floors, defeatable boss.** ✅

Key files: `game/encounters.go`, `game/biome.go`, `game/boss.game.go`, `entities/monster_behaviors.go`, `entities/effects.go`, `items/loot.go`

---

## Phase 3: NPCs & Dialogue ✅ COMPLETE

*Goal: Player can meet and talk to NPCs with branching dialogue.*

| # | Task | Status |
|---|---|---|
| 3.1 | NPC Entity — position, sprite, idle behavior, bob animation, `IsPlayerInRange()` | ✅ |
| 3.2 | Dialogue Panel — dark bordered UI, typewriter text, clickable responses, portrait | ✅ |
| 3.3 | Dialogue Data Model — `DialogueTree`, `DialogueNode`, conditions, actions | ✅ |
| 3.4 | Dialogue Engine — condition evaluation, action execution, tree navigation | ✅ |
| 3.5 | Quest Flag Store — `QuestFlags map[string]int` in RunState + `NPCMetaState` in MetaSave | ✅ |
| 3.6 | Minor NPCs — 5 ambient NPCs (Hollow Monk, Scavenger, Forgotten Soldier, Mad Scholar, Weeping Shade) | ✅ |
| 3.7 | NPC Spawning — 7-tier placement system with room-tag awareness | ✅ |
| 3.8 | Interaction Indicator — "[E] Talk" hint centered above NPC in isometric view | ✅ |
| 3.9 | Room Tagging — post-generation pass assigning semantic roles (sanctuary, dead_end, etc.) | ✅ |
| 3.10 | NPC Met Persistence — auto-set `NPCMeta.Met` on first dialogue, saved to meta.json | ✅ |

**Milestone: NPCs appear in dungeon and hub, player has branching conversations.** ✅

Key files: `entities/npc.go`, `entities/npc_behaviors.go`, `dialogue/types.go`, `dialogue/loader.go`, `ui/dialogue.go`, `game/npc.game.go`, `game/npc_data.go`, `levels/room_tagger.go`

Design docs: `room-tagging.md`, `spawn-placement.md`, `dialogue-system.md`, `npc-system.md`

---

## Phase 4: Ability Items, Economy & Equipment

*Goal: Abilities are earned, not given. The player starts with only melee and must find items to unlock spells, dash, and grapple. The hub economy gives reason to return between runs.*

This is the largest single phase. It transforms the game from "all abilities free" to "every ability is a discovery." It also builds the economic loop (Remnants → shop/upgrades → stronger starts) that makes the run cycle rewarding. Must complete before Phase 5 since NPC quest rewards need to give powerful ability items.

See: `ability-items.md` for full design.

### 4A: Ability-Item Gating (Core)

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.1 | `GrantsAbility` Field — add `GrantsAbility string`, `AbilitySlot string`, `SetID string` to `ItemTemplate` | None | — | `items/types.go` |
| 4.2 | Learned Abilities — add `LearnedAbilities []string` to Player, populated from equipped items via `OnEquip`/`OnUnequip` | 4.1 | — | `entities/player.go` |
| 4.3 | Gate Spell Casting — handler checks `LearnedAbilities` before allowing cast on keys 1-6; no ability = no cast | 4.2 | — | `game/handlers.game.go`, `game/spells.game.go` |
| 4.4 | Gate Dash — check if "dash" is in `LearnedAbilities` before allowing Shift dash | 4.2 | — | `game/handlers.game.go` |
| 4.5 | Gate Grapple — check if "grapple" is in `LearnedAbilities` before allowing F grapple | 4.2 | — | `game/handlers.game.go` |
| 4.6 | Dynamic Spell Bar — HUD shows only learned spells; empty slots shown as locked; spells shift to fill gaps on unequip | 4.3 | — | `hud/hud.go`, `game/draw.game.go` |
| 4.7 | Mana Costs — each spell deducts mana on cast; gray out spell icon when insufficient mana; mana regen scales with Intelligence | 4.3 | — | `game/spells.game.go`, `entities/player.go`, `hud/hud.go` |

### 4B: Ability Item Templates

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.8 | Spell-Granting Items — create 6 ability items: Grimoire of Fire (Fireball), Wand of Chaos (Chaos Ray), Storm Rod (Lightning), Tempest Tome (Lightning Storm), Fractal Seed (Fractal Bloom), Living Branch (Fractal Canopy) | 4.1 | `items/ability_items.go` | `items/registry.go` |
| 4.9 | Utility-Granting Items — Windrunner Boots (Dash, 3 charges), Shadowstep Cloak (Dash, 5 charges), Iron Grapple (Grapple, 8 tiles), Voidhook (Grapple, 12 tiles) | 4.1 | — | `items/ability_items.go` |
| 4.10 | Ability Items in Loot Tables — add spell/utility items to biome loot pools at Uncommon+ rarity | 4.8 | — | `items/loot.go` |
| 4.11 | Floor 1 Guaranteed Drop — ensure at least one spell-granting item drops on Floor 1 (chest or early enemy) so the player isn't stuck with only melee | 4.10 | — | `game/hub.go`, `items/loot.go` |
| 4.12 | Quest-Locked Items — mark Tempest Tome, Fractal Seed, Living Branch, Shadowstep Cloak, Voidhook as quest-only (excluded from loot tables) | 4.8, 4.9 | — | `items/ability_items.go` |

### 4C: Equipment & Stats

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.13 | Equipment Effects — equipped items apply `StatModifiers` via `RecalculateStats()`, update on equip/unequip | None | — | `entities/player.go`, `inventory/inventory.go` |
| 4.14 | Item Quality Tiers — Common/Uncommon/Rare/Legendary with color-coded names and stat multipliers | 4.13 | `items/quality.go` | `items/types.go`, `ui/tooltip.go` |
| 4.15 | Equipment Tooltip — show stat deltas (green +, red -) and granted ability when hovering equippable items | 4.13 | — | `ui/tooltip.go` |
| 4.16 | Gold Economy — monsters drop gold, gold persists within run, displayed in HUD | None | — | `entities/player.go`, `hud/hud.go`, `entities/monster.go` |
| 4.17 | Item Actions in Dialogue — `give_item`, `take_item`, `has_item` conditions work against player inventory | 4.13 | — | `game/npc.game.go`, `dialogue/types.go` |

### 4D: Item Sets

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.18 | Set Registry — `ItemSet` struct with ID, member item IDs, and `[]SetBonus` (pieces required, stat bonuses, bonus ability) | 4.1 | `items/sets.go` | — |
| 4.19 | Set Bonus Tracking — on equip/unequip, recalculate active set bonuses; apply stat bonuses and bonus abilities | 4.18, 4.2 | — | `entities/player.go`, `items/sets.go` |
| 4.20 | Set Bonus UI — equipment panel shows set name, pieces owned/required, active bonuses highlighted | 4.19 | — | `ui/tooltip.go`, `hud/hud.go` |
| 4.21 | Starter Sets — define 2-3 item sets: Stormcaller (3pc, lightning focus), Fractalist (2pc, bloom+canopy synergy), Chainbreaker (3pc, Varn quest-locked, melee AoE) | 4.18 | — | `items/sets.go` |

### 4E: Hub Shop & Upgrades

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.22 | Shop NPC — permanent hub NPC with shop interface, speech bubble prompt | 4.16 | — | `game/hub.go`, `game/npc_data.go` |
| 4.23 | Shop UI — scrollable item list, Remnant balance, purchase confirmation, stock limits per run | 4.22 | `ui/shop.go` | `game/game.go`, `game/draw.game.go` |
| 4.24 | Shop Inventory — starter consumables + basic ability items (Grimoire of Fire, Windrunner Boots, Iron Grapple) priced in Remnants | 4.23, 4.8 | `game/shop_data.go` | — |
| 4.25 | Run Loadout — purchased items appear in player inventory when run starts | 4.24 | — | `game/hub.go` |
| 4.26 | Upgrade Registry — `UpgradeDef` struct with ID, name, costs per level, max level, `Apply()` function | None | `game/upgrades.go` | — |
| 4.27 | Starter Upgrades — 8 upgrades: Iron Constitution (+HP), Sharpened Edge (+DMG), Deep Pockets (+inv row), Spell Affinity (-mana cost), Mana Well (+max mana), Quick Draw (-cooldown), Scavenger (+Remnants), Innate Dash (start with dash, no item) | 4.26 | — | `game/upgrades.go` |
| 4.28 | Upgrade Station NPC — permanent hub NPC near portal, opens upgrade UI | 4.26 | — | `game/hub.go`, `game/npc_data.go` |
| 4.29 | Upgrade UI — category tabs, level pips, cost display, purchase with Remnants | 4.28 | `ui/upgrade_station.go` | `game/game.go`, `game/draw.game.go` |
| 4.30 | Apply Upgrades at Run Start — iterate `Meta.Upgrades`, call `Apply()` on player before dungeon entry | 4.27 | — | `game/hub.go`, `entities/player.go` |
| 4.31 | Meta Save Extension — persist `Upgrades map[string]int` in MetaSave, init on load | 4.26 | — | `game/metasave.go` |

### 4F: Loot Refinement

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 4.32 | Weighted Loot by Quality — higher floors and rarer rooms produce better item tiers; ability items skew Uncommon+ | 4.14 | — | `items/loot.go` |
| 4.33 | Chest Variants — wooden (common), iron (uncommon+), gold (rare+) with locked chests requiring Iron Key | 4.32 | `entities/chest.go` | `game/handlers.game.go` |
| 4.34 | Treasure Room Loot — rooms tagged `treasure` always contain a chest with elevated quality | 4.32 | — | `game/hub.go`, `items/loot.go` |

**Milestone: Abilities are earned through items. Complete meta economy. Each run feels different based on what you find. Quest-locked items create NPC motivation.**

---

## Phase 5: Character Ascension

*Goal: Helping an NPC transforms them into the final boss. The player's moral choices have mechanical consequences.*

Requires Phase 4 (item actions in dialogue, quest flag economy) and Phase 3 (dialogue engine, NPC spawning).

### 5A: NPC Phase System

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 5.1 | NPC Phase Tracker — `NPCPhaseTracker` struct, advance phase based on quest flags + floor thresholds | Phase 3 | `game/npc_phases.go` | `entities/npc.go`, `game/runstate.go` |
| 5.2 | Phase-Conditional Dialogue Selection — `SelectTree()` reads `npc_phase` flag and `NPCMeta.DefeatCount` to pick correct tree | 5.1 | — | `dialogue/loader.go` |
| 5.3 | Trust System — `add_trust` dialogue action, trust thresholds gate responses, trust decays on betrayal | 5.1 | — | `game/npc.game.go`, `dialogue/types.go` |

### 5B: Varn's Arc (First Major NPC)

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 5.4 | Varn Template — major NPC definition with Chainkeeper sprite, 4-phase dialogue IDs, sanctuary placement | 5.1 | — | `game/npc_data.go` |
| 5.5 | Varn Phase 0: Introduction — player discovers Varn chained in a sanctuary room, learns his philosophy of order | 5.4 | `dialogues/varn_phase0.json` | — |
| 5.6 | Varn Phase 1: First Task — Varn asks for a key item, trust builds, reveals backstory | 5.5 | `dialogues/varn_phase1.json` | — |
| 5.7 | Varn Phase 2: Conflict — Varn's methods become questionable, player faces moral choice (support or challenge) | 5.6 | `dialogues/varn_phase2.json` | — |
| 5.8 | Varn Phase 3: Ascension Trigger — if trust is high enough, Varn "ascends" — chains break, power surges, exits dialogue changed | 5.7 | `dialogues/varn_phase3.json` | — |
| 5.9 | Varn Major NPC Spawning — Varn appears on floors 2+ in sanctuary rooms, phase determines sprite variant | 5.4 | — | `game/npc.game.go` |

### 5C: Boss Selection & Varn Boss Fight

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 5.10 | Boss Selection Engine — evaluate all major NPC quest flags, pick highest-phase NPC as final boss, fallback to generic boss | 5.8, 2.6 | `game/boss_selection.go` | `game/hub.go` |
| 5.11 | Varn Boss Form — 3-phase boss: chain whip (melee), chain eruptions (AoE), unchained frenzy (fast melee) | 5.10 | `entities/bosses/varn.go` | `entities/boss.go` |
| 5.12 | Boss Arena Theming — Varn's arena has chain decoration tiles, unique wall sprites, sealed entrance | 5.11 | — | `levels/generate64.go`, `game/boss.game.go` |
| 5.13 | Pre-Fight Dialogue — Varn speaks before combat begins, text varies by trust level and betrayal status | 5.11 | `dialogues/varn_boss_pre.json` | `game/boss.game.go` |
| 5.14 | Post-Fight Dialogue — after defeat, Varn has final words; different if first defeat vs. NG+ | 5.11 | `dialogues/varn_boss_post.json` | `game/boss.game.go` |

### 5D: Hub NPC Quarter

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 5.15 | Hub NPC Positions — define fixed tile positions in hub.json for up to 6 major NPCs | 5.4 | — | `levels/hub.json` |
| 5.16 | Hub NPC Spawning — check `NPCMeta.Met` for each major NPC, spawn in hub at assigned position | 5.15 | — | `game/hub.go`, `game/npc.game.go` |
| 5.17 | Hub NPC Dialogue — between-run conversations reflecting quest progress, hints for next encounter | 5.16 | `dialogues/varn_hub.json` | — |

**Milestone: Full Varn arc from introduction through ascension to boss fight. Hub reflects player choices.**

---

## Phase 6: NG+ & Memory

*Goal: The Dungeon remembers across runs. NPCs react to their own deaths. Lore accumulates.*

Requires Phase 5 (at least one defeated major NPC to trigger NG+ content).

### 6A: Full Meta Save

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 6.1 | Extended MetaSave — add `CompletedRuns`, `TotalDeaths`, `TotalRemnants` (lifetime), `LoreUnlocked []string`, `HubState` | Phase 5 | — | `game/metasave.go` |
| 6.2 | Milestone Unlocks — `CheckMilestones()` evaluates MetaSave thresholds, unlocks shop (run 1), upgrades (run 3), echo shrine (first death), etc. | 6.1 | `game/milestones.go` | `game/hub.go` |
| 6.3 | Milestone UI — small notification toast when a milestone unlocks ("Upgrade Station now available") | 6.2 | — | `ui/toast.go`, `game/draw.game.go` |

### 6B: NG+ Dialogue

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 6.4 | Meta-Flag Conditions — dialogue conditions can read `NPCMeta` fields (defeat count, betrayed, highest phase) via `meta_flag_gte`, `meta_flag_equals` | 6.1 | — | `dialogue/types.go`, `game/npc.game.go` |
| 6.5 | Varn NG+ Trees — post-defeat dialogue: recognition after 1 defeat, self-doubt after 2, meta-awareness after 3 | 6.4 | `dialogues/varn_ng1.json`, `varn_ng2.json`, `varn_ng3.json` | — |
| 6.6 | Betrayal Memory — if player betrayed Varn (chose against him in Phase 2), he is hostile/suspicious on re-encounter, requires more trust to progress | 6.4 | `dialogues/varn_betrayed.json` | — |
| 6.7 | Trust Accumulation — `NPCMeta.TotalTrust` accumulates across runs, high trust unlocks deeper lore dialogue | 6.4 | — | `game/npc.game.go` |

### 6C: Lore System

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 6.8 | Lore Registry — `LoreDef` struct with ID, title, category (character/cosmology/history/fragment), body text | 6.1 | `game/lore.go` | — |
| 6.9 | Lore Unlock Action — `unlock_lore` dialogue action adds entry to `Meta.LoreUnlocked` | 6.8 | — | `game/npc.game.go` |
| 6.10 | Lore Library UI — scrollable text panel in hub, categorized tabs, locked entries shown as "???" | 6.9 | `ui/lore_library.go` | `game/game.go`, `game/draw.game.go` |
| 6.11 | Lore Content — 15-20 lore entries covering Abaddon's nature, dungeon cosmology, NPC backstories, cryptic fragments | 6.8 | `data/lore.json` | — |

**Milestone: NG+ dialogue transforms repeat encounters. Lore rewards exploration and trust.**

---

## Phase 7: Echoes & Living Dungeon

*Goal: The player's past lives haunt the dungeon. The dungeon adapts to the player.*

These systems both depend on accumulated meta-save data (multiple completed runs). They can be developed in parallel.

### 7A: Echoes of Self

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 7.1 | Echo Recorder — capture player position, HP, equipment, and cause of death each tick during a run | Phase 6 | `game/echo_recorder.go` | `game/game.go` |
| 7.2 | Echo Data Model — `EchoRecord` struct with path, actions, HP history, death cause, serialized to `echoes/*.json` | 7.1 | `entities/echo.go` | `game/metasave.go` |
| 7.3 | Echo Spawner — spawn up to N echoes per floor (configurable via upgrades), placed near historical death locations | 7.2 | — | `game/hub.go` |
| 7.4 | Wicked Echo — enemy-type echo, same stats as death snapshot, miniboss difficulty, drops Remnants on defeat | 7.3 | — | `entities/echo.go`, `game/encounters.go` |
| 7.5 | Hero Echo — ally-type echo, assists in combat for a limited time, disappears after encounter | 7.3 | — | `entities/echo.go` |
| 7.6 | Memory Fragment — static ghost NPC, non-hostile, triggers lore dialogue or stat hint | 7.3 | — | `entities/echo.go`, `game/npc.game.go` |
| 7.7 | Echo Visual Style — ghost shader (blue/purple tint, transparency, flicker), distinct from living entities | 7.3 | — | `entities/echo.go`, `game/draw.game.go` |
| 7.8 | Echo Shrine UI — hub interface for viewing stored echoes, selecting echo type preference, banishing old echoes | 7.3 | `ui/echo_shrine.go` | `game/game.go`, `game/hub.go` |

### 7B: Living Dungeon AI

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 7.9 | Behavior Tracker — record combat style, spell usage, enemy avoidance, rooms skipped, risk tolerance per run | Phase 6 | `game/behavior_tracker.go` | `game/game.go` |
| 7.10 | Player Profile — aggregate behavior across last N runs into `PlayerProfile` struct, save in MetaSave | 7.9 | — | `game/metasave.go` |
| 7.11 | Dungeon Mood — infer `DungeonMood` traits (Spiteful, Chaotic, Cautious, Deceptive) from PlayerProfile | 7.10 | `game/dungeon_ai.go` | — |
| 7.12 | Adaptive GenParams — DungeonMood modifies `GenParams`: room density, corridor length, trap frequency, enemy composition bias | 7.11 | — | `game/hub.go`, `levels/generate64.go` |
| 7.13 | Counter-Strategy Spawning — if player favors melee, spawn more ranged; if player rushes, spawn more ambushes | 7.11 | — | `game/encounters.go`, `game/biome.go` |
| 7.14 | Dungeon Whispers — flavor text on floor entry reflecting dungeon mood ("The walls feel watchful...", "Something remembers your fire...") | 7.11 | — | `game/hub.go`, `ui/toast.go` |

**Milestone: Past runs manifest as echoes. The dungeon evolves to challenge the player's habits.**

---

## Phase 8: Polish & Feel

*Goal: The game feels complete. Transitions are smooth, feedback is satisfying, navigation is clear.*

Can be started after Phase 5. Many tasks are independent and can be parallelized.

### 8A: Transitions & Camera

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 8.1 | Screen Transitions — fade-to-black between floors, fade-in on hub load, configurable duration | Phase 1 | `game/transition.go` | `game/game.go`, `game/hub.go` |
| 8.2 | Boss Intro Cutscene — camera pans to boss room, brief dialogue, then combat begins | Phase 5 | — | `game/boss.game.go` |
| 8.3 | Screen Shake — on heavy hits, boss phase transitions, explosions; configurable intensity and decay | Phase 2 | `game/screenshake.go` | `game/game.go`, `game/draw.game.go` |

### 8B: Visual Effects

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 8.4 | Particle System — lightweight emitter/particle model, pooled allocation | Phase 2 | `game/particles.go` | `game/draw.game.go` |
| 8.5 | Spell Impact Particles — burst on spell hit (fire=orange, ice=blue, etc.) | 8.4 | — | `spells/` |
| 8.6 | Enemy Death Effects — fade + dissolve on kill, brief flash | 8.4 | — | `entities/monster.go` |
| 8.7 | Item Pickup Sparkle — small particle burst when picking up items | 8.4 | — | `game/handlers.game.go` |
| 8.8 | Damage Numbers — floating numbers on hit, color-coded by damage type | None | `game/damage_numbers.go` | `game/draw.game.go` |

### 8C: Audio

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 8.9 | Audio Engine — SFX playback via Ebiten audio, volume control, concurrent sound limit | None | `audio/audio.go` | `game/game.go` |
| 8.10 | Combat SFX — hit, miss, spell cast, enemy death, player damage | 8.9 | `audio/sfx/` | `entities/`, `spells/` |
| 8.11 | Ambient Loops — per-biome background music, crossfade on floor change | 8.9 | `audio/music/` | `game/hub.go` |
| 8.12 | Boss Music — unique track for boss encounters, triggered on boss room entry | 8.11 | — | `game/boss.game.go` |
| 8.13 | UI SFX — menu open/close, dialogue typewriter tick, purchase confirm, level up | 8.9 | — | `ui/` |

### 8D: HUD & Navigation

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 8.14 | Minimap — small overlay showing explored rooms, player position, exit direction, toggle with M | Phase 1 | `hud/minimap.go` | `hud/hud.go`, `game/draw.game.go` |
| 8.15 | Floor Indicator — persistent HUD element showing "Floor 2/5" and biome name | Phase 1 | — | `hud/hud.go` |
| 8.16 | Status Effect Icons — row of active buff/debuff icons with remaining duration | Phase 2 | — | `hud/hud.go` |

**Milestone: Professional game feel. Juice in every interaction.**

---

## Phase 9: Thematic Completion

*Goal: The narrative vision is fully realized. Multiple NPC arcs, Abaddon's meta-commentary, and the full biome catalogue.*

Requires Phase 6 (NG+ dialogue) and Phase 5 (ascension framework). Phase 7 (Living Dungeon) enhances but doesn't block.

### 9A: Additional Major NPCs

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 9.1 | Seris, The Ember — fire-themed NPC, philosophy of destruction/renewal, 4-phase arc | Phase 5 | `dialogues/seris_*.json`, `entities/bosses/seris.go` | `game/npc_data.go` |
| 9.2 | Mira, The Veil — illusion-themed NPC, philosophy of deception/truth, 4-phase arc | Phase 5 | `dialogues/mira_*.json`, `entities/bosses/mira.go` | `game/npc_data.go` |
| 9.3 | Kael, The Root — nature-themed NPC, philosophy of growth/stagnation, 4-phase arc | Phase 5 | `dialogues/kael_*.json`, `entities/bosses/kael.go` | `game/npc_data.go` |
| 9.4 | NPC Boss Variants — each major NPC has a unique 3-phase boss form with themed attacks and arena | 9.1-9.3 | — | `entities/boss.go`, `game/boss.game.go` |

### 9B: Abaddon & Meta-Narrative

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 9.5 | Abaddon Entity — meta-NPC, appears in hub after 10+ runs, cannot be fought, speaks in philosophical riddles | Phase 6 | `dialogues/abaddon_*.json` | `game/npc_data.go`, `game/hub.go` |
| 9.6 | Abaddon Dialogue Tiers — 5+ dialogue trees unlocked by run count, NPC defeats, lore collected, alignment | 9.5 | — | `dialogue/loader.go` |
| 9.7 | Philosophical Alignment — player choices across NPC arcs map to axes (Order↔Chaos, Creation↔Destruction), stored in MetaSave | Phase 5 | `game/alignment.go` | `game/runstate.go`, `game/metasave.go` |
| 9.8 | Alignment-Gated Dialogue — NPC responses and Abaddon commentary shift based on player's philosophical lean | 9.7 | — | `dialogue/types.go`, `game/npc.game.go` |
| 9.9 | Endgame Narrative — post-all-boss-defeat philosophical resolution dialogue per NPC archetype | 9.4, 9.5 | `dialogues/endgame_*.json` | `ui/dialogue.go` |

### 9C: World Expansion

| # | Task | Dependencies | New Files | Touches |
|---|---|---|---|---|
| 9.10 | Biome Expansion — add Gehenna (fire), Pandemonium (chaos), Cocytus (ice), Lapis (crystal) biomes | Phase 2 | — | `game/biome.go`, `levels/generate64.go` |
| 9.11 | Biome-Specific Encounters — unique encounter templates per biome (fire ambush, ice patrol, etc.) | 9.10 | — | `game/encounters.go` |
| 9.12 | Hub Evolution — hub visual changes based on progression milestones (torches lit, banners hung, NPC decorations) | Phase 6 | — | `levels/hub.json`, `game/hub.go` |
| 9.13 | Hub NPC Interactions — major NPCs in hub comment on each other, react to player alignment | 9.4, 9.7 | `dialogues/hub_cross_*.json` | — |

**Milestone: The full Dungeoneer vision is playable. Multiple NPC arcs, philosophical depth, evolved world.**

---

## Critical Path

The shortest path to a narratively meaningful, replayable game:

```
Phase 1 (Run Loop) ✅ ─────────────────────────┐
   │                                            │
Phase 2 (Combat) ✅      Phase 3 (NPCs) ✅      │
   │                       │                    │
   └──────────┬────────────┘                    │
              │                                 │
     Phase 4 (Hub Economy)                      │
              │                                 │
     Phase 5 (Ascension)                        │
              │                                 │
     Phase 6 (NG+ & Memory)                     │
              │                                 │
   ┌──────────┴──────────┐                      │
   │                     │                      │
Phase 7 (Echoes/AI)  Phase 8 (Polish)           │
   │                     │                      │
   └──────────┬──────────┘                      │
              │                                 │
     Phase 9 (Narrative)                        │
              │                                 │
              └─────────────────────────────────┘
                                           SHIP IT
```

**Parallel opportunities:**
- Phases 7 and 8 can be developed in parallel after Phase 6
- Phase 8 tasks (audio, particles, transitions) are mostly independent of each other
- Phase 9A (new NPCs) and 9B (Abaddon) can be developed in parallel
- Phase 8 can start earlier (after Phase 5) for non-NG+-dependent polish

---

## Backlog

*Items not tied to a specific phase — optimization and scalability work to be scheduled when needed.*

| ID | Task | Trigger | Files | Severity |
|---|---|---|---|---|
| B.1 | **Dynamic FOV ray length** — `rayLength=1000` is hardcoded; should scale with map diagonal (`sqrt(W²+H²) * 1.5`) to support larger maps | Before any map size increase beyond 64x64 | `src/fov/fov.go:39` | High |
| B.2 | **A* heap upgrade** — open list uses a slice with O(n) extraction; replace with `container/heap` for O(log n) | When monster count or map size causes frame drops | `src/pathing/astar.go:77-100` | Medium |
| B.3 | **Rename Generate64x64** — function accepts parametric `GenParams` but name implies 64x64; rename to `GenerateLevel` for clarity | Cosmetic, low priority | `src/levels/generate64.go` | Low |
| B.4 | **Layered World System** — multi-floor vertical layers within a single level (sub-basements, upper walkways) per `layered-world-system.md` | When level variety demands sub-areas or secret zones | `src/levels/` | Medium |
| B.5 | **Item crafting system** — merge items in crafting UI, per `inventory-system.md` expansion hooks | When item variety justifies a crafting loop | `src/items/`, `src/ui/` | Low |
| B.6 | **Map size scaling** — 128x128 support requires B.1 + B.2 first, then testing BFS/generation performance | When 64x64 feels too small for deeper runs | `src/levels/generate64.go` | Medium |

---

## Estimated Scope

| Phase | Status | New Go Files | New Data Files | Complexity |
|---|---|---|---|---|
| Phase 1 | ✅ Complete | — | — | — |
| Phase 2 | ✅ Complete | — | — | — |
| Phase 3 | ✅ Complete | — | — | — |
| Phase 4 | Next | 8-12 | 3-4 (shop/upgrade/set data) | High |
| Phase 5 | Planned | 4-6 | 8-12 (dialogue JSON) | High |
| Phase 6 | Planned | 3-5 | 5-8 (lore + NG+ dialogue JSON) | Medium-High |
| Phase 7 | Planned | 4-6 | 2-3 (echo data) | High |
| Phase 8 | Planned | 5-8 | Audio assets | Medium |
| Phase 9 | Planned | 5-8 | 20+ (dialogue + lore JSON) | Medium-High |

---

## Design Doc Index

All system designs that inform this roadmap:

| System | Document | Primary Phase |
|---|---|---|
| Run loop & floors | `run-loop.md` | Phase 1 |
| Hub world layout | `hub-world.md` | Phase 1, 4, 5 |
| Enemy behaviors | `enemy-placement.md` | Phase 2 |
| Biome theming | `biome-system.md` | Phase 2 |
| Boss encounters | `boss-system.md` | Phase 2, 5 |
| Items & inventory | `inventory-system.md`, `items-system.md` | Phase 2, 4 |
| Ability-item gating | `ability-items.md` | Phase 4 |
| Player stats | `player-stats.md` | Phase 4 |
| NPC framework | `npc-system.md` | Phase 3, 5 |
| Dialogue engine | `dialogue-system.md` | Phase 3, 5, 6 |
| Quest & ascension | `quest-system.md` | Phase 5 |
| Room tagging | `room-tagging.md` | Phase 3 |
| Spawn placement | `spawn-placement.md` | Phase 3 |
| Meta-progression | `meta-progression.md` | Phase 4, 6 |
| Echoes of Self | `echoes-of-self.md` | Phase 7 |
| Living Dungeon AI | `living-dungeon-ai.md` | Phase 7 |
| Layered worlds | `layered-world-system.md` | Backlog |
| Test cases | `test-cases.md` | All phases |
