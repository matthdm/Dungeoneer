# Hub World

The hub is the player's persistent home between runs. It is a hand-crafted level that grows and changes based on meta-progression. It is the only place in the game where time feels stable.

---

## Philosophy

The hub is a reprieve from the Dungeon — but it is still *inside* the Dungeon. It is not an escape. It is a pocket of false stability. NPCs who return here after being defeated may reflect on this. The hub should feel like a campfire in the dark: warm, but surrounded.

---

## Layout

The hub is a single `Level` (not procedurally generated). It uses the existing tile and sprite systems.

```
                    ┌─────────────┐
                    │ Lore Library│
                    └──────┬──────┘
                           │
         ┌─────────┐   ┌──┴──┐   ┌──────────┐
         │  Shop   │───│ Hub │───│ Echo     │
         │         │   │Center│   │ Shrine   │
         └─────────┘   └──┬──┘   └──────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
        ┌─────┴────┐  ┌───┴───┐  ┌─────┴─────┐
        │ Upgrade  │  │Dungeon│  │ NPC       │
        │ Station  │  │Portal │  │ Quarter   │
        └──────────┘  └───────┘  └───────────┘
```

The hub is spatially small — 20x20 to 30x30 tiles. The player can reach any interaction point within seconds. No enemies spawn here. FOV is fully lit (no fog of war in the hub).

---

## Interaction Points

### 1. Dungeon Portal (Run Entry)

The central feature of the hub. A visible rift or staircase that begins a new run.

**Interaction:**
- Player walks to portal and interacts
- Dialogue prompt: "Enter the Dungeon?" with confirmation
- On confirm: generate Floor 1, begin run
- Portal visual may change based on dungeon mood (Living Dungeon AI)

```go
type DungeonPortal struct {
    X, Y     int
    SpriteID string
    Active   bool // false during cutscenes or events
}
```

### 2. Shop

Sells items for meta currency. Inventory changes based on meta-progression unlocks.

**Interaction:**
- Opens a shop UI (grid of purchasable items with prices)
- Player spends meta currency (Remnants — see meta-progression.md)
- Purchased items go into a "run loadout" or are consumed immediately
- Stock refreshes after each run

**Stock Categories:**
- Consumables (health potions, mana potions, antidotes)
- Starter weapons (basic variants for different builds)
- Utility items (keys, torches, recall stones)
- Unlockable items (appear after meta-progression milestones)

```go
type ShopItem struct {
    Template  *ItemTemplate
    Cost      int    // meta currency price
    Stock     int    // -1 for unlimited
    Unlocked  bool   // requires meta-progression
    UnlockReq string // e.g., "defeat_3_bosses"
}
```

### 3. Upgrade Station

Permanent upgrades purchased with meta currency. These persist across all future runs.

**Upgrade Categories:**

| Category | Examples |
|---|---|
| Starting stats | +5 base HP, +1 base STR |
| Inventory | +1 inventory row, larger starting pouch |
| Abilities | Start with 1 dash charge pre-filled, grapple range +2 |
| Dungeon | Floor exits visible on minimap, extra loot room chance |
| Meta | +10% meta currency gain, echo spawn chance increase |

**Interaction:**
- Opens upgrade tree UI
- Each upgrade has a cost and optional prerequisite
- Purchased upgrades are stored in meta save

```go
type Upgrade struct {
    ID          string
    Name        string
    Description string
    Cost        int
    MaxLevel    int
    CurrentLevel int
    Prereqs     []string // other upgrade IDs
    Effect      func(player *Player) // applied at run start
}
```

### 4. Echo Shrine

Where the player interacts with Echoes of Self from past runs.

**Interaction:**
- Displays a list of recorded echoes (past deaths)
- Player can preview echo path (ghost replay on a minimap)
- Player can choose which echo type to spawn in next run:
  - **Wicked Echo** — spawns as a miniboss, drops rewards on defeat
  - **Hero Echo** — spawns as an ally, assists in combat
  - **Memory Fragment** — spawns as a lore ghost, provides hints
- Can also banish echoes (delete recording)

```go
type EchoShrine struct {
    X, Y     int
    SpriteID string
    Echoes   []EchoRecord // loaded from meta save
}
```

### 5. Lore Library

A collection of discovered lore fragments, NPC histories, and philosophical texts.

**Interaction:**
- Opens a scrollable text UI
- Entries unlock as the player encounters NPCs, defeats bosses, and discovers secrets
- Categories:
  - **Characters** — NPC bios that update as you learn more about them
  - **Cosmology** — Entries about Abaddon, Azazel, and the nature of the Dungeon
  - **History** — Records of past runs (who ascended, who fell)
  - **Fragments** — Cryptic text pieces found in the dungeon

```go
type LoreEntry struct {
    ID        string
    Title     string
    Category  string // "Characters", "Cosmology", "History", "Fragments"
    Text      string
    Unlocked  bool
    UnlockReq string // e.g., "met_varn_phase_2"
}
```

### 6. NPC Quarter

A section of the hub where major NPCs appear between runs. NPCs only appear here after the player has encountered them in the dungeon.

**Behavior:**
- NPCs stand at designated positions
- Player can talk to them (triggers dialogue system)
- Dialogue reflects current relationship state and quest phase
- NPCs may offer hints about the next run or comment on past events
- Defeated bosses may reappear here with altered dialogue

**NPC Slots:**
- 3-5 designated positions in the NPC quarter
- Each slot can hold one major NPC
- NPCs claim slots based on encounter order
- If an NPC ascends to boss and is defeated, they return here with new dialogue

---

## Hub State

The hub has persistent state stored in the meta save:

```go
type HubState struct {
    ShopStock     []ShopItem
    Upgrades      map[string]int    // upgrade_id → level
    LoreUnlocked  []string          // lore_entry IDs
    NPCsPresent   []string          // NPC IDs currently in hub
    PortalActive  bool
    EchoesStored  []EchoRecord
    MetaCurrency  int               // player's current balance
    RunCount      int
    TotalDeaths   int
}
```

---

## Hub Evolution

The hub changes visually and functionally based on meta-progression:

| Milestone | Hub Change |
|---|---|
| First run completed | Shop opens |
| 3 runs completed | Upgrade station activates |
| First echo recorded | Echo shrine appears |
| First NPC encountered | NPC quarter populated |
| First lore fragment found | Lore library opens |
| 10+ runs | Hub visuals darken/grow (dungeon influence) |
| NG+ cycles | Abaddon's presence felt (ambient effects, whispers) |

---

## Visual Design

- **Lighting:** Warm torchlight, no fog of war. Contrast with the dungeon's darkness.
- **Tiles:** Unique hub tileset — stone with warmth, maybe wooden beams. Not dungeon walls.
- **Ambiance:** Quiet. Maybe a faint hum or heartbeat. No combat sounds.
- **NPC placement:** NPCs face the center of the hub. They feel like refugees, not merchants.

---

## Implementation Notes

### Phase 1 (Minimum Viable Hub)
1. Hand-craft a 20x20 level with the level editor
2. Place dungeon portal entity at center
3. Portal interaction triggers run start
4. Player spawns here on game start and after death/victory
5. No shop, upgrades, or NPCs yet — just the portal

### Phase 2 (Functional Hub)
1. Add shop NPC with basic stock
2. Add upgrade station with 3-5 starter upgrades
3. Meta currency display on HUD
4. Death/victory returns player here with rewards

### Phase 3 (Living Hub)
1. NPC quarter with dynamic population
2. Echo shrine with echo management
3. Lore library with unlockable entries
4. Hub visual evolution based on progression
