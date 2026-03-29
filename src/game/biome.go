package game

import (
	"dungeoneer/items"
	"dungeoneer/sprites"

	"github.com/hajimehoshi/ebiten/v2"
)

// EnemyDef describes a single enemy archetype within a biome's pool.
type EnemyDef struct {
	ID         string // unique key, e.g. "crypt_melee"
	Name       string // display name, e.g. "Grey Knight"
	Role       string // "melee", "ranged", "elite", "swarm", "caster", "ambush"
	SpriteID   string // maps to a SpriteSheet field via SpriteMap
	BaseHP     int
	BaseDamage int
	BaseSpeed  int // MovementDuration in ticks (lower = faster)
	AttackRate int // ticks between attacks
	Behavior   string // "roaming", "ambush", "patrol", "ranged", "swarm"
}

// GenParamOverrides allows a biome to override specific generation parameters.
// Nil pointer fields mean "use default".
type GenParamOverrides struct {
	RoomCountMin   *int
	RoomCountMax   *int
	RoomWMin       *int
	RoomWMax       *int
	CorridorWidth  *int
	DoorLockChance *float64
	CoverageTarget *float64
}

// BiomeConfig defines the visual, mechanical, and thematic identity of a biome.
type BiomeConfig struct {
	ID           string
	Name         string
	WallFlavor   string
	FloorFlavor  string
	GenOverrides *GenParamOverrides
	EnemyPool    []EnemyDef
	LootTable    *items.LootTableDef // nil = auto-generate from registry
}

// EnemyByRole returns the first EnemyDef matching the given role, or nil.
func (bc *BiomeConfig) EnemyByRole(role string) *EnemyDef {
	for i := range bc.EnemyPool {
		if bc.EnemyPool[i].Role == role {
			return &bc.EnemyPool[i]
		}
	}
	return nil
}

// biomeLootSupplement returns extra loot entries that boost thematic ability items
// for a biome. These are merged on top of the default table so ability items
// matching the biome theme appear more often while remaining available elsewhere.
func biomeLootSupplement(entries ...items.LootEntry) *items.LootTableDef {
	return &items.LootTableDef{Entries: entries}
}

// BiomeConfigs maps each Biome constant to its configuration.
var BiomeConfigs = map[Biome]*BiomeConfig{
	BiomeCrypt: {
		ID: "crypt", Name: "Crypt",
		WallFlavor: "crypt", FloorFlavor: "crypt",
		// Dark arcane theme: fire, chaos, and grapple feel at home here.
		LootTable: biomeLootSupplement(
			items.LootEntry{ItemID: "item_2_24", Weight: 2.0, MinFloor: 1, Rarity: items.RarityUncommon}, // Fireball Emblem
			items.LootEntry{ItemID: "item_0_3", Weight: 2.0, MinFloor: 1, Rarity: items.RarityUncommon},  // Chaos Emblem
			items.LootEntry{ItemID: "item_1_12", Weight: 1.5, MinFloor: 2, Rarity: items.RarityUncommon}, // Grips of the Buried Flame → grapple
		),
		EnemyPool: []EnemyDef{
			{ID: "crypt_melee", Name: "Grey Knight", Role: "melee", SpriteID: "GreyKnight", BaseHP: 30, BaseDamage: 8, BaseSpeed: 30, AttackRate: 45, Behavior: "roaming"},
			{ID: "crypt_ranged", Name: "Sorcerer", Role: "ranged", SpriteID: "Sorcerer", BaseHP: 20, BaseDamage: 6, BaseSpeed: 35, AttackRate: 60, Behavior: "ranged"},
			{ID: "crypt_elite", Name: "Demon Knight", Role: "elite", SpriteID: "DemonKnight", BaseHP: 80, BaseDamage: 15, BaseSpeed: 25, AttackRate: 40, Behavior: "roaming"},
			{ID: "crypt_swarm", Name: "Apparition", Role: "swarm", SpriteID: "Apparition", BaseHP: 8, BaseDamage: 3, BaseSpeed: 20, AttackRate: 30, Behavior: "swarm"},
			{ID: "crypt_caster", Name: "Death", Role: "caster", SpriteID: "Death", BaseHP: 25, BaseDamage: 10, BaseSpeed: 35, AttackRate: 70, Behavior: "ranged"},
			{ID: "crypt_ambush", Name: "Chimera", Role: "ambush", SpriteID: "Chimera", BaseHP: 40, BaseDamage: 12, BaseSpeed: 25, AttackRate: 40, Behavior: "ambush"},
		},
	},
	BiomeMoss: {
		ID: "moss", Name: "Moss",
		WallFlavor: "moss", FloorFlavor: "moss",
		// Ancient nature theme: fractal spells and lightning fit the primal wilds.
		LootTable: biomeLootSupplement(
			items.LootEntry{ItemID: "item_2_63", Weight: 2.0, MinFloor: 1, Rarity: items.RarityUncommon}, // Verdant Tome → fractal_bloom
			items.LootEntry{ItemID: "item_2_55", Weight: 2.0, MinFloor: 2, Rarity: items.RarityUncommon}, // Necromancer's Tome → fractal_canopy
			items.LootEntry{ItemID: "item_0_26", Weight: 1.5, MinFloor: 2, Rarity: items.RarityUncommon}, // Rage Emblem → lightning
		),
		EnemyPool: []EnemyDef{
			{ID: "moss_melee", Name: "Caveman", Role: "melee", SpriteID: "Caveman", BaseHP: 35, BaseDamage: 9, BaseSpeed: 28, AttackRate: 45, Behavior: "roaming"},
			{ID: "moss_ranged", Name: "Oracle", Role: "ranged", SpriteID: "Oracle", BaseHP: 22, BaseDamage: 7, BaseSpeed: 32, AttackRate: 55, Behavior: "ranged"},
			{ID: "moss_elite", Name: "Minotaur", Role: "elite", SpriteID: "Minotaur", BaseHP: 100, BaseDamage: 18, BaseSpeed: 22, AttackRate: 50, Behavior: "patrol"},
			{ID: "moss_swarm", Name: "Blue Wisp", Role: "swarm", SpriteID: "BlueMan", BaseHP: 6, BaseDamage: 2, BaseSpeed: 18, AttackRate: 25, Behavior: "swarm"},
			{ID: "moss_caster", Name: "Absolem", Role: "caster", SpriteID: "Absolem", BaseHP: 28, BaseDamage: 9, BaseSpeed: 35, AttackRate: 65, Behavior: "ranged"},
			{ID: "moss_ambush", Name: "Manticore", Role: "ambush", SpriteID: "Manticore", BaseHP: 45, BaseDamage: 14, BaseSpeed: 22, AttackRate: 40, Behavior: "ambush"},
		},
	},
	BiomeGallery: {
		ID: "gallery", Name: "Gallery",
		WallFlavor: "gallery", FloorFlavor: "gallery",
		// Cursed aristocratic theme: dark power and blink suit the haunted halls.
		LootTable: biomeLootSupplement(
			items.LootEntry{ItemID: "item_0_35", Weight: 2.0, MinFloor: 2, Rarity: items.RarityRare},     // Azazel's Pentagram → lightning_storm
			items.LootEntry{ItemID: "item_0_3", Weight: 1.5, MinFloor: 1, Rarity: items.RarityUncommon},  // Chaos Emblem → chaos_ray
			items.LootEntry{ItemID: "item_2_35", Weight: 1.5, MinFloor: 1, Rarity: items.RarityUncommon}, // Haste Carriers → blink
		),
		EnemyPool: []EnemyDef{
			{ID: "gallery_melee", Name: "Red Champion", Role: "melee", SpriteID: "RedChampion", BaseHP: 32, BaseDamage: 10, BaseSpeed: 28, AttackRate: 42, Behavior: "roaming"},
			{ID: "gallery_ranged", Name: "Duchess", Role: "ranged", SpriteID: "Duchess", BaseHP: 18, BaseDamage: 7, BaseSpeed: 33, AttackRate: 55, Behavior: "ranged"},
			{ID: "gallery_elite", Name: "Blue Champion", Role: "elite", SpriteID: "BlueChampion", BaseHP: 90, BaseDamage: 16, BaseSpeed: 24, AttackRate: 45, Behavior: "patrol"},
			{ID: "gallery_swarm", Name: "Tortured Soul", Role: "swarm", SpriteID: "TorturedSoul", BaseHP: 7, BaseDamage: 3, BaseSpeed: 20, AttackRate: 28, Behavior: "swarm"},
			{ID: "gallery_caster", Name: "Celestial", Role: "caster", SpriteID: "Celestial", BaseHP: 24, BaseDamage: 11, BaseSpeed: 36, AttackRate: 68, Behavior: "ranged"},
			{ID: "gallery_ambush", Name: "Griffon", Role: "ambush", SpriteID: "Griffon", BaseHP: 38, BaseDamage: 13, BaseSpeed: 20, AttackRate: 38, Behavior: "ambush"},
		},
	},
	BiomeBrick: {
		ID: "brick", Name: "Brick",
		WallFlavor: "brick", FloorFlavor: "brick",
		// Demonic hellish theme: fire and storm dominate the deeper infernal halls.
		LootTable: biomeLootSupplement(
			items.LootEntry{ItemID: "item_2_24", Weight: 1.5, MinFloor: 1, Rarity: items.RarityUncommon}, // Fireball Emblem
			items.LootEntry{ItemID: "item_0_26", Weight: 2.0, MinFloor: 1, Rarity: items.RarityUncommon}, // Rage Emblem → lightning
			items.LootEntry{ItemID: "item_0_35", Weight: 1.5, MinFloor: 3, Rarity: items.RarityRare},     // Azazel's Pentagram → lightning_storm
		),
		EnemyPool: []EnemyDef{
			{ID: "brick_melee", Name: "Sentinel", Role: "melee", SpriteID: "Sentinel", BaseHP: 28, BaseDamage: 8, BaseSpeed: 30, AttackRate: 45, Behavior: "roaming"},
			{ID: "brick_ranged", Name: "Jester", Role: "ranged", SpriteID: "Jester", BaseHP: 20, BaseDamage: 6, BaseSpeed: 30, AttackRate: 50, Behavior: "ranged"},
			{ID: "brick_elite", Name: "Cyclops", Role: "elite", SpriteID: "Cyclops", BaseHP: 95, BaseDamage: 20, BaseSpeed: 28, AttackRate: 55, Behavior: "roaming"},
			{ID: "brick_swarm", Name: "Lesser Demon", Role: "swarm", SpriteID: "LesserDemon", BaseHP: 8, BaseDamage: 4, BaseSpeed: 22, AttackRate: 30, Behavior: "swarm"},
			{ID: "brick_caster", Name: "Greater Demon", Role: "caster", SpriteID: "GreaterDemon", BaseHP: 30, BaseDamage: 12, BaseSpeed: 34, AttackRate: 65, Behavior: "ranged"},
			{ID: "brick_ambush", Name: "Two Headed Ogre", Role: "ambush", SpriteID: "TwoHeadedOgre", BaseHP: 50, BaseDamage: 15, BaseSpeed: 28, AttackRate: 45, Behavior: "ambush"},
		},
	},
}

// SpriteMap builds a lookup from sprite name to image for the entity system.
func BuildSpriteMap(ss *sprites.SpriteSheet) map[string]*ebiten.Image {
	return map[string]*ebiten.Image{
		// Monsters (Row 7)
		"GreyKnight":   ss.GreyKnight,
		"Chimera":      ss.Chimera,
		"Sentinel":     ss.Sentinel,
		"Sorcerer":     ss.Sorcerer,
		"Duchess":      ss.Duchess,
		"Absolem":      ss.Absolem,
		"Death":        ss.Death,
		"Oracle":       ss.Oracle,
		"Jester":       ss.Jester,
		"GreaterDemon": ss.GreaterDemon,
		"DemonKnight":  ss.DemonKnight,
		"Abomination":  ss.Abomination,
		// Row 8
		"QueenOfDarkness": ss.QueenOfDarkness,
		"LesserDemon":     ss.LesserDemon,
		"TheTerror":       ss.TheTerror,
		"Celestial":       ss.Celestial,
		"Demon":           ss.Demon,
		"Apparition":      ss.Apparition,
		"Griffon":         ss.Griffon,
		"Manticore":       ss.Manticore,
		"Minotaur":        ss.Minotaur,
		"TorturedSoul":    ss.TorturedSoul,
		"ChaosTotem":      ss.ChaosTotem,
		// Row 9 (dragons, etc.)
		"LesserDragon":       ss.LesserDragon,
		"GoldenDragon":       ss.GoldenDragon,
		"Wyvern":             ss.Wyvern,
		"LesserCaveDragon":   ss.LesserCaveDragon,
		"GhostWyvern":        ss.GhostWyvern,
		"AncientDragonRed":   ss.AncientDragonRed,
		"AncientDragonBlack": ss.AncientDragonBlack,
		"AncientDragon":      ss.AncientDragon,
		"PetrifiedDragon":    ss.PetrifiedDragon,
		"GreaterDragon":      ss.GreaterDragon,
		"GreaterRedDragon":   ss.GreaterRedDragon,
		"BlueMan":            ss.BlueMan,
		"Cyclops":            ss.Cyclops,
		"TwoHeadedOgre":      ss.TwoHeadedOgre,
		// Row 10
		"RedChampion":  ss.RedChampion,
		"BlueChampion": ss.BlueChampion,
		"Caveman":      ss.Caveman,
		"RedMan":       ss.RedMan,
	}
}
