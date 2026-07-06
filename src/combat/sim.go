package combat

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

// Scenario defines one benchmarker test case.
type Scenario struct {
	Name          string    `json:"name"`
	Class         string    `json:"class"`          // "knight" | "mage"
	Artifacts     []string  `json:"artifacts"`      // artifact IDs, up to 7
	Stats         BaseStats `json:"stats"`
	Floor         int       `json:"floor"`
	Biome         string    `json:"biome"`
	EnemyPool     []string  `json:"enemy_pool"`     // enemy role names
	Iterations    int       `json:"iterations"`
	SkillRotation []string  `json:"skill_rotation"` // "auto" | "slot_N"
	EnemyBuildID  string   `json:"enemy_build"`    // optional: counter-meta enemy archetype ID
	// StatPriority defines which stat gets each of the 3 points awarded on a level-up.
	// Cycles: first entry gets point 1, second entry gets point 2, third gets point 3.
	// Valid values: "str" "vit" "int" "dex". Defaults to ["str","vit","int"] if empty.
	StatPriority []string `json:"stat_priority"`
}

// BaseStats mirrors entities.BaseStats but is defined here for engine independence.
type BaseStats struct {
	Strength     int `json:"Strength"`
	Dexterity    int `json:"Dexterity"`
	Vitality     int `json:"Vitality"`
	Intelligence int `json:"Intelligence"`
	Luck         int `json:"Luck"`
}

// LoadScenario reads a JSON scenario file from disk.
func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("combat: LoadScenario: %w", err)
	}
	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return Scenario{}, fmt.Errorf("combat: LoadScenario: %w", err)
	}
	if s.Iterations <= 0 {
		s.Iterations = 1
	}
	return s, nil
}

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

// deterministicRNG returns a seeded RNG derived from a scenario name.
func deterministicRNG(name string) *rand.Rand {
	var seed uint64
	for i, c := range name {
		seed ^= uint64(c) << (uint64(i) % 56)
	}
	if seed == 0 {
		seed = 0xdeadbeef
	}
	return rand.New(rand.NewPCG(seed, seed^0xcafebabe))
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
	var n int
	_, err := fmt.Sscanf(strings.TrimPrefix(entry, "slot_"), "%d", &n)
	if err != nil || n < 1 {
		return -1
	}
	return n - 1
}

const simDurationTicks = 7200 // 2 minutes at 60fps per iteration

// RunSimulation runs a Scenario N times and returns aggregate SimResult.
// Each iteration is a 2-minute continuous fight with enemy respawn — cooldowns
// persist across kills so the numbers reflect sustained DPS, not burst.
// Uses a seeded RNG so results are reproducible given the same scenario.
func RunSimulation(engine CombatEngine, scenario Scenario) SimResult {
	rng := deterministicRNG(scenario.Name)

	iterations := scenario.Iterations
	if iterations <= 0 {
		iterations = 1
	}

	var (
		totalDmgDealt float64
		totalDmgTaken float64
		totalKills    int
		survivedCount int
		totalStreak   float64
		skillCounts   = make(map[string]int)
		totalActions  int
	)

	rotLen := len(scenario.SkillRotation)

	for range iterations {
		state := buildState(scenario, rng)
		rotIdx := 0
		iterKills := 0
		iterDmgTaken := 0
		iterMaxStreak := 0
		enemyAtkTimer := enemyAttackInterval
		died := false

		for tick := range simDurationTicks {
			_ = tick
			// Build action from skill rotation.
			var actions []Action
			if rotLen > 0 {
				entry := scenario.SkillRotation[rotIdx%rotLen]
				rotIdx++
				if entry != "auto" {
					slot := parseSlotIdx(entry)
					if slot >= 0 {
						actions = append(actions, Action{
							Type:    ActionActivateSkill,
							SlotIdx: slot,
						})
						if slot < len(state.EquippedArtifacts) {
							id := state.EquippedArtifacts[slot]
							if id != "" {
								skillCounts[id]++
							} else {
								skillCounts[entry]++
							}
						}
						totalActions++
					}
				}
			}

			newState, events := engine.Tick(state, actions)
			state = newState

			for _, ev := range events {
				switch ev.Type {
				case EventDamageDealt:
					totalDmgDealt += float64(ev.Value)
				case EventTargetDied:
					iterKills++
					// Award EXP; level-up events are emitted into a local slice and consumed here.
					var lvlEvents []Event
					awardSimEXP(&state, scenario.Floor, statPriorityOrDefault(scenario), &lvlEvents)
					for _, lev := range lvlEvents {
						if lev.Type == EventLevelUp {
							// Level-up events counted but not accumulated in aggregate stats.
							_ = lev
						}
					}
					// Respawn next enemy; keep cooldowns so we measure sustained DPS.
					state.TargetHP = enemyHP(scenario.Floor)
					state.TargetMaxHP = state.TargetHP
					state.TargetIsDead = false
					state.HasTarget = true
					state.IsAutoAttacking = true
					state.TargetName = pickRandom(scenario.EnemyPool, rng, "enemy")
					enemyAtkTimer = enemyAttackInterval
				case EventStreakChange:
					if ev.Value > iterMaxStreak {
						iterMaxStreak = ev.Value
					}
				case EventHPSpent:
					if state.EnemySacrificeLeech && !state.TargetIsDead {
						state.TargetHP += ev.Value
						if state.TargetHP > state.TargetMaxHP {
							state.TargetHP = state.TargetMaxHP
						}
					}
				}
			}

			// Enemy attacks at intervals.
			enemyAtkTimer -= tickDuration
			if enemyAtkTimer <= 0 && state.HasTarget && !state.TargetIsDead {
				enemyAtkTimer = enemyAttackInterval
				dmg := enemyDmgPerAttack(scenario.Floor)

				if state.TargetRooted || state.InShadow {
					dmg = 0
				}

				if dmg > 0 {
					if state.DamageReductionPct > 0 {
						dmg = int(float64(dmg) * (1.0 - float64(state.DamageReductionPct)/100.0))
						if dmg < 0 {
							dmg = 0
						}
					}

					rawDmg := dmg
					for _, id := range state.EquippedArtifacts {
						if eff, ok := ArtifactEffects[id]; ok && eff.DamageCapPct > 0 {
							cap := int(float64(state.PlayerMaxHP) * float64(eff.DamageCapPct) / 100.0)
							if cap < 1 {
								cap = 1
							}
							if dmg > cap {
								dmg = cap
							}
							break
						}
					}

					if state.EnemyDamageCapBypass && state.PlayerHP*2 < state.PlayerMaxHP {
						dmg = rawDmg
					}

					if state.EnemyPackBonusPct > 0 {
						dmg += int(float64(dmg) * float64(state.EnemyPackBonusPct) / 100.0)
					}

					state.PlayerHP -= dmg
					iterDmgTaken += dmg

					if state.EnemyInstakillPct > 0 && state.PlayerHP > 0 {
						threshold := state.PlayerMaxHP * state.EnemyInstakillPct / 100
						if state.PlayerHP <= threshold {
							state.PlayerHP = 0
						}
					}

					if state.KillStreak > 0 {
						state.KillStreak = 0
						state.StreakTimer = 0
					}
				}
			}

			// Regenerator: enemy heals each tick.
			if state.EnemyHPRegenPerSec > 0 && !state.TargetIsDead {
				regen := int(float64(state.EnemyHPRegenPerSec) * tickDuration)
				if regen > 0 {
					state.TargetHP += regen
					if state.TargetHP > state.TargetMaxHP {
						state.TargetHP = state.TargetMaxHP
					}
				}
			}

			if state.PlayerHP <= 0 {
				died = true
				break
			}
		}

		totalKills += iterKills
		totalDmgDealt += 0 // already accumulated above
		totalDmgTaken += float64(iterDmgTaken)
		totalStreak += float64(iterMaxStreak)
		if !died {
			survivedCount++
		}
	}

	result := SimResult{
		ScenarioName:    scenario.Name,
		Iterations:      iterations,
		SkillUsageRatio: make(map[string]float64),
	}

	simDurSec := float64(simDurationTicks) * tickDuration // 120s
	avgKillsPerIter := float64(totalKills) / float64(iterations)
	result.KillsPerMinute = avgKillsPerIter / (simDurSec / 60.0)
	if avgKillsPerIter > 0 {
		result.AvgClearTimeSec = simDurSec / avgKillsPerIter
	}
	result.AvgDPSTick = totalDmgDealt / (float64(iterations) * float64(simDurationTicks)) * 60.0
	result.SurvivalRate = float64(survivedCount) / float64(iterations)
	result.AvgDamageTaken = totalDmgTaken / float64(iterations)
	result.StreakAvg = totalStreak / float64(iterations)

	if totalActions > 0 {
		for id, count := range skillCounts {
			result.SkillUsageRatio[id] = float64(count) / float64(totalActions)
		}
	}

	return result
}
