package entities

import (
	"dungeoneer/levels"
	"dungeoneer/pathing"
	"dungeoneer/sprites"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	Motion        EntityMotion
	LeftFacing    bool
	Path          []pathing.PathNode
	PathPreview   []pathing.PathNode
	TickCount     int
	BobOffset     float64
	Sprite        *ebiten.Image
	HP, MaxHP     int
	Damage        int
	AttackRate    int
	AttackTick    int
	IsDead        bool
	Width, Height float64
}

func NewPlayer(ss *sprites.SpriteSheet) *Player {
	return &Player{
		Motion: EntityMotion{
			X:       3,
			Y:       3,
			Speed:   3.0, // tunable speed (tiles/sec)
			TileX:   3,
			TileY:   3,
			InterpX: 3,
			InterpY: 3,
		},
		LeftFacing: true,
		Sprite:     ss.GreyKnight,
		HP:         50,
		MaxHP:      50,
		Damage:     2,
		AttackRate: 60,
		IsDead:     false,
		Width:      .5,
		Height:     .3,
	}
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if p.IsDead {
		return
	}

	// Optional: more accurate rendering
	op := &ebiten.DrawImageOptions{}
	x, y := isoToScreenFloat(p.Motion.InterpX, p.Motion.InterpY, 64)
	bounds := p.Sprite.Bounds()
	spriteW := float64(bounds.Dx())

	// Then apply the bob
	const verticalOffset = 1.0 // tweak until it feels good
	op.GeoM.Translate(0, -verticalOffset+p.BobOffset)

	// 2. Flip horizontally if facing right
	if !p.LeftFacing {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(spriteW, 0)
	}

	// 3. Move to isometric position
	op.GeoM.Translate(x, y)

	// 4. Apply camera transform
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	//Player Draw
	screen.DrawImage(p.Sprite, op)

	p.UpdateHealthBar(screen, x, y, camX, camY, camScale, cx, cy)

	//Player Bounding Box Debug
	//p.DrawLogicalTileDebug(screen, tileSize, isoToScreen, camX, camY, camScale, cx, cy)
}
func (p *Player) DrawLogicalTileDebug(screen *ebiten.Image, tileSize int, isoToScreen func(int, int) (float64, float64), camX, camY, camScale, cx, cy float64) {
	x, y := isoToScreen(p.Motion.TileX, p.Motion.TileY)

	// Build a transparent blue square
	img := ebiten.NewImage(tileSize, tileSize)
	img.Fill(color.RGBA{0, 0, 255, 0}) // Transparent base

	// Draw border by filling edges
	borderColor := color.RGBA{0, 0, 255, 255} // Solid blue
	for i := 0; i < tileSize; i++ {
		img.Set(i, 0, borderColor)          // top
		img.Set(i, tileSize-1, borderColor) // bottom
		img.Set(0, i, borderColor)          // left
		img.Set(tileSize-1, i, borderColor) // right
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	screen.DrawImage(img, op)
}

func (p *Player) UpdateHealthBar(screen *ebiten.Image, x, y float64, camX, camY, camScale, cx, cy float64) {
	if p.MaxHP > 0 {
		hpPercent := float64(p.HP) / float64(p.MaxHP)
		barWidth := 32.0
		barHeight := 4.0

		// Ensure the width is always at least 1 to avoid panic
		width := int(barWidth * hpPercent)
		if width < 1 {
			width = 1
		}

		hpBar := ebiten.NewImage(int(barWidth), int(barHeight))
		hpBar.Fill(color.RGBA{255, 0, 0, 255})
		hpBarFilled := ebiten.NewImage(int(width), int(barHeight))
		hpBarFilled.Fill(color.RGBA{0, 255, 0, 255})

		// Position HP bar
		barOp := &ebiten.DrawImageOptions{}
		barOp.GeoM.Translate(x-barWidth/2+30, y)
		barOp.GeoM.Translate(-camX, camY)
		barOp.GeoM.Scale(camScale, camScale)
		barOp.GeoM.Translate(cx, cy)

		screen.DrawImage(hpBar, barOp)
		screen.DrawImage(hpBarFilled, barOp)
	}
}

func (p *Player) CanMoveTo(x, y int, level *levels.Level) bool {
	return x >= 0 && y >= 0 && x < level.W && y < level.H
}

func (p *Player) Update(level *levels.Level, delta float64) {
	const bobAmplitude = 1.5
	const bobFrequency = 0.3
	const maxDelta = 0.1  // Prevent tunneling due to large frame skips
	const sweepSteps = 10 // More steps = more precise but slightly more expensive

	// Tick updates
	p.TickCount++
	p.AttackTick++
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFrequency) * bobAmplitude

	// Input
	inputX, inputY := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		inputX += 1
		inputY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		inputX -= 1
		inputY += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		inputX -= 1
		inputY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		inputX += 1
		inputY += 1
	}

	// Normalize input
	mag := math.Hypot(inputX, inputY)
	if mag > 0 {
		inputX /= mag
		inputY /= mag
		p.Motion.SetVelocity(inputX*p.Motion.Speed, inputY*p.Motion.Speed)
		p.LeftFacing = p.Motion.VelocityX < 0
	} else {
		p.Motion.Stop()
	}

	// Clamp delta for stability
	clampedDelta := math.Min(delta, maxDelta)

	// Predict next position
	feetX := p.Motion.X
	feetY := p.Motion.Y
	nextX := feetX + p.Motion.VelocityX*clampedDelta
	nextY := feetY + p.Motion.VelocityY*clampedDelta

	// First sweep X-axis independently
	dx := nextX - feetX
	allowedX := feetX
	for i := 1; i <= sweepSteps; i++ {
		t := float64(i) / float64(sweepSteps)
		testX := feetX + dx*t
		if CollidesWithMap(level, testX, feetY, p.Width, p.Height) {
			break
		}
		allowedX = testX
	}

	// Then sweep Y-axis independently (using updated X position)
	dy := nextY - feetY
	allowedY := feetY
	for i := 1; i <= sweepSteps; i++ {
		t := float64(i) / float64(sweepSteps)
		testY := feetY + dy*t
		if CollidesWithMap(level, p.Motion.X, testY, p.Width, p.Height) {
			break
		}
		allowedY = testY
	}
	testX, testY := allowedX, allowedY
	tileX := int(math.Floor(testX))
	tileY := int(math.Floor(testY))

	if !level.IsWalkable(tileX, tileY) {
		// Try X-only
		tileXonly := int(math.Floor(allowedX))
		tileYonly := int(math.Floor(feetY))
		if level.IsWalkable(tileXonly, tileYonly) && !CollidesWithMap(level, allowedX, feetY, p.Width, p.Height) {
			testX = allowedX
			testY = feetY
		} else if level.IsWalkable(int(math.Floor(feetX)), int(math.Floor(allowedY))) && !CollidesWithMap(level, feetX, allowedY, p.Width, p.Height) {
			// Try Y-only
			testX = feetX
			testY = allowedY
		} else {
			// Revert
			testX = feetX
			testY = feetY
		}
	}

	p.Motion.X = testX
	p.Motion.Y = testY

	// Final guard: don't allow standing on center of unwalkable tile
	finalX := int(math.Floor(p.Motion.X))
	finalY := int(math.Floor(p.Motion.Y))
	if !level.IsWalkable(finalX, finalY) {
		return // skip the rest of Update() to prevent updating tile position
	}

	// Update grid position
	p.Motion.TileX = int(math.Floor(p.Motion.X + 0.5))
	p.Motion.TileY = int(math.Floor(p.Motion.Y + 0.5))
	p.Motion.InterpX = p.Motion.X
	p.Motion.InterpY = p.Motion.Y
}

func (p *Player) CanAttack() bool {
	return p.AttackTick >= p.AttackRate
}

func (p *Player) TakeDamage(dmg int) {
	p.HP -= dmg
	if p.HP < 0 {
		p.HP = 0
		p.IsDead = true
	}
	// TODO: Add visual feedback / flashing / etc
}
