package ui

import (
	"dungeoneer/items"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ArtifactLoadout is the pre-run screen where the player selects which
// artifacts to bring. Slots 0-5 are regular, slot 6 is the elite slot.
type ArtifactLoadout struct {
	Visible    bool
	Loadout    [7]string       // current loadout (artifact IDs, "" = empty)
	Collection []string        // all owned artifact IDs
	ActiveSlot int             // which slot is being filled (-1 = none)
	PickerOpen bool            // true when slot picker browser is open
	PickerIdx  int             // selected index in picker list
	OnConfirm  func([7]string) // called when player clicks "Begin Run"
	OnCancel   func()          // called when player cancels

	screenW       int
	screenH       int
	pickerEntries []pickerEntry // filtered list shown in picker
	pickerScroll  int
}

// pickerEntry is one artifact shown in the picker list.
type pickerEntry struct {
	tmpl *items.ItemTemplate
}

// NewArtifactLoadout creates a closed loadout selector.
func NewArtifactLoadout(w, h int) *ArtifactLoadout {
	return &ArtifactLoadout{
		screenW:    w,
		screenH:    h,
		ActiveSlot: -1,
	}
}

// Resize updates screen dimensions.
func (al *ArtifactLoadout) Resize(w, h int) { al.screenW = w; al.screenH = h }

// Open prepopulates Loadout from the saved loadout and Collection from MetaSave.
func (al *ArtifactLoadout) Open(savedLoadout [7]string, collection []string) {
	al.Loadout = savedLoadout
	al.Collection = collection
	al.Visible = true
	al.ActiveSlot = -1
	al.PickerOpen = false
	al.PickerIdx = 0
	al.pickerScroll = 0
}

// Close hides the loadout screen without confirming.
func (al *ArtifactLoadout) Close() {
	al.Visible = false
	al.PickerOpen = false
	al.ActiveSlot = -1
}

// openPicker populates pickerEntries for the given slot and shows the picker.
func (al *ArtifactLoadout) openPicker(slot int) {
	al.ActiveSlot = slot
	al.PickerOpen = true
	al.PickerIdx = 0
	al.pickerScroll = 0

	eliteOnly := slot == 6
	al.pickerEntries = al.pickerEntries[:0]

	// Build a fast set of already-used IDs (to prevent double-equipping).
	used := map[string]bool{}
	for i, id := range al.Loadout {
		if id != "" && i != slot {
			used[id] = true
		}
	}

	for _, id := range al.Collection {
		tmpl, ok := items.Registry[id]
		if !ok || !tmpl.IsArtifact {
			continue
		}
		if used[id] {
			continue
		}
		if eliteOnly && !tmpl.IsElite {
			continue
		}
		if !eliteOnly && tmpl.IsElite {
			continue
		}
		al.pickerEntries = append(al.pickerEntries, pickerEntry{tmpl: tmpl})
	}
}

// Update handles input. Returns true if consumed.
func (al *ArtifactLoadout) Update() bool {
	if !al.Visible {
		return false
	}

	mx, my := ebiten.CursorPosition()

	// --- Picker mode ---
	if al.PickerOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			al.PickerOpen = false
			al.ActiveSlot = -1
			return true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			if al.PickerIdx < len(al.pickerEntries)-1 {
				al.PickerIdx++
				al.clampPickerScroll()
			}
			return true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			if al.PickerIdx > 0 {
				al.PickerIdx--
				al.clampPickerScroll()
			}
			return true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			al.confirmPickerSelection()
			return true
		}

		// Mouse click in picker.
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			px, py, pw, _ := al.pickerBounds()
			listY := py + 36
			rowH := 22
			if mx >= px && mx <= px+pw {
				relY := my - listY
				if relY >= 0 {
					row := relY / rowH
					idx := al.pickerScroll + row
					if idx >= 0 && idx < len(al.pickerEntries) {
						al.PickerIdx = idx
						al.confirmPickerSelection()
						return true
					}
				}
			} else {
				// Click outside picker closes it.
				al.PickerOpen = false
				al.ActiveSlot = -1
			}
			return true
		}
		return true
	}

	// --- Main loadout mode ---
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		al.Visible = false
		if al.OnCancel != nil {
			al.OnCancel()
		}
		return true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// "Begin Run" button.
		bx, by, bw, bh := al.beginRunButtonBounds()
		if mx >= bx && mx <= bx+bw && my >= by && my <= by+bh {
			if al.OnConfirm != nil {
				al.OnConfirm(al.Loadout)
			}
			al.Visible = false
			return true
		}

		// "Back" button.
		cx, cy, cw, ch := al.backButtonBounds()
		if mx >= cx && mx <= cx+cw && my >= cy && my <= cy+ch {
			al.Visible = false
			if al.OnCancel != nil {
				al.OnCancel()
			}
			return true
		}

		// Slot clicks.
		for i := 0; i < 7; i++ {
			sx, sy, sw, sh := al.slotBounds(i)
			if mx >= sx && mx <= sx+sw && my >= sy && my <= sy+sh {
				al.openPicker(i)
				return true
			}
		}
	}

	// Right-click a slot to clear it.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		for i := 0; i < 7; i++ {
			sx, sy, sw, sh := al.slotBounds(i)
			if mx >= sx && mx <= sx+sw && my >= sy && my <= sy+sh {
				al.Loadout[i] = ""
				return true
			}
		}
	}

	return true
}

func (al *ArtifactLoadout) confirmPickerSelection() {
	if al.PickerIdx >= 0 && al.PickerIdx < len(al.pickerEntries) {
		al.Loadout[al.ActiveSlot] = al.pickerEntries[al.PickerIdx].tmpl.ID
	}
	al.PickerOpen = false
	al.ActiveSlot = -1
}

func (al *ArtifactLoadout) clampPickerScroll() {
	const maxRows = 10
	if al.PickerIdx < al.pickerScroll {
		al.pickerScroll = al.PickerIdx
	}
	if al.PickerIdx >= al.pickerScroll+maxRows {
		al.pickerScroll = al.PickerIdx - maxRows + 1
	}
}

// slotBounds returns the bounding box for slot i (0-5 = regular grid, 6 = elite).
func (al *ArtifactLoadout) slotBounds(slot int) (x, y, w, h int) {
	const slotW, slotH = 130, 80
	const eliteW, eliteH = 280, 80

	gridStartX := (al.screenW - (slotW*3 + 20)) / 2
	gridStartY := al.screenH/2 - 110

	if slot < 6 {
		col := slot % 3
		row := slot / 3
		return gridStartX + col*(slotW+10), gridStartY + row*(slotH+10), slotW, slotH
	}
	// Elite slot: centered below the 2×3 grid.
	eliteX := (al.screenW - eliteW) / 2
	eliteY := gridStartY + 2*(slotH+10) + 20
	return eliteX, eliteY, eliteW, eliteH
}

func (al *ArtifactLoadout) beginRunButtonBounds() (x, y, w, h int) {
	return al.screenW - 160, al.screenH - 50, 140, 34
}

func (al *ArtifactLoadout) backButtonBounds() (x, y, w, h int) {
	return 20, al.screenH - 50, 100, 34
}

// pickerBounds returns the bounding box of the picker overlay.
func (al *ArtifactLoadout) pickerBounds() (x, y, w, h int) {
	w, h = 320, 280
	x = (al.screenW - w) / 2
	y = (al.screenH - h) / 2
	return
}

// Draw renders the loadout screen.
func (al *ArtifactLoadout) Draw(screen *ebiten.Image) {
	if !al.Visible {
		return
	}

	// Full-screen dark overlay.
	vector.DrawFilledRect(screen, 0, 0, float32(al.screenW), float32(al.screenH), color.NRGBA{5, 5, 10, 220}, false)

	// Title.
	title := "PREPARE YOUR ARSENAL"
	ebitenutil.DebugPrintAt(screen, title, al.screenW/2-len(title)*4, 20)

	// Subtitle hint.
	ebitenutil.DebugPrintAt(screen, "Click a slot to assign an artifact.  Right-click to clear.", al.screenW/2-180, 38)

	// Draw 6 regular slots.
	for i := 0; i < 6; i++ {
		al.drawSlot(screen, i)
	}

	// Draw elite slot.
	al.drawEliteSlot(screen)

	// "Begin Run" button.
	bx, by, bw, bh := al.beginRunButtonBounds()
	vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{30, 120, 50, 255}, false)
	vector.StrokeRect(screen, float32(bx), float32(by), float32(bw), float32(bh), 2, color.RGBA{80, 220, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Begin Run", bx+bw/2-36, by+10)

	// "Back" button.
	cx, cy, cw, ch := al.backButtonBounds()
	vector.DrawFilledRect(screen, float32(cx), float32(cy), float32(cw), float32(ch), color.RGBA{50, 50, 50, 255}, false)
	vector.StrokeRect(screen, float32(cx), float32(cy), float32(cw), float32(ch), 2, color.RGBA{120, 120, 120, 255}, false)
	ebitenutil.DebugPrintAt(screen, "Back", cx+cw/2-14, cy+10)

	// Picker overlay (drawn on top).
	if al.PickerOpen {
		al.drawPicker(screen)
	}
}

func (al *ArtifactLoadout) drawSlot(screen *ebiten.Image, slot int) {
	sx, sy, sw, sh := al.slotBounds(slot)
	id := al.Loadout[slot]

	bgClr := color.RGBA{20, 18, 30, 220}
	borderClr := color.RGBA{80, 80, 100, 200}
	if al.ActiveSlot == slot {
		borderClr = color.RGBA{220, 200, 80, 255}
	}

	vector.DrawFilledRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), bgClr, false)
	vector.StrokeRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), 2, borderClr, false)

	if id == "" {
		ebitenutil.DebugPrintAt(screen, itoa(slot+1), sx+4, sy+4)
		ebitenutil.DebugPrintAt(screen, "Empty", sx+sw/2-18, sy+sh/2-6)
		return
	}

	tmpl, ok := items.Registry[id]
	if !ok {
		ebitenutil.DebugPrintAt(screen, "Unknown", sx+4, sy+4)
		return
	}

	// Icon (if available).
	if tmpl.Icon != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx+4), float64(sy+4))
		screen.DrawImage(tmpl.Icon, op)
	}

	// Name.
	ebitenutil.DebugPrintAt(screen, tmpl.Name, sx+4, sy+4)

	// Domain badge.
	dc := domainColor(tmpl.ArtifactDomain)
	vector.DrawFilledRect(screen, float32(sx+4), float32(sy+20), 8, 8, dc, false)
	ebitenutil.DebugPrintAt(screen, tmpl.ArtifactDomain, sx+16, sy+18)

	// Effect one-liner.
	if tmpl.Effect != nil {
		ef := tmpl.Effect
		line := formatTrigger(ef.Trigger) + ": " + formatEffectType(ef.Type)
		ebitenutil.DebugPrintAt(screen, line, sx+4, sy+sh-16)
	} else if tmpl.GrantsAbility != "" {
		ebitenutil.DebugPrintAt(screen, "Ability: "+tmpl.GrantsAbility, sx+4, sy+sh-16)
	}
}

func (al *ArtifactLoadout) drawEliteSlot(screen *ebiten.Image) {
	sx, sy, sw, sh := al.slotBounds(6)
	id := al.Loadout[6]

	bgClr := color.RGBA{20, 16, 10, 220}
	borderClr := color.RGBA{180, 140, 40, 220} // gold border for elite
	if al.ActiveSlot == 6 {
		borderClr = color.RGBA{255, 220, 80, 255}
	}

	vector.DrawFilledRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), bgClr, false)
	vector.StrokeRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), 2, borderClr, false)

	// "Elite Artifact" label.
	ebitenutil.DebugPrintAt(screen, "ELITE ARTIFACT", sx+4, sy+4)

	if id == "" {
		ebitenutil.DebugPrintAt(screen, "Empty (Elite only)", sx+sw/2-60, sy+sh/2-6)
		return
	}

	tmpl, ok := items.Registry[id]
	if !ok {
		ebitenutil.DebugPrintAt(screen, "Unknown", sx+4, sy+20)
		return
	}

	if tmpl.Icon != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx+4), float64(sy+20))
		screen.DrawImage(tmpl.Icon, op)
	}

	ebitenutil.DebugPrintAt(screen, tmpl.Name, sx+4, sy+20)

	dc := domainColor(tmpl.ArtifactDomain)
	vector.DrawFilledRect(screen, float32(sx+4), float32(sy+34), 8, 8, dc, false)
	ebitenutil.DebugPrintAt(screen, tmpl.ArtifactDomain, sx+16, sy+32)

	if tmpl.Effect != nil {
		ef := tmpl.Effect
		line := formatTrigger(ef.Trigger) + ": " + formatEffectType(ef.Type)
		ebitenutil.DebugPrintAt(screen, line, sx+4, sy+sh-16)
	} else if tmpl.GrantsAbility != "" {
		ebitenutil.DebugPrintAt(screen, "Ability: "+tmpl.GrantsAbility, sx+4, sy+sh-16)
	}
}

func (al *ArtifactLoadout) drawPicker(screen *ebiten.Image) {
	px, py, pw, ph := al.pickerBounds()
	fpx, fpy, fpw, fph := float32(px), float32(py), float32(pw), float32(ph)

	// Panel background.
	vector.DrawFilledRect(screen, fpx, fpy, fpw, fph, color.NRGBA{10, 10, 25, 240}, false)
	vector.StrokeRect(screen, fpx, fpy, fpw, fph, 2, color.RGBA{160, 140, 60, 255}, false)

	// Title.
	eliteOnly := al.ActiveSlot == 6
	title := "Select Artifact"
	if eliteOnly {
		title = "Select Elite Artifact"
	}
	ebitenutil.DebugPrintAt(screen, title, px+8, py+8)
	ebitenutil.DebugPrintAt(screen, "[Esc] Cancel  [Enter] Select", px+8, py+20)

	// Divider.
	vector.DrawFilledRect(screen, fpx+4, float32(py+34), fpw-8, 1, color.RGBA{80, 80, 80, 200}, false)

	if len(al.pickerEntries) == 0 {
		ebitenutil.DebugPrintAt(screen, "(no artifacts available)", px+8, py+46)
		return
	}

	listY := py + 38
	rowH := 22
	maxRows := (ph - 46) / rowH

	for i, e := range al.pickerEntries {
		vis := i - al.pickerScroll
		if vis < 0 || vis >= maxRows {
			continue
		}
		ey := listY + vis*rowH

		if i == al.PickerIdx {
			vector.DrawFilledRect(screen, fpx+2, float32(ey-1), fpw-4, float32(rowH-2), color.RGBA{60, 50, 80, 200}, false)
		}

		// Domain badge.
		dc := domainColor(e.tmpl.ArtifactDomain)
		vector.DrawFilledRect(screen, float32(px+8), float32(ey+5), 8, 8, dc, false)

		// Elite marker.
		prefix := "  "
		if e.tmpl.IsElite {
			prefix = "* "
		}
		ebitenutil.DebugPrintAt(screen, prefix+e.tmpl.Name, px+20, ey+2)
		ebitenutil.DebugPrintAt(screen, e.tmpl.ArtifactDomain, px+pw-70, ey+2)
	}

	// Scroll indicator.
	if len(al.pickerEntries) > maxRows {
		barH := float32(ph-46) * float32(maxRows) / float32(len(al.pickerEntries))
		var barY float32
		if len(al.pickerEntries) > maxRows {
			barY = float32(py+38) + float32(ph-46-int(barH))*float32(al.pickerScroll)/float32(len(al.pickerEntries)-maxRows)
		}
		vector.DrawFilledRect(screen, float32(px+pw-6), float32(py+38), 4, float32(ph-46), color.RGBA{40, 40, 40, 200}, false)
		vector.DrawFilledRect(screen, float32(px+pw-6), barY, 4, barH, color.RGBA{160, 140, 100, 255}, false)
	}
}
