# Spawn Placement System

Spawn placement governs **where** NPCs and enemies are positioned on a floor. It replaces the current approach of iterating rooms in generation order with a tier-based system that uses room tags (see [room-tagging.md](room-tagging.md)) to make placement feel intentional.

---

## Design Goals

1. **Important NPCs get important rooms.** A quest-giver with branching dialogue should stand in a decorated sanctuary, not a random corridor.
2. **Ambient NPCs blend into the world.** A scavenger muttering in a hallway feels more natural than one standing in a room center.
3. **Hidden content rewards exploration.** A jester tucked in a dead-end gives players a reason to check every corner.
4. **Enemies and NPCs don't fight for space.** The systems coordinate through room tags — a `sanctuary` is cleared, a `guard_post` is hostile.
5. **The same vocabulary works for both systems.** `SpawnStrategy` is shared; the spawners interpret it differently.

---

## Spawn Strategy

A shared type used by both NPC templates and (eventually) encounter templates.

```go
type SpawnStrategy string

const (
    SpawnQuest        SpawnStrategy = "quest"
    SpawnAmbient      SpawnStrategy = "ambient"
    SpawnWandering    SpawnStrategy = "wandering"
    SpawnHidden       SpawnStrategy = "hidden"
    SpawnEntrance     SpawnStrategy = "entrance"
    SpawnExit         SpawnStrategy = "exit"
    SpawnBossAdjacent SpawnStrategy = "boss_adjacent"
)
```

This type lives in a shared package (e.g., `constants` or `levels`) so both the NPC and enemy spawners can reference it without import cycles.

---

## NPC Placement Tiers

### `quest` — Important NPCs with storylines

**Who:** Major NPCs (Varn), merchants, healers, NPCs with branching dialogue trees.

**Where:** Rooms tagged `sanctuary`. These rooms are monster-free (`cleared`), decorated, and positioned away from the critical path when possible. The NPC stands at the room center.

**Rules:**
- Maximum 1 quest NPC per floor
- Gets the best available sanctuary room (largest, farthest from spawn)
- If no sanctuary room exists (small floor), falls back to the largest `common` room and marks it `cleared`
- Room receives decoration props appropriate to the NPC (candles for a monk, bookshelves for a scholar)

**Feel:** Walking into this room should feel like finding an oasis. The player knows they're safe and something meaningful is here.

### `ambient` — NPCs with minor dialogue

**Who:** Hollow Monk, Forgotten Soldier, Mad Scholar — NPCs with 2-5 lines of lore or flavor text. No quests, no branching.

**Where:** Room centers in `common` or `crossroads` rooms. These NPCs are part of the room, not the reason for it.

**Rules:**
- 0-2 per floor
- Placed in room centers via existing `findWalkableNear` logic
- Skip `sanctuary`, `boss_arena`, `spawn`, and `exit` rooms
- Skip rooms that already contain a quest NPC
- Biome-filtered from the minor NPC pool

**Feel:** A natural encounter. "Oh, someone's in this room." The player can talk or walk past.

### `wandering` — Corridor dwellers

**Who:** Scavenger, lost souls, messengers, vagrants — NPCs that wouldn't logically be standing in a room center.

**Where:** Walkable tiles in corridors (tiles NOT inside any room's bounding box). Biased toward corridor junctions where the player is likely to pass through.

**Rules:**
- 0-1 per floor
- Find walkable tiles not contained by any room
- Prefer tiles with 3+ walkable neighbors (junctions, wide corridors) over narrow passages
- Minimum distance of 5 tiles from player spawn
- Must not block narrow (1-tile-wide) corridors — check that adjacent walkable tiles exist on both sides

**Feel:** Stumbling across someone in a hallway. Unexpected but not hidden. "What are you doing out here?"

### `hidden` — Secret and rare NPCs

**Who:** Jester, secret quest NPCs, rare reward encounters, Easter eggs.

**Where:** Dead-end rooms, small optional rooms off the critical path, or dead-end corridor branches. The player has to go out of their way to find them.

**Rules:**
- 0-1 per floor, low spawn chance (30-40% per eligible floor)
- Prefer rooms tagged `dead_end` + `optional`
- If no dead-end room: find dead-end corridor tiles (walkable tiles with exactly 1 walkable cardinal neighbor)
- Farthest from the critical path preferred
- Can have a special `SpawnChance` field on the template (e.g., jester = 15% per floor)

**Feel:** Genuine discovery. "I almost didn't check down here." Rewards thoroughness and curiosity.

### `entrance` — Near the player spawn

**Who:** Recurring guide NPCs, tutorial hints, a coward who won't go deeper, a merchant selling supplies.

**Where:** The `spawn` room itself (if large enough) or the first room adjacent to spawn on the critical path.

**Rules:**
- Maximum 1 per floor
- Placed 2-3 tiles from the player spawn point, not on the spawn tile itself
- Guaranteed visible when the floor loads (within starting FOV)

**Feel:** "Ah, you again." A reliable presence. The player learns to expect them.

### `exit` — Near the floor exit

**Who:** Gatekeepers, last-chance merchants, warning voices, NPCs who foreshadow the next floor.

**Where:** The `exit` room or the room immediately before it on the critical path.

**Rules:**
- Maximum 1 per floor
- Placed near the exit entity but not on it
- On boss floors, placed outside the `boss_arena` (see `boss_adjacent`)

**Feel:** A checkpoint moment. Rest, buy potions, hear a warning, then descend.

### `boss_adjacent` — Foreshadowing the boss

**Who:** Lore NPCs, shades of previous challengers, a spirit who reveals the boss's weakness.

**Where:** The room nearest to the `boss_arena` that isn't the arena itself.

**Rules:**
- Only spawns on boss floors
- 1 per boss floor
- Room is tagged `cleared` (no enemies between the player and the foreshadowing)
- NPC dialogue references the specific boss

**Feel:** "Something terrible is through that door. Here's what I know about it."

---

## NPC Template Changes

```go
type NPCTemplate struct {
    ID            string
    Name          string
    Title         string
    SpriteID      string
    PortraitID    string
    IsMajor       bool
    DialogueID    string
    Biomes        []Biome
    Placement     SpawnStrategy  // NEW — which tier this NPC uses
    SpawnChance   float64        // NEW — 0.0-1.0, probability per eligible floor (0 = always)
    SpawnMinFloor int            // NEW — earliest floor this NPC can appear
    SpawnMaxFloor int            // NEW — latest floor (0 = no limit)
}
```

Updated pool example:

```go
var npcPool = []NPCTemplate{
    {
        ID: "hollow_monk", Name: "Hollow Monk",
        SpriteID: "Sorcerer", PortraitID: "Sorcerer",
        DialogueID: "hollow_monk",
        Biomes: []Biome{BiomeCrypt, BiomeCatacomb},
        Placement: SpawnAmbient,
    },
    {
        ID: "scavenger", Name: "Scavenger",
        SpriteID: "Caveman", PortraitID: "Caveman",
        DialogueID: "scavenger",
        Biomes: nil, // all biomes
        Placement: SpawnWandering,
    },
    {
        ID: "mad_jester", Name: "The Cackling Jester",
        SpriteID: "Jester", PortraitID: "Jester",
        DialogueID: "mad_jester",
        Biomes: nil,
        Placement: SpawnHidden,
        SpawnChance: 0.15, // 15% chance per floor
    },
    {
        ID: "varn", Name: "Varn", Title: "The Chainkeeper",
        SpriteID: "Chainkeeper", PortraitID: "Chainkeeper",
        IsMajor: true,
        Biomes: nil,
        Placement: SpawnQuest,
        SpawnMinFloor: 1, SpawnMaxFloor: 6,
    },
}
```

---

## Spawner Algorithm

The NPC spawner runs after room tagging and before enemy spawning (so enemies can respect `cleared` rooms).

```
1. Collect eligible NPCs for this floor:
   - Filter by biome
   - Filter by floor range (SpawnMinFloor / SpawnMaxFloor)
   - Roll SpawnChance for probabilistic NPCs
   - Check quest flags for major NPCs (correct phase, not already met this run, etc.)

2. Sort by placement priority:
   quest > boss_adjacent > entrance > exit > ambient > wandering > hidden

3. For each eligible NPC, find placement:
   a. quest       → room with tag `sanctuary`, fallback to largest `common`
   b. boss_adjacent → room nearest `boss_arena` (boss floors only)
   c. entrance    → `spawn` room or first adjacent room
   d. exit        → `exit` room or preceding room
   e. ambient     → random `common` / `crossroads` room center
   f. wandering   → random corridor tile (not in any room)
   g. hidden      → `dead_end` + `optional` room, or dead-end corridor tile

4. For each placement, find a specific tile:
   - Use existing findWalkableNear() for room-based placements
   - Use new findCorridorTile() for wandering placements
   - Use new findDeadEndTile() for hidden placements

5. Validate:
   - Not on player spawn tile
   - Not on exit tile
   - Not on another NPC's tile
   - Minimum 3-tile separation between NPCs

6. Mark rooms with placed quest/entrance/exit NPCs as `cleared`
   (prevents enemy spawner from adding monsters to those rooms)
```

---

## Enemy Spawner Integration

The enemy encounter system benefits from the same strategy vocabulary but interprets it differently:

| Strategy | Enemy Interpretation |
|---|---|
| `quest` | Elite encounter guarding a valuable room (mini-boss + adds) |
| `ambient` | Standard room encounter (scaled to floor difficulty) |
| `wandering` | Corridor patrol (1-2 roaming enemies) |
| `hidden` | Ambush encounter in a dead-end (mimic, lurker) |
| `entrance` | Light encounter near spawn (tutorial-weight) |
| `exit` | Gatekeeper encounter near exit (slightly above average) |
| `boss_adjacent` | Pre-boss gauntlet (hard encounter before the arena) |

This mapping is documented here for future reference. The enemy spawner changes are **not** part of the current implementation — they happen when the encounter template system is upgraded to read room tags.

### Coordination Rules

- Rooms tagged `cleared` receive zero enemies, period
- Rooms tagged `sanctuary` are always `cleared`
- The enemy budget subtracts `cleared` rooms from available space
- If a room has both an NPC and enemies (e.g., `ambient` NPC in a `common` room), enemies spawn but maintain 3-tile distance from the NPC
- `guard_post` rooms near a `sanctuary` can receive *stronger* encounters — the guard post protects the sanctuary

---

## Helper Functions

New utility functions needed for the placement system:

```go
// findCorridorTile returns a walkable tile not inside any room.
// Prefers tiles with 3+ walkable neighbors (junctions).
func findCorridorTile(lvl *Level, rooms []Room, avoid []image.Point) (int, int)

// findDeadEndTile returns a walkable tile with exactly 1 walkable
// cardinal neighbor, or a tile in the smallest dead_end room.
func findDeadEndTile(lvl *Level, rooms []Room, avoid []image.Point) (int, int)

// roomsByTag returns all rooms with the given tag.
func roomsByTag(rooms []Room, tag RoomTag) []*Room

// nearestRoomTo returns the room closest to the target room by center distance.
func nearestRoomTo(rooms []Room, target *Room, exclude []*Room) *Room
```

---

## Implementation Plan

### Phase 3.5 (NPC Placement — current)
1. Add `Tags` field to Room struct
2. Basic tagging pass: `spawn`, `exit`, `boss_arena`, `sanctuary`, `common`
3. Add `Placement` field to NPCTemplate
4. Rewrite `spawnFloorNPCs` to use tag-based placement
5. Implement `findCorridorTile` and `findDeadEndTile`
6. Update existing NPC pool with placement tiers

### Phase 4 (Enemy Integration)
1. Full room tagging (guard_post, barracks, ambush, crossroads, dead_end)
2. Encounter templates read room tags for selection
3. `cleared` tag respected by enemy spawner
4. Guard post encounters scale with nearby sanctuary value

### Future
1. Decoration pass reads tags for prop placement
2. Locked rooms with key items
3. Living Dungeon AI modifies room tags based on player behavior
