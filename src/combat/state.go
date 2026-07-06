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
	TargetInRange    bool
	TargetIsDead     bool

	// Auto-attack
	AutoAttackTimer float64 // countdown to next auto-attack (seconds)
	IsAutoAttacking bool

	// Kill momentum
	KillStreak  int
	StreakTimer float64 // resets if player takes damage before this expires

	// Equipped artifact IDs and their cooldowns (indices 0–6, slot 6 = elite)
	EquippedArtifacts [7]string
	ArtifactCooldowns [7]float64

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
	ActiveDoTCount       int     // number of distinct DoTs currently ticking on target (for resonance_crystal)
}
