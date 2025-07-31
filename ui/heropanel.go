package ui

import (
	"dungeoneer/entities"
	"dungeoneer/progression"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// HeroPanel displays player stats and allows spending attribute points.
// It intentionally keeps layout simple for readability.
type HeroPanel struct {
	rect    image.Rectangle
	visible bool
	player  *entities.Player
	plus    map[string]image.Rectangle
	minus   map[string]image.Rectangle
}

// NewHeroPanel creates a hero panel.
func NewHeroPanel(r image.Rectangle, p *entities.Player) *HeroPanel {
	return &HeroPanel{
		rect:   r,
		player: p,
		plus:   make(map[string]image.Rectangle),
		minus:  make(map[string]image.Rectangle),
	}
}

func (hp *HeroPanel) SetRect(r image.Rectangle) { hp.rect = r }
func (hp *HeroPanel) Show()                     { hp.visible = true }
func (hp *HeroPanel) Hide()                     { hp.visible = false }
func (hp *HeroPanel) Toggle()                   { hp.visible = !hp.visible }
func (hp *HeroPanel) IsVisible() bool           { return hp.visible }

func (hp *HeroPanel) Update() {
	if !hp.visible || hp.player == nil {
		return
	}
	mx, my := ebiten.CursorPosition()
	if ebiten.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for attr, r := range hp.plus {
			if pointInRect(mx, my, r) && hp.player.UnspentPoints > 0 {
				switch attr {
				case "str":
					hp.player.Stats.Strength++
				case "int":
					hp.player.Stats.Intelligence++
				case "vit":
					hp.player.Stats.Vitality++
				case "dex":
					hp.player.Stats.Dexterity++
				}
				hp.player.UnspentPoints--
				hp.player.RecalculateStats()
				return
			}
		}
		for attr, r := range hp.minus {
			if pointInRect(mx, my, r) {
				switch attr {
				case "str":
					if hp.player.Stats.Strength > 1 {
						hp.player.Stats.Strength--
						hp.player.UnspentPoints++
					}
				case "int":
					if hp.player.Stats.Intelligence > 1 {
						hp.player.Stats.Intelligence--
						hp.player.UnspentPoints++
					}
				case "vit":
					if hp.player.Stats.Vitality > 1 {
						hp.player.Stats.Vitality--
						hp.player.UnspentPoints++
					}
				case "dex":
					if hp.player.Stats.Dexterity > 1 {
						hp.player.Stats.Dexterity--
						hp.player.UnspentPoints++
					}
				}
				hp.player.RecalculateStats()
				return
			}
		}
	}
}

func (hp *HeroPanel) Draw(screen *ebiten.Image) {
	if !hp.visible || hp.player == nil {
		return
	}

	DrawMenuOverlay(screen, DefaultOverlayColor)
	style := DefaultMenuStyles()
	DrawMenuWindow(screen, &style, float32(hp.rect.Min.X), float32(hp.rect.Min.Y), float32(hp.rect.Dx()), float32(hp.rect.Dy()))

	x := hp.rect.Min.X + 20
	y := hp.rect.Min.Y + 20

	// portrait
	if hp.player.Sprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(hp.player.Sprite, op)
	}
	infoY := y + 70
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Name: %s", hp.player.Name), x, infoY)
	infoY += 15
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d", hp.player.Level), x, infoY)
	infoY += 15
	expNeeded := progression.EXPToLevel(hp.player.Level)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("EXP: %d / %d", hp.player.EXP, expNeeded), x, infoY)

	statY := infoY + 20
	stats := []struct {
		key  string
		name string
		val  int
	}{
		{"str", "Strength", hp.player.Stats.Strength},
		{"int", "Intelligence", hp.player.Stats.Intelligence},
		{"vit", "Vitality", hp.player.Stats.Vitality},
		{"dex", "Dexterity", hp.player.Stats.Dexterity},
	}
	hp.plus = make(map[string]image.Rectangle)
	hp.minus = make(map[string]image.Rectangle)
	for i, s := range stats {
		lineY := statY + i*20
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: %d", s.name, s.val), x, lineY)
		pr := image.Rect(x+120, lineY-10, x+135, lineY+5)
		mr := image.Rect(x+138, lineY-10, x+153, lineY+5)
		hp.plus[s.key] = pr
		hp.minus[s.key] = mr
		if hp.player.UnspentPoints > 0 {
			ebitenutil.DebugPrintAt(screen, "[+]", pr.Min.X, pr.Min.Y)
		}
		if s.val > 1 {
			ebitenutil.DebugPrintAt(screen, "[-]", mr.Min.X, mr.Min.Y)
		}
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Unspent Points: %d", hp.player.UnspentPoints), x, statY+len(stats)*20+10)
}
