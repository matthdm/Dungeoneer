package game

// checkEnemyIntros fires a flavor toast the first time the player encounters
// a given role+biome combination this run. Called once per frame after monster updates.
func (g *Game) checkEnemyIntros() {
	if g.RunState == nil || g.FloorCtx == nil {
		return
	}
	biome := string(g.FloorCtx.Biome)
	for _, m := range g.Monsters {
		if m == nil || m.IsDead || m.IsEcho || m.IntroFired || m.Role == "" {
			continue
		}
		// Only fire when the monster has started chasing (has a path).
		if len(m.Path) == 0 {
			continue
		}
		key := m.Role + "_" + biome
		if g.RunState.EnemyIntrosSeen[key] {
			m.IntroFired = true
			continue
		}
		line := GetEnemyFlavorLine(m.Role, biome)
		if line != "" {
			g.pendingToasts = append(g.pendingToasts, line)
		}
		g.RunState.EnemyIntrosSeen[key] = true
		m.IntroFired = true
	}
}
