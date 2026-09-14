package game

import (
	gaudio "dungeoneer/audio"
	"dungeoneer/combat"
	"dungeoneer/coords"
	"dungeoneer/entities"
	"dungeoneer/items"
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

	// slotTrans maps spell-bar index (0-5) → EquippedArtifacts index for the
	// current tick, so keypresses land on the right engine slot when the
	// loadout order differs from the spell bar order.
	slotTrans [6]int

	burnParticleTimer float64 // seconds until next burn particle burst on target
}

// BarCooldown returns the engine cooldown remaining for a spell-bar slot, and
// whether that bar slot maps to an engine slot at all. Used by the HUD, since
// the engine owns cooldowns on the new path (the legacy Caster is never
// populated).
func (a *NewCombatAdapter) BarCooldown(barIdx int) (float64, bool) {
	if barIdx < 0 || barIdx >= len(a.slotTrans) {
		return 0, false
	}
	tr := a.slotTrans[barIdx]
	if tr < 0 || tr >= len(a.combatState.ArtifactCooldowns) {
		return 0, false
	}
	return a.combatState.ArtifactCooldowns[tr], true
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
	// Click radius: 1.3 tiles — the click lands on the ground plane while the
	// sprite is tall, so a tight radius made targeting feel pixel-precise.
	// Nearest-within wins, so adjacent monsters still resolve correctly.
	const clickRadiusSq = 1.3 * 1.3
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
	// arcane_spray is a held channel (hold key → cone drains mana and ticks
	// damage every frame via updateChanneledSpray, which runs unconditionally
	// regardless of combat adapter). That doesn't fit the engine's discrete
	// one-press activation model, so it stays on the legacy channel path on
	// both combat paths rather than firing as a single burst.
	if slotIdx >= 0 && slotIdx < len(g.player.SpellSlots) && g.player.SpellSlots[slotIdx] == "arcane_spray" {
		g.castSpellSlot(slotIdx)
		return
	}
	// Pure passives (Cooldown 0, IsPassive true — e.g. stone_skin_idol,
	// soul_harvest) occupy a bar slot for tooltip/UI purposes but have no
	// active effect to trigger. Pressing the key should be a no-op, not a
	// wasted particle burst.
	if slotIdx >= 0 && slotIdx < len(g.player.SpellSlots) {
		itemID := artifactIDForAbility(g.player.SpellSlots[slotIdx])
		if eff, ok := combat.ArtifactEffects[itemID]; ok && eff.IsPassive && eff.Cooldown == 0 {
			return
		}
	}
	// Casting with no target auto-acquires the monster nearest the cursor so
	// spells thrown into a pack still connect (within a generous 4-tile radius).
	if g.TargetedMonster == nil || g.TargetedMonster.IsDead {
		const acquireRadiusSq = 4.0 * 4.0
		cx := float64(g.hoverTileX) + coords.BodyDX
		cy := float64(g.hoverTileY) + coords.BodyDY
		var best *entities.Monster
		bestDist := acquireRadiusSq
		for _, m := range g.Monsters {
			if m == nil || m.IsDead {
				continue
			}
			dx := m.BodyX() - cx
			dy := m.BodyY() - cy
			d := dx*dx + dy*dy
			if d < bestDist {
				bestDist = d
				best = m
			}
		}
		if best != nil {
			g.TargetedMonster = best
		}
	}
	// New engine path: queue the action; ProcessTick will pass it to the engine.
	// Visuals are triggered in ProcessTick via EventSkillFired events. The cursor
	// position rides along so targetless casts (fireball at the ground) still
	// have a visual destination.
	a.pendingActions = append(a.pendingActions, combat.Action{
		Type:    combat.ActionActivateSkill,
		SlotIdx: slotIdx,
		TargetX: float64(g.hoverTileX),
		TargetY: float64(g.hoverTileY),
	})
}

// buildEquipped fills the engine's artifact state from what the player is
// actually running with:
//
//   - Active slots 0-5 mirror the spell bar (SpellSlots), which the game keeps
//     current across loadout application, mid-run pickups, and equips. Ability
//     IDs ("ironbreaker_slam") resolve to the item IDs ("ironbreaker_gauntlets")
//     the engine registry is keyed by, so slotTrans is the identity for
//     populated bar slots.
//   - Slot 6 is the elite from the MetaSave loadout.
//   - PassiveArtifacts carries worn equipment with passive engine effects
//     (ember_mantle burn, quicksilver attack speed, stone_skin cap, …) that
//     never occupy a bar slot.
func (a *NewCombatAdapter) buildEquipped(g *Game) {
	for i := range a.combatState.EquippedArtifacts {
		a.combatState.EquippedArtifacts[i] = ""
	}
	for i := range a.slotTrans {
		a.slotTrans[i] = -1
	}

	for si, ability := range g.player.SpellSlots {
		if si >= 6 {
			break
		}
		a.combatState.EquippedArtifacts[si] = artifactIDForAbility(ability)
		a.slotTrans[si] = si
	}
	if g.Meta != nil {
		a.combatState.EquippedArtifacts[6] = canonicalArtifactID(g.Meta.ArtifactLoadout[6])
	}

	a.combatState.PassiveArtifacts = a.combatState.PassiveArtifacts[:0]
	for _, it := range g.player.Equipment {
		if it == nil || it.ItemTemplate == nil {
			continue
		}
		id := canonicalArtifactID(it.ID)
		if _, ok := combat.ArtifactEffects[id]; !ok {
			continue
		}
		already := false
		for _, eq := range a.combatState.EquippedArtifacts {
			if eq == id {
				already = true
				break
			}
		}
		if !already {
			a.combatState.PassiveArtifacts = append(a.combatState.PassiveArtifacts, id)
		}
	}
}

func (a *NewCombatAdapter) ProcessTick(g *Game, dt float64) {
	if g.player == nil {
		return
	}

	// Sync last tick's engine conditions to the player BEFORE monster.Update /
	// updateMonsterProjectiles run later in this frame. This enforces shadow
	// invulnerability, taunt DR, and the stone_skin_idol damage cap on all
	// incoming monster attacks for the current frame.
	{
		cs := &a.combatState
		g.player.IsInvulnerable = cs.InShadow
		g.player.DamageReductionPct = cs.DamageReductionPct
		damageCap := 0
		applyCap := func(id string) {
			if eff, ok := combat.ArtifactEffects[id]; ok && eff.IsPassive && eff.DamageCapPct > 0 {
				capAmt := int(float64(g.player.MaxHP) * float64(eff.DamageCapPct) / 100.0)
				if damageCap == 0 || capAmt < damageCap {
					damageCap = capAmt
				}
			}
		}
		for _, id := range cs.EquippedArtifacts {
			applyCap(id)
		}
		for _, id := range cs.PassiveArtifacts {
			applyCap(id)
		}
		g.player.IncomingDamageCap = damageCap
	}

	// 1. Target bookkeeping. The engine ticks every frame — with or without a
	// target — so cooldowns, shadow/root/burn timers, and queued skill presses
	// keep working between fights (they used to freeze out of combat).
	if g.TargetedMonster == nil || g.TargetedMonster.IsDead {
		g.TargetedMonster = nil
		g.IsAutoAttacking = false
		// Clear root on all monsters when there is no target.
		for _, mon := range g.Monsters {
			if mon != nil {
				mon.IsRooted = false
			}
		}
	}
	m := g.TargetedMonster

	// 2. Determine attack range and interval for this class.
	attackRange := 1.5 // knight default (tile units)
	attackInterval := 0.8
	if g.player.Class == entities.ClassMage {
		attackRange = 6.0
		attackInterval = 1.0
	}

	// 3. Compute distance from player to monster in tile units.
	px := g.player.MoveController.InterpX
	py := g.player.MoveController.InterpY
	inRange := false
	mx, my := 0.0, 0.0
	if m != nil {
		mx = m.BodyX()
		my = m.BodyY()
		dist := math.Sqrt((mx-px)*(mx-px) + (my-py)*(my-py))
		inRange = dist <= attackRange
	}

	// 4. Move toward target if out of range and auto-attacking.
	movingIntoRange := false
	if m != nil && g.IsAutoAttacking && !inRange {
		movingIntoRange = true
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
	}

	// 5. Build the CombatState for this tick and pass to the engine.
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
	a.combatState.PlayerMana = g.player.Mana
	a.combatState.PlayerMaxMana = g.player.MaxMana
	a.combatState.PlayerDamage = g.player.Damage
	a.combatState.PlayerIntelligence = eff.Intelligence
	a.combatState.PlayerStrength = eff.Strength
	a.combatState.PlayerAttackInterval = attackInterval
	a.combatState.PlayerAttackRange = attackRange
	a.combatState.PlayerClass = string(g.player.Class)
	a.combatState.PlayerX = px
	a.combatState.PlayerY = py
	a.combatState.KillStreak = g.KillStreak
	a.combatState.DeltaTime = dt
	if m != nil {
		a.combatState.HasTarget = true
		a.combatState.TargetHP = m.HP
		a.combatState.TargetMaxHP = m.MaxHP
		a.combatState.TargetX = mx
		a.combatState.TargetY = my
		a.combatState.TargetInRange = inRange
		a.combatState.TargetIsDead = false
		a.combatState.IsAutoAttacking = g.IsAutoAttacking && !movingIntoRange
	} else {
		a.combatState.HasTarget = false
		a.combatState.TargetInRange = false
		a.combatState.TargetIsDead = false
		a.combatState.IsAutoAttacking = false
	}

	// Populate equipped artifacts (canonical item IDs) and the spell-bar →
	// engine-slot translation.
	a.buildEquipped(g)

	// Re-aggregate passive effects from equipped artifacts each tick so CDR,
	// attack speed, and skill duration always reflect the current loadout.
	a.combatState.CooldownReductionPct = 0
	a.combatState.AttackSpeedPct = 0
	a.combatState.SkillDurationPct = 0
	aggregatePassive := func(id string) {
		if eff, ok := combat.ArtifactEffects[id]; ok && eff.IsPassive {
			a.combatState.CooldownReductionPct += eff.CooldownReductionPct
			a.combatState.AttackSpeedPct += eff.AttackSpeedPct
			a.combatState.SkillDurationPct += eff.SkillDurationPct
		}
	}
	for _, id := range a.combatState.EquippedArtifacts {
		aggregatePassive(id)
	}
	for _, id := range a.combatState.PassiveArtifacts {
		aggregatePassive(id)
	}
	// Mana cost reduction comes from equipment passives; the dev InfMana flag
	// makes all casts free without touching the engine's checks.
	a.combatState.ManaCostReductionPct = int(items.EvalPassiveManaCostReduction(g.player.Equipment) * 100)
	if g.InfMana {
		a.combatState.ManaCostReductionPct = 100
	}

	// Drain the pending skill actions queued by HandleSkillActivation,
	// translating spell-bar indices to engine slots.
	actions := a.pendingActions
	a.pendingActions = a.pendingActions[:0]
	for i := range actions {
		if actions[i].Type == combat.ActionActivateSkill {
			si := actions[i].SlotIdx
			if si >= 0 && si < len(a.slotTrans) && a.slotTrans[si] >= 0 {
				actions[i].SlotIdx = a.slotTrans[si]
			}
		}
	}

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

	// 6. Consume events — drive game-layer effects.
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
			if ev.Tag != "" && ev.Value > 0 && m != nil {
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
			a.handleSkillFired(g, ev, m, mx, my)
		case combat.EventStreakChange:
			g.KillStreak = ev.Value
			a.combatState.KillStreak = ev.Value
		case combat.EventManaChanged:
			// Engine deducted mana; game-layer mana is authoritative, so mirror
			// the spend here.
			if ev.Value > 0 && !g.InfMana {
				g.player.Mana -= ev.Value
				if g.player.Mana < 0 {
					g.player.Mana = 0
				}
			}
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

// handleSkillFired drives all game-layer feedback for a skill activation:
// HUD slot flash + cooldown, blink-strike teleport, the legacy spell visual
// (projectiles, arcs, fields — in visual-only mode since the engine owns
// damage), sound, and a domain-colored particle burst as the fallback for
// skills with no bespoke visual.
func (a *NewCombatAdapter) handleSkillFired(g *Game, ev combat.Event, m *entities.Monster, mx, my float64) {
	effect, hasEffect := combat.ArtifactEffects[ev.Tag]

	// Activation flash on the matching HUD slot. slotTrans maps bar → engine
	// slot, so invert it to find which bar slot fired.
	if g.HUD != nil {
		engineSlot := -1
		for i, id := range a.combatState.EquippedArtifacts {
			if id == ev.Tag {
				engineSlot = i
				break
			}
		}
		for barIdx, tr := range a.slotTrans {
			if tr == engineSlot && barIdx < len(g.HUD.SkillSlots) {
				g.HUD.SkillSlots[barIdx].FlashTimer = 0.35
				if hasEffect {
					g.HUD.SkillSlots[barIdx].MaxCooldown = effect.Cooldown
				}
				break
			}
		}
	}

	// groundX/Y is the raw hovered tile the cast action carried — the cursor-
	// aim convention every legacy ground-cast spell (fireball, lightning, ...)
	// used, with no target-lock offset. lockX/Y is the current target's body
	// center (or the ground position if untargeted) — for artifact skills that
	// act on the locked target rather than the cursor.
	groundX, groundY := ev.X, ev.Y
	lockX, lockY := groundX+coords.BodyDX, groundY+coords.BodyDY
	if m != nil {
		lockX, lockY = mx, my
	}

	// Legacy spell visual in visual-only mode; artifact skills get their own
	// bespoke effect (shadowstep teleports as part of its own visual); fall
	// back to a particle burst for anything with no bespoke visual yet.
	spellID := ev.Tag
	if hasEffect {
		// Artifact item IDs (e.g. emblem items granting "fireball", or new
		// artifacts like "ironbreaker_gauntlets") resolve to the ability they
		// grant so the visual switch matches.
		if tmpl, ok := items.Registry[ev.Tag]; ok && tmpl.GrantsAbility != "" {
			spellID = tmpl.GrantsAbility
		}
	}
	if !g.spawnSkillVisual(spellID, groundX, groundY, lockX, lockY, a.combatState.TargetIsDead) {
		if g.Particles != nil && g.currentLevel != nil {
			tx, ty := lockX, lockY
			isoX, isoY := g.cartesianToIso(tx, ty)
			sx := (isoX-g.camX)*g.camScale + float64(g.w/2)
			sy := (isoY+g.camY)*g.camScale + float64(g.h/2)
			domain := ev.Tag
			if hasEffect {
				domain = effect.Domain
			}
			r, gf, b := SpellParticleColor(domain)
			// Activation burst: more particles than a regular auto-attack.
			g.Particles.Emit(sx, sy, 22, r, gf, b)
		}
	}

	if g.Audio != nil {
		switch spellID {
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
}
