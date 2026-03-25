# Room Tagging System

Room tagging is a post-generation pass that assigns semantic roles to rooms and corridors. Tags tell spawners **what a space is for** — a treasure vault, a guard post, an NPC sanctuary — so that placement decisions feel authored rather than random.

---

## Why Room Tagging

The level generator creates rooms with geometry (position, size) but no meaning. Currently, spawners choose rooms by iterating in generation order or picking the largest. This produces functional but flat results:

- Every room feels the same regardless of contents
- NPCs and enemies compete for the same spaces with no coordination
- There's no concept of "this room is special" — a treasure room looks like every other room
- The enemy placement doc already describes room zones (center, back edge, entry) but nothing assigns rooms to *roles*

Room tags bridge the gap between geometry and narrative.

---

## Tag Vocabulary

Tags are strings attached to rooms during the tagging pass. A room can have multiple tags.

### Primary Role Tags (mutually exclusive)

Each room receives exactly one primary role. The tagger assigns these based on room size, position in the level graph, and floor context.

| Tag | Meaning | Selection Criteria |
|---|---|---|
| `spawn` | Player start room | Contains the spawn point |
| `exit` | Contains the exit portal or stairs | Contains the exit entity |
| `boss_arena` | Boss fight room | Largest room on boss floors (existing logic) |
| `sanctuary` | Safe room for important NPCs | Medium/Large room, far from spawn and exit, no monsters |
| `treasure` | Loot room with decorations | Small/Medium room, off the critical path when possible |
| `guard_post` | Fortified enemy encounter | Medium room near a junction or chokepoint |
| `barracks` | Dense enemy room (swarm encounters) | Large room not used for boss or sanctuary |
| `ambush` | Trap/surprise encounter room | Small room with single entrance |
| `crossroads` | Hub room connecting multiple corridors | Room with 3+ exits |
| `dead_end` | Terminal room with one entrance | Room with exactly 1 exit |
| `common` | No special role — filler room | Default for untagged rooms |

### Modifier Tags (stackable)

Modifiers add properties on top of the primary role.

| Tag | Meaning | Applied When |
|---|---|---|
| `decorated` | Room receives prop sprites (furniture, candles, etc.) | Sanctuary, treasure, boss_arena |
| `loot` | Room has item drops or chests | Treasure, sometimes guard_post (reward for clearing) |
| `cleared` | No enemies spawn here | Sanctuary, spawn, exit |
| `dark` | Reduced FOV / no ambient light | Ambush, dead_end (atmosphere) |
| `locked` | Requires key or condition to enter | Treasure on later floors, optional |
| `optional` | Not on the critical path from spawn to exit | Dead ends, branches |

---

## Tagging Algorithm

The tagging pass runs after level generation but before any entity spawning. It operates on the room list and the level's walkability grid.

```
1. Tag `spawn` and `exit` rooms (rooms containing those entities)
2. Tag `boss_arena` on boss floors (largest room, existing logic)
3. Classify room connectivity:
   - Count exits per room (doorways / corridor connections)
   - Rooms with 1 exit → candidate `dead_end`
   - Rooms with 3+ exits → candidate `crossroads`
4. Build critical path (BFS from spawn to exit)
5. Rooms NOT on critical path are marked `optional`
6. Assign primary roles by priority:
   a. `sanctuary` — 1 per floor, Medium+ room, not on critical path preferred,
      farthest from spawn among candidates. Gets `cleared` + `decorated`.
   b. `treasure` — 1-2 per floor, Small/Medium `dead_end` or `optional` rooms.
      Gets `loot` + `decorated`.
   c. `guard_post` — Medium rooms on critical path near junctions.
   d. `barracks` — Large rooms not already tagged.
   e. `ambush` — Small rooms with 1 exit on critical path.
   f. `crossroads` — assigned from connectivity pass.
   g. `dead_end` — assigned from connectivity pass (if not already treasure/ambush).
   h. Everything else → `common`.
7. Apply modifier tags based on primary role (see table above).
8. On boss floors: `boss_arena` gets `decorated`, nearby room gets `boss_adjacent` modifier.
```

### Floor Scaling

| Floor Position | Sanctuaries | Treasure Rooms | Guard Posts |
|---|---|---|---|
| Floor 1-2 | 1 | 1 | 1 |
| Floor 3-5 | 1 | 1-2 | 2 |
| Floor 6+ | 0-1 | 1-2 | 2-3 |
| Boss floor | 0 | 1 | 1 |

Sanctuaries become rarer on deeper floors — safety is a luxury.

---

## Data Model

```go
// RoomTag represents a semantic role or modifier assigned to a room.
type RoomTag string

const (
    // Primary roles
    TagSpawn     RoomTag = "spawn"
    TagExit      RoomTag = "exit"
    TagBossArena RoomTag = "boss_arena"
    TagSanctuary RoomTag = "sanctuary"
    TagTreasure  RoomTag = "treasure"
    TagGuardPost RoomTag = "guard_post"
    TagBarracks  RoomTag = "barracks"
    TagAmbush    RoomTag = "ambush"
    TagCrossroads RoomTag = "crossroads"
    TagDeadEnd   RoomTag = "dead_end"
    TagCommon    RoomTag = "common"

    // Modifiers
    TagDecorated RoomTag = "decorated"
    TagLoot      RoomTag = "loot"
    TagCleared   RoomTag = "cleared"
    TagDark      RoomTag = "dark"
    TagLocked    RoomTag = "locked"
    TagOptional  RoomTag = "optional"
)

// Room gains a Tags field:
type Room struct {
    X, Y, W, H       int
    CenterX, CenterY int
    Size              RoomSize
    Index             int
    Tags              []RoomTag  // assigned during tagging pass
}

// Helper methods on Room:
func (r *Room) HasTag(tag RoomTag) bool
func (r *Room) AddTag(tag RoomTag)
func (r *Room) PrimaryTag() RoomTag  // returns first non-modifier tag
```

---

## Consumer Integration

### Enemy Spawner

The encounter template system reads room tags to decide what goes where:

| Room Tag | Encounter Selection |
|---|---|
| `guard_post` | Defensive encounters (Firing Line, Ambush Pair) |
| `barracks` | High-density encounters (Swarm Room, Elite + Adds) |
| `ambush` | Ambush encounters only |
| `crossroads` | Patrol encounters, enemies watching multiple exits |
| `cleared` | No enemies — skip room entirely |
| `common` | Standard random encounter selection |

### NPC Spawner

NPCs use tags to find appropriate placement (see [spawn-placement.md](spawn-placement.md)):

| NPC Tier | Preferred Room Tags |
|---|---|
| `quest` | `sanctuary` (exclusive) |
| `ambient` | `common`, `crossroads` |
| `wandering` | corridor tiles (no room tag) |
| `hidden` | `dead_end`, `optional` rooms |
| `entrance` | `spawn` room or adjacent |
| `exit` | `exit` room or adjacent |
| `boss_adjacent` | room nearest to `boss_arena` |

### Decoration Pass

Rooms tagged `decorated` receive visual props during a post-tag decoration pass:

| Room Role | Decoration Set |
|---|---|
| `sanctuary` | Candles, prayer mats, bookshelves, warm lighting |
| `treasure` | Chests, gold piles, weapon racks |
| `boss_arena` | Pillars, braziers, runic circles |
| `guard_post` | Weapon racks, armor stands, barricades |

This pass is future work but the tag system supports it from day one.

---

## Corridor Tagging

Corridors are not rooms but still benefit from lightweight tagging:

| Corridor Property | Tag |
|---|---|
| Long straight corridor (8+ tiles) | `patrol_route` — good for roaming enemies |
| Dead-end offshoot | `alcove` — good for hidden NPCs or ambush enemies |
| Junction (3+ connections) | `junction` — good for crossroads encounters |

Corridor tagging is a stretch goal. The room tagging system works without it.

---

## Implementation Notes

### Phase 3 (Current — NPC Placement)
1. Add `Tags []RoomTag` to Room struct
2. Implement basic tagging pass: spawn, exit, boss_arena, sanctuary, treasure, common
3. NPC spawner reads tags to place NPCs by tier
4. Enemy spawner reads `cleared` tag to skip sanctuary rooms

### Phase 4 (Enemy Encounter Integration)
1. Full tagging algorithm with connectivity analysis
2. Guard post, barracks, ambush, crossroads, dead_end classification
3. Encounter template selection reads room tags
4. Difficulty budget allocated per room based on tag

### Phase 5 (Decoration)
1. Decoration pass reads `decorated` tag
2. Prop placement within rooms based on role
3. Lighting variations based on `dark` tag
