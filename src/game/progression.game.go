package game

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"dungeoneer/progression"
)

// rollGoldDrop returns the gold amount for killing a monster of the given role on a given floor.
func rollGoldDrop(role string, floor int) int {
	base := 3 + floor*2
	switch role {
	case "swarm":
		return base / 2
	case "elite":
		return base * 3
	case "boss":
		return base * 8
	default:
		return base
	}
}

// awardGold adds gold to the player and tracks it in the run state.
func (g *Game) awardGold(m *entities.Monster) {
	if g.player == nil || g.FloorCtx == nil {
		return
	}
	amount := rollGoldDrop(m.Role, g.FloorCtx.FloorNumber)
	g.player.Gold += amount
	if g.RunState != nil {
		g.RunState.GoldEarned += amount
	}
}

func (g *Game) awardEXP(m *entities.Monster) {
	if g.player == nil || m == nil {
		return
	}
	exp := progression.CalculateEXPReward(m.Level, g.player.Level)
	g.player.AddEXP(exp)
}

// handleMonsterDeath handles all consequences of a monster dying:
// EXP, gold, kill count, and loot drop.
func (g *Game) handleMonsterDeath(m *entities.Monster) {
	g.awardEXP(m)
	g.awardGold(m)
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

	// Build the effective loot table: default registry items merged with any
	// biome-specific ability item boosts.
	table := items.BuildDefaultLootTable(string(g.FloorCtx.Biome))
	if g.FloorCtx.BiomeConfig.LootTable != nil {
		table.Entries = append(table.Entries, g.FloorCtx.BiomeConfig.LootTable.Entries...)
	}

	// On floor 1, the first elite or boss guarantees an ability item drop so
	// the player always leaves floor 1 with at least one new ability.
	if g.FloorCtx.FloorNumber == 1 && !g.FloorCtx.AbilityDropped &&
		(m.Role == "elite" || m.Role == "boss") {
		if result := items.RollAbilityItem(table, 1); result != nil {
			if tmpl, ok := items.Registry[result.ItemID]; ok {
				g.spawnDrop(m, tmpl, result.Count)
			}
		}
		g.FloorCtx.AbilityDropped = true
	}

	if !items.ShouldDrop(m.Role, g.FloorCtx.FloorNumber) {
		return
	}
	result := items.RollLoot(table, g.FloorCtx.FloorNumber)
	if result == nil {
		return
	}
	tmpl, ok := items.Registry[result.ItemID]
	if !ok {
		return
	}
	g.spawnDrop(m, tmpl, result.Count)
}

// spawnDrop places an item drop at the monster's tile.
func (g *Game) spawnDrop(m *entities.Monster, tmpl *items.ItemTemplate, count int) {
	it := &items.Item{ItemTemplate: tmpl, Count: count}
	g.ItemDrops = append(g.ItemDrops, &entities.ItemDrop{
		TileX: m.TileX,
		TileY: m.TileY,
		Item:  *it,
	})
}
