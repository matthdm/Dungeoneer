package entities

type EntityMotion struct {
	X, Y      float64 // World position
	VelocityX float64
	VelocityY float64
	Speed     float64 // Units per second (tunable)

	// Tile grid compatibility
	TileX, TileY int

	// Interpolation rendering
	InterpX, InterpY float64
	StartX, StartY   float64
	TargetX, TargetY float64
	InterpTicks      int
	Moving           bool
}

func (m *EntityMotion) SetVelocity(vx, vy float64) {
	m.VelocityX = vx
	m.VelocityY = vy
}

func (m *EntityMotion) Stop() {
	m.VelocityX = 0
	m.VelocityY = 0
}

func (m *EntityMotion) UpdatePosition(delta float64) {
	m.X += m.VelocityX * delta
	m.Y += m.VelocityY * delta

	m.TileX = int(m.X)
	m.TileY = int(m.Y)

	// For isometric render alignment
	m.InterpX = m.X
	m.InterpY = m.Y
}
