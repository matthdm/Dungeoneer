package game

import (
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
	}
}
