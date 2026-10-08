package combat

import (
	"io"
	"math/rand/v2"
)

// This file is the simulator's API for cmd/combatvis: one fight that can be
// stepped a tick at a time and inspected between ticks. It runs the same
// simRun the benchmarker aggregates, so what the visualizer shows is one of
// the iterations behind a benchmark number.

// EnemyAttackInterval is the seconds between one enemy's attacks.
const EnemyAttackInterval = enemyAttackInterval

// Sim is a single simulated fight.
type Sim struct {
	run    *simRun
	totals simTotals
}

// WaveEnemy is an enemy in the wave other than the one being fought.
type WaveEnemy struct {
	HP, MaxHP int
}

// NewSim starts the fight the benchmarker runs as iteration `iteration` of the
// scenario under the given seed. trace, if not nil, receives one line per event.
func NewSim(s Scenario, seed uint64, iteration int, trace io.Writer) *Sim {
	sim := &Sim{totals: simTotals{dmgByTag: make(map[string]int64)}}
	is := iterSeed(s.Name, seed, iteration)
	sim.run = newSimRun(s, NewDefaultCombatEngine(is), rand.New(rand.NewPCG(is, is^0xcafebabe)), &sim.totals, trace)
	return sim
}

// Step advances the fight by one tick (1/60 s). It returns false, and does
// nothing, once the player is dead.
func (s *Sim) Step() bool {
	if s.run.stats.died {
		return false
	}
	return s.run.step()
}

// State is the combat state after the last step. Treat it as read-only.
func (s *Sim) State() *CombatState { return &s.run.state }

func (s *Sim) Time() float64    { return s.run.time }
func (s *Sim) Dead() bool       { return s.run.stats.died }
func (s *Sim) Kills() int       { return s.run.stats.kills }
func (s *Sim) DamageTaken() int { return s.run.stats.dmgTaken }
func (s *Sim) Healed() int      { return s.run.stats.healed }

// DamageDealt is the damage that removed enemy HP (overkill excluded).
func (s *Sim) DamageDealt() int { return s.run.stats.effDmg }

// Floor is the floor enemies are currently scaled to (it rises under a ramp).
func (s *Sim) Floor() int { return s.run.floorNow() }

// EnemyBaseDamage is one enemy hit at the current floor, before variance,
// damage reduction and caps.
func (s *Sim) EnemyBaseDamage() int { return enemyDmgPerAttack(s.run.floorNow()) }

// AttackTimers is the seconds until each enemy's next attack. Index 0 is the
// enemy being fought; the rest line up with Wave. Treat it as read-only.
func (s *Sim) AttackTimers() []float64 { return s.run.atkTimers }

// Wave appends the enemies not currently being fought to dst and returns it.
func (s *Sim) Wave(dst []WaveEnemy) []WaveEnemy {
	for _, m := range s.run.others {
		dst = append(dst, WaveEnemy{HP: m.hp, MaxHP: m.maxHP})
	}
	return dst
}
