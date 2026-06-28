# Lore Ambient Injection — Design Spec

**Date:** 2026-06-24  
**Status:** Approved, pending implementation plan  
**Scope:** Four passive lore delivery systems that inject existing lore into normal gameplay without requiring the player to visit the lore library or reach dialogue trust thresholds.

---

## Motivation

The lore library and NPC dialogue trees require the player to actively seek out lore. Most runs never reach trust thresholds. Slay the Spire 2 demonstrated that ambient lore delivery — items as vessels, biome atmosphere, events with meaning, enemy codex entries — produces a richer sense of world without adding friction. This spec defines four systems that bake the existing `data/lore.json` content (and new writing) into the run loop passively.

---

## System 1: Item Flavor Text

### Data

New file: `data/items_flavor.json`

```json
[
  {
    "id": "item_2_24",
    "text": "Forged in the kiln-rooms beneath the Hollow Citadel, this emblem was carried by arcanists who never returned.",
    "line": "Still warm from fires long dead."
  }
]
```

- `text` — full flavor paragraph, rendered italic at the bottom of the item tooltip
- `line` — one-liner, queued as a toast on item pickup

**Launch coverage:** all 14 ability items + `iron_key` get both `text` and `line`. Common stat items get a `line` only. The code accepts any subset — content can be added incrementally without touching code.

**Voice:** each entry mixes two registers — an in-world factual voice (what the item is, where it's from) and the item's own uncanny voice (what it implies, what it wants). The `line` should land in one breath.

### Code Changes

| File | Change |
|------|--------|
| `items/types.go` | Add `FlavorText string`, `FlavorLine string` to `ItemTemplate` |
| `items/flavor.go` (new) | `LoadItemFlavor()` — unmarshal JSON, walk registry, patch matching entries |
| `items/load.go` | Call `LoadItemFlavor()` at end of `LoadDefaultItems()`, after `registerKeyItems()` |
| Tooltip render | Add italic `FlavorText` section below description if non-empty |
| Pickup handling | Queue `FlavorLine` as a second toast on item pickup if non-empty |

---

## System 2: Biome Entry Lines

### Data

New file: `data/biome_flavor.json`

```json
{
  "crypt": [
    "The dead here are older than the kingdom that buried them.",
    "Something in the stone remembers being alive.",
    "The air smells like sealed rooms and old decisions."
  ],
  "moss_cave": [
    "The green here is wrong. Too deliberate.",
    "Whatever grew this place is still growing.",
    "Light bends strangely where the roots are thickest."
  ]
}
```

3 lines per biome. Keys match the biome name constants already used in `RunState`.

### Trigger Logic

- `RunState` gets a new field: `BiomesIntroduced map[string]bool`
- In `startFloorWithContext()` (`game/hub.go`), after biome is set: if the current biome key is absent from `BiomesIntroduced`, mark it seen and queue a toast
- Line selection: `floorNumber % 3` — deterministic within a run, varies across runs as floor count drifts

### Layering with Mood Whispers

Biome line fires first, mood whisper fires after. Both go through `pendingToasts`. No collision — they stack in sequence. Biome line fires once per biome per run; mood whispers fire every floor.

---

## System 3: Event Rooms

### Data

New file: `data/events.json`

```json
[
  {
    "id": "weeping_statue",
    "title": "The Weeping Statue",
    "trigger": "object",
    "text": "A marble figure weeps black tears into a cracked basin. The water moves against gravity.",
    "choices": [
      {
        "label": "Drink from the basin",
        "preview": "Cold. Wrong. Powerful.",
        "effects": [{"type": "lose_hp", "value": 8}, {"type": "give_item", "id": "item_0_35"}]
      },
      {
        "label": "Shatter the basin",
        "preview": "Some offerings should not be accepted.",
        "effects": [{"type": "set_flag", "key": "shattered_weeping_statue"}]
      },
      {
        "label": "Walk away",
        "preview": "",
        "effects": []
      }
    ]
  }
]
```

### Room Assignment

- Room tagger (`levels/room_tagger.go`) gets a new `event` tag
- During `startFloorWithContext()`: if floor is eligible (not floor 1, not boss floor), one un-seen event is selected from the pool and a suitable room (dead-end or mid-size, not sanctuary) is tagged `event`
- `RunState.EventsSeen []string` prevents the same event appearing twice in a run

### Trigger Types

| Value | Behavior |
|-------|----------|
| `"entry"` | Panel auto-opens when player enters the tagged room |
| `"object"` | An interactable entity is placed at room center; player presses interact to open |

### UI

Reuses the existing dialogue panel. Speaker name is set to the event's `title`. No new panel required.

### New Event Actions

Extend the existing dialogue action dispatch:

| Action type | Effect |
|-------------|--------|
| `lose_hp` | Reduces player HP by `value` (cannot kill) |
| `gain_remnants` | Adds `value` remnants |
| `give_item` | Already exists in dialogue system |
| `curse_item` | Randomly downgrades one equipped item's Quality by one tier |
| `gain_buff` | Applies a named timed status effect from `entities/effects` |
| `set_flag` | Already exists in dialogue system |

### Content Target

8–12 events at launch across the 3 biomes. Each event must have:
- A **risk option** (meaningful cost for meaningful reward)
- A **safe option** (no cost, lesser or no reward)
- A **lore-only option** (no mechanical effect, pure atmosphere)

---

## System 4: Enemy First-Encounter Flavor

### Data

New file: `data/enemy_flavor.json`

```json
{
  "melee_crypt": "Hollow things. They move like they've forgotten what arms are for.",
  "ranged_moss_cave": "It doesn't aim. It anticipates.",
  "elite": "Something was done to this one. It survived.",
  "caster": "The words it speaks are older than language."
}
```

**Lookup order:** `role_biome` → `role` → nothing (silent). Biome-specific lines convey why this creature belongs here. Fallback lines convey the role archetype.

### Tracking

- `RunState` gets `EnemyIntrosSeen map[string]bool`
- Key: `role + "_" + biome` (e.g. `"melee_crypt"`)
- Same enemy type in a new biome triggers again (per-run, biome-aware cadence)

### Trigger Point

In `BasicChaseLogic()` or the aggro transition in `monster.go`: when a monster first transitions from idle to actively chasing the player — if its `role_biome` key is absent from `EnemyIntrosSeen`, queue the flavor line as a toast and mark it seen.

### Display

Delivered via `pendingToasts`. No new surface. The line appears as the monster closes in — timing feels like a field note snapping into place on first contact.

---

## Shared Architecture Notes

All four systems follow the pattern established by `data/lore.json`:
- Narrative content lives in JSON under `data/`
- Loaded at startup, patched into live game state
- No new UI panels (Systems 1–4 all use existing surfaces: tooltip, toast, dialogue panel)
- No new persistence beyond two `RunState` fields and two `ItemTemplate` fields

### New `RunState` fields

```go
BiomesIntroduced map[string]bool
EventsSeen       []string
EnemyIntrosSeen  map[string]bool
```

### New `ItemTemplate` fields

```go
FlavorText string
FlavorLine string
```

---

## Out of Scope

- Lore unlock integration (these lines are ambient, not gated behind `unlock_lore`)
- Cross-run persistence of events seen (events reset each run; `EventsSeen` lives on `RunState`, not `MetaSave`)
- Enemy flavor for echo entities (echoes are player-derived, not dungeon-native)
- New UI panels or HUD surfaces
