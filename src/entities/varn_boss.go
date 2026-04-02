package entities

import "dungeoneer/sprites"

// NewVarnChainkeeper creates the Varn ascended boss entity.
//
// Phase 0 (HP > 50%): GreyKnight sprite — controlled, chain whip + ranged throw.
// Phase 1 (50% → 25%): sprite permanently swaps to Sentinel — chain pull + eruption + frenzy.
// Phase 2 (HP < 25%): same Sentinel sprite — all attacks faster, higher damage.
//
// isNGPlus: if true (boss defeated before), phase 0 uses TorturedSoul sprite instead
// of GreyKnight, reflecting the toll of repeated death.
func NewVarnChainkeeper(ss *sprites.SpriteSheet, x, y int, isNGPlus bool) *Boss {
	phase1Sprite := ss.GreyKnight
	if isNGPlus {
		phase1Sprite = ss.TorturedSoul
	}

	m := &Monster{
		Name:             "Varn",
		TileX:            x,
		TileY:            y,
		InterpX:          float64(x),
		InterpY:          float64(y),
		Sprite:           phase1Sprite,
		MovementDuration: 18, // faster than the generic guardian
		LeftFacing:       true,
		HP:               250,
		MaxHP:            250,
		Damage:           12,
		HitRadius:        DefaultMonsterHitRadius,
		AttackRate:       25,
		Level:            12,
		Role:             "boss",
	}

	boss := &Boss{
		Monster:            m,
		Title:              "Warden Varn, The Chainkeeper",
		NPCID:               "varn",
		PreFightDialogueID:  "varn_boss_pre",
		PostFightDialogueID: "varn_boss_post",
		MaxPhases:          3,
		PhaseHP:            []float64{0.5, 0.25},
		Patterns: [][]BossAttack{
			// Phase 0 — The Chainkeeper restrained: whip, ranged chain bolt, and a slow telegraphed pull.
			{
				{ID: "chain_whip", Type: "melee", Damage: 12, Range: 2.5, Cooldown: 2.5},
				{ID: "chain_bolt", Type: "projectile", Damage: 8, Range: 7.0, Cooldown: 3.5},
				{ID: "chain_pull", Type: "pull_player", Damage: 6, Range: 6.0, Cooldown: 7.0},
			},
			// Phase 1 — Unchained: signature pull + eruption + rapid frenzy.
			{
				{ID: "chain_pull", Type: "pull_player", Damage: 10, Range: 8.0, Cooldown: 4.0},
				{ID: "chain_eruption", Type: "aoe", Damage: 15, AOERadius: 3, Cooldown: 5.0},
				{ID: "chain_frenzy", Type: "melee", Damage: 8, Range: 1.5, Cooldown: 1.0},
			},
			// Phase 2 — Final frenzy: longer reach, heavier damage, shorter cooldowns.
			{
				{ID: "chain_pull", Type: "pull_player", Damage: 15, Range: 10.0, Cooldown: 2.5},
				{ID: "chain_eruption", Type: "aoe", Damage: 20, AOERadius: 4, Cooldown: 3.5},
				{ID: "chain_frenzy", Type: "melee", Damage: 12, Range: 1.5, Cooldown: 0.7},
			},
		},
	}

	// Sprite swap on phase transition: GreyKnight/TorturedSoul → Sentinel.
	// This is permanent — once unchained, he cannot return to what he was.
	boss.OnPhaseTransition = func(newPhase int) {
		if newPhase >= 1 {
			m.Sprite = ss.Sentinel
		}
	}

	m.Behavior = NewBossBehavior(boss)
	return boss
}
