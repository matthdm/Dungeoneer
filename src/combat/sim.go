package combat

import (
	"math/rand/v2"
	"strings"
)

const tickDuration = 1.0 / 60.0 // simulate at 60 fps

// ── Progression helpers ───────────────────────────────────────────────────
// These mirror game/progression so the combat engine stays dependency-free.

// expToLevelSim returns the total EXP required to reach the *next* level.
// Matches entities/Player.AddEXP: threshold = 100*level + 50*level*level.
func expToLevelSim(level int) int {
	return 100*level + 50*level*level
}

// calcEXPRewardSim mirrors progression.CalculateEXPReward.
func calcEXPRewardSim(enemyLevel, playerLevel int) int {
	base := 50 + enemyLevel*10
	diff := float64(enemyLevel - playerLevel)
	mult := 1.0 + 0.2*diff
	if mult < 0.1 {
		mult = 0.1
	}
	return int(float64(base) * mult)
}

// applyStatPoint adds one allocated stat point to the CombatState and updates
// derived stats. HP is also restored by the amount gained (level-up heals).
func applyStatPoint(stat string, state *CombatState) {
	switch stat {
	case "str":
		state.PlayerStrength++
		state.PlayerDamage++
	case "vit":
		state.PlayerMaxHP += 5
		state.PlayerHP += 5
	case "int":
		state.PlayerIntelligence++
		if state.PlayerMaxMana > 0 { // only when the run models mana
			state.PlayerMaxMana += manaPerINT
			state.PlayerMana += manaPerINT
		}
	case "dex":
		state.AttackSpeedPct += 5
	}
}

// statPriorityOrDefault returns the scenario's stat priority, or a sensible default.
func statPriorityOrDefault(s Scenario) []string {
	if len(s.StatPriority) > 0 {
		return s.StatPriority
	}
	switch s.Class {
	case "mage":
		return []string{"int", "int", "vit"}
	default:
		return []string{"str", "str", "vit"}
	}
}

// awardSimEXP awards EXP for a kill, handles level-ups, allocates stat points,
// and appends EventLevelUp events. Call once per EventTargetDied.
func awardSimEXP(state *CombatState, enemyFloor int, priority []string, events *[]Event) {
	reward := calcEXPRewardSim(enemyFloor, state.PlayerLevel)
	state.PlayerEXP += reward
	for state.PlayerEXP >= state.PlayerEXPToNext {
		state.PlayerEXP -= state.PlayerEXPToNext
		state.PlayerLevel++
		state.PlayerEXPToNext = expToLevelSim(state.PlayerLevel)

		// Allocate 3 stat points using the priority list (cycles if shorter than 3).
		n := len(priority)
		pointsDesc := make([]string, 3)
		for i := range 3 {
			stat := priority[i%n]
			applyStatPoint(stat, state)
			pointsDesc[i] = stat
		}
		*events = append(*events, Event{
			Type:  EventLevelUp,
			Value: state.PlayerLevel,
			Tag:   strings.Join(pointsDesc, " "),
		})
	}
}

// enemyHP returns the starting HP for an enemy on the given floor.
func enemyHP(floor int) int {
	if floor < 1 {
		floor = 1
	}
	return 50 + floor*25
}

// enemyDmgPerAttack returns the damage one enemy attack deals.
// Enemies attack every enemyAttackInterval seconds, not every tick.
func enemyDmgPerAttack(floor int) int {
	if floor < 1 {
		floor = 1
	}
	return 10 + floor*3
}

const enemyAttackInterval = 2.0 // seconds between enemy attacks

// buildState constructs an initial CombatState from a scenario.
func buildState(s Scenario, rng *rand.Rand) CombatState {
	state := CombatState{
		PlayerHP:    100 + s.Stats.Vitality*5,
		PlayerMaxHP: 100 + s.Stats.Vitality*5,
		// Base damage includes Strength modifier (game-layer concern simulated here).
		PlayerDamage:       10 + s.Stats.Strength,
		PlayerLevel:        1,
		PlayerEXP:          0,
		PlayerEXPToNext:    expToLevelSim(1),
		PlayerIntelligence: s.Stats.Intelligence,
		PlayerStrength:     s.Stats.Strength,
		DeltaTime:          tickDuration,
	}

	switch s.Class {
	case "mage":
		state.PlayerAttackInterval = 1.0
		state.PlayerAttackRange = 6.0
		state.PlayerClass = "mage"
	default: // knight
		state.PlayerAttackInterval = 0.8
		state.PlayerAttackRange = 1.5
		state.PlayerClass = "knight"
	}

	// Populate artifact slots (up to 7).
	for i, id := range s.Artifacts {
		if i >= len(state.EquippedArtifacts) {
			break
		}
		state.EquippedArtifacts[i] = id
	}

	// Aggregate stat modifiers from equipped artifacts.
	for _, id := range state.EquippedArtifacts {
		eff, ok := ArtifactEffects[id]
		if !ok {
			continue
		}
		// Passive-only modifiers: CDR, attack speed, skill duration.
		if eff.IsPassive {
			state.CooldownReductionPct += eff.CooldownReductionPct
			state.SkillDurationPct += eff.SkillDurationPct
			state.AttackSpeedPct += eff.AttackSpeedPct
		}
		// Flat stat bonuses apply from all equipped items (active or passive).
		// Per design spec: "every item changes the character" — stat bonus is always-on.
		if eff.MaxHPBonus != 0 {
			state.PlayerMaxHP += eff.MaxHPBonus
			state.PlayerHP += eff.MaxHPBonus
		}
		if eff.MaxHPMod != 0 {
			state.PlayerMaxHP += eff.MaxHPMod
			if state.PlayerHP > state.PlayerMaxHP {
				state.PlayerHP = state.PlayerMaxHP
			}
		}
		if eff.StrBonus != 0 {
			state.PlayerStrength += eff.StrBonus
			state.PlayerDamage += eff.StrBonus
		}
	}
	if state.PlayerMaxHP < 1 {
		state.PlayerMaxHP = 1
	}
	if state.PlayerHP < 1 {
		state.PlayerHP = 1
	}

	// Apply enemy build capabilities if a counter-meta archetype is specified.
	if s.EnemyBuildID != "" {
		if eb, ok := EnemyBuilds[s.EnemyBuildID]; ok {
			state.EnemySilenceRadius = eb.Ability.SilenceRadiusTiles
			state.EnemyDetectionRadius = eb.Ability.DetectionRadiusTiles
			state.EnemyInstakillPct = eb.Ability.InstakillThresholdPct
			state.EnemyDamageCapBypass = eb.Ability.DamageCapBypass
			state.EnemySacrificeLeech = eb.Ability.SacrificeLeech
			state.EnemyHealReductionPct = eb.Ability.HealReductionPct
			state.EnemyHPRegenPerSec = eb.Ability.HPRegenPerSec
			state.EnemyBlockBlink = eb.Ability.BlockBlink
			state.EnemyPackBonusPct = eb.Ability.PackAuraBonusPct
		}
	}

	// Enemy starts in range and targeted.
	state.HasTarget = true
	state.TargetInRange = true
	state.IsAutoAttacking = true
	state.TargetHP = enemyHP(s.Floor)
	state.TargetMaxHP = state.TargetHP
	state.TargetLevel = s.Floor
	state.TargetName = pickRandom(s.EnemyPool, rng, "enemy")

	// Auto-attack fires immediately on first tick.
	state.AutoAttackTimer = 0

	return state
}
