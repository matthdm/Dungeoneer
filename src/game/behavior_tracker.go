package game

// BehaviorRecord captures one run's observed play patterns.
type BehaviorRecord struct {
	SpellUsage   map[string]int `json:"spell_usage"`   // spell ID → cast count
	MeleeKills   int            `json:"melee_kills"`
	RangedKills  int            `json:"ranged_kills"`
	TotalKills   int            `json:"total_kills"`
	TotalEnemies int            `json:"total_enemies"` // enemies spawned this run
	Deaths       int            `json:"deaths"`        // 0 = survived, 1 = died
}

// KillRatio returns kills/enemies, or 1.0 if no enemies.
func (r *BehaviorRecord) KillRatio() float64 {
	if r.TotalEnemies == 0 {
		return 1.0
	}
	return float64(r.TotalKills) / float64(r.TotalEnemies)
}

// RangedRatio returns ranged kills / total kills, or 0.0 if no kills.
func (r *BehaviorRecord) RangedRatio() float64 {
	if r.TotalKills == 0 {
		return 0.0
	}
	return float64(r.RangedKills) / float64(r.TotalKills)
}

// SpellVariety returns the number of distinct spells cast.
func (r *BehaviorRecord) SpellVariety() int {
	return len(r.SpellUsage)
}

// BehaviorTracker accumulates behavior data for the current run.
type BehaviorTracker struct {
	Record BehaviorRecord
	active bool
}

// Start initializes the tracker for a new run.
func (bt *BehaviorTracker) Start(totalEnemies int) {
	bt.Record = BehaviorRecord{
		SpellUsage:   make(map[string]int),
		TotalEnemies: totalEnemies,
	}
	bt.active = true
}

// RecordSpellCast records a successful spell cast.
func (bt *BehaviorTracker) RecordSpellCast(spellID string) {
	if !bt.active {
		return
	}
	bt.Record.SpellUsage[spellID]++
}

// RecordKill records an enemy kill. rangedRole should be true for "ranged"/"caster" roles.
func (bt *BehaviorTracker) RecordKill(rangedRole bool) {
	if !bt.active {
		return
	}
	bt.Record.TotalKills++
	if rangedRole {
		bt.Record.RangedKills++
	} else {
		bt.Record.MeleeKills++
	}
}

// RecordEnemySpawned increments total enemies for this run.
func (bt *BehaviorTracker) RecordEnemySpawned() {
	if !bt.active {
		return
	}
	bt.Record.TotalEnemies++
}

// Finalize stops tracking and returns the completed record.
func (bt *BehaviorTracker) Finalize(died bool) BehaviorRecord {
	bt.active = false
	if died {
		bt.Record.Deaths = 1
	}
	return bt.Record
}
