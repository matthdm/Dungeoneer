📄 Design Document: "Living Dungeon AI"
Feature Name: Living Dungeon AI
Genre: 2D Isometric Dungeon Crawler
Game: Dungeoneer
Priority: Medium (after Echoes of Self)
🧠 Concept Overview
The "Living Dungeon AI" turns the dungeon into a reactive, semi-sentient environment that learns from and adapts to the player’s actions across runs. Unlike static procedural generation, this system introduces meta-behavioral adaptation — the dungeon doesn’t just regenerate, it evolves based on how you play.

Each decision the player makes — their combat style, exploration pace, spell usage, preferred enemy types avoided, and even number of times they died — subtly influences how future dungeons are generated, both thematically and tactically.

🧩 Design Goals
Personalized Procedural Generation: Each run becomes tailored to your habits.

Anticipatory Level Design: The dungeon starts “playing against” you.

Replayability: No two players have the same dungeon, even with the same seed.

Meta-Narrative Depth: Dungeon as an entity with memory and personality.

🔮 Key Mechanisms
1. Player Behavior Profiling
Track player behaviors during each run, including:

Behavior Type	Examples
Combat Bias	Melee-heavy vs. Ranged-heavy
Spell Usage	Fire usage, healing spam, etc.
Avoidance Patterns	Always avoids ghosts or spiders
Preferred Paths	Sticks to lit areas, ignores side paths
Risk Tolerance	Opens every chest? Always low HP?

These are saved as a PlayerProfile struct and influence future generation logic.

2. Adaptive Dungeon Director
A lightweight AI director (think Left 4 Dead’s pacing system) that reacts based on PlayerProfile.

It can modify:

Room layouts (e.g., more bottlenecks vs. open arenas)

Enemy compositions (more ranged if player is a ranged abuser)

Trap frequency (more if player plays too fast)

Puzzle frequency (if player ignores combat)

Ambush triggers (if player rushes into rooms blindly)

3. Evolving Ruleset Over Time
As the player completes runs or dies repeatedly, the dungeon subtly mutates:

Traits like “Spiteful,” “Curious,” or “Paranoid” develop.

These change the “tone” of generation (more traps, more stalkers, more illusions).

Think of this like a roguelike mutation tree for the dungeon itself.

4. Lore and Visual Integration
Dungeon walls may show runes referring to prior mistakes.

Shadowy visions echo the dungeon's growing awareness.

Narrator whispers: “You thought that trick would work again?”

📦 Data Architecture
PlayerProfile
go
Copy
Edit
type PlayerProfile struct {
	RunCount            int
	AvgCombatStyle      string // "Melee", "Ranged", "Spellcaster"
	SpellUseCounts      map[string]int
	EnemiesAvoided      map[string]int
	RoomsSkipped        int
	ChestsOpened        int
	AvgHPOnRoomEntry    float64
	AvgTimePerRoom      float64
	LastDeathCause      string
	RecentTraits        []string // ["Reckless", "Cautious"]
}
Saved per run, then summarized into a meta-profile.

DungeonMood / Traits
go
Copy
Edit
type DungeonMood struct {
	Spiteful  bool // More enemies of types that killed player
	Chaotic   bool // Layouts more random, unstable
	Cautious  bool // Slower pacing, more patrols
	Deceptive bool // Fake exits, illusions
}
🛠️ Prototype Implementation Outline
🔹 1. Tracking Player Behavior
During Update() in game.go or player.go, log behavior like:

Time per room

Distance to enemies on engagement

Spell usage per combat

FOV coverage per room

Chest openings

Avoided vs. engaged enemies

Store in a BehaviorTracker struct.

🔹 2. Run Summary on Death/Win
On game over:

go
Copy
Edit
SaveRunProfile(BehaviorTracker, DungeonMood)
UpdateMetaProfile()
Aggregate past n runs to build a persistent PlayerProfile.

🔹 3. Adaptive Generator Hook
When generating a level:

go
Copy
Edit
profile := LoadPlayerProfile()
traits := InferDungeonTraits(profile)

level := GenerateLevelUsingTraits(traits)
This changes the procedural generation parameters.

Examples:

go
Copy
Edit
if traits.Spiteful {
    IncreaseSpawnRateOf(profile.LastDeathCause)
}
if traits.Paranoid {
    AddMoreTrapsAndDeadEnds()
}
🔹 4. Visual Feedback
In Draw() or level init:

Add flickering runes: “Too slow…”

Optional: dungeon introduces corrupted versions of previously-used spells

Entity-based feedback: a mimic chest uses your last known item

✨ Stretch Goals
Dynamic Dialogue / Narration: Dungeon taunts you based on profile.

“Dungeon Alignment” UI: Shows traits like an RPG character sheet.

Player-Driven Redemption: Overcome dungeon’s fear of fire by avoiding it for 3 runs.

Echoes + Dungeon AI Integration: Echoes influence dungeon evolution. (e.g., too many ghost kills = dungeon spawns more of them)

🧠 Codex Prompt
You are developing a dungeon crawler game in Go using the Ebiten engine. Implement a “Living Dungeon AI” system that tracks player behavior across runs — including combat styles, frequently used spells, avoided enemies, and risk patterns. Based on this behavior, modify procedural level generation: change enemy types, room layouts, trap density, and dungeon mood traits. Save behavior summaries in a PlayerProfile, evolve DungeonMood based on past behavior, and hook this into your level generator. Add subtle visual/lore cues that reflect the dungeon's shifting attitude.

