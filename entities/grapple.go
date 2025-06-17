package entities

// Grapple represents the player's grappling hook state.
type Grapple struct {
	Active      bool
	HookPos     Vec2 // current hook head position
	StartPos    Vec2 // starting position of the hook
	TargetTile  Vec2 // target tile in world coords
	Hooking     bool // true when extending
	Pulling     bool // true when pulling player
	MaxDistance float64
	Speed       float64
	Delay       float64 // delay before pull starts
}

// Vec2 is a simple 2D vector in tile coordinates.
type Vec2 struct {
	X, Y float64
}
