package ui

import (
	"fmt"
	"image"

	"dungeoneer/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ProcGenMenu allows tweaking parameters for procedural generation.
type ProcGenMenu struct {
	Menu   *Menu
	params levels.GenParams
	onGen  func(levels.GenParams)
	onBack func()
}

// NewProcGenMenu creates the configuration menu.
func NewProcGenMenu(w, h int, p levels.GenParams, onGen func(levels.GenParams), onBack func()) *ProcGenMenu {
	style := DefaultMenuStyles()
	mw, mh := 340, 480
	mx := (w - mw) / 2
	my := (h - mh) / 2
	rect := image.Rect(mx, my, mx+mw, my+mh)

	pg := &ProcGenMenu{params: p, onGen: onGen, onBack: onBack}
	options := make([]MenuOption, 0, 12)
	// placeholders for params; actions nil
	for i := 0; i < 11; i++ {
		options = append(options, MenuOption{Text: ""})
	}
	options = append(options,
		MenuOption{Text: "Generate", Action: func() { pg.onGen(pg.params) }},
		MenuOption{Text: "Back", Action: pg.onBack},
	)
	menu := NewMenu(rect, "PROC GEN", options, style)
	menu.SetInstructions([]string{"W/S Navigate", "\u2190/\u2192 Adjust", "Enter Generate", "Esc Cancel"})
	pg.Menu = menu
	pg.refresh()
	return pg
}

func (pg *ProcGenMenu) refresh() {
	pg.Menu.options[0].Text = fmt.Sprintf("Seed: %d", pg.params.Seed)
	pg.Menu.options[1].Text = fmt.Sprintf("RoomCountMin: %d", pg.params.RoomCountMin)
	pg.Menu.options[2].Text = fmt.Sprintf("RoomCountMax: %d", pg.params.RoomCountMax)
	pg.Menu.options[3].Text = fmt.Sprintf("RoomWMin: %d", pg.params.RoomWMin)
	pg.Menu.options[4].Text = fmt.Sprintf("RoomWMax: %d", pg.params.RoomWMax)
	pg.Menu.options[5].Text = fmt.Sprintf("RoomHMin: %d", pg.params.RoomHMin)
	pg.Menu.options[6].Text = fmt.Sprintf("RoomHMax: %d", pg.params.RoomHMax)
	pg.Menu.options[7].Text = fmt.Sprintf("CorridorWidth: %d", pg.params.CorridorWidth)
	pg.Menu.options[8].Text = fmt.Sprintf("DashLaneMinLen: %d", pg.params.DashLaneMinLen)
	pg.Menu.options[9].Text = fmt.Sprintf("GrappleRange: %d", pg.params.GrappleRange)
	pg.Menu.options[10].Text = fmt.Sprintf("Extras: %d", pg.params.Extras)
}

func (pg *ProcGenMenu) adjust(idx, delta int) {
	clamp := func(v, min, max int) int {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}
	switch idx {
	case 0:
		pg.params.Seed += int64(delta)
	case 1:
		pg.params.RoomCountMin = clamp(pg.params.RoomCountMin+delta, 1, pg.params.RoomCountMax)
	case 2:
		pg.params.RoomCountMax = clamp(pg.params.RoomCountMax+delta, pg.params.RoomCountMin, 30)
	case 3:
		pg.params.RoomWMin = clamp(pg.params.RoomWMin+delta, 4, pg.params.RoomWMax)
	case 4:
		pg.params.RoomWMax = clamp(pg.params.RoomWMax+delta, pg.params.RoomWMin, 20)
	case 5:
		pg.params.RoomHMin = clamp(pg.params.RoomHMin+delta, 4, pg.params.RoomHMax)
	case 6:
		pg.params.RoomHMax = clamp(pg.params.RoomHMax+delta, pg.params.RoomHMin, 20)
	case 7:
		pg.params.CorridorWidth = clamp(pg.params.CorridorWidth+delta, 3, 4)
	case 8:
		pg.params.DashLaneMinLen = clamp(pg.params.DashLaneMinLen+delta, 4, 20)
	case 9:
		pg.params.GrappleRange = clamp(pg.params.GrappleRange+delta, 4, 20)
	case 10:
		pg.params.Extras = clamp(pg.params.Extras+delta, 0, 10)
	}
	pg.refresh()
}

func (pg *ProcGenMenu) Show()           { pg.Menu.Show() }
func (pg *ProcGenMenu) Hide()           { pg.Menu.Hide() }
func (pg *ProcGenMenu) IsVisible() bool { return pg.Menu.IsVisible() }
func (pg *ProcGenMenu) Update() {
	pg.Menu.Update()
	if !pg.Menu.IsVisible() {
		return
	}
	idx := pg.Menu.SelectedIndex()
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		pg.adjust(idx, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		pg.adjust(idx, 1)
	}
}
func (pg *ProcGenMenu) Draw(screen *ebiten.Image) { pg.Menu.Draw(screen) }
func (pg *ProcGenMenu) SetRect(r image.Rectangle) { pg.Menu.rect = r }
