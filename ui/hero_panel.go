package ui

import (
	"dungeoneer/entities"
	"dungeoneer/progression"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HeroPanel displays player stats and allows spending attribute points.
type HeroPanel struct {
	rect      image.Rectangle
	player    *entities.Player
	visible   bool
	plusRects map[string]image.Rectangle
}

// NewHeroPanel creates a centered hero panel.
func NewHeroPanel(w, h int, p *entities.Player) *HeroPanel {
	width := 220
	height := 220
	x := (w - width) / 2
	y := (h - height) / 2
	return &HeroPanel{
		rect:      image.Rect(x, y, x+width, y+height),
		player:    p,
		plusRects: make(map[string]image.Rectangle),
	}
}

func (h *HeroPanel) Show()                     { h.visible = true }
func (h *HeroPanel) Hide()                     { h.visible = false }
func (h *HeroPanel) Toggle()                   { h.visible = !h.visible }
func (h *HeroPanel) IsVisible() bool           { return h.visible }
func (h *HeroPanel) SetRect(r image.Rectangle) { h.rect = r }

// Update handles button clicks when visible.
func (h *HeroPanel) Update() {
	if !h.visible || h.player == nil {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		h.Hide()
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && h.player.AttributePoints > 0 {
		mx, my := ebiten.CursorPosition()
		p := image.Pt(mx, my)
		for stat, r := range h.plusRects {
			if p.In(r) {
				switch stat {
				case "Strength":
					h.player.Stats.Strength++
				case "Intelligence":
					h.player.Stats.Intelligence++
				case "Vitality":
					h.player.Stats.Vitality++
				case "Dexterity":
					h.player.Stats.Dexterity++
				}
				h.player.AttributePoints--
				h.player.RecalculateStats()
				break
			}
		}
	}
}

// Draw renders the panel to the screen.
func (h *HeroPanel) Draw(screen *ebiten.Image) {
	if !h.visible || h.player == nil {
		return
	}
	style := DefaultMenuStyles()
	DrawMenuOverlay(screen, DefaultOverlayColor)
	DrawMenuWindow(screen, &style, float32(h.rect.Min.X), float32(h.rect.Min.Y), float32(h.rect.Dx()), float32(h.rect.Dy()))

	x := h.rect.Min.X + 10
	y := h.rect.Min.Y + 20
	ebitenutil.DebugPrintAt(screen, h.player.Name, x, y)
	y += 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d", h.player.Level), x, y)
	y += 20
	required := progression.EXPToLevel(h.player.Level)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("EXP: %d/%d", h.player.EXP, required), x, y)
	barW := h.rect.Dx() - 20
	barH := 8
	percent := float32(h.player.EXP) / float32(required)
	vector.DrawFilledRect(screen, float32(x), float32(y+10), float32(barW)*percent, float32(barH), color.RGBA{200, 200, 40, 255}, false)
	vector.StrokeRect(screen, float32(x), float32(y+10), float32(barW), float32(barH), 1, color.White, false)
	y += 30

	h.plusRects = map[string]image.Rectangle{}
	stats := []struct {
		Name  string
		Value int
	}{
		{"Strength", h.player.Stats.Strength},
		{"Intelligence", h.player.Stats.Intelligence},
		{"Vitality", h.player.Stats.Vitality},
		{"Dexterity", h.player.Stats.Dexterity},
	}

	for _, st := range stats {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: %d", st.Name, st.Value), x, y)
		if h.player.AttributePoints > 0 {
			px := x + 110
			py := y - 10
			vector.StrokeRect(screen, float32(px), float32(py), 12, 12, 1, color.White, false)
			ebitenutil.DebugPrintAt(screen, "+", px+3, py-2)
			h.plusRects[st.Name] = image.Rect(px, py, px+12, py+12)
		}
		y += 20
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Unspent Points: %d", h.player.AttributePoints), x, y)
}
