package combat

import "math/rand/v2"

// CombatEngine processes one tick of combat logic.
// Inputs: current state + player actions for this tick.
// Outputs: updated state + events for the game layer to consume.
type CombatEngine interface {
	Tick(state CombatState, actions []Action) (CombatState, []Event)
}

// Event is emitted by the engine and consumed by the game layer to trigger
// visual/audio effects (damage numbers, sounds, particles, screen shake).
type Event struct {
	Type   EventType
	Value  int
	X, Y   float64
	IsCrit bool
	Tag    string // artifact ID, effect name, etc.
}

// EventType classifies what happened so the game layer knows how to react.
type EventType string

const (
	EventDamageDealt  EventType = "damage_dealt"
	EventDamageTaken  EventType = "damage_taken"
	EventTargetDied   EventType = "target_died"
	EventAutoAttack   EventType = "auto_attack"
	EventSkillFired   EventType = "skill_fired"
	EventStreakChange EventType = "streak_changed"
	EventKillMomentum EventType = "kill_momentum"
	EventHPSpent      EventType = "hp_spent"  // blood_price, hollow_sigil; Value = HP deducted
	EventLevelUp      EventType = "level_up"  // Value = new level; Tag = stat points e.g. "str str vit"
)

// DefaultCombatEngine is the concrete implementation of CombatEngine.
// It is pure Go — no Ebiten or game-layer dependencies.
type DefaultCombatEngine struct {
	rng *rand.Rand
}

// NewDefaultCombatEngine creates a DefaultCombatEngine with the given seed.
// Pass 0 to use a fixed default seed for reproducible simulations.
func NewDefaultCombatEngine(seed uint64) *DefaultCombatEngine {
	if seed == 0 {
		seed = 12345
	}
	return &DefaultCombatEngine{
		rng: rand.New(rand.NewPCG(seed, seed^0xdeadbeef)),
	}
}

const (
	baseCritChance     = 0.05  // 5% base crit chance (Luck not yet wired in)
	critMultiplier     = 1.5   // 1.5× damage on crit
	skillStubCooldown  = 3.0   // fallback cooldown for skills not in ArtifactEffects (seconds)
	streakResetTime    = 5.0   // seconds before kill streak decays
	momentumPctPerKill = 0.05  // 5% attack interval reduction per streak level
	maxMomentumPct     = 0.25  // cap at 25% total reduction
)

// applyKillPassives checks equipped artifacts for on-kill passives (e.g. soul_harvest)
// and applies their effects. Call this immediately after any kill event is recorded.
// Multiple heal-on-kill artifacts each contribute — there is no break.
func applyKillPassives(state *CombatState, events *[]Event) {
	totalHeal := 0
	for _, id := range state.EquippedArtifacts {
		eff, ok := ArtifactEffects[id]
		if !ok || !eff.IsPassive {
			continue
		}
		if eff.HealOnKillPct > 0 {
			healAmt := int(float64(state.PlayerMaxHP) * float64(eff.HealOnKillPct) / 100.0)
			if state.EnemyHealReductionPct > 0 {
				healAmt = int(float64(healAmt) * (1.0 - float64(state.EnemyHealReductionPct)/100.0))
				if healAmt < 0 {
					healAmt = 0
				}
			}
			totalHeal += healAmt
		}
		if eff.HealOnKillFlat > 0 {
			healAmt := eff.HealOnKillFlat
			if state.EnemyHealReductionPct > 0 {
				healAmt = int(float64(healAmt) * (1.0 - float64(state.EnemyHealReductionPct)/100.0))
				if healAmt < 0 {
					healAmt = 0
				}
			}
			totalHeal += healAmt
		}
	}
	if totalHeal > 0 {
		state.PlayerHP += totalHeal
		if state.PlayerHP > state.PlayerMaxHP {
			state.PlayerHP = state.PlayerMaxHP
		}
		*events = append(*events, Event{Type: EventDamageTaken, Value: -totalHeal}) // negative = heal
	}
}

// recordKill marks the target as dead, increments the kill streak, and emits
// the standard kill events. It also fires applyKillPassives.
func recordKill(state *CombatState, events *[]Event) {
	state.TargetHP = 0
	state.TargetIsDead = true
	state.IsAutoAttacking = false
	state.KillStreak++
	state.StreakTimer = streakResetTime
	*events = append(*events, Event{Type: EventTargetDied, Value: state.KillStreak})
	*events = append(*events, Event{Type: EventStreakChange, Value: state.KillStreak})
	applyKillPassives(state, events)
}

// passiveDmgBonusMult returns the cumulative passive damage multiplier from all
// equipped artifacts. Call once per damage event to apply marrow_ring, blood_vow_amulet, etc.
func passiveDmgBonusMult(state *CombatState) float64 {
	bonus := 0
	for _, id := range state.EquippedArtifacts {
		if eff, ok := ArtifactEffects[id]; ok && eff.IsPassive {
			bonus += eff.DamageBonusPct
			if eff.DamageBonusLowHPPct > 0 && state.PlayerHP*2 < state.PlayerMaxHP {
				bonus += eff.DamageBonusLowHPPct
			}
			if eff.DoTDmgBonusPct > 0 && state.ActiveDoTCount > 0 {
				bonus += eff.DoTDmgBonusPct * state.ActiveDoTCount
			}
		}
	}
	return 1.0 + float64(bonus)/100.0
}

// Tick advances combat by one step (state.DeltaTime seconds).
func (e *DefaultCombatEngine) Tick(state CombatState, actions []Action) (CombatState, []Event) {
	var events []Event
	dt := state.DeltaTime

	// 1. Process actions.
	for _, a := range actions {
		switch a.Type {
		case ActionSelectTarget:
			state.HasTarget = true
			state.TargetX = a.TargetX
			state.TargetY = a.TargetY
			state.IsAutoAttacking = true

		case ActionClearTarget:
			state.HasTarget = false
			state.IsAutoAttacking = false
			state.TargetIsDead = false

		case ActionTargetNearest:
			// Game layer must have pre-populated TargetX/TargetY when it sends
			// this action. Here we simply accept the target.
			state.HasTarget = true
			state.IsAutoAttacking = true

		case ActionMoveToAttack:
			// Game layer handles pathing; engine acknowledges intent.
			state.IsAutoAttacking = true

		case ActionActivateSkill:
			idx := a.SlotIdx
			if idx < 0 || idx >= len(state.EquippedArtifacts) {
				break
			}
			artifactID := state.EquippedArtifacts[idx]
			if artifactID == "" {
				break
			}
			if state.ArtifactCooldowns[idx] > 0 {
				break // still on cooldown
			}
			// Silencer: block skill if player is within enemy silence radius.
			if state.EnemySilenceRadius > 0 {
				dx := state.PlayerX - state.TargetX
				dy := state.PlayerY - state.TargetY
				distSq := dx*dx + dy*dy
				r := state.EnemySilenceRadius
				if distSq <= r*r {
					break // silenced — skill does not fire
				}
			}
			// Gravity Warden: block blink-strike skills.
			if state.EnemyBlockBlink {
				if eff, ok := ArtifactEffects[artifactID]; ok && eff.IsBlinkStrike {
					break // blink blocked
				}
			}
			// Determine cooldown from registry; fall back to stub.
			cd := skillStubCooldown
			if effect, ok := ArtifactEffects[artifactID]; ok {
				cd = effect.Cooldown
			}
			state.ArtifactCooldowns[idx] = cd
			events = append(events, Event{
				Type: EventSkillFired,
				Tag:  artifactID,
			})

			// Apply skill effects.
			effect, hasEffect := ArtifactEffects[artifactID]
			if hasEffect {
				// Duration modifier from SkillDurationPct
				durationMult := 1.0 + float64(state.SkillDurationPct)/100.0

				if effect.IsBlinkStrike {
					// Enter shadow + guarantee next crit
					state.InShadow = true
					state.ShadowTimer = effect.DurationSec * durationMult
					state.NextCritGuaranteed = true
				}
				if effect.IsTaunt {
					state.DamageReductionPct = effect.DamageReductionPct
					state.TauntTimer = effect.DurationSec * durationMult
				}
				if effect.IsRoot {
					state.TargetRooted = true
					state.RootTimer = effect.DurationSec * durationMult
				}
				if effect.IsExecute {
					// Execute: instakill if target is below threshold
					threshold := int(float64(state.TargetMaxHP) * float64(effect.ExecuteThresholdPct) / 100.0)
					if state.TargetHP <= threshold && state.TargetHP > 0 {
						execDmg := state.TargetHP
						events = append(events, Event{Type: EventDamageDealt, Value: execDmg, X: state.TargetX, Y: state.TargetY})
						if !state.TargetIsDead {
							recordKill(&state, &events)
						}
					}
				}
				if effect.SurgeDmgPerCooldown > 0 {
					// Count how many artifact slots are currently on cooldown (excluding self)
					cdCount := 0
					for i, cd := range state.ArtifactCooldowns {
						if cd > 0 && state.EquippedArtifacts[i] != artifactID {
							cdCount++
						}
					}
					surgeDmg := effect.SurgeDmgPerCooldown * cdCount
					if surgeDmg > 0 {
						events = append(events, Event{Type: EventDamageDealt, Value: surgeDmg, X: state.TargetX, Y: state.TargetY})
						state.TargetHP -= surgeDmg
						if state.TargetHP <= 0 && !state.TargetIsDead {
							recordKill(&state, &events)
						}
					}
				}
				if effect.HPCostPct > 0 && effect.DamageFlat > 0 {
					// Blood price: spend % of current HP, deal flat damage
					hpCost := int(float64(state.PlayerHP) * float64(effect.HPCostPct) / 100.0)
					if hpCost < 1 {
						hpCost = 1
					}
					state.PlayerHP -= hpCost
					if state.PlayerHP < 1 {
						state.PlayerHP = 1 // don't let it kill you
					}
					events = append(events, Event{Type: EventHPSpent, Value: hpCost})
					dmg := effect.DamageFlat
					events = append(events, Event{Type: EventDamageDealt, Value: dmg, X: state.TargetX, Y: state.TargetY})
					state.TargetHP -= dmg
					if state.TargetHP <= 0 && !state.TargetIsDead {
						recordKill(&state, &events)
					}
				}
				// Spell damage: registered game spells with INT/STR scaling.
				// AoE fields use DurationSec as a time multiplier (total = DPS × secs).
				if effect.SpellDamageBase > 0 || effect.SpellDamagePerINT > 0 || effect.SpellDamagePerSTR > 0 {
					spellDmg := effect.SpellDamageBase
					spellDmg += int(float64(state.PlayerIntelligence) * effect.SpellDamagePerINT)
					spellDmg += int(float64(state.PlayerStrength) * effect.SpellDamagePerSTR)
					if effect.IsAOEField && effect.DurationSec > 0 {
						spellDmg = int(float64(spellDmg) * effect.DurationSec)
					}
					spellDmg = int(float64(spellDmg) * passiveDmgBonusMult(&state))
					finalDmg, isCrit := e.calcDamage(spellDmg, false)
					events = append(events, Event{
						Type:   EventDamageDealt,
						Value:  finalDmg,
						X:      state.TargetX,
						Y:      state.TargetY,
						IsCrit: isCrit,
						Tag:    artifactID,
					})
					state.TargetHP -= finalDmg
					if state.TargetHP <= 0 && !state.TargetIsDead {
						recordKill(&state, &events)
					}
				}
				// DamageMultiplier: melee/physical skills that strike for N× base weapon damage.
				// IsBlinkStrike sets NextCritGuaranteed before this runs, so the blink strike crits.
				if effect.DamageMultiplier > 0 {
					skillDmg := int(float64(state.PlayerDamage) * effect.DamageMultiplier * passiveDmgBonusMult(&state))
					finalDmg, isCrit := e.calcDamage(skillDmg, state.NextCritGuaranteed)
					state.NextCritGuaranteed = false
					events = append(events, Event{
						Type:   EventDamageDealt,
						Value:  finalDmg,
						X:      state.TargetX,
						Y:      state.TargetY,
						IsCrit: isCrit,
						Tag:    artifactID,
					})
					state.TargetHP -= finalDmg
					if state.TargetHP <= 0 && !state.TargetIsDead {
						recordKill(&state, &events)
					}
				}

				// hollow_sigil: void-domain skills cost HP instead of mana.
				// Skip skills that already pay an HP cost via HPCostPct (e.g. blood_price)
				// to avoid double-charging the same activation.
				if effect.Domain == "void" && effect.HPCostPct == 0 {
					for _, equippedID := range state.EquippedArtifacts {
						if eqEff, ok := ArtifactEffects[equippedID]; ok && eqEff.VoidCostsHP {
							const voidHPCost = 10
							hpCost := voidHPCost
							state.PlayerHP -= hpCost
							if state.PlayerHP < 1 {
								state.PlayerHP = 1
							}
							events = append(events, Event{Type: EventHPSpent, Value: hpCost, Tag: "hollow_sigil"})
							break // one hollow_sigil is enough; don't double-charge
						}
					}
				}
			}
		}
	}

	// 2. Tick down artifact cooldowns (CDR makes time pass faster).
	cdMultiplier := 1.0 + float64(state.CooldownReductionPct)/100.0
	for i := range state.ArtifactCooldowns {
		if state.ArtifactCooldowns[i] > 0 {
			state.ArtifactCooldowns[i] -= dt * cdMultiplier
			if state.ArtifactCooldowns[i] < 0 {
				state.ArtifactCooldowns[i] = 0
			}
		}
	}

	// 3. Auto-attack tick.
	if state.HasTarget && state.TargetInRange && state.IsAutoAttacking && !state.TargetIsDead {
		// Apply kill momentum: reduce attack interval up to maxMomentumPct.
		effectiveInterval := state.PlayerAttackInterval
		if state.KillStreak > 0 {
			reduction := float64(state.KillStreak) * momentumPctPerKill
			if reduction > maxMomentumPct {
				reduction = maxMomentumPct
			}
			effectiveInterval = state.PlayerAttackInterval * (1.0 - reduction)
		}

		// Apply AttackSpeedPct bonus.
		if state.AttackSpeedPct > 0 {
			effectiveInterval = effectiveInterval / (1.0 + float64(state.AttackSpeedPct)/100.0)
		}

		state.AutoAttackTimer -= dt
		if state.AutoAttackTimer <= 0 {
			// Compute damage (with passive bonuses from marrow_ring, blood_vow_amulet, etc.).
			dmgBonusMult := passiveDmgBonusMult(&state)
			baseDmg := int(float64(state.PlayerDamage) * dmgBonusMult)
			dmg, isCrit := e.calcDamage(baseDmg, state.NextCritGuaranteed)
			state.NextCritGuaranteed = false // consume the guaranteed crit

			// Deal damage.
			state.TargetHP -= dmg
			events = append(events, Event{
				Type: EventAutoAttack,
				X:    state.TargetX,
				Y:    state.TargetY,
			})
			events = append(events, Event{
				Type:   EventDamageDealt,
				Value:  dmg,
				X:      state.TargetX,
				Y:      state.TargetY,
				IsCrit: isCrit,
			})

			// Reset timer using the momentum-reduced interval.
			state.AutoAttackTimer = effectiveInterval

			// Check for passive burn from equipped artifacts (ember_mantle).
			for _, id := range state.EquippedArtifacts {
				if eff, ok := ArtifactEffects[id]; ok && eff.IsPassive && eff.BurnDPS > 0 {
					durationMult := 1.0 + float64(state.SkillDurationPct)/100.0
					if !state.BurnActive {
						state.ActiveDoTCount++ // new DoT stack started
					}
					state.BurnActive = true
					state.BurnDPS = eff.BurnDPS
					state.BurnTimer = eff.BurnDurationSec * durationMult
					break
				}
			}

			// Check for shadow hit → shroud_cloak cooldown reset (shadows_return).
			if state.InShadow {
				for i, id := range state.EquippedArtifacts {
					if eff, ok := ArtifactEffects[id]; ok && eff.ShroudCooldownReset > 0 {
						// Find shroud_cloak slot and reduce its cooldown.
						for j, sid := range state.EquippedArtifacts {
							if sid == "shroud_cloak" && state.ArtifactCooldowns[j] > 0 {
								state.ArtifactCooldowns[j] -= eff.ShroudCooldownReset
								if state.ArtifactCooldowns[j] < 0 {
									state.ArtifactCooldowns[j] = 0
								}
							}
						}
						_ = i
					}
				}
			}

			// Check kill from auto-attack.
			if state.TargetHP <= 0 && !state.TargetIsDead {
				recordKill(&state, &events)

				// Emit kill-momentum event when streak provides a bonus.
				if state.KillStreak > 0 {
					reduction := float64(state.KillStreak) * momentumPctPerKill
					if reduction > maxMomentumPct {
						reduction = maxMomentumPct
					}
					events = append(events, Event{
						Type:  EventKillMomentum,
						Value: int(reduction * 100), // e.g. 5, 10, 25 ...
					})
				}
			}
		}
	}

	// 4. Tick down streak timer; reset streak when it expires.
	if state.KillStreak > 0 {
		state.StreakTimer -= dt
		if state.StreakTimer <= 0 {
			state.KillStreak = 0
			state.StreakTimer = 0
			events = append(events, Event{
				Type:  EventStreakChange,
				Value: 0,
			})
		}
	}

	// 5. Tick shadow timer. Veilbane ejects player from shadow if within detection radius.
	if state.InShadow {
		if state.EnemyDetectionRadius > 0 {
			dx := state.PlayerX - state.TargetX
			dy := state.PlayerY - state.TargetY
			distSq := dx*dx + dy*dy
			r := state.EnemyDetectionRadius
			if distSq <= r*r {
				state.InShadow = false
				state.ShadowTimer = 0
			}
		}
		if state.InShadow {
			state.ShadowTimer -= dt
			if state.ShadowTimer <= 0 {
				state.InShadow = false
				state.ShadowTimer = 0
			}
		}
	}

	// 6. Tick root timer.
	if state.TargetRooted {
		state.RootTimer -= dt
		if state.RootTimer <= 0 {
			state.TargetRooted = false
			state.RootTimer = 0
		}
	}

	// 7. Tick taunt timer.
	if state.TauntTimer > 0 {
		state.TauntTimer -= dt
		if state.TauntTimer <= 0 {
			state.DamageReductionPct = 0
			state.TauntTimer = 0
		}
	}

	// 8. Passive HP drain (blood_vow_amulet: 5 HP/s).
	// We accumulate fractional drain per tick to avoid int(5 * 1/60) = 0 truncation.
	for _, id := range state.EquippedArtifacts {
		if eff, ok := ArtifactEffects[id]; ok && eff.IsPassive && eff.HPDrainPerSec > 0 {
			state.HPDrainAccum += float64(eff.HPDrainPerSec) * dt
			if state.HPDrainAccum >= 1.0 {
				drain := int(state.HPDrainAccum)
				state.HPDrainAccum -= float64(drain)
				state.PlayerHP -= drain
				if state.PlayerHP < 1 {
					state.PlayerHP = 1 // passive drain can't kill the player
				}
			}
		}
	}

	// 9. Tick burn DoT and emit damage.
	if state.BurnActive && state.BurnTimer > 0 && !state.TargetIsDead {
		state.BurnTimer -= dt
		burnThisTick := int(float64(state.BurnDPS) * dt)
		if burnThisTick > 0 {
			events = append(events, Event{Type: EventDamageDealt, Value: burnThisTick, X: state.TargetX, Y: state.TargetY})
			state.TargetHP -= burnThisTick
			if state.TargetHP <= 0 && !state.TargetIsDead {
				recordKill(&state, &events)
			}
		}
		if state.BurnTimer <= 0 {
			state.BurnActive = false
			state.BurnDPS = 0
			if state.ActiveDoTCount > 0 {
				state.ActiveDoTCount--
			}
		}
	}

	return state, events
}
