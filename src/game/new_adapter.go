package game

import (
	gaudio "dungeoneer/audio"
	"dungeoneer/combat"
	"dungeoneer/entities"
	"dungeoneer/pathing"
	"math"
	"sort"
)

// NewCombatAdapter routes to the src/combat engine.
type NewCombatAdapter struct {
	engine *combat.DefaultCombatEngine
	// combatState persists across ticks so timers and streak carry forward.
	combatState    combat.CombatState
	pendingActions []combat.Action // skill activations queued by HandleSkillActivation

	burnParticleTimer float64 // seconds until next burn particle burst on target
}

// getEngine returns the combat engine, initialising it lazily so that
// &NewCombatAdapter{} (zero-value) is safe to use without a constructor.
func (a *NewCombatAdapter) getEngine() *combat.DefaultCombatEngine {
	if a.engine == nil {
		a.engine = combat.NewDefaultCombatEngine(0)
	}
	return a.engine
}

func (a *NewCombatAdapter) HandleTargetSelect(g *Game, worldX, worldY float64) {
	// Find the monster closest to (worldX, worldY) within a click radius.
	// worldX/worldY are in cartesian tile-unit space (same space as BodyX/BodyY).
	// Click radius: 0.6 tiles.
	const clickRadiusSq = 0.6 * 0.6
	var best *entities.Monster
	bestDist := clickRadiusSq
	for _, m := range g.Monsters {
		if m == nil || m.IsDead {
			continue
		}
		dx := m.BodyX() - worldX
		dy := m.BodyY() - worldY
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			best = m
		}
	}
	g.TargetedMonster = best
	if best != nil {
		g.IsAutoAttacking = true
	}
}

func (a *NewCombatAdapter) HandleTargetNearest(g *Game) {
	// Select the nearest non-dead monster by cartesian distance from player.
	// Player origin is MoveController.InterpX/InterpY.
	if g.player == nil {
		return
	}
	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY

	var best *entities.Monster
	bestDistSq := -1.0
	for _, m := range g.Monsters {
		if m == nil || m.IsDead {
			continue
		}
		dx := m.InterpX - px
		dy := m.InterpY - py
		distSq := dx*dx + dy*dy
		if bestDistSq < 0 || distSq < bestDistSq {
			bestDistSq = distSq
			best = m
		}
	}
	g.TargetedMonster = best
}

func (a *NewCombatAdapter) HandleMoveToAttack(g *Game) {
	// If no target, fall back to interact (E key behaviour).
	if g.TargetedMonster == nil || g.TargetedMonster.IsDead {
		// No live target — no-op; caller can wire interact logic here in next phase.
		return
	}
	// Target is set: begin auto-attacking.
	g.IsAutoAttacking = true
}

func (a *NewCombatAdapter) HandleSkillActivation(g *Game, slotIdx int) {
	if g.player == nil {
		return
	}
	settings := LoadDevSettings()
	if settings.UseLegacyCombat {
		// Legacy path: spell visuals + damage handled entirely by castSpellSlot.
		g.castSpellSlot(slotIdx)
		return
	}
	// New engine path: queue the action; ProcessTick will pass it to the engine.
	// Visuals are triggered in ProcessTick via EventSkillFired events.
	a.pendingActions = append(a.pendingActions, combat.Action{
		Type:    combat.ActionActivateSkill,
		SlotIdx: slotIdx,
	})
}

func (a *NewCombatAdapter) ProcessTick(g *Game, dt float64) {
	// Sync last tick's engine conditions to the player BEFORE monster.Update /
	// updateMonsterProjectiles run later in this frame. This enforces shadow
	// invulnerability, taunt DR, and the stone_skin_idol damage cap on all
	// incoming monster attacks for the current frame.
	if g.player != nil {
		cs := a.combatState
		g.player.IsInvulnerable = cs.InShadow
		g.player.DamageReductionPct = cs.DamageReductionPct
		damageCap := 0
		for _, id := range cs.EquippedArtifacts {
			if eff, ok := combat.ArtifactEffects[id]; ok && eff.IsPassive && eff.DamageCapPct > 0 {
				capAmt := int(float64(g.player.MaxHP) * float64(eff.DamageCapPct) / 100.0)
				if damageCap == 0 || capAmt < damageCap {
					damageCap = capAmt
				}
			}
		}
		g.player.IncomingDamageCap = damageCap
	}

	// 1. Clear stale target.
	if g.TargetedMonster == nil || g.TargetedMonster.IsDead {
		g.TargetedMonster = nil
		g.IsAutoAttacking = false
		// Clear root on all monsters when there is no target.
		for _, mon := range g.Monsters {
			if mon != nil {
				mon.IsRooted = false
			}
		}
		// Reset persisted combat state targeting fields.
		a.combatState.HasTarget = false
		a.combatState.IsAutoAttacking = false
		a.combatState.TargetIsDead = false
		return
	}
	m := g.TargetedMonster

	// 2. Determine attack range and interval for this class.
	attackRange := 1.5  // knight default (tile units)
	attackInterval := 0.8
	if g.player.Class == entities.ClassMage {
		attackRange = 6.0
		attackInterval = 1.0
	}

	// 3. Compute distance from player to monster in tile units.
	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY
	mx := m.BodyX()
	my := m.BodyY()
	dist := math.Sqrt((mx-px)*(mx-px) + (my-py)*(my-py))
	inRange := dist <= attackRange

	// 4. Move toward target if out of range and auto-attacking.
	if g.IsAutoAttacking && !inRange {
		path := pathing.AStar(g.currentLevel, g.player.TileX, g.player.TileY, m.TileX, m.TileY)
		if len(path) > 0 {
			stopIdx := len(path) - 1
			if g.player.Class == entities.ClassKnight && len(path) > 1 {
				// Melee: stop one tile short to avoid standing on the monster.
				stopIdx = len(path) - 2
			} else if g.player.Class != entities.ClassKnight {
				// Ranged: walk to the first path node within attack range.
				// Pass the full sub-path (not just destination) so the movement
				// controller walks tile-by-tile instead of jumping across multiple
				// tiles in one interpolation step (the zip bug).
				for i, node := range path {
					dx := float64(node.X) - float64(m.TileX)
					dy := float64(node.Y) - float64(m.TileY)
					if math.Sqrt(dx*dx+dy*dy) <= attackRange-0.5 {
						stopIdx = i
						break
					}
				}
			}
			g.player.MoveController.SetPath(path[:stopIdx+1])
		}
		return // don't auto-attack while moving into range
	}

	// 5. Not auto-attacking or not in range — nothing to do.
	if !g.IsAutoAttacking || !inRange {
		return
	}

	// 6. Build the CombatState for this tick and pass to the engine.
	// Capture status that was active at the START of this tick (before the engine runs)
	// so we can retroactively block enemy hits that slipped through CombatCheck.
	// (CombatCheck runs before ProcessTick in the game loop, so a 1-tick lag is unavoidable
	// unless we snapshot here and restore below.)
	wasInShadow := a.combatState.InShadow
	wasTargetRooted := a.combatState.TargetRooted
	preTickHP := g.player.HP // HP after CombatCheck, before engine-computed costs

	eff := g.player.EffectiveStats()
	a.combatState.PlayerHP = g.player.HP
	a.combatState.PlayerMaxHP = g.player.MaxHP
	a.combatState.PlayerDamage = g.player.Damage
	a.combatState.PlayerIntelligence = eff.Intelligence
	a.combatState.PlayerStrength = eff.Strength
	a.combatState.PlayerAttackInterval = attackInterval
	a.combatState.PlayerAttackRange = attackRange
	a.combatState.PlayerClass = string(g.player.Class)
	a.combatState.PlayerX = px
	a.combatState.PlayerY = py
	a.combatState.HasTarget = true
	a.combatState.TargetHP = m.HP
	a.combatState.TargetMaxHP = m.MaxHP
	a.combatState.TargetX = mx
	a.combatState.TargetY = my
	a.combatState.TargetInRange = true
	a.combatState.IsAutoAttacking = true
	a.combatState.KillStreak = g.KillStreak
	a.combatState.DeltaTime = dt

	// Populate equipped artifacts and aggregate passive stat modifiers.
	for i := range a.combatState.EquippedArtifacts {
		a.combatState.EquippedArtifacts[i] = ""
	}
	for i, id := range g.player.SpellSlots {
		if i < 6 {
			a.combatState.EquippedArtifacts[i] = id
		}
	}
	if g.Meta != nil {
		a.combatState.EquippedArtifacts[6] = g.Meta.ArtifactLoadout[6]
	}
	// Re-aggregate passive effects from equipped artifacts each tick so CDR,
	// attack speed, and skill duration always reflect the current loadout.
	a.combatState.CooldownReductionPct = 0
	a.combatState.AttackSpeedPct = 0
	a.combatState.SkillDurationPct = 0
	for _, id := range a.combatState.EquippedArtifacts {
		if eff, ok := combat.ArtifactEffects[id]; ok && eff.IsPassive {
			a.combatState.CooldownReductionPct += eff.CooldownReductionPct
			a.combatState.AttackSpeedPct += eff.AttackSpeedPct
			a.combatState.SkillDurationPct += eff.SkillDurationPct
		}
	}

	// Drain the pending skill actions queued by HandleSkillActivation.
	actions := a.pendingActions
	a.pendingActions = a.pendingActions[:0]

	newState, events := a.getEngine().Tick(a.combatState, actions)
	a.combatState = newState

	// Root: propagate engine root state to the targeted monster so its movement
	// controller is frozen for the remainder of this frame and subsequent frames.
	if m != nil && !m.IsDead {
		m.IsRooted = a.combatState.TargetRooted
	}

	// Burn DoT particles: emit orange flame on the target while BurnActive so the
	// player gets continuous visual feedback that the DoT is ticking (casino signal).
	if a.combatState.BurnActive && !a.combatState.TargetIsDead && m != nil && !m.IsDead {
		a.burnParticleTimer -= dt
		if a.burnParticleTimer <= 0 {
			a.burnParticleTimer = 0.18
			if g.Particles != nil {
				isoX, isoY := g.cartesianToIso(mx, my)
				sx := (isoX-g.camX)*g.camScale + float64(g.w/2)
				sy := (isoY+g.camY)*g.camScale + float64(g.h/2)
				g.Particles.Emit(sx, sy, 5, 1.0, 0.38, 0.04) // orange-red flame
			}
		}
	} else {
		a.burnParticleTimer = 0
	}

	// mx/my (target body position) were captured above for distance computation;
	// they're reused here for AoE/chain radius checks in the event loop and remain
	// valid even if the target dies mid-loop.

	// 7. Consume events — drive game-layer effects.
	totalHPSpent := 0 // sum of EventHPSpent values this tick (blood_price, etc.)
	for _, ev := range events {
		switch ev.Type {
		case combat.EventDamageDealt:
			killed := m != nil && !m.IsDead && m.TakeDamage(ev.Value, &g.HitMarkers, &g.DamageNumbers)
			if killed {
				// Heal-on-kill: engine tracks it internally but game-layer HP is authoritative.
				// Apply any heal the engine computed before we clear the target.
				if a.combatState.PlayerHP > g.player.HP {
					healAmt := a.combatState.PlayerHP - g.player.HP
					if healAmt > 0 {
						g.player.Heal(healAmt)
					}
				}
				g.handleMonsterDeath(m)
				g.TargetedMonster = nil
				g.IsAutoAttacking = false
				a.combatState.HasTarget = false
				a.combatState.IsAutoAttacking = false
				a.combatState.TargetIsDead = true
			}
			// AoE + chain splash: apply the same damage value to nearby monsters.
			if ev.Tag != "" && ev.Value > 0 {
				if eff, ok := combat.ArtifactEffects[ev.Tag]; ok {
					// Area-of-effect: all monsters within radius.
					aoeRadius := eff.AOERadius
					if aoeRadius == 0 && eff.IsAOEField {
						aoeRadius = 2.5
					}
					if aoeRadius > 0 {
						for _, other := range g.Monsters {
							if other == nil || other.IsDead || other == m {
								continue
							}
							dx := other.BodyX() - mx
							dy := other.BodyY() - my
							if math.Sqrt(dx*dx+dy*dy) <= aoeRadius {
								if other.TakeDamage(ev.Value, &g.HitMarkers, &g.DamageNumbers) {
									g.handleMonsterDeath(other)
								}
							}
						}
					}
					// Chain lightning: jump to nearest N monsters within chain range.
					if eff.IsChain && eff.ChainCount > 0 {
						type distMon struct {
							dist float64
							mon  *entities.Monster
						}
						var candidates []distMon
						for _, other := range g.Monsters {
							if other == nil || other.IsDead || other == m {
								continue
							}
							dx := other.BodyX() - mx
							dy := other.BodyY() - my
							d := math.Sqrt(dx*dx + dy*dy)
							if d <= 4.0 {
								candidates = append(candidates, distMon{d, other})
							}
						}
						sort.Slice(candidates, func(i, j int) bool {
							return candidates[i].dist < candidates[j].dist
						})
						n := eff.ChainCount
						if n > len(candidates) {
							n = len(candidates)
						}
						for i := 0; i < n; i++ {
							if candidates[i].mon.TakeDamage(ev.Value, &g.HitMarkers, &g.DamageNumbers) {
								g.handleMonsterDeath(candidates[i].mon)
							}
						}
					}
				}
			}

		case combat.EventSkillFired:
			// Activation flash on the matching HUD slot + larger particle burst.
			if g.HUD != nil {
				for slotIdx, id := range a.combatState.EquippedArtifacts {
					if id == ev.Tag && slotIdx < 6 {
						g.HUD.SkillSlots[slotIdx].FlashTimer = 0.35
						if eff, ok := combat.ArtifactEffects[id]; ok {
							g.HUD.SkillSlots[slotIdx].MaxCooldown = eff.Cooldown
						}
						break
					}
				}
			}
			// Emit particles at target position using the artifact's domain color.
			if g.Particles != nil && g.currentLevel != nil && m != nil {
				isoX, isoY := g.cartesianToIso(mx, my)
				sx := (isoX-g.camX)*g.camScale + float64(g.w/2)
				sy := (isoY+g.camY)*g.camScale + float64(g.h/2)
				domain := ev.Tag
				if eff, ok := combat.ArtifactEffects[ev.Tag]; ok {
					domain = eff.Domain
				}
				r, gf, b := SpellParticleColor(domain)
				// Activation burst: more particles than a regular auto-attack.
				g.Particles.Emit(sx, sy, 22, r, gf, b)
			}
			if g.Audio != nil {
				switch ev.Tag {
				case "fireball":
					g.Audio.PlaySFX(gaudio.SFXSpellFireball)
				case "lightning", "lightning_storm":
					g.Audio.PlaySFX(gaudio.SFXSpellLightning)
				case "chaos_ray":
					g.Audio.PlaySFX(gaudio.SFXSpellChaos)
				default:
					g.Audio.PlaySFX(gaudio.SFXSpellGeneric)
				}
			}
		case combat.EventStreakChange:
			g.KillStreak = ev.Value
			a.combatState.KillStreak = ev.Value
		case combat.EventHPSpent:
			// blood_price and hollow_sigil route HP costs to the game-layer player.
			// Don't use player.TakeDamage — that applies item DR which is wrong for
			// an explicit HP cost (not incoming damage).
			if ev.Value > 0 {
				g.player.HP -= ev.Value
				if g.player.HP < 1 {
					g.player.HP = 1
				}
				totalHPSpent += ev.Value
			}
		}
	}

	// Shadow/root invulnerability: CombatCheck runs before ProcessTick in the game
	// loop, so targeted-enemy attacks that fired while shadow or root was already
	// active this frame need to be reversed. Engine-computed HP costs (blood_price)
	// are deducted via totalHPSpent and excluded from the reversal.
	if wasInShadow || wasTargetRooted {
		expectedHP := preTickHP - totalHPSpent
		if g.player.HP < expectedHP {
			g.player.HP = expectedHP
		}
	}
}
