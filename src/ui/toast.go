package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const toastDuration = 3.0

// Toast is a transient text overlay shown for toastDuration seconds then faded.
type Toast struct {
	Message string
	TTL     float64
}

// NewToast creates a toast with the default display duration.
func NewToast(msg string) *Toast {
	return &Toast{Message: msg, TTL: toastDuration}
}

// Update ticks the toast TTL down by dt seconds. Returns true when expired.
func (t *Toast) Update(dt float64) bool {
	t.TTL -= dt
	return t.TTL <= 0
}

// Draw renders the toast near the bottom centre of the screen with fade-out.
func (t *Toast) Draw(screen *ebiten.Image, screenW, screenH int) {
	if t.TTL <= 0 {
		return
	}
	alpha := float32(1.0)
	if t.TTL < 0.5 {
		alpha = float32(t.TTL / 0.5)
	}

	charW := 7
	padX, padY := 14, 8
	textW := len(t.Message) * charW
	panelW := textW + padX*2
	panelH := 16 + padY*2
	panelX := (screenW - panelW) / 2
	panelY := screenH * 3 / 4

	// Background
	bg := ebiten.NewImage(panelW, panelH)
	bg.Fill(color.NRGBA{0, 0, 0, 180})
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(float64(panelX), float64(panelY))
	bgOp.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(bg, bgOp)

	// Border (1px outline via 4 edge lines)
	borderCol := color.NRGBA{180, 150, 100, 200}
	top := ebiten.NewImage(panelW, 1)
	top.Fill(borderCol)
	bot := ebiten.NewImage(panelW, 1)
	bot.Fill(borderCol)
	left := ebiten.NewImage(1, panelH)
	left.Fill(borderCol)
	right := ebiten.NewImage(1, panelH)
	right.Fill(borderCol)
	drawEdge := func(img *ebiten.Image, x, y float64) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(panelX)+x, float64(panelY)+y)
		op.ColorScale.ScaleAlpha(alpha)
		screen.DrawImage(img, op)
	}
	drawEdge(top, 0, 0)
	drawEdge(bot, 0, float64(panelH-1))
	drawEdge(left, 0, 0)
	drawEdge(right, float64(panelW-1), 0)

	// Text
	textImg := ebiten.NewImage(textW+2, 14)
	ebitenutil.DebugPrintAt(textImg, t.Message, 0, 0)
	textOp := &ebiten.DrawImageOptions{}
	textOp.GeoM.Translate(float64(panelX+padX), float64(panelY+padY))
	textOp.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(textImg, textOp)
}
