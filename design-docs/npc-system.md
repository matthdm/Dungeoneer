# NPC System

NPCs are the narrative heart of Dungeoneer. They are the beings trapped in the Dungeon who have found — or are still searching for — purpose. The player's interactions with them drive the story, shape boss encounters, and give meaning to the run loop.

---

## NPC Categories

### Major NPCs

Major NPCs have:
- Multi-phase quest arcs (see [quest-system.md](quest-system.md))
- Branching dialogue trees (see [dialogue-system.md](dialogue-system.md))
- The potential to ascend and become the final boss
- Persistent memory across NG+ cycles
- A distinct philosophy and motivation
- Presence in the hub between runs

**Starting roster: 1 major NPC** (Varn, The Chainkeeper), with the system designed to support 3-5.

### Minor NPCs

Minor NPCs have:
- 2-5 lines of non-branching dialogue
- No quest arcs
- No ascension potential
- Lore and atmosphere delivery
- May sell items, give hints, or provide ambient texture
- Spawned procedurally or placed per biome

Examples:
- A hollow monk who prays endlessly
- A scavenger who trades gossip for scraps
- A soldier who guards an empty doorway out of habit
- A scholar cataloguing stones that keep changing

---

## NPC Entity

NPCs are a new entity type, distinct from `Monster` and `ItemDrop`.

```go
type NPC struct {
    // Identity
    ID          string // unique identifier, e.g., "varn"
    Name        string // display name
    Title       string // e.g., "The Chainkeeper"
    Philosophy  string // e.g., "order", "chaos", "knowledge"

    // Position
    TileX, TileY int
    InterpX, InterpY float64
    LeftFacing   bool

    // Visuals
    Sprite       *ebiten.Image
    PortraitID   string           // for dialogue panel
    BobOffset    float64

    // State
    IsMajor       bool
    Phase         int              // current quest phase (0-3 for major NPCs)
    DialogueID    string           // current dialogue tree ID
    Interactable  bool             // can player talk to them
    InteractRange float64          // tile distance for interaction (default 1.5)

    // Behavior
    Behavior     NPCBehavior      // idle, patrol, follow
    OnInteract   func(player *Player, flags map[string]int)
}

type NPCBehavior interface {
    Update(npc *NPC, level *Level, dt float64)
}
```

### NPC Behaviors

```go
// Stands still, faces player when nearby
type IdleBehavior struct {
    FacePlayerRadius float64
}

// Walks a fixed patrol route
type PatrolBehavior struct {
    Waypoints []PathNode
    Current   int
    Speed     float64
}

// Follows the player at a distance
type FollowBehavior struct {
    Distance float64
    Speed    float64
}
```

---

## Interaction

Player interacts with NPCs by:
1. Walking adjacent to the NPC (within `InteractRange`)
2. Pressing the interact key (default: E)
3. Dialogue panel opens with the appropriate dialogue tree

The interact key is a new `ActionInteract` in the controls system.

### Interaction Indicator

When the player is within range of an interactable NPC, display:
- A small icon above the NPC (speech bubble or "E" key hint)
- Subtle highlight or glow on the NPC sprite

---

## Major NPC: Varn, The Chainkeeper

The first major NPC to implement. Serves as the template for all future major NPCs.

### Character Profile

| Attribute | Value |
|---|---|
| ID | `varn` |
| Name | Varn |
| Title | The Chainkeeper |
| Philosophy | Order — believes structure prevents suffering |
| Motivation | Restore the "chains" that once held the Dungeon's chaos in check |
| Flaw | Cannot accept that order itself becomes oppression |
| Arc | Begins as a reasonable authority figure, ends as a tyrant |

### Phase Progression

| Phase | Floor Range | Summary |
|---|---|---|
| 0 — Introduction | Floor 1-2 | Player meets Varn. He explains his philosophy. Asks for small help (clear a corrupted area). |
| 1 — Task | Floor 3-4 | Varn asks player to recover a relic. Player can comply or refuse. Trust builds or erodes. |
| 2 — Conflict | Floor 5-6 | Varn's methods escalate. He asks player to imprison or eliminate a rival NPC. Moral choice. |
| 3 — Ascension | Floor 7+ | If trust is high enough, Varn has seized control of a section of the Dungeon. He becomes the final boss. |

### Dialogue Trees

- `varn_phase0` — First meeting (see dialogue-system.md for example)
- `varn_phase1` — Task assignment and progress
- `varn_phase2` — Conflict escalation, moral dilemma
- `varn_phase3` — Pre-boss confrontation
- `varn_ng_plus` — Post-defeat reflection (NG+ only)
- `varn_hub` — Hub world idle dialogue (between runs)

### Quest Flags

```
varn_met          = 0/1
varn_trust        = 0-5
varn_phase        = 0-3
varn_task_done    = 0/1
varn_betrayed     = 0/1
varn_rival_killed = 0/1
varn_ascended     = 0/1
```

---

## Minor NPC Examples

### The Hollow Monk

```go
NPC{
    ID: "hollow_monk",
    Name: "Hollow Monk",
    IsMajor: false,
    Behavior: &IdleBehavior{FacePlayerRadius: 3},
    DialogueID: "hollow_monk_ambient",
}
```

Dialogue (simple, non-branching):
1. "I have prayed every day since I arrived. No answer has come."
2. "But I continue. What else would I do?"
3. "Perhaps you will find what I cannot."

### The Scavenger

Dialogue:
1. "Psst. I found something shiny three rooms back. Cost you a key."
2. "No key? Then we have nothing to discuss."
3. *[If player has a key]* "Pleasure doing business."

Can trigger a simple trade action (take key, reveal hidden room location).

### The Forgotten Soldier

Dialogue:
1. "I guard this door. I have always guarded this door."
2. "I don't remember who told me to. But orders are orders."
3. "You may pass. I never said you couldn't."

---

## NPC Spawning

### In Dungeon (During Runs)

Major NPCs spawn based on quest phase and floor number:

```go
type NPCSpawnRule struct {
    NPCID       string
    MinFloor    int
    MaxFloor    int
    Phase       int    // which phase triggers this spawn
    Condition   string // quest flag condition
    Placement   string // "room_center", "room_edge", "corridor", "near_exit"
    Unique      bool   // only one instance per run
}
```

Spawn rules for Varn:
```go
[]NPCSpawnRule{
    {NPCID: "varn", MinFloor: 1, MaxFloor: 2, Phase: 0, Placement: "room_center", Unique: true},
    {NPCID: "varn", MinFloor: 3, MaxFloor: 4, Phase: 1, Placement: "room_center", Unique: true},
    {NPCID: "varn", MinFloor: 5, MaxFloor: 6, Phase: 2, Placement: "room_edge",   Unique: true},
    // Phase 3: Varn doesn't spawn as NPC — he's the boss
}
```

Minor NPCs spawn procedurally:
- 1-2 minor NPCs per floor
- Selected from a pool based on biome
- Placed in rooms (not corridors)
- Never in the same room as a major NPC

### In Hub (Between Runs)

Major NPCs appear in the hub's NPC Quarter after being met:
- Spawn condition: `{npc_id}_met == 1`
- Position: predetermined slot in NPC Quarter
- Dialogue: `{npc_id}_hub` tree (reflective, between-run dialogue)

---

## NPC Persistence

### Within a Run

NPCs that the player has interacted with may reappear on later floors. The `RunState.QuestFlags` map tracks all NPC state for the current run.

### Across Runs (Meta Save)

```go
// Stored in MetaSave
type NPCMetaState struct {
    Met           bool
    DefeatCount   int    // how many times defeated as boss
    HighestPhase  int    // furthest phase reached across all runs
    TotalTrust    int    // cumulative trust across runs (for NG+ dialogue)
}
```

---

## Ascension System

See [quest-system.md](quest-system.md) for the full ascension pipeline.

Summary:
1. Player helps major NPC through phases 0-2
2. If trust threshold met, NPC enters Phase 3 (Ascension)
3. Ascended NPC becomes the final boss on the boss floor
4. If multiple NPCs could ascend, the one with highest trust wins
5. Defeated boss returns to hub with NG+ dialogue

---

## Implementation Notes

### Phase 1 (Minimum Viable NPC)
1. `NPC` entity struct with position, sprite, interaction
2. `IdleBehavior` (stand still, face player)
3. Interact key (E) triggers dialogue panel
4. One hardcoded minor NPC with 3 lines
5. NPC rendering in the isometric pipeline (same depth sorting as monsters)

### Phase 2 (Major NPC Pipeline)
1. Varn fully implemented with Phase 0-1 dialogue trees
2. Quest flag store in RunState
3. NPC spawn rules per floor
4. Hub NPC Quarter with post-run Varn dialogue

### Phase 3 (Full System)
1. Varn Phases 2-3 (conflict + ascension)
2. Boss binding from ascension
3. 2-3 additional minor NPCs with biome-specific pools
4. NPC meta persistence (NG+ memory)
5. Additional major NPCs (future)
