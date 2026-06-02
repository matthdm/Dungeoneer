# Phases 7A → 9D Complete: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship Phases 7–9: Echoes of Self (past-run ghost entities), Living Dungeon AI (adaptive generation), Item Sets (set bonuses), Hub Shop & Upgrades (meta economy), and full Polish pass (transitions, screen shake, particles, damage numbers, audio, minimap).

**Architecture:** Sequential waves respecting file-ownership constraints. 7A and 7B share `metasave.go`/`hub.go`/`game.go` — they run sequentially. 8A is isolated (items/player/HUD). 8B adds hub economy entities. Polish phases (9A–9D) are sequential within the render/camera/audio stacks. Q-gathering for all agents is done in parallel (read-only); execution respects the dependency chain.

**Tech Stack:** Go 1.23, Ebiten v2.8, `ebitenutil` for debug text, `ebiten.Image` for panels, JSON for echo records. No new external dependencies.

---

## Status at plan start (2026-06-02)

| Phase | Code | Unit Tests | Manual Tests |
|---|---|---|---|
| 6A MetaSave + Milestones + Toast | ✅ written | ✅ 12 tests pass | ⚠️ untested (T3, T4) |
| 6B NG+ Dialogue + Conditions | ✅ written | ✅ 10 tests pass | ⚠️ untested (T5–T8) |
| 6C Lore System | ✅ written | ✅ 5 tests pass | ⚠️ untested (T4, T7) |
| 7A–9D | ❌ not started | — | — |

Phase 6 manual testing (T1–T9 in session test plan) **must be completed before Phase 7 ships** to confirm the MetaSave/milestone foundation is sound.

---

## Execution Waves

```
Wave 1:  [7A]           ← echoes (touches game.go, hub.go, metasave.go)
Wave 2:  [7B]           ← dungeon AI (same shared files — must follow 7A)
Wave 3:  [8A] ‖ [9A]   ← item sets + transitions run in parallel (no file overlap)
Wave 4:  [8B]           ← hub shop/upgrades (depends on 8A's MetaSave fields)
Wave 5:  [9D]           ← HUD nav (depends on 9A's camera offsets; touches draw.game.go)
Wave 6:  [9B]           ← particles + VFX (touches draw.game.go, game.go)
Wave 7:  [9C]           ← audio (touches game.go, audio/ new directory)
```

Wave 3 parallelism is verified: 8A touches `items/sets.go`, `entities/player.go`, `ui/tooltip.go`, `hud/hud.go`. 9A touches `game/transition.go`, `game/screenshake.go`, `game/game.go`, `game/draw.game.go`, `game/hub.go`, `game/boss.game.go`. Zero file overlap.

---

## File Map

| File | Phase | Action |
|---|---|---|
| `src/game/echo_recorder.go` | 7A | New |
| `src/entities/echo.go` | 7A | New |
| `src/ui/echo_shrine.go` | 7A | New |
| `src/game/metasave.go` | 7A then 7B | 7A adds EchoFiles; 7B adds PlayerProfile/RecentBehavior |
| `src/game/game.go` | 7A then 7B then 9A | Sequential additions |
| `src/game/hub.go` | 7A then 7B then 8B then 9A | Sequential additions |
| `src/game/draw.game.go` | 7A then 9A then 9D then 9B | Sequential additions |
| `src/game/behavior_tracker.go` | 7B | New |
| `src/game/dungeon_ai.go` | 7B | New |
| `src/levels/generate64.go` | 7B | Add GenParams fields |
| `src/game/encounters.go` | 7B | Mood-based enemy bias |
| `src/game/biome.go` | 7B | Mood-based pool weights |
| `src/items/sets.go` | 8A | New |
| `src/entities/player.go` | 8A | Set bonus integration |
| `src/ui/tooltip.go` | 8A | Set membership display |
| `src/hud/hud.go` | 8A then 9D | 8A adds set bonuses; 9D adds minimap |
| `src/game/shop_data.go` | 8B | New |
| `src/ui/shop.go` | 8B | New |
| `src/game/upgrades.go` | 8B | New |
| `src/ui/upgrade_station.go` | 8B | New |
| `src/game/npc_data.go` | 8B | Shop/upgrade NPC defs |
| `src/game/transition.go` | 9A | New |
| `src/game/screenshake.go` | 9A | New |
| `src/game/boss.game.go` | 9A | Boss intro camera pan |
| `src/hud/minimap.go` | 9D | New |
| `src/game/particles.go` | 9B | New |
| `src/game/damage_numbers.go` | 9B | New (existing file may exist — check) |
| `src/spells/` | 9B then 9C | 9B: particle hooks; 9C: SFX hooks |
| `src/entities/monster.go` | 9B then 9C | 9B: death fade; 9C: SFX |
| `src/audio/audio.go` | 9C | New (new directory) |

---

## Wave 1: Phase 7A — Echoes of Self

**Acceptance criteria (from `plans/7A-echoes-of-self.md`):**
- After every run, `EchoRecord` written to `echoes/run_N.json` (position snapshots every 2s, equipped items, death cause, death floor)
- Echo spawner places 0–3 echo entities per floor near historical death locations
- `WickedEcho`: melee enemy, stats from source run, drops 5–15 Remnants
- `HeroEcho`: follows player, attacks enemies, 60s lifetime
- `MemoryFragment`: static ghost NPC, 1–2 line message on interact
- All echoes: blue/purple tint, alpha 0.6, flicker ±0.1 at 2Hz
- Echo Shrine in hub: lists echo runs, allows banishing
- `go build ./...` passes; unit tests for EchoRecord serialization and echo spawner sampling

**File envelope:**
- `src/game/echo_recorder.go` (new)
- `src/entities/echo.go` (new)
- `src/ui/echo_shrine.go` (new)
- `src/game/metasave.go` — add `EchoFiles []string`
- `src/game/game.go` — start/stop recorder; store active echoes list
- `src/game/hub.go` — echo spawner pass on floor gen; echo shrine spawn
- `src/game/draw.game.go` — ghost tint/flicker draw pass

**Forbidden:** `src/coords/`, `src/collision/`, `src/dialogue/` (loader), `src/game/behavior_tracker.go` (7B scope)

**Key design decisions (finalized):**
- Position snapshots: every 2 seconds of gameplay, max 50 per floor
- Echo floor matching: death floor = spawn floor, random sample if >3 candidates
- Echo type weighting per spawn: 50% Wicked, 30% Hero, 20% Memory
- Max echoes stored: 10 runs (cap + oldest-first eviction)
- Ghost flicker: `time.Now().UnixNano()` in draw pass (matches existing item-sparkle pattern)
- `IsEcho bool` added to `Monster` struct (not forbidden; cleanest path for combat integration)
- WickedEcho/HeroEcho live in `g.Monsters`; HeroEcho behavior holds `*[]*Monster` for ally targeting
- "Drops Remnants" = increment `g.Meta.Remnants` directly (no item spawn)
- MemoryFragment uses `NPC.OnInteract` + runtime `SimpleDialogue.ToTree()` — no pre-registered tree ID
- `echoes/` directory created via `os.MkdirAll` at first write; path relative to working dir (same as `meta.json`)
- Echo Shrine follows `LoreLibrary` full-screen overlay pattern
- HeroEcho despawn: `IsDead = true` after 60s lifetime counter (DeltaTime accumulator on Echo struct)

**Tests to write:**
- `TestEchoRecord_Serialize`: write record, read it back, fields match
- `TestEchoSpawner_CapsAtThreePerFloor`: >3 matching records → 3 spawned
- `TestEchoSpawner_EmptyRecords_SpawnsNone`

---

## Wave 2: Phase 7B — Living Dungeon AI

**Acceptance criteria (from `plans/7B-living-dungeon-ai.md`):**
- `BehaviorTracker` records per run: spell usage map, melee/ranged kill counts, room skip ratio, enemies avoided
- `PlayerProfile` aggregated from last 5 runs (exponential decay, most recent weighted 2×), saved in MetaSave
- `DungeonMood` inferred each run start: Spiteful/Chaotic/Cautious/Deceptive/Neutral (priority order: Spiteful > Deceptive > Cautious > Chaotic > Neutral)
- Mood modifies `GenParams`: Spiteful → tighter corridors/more dead ends; Chaotic → larger rooms/more enemies; Cautious → more ambush placements; Deceptive → ranged enemy weight ×1.5
- Dungeon whispers: 4-second flavor toast on floor 1 entry (3 variants per mood)
- `go build ./...` passes; unit tests for `InferMood`, `BuildProfile`, `MoodGenModifiers`

**File envelope:**
- `src/game/behavior_tracker.go` (new)
- `src/game/dungeon_ai.go` (new)
- `src/game/metasave.go` — add `RecentBehavior []BehaviorRecord`, `PlayerProfile`
- `src/game/game.go` — start tracker on run start; finalize on run end
- `src/game/hub.go` — compute mood at run start; apply GenParamsDelta; add whisper on floor 1
- `src/levels/generate64.go` — add `RoomDensityMod`, `CorridorLengthMod` fields to GenParams
- `src/game/encounters.go` — accept enemy bias struct from dungeon mood
- `src/game/biome.go` — read mood to shift enemy pool weights

**Forbidden:** `src/coords/`, `src/collision/`, `src/entities/echo.go` (7A scope)

**Key design decisions (finalized):**
- `InferMood` priority: Spiteful (death rate >60%) > Deceptive (ranged ratio >70%) > Cautious (kill ratio <50%) > Chaotic (spell variety >4) > Neutral
- GenParams delta is additive (never replaces base, caps at ±30% deviation)
- **CorridorWidth fix IS in scope**: the hardcoded override in `generate64.go` that resets CorridorWidth to 1 is removed; Spiteful mood sets CorridorWidth to 2 for tighter corridors
- "Rooms skipped" = per-floor kill ratio (kills / total enemies spawned). Low ratio (<50%) = Cautious. No new room-visit tracking needed.
- Kill source = infer from monster Role field (`"ranged"`, `"caster"` role = ranged kill; other roles = melee). No call-site changes.
- Spell tracking = `castSpellSlot` only (7 ability slots). Arcane bolt and slash combo not counted.
- GenParamsDelta applied in `hub.go` after `BuildFloorContext` returns each floor (not in `runstate.go`)
- BehaviorRecord max: 5 entries in MetaSave as v2 field (migrateMetaSave handles nil slice)
- Whisper flavor text table defined in `dungeon_ai.go` (3 per mood × 5 moods = 15 strings)

**Tests to write:**
- `TestInferMood_SpitefulWinsOverDeceptive`
- `TestBuildProfile_WeightsMostRecentRun`
- `TestMoodGenModifiers_DeceptiveBoostsRanged`

---

## Wave 3A (parallel): Phase 8A — Item Sets

**Acceptance criteria (from `plans/8A-item-sets.md`):**
- `ItemSet` struct: ID, `[]string` member item IDs, `[]SetBonus` (pieces required, stat modifier, optional bonus ability)
- `RecalculateSetBonuses(equippedIDs []string) []ActiveSetBonus` — pure function, testable
- Set bonuses applied in `RecalculateStats()` on equip/unequip
- Three sets: Stormcaller (3pc, +20% lightning, `lightning_aura`), Fractalist (2pc, extra canopy fragment), Chainbreaker (3pc, +30% melee, `chain_sweep`, quest-locked)
- Tooltip: "Part of Stormcaller Set (1/3)" — active bonuses green
- Equipment panel: Set Bonuses section showing partial + full sets

**File envelope:** `src/items/sets.go` (new), `src/entities/player.go`, `src/ui/tooltip.go`, `src/hud/hud.go`

**Forbidden:** `src/game/hub.go`, `src/game/game.go`, `src/items/load.go`, `src/dialogue/`

**Key design decisions (finalized):**
- Set bonus abilities appended to `LearnedAbilities` (reuse existing ability system, no new field)
- **Stormcaller is 2-piece**: `item_0_26` (Rage Emblem → `lightning`) + `item_0_35` (Azazel's Pentagram → `lightning_storm`). 2pc bonus: +20% lightning damage + `lightning_aura` ability.
- **Fractalist**: `item_2_63` (Verdant Tome → `fractal_bloom`) + `item_2_55` (Necromancer's Tome → `fractal_canopy`). 2pc bonus: fractal bloom summons an additional canopy fragment.
- **Chainbreaker**: Runtime check — set bonus activates only when `MetaSave.NPCMeta["varn"].DefeatCount > 0`. Items remain droppable. Members to be chosen from Varn quest items + one melee item (agent checks `load.go` for exact IDs).
- Shop/upgrade NPC tile positions hardcoded in hub.go (same pattern as existing NPC slots)
- Upgrade key format: snake_case constants in `upgrades.go` (e.g., `"iron_constitution"`)

**Tests to write:**
- `TestRecalculateSetBonuses_PartialSetNoBonus`
- `TestRecalculateSetBonuses_FullSetAppliesBonus`
- `TestRecalculateSetBonuses_QuestLockedSetWorks`

---

## Wave 3B (parallel with 8A): Phase 9A — Transitions & Camera

**Acceptance criteria (from `plans/9A-polish-transitions-camera.md`):**
- Fade-to-black over 0.4s on floor change; fade-in on new floor; hub load fades in from black
- Screen shake: triggers on heavy hit (>20% max HP), boss phase transition, explosion spell; decays over 0.3–0.5s
- Boss intro: camera pans to boss spawn over 1.5s, holds 0.5s, then combat begins (pre-fight dialogue fires after pan)
- All effects purely cosmetic — no gameplay events during transition frames

**File envelope:** `src/game/transition.go` (new), `src/game/screenshake.go` (new), `src/game/game.go`, `src/game/draw.game.go`, `src/game/hub.go`, `src/game/boss.game.go`

**Forbidden:** `src/items/`, `src/entities/player.go`, `src/dialogue/`, `src/coords/`, `src/collision/`

**Key design decisions (finalized):**
- Shake = `shakeOffsetX/Y float64` pair on Game; injected additively in `getDrawOp` and the 2 manual Translate calls. Does NOT mutate `camX/camY`.
- Camera math: screen uses `(isoX - camX)*scale + cx` and `(isoY + camY)*scale + cy` (X negated, Y positive). Shake must respect this asymmetry.
- Transition = `transitionState int` + `transitionTimer float64` on Game (no new GameState value). `advanceFloor` deferred until fade-out completes.
- Transition overlay = pre-allocated `*ebiten.Image` on Game (never per-frame allocation).
- **Boss intro sequence**: player enters room → pre-fight dialogue fires first (existing behavior kept) → on dialogue close, camera pans to boss (1.5s lerp + 0.5s hold) → combat begins. Pan fires AFTER dialogue.
- Pan uses `lerp(current, target, dt*3.0)` on `camX/Y` toward `cartesianToIso(bx, by)` target.

**Tests to write:**
- `TestTransition_CompletesAfterDuration`
- `TestScreenShake_DecaysToZero`
- `TestScreenShake_TriggerSetsIntensity`

---

## Wave 4: Phase 8B — Hub Shop & Upgrades

**Acceptance criteria (from `plans/8B-hub-shop-upgrades.md`):**
- Shop NPC in hub (milestone-gated by `HubState["shop_unlocked"]`): scrollable item list, Remnant balance, per-run stock limits
- Upgrade station (milestone-gated by `HubState["upgrades_unlocked"]`): 8 upgrades, level pips, Remnant cost per level
- Purchased items appear in player inventory at run start
- Upgrades apply before dungeon entry via `Apply(player)`
- `MetaSave.Upgrades map[string]int` already exists (6A); `ShopPurchases []string` added here for per-run reset

**File envelope:** `src/game/shop_data.go` (new), `src/ui/shop.go` (new), `src/game/upgrades.go` (new), `src/ui/upgrade_station.go` (new), `src/game/hub.go`, `src/game/npc_data.go`, `src/game/metasave.go`

**8 starter upgrades:**
1. Iron Constitution — +20 max HP per level (3 levels)
2. Sharpened Edge — +10% melee damage per level (3 levels)
3. Deep Pockets — +1 inventory row per level (2 levels)
4. Spell Affinity — −10% mana cost per level (3 levels)
5. Mana Well — +20 max mana per level (3 levels)
6. Quick Draw — −0.1s spell cooldown per level (3 levels)
7. Scavenger — +15% Remnants earned per level (3 levels)
8. Innate Dash — start run with dash ability (1 level, replaces requiring Windrunner Boots)

**Tests to write:**
- `TestUpgrade_IronConstitution_AppliesHP`
- `TestUpgrade_InnateDash_GrantsAbility`
- `TestShopPurchases_ResetsEachRun`

---

## Wave 5: Phase 9D — HUD Navigation

**Acceptance criteria (from `plans/9D-polish-hud-navigation.md`):**
- Minimap: small overlay in corner showing explored rooms (gray squares), player dot, exit direction; toggle with M key
- Floor indicator: persistent HUD element showing "Floor 2 / 7" and biome name
- Status effect icons: row of active buff/debuff icons with remaining duration tick-down

**File envelope:** `src/hud/minimap.go` (new), `src/hud/hud.go`, `src/game/draw.game.go`

**Forbidden:** `src/game/game.go`, `src/game/hub.go` — already modified by 9A; work within hud package only

**Key design decisions (finalized):**
- Room explored check: `SeenTiles[y][x]` heuristic — room considered seen if any tile inside its rect has been seen. No new `ExploredRooms` map.
- Minimap `Draw(screen, rooms []levels.Room, seenTiles [][]bool, playerTX, playerTY, exitTX, exitTY int)` — parameters passed each frame, no pointer held.
- Player dot: `TileX/Y` integers (not interpolated).
- Status effect icons: colored 10×10 squares with 1-char abbreviation (P=poison, B=burn, S=slow, W=weaken, H=haste, +=shield) — no art needed.
- Biome name for floor indicator: `string(FloorCtx.Biome)` title-cased. `FloorCtx` passed to HUD each frame.
- Floor indicator: `"Floor %d / %d — %s"` via `RunState.CurrentFloor`, `RunState.TotalFloors`, biome name.
- `game.go` must pass `FloorCtx` and `RunState` to HUD.Draw each frame (small addition to 9A's game.go changes).

**Tests to write:**
- `TestMinimap_ShowsExploredRooms`
- `TestMinimap_ToggleOnKeyPress`

---

## Wave 6: Phase 9B — Particles & Visual Effects

**Acceptance criteria (from `plans/9B-polish-visual-effects.md`):**
- Pooled particle system (no per-frame heap alloc): pre-allocated pool of 512 particles
- Spell impact burst: fire=orange, lightning=blue/white, nature=green, chaos=purple
- Enemy death: dissolve fade over 0.5s
- Item pickup: sparkle burst (already partially exists in `render_collect.game.go` — check for conflict)
- Floating damage numbers: white=physical, spell color=spell type, yellow+large=crit

**File envelope:** `src/game/particles.go` (new), `src/game/draw.game.go`, `src/game/game.go`, `src/spells/`, `src/entities/monster.go`, `src/game/handlers.game.go`

Note: `src/game/damage_numbers.game.go` ALREADY EXISTS with a `DamageNumber` struct and `drawDamageNumbers`. Extend it — do not create a new file. Add `Type string` and `IsCrit bool` to the existing `DamageNumber` struct in `src/entities/damage_number.go`.

**Tests to write:**
- `TestParticlePool_NeverAllocatesAfterInit`
- `TestParticle_UpdateMovesPosition`

---

## Wave 7: Phase 9C — Audio

**Acceptance criteria (from `plans/9C-polish-audio.md`):**
- Ebiten audio engine (`audio/audio.go`): SFX playback, per-category volume control, max 8 concurrent sounds
- Combat SFX: hit, miss, spell cast (per type), enemy death, player damage
- Per-biome ambient loop: crossfade on floor change
- Boss music: unique track triggered on boss room entry, stops on boss defeat
- UI SFX: menu navigation, dialogue typewriter tick, purchase confirm
- All audio files stored in `assets/audio/` as `.ogg`; stubbed with `nil` checks if files absent

**File envelope:** `src/audio/audio.go` (new directory), `src/game/game.go`, `src/entities/monster.go`, `src/spells/`, `src/game/boss.game.go`, `src/ui/dialogue.go`, `src/game/hub.go`

**Note:** Audio files are NOT in scope. Stubs with nil/file-absent guards are required — the engine must not crash when `assets/audio/` is absent. Target format for future real assets: **OGG Vorbis** (`.ogg`), Ebiten's native preferred format.

**Key design decisions (finalized):**
- Audio = `*audio.Engine` field on `Game` struct (not a package-level global).
- Typewriter SFX hook inside `DialoguePanel.Update()` — compare `int(dp.TextProgress)` before/after increment.
- Boss music start: `activateBoss()` in `boss.game.go` (room entry trigger). Stop: existing `onBossDefeated()`.
- Nil-guard all audio calls: `if g.Audio != nil { g.Audio.PlaySFX(...) }`

**Tests to write:**
- `TestAudioEngine_PlayDoesNotPanic`
- `TestAudioEngine_ConcurrentLimitEnforced`

---

## Self-Review

### Spec coverage
| Plan | Covered | Notes |
|---|---|---|
| 7A echoes | ✅ | All 7.1–7.8 tasks |
| 7B dungeon AI | ✅ | All 7.9–7.14 tasks |
| 8A item sets | ✅ | All 8.1–8.4 tasks |
| 8B shop/upgrades | ✅ | All 8.5–8.14 tasks |
| 9A transitions | ✅ | All Phase 9 transition tasks |
| 9B particles | ✅ | All Phase 9 visual tasks |
| 9C audio | ✅ | All Phase 9 audio tasks |
| 9D HUD | ✅ | All Phase 9 navigation tasks |

### Wave 3 parallelism verification
8A touches: `items/sets.go`, `entities/player.go`, `ui/tooltip.go`, `hud/hud.go`
9A touches: `game/transition.go`, `game/screenshake.go`, `game/game.go`, `game/draw.game.go`, `game/hub.go`, `game/boss.game.go`
**Zero file overlap confirmed.**

### Critical path
Without audio assets, 9C will have stub implementations. This is acceptable — the architecture must be wired correctly so swapping in real `.ogg` files requires no code change, only adding files to `assets/audio/`.

### Phase 6 untested warning
7A records echoes and writes to `echoes/run_N.json`. This uses `MetaSave.CompletedRuns` as the run counter. If Phase 6 MetaSave fields are not committed and confirmed working, echo file naming will be wrong (counter stays 0). **Confirm Phase 6 is committed and passes T3 (milestone toast) before starting Wave 1.**
