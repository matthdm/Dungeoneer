package game

// BossType identifies which boss entity to spawn on the final floor.
type BossType int

const (
	BossGeneric BossType = iota // fallback dungeon guardian
	BossVarn                    // Varn the Chainkeeper (full ascension arc)
)

// SelectBoss returns the appropriate boss for the current run state.
// BossVarn is chosen when the player has completed Varn's phase 3 ascension
// dialogue (varn_p3_done flag set in varn_phase3.json farewell node).
func SelectBoss(rs *RunState) BossType {
	if rs != nil && rs.QuestFlags["varn_p3_done"] >= 1 {
		return BossVarn
	}
	return BossGeneric
}
