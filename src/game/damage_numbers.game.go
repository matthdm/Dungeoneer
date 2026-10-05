package game

import (
	"dungeoneer/entities"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

func (g *Game) handleDamageNumbers() {
	var remaining []entities.DamageNumber
	for _, dmg := range g.DamageNumbers {
		dmg.Ticks++
		if dmg.Ticks < dmg.MaxTicks {
			remaining = append(remaining, dmg)
		}
	}
	g.DamageNumbers = remaining
}

func (g *Game) drawDamageNumbers(target *ebiten.Image, scale, cx, cy float64) {
	for _, d := range g.DamageNumbers {
		// d.X/d.Y is the damaged entity's BodyCenter(); numbers rise from
		// just above its head.
		drawX, drawY := g.overheadTextPos(d.X, d.Y, scale, cx, cy)
		drawY -= float64(d.Ticks) // floats up

		alpha := 1.0 - float64(d.Ticks)/float64(d.MaxTicks)
		base := DamageNumberColor(d.Type, d.IsCrit)
		clr := color.NRGBA{base.R, base.G, base.B, uint8(alpha * float64(base.A))}

		msg := fmt.Sprintf("%d", d.Value)
		text.Draw(target, msg, basicfont.Face7x13, int(drawX), int(drawY), clr)
	}
}

func (g *Game) handleHealNumbers() {
	var remaining []entities.DamageNumber
	for _, heal := range g.HealNumbers {
		heal.Ticks++
		if heal.Ticks < heal.MaxTicks {
			remaining = append(remaining, heal)
		}
	}
	g.HealNumbers = remaining
}

func (g *Game) drawHealNumbers(target *ebiten.Image, scale, cx, cy float64) {
	for _, h := range g.HealNumbers {
		drawX, drawY := g.overheadTextPos(h.X, h.Y, scale, cx, cy)
		drawY -= float64(h.Ticks)

		alpha := 1.0 - float64(h.Ticks)/float64(h.MaxTicks)
		clr := color.NRGBA{0, 255, 0, uint8(alpha * 255)} // Green!

		msg := fmt.Sprintf("+%d", h.Value)
		text.Draw(target, msg, basicfont.Face7x13, int(drawX), int(drawY), clr)
	}
}

// overheadTextPos returns the screen position for a short piece of floating
// text above the head of a body standing on world point (x, y).
func (g *Game) overheadTextPos(x, y, scale, cx, cy float64) (float64, float64) {
	sx, sy := g.bodyToScreen(x, y, scale, cx, cy)
	// One more body-height up clears the head; nudge left so 1–2 digit
	// numbers in the 7px font read as centred.
	return sx - 6, sy - 1.5*float64(g.currentLevel.TileSize/4)*scale
}
