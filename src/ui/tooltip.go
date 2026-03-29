package ui

import (
	"dungeoneer/items"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// QualityColor returns the display color for a given quality tier.
func QualityColor(quality string) color.RGBA {
	switch quality {
	case items.RarityUncommon:
		return color.RGBA{80, 220, 80, 255}
	case items.RarityRare:
		return color.RGBA{80, 140, 255, 255}
	case items.RarityLegendary:
		return color.RGBA{255, 160, 0, 255}
	default: // common or empty
		return color.RGBA{200, 200, 200, 255}
	}
}

// DrawItemTooltip renders an item tooltip anchored at (x, y), clamped to the
// screen so it never overflows the window bounds.
func DrawItemTooltip(dst *ebiten.Image, it *items.Item, x, y int) {
	const (
		lineH  = 15
		padX   = 6
		padY   = 6
		charW  = 7
		minW   = 120
		ascent = 11 // basicfont.Face7x13 ascent in pixels
	)

	type tline struct {
		text string
		clr  color.RGBA
	}
	var lines []tline

	nameClr := QualityColor(it.Quality)
	lines = append(lines, tline{it.Name, nameClr})

	typeStr := string(it.Type)
	if it.GrantsAbility != "" {
		typeStr += " — grants " + it.GrantsAbility
	}
	if typeStr != "" {
		lines = append(lines, tline{typeStr, color.RGBA{160, 160, 160, 255}})
	}

	if it.Description != "" {
		lines = append(lines, tline{it.Description, color.RGBA{220, 220, 180, 255}})
	}

	if len(it.Stats) > 0 {
		order := []string{"Strength", "Dexterity", "Vitality", "Intelligence", "Luck"}
		for _, stat := range order {
			if v, ok := it.Stats[stat]; ok {
				clr := color.RGBA{80, 220, 80, 255}
				if v < 0 {
					clr = color.RGBA{220, 80, 80, 255}
				}
				lines = append(lines, tline{fmt.Sprintf("%s %+d", stat, v), clr})
			}
		}
	}

	if it.Effect != nil {
		txt := fmt.Sprintf("%s: %s %d%%", it.Effect.Trigger, it.Effect.Type, it.Effect.MagnitudePct)
		if it.Effect.ChancePct != 0 {
			txt += fmt.Sprintf(" (%d%% chance)", it.Effect.ChancePct)
		}
		lines = append(lines, tline{txt, color.RGBA{200, 180, 255, 255}})
	}

	// Measure width and height.
	w := minW
	for _, ln := range lines {
		if cw := len(ln.text)*charW + padX*2; cw > w {
			w = cw
		}
	}
	h := len(lines)*lineH + padY*2

	// Clamp to screen edges.
	sw, sh := dst.Bounds().Dx(), dst.Bounds().Dy()
	if x+w > sw {
		x = sw - w - 4
	}
	if y+h > sh {
		y = sh - h - 4
	}

	// Background panel.
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{10, 10, 20, 220}, false)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, nameClr, false)

	// Draw each line with its color. text.Draw baseline = y + ascent.
	for i, ln := range lines {
		ty := y + padY + i*lineH + ascent
		text.Draw(dst, ln.text, basicfont.Face7x13, x+padX, ty, ln.clr)
	}
}
