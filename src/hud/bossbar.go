package hud

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// BossHealthBar renders a wide health bar at the top of the screen for boss fights.
type BossHealthBar struct {
	Name         string
	Title        string
	MaxHP        int
	CurrentHP    int
	PhaseMarkers []float64 // HP% markers showing phase thresholds
	Visible      bool
}

// Draw renders the boss bar at the top-center of the screen.
func (b *BossHealthBar) Draw(screen *ebiten.Image, screenW int) {
	if !b.Visible || b.MaxHP <= 0 {
		return
	}

	barW := float32(screenW) * 0.6
	barH := float32(12)
	x := (float32(screenW) - barW) / 2
	y := float32(30)

	// Background (dark).
	vector.DrawFilledRect(screen, x, y, barW, barH, color.RGBA{30, 30, 30, 220}, false)

	// Fill (red → orange → green based on HP%).
	hpPct := float64(b.CurrentHP) / float64(b.MaxHP)
	if hpPct < 0 {
		hpPct = 0
	}
	fillW := float32(hpPct) * barW
	fillColor := bossBarColor(hpPct)
	vector.DrawFilledRect(screen, x, y, fillW, barH, fillColor, false)

	// Phase markers (notches).
	for _, pct := range b.PhaseMarkers {
		mx := x + float32(pct)*barW
		vector.DrawFilledRect(screen, mx-1, y, 2, barH, color.RGBA{255, 255, 255, 180}, false)
	}

	// Border.
	vector.StrokeRect(screen, x, y, barW, barH, 1, color.RGBA{200, 200, 200, 255}, false)

	// Name and title above bar.
	label := b.Name
	if b.Title != "" {
		label = b.Name + " — " + b.Title
	}
	face := basicfont.Face7x13
	textW := len(label) * 7 // approximate width
	tx := (screenW - textW) / 2
	ty := int(y) - 6
	text.Draw(screen, label, face, tx, ty, color.RGBA{220, 200, 180, 255})
}

// bossBarColor returns a fill color based on HP percentage.
func bossBarColor(pct float64) color.RGBA {
	switch {
	case pct > 0.6:
		return color.RGBA{180, 30, 30, 255} // deep red
	case pct > 0.25:
		return color.RGBA{200, 120, 20, 255} // orange
	default:
		return color.RGBA{200, 40, 40, 255} // bright red
	}
}
