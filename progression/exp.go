package progression

// EXPToLevel returns the experience required to reach the given level.
func EXPToLevel(level int) int {
	return 100*level + 50*level*level
}

// CalculateEXPReward computes the amount of EXP granted for defeating an enemy
// of a particular level relative to the player's level.
func CalculateEXPReward(enemyLevel, playerLevel int) int {
	base := 50
	diff := enemyLevel - playerLevel
	multiplier := 1.0 + 0.2*float64(diff)
	if multiplier < 0.1 {
		multiplier = 0.1
	}
	return int(float64(base+enemyLevel*10) * multiplier)
}
