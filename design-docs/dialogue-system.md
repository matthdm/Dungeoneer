# Dialogue System

The dialogue system is the primary vehicle for narrative delivery. It handles NPC conversations, quest progression, lore exposition, and player choice.

---

## Visual Style

Inspired by **Baldur's Gate 1** and **RuneScape** dialogue systems:

- Dark bordered panel at the bottom or center of the screen
- NPC portrait on the left (sprite-based, not high-res art)
- NPC name displayed above text
- Dialogue text appears with a typewriter effect (skippable)
- Player response options listed below as clickable text lines
- Selecting a response advances the conversation tree

```
┌──────────────────────────────────────────────────────┐
│ ┌────────┐                                           │
│ │        │  Varn, The Chainkeeper                     │
│ │ [sprite]│                                           │
│ │        │  "Order is the only thing that separates   │
│ └────────┘   us from the void. Help me restore the    │
│              chains, and I will show you what          │
│              stability means in this place."           │
│                                                       │
│  ► I'll help you. What do you need?                   │
│  ► What chains are you talking about?                 │
│  ► I'm not interested. [Leave]                        │
│                                                       │
└──────────────────────────────────────────────────────┘
```

---

## Data Model

### Dialogue Tree

A dialogue tree is a directed graph of nodes. Each node contains text and optional responses that link to other nodes.

```go
type DialogueTree struct {
    ID    string              // unique tree identifier, e.g., "varn_phase1_intro"
    Nodes map[string]*DialogueNode
    Root  string              // starting node ID
}

type DialogueNode struct {
    ID        string
    Speaker   string          // NPC name (or "" for narration)
    Text      string          // the dialogue line
    Portrait  string          // sprite ID for the speaker portrait
    Responses []DialogueResponse
    OnEnter   []DialogueAction // actions triggered when this node is reached
}

type DialogueResponse struct {
    Text      string          // what the player sees as a clickable option
    NextNode  string          // node ID to jump to ("" = end conversation)
    Condition *DialogueCondition // nil = always shown
    OnSelect  []DialogueAction   // actions triggered when selected
}
```

### Conditions

Conditions control which responses are visible to the player. Hidden responses keep the UI clean and prevent impossible choices.

```go
type DialogueCondition struct {
    Type     string // "flag_equals", "flag_gte", "flag_lte", "has_item", "not_flag"
    Flag     string // quest flag key, e.g., "varn_trust"
    Value    int    // comparison value
    ItemID   string // for "has_item" checks
}
```

Examples:
- `{Type: "flag_gte", Flag: "varn_trust", Value: 3}` — only show if trust is 3+
- `{Type: "has_item", ItemID: "broken_chain"}` — only show if player carries item
- `{Type: "not_flag", Flag: "betrayed_varn"}` — hide if player betrayed Varn

### Actions

Actions are side effects triggered by entering a node or selecting a response. They modify game state.

```go
type DialogueAction struct {
    Type    string // "set_flag", "add_flag", "give_item", "take_item",
                   // "give_exp", "give_meta_currency", "unlock_lore",
                   // "spawn_enemy", "heal_player", "start_quest"
    Flag    string
    Value   int
    ItemID  string
    LoreID  string
    Amount  int
}
```

Examples:
- `{Type: "set_flag", Flag: "varn_met", Value: 1}` — mark NPC as met
- `{Type: "add_flag", Flag: "varn_trust", Value: 1}` — increment trust
- `{Type: "give_item", ItemID: "rusty_key", Amount: 1}` — give player an item
- `{Type: "unlock_lore", LoreID: "varn_origin"}` — unlock lore entry

---

## Dialogue Flow

```
Player interacts with NPC (walks adjacent + presses interact key)
  │
  ▼
Load DialogueTree based on NPC ID + current quest phase
  │
  ▼
Display root node
  │
  ▼
Player reads text → clicks response
  │
  ├─ Execute OnSelect actions (set flags, give items, etc.)
  │
  ▼
Navigate to NextNode
  │
  ├─ If NextNode == "" → End conversation, close dialogue UI
  ├─ If NextNode has responses → Display next node
  └─ If NextNode has no responses → Display text, auto-close after click
```

---

## Quest Flag Store

All dialogue state is stored in the run's quest flag map:

```go
// Part of RunState
QuestFlags map[string]int
```

Flag naming convention:
```
{npc_id}_{flag_name}

Examples:
  varn_met        = 1 (has met Varn)
  varn_trust      = 3 (trust level 0-5)
  varn_phase      = 2 (current quest phase)
  varn_betrayed   = 0 (boolean as int)
  helped_prisoner = 1
  found_key_crypt = 1
```

Flags reset each run. Meta flags (for NG+ memory) are stored separately in the meta save.

---

## Dialogue Tree Selection

When the player talks to an NPC, the system selects the correct dialogue tree:

```go
func SelectDialogueTree(npcID string, flags map[string]int, metaFlags map[string]int) string {
    phase := flags[npcID + "_phase"]
    defeated := metaFlags[npcID + "_defeat_count"]

    // NG+ override: defeated NPCs have special dialogue
    if defeated > 0 {
        return npcID + "_ng_plus"
    }

    // Phase-specific dialogue
    return fmt.Sprintf("%s_phase%d", npcID, phase)
}
```

This allows each NPC to have multiple dialogue trees:
- `varn_phase0` — first encounter
- `varn_phase1` — after helping with first task
- `varn_phase2` — conflict phase
- `varn_phase3` — ascension phase
- `varn_ng_plus` — post-defeat (NG+ cycles)

---

## Dialogue File Format

Dialogue trees are defined in JSON for easy editing:

```json
{
  "id": "varn_phase0",
  "root": "greeting",
  "nodes": {
    "greeting": {
      "speaker": "Varn",
      "portrait": "varn_neutral",
      "text": "Another soul cast into this pit. Tell me — do you believe in order?",
      "responses": [
        {
          "text": "Order? In a place like this?",
          "next_node": "explain_order"
        },
        {
          "text": "Who are you?",
          "next_node": "introduce"
        },
        {
          "text": "I don't have time for philosophy. [Leave]",
          "next_node": "",
          "on_select": [
            {"type": "set_flag", "flag": "varn_met", "value": 1}
          ]
        }
      ],
      "on_enter": [
        {"type": "set_flag", "flag": "varn_met", "value": 1}
      ]
    },
    "introduce": {
      "speaker": "Varn",
      "portrait": "varn_neutral",
      "text": "I am Varn. Some call me the Chainkeeper. I held the old laws together — before this place swallowed them.",
      "responses": [
        {
          "text": "What laws?",
          "next_node": "explain_order"
        },
        {
          "text": "Sounds like a losing battle.",
          "next_node": "losing_battle"
        }
      ]
    },
    "explain_order": {
      "speaker": "Varn",
      "portrait": "varn_intense",
      "text": "Without order, every creature in this Dungeon devours the next. I've seen it. I intend to stop it.",
      "responses": [
        {
          "text": "How can I help?",
          "next_node": "accept_task",
          "on_select": [
            {"type": "add_flag", "flag": "varn_trust", "value": 1},
            {"type": "set_flag", "flag": "varn_phase", "value": 1}
          ]
        },
        {
          "text": "That's not my problem.",
          "next_node": ""
        }
      ]
    }
  }
}
```

---

## Minor NPC Dialogue

Minor NPCs (ambient characters with no quest lines) use simplified dialogue:

```go
type SimpleDialogue struct {
    Speaker  string
    Portrait string
    Lines    []string // sequential lines, click to advance
}
```

These NPCs have no branching, no conditions, and no actions. They deliver 2-5 lines of lore or atmosphere and close. Example:

> **Hollow Monk:** "I have prayed every day since I arrived. No answer has come. But I continue. What else would I do?"

---

## UI Implementation

### DialoguePanel

```go
type DialoguePanel struct {
    Active       bool
    Tree         *DialogueTree
    CurrentNode  *DialogueNode
    QuestFlags   map[string]int
    MetaFlags    map[string]int

    // Rendering state
    TextProgress int     // characters revealed (typewriter)
    TextSpeed    float64 // characters per second
    HoverIndex   int     // which response is hovered
    Responses    []DialogueResponse // filtered by conditions
}
```

### Input Handling

- **Click on response** → select it, fire actions, advance to next node
- **Click anywhere during typewriter** → reveal full text immediately
- **Escape** → close dialogue (only if no responses — can't skip choices)
- **Mouse hover** → highlight response option

### Rendering

1. Draw semi-transparent dark background panel
2. Draw NPC portrait (64x64 or 128x128 sprite)
3. Draw speaker name in accent color
4. Draw dialogue text with typewriter reveal
5. Draw response options below text (highlighted on hover)
6. Response text uses a distinct color (gold/amber) to distinguish from NPC text

---

## Integration Points

| System | Connection |
|---|---|
| NPC System | NPCs reference dialogue tree IDs; phase determines which tree loads |
| Quest System | Actions modify quest flags; conditions read quest flags |
| Item System | Actions can give/take items |
| Lore Library | Actions unlock lore entries |
| Meta Progression | Meta flags enable NG+ dialogue branches |
| Boss System | Ascension dialogue plays before boss fight |

---

## Implementation Notes

### Phase 1 (Minimum Viable Dialogue)
1. `DialoguePanel` struct with text rendering and response selection
2. `DialogueTree` loaded from hardcoded Go structs (not JSON yet)
3. Single NPC with 3-4 node conversation
4. Basic typewriter effect
5. Click-to-select responses

### Phase 2 (Full System)
1. JSON-based dialogue tree loading
2. Condition evaluation engine
3. Action execution engine
4. Quest flag integration
5. Portrait rendering
6. Minor NPC simple dialogue support

### Phase 3 (Polish)
1. Typewriter sound effect per character
2. Response hover animations
3. Dialogue history (scroll back)
4. NPC mood/expression changes mid-conversation (portrait swaps)
