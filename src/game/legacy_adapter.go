package game

// LegacyCombatAdapter wraps the existing combat handlers unchanged.
// It is the safe fallback when dev_settings.json has use_legacy_combat: true.
// All new combat inputs are no-ops so existing behaviour is fully preserved.
type LegacyCombatAdapter struct{}

func (a *LegacyCombatAdapter) HandleTargetSelect(g *Game, worldX, worldY float64) {
	// Legacy: left-click attack is already handled by handlePrimaryAttack — no-op here.
}

func (a *LegacyCombatAdapter) HandleTargetNearest(g *Game) {
	// Legacy: no targeting system — no-op.
}

func (a *LegacyCombatAdapter) HandleMoveToAttack(g *Game) {
	// Legacy: space was used for menu confirm — no-op in combat context.
}

func (a *LegacyCombatAdapter) HandleSkillActivation(g *Game, slotIdx int) {
	// Legacy: delegate to existing spell slot activation.
	if g.player == nil {
		return
	}
	g.castSpellSlot(slotIdx)
}

func (a *LegacyCombatAdapter) ProcessTick(g *Game, dt float64) {
	// Legacy: existing combat is handled directly in the Update loop — no-op here.
}
