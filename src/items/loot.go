package items

import "math/rand/v2"

// Rarity tiers for loot.
const (
	RarityCommon    = "common"
	RarityUncommon  = "uncommon"
	RarityRare      = "rare"
	RarityLegendary = "legendary"
)

// LootEntry defines a single possible drop.
type LootEntry struct {
	ItemID   string
	Weight   float64
	MinFloor int
	Rarity   string
}

// LootTableDef defines loot for a biome.
type LootTableDef struct {
	BiomeID string
	Entries []LootEntry
}

// LootResult is the outcome of a loot roll.
type LootResult struct {
	ItemID string
	Count  int
}

// ShouldDrop returns true if a killed monster drops loot based on role and floor.
func ShouldDrop(role string, floor int) bool {
	base := 0.30
	switch role {
	case "elite":
		base = 0.60
	case "boss":
		return true
	}
	chance := base + float64(floor)*0.02
	if chance > 0.90 {
		chance = 0.90
	}
	return rand.Float64() < chance
}

// adjustedWeight scales an entry's weight by floor for rarity progression.
func adjustedWeight(e LootEntry, floor int) float64 {
	w := e.Weight
	switch e.Rarity {
	case RarityCommon:
		w *= max(0.1, 1.0-0.03*float64(floor))
	case RarityUncommon:
		w *= 1.0 + 0.01*float64(floor)
	case RarityRare:
		w *= 1.0 + 0.015*float64(floor)
	case RarityLegendary:
		w *= 1.0 + 0.005*float64(floor)
	}
	return w
}

// RollLoot picks an item from the loot table based on floor and luck.
// Returns nil if no valid entry found.
func RollLoot(table *LootTableDef, floor int) *LootResult {
	if table == nil || len(table.Entries) == 0 {
		return nil
	}

	// Filter eligible entries.
	type candidate struct {
		entry  LootEntry
		weight float64
	}
	var pool []candidate
	total := 0.0
	for _, e := range table.Entries {
		if floor < e.MinFloor {
			continue
		}
		// Verify item exists in registry.
		if _, ok := Registry[e.ItemID]; !ok {
			continue
		}
		w := adjustedWeight(e, floor)
		pool = append(pool, candidate{e, w})
		total += w
	}
	if len(pool) == 0 {
		return nil
	}

	r := rand.Float64() * total
	for _, c := range pool {
		r -= c.weight
		if r <= 0 {
			return &LootResult{ItemID: c.entry.ItemID, Count: 1}
		}
	}
	return &LootResult{ItemID: pool[len(pool)-1].entry.ItemID, Count: 1}
}

// BuildDefaultLootTable creates a generic loot table from all items in the
// registry. This is used when a biome doesn't define its own table.
func BuildDefaultLootTable(biomeID string) *LootTableDef {
	table := &LootTableDef{BiomeID: biomeID}
	for id := range Registry {
		table.Entries = append(table.Entries, LootEntry{
			ItemID:   id,
			Weight:   1.0,
			MinFloor: 1,
			Rarity:   RarityCommon,
		})
	}
	return table
}
