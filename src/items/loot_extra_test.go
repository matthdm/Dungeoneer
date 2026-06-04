package items

import "testing"

func TestRollLoot(t *testing.T) {
	// Need mock registry
	RegisterItem(&ItemTemplate{ID: "loot_1", QuestLocked: false, Quality: RarityCommon})
	RegisterItem(&ItemTemplate{ID: "loot_quest", QuestLocked: true, Quality: RarityCommon})

	table := &LootTableDef{
		BiomeID: "test",
		Entries: []LootEntry{
			{ItemID: "loot_1", Weight: 10, MinFloor: 1, Rarity: RarityCommon},
			{ItemID: "loot_quest", Weight: 10, MinFloor: 1, Rarity: RarityCommon},
			{ItemID: "missing", Weight: 10, MinFloor: 1, Rarity: RarityCommon},
		},
	}

	// Should only pick loot_1
	res := RollLoot(table, 1, 0)
	if res == nil || res.ItemID != "loot_1" {
		t.Errorf("expected loot_1, got %v", res)
	}

	// Null table
	if RollLoot(nil, 1, 0) != nil {
		t.Errorf("expected nil result")
	}

	// Floor too low
	if RollLoot(table, 0, 0) != nil {
		t.Errorf("expected nil result due to MinFloor")
	}
}

func TestRollAbilityItem(t *testing.T) {
	RegisterItem(&ItemTemplate{ID: "ab_1", GrantsAbility: "dash", QuestLocked: false, Quality: RarityUncommon})
	RegisterItem(&ItemTemplate{ID: "no_ab", GrantsAbility: "", QuestLocked: false, Quality: RarityUncommon})

	table := &LootTableDef{
		BiomeID: "test",
		Entries: []LootEntry{
			{ItemID: "ab_1", Weight: 10, MinFloor: 1, Rarity: RarityUncommon},
			{ItemID: "no_ab", Weight: 10, MinFloor: 1, Rarity: RarityUncommon},
		},
	}

	res := RollAbilityItem(table, 1, 0)
	if res == nil || res.ItemID != "ab_1" {
		t.Errorf("expected ab_1, got %v", res)
	}
}

func TestRollChestLoot(t *testing.T) {
	RegisterItem(&ItemTemplate{ID: "ab_item", GrantsAbility: "dash", Quality: RarityRare})
	RegisterItem(&ItemTemplate{ID: "norm_item", Quality: RarityCommon})

	table := &LootTableDef{
		BiomeID: "test",
		Entries: []LootEntry{
			{ItemID: "ab_item", Weight: 10, MinFloor: 1, Rarity: RarityRare},
			{ItemID: "norm_item", Weight: 10, MinFloor: 1, Rarity: RarityCommon},
		},
	}

	// Wooden
	res := RollChestLoot(table, "wooden", 1, 0)
	if len(res) == 0 {
		t.Errorf("expected wooden drop")
	}

	// Gold
	res = RollChestLoot(table, "gold", 1, 0)
	if len(res) != 2 {
		t.Errorf("expected 2 drops for gold, got %v", len(res))
	}

	// Iron
	res = RollChestLoot(table, "iron", 1, 0)
	if len(res) != 1 {
		t.Errorf("expected 1 drop for iron, got %v", len(res))
	}

	// Nil table
	if RollChestLoot(nil, "wooden", 1, 0) != nil {
		t.Errorf("expected nil")
	}
}

func TestBuildDefaultLootTable(t *testing.T) {
	RegisterItem(&ItemTemplate{ID: "def_item", QuestLocked: false, Quality: RarityCommon})
	RegisterItem(&ItemTemplate{ID: "def_quest", QuestLocked: true, Quality: RarityCommon})

	table := BuildDefaultLootTable("any")
	if table == nil {
		t.Fatalf("expected table")
	}
	
	hasDefItem := false
	for _, e := range table.Entries {
		if e.ItemID == "def_item" {
			hasDefItem = true
		}
		if e.ItemID == "def_quest" {
			t.Errorf("quest item should not be in default table")
		}
	}
	if !hasDefItem {
		t.Errorf("expected def_item in table")
	}
}

func TestRarityRank(t *testing.T) {
	RegisterItem(&ItemTemplate{ID: "r_leg", Quality: RarityLegendary})
	RegisterItem(&ItemTemplate{ID: "r_rare", Quality: RarityRare})
	RegisterItem(&ItemTemplate{ID: "r_unc", Quality: RarityUncommon})
	RegisterItem(&ItemTemplate{ID: "r_com", Quality: RarityCommon})

	if rarityRank(nil, "r_leg") != 3 { t.Errorf("expected 3") }
	if rarityRank(nil, "r_rare") != 2 { t.Errorf("expected 2") }
	if rarityRank(nil, "r_unc") != 1 { t.Errorf("expected 1") }
	if rarityRank(nil, "r_com") != 0 { t.Errorf("expected 0") }
	if rarityRank(nil, "missing") != 0 { t.Errorf("expected 0") }
}
