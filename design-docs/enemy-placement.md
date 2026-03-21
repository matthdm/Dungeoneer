# Enemy Placement Philosophy

Enemy placement in Dungeoneer follows Dark Souls design principles: deliberate, often punishing setups that teach through repetition, reward patience, and punish recklessness. Every enemy placement should feel intentional, not random.

---

## Core Principles

### 1. Environmental Context & Lore

Enemies are placed where they *make sense*. They are not scattered randomly — they inhabit the space.

| Placement Logic | Example |
|---|---|
| Guards defend something | Enemies near locked doors, treasure rooms, or NPC locations |
| Predators stake territory | Beast-type enemies in lairs, chokepoints, dead ends |
| Scouts patrol routes | Roaming enemies in corridors and intersections |
| Ambushers hide | Enemies behind corners, in alcoves, behind pillars |
| Leaders command from advantage | Ranged enemies on elevated platforms or room edges while melee guards the approach |

### 2. The "Cruel Teacher" Approach

Each placement teaches something. The first time the player encounters a setup, it may kill them. The second time, they know what to expect.

**Teaching setups:**
- A single enemy in a wide room → teaches basic combat
- Two enemies in a narrow corridor → teaches crowd control
- A ranged enemy behind a melee blocker → teaches threat prioritization
- An ambush enemy behind a corner after a long safe hallway → teaches vigilance
- An enemy guarding an item on a ledge → teaches risk/reward evaluation

The Dungeon is not fair. But it is *consistent*. Players learn its language.

### 3. Asymmetrical Threat Pairing

Never place enemies in isolation when a more complex setup is available. Pair different threat types to create multi-dimensional encounters.

| Pairing | Player Must... |
|---|---|
| Melee + Ranged | Close gap on ranged while managing melee |
| Fast + Slow | Choose which to focus — burst the fast one or kite the slow one |
| Swarm + Elite | Clear the swarm quickly before the elite overwhelms |
| Caster + Shield | Flank the shield to reach the caster |
| Ambush + Patrol | Handle the surprise while a patrol closes in |

### 4. Exploiting Player Vulnerability

Place enemies where the player is weakest:
- **After a long corridor** — player is mid-sprint, committed to a direction
- **In rooms with one exit** — player can't retreat easily
- **Near loot** — player is distracted by greed
- **After a safe section** — player's guard is down
- **In dark corners** — FOV reveals them late

### 5. Optional Fair Solutions

Every "unfair" encounter has an answer. The player may not see it the first time, but it exists.

| Unfair Setup | Fair Solution |
|---|---|
| Ambush behind corner | Slow approach, FOV reveals shadow before enemy |
| Ranged + melee combo | Use dash to burst through melee, close on ranged |
| Room full of enemies | Pull one at a time from the doorway |
| Enemy near trap tiles | Kite enemy into their own trap |
| Elite in narrow space | Use spells for ranged damage before entering |

---

## Placement Zones

The procedural generator creates rooms and corridors. Enemy placement uses the structure of these spaces to assign roles.

### Room Zones

```
┌──────────────────┐
│    Back Edge     │  ← Ranged enemies, casters, NPC guards
│                  │
│   Center Zone    │  ← Elite enemies, bosses, quest NPCs
│                  │
│   Entry Zone     │  ← First enemies the player sees
│      ┌──┐        │
│      │  │ Door   │  ← Ambush enemies flanking doorway
└──────┴──┴────────┘
```

### Corridor Zones

```
Wall ████████████████████████████ Wall
     ← Patrol Route →
     │          │         │
   Alcove    Open      Alcove
   (ambush)  (patrol)  (ranged)
Wall ████████████████████████████ Wall
```

---

## Encounter Templates

Reusable enemy placement patterns that the procedural generator selects from.

### Solo Guardian

One enemy in the center of a room or blocking a corridor. Used for early floors and tutorial moments.

```go
EncounterTemplate{
    ID: "solo_guardian",
    MinFloor: 1,
    Enemies: []EnemySlot{
        {Role: "melee", Position: "room_center"},
    },
}
```

### Ambush Pair

Two enemies flanking a doorway. Triggered when the player enters.

```go
EncounterTemplate{
    ID: "ambush_pair",
    MinFloor: 2,
    Enemies: []EnemySlot{
        {Role: "melee", Position: "door_left", Behavior: "ambush"},
        {Role: "melee", Position: "door_right", Behavior: "ambush"},
    },
}
```

### Firing Line

Ranged enemy at the back of a room with melee blockers in front.

```go
EncounterTemplate{
    ID: "firing_line",
    MinFloor: 3,
    Enemies: []EnemySlot{
        {Role: "ranged", Position: "room_back_center"},
        {Role: "melee", Position: "room_mid_left"},
        {Role: "melee", Position: "room_mid_right"},
    },
}
```

### Corridor Patrol

A roaming enemy in a long corridor. Player may encounter them head-on or from behind.

```go
EncounterTemplate{
    ID: "corridor_patrol",
    MinFloor: 1,
    Enemies: []EnemySlot{
        {Role: "melee", Position: "corridor_patrol", Behavior: "roaming"},
    },
}
```

### Swarm Room

A room with many weak enemies. Tests AoE and crowd control.

```go
EncounterTemplate{
    ID: "swarm_room",
    MinFloor: 4,
    Enemies: []EnemySlot{
        {Role: "swarm", Position: "room_scattered", Count: 5},
    },
}
```

### Elite + Adds

One strong enemy with supporting weaker enemies.

```go
EncounterTemplate{
    ID: "elite_and_adds",
    MinFloor: 5,
    Enemies: []EnemySlot{
        {Role: "elite", Position: "room_center"},
        {Role: "melee", Position: "room_edges", Count: 2},
    },
}
```

### Trap Room

Enemies positioned near environmental hazards. Player must fight without stepping on traps.

```go
EncounterTemplate{
    ID: "trap_room",
    MinFloor: 4,
    Enemies: []EnemySlot{
        {Role: "ranged", Position: "behind_traps"},
        {Role: "melee", Position: "flanking_traps"},
    },
    RequiresHazards: true,
}
```

---

## Encounter Data Model

```go
type EncounterTemplate struct {
    ID              string
    MinFloor        int            // earliest floor this can appear
    MaxFloor        int            // 0 = no limit
    MinRoomSize     int            // minimum room dimension to fit this encounter
    Enemies         []EnemySlot
    RequiresHazards bool           // room must have trap tiles
    Biomes          []string       // restrict to specific biomes ("" = any)
    Weight          float64        // selection probability weight
}

type EnemySlot struct {
    Role       string // "melee", "ranged", "elite", "swarm", "caster", "ambush"
    Position   string // "room_center", "room_back_center", "door_left", "corridor_patrol", etc.
    Behavior   string // "idle", "ambush", "roaming", "patrol"
    Count      int    // for multi-spawn slots (default 1)
    LevelBonus int    // added to floor-based enemy level
}
```

### Enemy Role → Monster Mapping

The `Role` field maps to actual monster types at runtime based on biome:

| Role | Crypt Monster | Lair Monster | Gehenna Monster |
|---|---|---|---|
| melee | Skeleton Warrior | Wolf | Imp |
| ranged | Bone Archer | Spitting Spider | Fire Elemental |
| elite | Death Knight | Alpha Beast | Infernal Guard |
| swarm | Bone Rats | Bats | Ember Sprites |
| caster | Lich Acolyte | Druid | Flame Sorcerer |
| ambush | Coffin Mimic | Lurking Stalker | Lava Lurker |

---

## Floor-Level Placement Algorithm

When a floor is generated, enemies are placed in this order:

```
1. Identify all rooms and corridors
2. Classify rooms by size (small, medium, large)
3. Assign encounter templates based on:
   - Floor number (difficulty gating)
   - Room size (encounter must fit)
   - Biome (biome-restricted encounters)
   - Budget (total enemy count per floor capped by difficulty)
4. Place enemies at template-specified positions within rooms
5. Place corridor patrols in qualifying corridors
6. Validate: no enemy spawns on unwalkable tiles, within 3 tiles of player spawn,
   or in rooms designated for NPCs
7. Set enemy levels based on floor difficulty
```

### Enemy Budget

Total enemies per floor scale with difficulty:

| Floor Position | Budget |
|---|---|
| Floor 1 | 8-12 enemies |
| Mid floors | 15-20 enemies |
| Pre-boss floor | 20-25 enemies |
| Boss floor | Boss + 5-8 arena adds |

---

## Integration with Living Dungeon AI

The Living Dungeon AI (see [living-dungeon-ai.md](living-dungeon-ai.md)) modifies placement based on player behavior:

| Player Behavior | AI Response |
|---|---|
| Always rushes through | More ambush placements, corridor traps |
| Kites from range | Faster enemies, gap-closers |
| Relies on AoE spells | Spread-out formations, anti-AoE positioning |
| Methodical, slow | Timed encounters, patrol convergence |
| Avoids certain enemies | More of those enemies, front-loaded |

---

## Implementation Notes

### Phase 1 (Basic Placement)
1. Simple per-room enemy spawning: 1-3 enemies per room based on floor number
2. Use existing Ambush and Roaming behaviors
3. Enemy count scales with floor difficulty
4. No encounter templates yet — simple count-based placement

### Phase 2 (Template-Based Placement)
1. Define 5-6 encounter templates
2. Room classification (small/medium/large)
3. Template selection based on floor and room size
4. Position mapping within rooms (center, edges, doorways)
5. Enemy role → monster type mapping per biome

### Phase 3 (Intelligent Placement)
1. Full encounter template library (15+ templates)
2. Living Dungeon AI integration
3. Difficulty balancing and budget system
4. Trap tile integration for hazard encounters
5. Quest-aware placement (NPC rooms cleared of hostiles, or guarded)
