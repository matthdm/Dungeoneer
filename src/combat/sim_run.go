package combat

import (
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"math/rand/v2"
	"runtime"
	"strings"
	"sync"
)

// Simulation model constants. Mana mirrors entities.Player (MaxMana =
// 20 + INT*5, regen = 1 + INT/5 per second) so the engine stays dependency-free.
const (
	defaultSimDurationSec = 120.0
	enemyDmgVariance      = 0.2 // each enemy hit lands within ±20% of its base damage
	baseMana              = 20
	manaPerINT            = 5
	knightSimDistance     = 1.0 // tiles from the enemy: toe to toe
	mageSimDistance       = 5.0 // tiles from the enemy: inside every spell's CastRange
)

// SimOptions tunes a Simulate call. The zero value runs the scenario as written.
type SimOptions struct {
	// Engine is used for every iteration when set. When nil each iteration
	// gets its own DefaultCombatEngine seeded from the scenario name, Seed and
	// the iteration number, so a result does not depend on which scenarios ran
	// before it and any single iteration can be replayed.
	Engine CombatEngine
	// Seed varies the random stream. The same scenario and seed always give
	// the same result.
	Seed uint64
	// DurationSec is the length of one iteration. 0 means two minutes.
	DurationSec float64
	// Trace, when set, receives a timestamped event log of iteration
	// TraceIteration.
	Trace          io.Writer
	TraceIteration int
}

// Reasons a ready skill did not fire, beyond the engine's Fail* constants.
const (
	BlockNoMana = "no_mana"       // not enough mana for the cast
	BlockEnemy  = "enemy_blocked" // silenced or blink-blocked by the enemy build
	AutoAttack  = "auto_attack"   // SimResult.Skills key for basic attacks
)

// blockReasons indexes the reasons a ready skill press can come to nothing.
var blockReasons = [...]string{
	FailNoTarget, FailOutOfRange, FailNoLineOfSight, FailTargetNotLow,
	FailNothingOnCooldown, BlockNoMana, BlockEnemy,
}

func blockReasonIdx(reason string) int {
	for i, r := range blockReasons {
		if r == reason {
			return i
		}
	}
	return len(blockReasons) - 1
}

// splitmix64 is a bit mixer used to derive independent seeds.
func splitmix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

// iterSeed derives the seed for one iteration of one scenario.
func iterSeed(name string, seed uint64, iter int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(name))
	v := splitmix64(h.Sum64() ^ splitmix64(seed) ^ splitmix64(uint64(iter)+1))
	if v == 0 {
		v = 0xdeadbeef
	}
	return v
}

// pickRandom returns a random element from a slice, or fallback if empty.
func pickRandom(pool []string, rng *rand.Rand, fallback string) string {
	if len(pool) == 0 {
		return fallback
	}
	return pool[rng.IntN(len(pool))]
}

// parseSlotIdx parses "slot_N" and returns N-1 (0-based). Returns -1 on error.
func parseSlotIdx(entry string) int {
	rest, ok := strings.CutPrefix(entry, "slot_")
	if !ok {
		return -1
	}
	var n int
	if _, err := fmt.Sscanf(rest, "%d", &n); err != nil || n < 1 || fmt.Sprint(n) != rest {
		return -1
	}
	return n - 1
}

// iterStats is what one iteration measured.
type iterStats struct {
	kills     int
	rawDmg    int // damage rolled, including overkill
	effDmg    int // damage that actually removed enemy HP
	dmgTaken  int
	healed    int
	maxStreak int
	minHPPct  float64
	timeAlive float64
	died      bool
	level     int
}

// simTotals accumulates per-skill figures across every iteration of a run.
type simTotals struct {
	dmgByTag map[string]int64
	casts    [maxArtifactSlots]int64
	blocked  [maxArtifactSlots][len(blockReasons)]int64 // ready presses that did nothing, by reason
}

// packMember is an enemy in the wave that the player is not currently fighting.
type packMember struct {
	hp, maxHP int
}

// simRun is one iteration: a single player build fighting a stream of enemies.
type simRun struct {
	s        Scenario
	engine   CombatEngine
	rng      *rand.Rand
	state    CombatState
	priority []string
	totals   *simTotals
	trace    io.Writer

	rotIdx          int
	time            float64
	atkTimers       []float64    // one per enemy in the wave; index 0 is the current target
	others          []packMember // the rest of the wave (Pack-1 enemies)
	enemyRegenAccum float64
	manaRegenAccum  float64
	minHP           int
	lastBlock       [maxArtifactSlots]string // trace: last reported block reason per slot
	stats           iterStats
}

func newSimRun(s Scenario, engine CombatEngine, rng *rand.Rand, totals *simTotals, trace io.Writer) *simRun {
	r := &simRun{
		s:        s,
		engine:   engine,
		rng:      rng,
		state:    buildState(s, rng),
		priority: statPriorityOrDefault(s),
		totals:   totals,
		trace:    trace,
	}
	st := &r.state

	// The wave: every enemy attacks on its own timer, spread evenly across the
	// attack interval so a pack lands a steady stream of hits rather than volleys.
	pack := max(s.Pack, 1)
	r.atkTimers = make([]float64, pack)
	for i := range r.atkTimers {
		r.atkTimers[i] = enemyAttackInterval * float64(i+1) / float64(pack)
	}
	r.others = make([]packMember, pack-1)
	for i := range r.others {
		r.others[i] = r.freshEnemy()
	}

	// Mana: a build that cannot pay for its rotation should show it.
	st.PlayerMaxMana = baseMana + st.PlayerIntelligence*manaPerINT
	if st.PlayerMaxMana < 1 {
		st.PlayerMaxMana = 1 // 0 would switch the engine's mana rules off
	}
	st.PlayerMana = st.PlayerMaxMana

	// Positions: the player stands still at a class-appropriate distance, which
	// is what cast ranges and the enemy silence/detection radii are tested against.
	dist := s.Distance
	if dist == 0 {
		dist = knightSimDistance
		if st.PlayerClass == "mage" {
			dist = mageSimDistance
		}
	}
	st.PlayerX, st.PlayerY = 0, 0
	st.TargetX, st.TargetY = dist, 0
	st.TargetDist = dist
	st.TargetInRange = dist <= st.PlayerAttackRange

	r.minHP = st.PlayerHP
	r.stats.minHPPct = 1
	r.stats.level = st.PlayerLevel
	return r
}

func (r *simRun) logf(format string, args ...any) {
	if r.trace == nil {
		return
	}
	// One Write per line, so a trace sink can treat each call as an event.
	fmt.Fprintf(r.trace, "[%7.2fs] %s\n", r.time, fmt.Sprintf(format, args...))
}

// floorNow is the floor enemies are currently scaled to: the scenario floor,
// plus one for every RampSec seconds the fight has lasted.
func (r *simRun) floorNow() int {
	if r.s.RampSec <= 0 {
		return r.s.Floor
	}
	return r.s.Floor + int(r.time/r.s.RampSec)
}

func (r *simRun) freshEnemy() packMember {
	hp := enemyHP(r.floorNow())
	return packMember{hp: hp, maxHP: hp}
}

// nextTarget puts the next enemy in front of the player after a kill. In a
// wave the player turns to the most wounded enemy left and a fresh one joins.
func (r *simRun) nextTarget() {
	next := r.freshEnemy()
	if len(r.others) > 0 {
		low := 0
		for i, m := range r.others {
			if m.hp < r.others[low].hp {
				low = i
			}
		}
		next, r.others[low] = r.others[low], next
	}
	st := &r.state
	st.TargetHP, st.TargetMaxHP = next.hp, next.maxHP
	st.TargetIsDead = false
	st.HasTarget = true
	st.IsAutoAttacking = true
	st.TargetName = pickRandom(r.s.EnemyPool, r.rng, "enemy")
	r.enemyRegenAccum = 0
}

// countKill records a kill and, when the scenario allows it, the EXP for it.
func (r *simRun) countKill() {
	r.stats.kills++
	if !r.s.Leveling {
		return
	}
	var lvl []Event
	awardSimEXP(&r.state, r.s.Floor, r.priority, &lvl)
	for _, lev := range lvl {
		r.logf("LEVEL %d (+%s)", lev.Value, lev.Tag)
	}
	r.stats.level = r.state.PlayerLevel
}

// splashCount is how many enemies besides the target a hit from this artifact
// also strikes: all of them for an area skill, ChainCount-1 for a chain.
func splashCount(eff ArtifactEffect, others int) int {
	switch {
	case eff.AOERadius > 0 || eff.IsAOEField:
		return others
	case eff.IsChain:
		return min(max(eff.ChainCount-1, 0), others)
	}
	return 0
}

// splash repeats an area or chain hit on the rest of the wave. The engine only
// knows about one target, so kills made here apply the streak and on-kill
// passives the engine would have applied.
func (r *simRun) splash(ev Event) {
	n := splashCount(ArtifactEffects[ev.Tag], len(r.others))
	st := &r.state
	for i := range n {
		m := &r.others[i]
		eff := min(ev.Value, m.hp)
		m.hp -= ev.Value
		r.stats.rawDmg += ev.Value
		r.stats.effDmg += eff
		r.totals.dmgByTag[ev.Tag] += int64(eff)
		if m.hp > 0 {
			continue
		}
		st.KillStreak++
		st.StreakTimer = streakResetTime
		r.stats.maxStreak = max(r.stats.maxStreak, st.KillStreak)
		var passive []Event
		applyKillPassives(st, &passive)
		for _, p := range passive {
			if p.Type == EventDamageTaken && p.Value < 0 {
				r.stats.healed += -p.Value
			}
		}
		r.countKill()
		r.logf("KILL #%d by %s splash (streak x%d)", r.stats.kills, ev.Tag, st.KillStreak)
		*m = r.freshEnemy()
	}
}

// manaShort reports whether the artifact's cast would be refused for mana,
// following the engine's rules (hollow_sigil makes void skills cost HP instead).
func manaShort(state *CombatState, eff ArtifactEffect) bool {
	if eff.ManaCost <= 0 || state.PlayerMaxMana <= 0 {
		return false
	}
	if eff.Domain == "void" && eff.HPCostPct == 0 {
		paysHP := false
		scanArtifacts(state, func(_ string, eq ArtifactEffect) bool {
			paysHP = eq.VoidCostsHP
			return !paysHP
		})
		if paysHP {
			return false
		}
	}
	cost := eff.ManaCost
	if state.ManaCostReductionPct > 0 {
		cost = int(float64(cost) * (1.0 - float64(state.ManaCostReductionPct)/100.0))
	}
	return state.PlayerMana < cost
}

// step advances the fight by one tick. It returns false once the player is dead.
func (r *simRun) step() bool {
	st := &r.state

	// Player input: the rotation presses one entry per tick, so a skill goes
	// off the moment its cooldown, mana and target allow.
	var actions []Action
	slot := -1
	if n := len(r.s.SkillRotation); n > 0 {
		entry := r.s.SkillRotation[r.rotIdx%n]
		r.rotIdx++
		if entry != "auto" {
			slot = parseSlotIdx(entry)
		}
	}
	ready, short := false, false
	if slot >= 0 && slot < maxArtifactSlots && st.EquippedArtifacts[slot] != "" {
		actions = []Action{{Type: ActionActivateSkill, SlotIdx: slot}}
		ready = st.ArtifactCooldowns[slot] <= 0
		short = manaShort(st, ArtifactEffects[st.EquippedArtifacts[slot]])
	} else {
		slot = -1
	}

	targetHP := st.TargetHP
	newState, events := r.engine.Tick(*st, actions)
	*st = newState
	r.time += tickDuration

	fired, failReason := false, ""
	for _, ev := range events {
		switch ev.Type {
		case EventDamageDealt:
			eff := min(ev.Value, max(targetHP, 0))
			targetHP -= ev.Value
			r.stats.rawDmg += ev.Value
			r.stats.effDmg += eff
			tag := ev.Tag
			if tag == "" {
				tag = AutoAttack
			}
			r.totals.dmgByTag[tag] += int64(eff)
			if r.trace != nil {
				crit := ""
				if ev.IsCrit {
					crit = " CRIT"
				}
				r.logf("%-22s hits for %d%s (enemy %d/%d)", tag, ev.Value, crit, max(targetHP, 0), st.TargetMaxHP)
			}
			if len(r.others) > 0 && ev.Tag != "" {
				r.splash(ev)
			}
		case EventTargetDied:
			r.countKill()
			r.logf("KILL #%d (streak x%d)", r.stats.kills, ev.Value)
			// The next enemy steps in. Cooldowns carry over so the numbers
			// reflect sustained play, and so do the enemy attack timers: a
			// kill does not buy a free two seconds without incoming damage.
			r.nextTarget()
			targetHP = st.TargetHP
		case EventStreakChange:
			if ev.Value > r.stats.maxStreak {
				r.stats.maxStreak = ev.Value
			}
		case EventHPSpent:
			r.logf("spent %d HP (player %d/%d)", ev.Value, st.PlayerHP, st.PlayerMaxHP)
			if st.EnemySacrificeLeech && !st.TargetIsDead {
				st.TargetHP = min(st.TargetHP+ev.Value, st.TargetMaxHP)
			}
		case EventDamageTaken:
			if ev.Value < 0 { // negative = heal
				r.stats.healed += -ev.Value
				r.logf("healed %d (player %d/%d)", -ev.Value, st.PlayerHP, st.PlayerMaxHP)
			}
		case EventSkillFired:
			fired = true
			r.logf("CAST %s", ev.Tag)
		case EventSkillFailed:
			failReason = ev.Reason
		}
	}

	if slot >= 0 {
		switch {
		case fired:
			r.totals.casts[slot]++
			r.lastBlock[slot] = ""
		case ready:
			// The skill was off cooldown and still did nothing.
			reason := failReason
			if reason == "" {
				reason = BlockEnemy
				if short {
					reason = BlockNoMana
				}
			}
			r.totals.blocked[slot][blockReasonIdx(reason)]++
			if r.trace != nil && r.lastBlock[slot] != reason {
				r.lastBlock[slot] = reason
				r.logf("%s ready but not cast: %s", st.EquippedArtifacts[slot], reason)
			}
		}
	}

	r.enemyAttack()
	r.enemyRegen()
	r.manaRegen()

	if st.PlayerHP < r.minHP {
		r.minHP = st.PlayerHP
		r.stats.minHPPct = math.Max(0, float64(st.PlayerHP)/float64(st.PlayerMaxHP))
	}
	if st.PlayerHP <= 0 {
		r.stats.died = true
		r.logf("PLAYER DIED after %d kills", r.stats.kills)
		return false
	}
	return true
}

// enemyAttack advances every enemy's attack timer and lands the hits that are due.
func (r *simRun) enemyAttack() {
	st := &r.state
	for i := range r.atkTimers {
		r.atkTimers[i] -= tickDuration
		if r.atkTimers[i] > 0 || !st.HasTarget || st.TargetIsDead || st.PlayerHP <= 0 {
			continue
		}
		r.atkTimers[i] += enemyAttackInterval
		r.enemyHit(i == 0)
	}
}

// enemyHit resolves one enemy attack. Only the enemy the player is fighting
// can be held by a root; the rest of the wave attacks regardless.
func (r *simRun) enemyHit(isTarget bool) {
	st := &r.state
	if isTarget && st.TargetRooted {
		r.logf("enemy attack missed (rooted)")
		return
	}
	if st.InShadow {
		r.logf("enemy attack missed (player in shadow)")
		return
	}

	spread := 1 + (r.rng.Float64()*2-1)*enemyDmgVariance
	dmg := int(math.Round(float64(enemyDmgPerAttack(r.floorNow())) * spread))
	if st.DamageReductionPct > 0 {
		dmg = max(int(float64(dmg)*(1.0-float64(st.DamageReductionPct)/100.0)), 0)
	}
	if dmg <= 0 {
		return
	}

	rawDmg := dmg
	for _, id := range st.EquippedArtifacts {
		if eff, ok := ArtifactEffects[id]; ok && eff.DamageCapPct > 0 {
			dmg = min(dmg, max(int(float64(st.PlayerMaxHP)*float64(eff.DamageCapPct)/100.0), 1))
			break
		}
	}
	if st.EnemyDamageCapBypass && st.PlayerHP*2 < st.PlayerMaxHP {
		dmg = rawDmg
	}
	if st.EnemyPackBonusPct > 0 {
		// Pack Leader: the bonus is per ally. A lone enemy is assumed to have one.
		allies := max(len(r.others), 1)
		dmg += int(float64(dmg) * float64(st.EnemyPackBonusPct*allies) / 100.0)
	}

	st.PlayerHP -= dmg
	r.stats.dmgTaken += dmg
	r.logf("enemy hits for %d (player %d/%d)", dmg, max(st.PlayerHP, 0), st.PlayerMaxHP)

	if st.EnemyInstakillPct > 0 && st.PlayerHP > 0 && st.PlayerHP <= st.PlayerMaxHP*st.EnemyInstakillPct/100 {
		st.PlayerHP = 0
		r.logf("enemy executes the player (HP at or below %d%%)", st.EnemyInstakillPct)
	}
	if st.KillStreak > 0 {
		st.KillStreak = 0
		st.StreakTimer = 0
	}
}

// enemyRegen heals the enemy (Regenerator). Fractions accumulate across ticks:
// at 60 Hz a 40 HP/s regen is under one point per tick.
func (r *simRun) enemyRegen() {
	st := &r.state
	if st.EnemyHPRegenPerSec <= 0 || st.TargetIsDead {
		return
	}
	r.enemyRegenAccum += float64(st.EnemyHPRegenPerSec) * tickDuration
	if whole := int(r.enemyRegenAccum); whole > 0 {
		r.enemyRegenAccum -= float64(whole)
		st.TargetHP = min(st.TargetHP+whole, st.TargetMaxHP)
	}
}

// manaRegen mirrors entities.Player's passive mana regeneration.
func (r *simRun) manaRegen() {
	st := &r.state
	if st.PlayerMana >= st.PlayerMaxMana {
		r.manaRegenAccum = 0
		return
	}
	r.manaRegenAccum += float64(1+st.PlayerIntelligence/5) * tickDuration
	if whole := int(r.manaRegenAccum); whole > 0 {
		r.manaRegenAccum -= float64(whole)
		st.PlayerMana = min(st.PlayerMana+whole, st.PlayerMaxMana)
	}
}

// SimulateAll simulates every scenario, in parallel, and returns results in
// the same order. Each simulation is self-seeded, so the order of execution
// does not change any result. opts.Engine and opts.Trace must be nil: they
// are not safe to share between goroutines.
func SimulateAll(scenarios []Scenario, opts SimOptions) []SimResult {
	results := make([]SimResult, len(scenarios))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(runtime.NumCPU(), len(scenarios)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = Simulate(scenarios[i], opts)
			}
		}()
	}
	for i := range scenarios {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// RunSimulation runs a Scenario with one shared engine. Kept for callers that
// supply their own engine; Simulate is the reproducible entry point.
func RunSimulation(engine CombatEngine, scenario Scenario) SimResult {
	return Simulate(scenario, SimOptions{Engine: engine})
}

// Simulate runs a Scenario for its iteration count and returns the aggregate
// SimResult. Each iteration is one continuous fight against a stream of
// enemies; cooldowns persist across kills so the numbers reflect sustained
// play, not burst. The same scenario and options always give the same result.
func Simulate(scenario Scenario, opts SimOptions) SimResult {
	iterations := max(scenario.Iterations, 1)
	duration := opts.DurationSec
	if duration <= 0 {
		duration = defaultSimDurationSec
	}
	ticks := int(math.Round(duration / tickDuration))

	totals := &simTotals{dmgByTag: make(map[string]int64)}
	runs := make([]iterStats, iterations)

	for i := range iterations {
		seed := iterSeed(scenario.Name, opts.Seed, i)
		engine := opts.Engine
		if engine == nil {
			engine = NewDefaultCombatEngine(seed)
		}
		var trace io.Writer
		if i == opts.TraceIteration {
			trace = opts.Trace
		}
		run := newSimRun(scenario, engine, rand.New(rand.NewPCG(seed, seed^0xcafebabe)), totals, trace)
		for range ticks {
			if !run.step() {
				break
			}
		}
		run.stats.timeAlive = run.time
		runs[i] = run.stats
	}

	return aggregate(scenario, opts.Seed, duration, runs, totals)
}
