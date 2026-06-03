package entities

type DamageNumber struct {
	X, Y     float64
	Value    int
	Ticks    int
	MaxTicks int
	Type     string // spell type for color coding (e.g. "fire", "lightning")
	IsCrit   bool   // crits render yellow and larger
}
