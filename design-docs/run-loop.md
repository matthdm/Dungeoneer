# Run Loop & Floor Progression

The run loop is the core gameplay cycle. Everything else — NPCs, bosses, echoes, dungeon AI — lives inside this structure.

---

## Run Structure

A **run** is a single attempt at the dungeon. The player enters from the hub, descends through procedurally generated floors, and either dies or defeats the final boss.

```
Hub World
  │
  ▼
Floor 1 (entry)
  │
Floor 2
  │
  ...
  │
Floor N (boss floor)
  │
  ▼
Boss Encounter
  │
  ├─ Victory → Return to Hub (with meta rewards)
  └─ Death   → Return to Hub (with partial meta rewards)
```

---

## Variable Run Length

Run length is **not fixed**. It scales with meta-progression.

| Progression Stage | Floor Count | Unlocked By |
|---|---|---|
| First runs | 3-4 floors | Default |
| Early progression | 5-6 floors | Completing first run |
| Mid progression | 7-8 floors | Defeating N unique bosses |
| Late progression | 9-10+ floors | NG+ cycles, dungeon AI evolution |

The final floor is always the **boss floor**. Additional floors mean more biome variety, more NPC encounters, and more enemy complexity — not padding.

---

## Floor Generation

Each floor is generated on-the-fly when the player reaches the exit of the previous floor. The generator receives a `FloorContext`:

```go
type FloorContext struct {
    FloorNumber    int            // 1-indexed
    TotalFloors    int            // how many floors this run
    RunSeed        int64          // deterministic seed for the run
    Biome          string         // assigned biome for this floor
    Difficulty     float64        // 0.0-1.0, scales with floor number
    QuestFlags     map[string]int // active quest state (affects NPC spawns)
    PlayerProfile  *PlayerProfile // from Living Dungeon AI (future)
    DungeonMood    *DungeonMood   // from Living Dungeon AI (future)
}
```

### Difficulty Scaling

Difficulty affects `GenParams` and enemy spawning:

| Parameter | Floor 1 | Mid Floor | Final Floor |
|---|---|---|---|
| Room count | 6-8 | 10-14 | 14-18 |
| Enemy density | Low | Medium | High |
| Enemy level | 1-2 | scales | player_level + 1-2 |
| Locked door chance | 0.1 | 0.25 | 0.4 |
| Trap density | None | Low | Medium |
| Loot quality | Common | Common-Uncommon | Uncommon-Rare |

---

## Floor Traversal

### Exit Entity

Each floor contains an **exit entity** — a stairwell, portal, or rift that advances the player to the next floor.

```go
type FloorExit struct {
    X, Y       int
    SpriteID   string
    Locked     bool       // may require clearing enemies or a key
    NextBiome  string     // preview of what's below
}
```

Exit placement rules:
- Placed in the room **farthest from the player spawn** (maximizes exploration)
- May be locked until a condition is met (clear all enemies in room, find a key, defeat a miniboss)
- Visual hint distinguishes it from regular stairwells (glow, particle effect)

### Spawn Point

Player spawns at a designated point on each floor:
- Floor 1: center of the first room generated
- Subsequent floors: near the entry stairwell (reverse of previous exit)

---

## Death Handling

When the player dies:

1. **Run ends immediately** — no revive, no second chance
2. **Record echo data** — player path, actions, and death location are saved for Echoes of Self
3. **Award meta currency** — based on floors cleared, enemies killed, NPCs encountered
4. **Display death screen** — run summary with stats
5. **Return to hub** — player spawns at hub with full reset (no gold, items, or stats carry over)

### Death Screen

The death screen shows:
- Floors cleared (e.g., "Reached Floor 6 of 8")
- Enemies defeated
- NPCs encountered and choices made
- Cause of death
- Meta currency earned
- Echo spawned (if applicable)

---

## Victory Handling

When the player defeats the boss on the final floor:

1. **Boss defeat dialogue** — philosophical exchange based on which NPC ascended
2. **Award meta currency** — bonus for completion
3. **Record run completion** — increments run count, logs NPC defeat
4. **NG+ flag** — next run may have modified dialogue, harder floors, new NPC phases
5. **Return to hub** — with completion rewards

---

## Run State

The run maintains persistent state across floors within a single run:

```go
type RunState struct {
    RunID          string
    Seed           int64
    CurrentFloor   int
    TotalFloors    int
    FloorsBiomes   []string         // pre-assigned biome sequence
    QuestFlags     map[string]int   // NPC interaction state
    EnemiesKilled  int
    ItemsCollected []string
    NPCsMet        []string
    MetaCurrency   int              // accumulated this run
    StartTime      time.Time
    PlayerSnapshot PlayerSave       // snapshot at run start (for echo)
    ActionLog      []RecordedAction // for echo recording
}
```

This state is **not saved to disk** during a run (roguelike — no save-scumming). It exists only in memory. If the game crashes or is closed, the run is lost.

---

## Enemy Placement Philosophy

Enemy placement follows Dark Souls design principles. See [enemy-placement.md](enemy-placement.md) for the full philosophy.

Key integration points with the run loop:
- Floor difficulty scales enemy count and level
- Biome determines enemy pool
- Procedural generator places enemies using placement rules (ambush points, chokepoints, threat pairing)
- Quest flags may add or remove specific enemies (NPC-related encounters)

---

## Floor Transition Flow

```
Player reaches exit entity
  │
  ├─ Exit locked? → Must fulfill condition (clear room, find key)
  │
  ▼
Fade to black
  │
  ▼
Generate next floor (FloorContext)
  │
  ▼
Place player at spawn point
  │
  ▼
Initialize FOV, spawn enemies, place NPCs
  │
  ▼
Fade in → Resume play
```

---

## Integration Points

| System | How It Connects |
|---|---|
| Hub World | Entry point for runs, return point on death/victory |
| Biome System | Assigns biome per floor, affects generation params |
| NPC System | Quest flags determine NPC spawns per floor |
| Boss System | Final floor generates boss arena based on ascension state |
| Enemy Placement | Floor difficulty drives placement rules |
| Echoes of Self | Death triggers echo recording |
| Living Dungeon AI | Player profile modifies floor generation |
| Meta Progression | Run rewards feed hub upgrades |

---

## Implementation Notes

### Phase 1 (Minimum Viable Run)
1. Hardcode 3-floor run
2. Generate each floor with increasing `GenParams` difficulty
3. Place exit entity in farthest room
4. On exit interaction: generate next floor, switch level
5. Floor 3: spawn a placeholder boss
6. Death: return to main menu (hub comes later)
7. Victory: return to main menu

### Phase 2 (Full Run Loop)
1. Variable floor count based on meta-progression
2. Biome assignment per floor
3. NPC spawn integration
4. Death screen with stats
5. Hub world integration
6. Meta currency rewards
