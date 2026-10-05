package game

import (
	"math"

	"dungeoneer/combat"
	"dungeoneer/coords"
	"dungeoneer/items"
	"dungeoneer/spells"
)

// abilityToArtifactID caches the ability-ID → artifact-item-ID resolution.
// SpellSlots store GrantsAbility values ("ironbreaker_slam"), while the combat
// engine's ArtifactEffects registry is keyed by item IDs ("ironbreaker_gauntlets").
// Built lazily once — the item registry is static after load.
var abilityToArtifactID map[string]string

// artifactIDForAbility resolves an ability ID to the artifact item that grants
// it. IDs with no granting artifact (legacy spells like "fireball", or direct
// item IDs written by dev tools) are returned unchanged.
func artifactIDForAbility(abilityID string) string {
	if abilityToArtifactID == nil {
		abilityToArtifactID = make(map[string]string, len(items.Registry))
		for _, tmpl := range items.Registry {
			if tmpl.IsArtifact && tmpl.GrantsAbility != "" {
				// Only map abilities the engine knows by item ID; spells like
				// "fireball" are registered in ArtifactEffects under the ability
				// ID itself and must not be remapped to their emblem item.
				if _, ok := combat.ArtifactEffects[tmpl.ID]; ok {
					abilityToArtifactID[tmpl.GrantsAbility] = tmpl.ID
				}
			}
		}
	}
	if itemID, ok := abilityToArtifactID[abilityID]; ok {
		return itemID
	}
	return abilityID
}

// canonicalArtifactID maps any loadout entry to the ID the engine registry is
// keyed by. New artifacts are registered under their item ID; legacy spell
// emblems (e.g. item_2_24 → "fireball") are registered under the ability they
// grant.
func canonicalArtifactID(id string) string {
	if _, ok := combat.ArtifactEffects[id]; ok {
		return id
	}
	if tmpl, ok := items.Registry[id]; ok && tmpl.GrantsAbility != "" {
		if _, ok2 := combat.ArtifactEffects[tmpl.GrantsAbility]; ok2 {
			return tmpl.GrantsAbility
		}
	}
	return id
}

// spawnSkillVisual spawns the legacy spell visual for a skill fired through the
// combat engine, in visual-only mode: the engine has already dealt the damage
// (EventDamageDealt), so the projectile/field only animates. Returns false when
// the skill has no bespoke visual (caller falls back to a particle burst).
//
// All positions are world ground points (package coords):
//
//   - cursor is the point under the mouse when the skill was pressed.
//   - target is where the engine applied the skill — the locked monster's
//     BodyCenter() when hasTarget, otherwise the cursor point.
//
// Every damaging visual lands on target, because that is where the damage
// landed. Only effects whose game logic really is placed by the cursor (the
// healing canopy) use cursor.
func (g *Game) spawnSkillVisual(id string, cursor, target coords.WorldPos, hasTarget, killed bool) bool {
	if g.player == nil || g.currentLevel == nil {
		return false
	}
	bx := g.player.BodyX()
	by := g.player.BodyY()
	info := spells.SpellInfo{Name: id, Level: 1, Damage: 0}

	// Ground effects strike the floor at the target. Body-height effects
	// (projectiles, beams) fly body to body; with no body to hit they are
	// aimed so that they pass through the cursor.
	groundX, groundY := target.X, target.Y
	lockX, lockY := target.X, target.Y
	flyX, flyY := target.X, target.Y
	if !hasTarget {
		aim := cursor.AtBodyHeight()
		flyX, flyY = aim.X, aim.Y
	}

	switch id {
	case "fireball":
		fb := spells.NewFireball(info, bx, by, flyX, flyY, g.fireballSprites, g.spriteSheet.FireBurst)
		fb.VisualOnly = true
		g.ActiveSpells = append(g.ActiveSpells, fb)
	case "chaos_ray":
		cr := spells.NewChaosRay(info, bx, by, flyX, flyY)
		g.ActiveSpells = append(g.ActiveSpells, cr)
	case "lightning":
		ls := spells.NewLightningStrike(info, groundX, groundY, g.spriteSheet.ArcaneBurst)
		ls.DamageApplied = true // engine owns damage
		g.ActiveSpells = append(g.ActiveSpells, ls)
	case "lightning_storm":
		storm := spells.NewLightningStorm(info, groundX, groundY, 3, 0.2, 3.0, g.player.Caster, g.spriteSheet.ArcaneBurst, g.currentLevel)
		storm.VisualOnly = true
		g.ActiveSpells = append(g.ActiveSpells, storm)
	case "fractal_bloom":
		bloom := spells.NewFractalBloom(info, groundX, groundY, g.player.Caster, g.spriteSheet.ArcaneBurst, g.currentLevel, 3, 0.7, 0.2)
		bloom.VisualOnly = true
		g.ActiveSpells = append(g.ActiveSpells, bloom)
	case "fractal_canopy":
		// Canopy heals through the game layer on both combat paths — the engine
		// does not model healing fields — so this is the full legacy spawn.
		fc := &spells.FractalCanopy{
			MaxGrowTime: 5,
			MaxDuration: 10,
			MaxRadius:   5,
			HealingMin:  3,
			HealingMax:  15,
			X:           cursor.X,
			Y:           cursor.Y,
			Visual:      spells.NewFractalCanopyVisual(cursor.X, cursor.Y, 10),
		}
		g.ActiveSpells = append(g.ActiveSpells, fc)
	case "arcane_bolt":
		ab := spells.NewArcaneBolt(info, bx, by, flyX, flyY)
		ab.VisualOnly = true
		g.ActiveSpells = append(g.ActiveSpells, ab)

	// ── Artifact skills: bespoke visuals matching each skill's identity.
	// These act on the locked target, not the cursor. ──
	case "ironbreaker_slam":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewGroundSlam(bx, by, 1.5))
	case "shadowstep":
		g.blinkBehindTarget(lockX, lockY) // teleport + its own ShadowStrike visual
	case "warden_taunt":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewTauntPulse(bx, by, 4.0))
	case "void_bind":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewRootBind(lockX, lockY))
	case "execute":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewReaperExecute(lockX, lockY, killed))
	case "arcane_surge_blast":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewArcaneSurgeNuke(lockX, lockY))
	case "blood_price_strike":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewBloodPriceStrike(bx, by, lockX, lockY))
	case "void_rift_blast":
		g.ActiveSpells = append(g.ActiveSpells, spells.NewVoidRift(lockX, lockY))
	default:
		return false
	}
	return true
}

// blinkBehindTarget teleports the player to the far side of the target — the
// game-layer half of shroud_cloak's blink strike (the engine sets InShadow and
// the guaranteed crit; movement is a game-layer concern). Spawns a
// ShadowStrike (teleport smoke + crit-flash) rather than the plain blink
// trail used by the "blink" dash ability, to sell shroud_cloak's darker,
// guaranteed-critical identity.
func (g *Game) blinkBehindTarget(mx, my float64) {
	if g.player == nil || g.currentLevel == nil {
		return
	}
	// Everything here is in world space: the player's body, the target's
	// body (mx, my) and the point 0.9 tiles beyond it.
	from := g.player.Pos().BodyCenter()
	dx := mx - from.X
	dy := my - from.Y
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return
	}
	// Aim ~0.9 tiles past the target so the destination lands on the far
	// side (or as close as walls allow).
	aimX := mx + dx/dist*0.9
	aimY := my + dy/dist*0.9
	dest, ok := g.blinkDestination(from, coords.WorldPos{X: aimX, Y: aimY})
	if !ok {
		return
	}
	g.teleportPlayerBodyTo(dest)
	g.ActiveSpells = append(g.ActiveSpells, spells.NewShadowStrike(from.X, from.Y, dest.X, dest.Y))
}

// blinkDestination returns where a blink from one world point toward another
// ends: the centre of the last walkable tile on the way (never inside or past
// a wall). ok is false when the blink would not move the player.
func (g *Game) blinkDestination(from, toward coords.WorldPos) (coords.WorldPos, bool) {
	x, y := spells.FindBlinkTarget(g.currentLevel, from.X, from.Y, toward.X, toward.Y)
	reached := coords.WorldPos{X: x, Y: y}
	dest := coords.TileCenter(reached.TileX(), reached.TileY())
	if dest.DistTo(from) < 0.5 {
		return from, false
	}
	return dest, true
}

// teleportPlayerBodyTo moves the player so that their body stands on the
// given world point.
func (g *Game) teleportPlayerBodyTo(body coords.WorldPos) {
	pos := body.EntityPos()
	g.player.MoveController.Stop()
	g.player.MoveController.InterpX = pos.X
	g.player.MoveController.InterpY = pos.Y
	g.player.TileX = body.TileX()
	g.player.TileY = body.TileY()
	g.player.CollisionBox.X = pos.X
	g.player.CollisionBox.Y = pos.Y - (g.player.CollisionBox.Height / 2)
}
