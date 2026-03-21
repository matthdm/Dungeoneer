# Boss System

Bosses are the narrative climax of each run. They are not random encounters — they are the NPCs the player helped, ascended to power, and now stand in the player's way. Defeating them is the point. The irony is that the player made them.

---

## Boss Origin

Every boss originates from the Character Ascension System (see [quest-system.md](quest-system.md) and [npc-system.md](npc-system.md)).

```
NPC Phase 0 (Introduction)
  │
NPC Phase 1 (Task)
  │
NPC Phase 2 (Conflict)
  │
NPC Phase 3 (Ascension)
  │
  ▼
BOSS
```

If no NPC reaches ascension (player ignored or refused all NPCs), a **fallback boss** spawns — a generic dungeon entity that represents the Dungeon itself fighting back.

---

## Boss Selection

At the start of a run, the system evaluates which NPCs are eligible for ascension:

```go
func SelectBoss(flags map[string]int, metaFlags map[string]int, npcs []MajorNPCDef) string {
    candidates := []struct {
        ID    string
        Trust int
    }{}

    for _, npc := range npcs {
        phase := flags[npc.ID + "_phase"]
        trust := flags[npc.ID + "_trust"]
        if phase >= 3 && trust >= npc.AscensionThreshold {
            candidates = append(candidates, struct{ID string; Trust int}{npc.ID, trust})
        }
    }

    if len(candidates) == 0 {
        return "fallback_dungeon_guardian"
    }

    // Highest trust wins. Ties broken by encounter order.
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Trust > candidates[j].Trust
    })

    return candidates[0].ID
}
```

The boss is selected when the player reaches the final floor. The boss floor is generated specifically for this boss.

---

## Boss Entity

```go
type Boss struct {
    // Inherits from Monster base
    *Monster

    // Boss-specific
    NPCID         string          // origin NPC ID (e.g., "varn")
    Title         string          // display title (e.g., "Varn, Unchained Tyrant")
    Philosophy    string          // thematic tag
    CurrentPhase  int             // combat phase (not quest phase)
    MaxPhases     int             // total combat phases (usually 2-3)
    PhaseHP       []int           // HP threshold for each phase transition
    Patterns      [][]BossAttack  // attack patterns per phase
    Arena         *ArenaConfig    // boss arena parameters
    PreFightTree  string          // dialogue tree before fight
    PostFightTree string          // dialogue tree after defeat
    Ascended      bool            // true if this boss came from NPC ascension

    // Visual
    BossBar       *BossHealthBar
    PhaseTransAnim string         // animation played on phase change
}
```

---

## Combat Phases

Boss fights have multiple phases. Each phase has:
- A distinct attack pattern set
- A health threshold for transition
- A visual/audio change on transition
- Optional dialogue mid-fight

### Phase Transitions

```go
type BossPhaseConfig struct {
    HPPercent     float64        // transition at this % of max HP (e.g., 0.5 = 50%)
    NewPatterns   []BossAttack   // attack set for this phase
    OnEnter       func(boss *Boss, level *Level) // spawn adds, change arena, etc.
    Dialogue      string         // one-liner displayed on transition ("You think this ends here?")
    SpeedModifier float64        // attack speed multiplier (1.0 = normal, 1.5 = faster)
}
```

Example for Varn (3 phases):

| Phase | HP Range | Behavior |
|---|---|---|
| 1 | 100%-60% | Methodical, predictable swings. Chains sweep in arcs. Tests the player. |
| 2 | 60%-25% | Chains lash unpredictably. Summons bound prisoners as minions. Faster. |
| 3 | 25%-0% | Desperate. Chains shatter, becomes melee brawler. Deals massive damage but takes more. |

---

## Boss Attacks

```go
type BossAttack struct {
    ID          string
    Name        string
    Type        string     // "melee", "ranged", "aoe", "summon", "dash"
    Damage      int
    Range       float64    // in tiles
    Cooldown    float64    // seconds
    WindupTime  float64    // telegraph duration before hit
    AOERadius   float64    // for area attacks
    SpellID     string     // if attack casts a spell (reuses spell system)
    SummonID    string     // if attack summons an enemy
    SummonCount int
}
```

Boss attacks are assembled from:
1. **NPC philosophy** — determines attack themes (Varn uses chain/binding attacks)
2. **Player choices** — spells or abilities the player taught the NPC become boss attacks
3. **Ascension level** — higher trust = more complex patterns

### Attack Philosophy Mapping

| Philosophy | Attack Theme | Example Attacks |
|---|---|---|
| Order (Varn) | Chains, binding, area denial | Chain sweep, cage trap, shackle bolt |
| Chaos | Random, unpredictable, explosive | Wild barrage, chaos rift, spawn swarm |
| Knowledge | Calculated, spell-heavy, debuffs | Mind spike, silence field, mirror image |
| Power | Brute force, high damage, charge | Ground slam, charge rush, seismic wave |
| Redemption | Defensive, healing, reflects damage | Holy barrier, retribution aura, martyr burst |

---

## Boss Arena

Each boss fight takes place in a generated arena on the final floor. The arena is shaped by the boss's philosophy.

```go
type ArenaConfig struct {
    Width, Height  int            // arena room size
    Shape          string         // "rect", "circle", "cross", "ring"
    Pillars        int            // destructible cover pillars
    Hazards        []ArenaHazard  // environmental dangers
    Exits          bool           // sealed during fight
    Biome          string         // visual theme override
}

type ArenaHazard struct {
    Type     string  // "lava", "spikes", "void", "chains"
    Tiles    []Point // affected tiles
    Damage   int     // per-tick damage
    Active   bool    // can toggle during phase transitions
}
```

Arena examples:

| Boss | Arena Shape | Features |
|---|---|---|
| Varn | Rectangle with pillars | Chain hazards along walls, pillars for cover |
| Chaos NPC | Irregular, shifting | Tiles randomly become void, no safe spots |
| Knowledge NPC | Circular with runes | Rune circles that buff/debuff based on position |

---

## Boss Health Bar

```go
type BossHealthBar struct {
    Name         string
    Title        string
    MaxHP        int
    CurrentHP    int
    PhaseMarkers []float64 // HP% markers showing phase transitions
    Visible      bool
    FadeTimer    float64   // fades in when fight starts
}
```

Rendered at the **top center** of the screen:
- Wide horizontal bar (60-70% of screen width)
- Boss name and title above
- Phase transition markers as notches on the bar
- Smooth HP drain animation
- Color shifts per phase (green → yellow → red)

---

## Pre-Fight & Post-Fight Dialogue

### Pre-Fight

When the player enters the boss arena, a brief dialogue exchange plays:

```
Varn: "You helped me build these chains. Now they hold this entire floor together."
Varn: "I can't let you undo what we've created. Surely you understand."

  ► "I understand. But I have to."
  ► "You've become what you feared, Varn."

Varn: "Then let the chains decide."

[Boss fight begins]
```

Pre-fight dialogue is non-branching in terms of outcome — it always leads to the fight — but the player's response is logged for NG+ memory.

### Post-Fight

After the boss is defeated:

```
Varn: "The chains... they're breaking. All of them."
Varn: "I should have known. Order was never mine to impose."
Varn: "I'll be back in that pit soon enough. And I'll remember this."

[Screen fades. Run complete.]
```

Post-fight dialogue feeds into the Lore Library and NG+ meta flags.

---

## Fallback Boss

If no NPC ascended (player ignored all NPCs), a Dungeon Guardian spawns:

```go
Boss{
    NPCID: "dungeon_guardian",
    Name: "The Warden",
    Title: "Hollow Sentinel of the Deep",
    Philosophy: "none",
    MaxPhases: 2,
    // Generic dungeon-themed attacks
    // No pre/post-fight dialogue (it has no philosophy to argue)
    // Represents the Dungeon rejecting the player's passivity
}
```

This creates an incentive to engage with NPCs — the default boss is mechanically functional but narratively hollow.

---

## NG+ Boss Mutations

Bosses defeated in previous runs return with modifications:

| Defeat Count | Mutation |
|---|---|
| 1st defeat | NPC remembers in hub dialogue, boss unchanged |
| 2nd defeat | Boss has +20% HP, new attack in phase 2 |
| 3rd defeat | Boss arena gains new hazards, unique NG+ dialogue |
| 4th+ | Boss philosophy begins to fracture (dialogue becomes unstable, erratic) |

This reinforces the theme: even bosses can't escape the cycle.

---

## Implementation Notes

### Phase 1 (Minimum Viable Boss)
1. `Boss` struct extending `Monster` with `CurrentPhase`, `PhaseHP`, health bar
2. `BossHealthBar` rendered at top of screen
3. Single fallback boss (Dungeon Guardian) with 2 phases and basic melee/AoE
4. Boss arena: rectangular room, sealed exits
5. Defeating boss triggers run completion

### Phase 2 (Ascension-Bound Boss)
1. Varn boss form with 3 phases and chain-themed attacks
2. Boss selection from quest flags
3. Pre-fight and post-fight dialogue
4. Arena customization per boss philosophy

### Phase 3 (Full System)
1. Multiple boss forms for multiple major NPCs
2. Player-choice-influenced attack patterns
3. NG+ mutations
4. Arena hazards and phase-triggered environmental changes
5. Boss music system integration
