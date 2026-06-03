package game

// ShopItem defines an item available for purchase in the hub shop.
type ShopItem struct {
	ItemID   string
	Name     string
	Cost     int // Remnants
	StockMax int // max purchases per run (0 = unlimited)
}

// ShopCatalog is the list of items available in the hub shop.
// These are ability items sold between runs.
// TODO: add consumable item IDs (health_potion, mana_potion) once those are registered.
var ShopCatalog = []ShopItem{
	{ItemID: "item_0_26", Name: "Rage Emblem", Cost: 120, StockMax: 1},
	{ItemID: "item_0_35", Name: "Azazel's Pentagram", Cost: 150, StockMax: 1},
	{ItemID: "item_2_24", Name: "Fireball Emblem", Cost: 100, StockMax: 1},
	{ItemID: "item_0_3", Name: "Chaos Emblem", Cost: 110, StockMax: 1},
	{ItemID: "iron_key", Name: "Iron Key", Cost: 40, StockMax: 3},
}
