**Design Document: "Echoes of Self"**
Feature Name: Echoes of Self
Genre: 2D Isometric Dungeon Crawler
Game: Dungeoneer
Priority: High
**Concept Overview**
"Echoes of Self" is a system that records the players actions from previous runs or deaths and spawns ghostly versions of them in future playthroughs. These echoes can:

Replay their movements and actions (like "ghosts" in racing games),

Act as allies, enemies, or neutral NPCs, depending on game logic,

Serve as narrative triggers, unlocking lore, shortcuts, or secrets,

Provide unique meta-progression based on how you interact with them.

This system personalizes the dungeon by reflecting the players past decisions into their current run, enhancing immersion and replayability.

Design Goals:
Personalized Memory System: The dungeon "remembers" your deaths and decisions.

Visual Feedback & Narrative Hooks: Players see how they died, what they missed, or how far they got.

Gameplay Variety: Echoes may help or hinder the player depending on an item held. Echo types are determines by the "Echo Shard" held by the player. The types of shards are "Hero", "Wicked", "Memory"

Player-Driven Progression: Interacting with echoes could unlock new spells, items, or secrets.

Echo Types:
1. Wicked Echo
Same stats as player at death.

Becomes an enemy.

Should be very difficult to defeat. The "Wicked Echo should be like a miniboss that would allow the player to get an advantage in the next run.

Spawns on a random level in the next run if the player is holding the "Wicked Echo Shard"

Drops items and gives buffs when defeated.

2. Hero Echo
Same stats as player at death.

Becomes an ally.

Should be very difficult to defeat. The "Wicked Echo should be like a miniboss that would allow the player to get an advantage in the next run.

Spawns on a random level in the next run if the player is holding the "Hero Echo Shard". 

Once found the "Hero Echo" will follow the player and assist in combat. After defeating enemies the player is fighting it should return to following the player. This behavior should continue until defeated. 

Drops NO items and gives NO buffs when defeated.

3. Memory Fragment
Static ghost NPC that triggers lore/dialogue/memory.

Non-hostile.

Used in puzzles, sidequests, or narrative sequences.

Core Mechanics:

Data to Record:
Each time the player dies, the following is saved:

Player position history: array of (x, y) positions over time

Timestamped actions: attacks, spells, item use

HP and status timeline

Cause of death / final attacker

Optional: equipped items, tile visibility (FOV), level seed

This gets saved as an EchoRecord struct, serialized to disk.

Spawn Rules:
Spawn an Echo in a matching dungeon room or near players old death location.

Allow max N echoes per level, oldest are purged first.

Echoes may spawn after a certain progression threshold (e.g., Floor 2+).

Visual Design:
Ghost-like shader effect (tint blue/purple, partial transparency)

Flickering playback line when retracing path

Optional: a soft ghostly whisper SFX

Lore Integration:
Echoes can whisper secrets: "Theres a hidden door nearby"

"You died here once. Are you ready this time?"

Dialogue fragments from past runs (stored as small templates)

Interaction:
Walk near: Replay path or animation

Attack: May trigger a fight or banish it

Use 'Commune' skill: Talk and gain lore or buffs

Ignore: Disappears over time, or haunts you later

Save/Load Design
go
Copy
Edit
type EchoRecord struct {
	RunID         string
	LevelID       string
	PlayerName    string
	FinalX, FinalY int
	Path          []PlayerStep
	Actions       []RecordedAction
	HPHistory     []int
	CauseOfDeath  string
	Timestamp     time.Time
}
 Prototype Implementation Outline
 1. Data Capture (on player death)
Track players position/actions every n ticks.

Save to a JSON file ./echoes/<RunID>.json using Gos standard lib.

go
Copy
Edit
type PlayerStep struct {
	X, Y     float64
	Tick     int
	LeftFacing bool
}

type RecordedAction struct {
	Action string // "Attack", "Spell", etc.
	Tick   int
}
 2. Echo Spawner
In your level loader:

go
Copy
Edit
func LoadEchoes(levelID string) []*Echo {
	files := getJSONFilesForLevel(levelID)
	for each file:
		parse EchoRecord
		create new Echo entity
}
 3. Echo Entity
go
Copy
Edit
type Echo struct {
	Path      []PlayerStep
	Actions   []RecordedAction
	TickIndex int
	Mode      EchoMode // Ghost, Combat, Memory
}
In Update():

Replay movement via TickIndex++

Optional: trigger Actions like attack animations

Switch to combat mode if interacted with

 4. Drawing
Draw semi-transparent sprite following Path[TickIndex]. Add flicker or fade over time.

Optional Extensions: 
Echo Fusion: Reclaim a ghost to get a perk.


 Codex Prompt
You are working on a 2D isometric dungeon crawler in Go using the Ebiten engine. Implement a system called "Echoes of Self" that records a players position, actions, and cause of death during a run. On future runs, spawn ghost entities that replay the recorded data, either visually or interactively. Echoes should be loaded from disk, follow the players past movement path, and optionally trigger past actions (e.g., attack swings). Design the EchoRecord, replay system, and rendering logic for a basic prototype.

