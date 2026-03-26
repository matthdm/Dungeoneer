package game

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"dungeoneer/levels"
	"dungeoneer/ui"
)

// buildDevEntries constructs the DevOverlay entries wired to live game state.
// All closures capture g, so toggles take effect immediately.
func (g *Game) buildDevEntries() []ui.DevEntry {
	return []ui.DevEntry{
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
			Toggle:   func() { g.devToggleAbility("slash_combo", items.AbilitySlotPrimary) },
		},
		{
			Label:    "Grant: Arcane Bolt",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("arcane_bolt") },
			Toggle:   func() { g.devToggleAbility("arcane_bolt", items.AbilitySlotPrimary) },
		},
		{
			Label:    "Grant: Arcane Spray",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("arcane_spray") },
			Toggle:   func() { g.devToggleAbility("arcane_spray", items.AbilitySlotSpell) },
		},
		{
			Label:    "Grant: Dash",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("dash") },
			Toggle:   func() { g.devToggleAbility("dash", items.AbilitySlotDash) },
		},
		{
			Label:    "Grant: Blink",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("blink") },
			Toggle:   func() { g.devToggleAbility("blink", items.AbilitySlotDash) },
		},
		{
			Label:    "Grant: Grapple",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("grapple") },
			Toggle:   func() { g.devToggleAbility("grapple", items.AbilitySlotGrapple) },
		},
		{
			Label:    "Grant: Fireball",
			IsActive: func() bool { return g.player != nil && g.player.HasAbility("fireball") },
			Toggle:   func() { g.devToggleAbility("fireball", items.AbilitySlotSpell) },
		},

		// ── Class Switcher ────────────────────────────────────────────────
		{Label: "CLASS", IsHeader: true},
		{
			Label:    "Play as Knight",
			IsActive: func() bool { return g.player != nil && g.player.Class == entities.ClassKnight },
			Toggle: func() {
				if g.player != nil {
					g.player.Class = entities.ClassKnight
					g.player.ClearAbilities()
					for slot := range g.player.Equipment {
						g.player.Equipment[slot] = nil
					}
					g.player.EquipStarter()
				}
			},
		},
		{
			Label:    "Play as Mage",
			IsActive: func() bool { return g.player != nil && g.player.Class == entities.ClassMage },
			Toggle: func() {
				if g.player != nil {
					g.player.Class = entities.ClassMage
					g.player.ClearAbilities()
					for slot := range g.player.Equipment {
						g.player.Equipment[slot] = nil
					}
					g.player.EquipStarter()
				}
			},
		},
	}
}

// devToggleAbility grants or revokes an ability directly (bypassing equipment).
func (g *Game) devToggleAbility(abilityID string, slot items.AbilitySlotType) {
	if g.player == nil {
		return
	}
	if g.player.HasAbility(abilityID) {
		delete(g.player.Abilities, abilityID)
		// Remove from spell slots if present.
		filtered := g.player.SpellSlots[:0]
		for _, s := range g.player.SpellSlots {
			if s != abilityID {
				filtered = append(filtered, s)
			}
		}
		g.player.SpellSlots = filtered
	} else {
		g.player.Abilities[abilityID] = true
		if slot == items.AbilitySlotSpell && len(g.player.SpellSlots) < 6 {
			g.player.SpellSlots = append(g.player.SpellSlots, abilityID)
		}
	}
}
