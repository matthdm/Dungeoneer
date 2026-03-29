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
		// Verify item exists in registry and is not quest-locked.
		tmpl, ok := Registry[e.ItemID]
		if !ok || tmpl.QuestLocked {
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

// RollAbilityItem picks a random ability-granting item from the table,
// ignoring quest-locked entries. Returns nil if no eligible ability item exists.
func RollAbilityItem(table *LootTableDef, floor int) *LootResult {
	if table == nil {
		return nil
	}
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
		tmpl, ok := Registry[e.ItemID]
		if !ok || tmpl.QuestLocked || tmpl.GrantsAbility == "" {
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

// RollChestLoot rolls loot for a chest based on its variant.
// Wooden chests roll normal loot. Iron chests bias toward uncommon+. Gold/locked
// chests guarantee ability items when possible, falling back to normal loot.
func RollChestLoot(table *LootTableDef, variant string, floor int) []*LootResult {
	if table == nil {
		return nil
	}
	switch variant {
	case "gold", "locked":
		// Try for an ability item first; always produce at least one drop.
		var results []*LootResult
		if r := RollAbilityItem(table, floor); r != nil {
			results = append(results, r)
		}
		// Second roll: normal loot (bonus drop for premium chests).
		if r := RollLoot(table, floor); r != nil {
			results = append(results, r)
		}
		return results
	case "iron":
		// Bias toward uncommon/rare by re-rolling once and keeping the
		// result with higher rarity weight.
		r1 := RollLoot(table, floor)
		r2 := RollLoot(table, floor)
		if r1 == nil {
			return nil
		}
		if r2 == nil {
			return []*LootResult{r1}
		}
		// Pick the rarer of the two.
		if rarityRank(table, r2.ItemID) > rarityRank(table, r1.ItemID) {
			return []*LootResult{r2}
		}
		return []*LootResult{r1}
	default: // wooden
		if r := RollLoot(table, floor); r != nil {
			return []*LootResult{r}
		}
		return nil
	}
}

// rarityRank returns a numeric rank for an item's rarity (higher = rarer).
func rarityRank(table *LootTableDef, itemID string) int {
	tmpl, ok := Registry[itemID]
	if !ok {
		return 0
	}
	switch tmpl.Quality {
	case RarityLegendary:
		return 3
	case RarityRare:
		return 2
	case RarityUncommon:
		return 1
	default:
		return 0
	}
}

// BuildDefaultLootTable creates a generic loot table from all items in the
// registry. This is used when a biome doesn't define its own table.
func BuildDefaultLootTable(biomeID string) *LootTableDef {
	table := &LootTableDef{BiomeID: biomeID}
	for id, tmpl := range Registry {
		if tmpl.QuestLocked {
			continue
		}
		table.Entries = append(table.Entries, LootEntry{
			ItemID:   id,
			Weight:   1.0,
			MinFloor: 1,
			Rarity:   RarityCommon,
		})
	}
	return table
}
