package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawTargetRing draws a pulsing red ring under the targeted monster.
// Called from drawPlaying before/during the entity layer draw.
//
// Coordinate transform mirrors worldToScreenPoint (tileCentered=false):
//
//	isoX, isoY = cartesianToIso(worldX, worldY)
//	sx = (isoX - camX) * scale + cx
//	sy = (isoY + camY) * scale + cy
func (g *Game) drawTargetRing(target *ebiten.Image, scale, cx, cy float64) {
	if g.TargetedMonster == nil || g.TargetedMonster.IsDead {
		return
	}
	m := g.TargetedMonster

	// Center the ring on the tile diamond center (same tileCentered=true offset that
	// entity renders apply), not the raw tile anchor which is the upper-left corner.
	sx, sy := g.worldToScreenPoint(m.InterpX, m.InterpY, scale, cx, cy, true)

	// Horizontal radius in screen pixels: 0.5 tiles wide in iso space.
	ts := float64(g.currentLevel.TileSize)
	rx := float32(ts * 0.5 * scale)
	// Iso projection halves the Y axis, so vertical radius is ~half horizontal.
	ry := rx * 0.5

	// Pulse: simple sine wave on game tick.
	pulse := float32(0.7 + 0.3*math.Sin(float64(g.gameTick)*0.18))
	alpha := uint8(200 * pulse)

	ringColor := color.NRGBA{R: 220, G: 40, B: 40, A: alpha}

	// Draw an ellipse approximated by line segments.
	const segments = 32
	var prevX, prevY float32
	for i := 0; i <= segments; i++ {
		t := 2 * math.Pi * float64(i) / float64(segments)
		ex := sx + rx*float32(math.Cos(t))
		ey := sy + ry*float32(math.Sin(t))
		if i > 0 {
			vector.StrokeLine(target, prevX, prevY, ex, ey, 1.5, ringColor, false)
		}
		prevX, prevY = ex, ey
	}
}

// drawTargetPanel draws the target health/name panel at the top-center of the screen.
// Called from drawPlaying during the UI layer draw (after HUD, before toast).
//
// Panel: 280px wide, 36px tall, centered horizontally, 8px from top.
// Shows: monster name (left-aligned), health bar (right section).
func (g *Game) drawTargetPanel(screen *ebiten.Image) {
	if g.TargetedMonster == nil {
		return
	}
	m := g.TargetedMonster

	const panelW = 280
	const panelH = 36
	panelX := (g.w - panelW) / 2
	panelY := 8

	// Background panel.
	panelImg := ebiten.NewImage(panelW, panelH)
	panelImg.Fill(color.NRGBA{R: 10, G: 10, B: 20, A: 200})
	panelOp := &ebiten.DrawImageOptions{}
	panelOp.GeoM.Translate(float64(panelX), float64(panelY))
	screen.DrawImage(panelImg, panelOp)

	// Border.
	borderColor := color.NRGBA{R: 160, G: 40, B: 40, A: 220}
	vector.StrokeRect(screen,
		float32(panelX), float32(panelY),
		float32(panelW), float32(panelH),
		1, borderColor, false)

	// Health bar background.
	const barMargin = 4
	const barH = 8
	barW := panelW - barMargin*2
	barY := panelY + panelH - barH - barMargin
	barBg := ebiten.NewImage(barW, barH)
	barBg.Fill(color.NRGBA{R: 60, G: 10, B: 10, A: 200})
	barBgOp := &ebiten.DrawImageOptions{}
	barBgOp.GeoM.Translate(float64(panelX+barMargin), float64(barY))
	screen.DrawImage(barBg, barBgOp)

	// Health bar fill.
	hpPct := 1.0
	if m.MaxHP > 0 {
		hpPct = math.Max(0, math.Min(1.0, float64(m.HP)/float64(m.MaxHP)))
	}
	fillW := int(float64(barW) * hpPct)
	if fillW > 0 {
		barFill := ebiten.NewImage(fillW, barH)
		barFill.Fill(color.NRGBA{R: 200, G: 40, B: 40, A: 230})
		barFillOp := &ebiten.DrawImageOptions{}
		barFillOp.GeoM.Translate(float64(panelX+barMargin), float64(barY))
		screen.DrawImage(barFill, barFillOp)
	}

	// Monster name — top-left of the panel.
	nameText := m.Name
	if m.Level > 0 {
		nameText = fmt.Sprintf("%s (Lv %d)", m.Name, m.Level)
	}
	ebitenutil.DebugPrintAt(screen, nameText, panelX+6, panelY+4)

	// Kill streak counter drawn inside the panel (bottom-right corner).
	g.drawKillStreak(screen)
}

// drawKillStreak draws the current kill streak counter.
// It is called from drawTargetPanel which is already in the draw loop.
// Colour escalation: white (3–4), orange (5–9), red (10+).
func (g *Game) drawKillStreak(screen *ebiten.Image) {
	if g.KillStreak < 3 {
		return
	}

	label := fmt.Sprintf("x%d", g.KillStreak)

	var col color.NRGBA
	switch {
	case g.KillStreak >= 10:
		col = color.NRGBA{R: 220, G: 40, B: 40, A: 255} // red
	case g.KillStreak >= 5:
		col = color.NRGBA{R: 255, G: 160, B: 30, A: 255} // orange
	default:
		col = color.NRGBA{R: 240, G: 240, B: 240, A: 255} // white
	}

	// Draw a small tinted background chip at bottom-right, then the label.
	const chipW, chipH = 36, 14
	chipX := g.w - chipW - 8
	chipY := g.h - chipH - 8

	chip := ebiten.NewImage(chipW, chipH)
	chip.Fill(color.NRGBA{R: col.R / 4, G: col.G / 4, B: col.B / 4, A: 200})
	chipOp := &ebiten.DrawImageOptions{}
	chipOp.GeoM.Translate(float64(chipX), float64(chipY))
	screen.DrawImage(chip, chipOp)

	vector.StrokeRect(screen,
		float32(chipX), float32(chipY),
		float32(chipW), float32(chipH),
		1, col, false)

	ebitenutil.DebugPrintAt(screen, label, chipX+4, chipY+1)
}
