package entities

import (
	"dungeoneer/collision"
	"dungeoneer/constants"
	"dungeoneer/images"
	"dungeoneer/inventory"
	"dungeoneer/items"
	"dungeoneer/levels"
	"dungeoneer/movement"
	"dungeoneer/pathing"
	"dungeoneer/progression"
	"dungeoneer/spells"
	"dungeoneer/sprites"
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// BaseStats are the fundamental RPG attributes.
type BaseStats struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Vitality     int `json:"vitality"`
	Intelligence int `json:"intelligence"`
	Luck         int `json:"luck"`
}

// StatModifiers represent temporary or equipment-derived stat bonuses.
type StatModifiers struct {
	StrengthMod     int `json:"strength_mod"`
	DexterityMod    int `json:"dexterity_mod"`
	VitalityMod     int `json:"vitality_mod"`
	IntelligenceMod int `json:"intelligence_mod"`
	LuckMod         int `json:"luck_mod"`
}

type Player struct {
	TileX, TileY int

	Sprite     *ebiten.Image
	LeftFacing bool

	PathPreview []pathing.PathNode
	TickCount   int
	BobOffset   float64

	HP, MaxHP  int
	Damage     int
	AttackRate int
	AttackTick int
	IsDead     bool

	MoveController *movement.MovementController

	CollisionBox collision.Box

	DashCharges   int
	DashCooldowns [constants.MaxDashCharges]float64
	IsDashing     bool
	DashTimer     float64

	Grapple Grapple

	Caster *spells.Caster

	Inventory     *inventory.Inventory
	Stats         BaseStats
	TempModifiers StatModifiers
	Equipment     map[string]*items.Item

	Level         int
	EXP           int
	UnspentPoints int

	Mana, MaxMana int
	Gold          int

	Name string

	LastMoveDirX float64
	LastMoveDirY float64
}

func NewPlayer(ss *sprites.SpriteSheet) *Player {
	mc := movement.NewMovementController(5) // Speed = 3 tiles/sec
	mc.InterpX = 3
	mc.InterpY = 3
	blackMage, _ := images.LoadEmbeddedImage(images.Black_Mage_Full_png)
	p := &Player{
		TileX:          3,
		TileY:          3,
		LeftFacing:     true,
		Sprite:         blackMage, //ss.GreyKnight,
		HP:             100,
		MaxHP:          100,
		Damage:         8,
		AttackRate:     60,
		IsDead:         false,
		MoveController: mc,
		CollisionBox:   collision.Box{X: 3, Y: 3, Width: 0.55, Height: 0.8},
		DashCharges:    constants.MaxDashCharges,
		Grapple: Grapple{
			MaxDistance: constants.GrappleMaxDistance,
			Speed:       constants.GrappleSpeed,
			Delay:       constants.GrappleDelay,
		},
		Caster:    spells.NewCaster(),
		Inventory: inventory.New(inventory.Width, inventory.Height),
		Stats: BaseStats{
			Strength:     1,
			Dexterity:    1,
			Vitality:     1,
			Intelligence: 1,
			Luck:         1,
		},
		TempModifiers: StatModifiers{},
		Equipment: map[string]*items.Item{
			"Head":    nil,
			"Chest":   nil,
			"Weapon":  nil,
			"Offhand": nil,
			"Ring1":   nil,
			"Ring2":   nil,
		},
		Level:         1,
		EXP:           0,
		UnspentPoints: 0,
		Mana:          20,
		MaxMana:       20,
		Gold:          0,
		Name:          "Hero",
		LastMoveDirX:  -1,
		LastMoveDirY:  0,
	}

	// Whenever InterpX/InterpY crosses into a new tile, update TileX/TileY
	mc.OnStep = func(x, y int) {
		p.TileX = x
		p.TileY = y
	}

	p.RecalculateStats()
	return p
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if p.IsDead {
		return
	}

	// 1) Compute screen coords from continuous (InterpX/InterpY)
	sx, sy := isoToScreenFloat(p.MoveController.InterpX, p.MoveController.InterpY, tileSize)

	// 2) Vertically bob up/down
	const verticalOffset = 1.0
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, -verticalOffset+p.BobOffset)
	if p.IsDashing {
		op.ColorScale.Scale(1.3, 1.3, 1.3, 1)
	}

	// 3) Flip if facing right
	b := p.Sprite.Bounds()
	if !p.LeftFacing {
		w := float64(b.Dx())
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(w, 0)
	}

	// 4) Position at feet, then apply camera
	op.GeoM.Translate(sx, sy)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)

	screen.DrawImage(p.Sprite, op)

	// Draw health bar above sprite
	p.drawHealthBar(screen, sx, sy, camX, camY, camScale, cx, cy)
}

func (p *Player) drawHealthBar(screen *ebiten.Image, sx, sy, camX, camY, camScale, cx, cy float64) {
	if p.MaxHP <= 0 {
		return
	}
	frac := float64(p.HP) / float64(p.MaxHP)
	barW := 32.0
	barH := 4.0
	filled := int(barW * frac)
	if filled < 1 {
		filled = 1
	}

	full := ebiten.NewImage(int(barW), int(barH))
	full.Fill(color.RGBA{255, 0, 0, 255})
	part := ebiten.NewImage(filled, int(barH))
	part.Fill(color.RGBA{0, 255, 0, 255})

	barOp := &ebiten.DrawImageOptions{}
	barOp.GeoM.Translate(sx-barW/2+30, sy)
	barOp.GeoM.Translate(-camX, camY)
	barOp.GeoM.Scale(camScale, camScale)
	barOp.GeoM.Translate(cx, cy)

	screen.DrawImage(full, barOp)
	screen.DrawImage(part, barOp)
}

func (p *Player) CanMoveTo(x, y int, level *levels.Level) bool {
	return x >= 0 && y >= 0 && x < level.W && y < level.H
}

func (p *Player) Update(level *levels.Level, dt float64) {
	// Bobbing
	const bobFreq = 0.3
	const bobAmp = 1.5
	p.TickCount++
	p.AttackTick++
	p.BobOffset = math.Sin(float64(p.TickCount)*bobFreq) * bobAmp

	p.updateGrapple(level, dt)
	if p.Caster != nil {
		p.Caster.Update(dt)
	}

	// Passive mana regeneration
	regen := 1 + (p.Stats.Intelligence / 5)
	p.Mana += int(float64(regen) * dt)
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}

	// Track last movement direction
	if p.MoveController.Mode == movement.VelocityMode {
		if p.MoveController.VelocityX != 0 || p.MoveController.VelocityY != 0 {
			mag := math.Hypot(p.MoveController.VelocityX, p.MoveController.VelocityY)
			if mag != 0 {
				p.LastMoveDirX = p.MoveController.VelocityX / mag
				p.LastMoveDirY = p.MoveController.VelocityY / mag
			}
		}
	} else if p.MoveController.Mode == movement.PathingMode && p.MoveController.Moving {
		dx := p.MoveController.TargetX - p.MoveController.InterpX
		dy := p.MoveController.TargetY - p.MoveController.InterpY
		mag := math.Hypot(dx, dy)
		if mag != 0 {
			p.LastMoveDirX = dx / mag
			p.LastMoveDirY = dy / mag
		}
	}

	// Recharge dash charges
	for i := range p.DashCooldowns {
		if p.DashCooldowns[i] > 0 {
			p.DashCooldowns[i] -= dt
			if p.DashCooldowns[i] <= 0 {
				p.DashCooldowns[i] = 0
				if p.DashCharges < constants.MaxDashCharges {
					p.DashCharges++
				}
			}
		}
	}

	// Handle active dash timer
	if p.IsDashing {
		p.DashTimer -= dt
		if p.DashTimer <= 0 {
			p.IsDashing = false
			p.MoveController.Stop()
		}
	}

	// For pathing, let the controller interpolate positions
	if p.MoveController.Mode == movement.PathingMode {
		p.MoveController.Update(dt)
	}

	// If PathingMode, check validity of next node and flip sprite direction
	if p.MoveController.Mode == movement.PathingMode && len(p.MoveController.Path) > 0 && !p.MoveController.Moving {
		next := p.MoveController.Path[0]
		if !p.CanMoveTo(next.X, next.Y, level) {
			p.MoveController.Path = nil
			return
		}
		p.LeftFacing = next.X < p.TileX
	}
	// If pure VelocityMode we will update tile coords after resolving movement

	// Update collision box so its center is at “feet” (InterpX,InterpY – half Height)
	p.CollisionBox.X = p.MoveController.InterpX
	p.CollisionBox.Y = p.MoveController.InterpY - (p.CollisionBox.Height / 2)

	// If moving by velocity, resolve collisions each frame
	if p.MoveController.Mode == movement.VelocityMode {
		vx := p.MoveController.VelocityX * dt
		vy := p.MoveController.VelocityY * dt

		// Clamp displacement to avoid tunneling
		maxStep := 0.25
		if math.Abs(vx) > maxStep {
			vx = math.Copysign(maxStep, vx)
		}
		if math.Abs(vy) > maxStep {
			vy = math.Copysign(maxStep, vy)
		}

		// Sweep box and clip movement against the tile map
		finalBox, hitX, hitY := collision.PredictAndClip(level, p.CollisionBox, vx, vy)

		// Stop velocity on the axes we collided with to prevent wall clipping
		if hitX {
			p.MoveController.VelocityX = 0
		}
		if hitY {
			p.MoveController.VelocityY = 0
		}

		// Apply resolved position back to movement controller
		p.MoveController.InterpX = finalBox.X
		p.MoveController.InterpY = finalBox.Y + (p.CollisionBox.Height / 2)

		// Sync collision box with new position
		p.CollisionBox = finalBox

		// Update tile coordinates from final position
		p.TileX = int(p.MoveController.InterpX)
		p.TileY = int(p.MoveController.InterpY)
	}
}

func (p *Player) StartGrapple(tx, ty float64) {
	// Stop any current movement so the rope begins at the player's
	// actual position rather than where they were headed.
	//p.MoveController.Stop()
	p.Grapple.Active = true
	p.Grapple.Hooking = true
	p.Grapple.Pulling = false
	p.Grapple.StartPos = Vec2{p.MoveController.InterpX, p.MoveController.InterpY}
	p.Grapple.HookPos = p.Grapple.StartPos
	p.Grapple.TargetTile = Vec2{tx, ty}
	p.Grapple.Delay = constants.GrappleDelay
}

func (p *Player) CancelGrapple() {
	p.Grapple.Active = false
	p.Grapple.Hooking = false
	p.Grapple.Pulling = false
	p.MoveController.Stop()
}

func (p *Player) updateGrapple(level *levels.Level, dt float64) {
	g := &p.Grapple
	if !g.Active {
		return
	}

	if g.Hooking {
		// Extend the hook toward the highlighted tile but do not stop there.
		dx := g.TargetTile.X - g.StartPos.X
		dy := g.TargetTile.Y - g.StartPos.Y
		dist := math.Hypot(dx, dy)
		if dist == 0 {
			g.Active = false
			return
		}
		dx /= dist
		dy /= dist
		step := g.Speed * dt
		g.HookPos.X += dx * step
		g.HookPos.Y += dy * step
		traveled := math.Hypot(g.HookPos.X-g.StartPos.X, g.HookPos.Y-g.StartPos.Y)
		if traveled >= g.MaxDistance {
			// Reached max range without hitting a surface
			g.Active = false
			g.Hooking = false
			return
		}
		tx := int(math.Floor(g.HookPos.X))
		ty := int(math.Floor(g.HookPos.Y))
		if !level.IsWalkable(tx, ty) {
			g.Hooking = false
			g.Pulling = true
			g.HookPos = Vec2{float64(tx), float64(ty)}
		}
	} else if g.Pulling {
		if g.Delay > 0 {
			g.Delay -= dt
		}
		dx := g.HookPos.X - p.MoveController.InterpX
		dy := g.HookPos.Y - p.MoveController.InterpY
		dist := math.Hypot(dx, dy)

		// Auto-cancel if we've reached the wall or, after pulling has
		// begun, we've essentially stopped moving
		closeEnough := 1.5 > dist
		//stalled := false
		//if g.Delay <= 0 {
		//	stalled = math.Abs(p.MoveController.VelocityX)+math.Abs(p.MoveController.VelocityY) < 0.01
		//}
		fmt.Println("dist: ", closeEnough, dist)
		if closeEnough { //| stalled {
			//p.MoveController.Stop()
			g.Active = false
			g.Pulling = false
			return
		}

		if g.Delay <= 0 {
			dx /= dist
			dy /= dist
			p.MoveController.Mode = movement.VelocityMode
			p.MoveController.VelocityX = dx * g.Speed
			p.MoveController.VelocityY = dy * g.Speed
		}
	}
}
func (p *Player) StartDash(dirX, dirY float64) {
	if p.DashCharges <= 0 || p.IsDashing {
		return
	}

	mag := math.Hypot(dirX, dirY)
	if mag == 0 {
		if p.LeftFacing {
			dirX = -1
			dirY = 0
		} else {
			dirX = 1
			dirY = 0
		}
		mag = 1
	}

	dirX /= mag
	dirY /= mag

	p.LastMoveDirX = dirX
	p.LastMoveDirY = dirY

	p.IsDashing = true
	p.DashTimer = constants.DashDuration
	dashSpeed := p.MoveController.Speed * constants.DashSpeedMultiplier
	p.MoveController.Mode = movement.VelocityMode
	p.MoveController.VelocityX = dirX * dashSpeed
	p.MoveController.VelocityY = dirY * dashSpeed

	p.DashCharges--
	for i := range p.DashCooldowns {
		if p.DashCooldowns[i] == 0 {
			p.DashCooldowns[i] = constants.DashRecharge
			break
		}
	}
}

func (p *Player) SetPath(path []pathing.PathNode) {
	p.MoveController.SetPath(path)
}

func (p *Player) CanAttack() bool {
	return p.AttackTick >= p.AttackRate
}

func (p *Player) TakeDamage(dmg int) {
	p.HP -= dmg
	if p.HP <= 0 {
		p.HP = 0
		p.IsDead = true
	}
}

// getEquipmentStatModifiers sums stat bonuses from equipped items.
func (p *Player) getEquipmentStatModifiers() StatModifiers {
	mod := StatModifiers{}
	for _, it := range p.Equipment {
		if it == nil {
			continue
		}
		if v, ok := it.Stats["Strength"]; ok {
			mod.StrengthMod += v
		}
		if v, ok := it.Stats["Dexterity"]; ok {
			mod.DexterityMod += v
		}
		if v, ok := it.Stats["Vitality"]; ok {
			mod.VitalityMod += v
		}
		if v, ok := it.Stats["Intelligence"]; ok {
			mod.IntelligenceMod += v
		}
		if v, ok := it.Stats["Luck"]; ok {
			mod.LuckMod += v
		}
	}
	return mod
}

// EffectiveStats returns the player's stats including equipment and temporary modifiers.
func (p *Player) EffectiveStats() BaseStats {
	equip := p.getEquipmentStatModifiers()
	return BaseStats{
		Strength:     p.Stats.Strength + p.TempModifiers.StrengthMod + equip.StrengthMod,
		Dexterity:    p.Stats.Dexterity + p.TempModifiers.DexterityMod + equip.DexterityMod,
		Vitality:     p.Stats.Vitality + p.TempModifiers.VitalityMod + equip.VitalityMod,
		Intelligence: p.Stats.Intelligence + p.TempModifiers.IntelligenceMod + equip.IntelligenceMod,
		Luck:         p.Stats.Luck + p.TempModifiers.LuckMod + equip.LuckMod,
	}
}

// RecalculateStats updates derived fields like MaxHP, Damage, and AttackRate.
func (p *Player) RecalculateStats() {
	equip := p.getEquipmentStatModifiers()
	p.MaxHP = 100 + (p.Stats.Vitality+p.TempModifiers.VitalityMod+equip.VitalityMod)*5
	p.MaxMana = 20 + (p.Stats.Intelligence+p.TempModifiers.IntelligenceMod+equip.IntelligenceMod)*5
	p.Damage = 5 + (p.Stats.Strength+equip.StrengthMod)*2
	p.AttackRate = 60 - (p.Stats.Dexterity+equip.DexterityMod)*2
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}
}

// AddEXP grants experience and handles level ups.
func (p *Player) AddEXP(amount int) {
	p.EXP += amount
	for p.EXP >= progression.EXPToLevel(p.Level) {
		p.EXP -= progression.EXPToLevel(p.Level)
		p.Level++
		p.UnspentPoints += 3
	}
}
