package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DevEntry is one row in the developer overlay.
//   - IsHeader=true rows are visual section dividers with no interactivity.
//   - Toggle=nil rows are informational only.
//   - IsActive=nil but Toggle!=nil rows are one-shot actions (show RUN button).
type DevEntry struct {
	Label    string
	Key      string       // display shortcut shown next to the label (informational)
	IsActive func() bool  // nil for headers and action-only rows
	Toggle   func()       // nil for headers
	IsHeader bool
}

// DevOverlay is the F12 developer tools panel. It aggregates all dev-only
// toggles and overlays in one place so they are never accidentally exposed
// through the normal player controls screen.
type DevOverlay struct {
	visible     bool
	rect        image.Rectangle
	entries     []DevEntry
	selectedIdx int
	style       MenuStyle
	scrollY     int // pixel scroll offset into the content area
	contentH    int // total content height (may exceed rect height)
	screenH     int // last known screen height for clamping
}

const (
	devTitleH  = 28
	devPadV    = 8
	devItemH   = 20
	devHeaderH = 22
	devPanelW  = 316
	devBadgeW  = 36
	devBadgeH  = 13
)

// NewDevOverlay creates a DevOverlay anchored to the top-right of the screen.
func NewDevOverlay(w, h int, entries []DevEntry) *DevOverlay {
	d := &DevOverlay{
		entries: entries,
		style:   DefaultMenuStyles(),
	}
	d.selectedIdx = d.firstSelectable()
	d.computeRect(w, h)
	return d
}

func (d *DevOverlay) computeRect(w, h int) {
	d.screenH = h
	d.contentH = devTitleH + devPadV
	for _, e := range d.entries {
		if e.IsHeader {
			d.contentH += devHeaderH
		} else {
			d.contentH += devItemH
		}
	}
	d.contentH += devPadV

	maxH := h - 20
	panelH := d.contentH
	if panelH > maxH {
		panelH = maxH
	}

	x := w - devPanelW - 10
	d.rect = image.Rect(x, 10, x+devPanelW, 10+panelH)
	d.clampScroll()
}

func (d *DevOverlay) Resize(w, h int)    { d.computeRect(w, h) }
func (d *DevOverlay) IsVisible() bool    { return d.visible }
func (d *DevOverlay) Show()              { d.visible = true }
func (d *DevOverlay) Hide()              { d.visible = false }
func (d *DevOverlay) Toggle()            { d.visible = !d.visible }

func (d *DevOverlay) clampScroll() {
	visibleH := d.rect.Dy()
	maxScroll := d.contentH - visibleH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.scrollY < 0 {
		d.scrollY = 0
	}
	if d.scrollY > maxScroll {
		d.scrollY = maxScroll
	}
}

func (d *DevOverlay) Update() {
	if !d.visible {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		d.Hide()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		d.selectedIdx = d.advance(d.selectedIdx, 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		d.selectedIdx = d.advance(d.selectedIdx, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		d.activate(d.selectedIdx)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		d.handleClick(mx, my)
	}
	// Mouse wheel scrolling when cursor is over the panel.
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		mx, my := ebiten.CursorPosition()
		if mx >= d.rect.Min.X && mx <= d.rect.Max.X && my >= d.rect.Min.Y && my <= d.rect.Max.Y {
			d.scrollY -= int(wheelY * 20)
			d.clampScroll()
		}
	}
}

func (d *DevOverlay) firstSelectable() int {
	for i, e := range d.entries {
		if !e.IsHeader && e.Toggle != nil {
			return i
		}
	}
	return 0
}

func (d *DevOverlay) advance(cur, dir int) int {
	n := len(d.entries)
	if n == 0 {
		return cur
	}
	idx := cur
	for range d.entries {
		idx = (idx + dir + n) % n
		if !d.entries[idx].IsHeader && d.entries[idx].Toggle != nil {
			return idx
		}
	}
	return cur
}

func (d *DevOverlay) activate(idx int) {
	if idx < 0 || idx >= len(d.entries) {
		return
	}
	if d.entries[idx].Toggle != nil {
		d.entries[idx].Toggle()
	}
}

func (d *DevOverlay) handleClick(mx, my int) {
	contentTop := d.rect.Min.Y + devTitleH + devPadV
	y := contentTop - d.scrollY
	for i, e := range d.entries {
		h := devItemH
		if e.IsHeader {
			h = devHeaderH
		}
		// Only register clicks within the visible content area.
		if my >= y && my < y+h && mx >= d.rect.Min.X && mx <= d.rect.Max.X &&
			my >= contentTop && my <= d.rect.Max.Y {
			if !e.IsHeader && e.Toggle != nil {
				d.selectedIdx = i
				e.Toggle()
			}
			return
		}
		y += h
	}
}

var (
	devOnColor     = color.RGBA{40, 170, 70, 230}
	devOffColor    = color.RGBA{65, 65, 65, 210}
	devActionColor = color.RGBA{55, 85, 130, 220}
	devSelColor    = color.RGBA{75, 75, 115, 160}
	devDivColor    = color.RGBA{90, 90, 110, 160}
)

func (d *DevOverlay) Draw(screen *ebiten.Image) {
	if !d.visible {
		return
	}

	DrawMenuWindow(screen, &d.style,
		float32(d.rect.Min.X), float32(d.rect.Min.Y),
		float32(d.rect.Dx()), float32(d.rect.Dy()))

	lx := d.rect.Min.X + 10
	y := d.rect.Min.Y + 8

	ebitenutil.DebugPrintAt(screen, "DEV TOOLS              [F12]", lx, y)
	y += devTitleH

	// Content area starts below the title; scroll offset applied here.
	contentTop := y + devPadV
	y = contentTop - d.scrollY

	// Draw a scroll indicator if content overflows.
	if d.contentH > d.rect.Dy() {
		scrollFrac := float64(d.scrollY) / float64(d.contentH-d.rect.Dy())
		trackH := float32(d.rect.Dy() - devTitleH - devPadV*2)
		thumbH := float32(4)
		thumbY := float32(contentTop) + float32(scrollFrac)*float32(trackH-thumbH)
		vector.DrawFilledRect(screen,
			float32(d.rect.Max.X-4), float32(contentTop),
			3, trackH, color.RGBA{60, 60, 80, 180}, false)
		vector.DrawFilledRect(screen,
			float32(d.rect.Max.X-4), thumbY,
			3, thumbH, color.RGBA{160, 160, 200, 220}, false)
	}

	for i, e := range d.entries {
		h := devItemH
		if e.IsHeader {
			h = devHeaderH
		}
		// Skip entries completely outside the visible content area.
		if y+h <= contentTop || y >= d.rect.Max.Y {
			y += h
			continue
		}

		if e.IsHeader {
			// Horizontal rule + section label
			ry := float32(y + devHeaderH/2)
			vector.StrokeLine(screen,
				float32(d.rect.Min.X+6), ry,
				float32(d.rect.Max.X-6), ry,
				1, devDivColor, false)
			ebitenutil.DebugPrintAt(screen, e.Label, lx, y+3)
			y += devHeaderH
			continue
		}

		// Row selection highlight
		if i == d.selectedIdx && e.Toggle != nil {
			vector.DrawFilledRect(screen,
				float32(d.rect.Min.X+3), float32(y),
				float32(d.rect.Dx()-6), float32(devItemH-2),
				devSelColor, false)
		}

		// Label (with optional key badge prefix)
		label := e.Label
		if e.Key != "" {
			label = "[" + e.Key + "] " + label
		}
		ebitenutil.DebugPrintAt(screen, label, lx+4, y+3)

		// State badge on the right edge
		bx := float32(d.rect.Max.X - devBadgeW - 6)
		by := float32(y + 3)
		if e.IsActive != nil {
			if e.IsActive() {
				vector.DrawFilledRect(screen, bx, by, devBadgeW, devBadgeH, devOnColor, false)
				ebitenutil.DebugPrintAt(screen, " ON", int(bx)+3, int(by))
			} else {
				vector.DrawFilledRect(screen, bx, by, devBadgeW, devBadgeH, devOffColor, false)
				ebitenutil.DebugPrintAt(screen, "OFF", int(bx)+3, int(by))
			}
		} else if e.Toggle != nil {
			vector.DrawFilledRect(screen, bx, by, devBadgeW, devBadgeH, devActionColor, false)
			ebitenutil.DebugPrintAt(screen, "RUN", int(bx)+3, int(by))
		}

		y += devItemH
	}
}
