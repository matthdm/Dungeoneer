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
		xi, yi := g.cartesianToIso(d.X, d.Y)
		drawX := (xi-g.camX)*scale + cx
		drawY := (yi+g.camY)*scale + cy - float64(d.Ticks) // floats up

		alpha := 1.0 - float64(d.Ticks)/float64(d.MaxTicks)
		clr := color.NRGBA{255, 255, 0, uint8(alpha * 255)}

		msg := fmt.Sprintf("%d", d.Value)
		text.Draw(target, msg, basicfont.Face7x13, int(drawX), int(drawY), clr)
	}
}
