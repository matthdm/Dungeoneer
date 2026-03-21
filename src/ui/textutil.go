package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// DrawBigText renders s scaled by the given factor.
func DrawBigText(dst *ebiten.Image, s string, x, y int, scale float64, clr color.Color) {
	img := ebiten.NewImage(len(s)*7, 13)
	text.Draw(img, s, basicfont.Face7x13, 0, 13, clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}

// WrapText splits the string into lines not exceeding max characters.
func WrapText(s string, max int) []string {
	if max <= 0 {
		return []string{s}
	}
	words := strings.Fields(s)
	var lines []string
	line := ""
	for _, w := range words {
		if len(line)+len(w)+1 > max && line != "" {
			lines = append(lines, line)
			line = w
		} else {
			if line != "" {
				line += " "
			}
			line += w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
