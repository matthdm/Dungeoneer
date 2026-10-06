package combat

// CombatState holds all data the combat engine needs for a single tick.
// Pure data — no Ebiten or game-layer dependencies.
type CombatState struct {
	// Player
	PlayerHP             int
	PlayerMaxHP          int
	PlayerMana           int
	PlayerMaxMana        int
	PlayerDamage         int
	PlayerAttackRange    float64 // tiles
	PlayerAttackInterval float64 // seconds between auto-attacks
	PlayerClass          string  // "knight" | "mage"
	PlayerX, PlayerY    float64 // world position (tile units)

	// Target
	HasTarget        bool
	TargetHP         int
	TargetMaxHP      int
	TargetX, TargetY float64
	TargetLevel      int
	TargetName       string
	TargetInRange    bool // within the player's basic attack reach, with line of sight
	TargetIsDead     bool

	// TargetDist is the distance from player to target in tiles and
	// TargetLOSBlocked is true when a wall stands between them. The game layer
	// fills both every tick; skills with their own CastRange are gated on them.
	// The zero values mean "adjacent and visible", so states that do not model
	// positions (simulations, most tests) are unaffected.
	TargetDist       float64
	TargetLOSBlocked bool

	// Auto-attack
	AutoAttackTimer float64 // countdown to next auto-attack (seconds)
	IsAutoAttacking bool

	// Kill momentum
	KillStreak  int
	StreakTimer float64 // resets if player takes damage before this expires

	// Equipped artifact IDs and their cooldowns (indices 0–6, slot 6 = elite)
	EquippedArtifacts [7]string
	ArtifactCooldowns [7]float64

	// PassiveArtifacts holds artifact IDs whose passive effects apply but which
	// occupy no activation slot — in the game these come from worn equipment
	// (the spell bar only holds actives). Sims may leave this nil and put
	// passives in EquippedArtifacts as before; the engine scans both.
	PassiveArtifacts []string

	DeltaTime float64

	// Raw stat values — used by spell damage scaling (INT) and melee skill scaling (STR).
	// These are populated from BaseStats before combat; they are NOT the same as
	// PlayerDamage (which already bakes in STR for auto-attacks).
	PlayerIntelligence int
	PlayerStrength     int

	// Item-derived stat modifiers (aggregated from equipped items before combat)
	CooldownReductionPct int  // -N% to all active skill cooldowns (e.g. 20 = -20%)
	AttackSpeedPct       int  // +N% attack speed bonus (e.g. 15 = attacks 15% faster)
	SkillDurationPct     int  // +N% to all skill/effect durations
	ManaCostReductionPct int  // -N% to skill mana costs (item passives; 100 = free casting)

	// Enemy capability fields — set when fighting a named EnemyBuild archetype.
	// The Living Dungeon AI populates these before combat begins.
	EnemySilenceRadius    float64 // Silencer: distance within which player skills are blocked
	EnemyDetectionRadius  float64 // Veilbane: ejects player from shadow within this radius
	EnemyInstakillPct     int     // Judge: instakill if player HP% ≤ this (0 = disabled)
	EnemyDamageCapBypass  bool    // Bloodhound: bypasses stone_skin_idol damage cap
	EnemySacrificeLeech   bool    // Crucible: enemy heals on player HP-spend (blood_price)
	EnemyHealReductionPct int     // Nullifier: reduces player healing by N%
	EnemyHPRegenPerSec    int     // Regenerator: enemy heals this many HP per second
	EnemyBlockBlink       bool    // Gravity Warden: prevents IsBlinkStrike skill effects
	EnemyPackBonusPct     int     // Pack Leader: flat damage bonus from pack aura

	// Progression — mirrored from game/progression so the engine stays dependency-free.
	PlayerLevel       int     // starts at 1
	PlayerEXP         int     // EXP accumulated toward the next level
	PlayerEXPToNext   int     // EXP required to reach the next level (cached each level-up)

	// Float accumulator for sub-integer HP drain (blood_vow_amulet drains 5 HP/s;
	// at 60 Hz each tick contributes 0.083 HP — integer truncation would lose it all).
	HPDrainAccum float64

	// Runtime combat state — set by skills, cleared by timers
	InShadow             bool    // true while shroud_cloak shadow form is active
	ShadowTimer          float64 // remaining shadow duration (seconds)
	TargetRooted         bool    // true while target is rooted (cannot attack player)
	RootTimer            float64 // remaining root duration
	DamageReductionPct   int     // current incoming damage reduction % (from taunt)
	TauntTimer           float64 // remaining taunt duration
	BurnActive           bool    // true while burn DoT is ticking on target
	BurnDPS              int     // damage per second of active burn
	BurnTimer            float64 // remaining burn duration
	NextCritGuaranteed   bool    // shroud_cloak blink guarantees next auto-attack is a crit
	ActiveDoTCount       int     // number of distinct DoTs currently active on target (for resonance_crystal); derived from DoTs each tick

	// DoTs is the list of lingering effects on the current target: the burn and
	// any persistent AoE field. It is the source of truth; BurnActive, BurnDPS,
	// BurnTimer and ActiveDoTCount above are mirrors refreshed every tick for
	// the HUD and for callers that predate this list.
	DoTs [MaxDoTs]DoT
}

// MaxDoTs is how many distinct lingering effects one target can carry.
const MaxDoTs = 4

// DoT is one lingering effect on the target. Each source occupies at most one
// entry; re-applying it refreshes the duration instead of adding a stack.
type DoT struct {
	Source    string  // artifact ID that applied it; "" = empty slot
	DPS       int     // damage per second; 0 for effects whose damage was dealt up front
	Remaining float64 // seconds left
	accum     float64 // damage earned but not yet dealt
	tickTimer float64 // time since the last damage tick
}
