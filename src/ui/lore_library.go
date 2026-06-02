package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// LoreEntry is the display model for a lore entry — passed from game layer.
type LoreEntry struct {
	ID       string
	Title    string
	Category string
	Body     string
	Unlocked bool
}

var loreCategories = []string{"character", "cosmology", "history", "fragment"}
var loreCategoryLabels = []string{"1:Character", "2:Cosmology", "3:History", "4:Fragment"}

// LoreLibrary is a full-screen lore reading UI with category tabs and scroll.
type LoreLibrary struct {
	Active         bool
	Entries        []LoreEntry
	ActiveCategory string
	ScrollOffset   int
	screenW        int
	screenH        int
}

// NewLoreLibrary creates a closed lore library sized to the given screen dimensions.
func NewLoreLibrary(w, h int) *LoreLibrary {
	return &LoreLibrary{
		screenW:        w,
		screenH:        h,
		ActiveCategory: "character",
	}
}

// Open shows the library with the given entries.
func (l *LoreLibrary) Open(entries []LoreEntry) {
	l.Entries = entries
	l.Active = true
	l.ScrollOffset = 0
}

// Close hides the library.
func (l *LoreLibrary) Close() {
	l.Active = false
}

// Resize updates screen dimensions (call on window resize).
func (l *LoreLibrary) Resize(w, h int) {
	l.screenW = w
	l.screenH = h
}

// Update handles input. Call only when library is active.
func (l *LoreLibrary) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		l.Active = false
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		l.ScrollOffset++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		if l.ScrollOffset > 0 {
			l.ScrollOffset--
		}
	}
	for i, cat := range loreCategories {
		key := ebiten.Key(int(ebiten.Key1) + i)
		if inpututil.IsKeyJustPressed(key) {
			l.ActiveCategory = cat
			l.ScrollOffset = 0
		}
	}
}

// Draw renders the lore library UI.
func (l *LoreLibrary) Draw(screen *ebiten.Image) {
	if !l.Active {
		return
	}
	w, h := l.screenW, l.screenH

	// Full-screen dark background.
	bg := ebiten.NewImage(w, h)
	bg.Fill(color.NRGBA{10, 8, 15, 235})
	screen.DrawImage(bg, &ebiten.DrawImageOptions{})

	ebitenutil.DebugPrintAt(screen, "LORE LIBRARY  [Esc] Close", w/2-70, 12)
	ebitenutil.DebugPrintAt(screen, "[1] Character  [2] Cosmology  [3] History  [4] Fragment  [Up/Down] Scroll", 20, 28)

	// Category tabs.
	for i, cat := range loreCategories {
		tx := 20 + i*132
		ty := 50
		if cat == l.ActiveCategory {
			hl := ebiten.NewImage(124, 16)
			hl.Fill(color.NRGBA{100, 80, 40, 180})
			hlOp := &ebiten.DrawImageOptions{}
			hlOp.GeoM.Translate(float64(tx-4), float64(ty-2))
			screen.DrawImage(hl, hlOp)
		}
		ebitenutil.DebugPrintAt(screen, loreCategoryLabels[i], tx, ty)
	}

	// Entry list.
	const lineH = 64
	y := 74
	shown := 0
	for _, entry := range l.Entries {
		if entry.Category != l.ActiveCategory {
			continue
		}
		shown++
		if shown-1 < l.ScrollOffset {
			continue
		}
		if y > h-20 {
			break
		}

		entryBg := ebiten.NewImage(w-40, lineH-4)
		entryBg.Fill(color.NRGBA{30, 25, 40, 200})
		ebOp := &ebiten.DrawImageOptions{}
		ebOp.GeoM.Translate(20, float64(y))
		screen.DrawImage(entryBg, ebOp)

		if entry.Unlocked {
			ebitenutil.DebugPrintAt(screen, entry.Title, 28, y+4)
			// Render body text with wrapping.
			bodyY := y + 18
			body := entry.Body
			for len(body) > 0 && bodyY < y+lineH-4 {
				cut := 100
				if cut > len(body) {
					cut = len(body)
				}
				if cut < len(body) {
					if idx := strings.LastIndex(body[:cut], " "); idx > 0 {
						cut = idx
					}
				}
				ebitenutil.DebugPrintAt(screen, body[:cut], 28, bodyY)
				body = strings.TrimSpace(body[cut:])
				bodyY += 12
			}
		} else {
			ebitenutil.DebugPrintAt(screen, "???", 28, y+4)
			ebitenutil.DebugPrintAt(screen, "Unlock through exploration or NPC dialogue.", 28, y+18)
		}
		y += lineH
	}

	if shown == 0 {
		ebitenutil.DebugPrintAt(screen, "(no entries in this category)", w/2-100, h/2)
	}
}
