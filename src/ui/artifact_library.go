package ui

import (
	"dungeoneer/items"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// domainColor returns the badge color for an artifact domain.
func domainColor(domain string) color.RGBA {
	switch domain {
	case "iron":
		return color.RGBA{180, 160, 120, 255}
	case "shadow":
		return color.RGBA{100, 80, 180, 255}
	case "flame":
		return color.RGBA{220, 80, 40, 255}
	case "void":
		return color.RGBA{60, 20, 100, 255}
	case "nature":
		return color.RGBA{60, 160, 80, 255}
	case "arcane":
		return color.RGBA{80, 140, 255, 255}
	default:
		return color.RGBA{120, 120, 120, 255}
	}
}

var artifactDomains = []string{"all", "iron", "shadow", "flame", "void", "nature", "arcane"}
var artifactDomainLabels = []string{"1:All", "2:Iron", "3:Shadow", "4:Flame", "5:Void", "6:Nature", "7:Arcane"}

// artifactEntry is an internal display model for one artifact row.
type artifactEntry struct {
	tmpl    *items.ItemTemplate
	owned   bool
}

// ArtifactLibrary displays all artifacts the player owns, with domain filtering.
type ArtifactLibrary struct {
	Visible      bool
	Collection   []string // artifact IDs from MetaSave.ArtifactCollection
	SelectedIdx  int
	FilterDomain string // "" / "all" = all, "shadow", "flame", etc.
	ScrollOffset int

	screenW  int
	screenH  int
	entries  []artifactEntry // filtered+sorted view
	owned    map[string]bool // fast lookup
}

// NewArtifactLibrary creates a closed library.
func NewArtifactLibrary(w, h int) *ArtifactLibrary {
	return &ArtifactLibrary{
		screenW:      w,
		screenH:      h,
		FilterDomain: "all",
	}
}

// Resize updates screen dimensions.
func (al *ArtifactLibrary) Resize(w, h int) { al.screenW = w; al.screenH = h }

// Open populates Collection from MetaSave and makes the library visible.
func (al *ArtifactLibrary) Open(collection []string) {
	al.Collection = collection
	al.Visible = true
	al.SelectedIdx = 0
	al.ScrollOffset = 0
	al.rebuildEntries()
}

// Close hides the library.
func (al *ArtifactLibrary) Close() { al.Visible = false }

// rebuildEntries refreshes the filtered+sorted entry list.
func (al *ArtifactLibrary) rebuildEntries() {
	al.owned = make(map[string]bool, len(al.Collection))
	for _, id := range al.Collection {
		al.owned[id] = true
	}

	al.entries = al.entries[:0]
	for _, tmpl := range items.Registry {
		if !tmpl.IsArtifact {
			continue
		}
		if al.FilterDomain != "all" && al.FilterDomain != "" && tmpl.ArtifactDomain != al.FilterDomain {
			continue
		}
		al.entries = append(al.entries, artifactEntry{
			tmpl:  tmpl,
			owned: al.owned[tmpl.ID],
		})
	}

	// Sort: owned first, then alphabetically by name.
	sort.Slice(al.entries, func(i, j int) bool {
		oi, oj := al.entries[i].owned, al.entries[j].owned
		if oi != oj {
			return oi
		}
		return al.entries[i].tmpl.Name < al.entries[j].tmpl.Name
	})

	if al.SelectedIdx >= len(al.entries) {
		al.SelectedIdx = 0
	}
}

// Update handles mouse/keyboard input. Returns true if input was consumed.
func (al *ArtifactLibrary) Update() bool {
	if !al.Visible {
		return false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		al.Close()
		return true
	}

	// Domain filter keys 1-7.
	for i, dom := range artifactDomains {
		key := ebiten.Key(int(ebiten.Key1) + i)
		if inpututil.IsKeyJustPressed(key) {
			al.FilterDomain = dom
			al.ScrollOffset = 0
			al.SelectedIdx = 0
			al.rebuildEntries()
			return true
		}
	}

	// Arrow navigation.
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if al.SelectedIdx < len(al.entries)-1 {
			al.SelectedIdx++
			al.clampScroll()
		}
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if al.SelectedIdx > 0 {
			al.SelectedIdx--
			al.clampScroll()
		}
		return true
	}

	// Mouse click on list rows.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		panelX, panelY, panelW, _ := al.panelBounds()
		listX := panelX + 8
		listY := panelY + 70 // below title + tabs
		rowH := 22
		col := 2 // two columns
		colW := (panelW/2 - 16)

		// Check click on list entries (left half panel).
		for i, _ := range al.entries {
			vis := i - al.ScrollOffset
			if vis < 0 || vis >= al.maxVisible() {
				continue
			}
			col0 := vis % col
			row0 := vis / col
			ex := listX + col0*colW
			ey := listY + row0*rowH
			if mx >= ex && mx < ex+colW && my >= ey && my < ey+rowH {
				al.SelectedIdx = i
				return true
			}
		}

		// Click outside panel closes.
		if mx < panelX || mx > panelX+panelW {
			al.Close()
			return true
		}
		return true
	}

	return true // consume all input while open
}

func (al *ArtifactLibrary) clampScroll() {
	const rowsVisible = 6 // 6 rows × 2 columns = 12 entries
	vis := al.SelectedIdx / 2 // which row the selection is in
	if vis < al.ScrollOffset {
		al.ScrollOffset = vis
	}
	if vis >= al.ScrollOffset+rowsVisible {
		al.ScrollOffset = vis - rowsVisible + 1
	}
}

func (al *ArtifactLibrary) maxVisible() int { return 12 } // 6 rows × 2 cols

// panelBounds returns the panel's top-left x, y and w, h.
func (al *ArtifactLibrary) panelBounds() (int, int, int, int) {
	w, h := 600, 400
	x := (al.screenW - w) / 2
	y := (al.screenH - h) / 2
	return x, y, w, h
}

// Draw renders the library panel.
func (al *ArtifactLibrary) Draw(screen *ebiten.Image) {
	if !al.Visible {
		return
	}

	// Dim background.
	vector.DrawFilledRect(screen, 0, 0, float32(al.screenW), float32(al.screenH), color.RGBA{0, 0, 0, 160}, false)

	px, py, pw, ph := al.panelBounds()
	fpx, fpy, fpw, fph := float32(px), float32(py), float32(pw), float32(ph)

	// Panel background.
	vector.DrawFilledRect(screen, fpx, fpy, fpw, fph, color.NRGBA{10, 10, 20, 230}, false)
	vector.StrokeRect(screen, fpx, fpy, fpw, fph, 2, color.RGBA{100, 80, 60, 255}, false)

	// Title.
	ebitenutil.DebugPrintAt(screen, "ARTIFACT LIBRARY  [Esc] Close", px+pw/2-110, py+8)

	// Domain filter tabs.
	tabY := py + 26
	tabW := pw / len(artifactDomainLabels)
	for i, lbl := range artifactDomainLabels {
		tx := px + i*tabW
		active := artifactDomains[i] == al.FilterDomain
		if active {
			vector.DrawFilledRect(screen, float32(tx), float32(tabY-2), float32(tabW-2), 18, color.RGBA{80, 60, 30, 200}, false)
			vector.StrokeRect(screen, float32(tx), float32(tabY-2), float32(tabW-2), 18, 1, color.RGBA{200, 160, 60, 255}, false)
		}
		ebitenutil.DebugPrintAt(screen, lbl, tx+2, tabY)
	}

	// Divider.
	vector.DrawFilledRect(screen, fpx+4, float32(py+48), fpw-8, 1, color.RGBA{80, 80, 80, 200}, false)

	// Left column: entry list (half panel width).
	listX := px + 8
	listY := py + 56
	rowH := 22
	cols := 2
	colW := (pw/2 - 12)

	for i, e := range al.entries {
		vis := i - al.ScrollOffset
		if vis < 0 || vis >= al.maxVisible() {
			continue
		}
		col := vis % cols
		row := vis / cols
		ex := listX + col*colW
		ey := listY + row*rowH

		// Selection highlight.
		if i == al.SelectedIdx {
			vector.DrawFilledRect(screen, float32(ex-2), float32(ey-1), float32(colW-2), float32(rowH-2), color.RGBA{60, 50, 80, 200}, false)
		}

		// Elite gold border on the entry.
		if e.tmpl.IsElite {
			vector.StrokeRect(screen, float32(ex-2), float32(ey-1), float32(colW-2), float32(rowH-2), 1, color.RGBA{255, 200, 60, 180}, false)
		}

		nameClr := color.RGBA{220, 210, 200, 255}
		if !e.owned {
			nameClr = color.RGBA{80, 80, 80, 180}
		}

		// Domain badge (small colored square).
		dc := domainColor(e.tmpl.ArtifactDomain)
		if !e.owned {
			dc = color.RGBA{40, 40, 40, 140}
		}
		vector.DrawFilledRect(screen, float32(ex), float32(ey+4), 8, 8, dc, false)

		label := e.tmpl.Name
		if !e.owned {
			label = "???"
		}
		ebitenutil.DebugPrintAt(screen, label, ex+12, ey+2)
		_ = nameClr // color applied via badge only for now (DebugPrintAt is always white)
	}

	// Vertical divider between list and detail.
	midX := px + pw/2
	vector.DrawFilledRect(screen, float32(midX), float32(py+48), 1, float32(ph-56), color.RGBA{80, 80, 80, 200}, false)

	// Right column: detail panel.
	al.drawDetail(screen, midX+8, py+56, pw/2-16, ph-68)

	// Scroll indicator.
	if len(al.entries) > al.maxVisible() {
		total := (len(al.entries) + 1) / cols
		visible := al.maxVisible() / cols
		if total > visible {
			barH := float32(ph-68) * float32(visible) / float32(total)
			barY := float32(py+56) + float32(ph-68-int(barH))*float32(al.ScrollOffset)/float32(total-visible)
			vector.DrawFilledRect(screen, float32(px+pw-6), float32(py+56), 4, float32(ph-68), color.RGBA{40, 40, 40, 200}, false)
			vector.DrawFilledRect(screen, float32(px+pw-6), barY, 4, barH, color.RGBA{160, 140, 100, 255}, false)
		}
	}

	// Footer hint.
	ebitenutil.DebugPrintAt(screen, "[Up/Down] Navigate  [1-7] Filter Domain", px+8, py+ph-14)
}

// drawDetail renders the selected artifact's full detail.
func (al *ArtifactLibrary) drawDetail(screen *ebiten.Image, x, y, w, h int) {
	if len(al.entries) == 0 {
		ebitenutil.DebugPrintAt(screen, "(no artifacts)", x, y+10)
		return
	}
	if al.SelectedIdx < 0 || al.SelectedIdx >= len(al.entries) {
		return
	}
	e := al.entries[al.SelectedIdx]
	tmpl := e.tmpl

	if !e.owned {
		ebitenutil.DebugPrintAt(screen, "???", x, y)
		ebitenutil.DebugPrintAt(screen, "Not yet discovered.", x, y+16)
		return
	}

	// Name.
	ebitenutil.DebugPrintAt(screen, tmpl.Name, x, y)
	y += 16

	// Elite badge.
	if tmpl.IsElite {
		ebitenutil.DebugPrintAt(screen, "[ELITE]", x, y)
		y += 14
	}

	// Domain badge.
	dc := domainColor(tmpl.ArtifactDomain)
	vector.DrawFilledRect(screen, float32(x), float32(y+2), 8, 8, dc, false)
	domainName := tmpl.ArtifactDomain
	if domainName == "" {
		domainName = "unknown"
	}
	ebitenutil.DebugPrintAt(screen, domainName, x+12, y)
	y += 16

	// Divider.
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 1, color.RGBA{80, 80, 80, 200}, false)
	y += 6

	// Grants ability.
	if tmpl.GrantsAbility != "" {
		ebitenutil.DebugPrintAt(screen, "Ability: "+tmpl.GrantsAbility, x, y)
		y += 14
	}

	// Description.
	if tmpl.Description != "" {
		for _, line := range WrapText(tmpl.Description, w/7) {
			ebitenutil.DebugPrintAt(screen, line, x, y)
			y += 13
		}
		y += 4
	}

	// Effect one-liner.
	if tmpl.Effect != nil {
		ef := tmpl.Effect
		var effLine string
		if ef.MagnitudePct != 0 {
			effLine = formatEffectType(ef.Type) + ": " + itoa(ef.MagnitudePct) + "%"
		} else if ef.MagnitudeFlat != 0 {
			effLine = formatEffectType(ef.Type) + ": +" + itoa(ef.MagnitudeFlat)
		} else {
			effLine = formatEffectType(ef.Type)
		}
		if ef.Trigger != "" && ef.Trigger != "passive" {
			effLine = formatTrigger(ef.Trigger) + ": " + effLine
		}
		ebitenutil.DebugPrintAt(screen, effLine, x, y)
		y += 14
	}

	// Flavor text.
	if tmpl.FlavorText != "" {
		y += 4
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 1, color.RGBA{60, 60, 60, 180}, false)
		y += 6
		for _, line := range WrapText(tmpl.FlavorText, w/7) {
			ebitenutil.DebugPrintAt(screen, line, x, y)
			y += 13
		}
	}
}

// itoa is a tiny helper to avoid importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
