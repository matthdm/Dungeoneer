package combat

// This file exports internal symbols needed by cmd/combatvis.
// It is NOT part of the benchmarker or game-layer API.

// EnemyAttackInterval is the seconds between enemy attacks (exported for combatvis).
const EnemyAttackInterval = enemyAttackInterval

// TickDurationExported is the simulation tick duration in seconds.
const TickDurationExported = tickDuration

// EnemyHPExported returns starting HP for an enemy on the given floor.
func EnemyHPExported(floor int) int { return enemyHP(floor) }

// EnemyDmgPerAttackExported returns the damage per enemy attack on the given floor.
func EnemyDmgPerAttackExported(floor int) int { return enemyDmgPerAttack(floor) }

// BuildStateExported builds a CombatState from a scenario using a deterministic seed.
// Used by combatvis for single-run replays.
func BuildStateExported(s Scenario) CombatState {
	return buildState(s, deterministicRNG(s.Name))
}

// ParseSlotIdxExported parses "slot_N" and returns the 0-based index.
func ParseSlotIdxExported(entry string) int { return parseSlotIdx(entry) }

// AwardSimEXPExported calls awardSimEXP for combatvis. The events slice may grow
// with EventLevelUp entries that the caller should process after the call.
func AwardSimEXPExported(state *CombatState, floor int, s Scenario, events *[]Event) {
	awardSimEXP(state, floor, statPriorityOrDefault(s), events)
}
