# Meta Progression & Death

Meta progression is the layer of persistence that sits above individual runs. Within a run, the player accumulates power through items, levels, and stats. On death or victory, **all of that resets**. What persists is meta-progression: currency, upgrades, NPC memory, echoes, and dungeon intelligence.

---

## Death Philosophy

Death is a full reset. The player loses:
- All gold
- All items and equipment
- All stats and levels
- All quest flag progress within the run

Death is **not punishing** — it is **clarifying**. The player failed this attempt. The Dungeon remembers. The player returns to the hub and tries again, now informed by what they learned.

The only things that survive death are the things that exist outside the run.

---

## Meta Currency: Remnants

**Remnants** are the meta currency. They are the residue of experience — what the Dungeon allows you to keep.

### Earning Remnants

| Source | Amount | Notes |
|---|---|---|
| Enemy killed | 1-3 | Scales with enemy level |
| Floor cleared | 10-20 | Bonus for clearing all rooms |
| NPC encountered | 5 | First encounter per NPC per run |
| Quest phase completed | 15-25 | Major NPC quest milestones |
| Boss defeated | 50-100 | Victory reward |
| Run completed | 25 | Flat completion bonus |
| Death | floor * 5 | Partial reward based on progress |

### Spending Remnants

Remnants are spent in the hub at:
1. **Shop** — consumables, starter items (see [hub-world.md](hub-world.md))
2. **Upgrade Station** — permanent upgrades (see below)

Remnants are stored in the meta save and persist indefinitely.

```go
type MetaSave struct {
    Remnants       int
    TotalRemnants  int               // lifetime earned (for milestones)
    RunCount       int
    CompletedRuns  int
    TotalDeaths    int
    Upgrades       map[string]int    // upgrade_id → level
    NPCMeta        map[string]*NPCMetaState
    EchoRecords    []EchoRecord
    PlayerProfile  *PlayerProfile    // for Living Dungeon AI
    LoreUnlocked   []string
    HubState       *HubState
}
```

---

## Permanent Upgrades

Purchased at the Upgrade Station in the hub. These modify the player's starting state for all future runs.

### Upgrade Tree

Upgrades are organized into categories. Each upgrade has multiple levels with increasing cost.

#### Vitality Branch

| Upgrade | Levels | Effect Per Level | Cost |
|---|---|---|---|
| Iron Constitution | 3 | +10 starting MaxHP | 30, 60, 120 |
| Second Wind | 2 | Slow HP regen when below 25% HP | 50, 100 |
| Thick Skin | 3 | +1 damage reduction (flat) | 40, 80, 160 |

#### Combat Branch

| Upgrade | Levels | Effect Per Level | Cost |
|---|---|---|---|
| Sharpened Edge | 3 | +2 starting damage | 30, 60, 120 |
| Quick Hands | 2 | -5 starting attack rate (faster) | 50, 100 |
| Spell Affinity | 3 | -0.5s spell cooldown reduction | 40, 80, 160 |

#### Utility Branch

| Upgrade | Levels | Effect Per Level | Cost |
|---|---|---|---|
| Deep Pockets | 2 | +1 inventory row | 60, 150 |
| Dash Recovery | 2 | -1s dash recharge time | 40, 80 |
| Grapple Reach | 2 | +2 grapple range | 40, 80 |

#### Fortune Branch

| Upgrade | Levels | Effect Per Level | Cost |
|---|---|---|---|
| Scavenger | 3 | +15% Remnant gain | 50, 100, 200 |
| Lucky Find | 2 | +10% item drop rate | 60, 120 |
| Echo Attunement | 2 | +1 max echo spawns per run | 75, 150 |

#### Dungeon Branch

| Upgrade | Levels | Effect Per Level | Cost |
|---|---|---|---|
| Cartographer | 1 | Floor exits glow faintly through fog | 100 |
| Deep Descent | 3 | +1 floor per run (extends run length) | 80, 160, 320 |
| Dungeon Sense | 1 | Hidden rooms have a faint visual tell | 150 |

### Upgrade Data Model

```go
type UpgradeDef struct {
    ID          string
    Name        string
    Description string
    Category    string   // "vitality", "combat", "utility", "fortune", "dungeon"
    MaxLevel    int
    Costs       []int    // cost per level
    Prereqs     []string // upgrade IDs required first
    Apply       func(player *Player, level int)
}
```

Upgrades are applied at run start, before the player enters the dungeon:

```go
func ApplyUpgrades(player *Player, upgrades map[string]int) {
    for id, level := range upgrades {
        def := UpgradeRegistry[id]
        if def != nil && level > 0 {
            def.Apply(player, level)
        }
    }
}
```

---

## Run Loadout

Before entering the dungeon, the player can spend Remnants at the shop for **run consumables** — items that go into inventory at the start of the run.

These are one-time purchases per run (not permanent). The player chooses what to bring:

| Item | Cost | Effect |
|---|---|---|
| Health Potion | 5 | Restores 50 HP |
| Mana Elixir | 5 | Restores 30 Mana |
| Iron Key | 10 | Opens one locked door |
| Recall Stone | 15 | Teleport to floor entrance |
| Echo Shard | 20 | Choose echo type on next death |

Run loadout items are lost on death like everything else.

---

## NG+ Memory

The most narratively important meta-progression. NPCs remember across runs.

### NPC Meta State

```go
type NPCMetaState struct {
    Met            bool
    DefeatCount    int    // times defeated as boss
    HighestPhase   int    // furthest quest phase reached
    TotalTrust     int    // cumulative trust across all runs
    LastDefeatRun  int    // which run number they were last defeated
    Betrayed       bool   // player ever betrayed this NPC
}
```

### NG+ Dialogue Effects

| Condition | Dialogue Change |
|---|---|
| First meeting after defeat | NPC recognizes player, references past fight |
| Defeated 2+ times | NPC's philosophy shows cracks, self-doubt |
| Defeated 3+ times | NPC questions the Dungeon itself, approaches meta-awareness |
| Betrayed then re-encountered | NPC is hostile or suspicious, requires more effort to build trust |
| High cumulative trust | NPC opens up about deeper motivations, unlocks lore |

Example NG+ dialogue for Varn after 2 defeats:

> "You again. I remember the chains breaking. Twice now."
> "Tell me — when you toppled me, did you feel order return? Or just more chaos?"
> "Perhaps my chains were never the answer. But I don't know another way."

---

## Milestone Unlocks

Certain thresholds unlock new content:

| Milestone | Unlock |
|---|---|
| First run completed | Shop opens in hub |
| 3 runs completed | Upgrade Station activates |
| First death | Echo Shrine appears |
| Defeat first boss | NG+ dialogue for that NPC |
| 5 total runs | New minor NPC pool |
| 10 total runs | Abaddon appears in hub |
| Defeat all major NPCs (1 each) | Cosmological lore entries |
| 50 total Remnants earned | Fortune upgrade branch |
| 100 total Remnants earned | Dungeon upgrade branch |

---

## Data Persistence

Meta save is stored as a single JSON file:

```
saves/
  meta.json           ← meta-progression (permanent)
  echoes/
    echo_001.json     ← echo recordings
    echo_002.json
```

The meta save is **never deleted** (unless the player explicitly resets). It grows over time, accumulating NPC memory, echoes, and upgrade state.

---

## Implementation Notes

### Phase 1 (Basic Meta)
1. `MetaSave` struct with Remnants, RunCount, Upgrades
2. Save/load meta.json on run end
3. Remnant calculation on death and victory
4. Display Remnant balance in hub HUD
5. 3-5 starter upgrades in the Upgrade Station

### Phase 2 (Full Economy)
1. Shop with run loadout items
2. Full upgrade tree (15+ upgrades across 5 categories)
3. Milestone unlock system
4. NPC meta state tracking

### Phase 3 (NG+ Narrative)
1. NG+ dialogue trees for all major NPCs
2. Boss mutations based on defeat count
3. Abaddon meta-NPC appearances
4. Cumulative trust effects on NPC behavior
