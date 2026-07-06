package combat

// ArtifactEffect describes the engine-side behaviour of one artifact.
// Visual and audio reactions are handled by the game layer via events.
type ArtifactEffect struct {
	Cooldown            float64 // seconds; 0 = passive (no active use)
	Domain              string
	DamageMultiplier    float64
	AOERadius           float64
	IsBlinkStrike       bool
	IsTaunt             bool
	IsRoot              bool
	IsExecute           bool
	IsPassive           bool
	ExecuteThresholdPct int
	DurationSec         float64
	BurnDPS             int
	BurnDurationSec     float64
	DamageReductionPct  int

	// New fields for meta-build artifacts
	DamageCapPct         int     // stone_skin_idol: incoming damage capped at N% of max HP
	ShroudCooldownReset  float64 // shadows_return: reduce Shroud Cloak CD by N seconds on hit during shadow
	SkillDurationPct     int     // void_mirror_pendant / voidweave_wraps: +N% to all skill durations
	CooldownReductionPct int     // arcane_tempo_ring: -N% to all active cooldowns
	HPCostPct            int     // blood_price: costs N% of current HP on activation
	DamageFlat           int     // blood_price: flat void damage dealt
	HealOnKillPct        int     // soul_harvest: restore N% max HP on each kill
	HealOnKillFlat       int     // lifedrinker_robe: restore N flat HP on each kill
	SurgeDmgPerCooldown  int     // arcane_surge: damage per artifact currently on cooldown

	// Stat bonuses applied from ALL equipped items (passive or active with a Cooldown).
	// These are always-on bonuses regardless of whether the item has an active skill.
	AttackSpeedPct int // quicksilver_talisman: +N% attack speed (reduces attack interval)
	MaxHPBonus     int // flat +N max HP (stone_skin_idol, wardens_medallion, lifedrinker_robe, etc.)
	MaxHPMod       int // flat modifier to max HP, can be negative (marrow_ring: −15)
	StrBonus       int // iron_will_band: +N Strength (increases auto-attack damage)

	// Offensive passive bonuses
	DamageBonusPct     int // marrow_ring: flat +N% to all damage always
	DamageBonusLowHPPct int // blood_vow_amulet: +N% damage when player HP < 50%
	HPDrainPerSec      int // blood_vow_amulet: drain N HP per second (passive)
	DoTDmgBonusPct     int // resonance_crystal: +N% spell damage per active DoT stack

	// hollow_sigil routing
	VoidCostsHP bool // hollow_sigil: void-domain skill costs HP instead of mana

	// Spell-type fields — registered game spells (fireball, lightning, etc.).
	// Visuals stay in spells.game.go; the engine owns damage calculation.
	IsProjectile      bool    // single-target projectile (fireball, arcane_bolt)
	IsChain           bool    // chains to multiple targets (lightning)
	IsAOEField        bool    // persistent AoE field (lightning_storm, fractal_canopy)
	ChainCount        int     // targets hit by IsChain spells
	SpellDamageBase   int     // base spell damage before stat scaling
	SpellDamagePerINT float64 // damage per point of Intelligence
	SpellDamagePerSTR float64 // damage per point of Strength (slash_combo)
}

// ArtifactEffects is the engine's registry of known artifact IDs → effects.
// The game layer populates this via RegisterArtifact during startup so that
// artifacts defined in JSON are available to the simulation without importing
// game packages.
var ArtifactEffects = map[string]ArtifactEffect{
	// ── Wave 1: melee artifacts ────────────────────────────────────────────
	"ironbreaker_gauntlets": {
		Cooldown:         3.0,
		Domain:           "iron",
		DamageMultiplier: 1.5,
		AOERadius:        1.5,
	},
	"shroud_cloak": {
		Cooldown:         8.0,
		Domain:           "shadow",
		DamageMultiplier: 2.0,
		IsBlinkStrike:    true,
		DurationSec:      3.0, // shadow window; shadows_return resets CD mid-window for perma-shadow
	},
	"wardens_medallion": {
		Cooldown:           10.0,
		Domain:             "nature",
		IsTaunt:            true,
		DurationSec:        4.0,
		DamageReductionPct: 30,
		MaxHPBonus:         8,
	},
	"ashbound_chain": {
		Cooldown:    10.0,
		Domain:      "void",
		IsRoot:      true,
		DurationSec: 2.0,
	},
	"grave_reaper": {
		Cooldown:            20.0,
		Domain:              "void",
		IsExecute:           true,
		ExecuteThresholdPct: 20,
	},
	"ember_mantle": {
		Cooldown:        0,
		Domain:          "flame",
		IsPassive:       true,
		BurnDPS:         5,
		BurnDurationSec: 3.0,
	},

	// ── Wave 2: meta-build skill artifacts ────────────────────────────────

	// "The 55" tank core: incoming damage cannot exceed 12% of max HP
	"stone_skin_idol": {
		Cooldown:     0,
		Domain:       "nature",
		IsPassive:    true,
		DamageCapPct: 12,
		MaxHPBonus:   8,
	},
	// Perma-Shadow enabler: hitting an enemy while in shadow reduces Shroud Cloak CD by 3s
	"shadows_return": {
		Cooldown:            0,
		Domain:              "shadow",
		IsPassive:           true,
		ShroudCooldownReset: 3.0,
	},
	// Duration extender: passive +20% to all skill durations
	"void_mirror_pendant": {
		Cooldown:         0,
		Domain:           "void",
		IsPassive:        true,
		SkillDurationPct: 20,
	},
	// Surge Nuke elite: deal 12 × (artifacts currently on cooldown) arcane damage
	"arcane_surge": {
		Cooldown:            8.0,
		Domain:              "arcane",
		SurgeDmgPerCooldown: 12,
	},
	// CDR ring: passive -10% to all active cooldowns
	"arcane_tempo_ring": {
		Cooldown:             0,
		Domain:               "arcane",
		IsPassive:            true,
		CooldownReductionPct: 10,
	},
	// Sacrifice nuke: spend 25% current HP → 200 void damage
	"blood_price": {
		Cooldown:   12.0,
		Domain:     "void",
		HPCostPct:  25,
		DamageFlat: 200,
	},
	// Sacrifice sustain: passive on-kill heal (20% max HP)
	"soul_harvest": {
		Cooldown:      0,
		Domain:        "void",
		IsPassive:     true,
		HealOnKillPct: 20,
	},

	// ── Game spells (registered for engine damage; visuals stay in spells.game.go) ──
	"fireball": {
		Cooldown: 3.0, Domain: "flame", IsProjectile: true, AOERadius: 2.0,
		SpellDamageBase: 25, SpellDamagePerINT: 2.0,
	},
	"arcane_bolt": {
		Cooldown: 0.8, Domain: "arcane", IsProjectile: true,
		SpellDamageBase: 12, SpellDamagePerINT: 1.5,
	},
	"arcane_spray": {
		Cooldown: 1.5, Domain: "arcane", AOERadius: 2.0,
		SpellDamageBase: 8, SpellDamagePerINT: 1.0,
	},
	"lightning": {
		Cooldown: 2.0, Domain: "arcane", IsChain: true, ChainCount: 3,
		SpellDamageBase: 20, SpellDamagePerINT: 1.5,
	},
	// AoE fields: SpellDamageBase is DPS; engine applies DPS × DurationSec as one hit.
	"lightning_storm": {
		Cooldown: 6.0, Domain: "arcane", IsAOEField: true, DurationSec: 3.0,
		SpellDamageBase: 15, SpellDamagePerINT: 1.0,
	},
	"fractal_bloom": {
		Cooldown: 3.0, Domain: "nature", AOERadius: 2.0,
		SpellDamageBase: 18, SpellDamagePerINT: 1.0,
	},
	// fractal_canopy: nature AoE field, 4s; total dmg = DPS × duration
	"fractal_canopy": {
		Cooldown: 8.0, Domain: "nature", IsAOEField: true, DurationSec: 4.0,
		SpellDamageBase: 10, SpellDamagePerINT: 0.8,
	},
	// chaos_ray: high-damage void beam (elite-tier)
	"chaos_ray": {
		Cooldown: 5.0, Domain: "void",
		SpellDamageBase: 40, SpellDamagePerINT: 2.5,
	},
	// slash_combo: rapid melee strikes — knight's active skill, STR-scaling
	"slash_combo": {
		Cooldown: 0.6, Domain: "iron",
		SpellDamageBase: 15, SpellDamagePerSTR: 1.5,
	},

	// ── Build-enabling stat items ───────────────────────────────────────────────
	// arcane_tempo_belt: CDR belt (pairs with arcane_tempo_ring in mage builds)
	"arcane_tempo_belt": {IsPassive: true, Domain: "arcane", CooldownReductionPct: 10},
	// quicksilver_talisman: iron speed talisman — +15% attack speed
	"quicksilver_talisman": {IsPassive: true, Domain: "iron", AttackSpeedPct: 15},
	// iron_will_band: +2 STR ring; adrenaline stacking is future work
	"iron_will_band": {IsPassive: true, Domain: "iron", StrBonus: 2},
	// lifedrinker_robe: +10 MaxHP robe, heal 6 HP flat on kill
	"lifedrinker_robe": {IsPassive: true, Domain: "nature", MaxHPBonus: 10, HealOnKillFlat: 6},
	// hollow_sigil: +6 MaxMana sigil; void skills cost HP instead of mana
	"hollow_sigil": {IsPassive: true, Domain: "void", VoidCostsHP: true},
	// blood_vow_amulet: drain 5 HP/s passively; +20% damage when below 50% HP
	"blood_vow_amulet": {IsPassive: true, Domain: "void", HPDrainPerSec: 5, DamageBonusLowHPPct: 20},
	// marrow_ring: −15 max HP, +12% damage always
	"marrow_ring": {IsPassive: true, Domain: "void", MaxHPMod: -15, DamageBonusPct: 12},
	// resonance_crystal: CDR + skill duration + 8% spell damage per active DoT
	"resonance_crystal": {IsPassive: true, Domain: "arcane", CooldownReductionPct: 8, SkillDurationPct: 10, DoTDmgBonusPct: 8},
	// thornweave_vest: +8 MaxHP; thorn return damage is future work
	"thornweave_vest": {IsPassive: true, Domain: "nature", MaxHPBonus: 8},
	// voidweave_wraps: +12 MaxHP, +20% skill duration
	"voidweave_wraps": {IsPassive: true, Domain: "void", MaxHPBonus: 12, SkillDurationPct: 20},
}

// RegisterArtifact adds or replaces an entry in the engine registry.
// Call this from the game layer after loading artifact JSON so the simulation
// uses real cooldown values rather than the stub fallback.
func RegisterArtifact(id string, effect ArtifactEffect) {
	ArtifactEffects[id] = effect
}

// calcDamage returns the final damage value and whether the hit was a crit.
// baseDmg already includes Strength (baked in by the game layer).
// Crit chance is 5% base; Luck integration is stubbed until the stat is wired.
// Pass guaranteedCrit=true to force a crit (e.g. from shroud_cloak blink).
func (e *DefaultCombatEngine) calcDamage(baseDmg int, guaranteedCrit bool) (int, bool) {
	isCrit := guaranteedCrit || e.rng.Float64() < baseCritChance
	dmg := baseDmg
	if isCrit {
		dmg = int(float64(baseDmg) * critMultiplier)
	}
	return dmg, isCrit
}
