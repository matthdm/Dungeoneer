# Combat Redesign — Design Spec

**Date:** 2026-06-28
**Status:** Approved
**Scope:** Full combat system redesign — GW1/D2-style targeting, auto-attack loop, artifact library, combat engine extraction, and benchmarker utility.

---

## Motivation

Current combat is stale: no targeting system, no auto-approach, melee has only `slash_combo`, and the kill loop lacks the feedback density that drives replayability. The goal is to produce a dopamine-driven farming loop — targeted item acquisition, build crafting around artifacts, and satisfying moment-to-moment kill feel — without competing with action-game combat depth. The model is Guild Wars 1: simple baseline with deep build space.

---

## Architecture: Combat Engine as Library

### Core Principle

Combat logic is extracted into `src/combat/` — a pure Go package with zero Ebiten dependencies. The game loop and the benchmarker are both consumers of this package. Changing a damage formula in `src/combat/` is immediately reflected in both.

### Package Structure

```
src/combat/
    engine.go       ← CombatEngine interface + tick advancement
    state.go        ← CombatState (pure data, no Ebiten)
    actions.go      ← Action types (what can be applied per tick)
    skills.go       ← Skill/artifact execution logic
    sim.go          ← Headless simulation runner, seeded RNG
    metrics.go      ← DPS, survivability, clear speed, rotation efficiency

src/game/
    combat_adapter.go   ← CombatAdapter interface + toggle routing
    legacy_adapter.go   ← Wraps existing handlers.game.go (untouched)
    new_adapter.go      ← Routes to src/combat engine

cmd/benchmarker/
    main.go             ← Standalone binary entry point
    ui.go               ← Terminal or simple Ebiten UI
    scenarios/          ← JSON scenario definitions
```

### CombatAdapter Interface

```go
// src/game/combat_adapter.go
type CombatAdapter interface {
    HandleTargetSelect(g *Game, worldX, worldY float64)
    HandleMoveToAttack(g *Game)
    ProcessTick(g *Game, dt float64)
    HandleSkillActivation(g *Game, slotIdx int)
}
```

The game loop holds a `CombatAdapter` field. At startup, `dev_settings.json` determines which implementation is loaded. `LegacyCombatAdapter` wraps existing code untouched. `NewCombatAdapter` routes to `src/combat`.

### Toggle

```json
// src/dev_settings.json (gitignored, dev-only)
{ "use_legacy_combat": false }
```

Accessible from the debug overlay (F2 menu) at runtime. Legacy combat stays fully intact; switching is instantaneous with a game restart.

---

## Section 1: Targeting Model

### Inputs

| Input | Action |
|-------|--------|
| Left-click on enemy | Select that enemy as target |
| `C` key | Select nearest enemy (screen-distance) |
| Left-click on empty space | Deselect target |
| Spacebar | Move-to-attack current target (or interact if no target) |

### Target State

`Game` gains:
```go
TargetedMonster *entities.Monster  // nil = no target
```

Only one target at a time. AOE skills still hit multiple enemies — targeting determines the *primary* target for auto-attack and single-target skills, not the hit zone.

### Visual Indicator

A pulsing ring drawn under the targeted monster's feet. Color: red. Drawn at world position before the entity sprite so it appears on the floor. Reuses existing `vector.DrawFilledCircle` calls.

### Target Display (top of screen)

When `TargetedMonster != nil`, a panel renders at top-center:
- Enemy name (left)
- Level badge (e.g. `Lv.4`)
- Health bar (fills left to right, red, shows current/max HP)
- Elite indicator if applicable

Panel fades out 2 seconds after the target dies or is deselected.

---

## Section 2: Auto-Attack Loop

### Move-to-Attack

Spacebar with a live target: if the player is outside attack range, pathfind toward the target. Auto-attack begins the moment the player enters range. Pathfinding uses the existing A* system — no new pathing code.

### Attack Range

Determined by class (weapon):

| Class | Attack type | Range (tiles) | Attack interval |
|-------|-------------|---------------|-----------------|
| Knight | Melee | 1.5 | 0.8s base |
| Mage | Ranged (arcane bolt) | 6.0 | 1.0s base |

Range and interval are stats on the player — artifacts can modify them.

### Auto-Attack Execution

Once in range, the engine fires auto-attacks on a timer (`attackTimer float64` on CombatState). Each auto-attack:
1. Checks line-of-sight to target (existing FOV system)
2. Applies base weapon damage + relevant stat modifiers
3. Triggers `on_hit` item effects
4. Resets `attackTimer`

Auto-attack stops when: target dies, target is deselected, player moves out of range, player is CCed.

### Mage Range Bubble

When Mage has a target and is approaching, draw a semi-transparent circle at player position showing attack range (6 tiles radius). Fades when in range. Uses existing `vector.StrokeCircle`.

### Knight Auto-Attack Feel

Knight auto-attack is the backbone — skills amplify it. The auto-attack must feel weighty:
- Slight screen nudge toward target on swing (0.5px shake, 1 frame)
- Enemy flinch sprite flash on hit (white flash 1 frame, already exists for damage)
- Hit sound (existing)

---

## Section 3: Kill Feel (Dopamine Layer)

These systems live in `src/combat/` and fire through events the game loop consumes.

### Floating Damage Numbers

`src/combat/` emits a `DamageEvent` (value, world position, is_crit) — no rendering types in the engine. The game layer converts these into visual floats:

```go
// src/game/damage_number.go
type DamageNumber struct {
    Value    int
    X, Y     float64
    Life     float64  // seconds remaining
    IsCrit   bool
}
```
Numbers float upward and fade over 0.8s. Crits are larger and gold-colored. Stored in `Game.DamageNumbers []DamageNumber`, drawn in `draw.game.go` after entity layer.

### Kill Momentum

On enemy kill:
- `KillStreak int` increments, resets on taking damage
- At streak 3/6/10: brief screen vignette pulse (dark flash, 0.1s)
- Kill streak count shown briefly near the kill counter HUD element
- Streak > 3: auto-attack interval reduced by 5% per additional kill (caps at 25% reduction), resets on damage taken

### Gold/Drop Pop Animation

On enemy death, gold amount floats up in yellow text (same `DamageNumber` system, `IsCrit = false`, gold color). Item drops get a brief glow pulse on the item sprite for 1.5s.

### Kill Counter

Existing kill counter in HUD gains a tick animation on increment (scale up 1.1x → 1.0x over 0.15s).

---

## Section 4: Artifact Library

### Philosophy

Artifacts ARE skills. Finding the Chaos Emblem IS how you get Chaos Ray. Artifacts persist across runs in `MetaSave.ArtifactCollection []string` (artifact IDs). The Artifact Library is a hub UI that shows everything you own, filterable by domain.

### Artifact Acquisition

| Source | Behavior |
|--------|----------|
| Quest-unlocked drops | Item enters drop pool only after a quest flag is set |
| Boss greens | Unique artifact per boss, ~30% drop chance per kill |
| Event room rewards | Specific events always award specific artifacts |
| Common in-run drops | Low-tier artifacts, always in pool |

### Pre-Run Loadout

In the hub, before starting a run: a loadout screen lets the player equip up to **6 artifact slots** from their collection. This replaces the current random ability item floor drops as the *primary* source of abilities. Slot count is fixed at 6 for now (matches existing HUD skill slots).

One slot is the **Elite Artifact slot** (slot 6): only artifacts marked `"elite": true` can go here. These are the most powerful and are boss greens or end-of-quest rewards.

### Mid-Run Artifact Swaps (Hybrid)

During a run:
- Event rooms can offer a temporary artifact add (occupies a bonus 7th slot, run-only, shown with a distinct gold background in the HUD)
- Boss kills can drop their green into your active run loadout (added as the 7th slot or replaces a chosen slot — player is prompted)
- Trial artifacts (items you don't own) are usable for the run, but not added to collection — you must hunt them properly to own them. Trial artifacts show a `TRIAL` badge on the HUD slot icon

### Artifact Library UI

Located in the hub (L key, or pedestal interaction). GW1-style grid:
- Icon (existing item art)
- Name
- Domain tag (e.g. `shadow`, `flame`, `void`, `nature`)
- Effect one-liner
- Lock icon if not yet unlocked
- Filter tabs: All / Domain / Elite / Unowned

### Existing Items → Artifacts

All 14 existing ability items (`item_2_*`) are retroactively reframed as artifacts. Their `GrantsAbility` field remains the skill they provide. No data migration needed — they already have IDs, icons, and effects.

New field on `ItemTemplate`:
```go
IsArtifact  bool   `json:"is_artifact"`
ArtifactDomain string `json:"artifact_domain"`
IsElite     bool   `json:"is_elite"`
```

---

## Section 5: Benchmarker

### Purpose

A standalone application that imports `src/combat/` and runs build simulations headlessly. Gives the developer full "pit boss" visibility into build performance without manual playtesting.

### Standalone Binary

```
cmd/benchmarker/main.go
```

Built with `go build ./cmd/benchmarker/`. Runs independently of the game.

### In-Game Access

Accessible from the debug overlay (F2 → "Open Benchmarker"). Launches as a separate window or embedded panel showing live sim output.

### Scenario Format

```json
// cmd/benchmarker/scenarios/knight_slash_build.json
{
  "name": "Knight Slash Build — Floor 5",
  "class": "knight",
  "artifacts": ["shadow_blade", "iron_resolve", "ash_pauldrons"],
  "stats": { "Strength": 18, "Vitality": 12 },
  "floor": 5,
  "biome": "crypt",
  "enemy_pool": ["melee", "ranged", "elite"],
  "iterations": 1000,
  "skill_rotation": ["slot_1", "slot_3", "auto", "slot_1", "auto", "auto"]
}
```

### Metrics Output

Per scenario run, the sim collects:
```go
type SimResult struct {
    AvgDPSTick       float64
    AvgClearTimeSec  float64
    SurvivalRate     float64  // % of iterations player survived
    AvgDamageTaken   float64
    KillsPerMinute   float64
    SkillUsageRatio  map[string]float64
    StreakAvg        float64
}
```

Output: JSON to stdout (pipe to file), or rendered as a table in the benchmarker UI.

### Go Benchmark Integration

```go
// src/combat/sim_bench_test.go
func BenchmarkKnightSlashBuild(b *testing.B) {
    scenario := LoadScenario("knight_slash_build.json")
    for i := 0; i < b.N; i++ {
        RunSimulation(scenario)
    }
}
```

`go test -bench=. ./combat/...` produces the pit boss report.

---

## Section 6: HUD Changes

### Skill Bar

6 slots remain. Slot 6 is visually distinct (gold border = Elite slot). Slots show artifact icon, not ability icon — same art, different framing.

Pre-run: slots show the equipped loadout. In-run: slots show cooldown states, mana costs.

### Target Display Panel

Top-center, 280×36px panel:
```
[ Enemy Name        Lv.4  [========  ] 68/100 HP ]
```
Health bar color: red → dark red as HP drops. Fades 2s after death.

### Kill Streak Counter

Bottom-right, below kill count. Shows current streak as `x3`, `x6`, `x10` with color escalation (white → orange → red).

---

## New Melee Skills (Knight)

These are the first wave of artifact-backed melee skills to address the "stale melee" problem. Each is a new artifact with a `GrantsAbility` pointing to a new skill implementation.

| Artifact | Domain | Effect |
|----------|--------|--------|
| Ironbreaker Gauntlets | iron | Slam — AoE melee, 1.5 tile radius, 3s CD |
| Shroud Cloak | shadow | Shadowstep — blink behind target, next auto crits, 8s CD |
| Warden's Medallion | nature | Taunt — all nearby enemies target you for 4s, you gain 30% damage reduction |
| Ashbound Chain | void | Bind — root target in place for 2s, 10s CD |
| Grave Reaper | void | Execute — instant kill below 20% HP, 20s CD |
| Ember Mantle | flame | Ignite — auto-attacks apply burn (5 dmg/s for 3s) passively |

---

## Out of Scope

- Secondary professions / hybrid class system (future)
- Summoner, assassin, ranger classes (future)
- Full dungeon meta AI (future — benchmarker is the prerequisite)
- Skill synergy system (future — build space first)
- Artifact crafting or upgrading
- Multiplayer
