package combat

// EnemyAbility describes one counter-mechanic an enemy can use against the player.
// These mirror ArtifactEffect fields but apply in reverse: they constrain or
// punish specific player build strategies rather than buffing the player.
type EnemyAbility struct {
	SilenceRadiusTiles   float64 // Silencer: player skills blocked within N tiles
	DetectionRadiusTiles float64 // Veilbane: ejects player from shadow within N tiles
	InstakillThresholdPct int    // Judge: if player HP% ≤ this, next hit kills
	DamageCapBypass      bool    // Bloodhound: ignores stone_skin_idol cap
	SacrificeLeech       bool    // Crucible: heals when player spends HP on blood_price
	HealReductionPct     int     // Nullifier: reduces player healing by N%
	HPRegenPerSec        int     // Regenerator: enemy HP regen each second
	BlockBlink           bool    // Gravity Warden: prevents shroud_cloak teleport
	PackAuraBonusPct     int     // Pack Leader: +N% damage per additional pack member in range
}

// EnemyBuild defines an enemy archetype that counter-picks specific player strategies.
// The Living Dungeon AI (Phase 7B) selects from this pool based on player behavior tracking.
type EnemyBuild struct {
	ID            string
	Domain        string
	Ability       EnemyAbility
	CounterBuilds []string // player build IDs this enemy is designed to punish
}

// EnemyBuilds is the registry of counter-meta enemy archetypes.
// Keyed by enemy build ID.
var EnemyBuilds = map[string]EnemyBuild{
	// The Silencer — punishes builds that rely on tight cooldown timing.
	// Blocks player skill casts within 3 tiles. Counter to Arcane Surge (M5) and CC Chain (M1).
	"the_silencer": {
		ID:     "the_silencer",
		Domain: "arcane",
		Ability: EnemyAbility{
			SilenceRadiusTiles: 3.0,
		},
		CounterBuilds: []string{"arcane_surge_build", "cc_chain"},
	},

	// The Veilbane — detection field ejects players from shadow form.
	// Counter to Perma Shadow (M7), Shadow Burst (W4), CC Chain (M1).
	"the_veilbane": {
		ID:     "the_veilbane",
		Domain: "shadow",
		Ability: EnemyAbility{
			DetectionRadiusTiles: 4.0,
		},
		CounterBuilds: []string{"perma_shadow", "shadow_burst", "cc_chain"},
	},

	// The Judge — next hit instakills if player HP drops below 25%.
	// Counter to The 55 (M6) which intentionally runs at ~30% HP, and Void Sacrifice (M8).
	"the_judge": {
		ID:     "the_judge",
		Domain: "void",
		Ability: EnemyAbility{
			InstakillThresholdPct: 25,
		},
		CounterBuilds: []string{"the_55", "void_sacrifice"},
	},

	// The Regenerator — 40 HP/s regen; slow DPS builds can't out-damage it.
	// Forces burst builds; counters Arcane Farmer (W2), Tank Mage (W6), Nature Sustain (W3).
	"the_regenerator": {
		ID:     "the_regenerator",
		Domain: "nature",
		Ability: EnemyAbility{
			HPRegenPerSec: 40,
		},
		CounterBuilds: []string{"arcane_farmer", "tank_mage", "nature_sustain"},
	},

	// The Pack Leader — aura gives pack members +15% damage per ally in range.
	// 3 Pack Leaders in range = +45% damage. Forces AoE; counters single-target builds.
	"the_pack_leader": {
		ID:     "the_pack_leader",
		Domain: "iron",
		Ability: EnemyAbility{
			PackAuraBonusPct: 15,
		},
		CounterBuilds: []string{"shadow_burst", "the_55", "grapple_momentum"},
	},

	// The Bloodhound — bypasses all DR and damage caps when player HP < 50%.
	// Kites at range. Counter to The 55 (M6) and Void Sacrifice (M8).
	"the_bloodhound": {
		ID:     "the_bloodhound",
		Domain: "void",
		Ability: EnemyAbility{
			DamageCapBypass: true,
		},
		CounterBuilds: []string{"the_55", "void_sacrifice"},
	},

	// The Nullifier — reduces all player healing by 75% for 8s.
	// Counter to any sustain build: Grapple Momentum (M3), The 55 (M6), soul_harvest builds.
	"the_nullifier": {
		ID:     "the_nullifier",
		Domain: "arcane",
		Ability: EnemyAbility{
			HealReductionPct: 75,
		},
		CounterBuilds: []string{"grapple_momentum", "the_55", "void_sacrifice", "nature_sustain"},
	},

	// The Crucible — heals when the player spends HP on blood_price or void skills.
	// Counter to Chaos Knight (M4) and Void Sacrifice (M8).
	"the_crucible": {
		ID:     "the_crucible",
		Domain: "void",
		Ability: EnemyAbility{
			SacrificeLeech: true,
		},
		CounterBuilds: []string{"chaos_knight", "void_sacrifice"},
	},

	// The Gravity Warden — creates a zone that blocks blink/teleport skills.
	// Counter to Shadow Burst (W4) and Perma Shadow (M7) — both rely on shroud_cloak.
	"the_gravity_warden": {
		ID:     "the_gravity_warden",
		Domain: "iron",
		Ability: EnemyAbility{
			BlockBlink: true,
		},
		CounterBuilds: []string{"shadow_burst", "perma_shadow", "cc_chain"},
	},
}

// GetCounterBuilds returns enemy build IDs that counter the given player build ID.
// Used by the Living Dungeon AI (Phase 7B) to select a Nemesis enemy.
func GetCounterBuilds(playerBuildID string) []string {
	var result []string
	for id, build := range EnemyBuilds {
		for _, c := range build.CounterBuilds {
			if c == playerBuildID {
				result = append(result, id)
				break
			}
		}
	}
	return result
}
