package entities

type HitMarker struct {
	X, Y     float64 // world position in tile coords (float64 for Interp support)
	Ticks    int     // how long it's been alive
	MaxTicks int     // how long to display
}
