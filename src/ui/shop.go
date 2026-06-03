package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ShopEntry is the display model for one shop item.
type ShopEntry struct {
	ItemID    string
	Name      string
	Cost      int
	Purchased int
	StockMax  int
	CanAfford bool
}

// Shop is the hub shop UI.
type Shop struct {
	Active      bool
	Entries     []ShopEntry
	SelectedIdx int
	Remnants    int // player's current balance
	screenW     int
	screenH     int
	OnPurchase  func(itemID string, cost int) bool // returns true if purchase succeeded
}

// NewShop creates a closed shop sized to the given screen dimensions.
func NewShop(w, h int) *Shop {
	return &Shop{screenW: w, screenH: h}
}

// Open shows the shop with the given entries and current remnant balance.
func (s *Shop) Open(entries []ShopEntry, remnants int) {
	s.Entries = entries
	s.Remnants = remnants
	s.Active = true
	s.SelectedIdx = 0
}

// Close hides the shop.
func (s *Shop) Close() { s.Active = false }

// Resize updates screen dimensions (call on window resize).
func (s *Shop) Resize(w, h int) { s.screenW = w; s.screenH = h }

// Update handles input. Returns early if shop is not active.
func (s *Shop) Update() {
	if !s.Active {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.Active = false
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && s.SelectedIdx < len(s.Entries)-1 {
		s.SelectedIdx++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && s.SelectedIdx > 0 {
		s.SelectedIdx--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) && len(s.Entries) > 0 {
		entry := &s.Entries[s.SelectedIdx]
		if entry.CanAfford && (entry.StockMax == 0 || entry.Purchased < entry.StockMax) {
			if s.OnPurchase != nil && s.OnPurchase(entry.ItemID, entry.Cost) {
				entry.Purchased++
				s.Remnants -= entry.Cost
				// Refresh affordability for all entries.
				for i := range s.Entries {
					s.Entries[i].CanAfford = s.Remnants >= s.Entries[i].Cost
				}
			}
		}
	}
}

// Draw renders the shop UI.
func (s *Shop) Draw(screen *ebiten.Image) {
	if !s.Active {
		return
	}
	w, h := s.screenW, s.screenH

	bg := ebiten.NewImage(w, h)
	bg.Fill(color.NRGBA{5, 10, 20, 240})
	screen.DrawImage(bg, &ebiten.DrawImageOptions{})

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HUB SHOP   Remnants: %d   [E] Buy  [Esc] Close", s.Remnants), 20, 12)

	if len(s.Entries) == 0 {
		ebitenutil.DebugPrintAt(screen, "Shop is empty.", w/2-50, h/2)
		return
	}

	y := 32
	for i, entry := range s.Entries {
		rowH := 36
		rowBg := ebiten.NewImage(w-40, rowH-2)
		if i == s.SelectedIdx {
			rowBg.Fill(color.NRGBA{60, 90, 40, 200})
		} else {
			rowBg.Fill(color.NRGBA{20, 30, 20, 180})
		}
		rop := &ebiten.DrawImageOptions{}
		rop.GeoM.Translate(20, float64(y))
		screen.DrawImage(rowBg, rop)

		stockStr := ""
		if entry.StockMax > 0 {
			stockStr = fmt.Sprintf(" [%d/%d]", entry.Purchased, entry.StockMax)
		}
		affordStr := ""
		if !entry.CanAfford {
			affordStr = " (can't afford)"
		}
		label := fmt.Sprintf("%s — %d Remnants%s%s", entry.Name, entry.Cost, stockStr, affordStr)
		ebitenutil.DebugPrintAt(screen, label, 28, y+4)

		if i == s.SelectedIdx && entry.CanAfford && (entry.StockMax == 0 || entry.Purchased < entry.StockMax) {
			ebitenutil.DebugPrintAt(screen, "[E] Purchase", 28, y+18)
		}
		y += rowH
		if y > h-20 {
			break
		}
	}
}
