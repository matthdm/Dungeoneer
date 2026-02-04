package levels

type DoorDensityConfig struct {
	RoomCorridorChance float64
	RoomRoomChance     float64
	MaxDoorsPerRoom    int
	MinThroatSpacing   int
}

func DefaultDoorDensityConfig() DoorDensityConfig {
	return DoorDensityConfig{
		RoomCorridorChance: 0.6,
		RoomRoomChance:     0.25,
		MaxDoorsPerRoom:    2,
		MinThroatSpacing:   2,
	}
}

func normalizeDoorDensity(cfg DoorDensityConfig) DoorDensityConfig {
	// If zeroed, apply defaults.
	if cfg.RoomCorridorChance == 0 && cfg.RoomRoomChance == 0 && cfg.MaxDoorsPerRoom == 0 && cfg.MinThroatSpacing == 0 {
		cfg = DefaultDoorDensityConfig()
	}
	if cfg.RoomCorridorChance < 0 {
		cfg.RoomCorridorChance = 0
	}
	if cfg.RoomCorridorChance > 1 {
		cfg.RoomCorridorChance = 1
	}
	if cfg.RoomRoomChance < 0 {
		cfg.RoomRoomChance = 0
	}
	if cfg.RoomRoomChance > 1 {
		cfg.RoomRoomChance = 1
	}
	if cfg.MinThroatSpacing < 0 {
		cfg.MinThroatSpacing = 0
	}
	if cfg.MaxDoorsPerRoom < 0 {
		cfg.MaxDoorsPerRoom = 0
	}
	return cfg
}
