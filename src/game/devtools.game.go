package game

import (
	"dungeoneer/combat"
	"dungeoneer/entities"
	"dungeoneer/images"
	"dungeoneer/items"
	"dungeoneer/levels"
	"dungeoneer/ui"
	"sort"
)

// buildDevEntries constructs the DevOverlay entries wired to live game state.
// All closures capture g, so toggles take effect immediately.
func (g *Game) buildDevEntries() []ui.DevEntry {
	entries := []ui.DevEntry{
		// ── Rendering ──────────────────────────────────────────────────────
		{Label: "RENDERING", IsHeader: true},
		{
			Label:    "Fullbright (no fog)",
			IsActive: func() bool { return g.FullBright },
			Toggle:   func() { g.FullBright = !g.FullBright },
		},
		{
			Label:    "Debug Rays",
			IsActive: func() bool { return g.ShowRays },
			Toggle:   func() { g.ShowRays = !g.ShowRays },
		},
		{
			Label:    "Debug Walls",
			IsActive: func() bool { return g.ShowWalls },
			Toggle:   func() { g.ShowWalls = !g.ShowWalls },
		},
		{
			Label:    "Spell Debug",
			IsActive: func() bool { return g.SpellDebug },
			Toggle:   func() { g.SpellDebug = !g.SpellDebug },
		},
		{
			Label:    "HUD",
			Key:      "F10",
			IsActive: func() bool { return g.ShowHUD },
			Toggle:   func() { g.ShowHUD = !g.ShowHUD },
		},

		// ── Editor ─────────────────────────────────────────────────────────
		{Label: "EDITOR", IsHeader: true},
		{
			Label:    "Level Editor",
			Key:      "F3",
			IsActive: func() bool { return g.editor != nil && g.editor.Active },
			Toggle: func() {
				if g.editor != nil {
					g.editor.Active = !g.editor.Active
				}
			},
		},
		{
			Label:    "Sprite Palette",
			IsActive: func() bool { return g.editor != nil && g.editor.PaletteOpen },
			Toggle: func() {
				if g.editor != nil && g.editor.Active {
					g.editor.TogglePalette()
				}
			},
		},
		{
			Label:    "Entity Spawner",
			IsActive: func() bool { return g.editor != nil && g.editor.EntityPaletteOpen },
			Toggle: func() {
				if g.editor != nil && g.editor.Active {
					g.editor.ToggleEntityPalette()
				}
			},
		},
		{
			Label:    "Item Spawner",
			Key:      "F2",
			IsActive: func() bool { return g.DevMenu != nil && g.DevMenu.IsVisible() },
			Toggle: func() {
				if g.DevMenu != nil {
					g.DevMenu.TogglePalette()
				}
			},
		},

		// ── Debug Overlays ─────────────────────────────────────────────────
		{Label: "DEBUG OVERLAYS", IsHeader: true},
		{
			Label:    "Throat Valid",
			IsActive: func() bool { return g.ShowThroatValid },
			Toggle:   func() { g.ShowThroatValid = !g.ShowThroatValid },
		},
		{
			Label:    "Throat Invalid",
			IsActive: func() bool { return g.ShowThroatInvalid },
			Toggle:   func() { g.ShowThroatInvalid = !g.ShowThroatInvalid },
		},
		{
			Label:    "Region Debug",
			IsActive: func() bool { return g.ShowRegionDebug },
			Toggle:   func() { g.ShowRegionDebug = !g.ShowRegionDebug },
		},
		{
			Label:    "Door Debug",
			IsActive: func() bool { return g.ShowDoorDebug },
			Toggle:   func() { g.ShowDoorDebug = !g.ShowDoorDebug },
		},
		{
			Label:    "Show Hitboxes",
			IsActive: func() bool { return g.ShowHitboxes },
			Toggle:   func() { g.ShowHitboxes = !g.ShowHitboxes },
		},
		{
			Label:    "Show Interaction Ranges",
			IsActive: func() bool { return g.ShowInteractionRadii },
			Toggle:   func() { g.ShowInteractionRadii = !g.ShowInteractionRadii },
		},

		// ── Level Generation ───────────────────────────────────────────────
		{Label: "LEVEL GENERATION", IsHeader: true},
		{
			Label: "Generate Maze",
			Toggle: func() {
				if l, err := levels.NewMazeLevel(); err == nil {
					g.currentLevel = l
					g.UpdateSeenTiles(*l)
				}
			},
		},
		{
			Label: "Generate Forest",
			Toggle: func() {
				if l, err := levels.NewForestLevel(); err == nil {
					g.currentLevel = l
					g.UpdateSeenTiles(*l)
				}
			},
		},

		// ── Gameplay Cheats ────────────────────────────────────────────────
		{Label: "GAMEPLAY", IsHeader: true},
		{
			Label:    "God Mode (Inf HP)",
			IsActive: func() bool { return g.GodMode },
			Toggle:   func() { g.GodMode = !g.GodMode },
		},
		{
			Label:    "Infinite Mana",
			IsActive: func() bool { return g.InfMana },
			Toggle:   func() { g.InfMana = !g.InfMana },
		},
		{
			Label: "Full Heal",
			Toggle: func() {
				if g.player != nil {
					g.player.HP = g.player.MaxHP
					g.player.Mana = g.player.MaxMana
				}
			},
		},

		// ── Ability Grants ────────────────────────────────────────────────
		{Label: "ABILITIES", IsHeader: true},
		{
			Label:    "Grant: Slash Combo",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("slash_combo") },
			Toggle:   func() { g.devToggleAbility("slash_combo") },
		},
		{
			Label:    "Grant: Arcane Bolt",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("arcane_bolt") },
			Toggle:   func() { g.devToggleAbility("arcane_bolt") },
		},
		{
			Label:    "Grant: Arcane Spray",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("arcane_spray") },
			Toggle:   func() { g.devToggleAbility("arcane_spray") },
		},
		{
			Label:    "Grant: Dash",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("dash") },
			Toggle:   func() { g.devToggleAbility("dash") },
		},
		{
			Label:    "Grant: Blink",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("blink") },
			Toggle:   func() { g.devToggleAbility("blink") },
		},
		{
			Label:    "Grant: Grapple",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("grapple") },
			Toggle:   func() { g.devToggleAbility("grapple") },
		},
		{
			Label:    "Grant: Fireball",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("fireball") },
			Toggle:   func() { g.devToggleAbility("fireball") },
		},
		{
			Label:    "Grant: Chaos Ray",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("chaos_ray") },
			Toggle:   func() { g.devToggleAbility("chaos_ray") },
		},
		{
			Label:    "Grant: Lightning",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("lightning") },
			Toggle:   func() { g.devToggleAbility("lightning") },
		},
		{
			Label:    "Grant: Lightning Storm",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("lightning_storm") },
			Toggle:   func() { g.devToggleAbility("lightning_storm") },
		},
		{
			Label:    "Grant: Fractal Bloom",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("fractal_bloom") },
			Toggle:   func() { g.devToggleAbility("fractal_bloom") },
		},
		{
			Label:    "Grant: Fractal Canopy",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("fractal_canopy") },
			Toggle:   func() { g.devToggleAbility("fractal_canopy") },
		},

		// ── Combat Testing ───────────────────────────────────────────────
		{Label: "COMBAT TESTING", IsHeader: true},
		{
			Label:    "Combat State Overlay",
			IsActive: func() bool { return g.ShowCombatDebug },
			Toggle:   func() { g.ShowCombatDebug = !g.ShowCombatDebug },
		},
		{
			Label: "Spawn: Melee Enemy",
			Toggle: func() {
				floor := 1
				if g.RunState != nil {
					floor = g.RunState.CurrentFloor
				}
				g.devSpawnCombatEnemy("melee", floor)
			},
		},
		{
			Label: "Spawn: Ranged Enemy",
			Toggle: func() {
				floor := 1
				if g.RunState != nil {
					floor = g.RunState.CurrentFloor
				}
				g.devSpawnCombatEnemy("ranged", floor)
			},
		},
		{
			Label: "Spawn: Elite Enemy",
			Toggle: func() {
				floor := 1
				if g.RunState != nil {
					floor = g.RunState.CurrentFloor
				}
				g.devSpawnCombatEnemy("elite", floor)
			},
		},
		{
			Label:  "Reset Player State",
			Toggle: func() { g.devResetPlayerState() },
		},
		{Label: "-- BUILDS --", IsHeader: true},
	}

	// Build entries generated from scenario files. Each entry reads the JSON,
	// applies the build on activation. Using a closure-captured copy of path.
	scenarioPairs := []struct{ label, file string }{
		{"W1 Iron Flurry", "cmd/benchmarker/scenarios/iron_flurry.json"},
		{"W2 Arcane Farmer", "cmd/benchmarker/scenarios/arcane_farmer.json"},
		{"W3 Nature Sustain", "cmd/benchmarker/scenarios/nature_sustain.json"},
		{"W4 Shadow Burst", "cmd/benchmarker/scenarios/shadow_burst.json"},
		{"W5 Burn DoT", "cmd/benchmarker/scenarios/burn_dot.json"},
		{"W6 Tank Mage", "cmd/benchmarker/scenarios/tank_mage.json"},
		{"M1 CC Chain", "cmd/benchmarker/scenarios/cc_chain.json"},
		{"M2 Nature Bloom Farm", "cmd/benchmarker/scenarios/nature_bloom_farm.json"},
		{"M3 Grapple Momentum", "cmd/benchmarker/scenarios/grapple_momentum.json"},
		{"M4 Chaos Knight", "cmd/benchmarker/scenarios/chaos_knight.json"},
		{"M5 Arcane Surge", "cmd/benchmarker/scenarios/arcane_surge_build.json"},
		{"M6 The 55", "cmd/benchmarker/scenarios/the_55.json"},
		{"M7 Perma Shadow", "cmd/benchmarker/scenarios/perma_shadow.json"},
		{"M8 Void Sacrifice", "cmd/benchmarker/scenarios/void_sacrifice.json"},
		{"Void Rift Test", "cmd/benchmarker/scenarios/void_rift_test.json"},
	}
	for _, sp := range scenarioPairs {
		label := sp.label
		file := sp.file
		entries = append(entries, ui.DevEntry{
			Label: "Load: " + label,
			Toggle: func() {
				s, err := combat.LoadScenario(file)
				if err != nil {
					return
				}
				g.devLoadBuild(s)
			},
		})
	}

	entries = append(entries,
		// ── Class Switcher ────────────────────────────────────────────────
		ui.DevEntry{Label: "CLASS", IsHeader: true},
		ui.DevEntry{
			Label:    "Play as Knight",
			IsActive: func() bool { return g.player != nil && g.player.Class == entities.ClassKnight },
			Toggle: func() {
				if g.player == nil {
					return
				}
				g.player.Class = entities.ClassKnight
				if g.spriteSheet != nil {
					g.player.Sprite = g.spriteSheet.GreyKnight
				}
				g.player.ClearAbilities()
				for slot := range g.player.Equipment {
					g.player.Equipment[slot] = nil
				}
				g.player.EquipStarter()
			},
		},
		ui.DevEntry{
			Label:    "Play as Mage",
			IsActive: func() bool { return g.player != nil && g.player.Class == entities.ClassMage },
			Toggle: func() {
				if g.player == nil {
					return
				}
				g.player.Class = entities.ClassMage
				if img, err := images.LoadEmbeddedImage(images.Black_Mage_Full_png); err == nil {
					g.player.Sprite = img
				}
				g.player.ClearAbilities()
				for slot := range g.player.Equipment {
					g.player.Equipment[slot] = nil
				}
				g.player.EquipStarter()
			},
		},
	)
	return entries
}

// devToggleAbility grants or revokes an ability by equipping (or unequipping)
// the real item that GrantsAbility == abilityID — the same
// player.Equipment + RefreshAbilities path a real pickup/equip uses, so a
// dev-granted ability is indistinguishable from a legitimately-earned one to
// HasAbility, item stat bonuses, and set-bonus counting. A previous version
// wrote directly into player.Abilities, which desynced HasAbility from the
// equipment the inventory screen shows and let elite-only abilities (e.g.
// chaos_ray, lightning_storm, fractal_canopy) be granted with no backing
// elite artifact ever placed in the loadout. Elite items are equipped into
// the elite slot (EquipmentSlotOrder[6]) and mirrored into
// Meta.ArtifactLoadout[6] so the elite HUD bar position picks them up,
// matching equipArtifactLoadout/devLoadBuild.
func (g *Game) devToggleAbility(abilityID string) {
	if g.player == nil {
		return
	}
	if g.player.Equipment == nil {
		g.player.Equipment = entities.NewEquipmentSlots()
	}

	if g.player.HasAbility(abilityID) {
		for _, slot := range entities.EquipmentSlotOrder {
			it := g.player.Equipment[slot]
			if it == nil || it.GrantsAbility != abilityID {
				continue
			}
			g.player.Equipment[slot] = nil
			if slot == entities.EquipmentSlotOrder[6] && g.Meta != nil && g.Meta.ArtifactLoadout[6] == it.ID {
				g.Meta.ArtifactLoadout[6] = ""
			}
		}
	} else {
		id, tmpl := itemTemplateForAbility(abilityID)
		if tmpl == nil {
			return
		}
		if tmpl.IsElite {
			slot := entities.EquipmentSlotOrder[6]
			g.player.Equipment[slot] = items.NewItem(id)
			if g.Meta != nil {
				g.Meta.ArtifactLoadout[6] = id
			}
		} else {
			placed := false
			for i := 0; i < 6; i++ {
				slot := entities.EquipmentSlotOrder[i]
				if g.player.Equipment[slot] == nil {
					g.player.Equipment[slot] = items.NewItem(id)
					placed = true
					break
				}
			}
			if !placed {
				return // no free non-elite slot to equip into
			}
		}
	}

	g.player.RefreshAbilities()
	g.player.RecalculateStats()
	if g.HUD != nil {
		g.syncHUDSpellSlots()
	}
}

// itemTemplateForAbility returns the (id, template) of the first item in the
// registry — in stable ID order — whose GrantsAbility matches abilityID.
// Some abilities (dash, blink) are granted by more than one item; any of
// them satisfies HasAbility, so picking deterministically just keeps dev
// tool behavior reproducible across runs.
func itemTemplateForAbility(abilityID string) (string, *items.ItemTemplate) {
	ids := make([]string, 0, len(items.Registry))
	for id := range items.Registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if tmpl := items.Registry[id]; tmpl.GrantsAbility == abilityID {
			return id, tmpl
		}
	}
	return "", nil
}
