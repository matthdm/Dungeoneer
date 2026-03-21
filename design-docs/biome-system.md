# Biome System

Biomes define the visual, thematic, and mechanical identity of each dungeon floor. They determine what the player sees, what enemies they fight, what loot they find, and what the atmosphere feels like.

---

## Biome Definitions

| Biome | Theme | Atmosphere | Tilesheet |
|---|---|---|---|
| Crypt | Necromancy, undeath | Cold stone, flickering torches, coffins | `crypt.png` |
| Lair | Beasts, predators | Organic walls, bones, nests, dim light | `lair.png` |
| Catacomb | Tombs, buried history | Ancient brick, dusty air, narrow passages | `catacomb.png` |
| Moss | Nature reclaiming | Overgrown stone, green light, water | `moss.png` |
| Gehenna | Hellfire, punishment | Cracked stone, lava glow, heat distortion | `gehena.png` |
| Pandemonium | Chaos, madness | Shifting geometry, impossible colors, visual noise | `pandem*.png` |

### Future Biomes

| Biome | Theme | Unlock |
|---|---|---|
| Cocytus | Frozen despair | NG+ cycle 2 |
| Lapis | Crystalline knowledge | Encounter knowledge-philosophy NPC |
| Hive | Collective will, swarm | Mid-progression |
| Gallery | Curated memories, art | High lore collection |
| Tunnel | Transition, liminal | Between major biome sets |

---

## Biome → GenParams Mapping

Each biome modifies the procedural generation parameters to create distinct spatial feelings.

```go
type BiomeConfig struct {
    ID             string
    Name           string
    WallFlavor     string          // maps to spritesheet flavor
    FloorFlavor    string
    GenOverrides   GenParamOverrides
    EnemyPool      []EnemyDef
    LootTable      *LootTable
    AmbientColor   color.RGBA      // tint overlay
    FOVRadius      float64         // override default FOV range
    MusicTrack     string          // ambient music ID
}

type GenParamOverrides struct {
    RoomCountMin   *int
    RoomCountMax   *int
    RoomWMin       *int
    RoomWMax       *int
    RoomHMin       *int
    RoomHMax       *int
    CorridorWidth  *int
    DoorLockChance *float64
    CoverageTarget *float64
}
```

### Parameter Profiles

| Biome | Room Count | Room Size | Corridors | Doors | Coverage | Feel |
|---|---|---|---|---|---|---|
| Crypt | 8-12 | Medium (6-10) | Narrow (1) | Many, often locked | 0.40 | Claustrophobic, structured |
| Lair | 6-10 | Large (8-14) | Wide (2) | Few | 0.50 | Open, dangerous |
| Catacomb | 10-16 | Small (4-8) | Narrow (1) | Many | 0.35 | Maze-like, dense |
| Moss | 8-12 | Mixed (5-12) | Medium (1-2) | Few | 0.55 | Natural, flowing |
| Gehenna | 6-8 | Large (10-16) | Narrow (1) | Few, always locked | 0.45 | Arena-like, hot |
| Pandemonium | 10-14 | Random (3-15) | Random (1-3) | Random | 0.50 | Unpredictable |

---

## Enemy Pools

Each biome defines which enemy types spawn there. Enemy roles (melee, ranged, elite, etc.) map to biome-specific monsters.

```go
type EnemyDef struct {
    ID          string
    Name        string
    Role        string   // "melee", "ranged", "elite", "swarm", "caster", "ambush"
    SpriteID    string
    BaseHP      int
    BaseDamage  int
    BaseLevel   int
    Behavior    string   // "ambush", "roaming", "patrol", "ranged"
    SpellIDs    []string // for casters
}
```

### Crypt Enemy Pool

| Role | Enemy | HP | Damage | Behavior |
|---|---|---|---|---|
| melee | Skeleton Warrior | 30 | 8 | roaming |
| ranged | Bone Archer | 20 | 6 | ranged |
| elite | Death Knight | 80 | 15 | ambush |
| swarm | Bone Rats | 8 | 3 | swarm |
| caster | Lich Acolyte | 25 | 10 | ranged (spells) |
| ambush | Coffin Mimic | 40 | 12 | ambush |

### Lair Enemy Pool

| Role | Enemy | HP | Damage | Behavior |
|---|---|---|---|---|
| melee | Wolf | 25 | 10 | roaming |
| ranged | Spitting Spider | 18 | 5 | ranged |
| elite | Alpha Beast | 100 | 18 | patrol |
| swarm | Bats | 5 | 2 | swarm |
| caster | Druid | 30 | 8 | ranged (spells) |
| ambush | Lurking Stalker | 35 | 14 | ambush |

### Catacomb Enemy Pool

| Role | Enemy | HP | Damage | Behavior |
|---|---|---|---|---|
| melee | Tomb Warden | 35 | 9 | patrol |
| ranged | Spirit Archer | 22 | 7 | ranged |
| elite | Entombed Champion | 90 | 16 | ambush |
| swarm | Scarabs | 6 | 2 | swarm |
| caster | Cursed Priest | 28 | 11 | ranged (spells) |
| ambush | Wall Crawler | 30 | 10 | ambush |

### Moss Enemy Pool

| Role | Enemy | HP | Damage | Behavior |
|---|---|---|---|---|
| melee | Vine Brute | 40 | 10 | roaming |
| ranged | Thorn Spitter | 20 | 6 | ranged |
| elite | Ancient Treant | 120 | 14 | patrol |
| swarm | Sporelings | 8 | 3 | swarm |
| caster | Moss Shaman | 25 | 9 | ranged (spells) |
| ambush | Root Trap | 45 | 8 | ambush (stationary) |

### Gehenna Enemy Pool

| Role | Enemy | HP | Damage | Behavior |
|---|---|---|---|---|
| melee | Imp | 20 | 12 | roaming |
| ranged | Fire Elemental | 30 | 10 | ranged |
| elite | Infernal Guard | 100 | 20 | patrol |
| swarm | Ember Sprites | 6 | 4 | swarm |
| caster | Flame Sorcerer | 25 | 15 | ranged (spells) |
| ambush | Lava Lurker | 50 | 16 | ambush |

### Pandemonium Enemy Pool

| Role | Enemy | HP | Damage | Behavior |
|---|---|---|---|---|
| melee | Chaos Spawn | 30 | 11 | roaming |
| ranged | Void Eye | 15 | 8 | ranged |
| elite | Abyssal Knight | 110 | 22 | patrol |
| swarm | Maddened Rats | 7 | 3 | swarm |
| caster | Reality Bender | 20 | 14 | ranged (spells) |
| ambush | Mimic | 60 | 15 | ambush |

---

## Loot Tables

Each biome has a weighted loot table that determines what items drop.

```go
type LootTable struct {
    BiomeID string
    Entries []LootEntry
}

type LootEntry struct {
    ItemID   string
    Weight   float64 // relative probability
    MinFloor int     // earliest floor this can drop
    Rarity   string  // "common", "uncommon", "rare", "legendary"
}
```

Rarity weights scale with floor depth:

| Rarity | Base Weight | Floor Modifier |
|---|---|---|
| Common | 60% | -3% per floor |
| Uncommon | 30% | +1% per floor |
| Rare | 9% | +1.5% per floor |
| Legendary | 1% | +0.5% per floor |

---

## Floor → Biome Assignment

Biomes are assigned to floors at the start of each run. Assignment depends on run length and progression.

### Assignment Algorithm

```go
func AssignBiomes(totalFloors int, profile *PlayerProfile, unlocked []string) []string {
    // Available biome pool (starts with base 4, grows with progression)
    pool := []string{"crypt", "lair", "catacomb", "moss"}

    // Add unlocked biomes
    for _, b := range unlocked {
        pool = append(pool, b)
    }

    biomes := make([]string, totalFloors)

    // First floor: always an introductory biome (crypt or catacomb)
    biomes[0] = pickFrom([]string{"crypt", "catacomb"})

    // Last floor (boss): determined by ascending NPC's philosophy
    biomes[totalFloors-1] = bossFloorBiome(profile)

    // Middle floors: weighted random from pool, no immediate repeats
    for i := 1; i < totalFloors-1; i++ {
        biomes[i] = pickWeighted(pool, biomes[i-1]) // exclude previous
    }

    return biomes
}
```

### Biome Unlock Conditions

| Biome | Unlock Condition |
|---|---|
| Crypt | Always available |
| Lair | Always available |
| Catacomb | Always available |
| Moss | Always available |
| Gehenna | Complete 3 runs |
| Pandemonium | Complete 5 runs or reach floor 8+ |
| Cocytus | NG+ cycle 2 |
| Lapis | Encounter knowledge-NPC |
| Hive | Complete 8 runs |

### Boss Floor Biome

The boss floor biome is determined by the ascending NPC's philosophy:

| Philosophy | Boss Biome |
|---|---|
| Order | Catacomb (structured, rigid) |
| Chaos | Pandemonium |
| Knowledge | Lapis (if unlocked) or Catacomb |
| Power | Gehenna |
| Redemption | Moss (nature reclaiming) |
| No ascension (fallback) | Random from unlocked pool |

---

## Biome-Specific Features

### Environmental Hazards

| Biome | Hazard | Effect |
|---|---|---|
| Crypt | Bone piles | Crunch sound alerts nearby enemies |
| Lair | Web tiles | Slows player movement by 50% |
| Catacomb | Crumbling floor | Tile collapses after stepping (one-way) |
| Moss | Poison spores | Periodic poison damage in marked areas |
| Gehenna | Lava tiles | Continuous fire damage while standing |
| Pandemonium | Reality tears | Teleports player to random room tile |

### Special Room Types

| Biome | Special Room | Contents |
|---|---|---|
| Crypt | Ossuary | Bonus loot, guarded by elite |
| Lair | Den | Ambush encounter, guaranteed rare drop |
| Catacomb | Burial Chamber | Lore fragment + minor NPC |
| Moss | Spring | Heals player to full HP on first visit |
| Gehenna | Forge | Upgrade one equipped item |
| Pandemonium | Void Pocket | Random effect (buff, debuff, teleport, loot) |

---

## Visual Integration

### Tile Mapping

Each biome specifies its wall and floor sprite flavors, which map to the existing `sprites.SpriteSheet` flavor system:

```go
biomeConfigs := map[string]*BiomeConfig{
    "crypt": {
        WallFlavor:  "crypt",
        FloorFlavor: "crypt",
        AmbientColor: color.RGBA{40, 30, 50, 20},
    },
    "moss": {
        WallFlavor:  "moss",
        FloorFlavor: "moss",
        AmbientColor: color.RGBA{20, 50, 30, 15},
    },
    // ...
}
```

This integrates directly with the existing `GenParams.WallFlavor` and `FloorFlavor` fields in the procedural generator.

---

## Implementation Notes

### Phase 1 (Basic Biomes)
1. `BiomeConfig` struct with wall/floor flavor mapping
2. Biome assigned per floor at run start (simple rotation)
3. Pass biome flavor to `GenParams` during floor generation
4. 4 starter biomes (Crypt, Lair, Catacomb, Moss) using existing tilesets

### Phase 2 (Enemy Pools)
1. Enemy pool definitions per biome
2. Encounter templates reference biome enemy pools
3. Loot tables per biome
4. Difficulty-scaled enemy stats

### Phase 3 (Full Biomes)
1. Environmental hazards per biome
2. Special room types
3. Biome-specific ambient color overlays
4. Biome unlock system tied to meta-progression
5. Additional biomes (Gehenna, Pandemonium, etc.)
6. Audio integration (ambient tracks per biome)
