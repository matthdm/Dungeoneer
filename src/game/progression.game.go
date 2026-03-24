package game

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"dungeoneer/progression"
)

func (g *Game) awardEXP(m *entities.Monster) {
	if g.player == nil || m == nil {
		return
	}
	exp := progression.CalculateEXPReward(m.Level, g.player.Level)
	g.player.AddEXP(exp)
}

// handleMonsterDeath handles all consequences of a monster dying:
// EXP, kill count, and loot drop.
func (g *Game) handleMonsterDeath(m *entities.Monster) {
	g.awardEXP(m)
	if g.RunState != nil && g.RunState.Active {
		g.RunState.KillCount++
	}
	g.rollAndDropLoot(m)

	// Check if the killed monster is the boss.
	if g.CurrentBoss != nil && g.CurrentBoss.Monster == m {
		g.onBossDefeated()
	}
}

// rollAndDropLoot checks if the monster drops loot and spawns an item drop.
func (g *Game) rollAndDropLoot(m *entities.Monster) {
	if g.FloorCtx == nil || g.FloorCtx.BiomeConfig == nil {
		return
	}
	if !items.ShouldDrop(m.Role, g.FloorCtx.FloorNumber) {
		return
	}

	// Use biome loot table or build a default one.
	table := g.FloorCtx.BiomeConfig.LootTable
	if table == nil {
		table = items.BuildDefaultLootTable(string(g.FloorCtx.Biome))
	}

	result := items.RollLoot(table, g.FloorCtx.FloorNumber)
	if result == nil {
		return
	}
	tmpl, ok := items.Registry[result.ItemID]
	if !ok {
		return
	}
	it := &items.Item{ItemTemplate: tmpl, Count: result.Count}
	g.ItemDrops = append(g.ItemDrops, &entities.ItemDrop{
		TileX: m.TileX,
		TileY: m.TileY,
		Item:  *it,
	})
}
