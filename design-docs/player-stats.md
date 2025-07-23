 Design Document: Player Stats & Save System Refactor

Feature: Character Stats + Save/Load Support
Game: Dungeoneer
Goal: Extend the Player struct to support RPG stats (Strength, Dexterity, etc.), enable full serialization of the character, and integrate with inventory and combat systems.
 Design Goals

    Introduce a flexible stat system to the Player struct

    Cleanly separate base stats, equipment-modified stats, and temporary modifiers

    Ensure all character state (stats, inventory, equipment) is serializable to/from JSON

    Lay groundwork for future systems: leveling, buffs, status effects

 Core Concepts
Stat Categories
Stat	Purpose
Strength	Affects melee damage, carry weight
Dexterity	Affects accuracy, ranged weapons
Vitality	Affects max HP, resistance
Intelligence	Affects spell power, mana (future)
Luck	Influences crits, item drops
 Player Struct Refactor
 New Fields

type Player struct {
	// Existing
	TileX, TileY int
	MoveController *movement.MovementController
	CollisionBox collision.Box

	// Core Stats
	Stats BaseStats
	TempModifiers StatModifiers // Optional effects (buffs/debuffs)
	Inventory *inventory.Inventory
	Equipment map[string]*items.Item // e.g. "Weapon", "Armor"

	// Derived
	HP, MaxHP int
	Damage int
	AttackRate int

	// Save ID
	Name string
}

BaseStats

type BaseStats struct {
	Strength     int
	Dexterity    int
	Vitality     int
	Intelligence int
	Luck         int
}

StatModifiers

type StatModifiers struct {
	StrengthMod     int
	DexterityMod    int
	VitalityMod     int
	IntelligenceMod int
	LuckMod         int
}

 Derived Stats Calculations

Add a helper function to recalculate derived stats from base + modifiers + equipment:

func (p *Player) RecalculateStats() {
	equipMod := p.getEquipmentStatModifiers()

	p.MaxHP = 100 + (p.Stats.Vitality + p.TempModifiers.VitalityMod + equipMod.VitalityMod) * 5
	p.Damage = 5 + (p.Stats.Strength + equipMod.StrengthMod) * 2
	p.AttackRate = 60 - (p.Stats.Dexterity + equipMod.DexterityMod) * 2
}

 Serialization
Save Format

type PlayerSave struct {
	Name       string
	TileX      int
	TileY      int
	Stats      BaseStats
	Inventory  [][]items.ItemSave
	Equipment  map[string]items.ItemSave
	HP         int
}

Save Function

func (p *Player) ToSaveData() PlayerSave {
	return PlayerSave{
		Name:      p.Name,
		TileX:     p.TileX,
		TileY:     p.TileY,
		Stats:     p.Stats,
		HP:        p.HP,
		Inventory: p.Inventory.ToSaveData(),
		Equipment: serializeEquipment(p.Equipment),
	}
}

Load Function

func LoadPlayer(data PlayerSave) *Player {
	player := &Player{
		Name:      data.Name,
		TileX:     data.TileX,
		TileY:     data.TileY,
		Stats:     data.Stats,
		Inventory: inventory.FromSaveData(data.Inventory),
		Equipment: items.DeserializeEquipment(data.Equipment),
		HP:        data.HP,
	}
	player.RecalculateStats()
	return player
}

 Usage in Combat / Effects

    Player.Damage is used during combat  derived from Strength

    Spells in the future may use Intelligence

    Buff items can temporarily adjust TempModifiers, then trigger RecalculateStats()

 Codex Prompt

    You are refactoring a Player struct in an isometric 2D dungeon crawler (Ebiten + Go). Add support for base RPG stats: Strength, Dexterity, Vitality, Intelligence, and Luck. Imlpement the mana and passive regen. More intelligence means more power to spells, more mana pool; mana should have passive regen, unlike health. Include support for equipment bonuses and temporary modifiers. Also implement full JSON-compatible save/load for the player, including stats, inventory, equipment, and HP. Use a PlayerSave struct and helper methods like ToSaveData() and LoadPlayer().