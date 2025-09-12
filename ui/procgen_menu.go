package ui

import (
	"fmt"
	"image"

	"dungeoneer/levels"
	"dungeoneer/sprites"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ProcGenMenu presents and edits procedural generation parameters using a ScrollList.
type ProcGenMenu struct {
	List   *ScrollList
	params levels.GenParams
	onGen  func(levels.GenParams)
	onBack func()
}

// Indices into the scroll list (keep actions last)
const (
	idxSeed = iota
	idxRoomCountMin
	idxRoomCountMax
	idxRoomWMin
	idxRoomWMax
	idxRoomHMin
	idxRoomHMax
	idxCorridorWidth
	idxDashLaneMinLen
	idxGrappleRange
	idxExtras
	idxCoverageTarget
	idxFillerRoomsMax
	idxWallFlavor
	idxFloorFlavor
	idxGenerate
	idxBack
)

var (
	// Use predefined flavors already in the project.
	wallFlavors  = sprites.WallFlavors
	floorFlavors = sprites.WallFlavors // floors mirror wall flavor set for now
)

// NewProcGenMenu creates the configuration menu.
func NewProcGenMenu(w, h int, p levels.GenParams, onGen func(levels.GenParams), onBack func()) *ProcGenMenu {
	style := DefaultMenuStyles()

	mw, mh := 340, 480
	mx := (w - mw) / 2
	my := (h - mh) / 2
	rect := image.Rect(mx, my, mx+mw, my+mh)

	pg := &ProcGenMenu{
		params: p,
		onGen:  onGen,
		onBack: onBack,
	}

	opts := []MenuOption{
		{Text: ""}, // Seed
		{Text: ""}, // RoomCountMin
		{Text: ""}, // RoomCountMax
		{Text: ""}, // RoomWMin
		{Text: ""}, // RoomWMax
		{Text: ""}, // RoomHMin
		{Text: ""}, // RoomHMax
		{Text: ""}, // CorridorWidth
		{Text: ""}, // DashLaneMinLen
		{Text: ""}, // GrappleRange
		{Text: ""}, // Extras
		{Text: ""}, // CoverageTarget
		{Text: ""}, // FillerRoomsMax
		{Text: ""}, // WallFlavor
		{Text: ""}, // FloorFlavor

		{Text: "Generate", Action: func() { pg.onGen(pg.params) }},
		{Text: "Back", Action: func() {
			if pg.onBack != nil {
				pg.onBack()
			}
		}},
	}

	list := NewScrollList(rect, "PROC GEN", opts, style)
	list.SetInstructions([]string{
		"W/S or Wheel Navigate",
		"←/→ Adjust",
		"Enter Generate",
		"Esc Cancel",
	})
	pg.List = list
	pg.refresh()
	return pg
}

// --- Public controls ---

func (pg *ProcGenMenu) Show()           { pg.List.Show() }
func (pg *ProcGenMenu) Hide()           { pg.List.Hide() }
func (pg *ProcGenMenu) IsVisible() bool { return pg.List.IsVisible() }
func (pg *ProcGenMenu) SetRect(r image.Rectangle) {
	pg.List.rect = r
	pg.List.calcVisible() // same package; allowed to call
}

// Update input and apply left/right adjustments.
func (pg *ProcGenMenu) Update() {
	pg.List.Update()
	if !pg.List.IsVisible() {
		return
	}
	idx := pg.List.selected

	// Back on ESC (kept here so menu is self-contained)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if pg.onBack != nil {
			pg.onBack()
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		pg.adjust(idx, -1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		pg.adjust(idx, +1)
	}
}

// Draw renders the scroll list.
func (pg *ProcGenMenu) Draw(screen *ebiten.Image) {
	pg.List.Draw(screen)
}

// --- Internals ---

func (pg *ProcGenMenu) refresh() {
	opts := pg.List.options
	opts[idxSeed].Text = fmt.Sprintf("Seed: %d", pg.params.Seed)

	opts[idxRoomCountMin].Text = fmt.Sprintf("RoomCountMin: %d", pg.params.RoomCountMin)
	opts[idxRoomCountMax].Text = fmt.Sprintf("RoomCountMax: %d", pg.params.RoomCountMax)

	opts[idxRoomWMin].Text = fmt.Sprintf("RoomWMin: %d", pg.params.RoomWMin)
	opts[idxRoomWMax].Text = fmt.Sprintf("RoomWMax: %d", pg.params.RoomWMax)
	opts[idxRoomHMin].Text = fmt.Sprintf("RoomHMin: %d", pg.params.RoomHMin)
	opts[idxRoomHMax].Text = fmt.Sprintf("RoomHMax: %d", pg.params.RoomHMax)

	opts[idxCorridorWidth].Text = fmt.Sprintf("CorridorWidth: %d", pg.params.CorridorWidth)
	opts[idxDashLaneMinLen].Text = fmt.Sprintf("DashLaneMinLen: %d", pg.params.DashLaneMinLen)
	opts[idxGrappleRange].Text = fmt.Sprintf("GrappleRange: %d", pg.params.GrappleRange)
	opts[idxExtras].Text = fmt.Sprintf("Extras: %d", pg.params.Extras)
	opts[idxCoverageTarget].Text = fmt.Sprintf("CoverageTarget: %.2f", pg.params.CoverageTarget)
	opts[idxFillerRoomsMax].Text = fmt.Sprintf("FillerRoomsMax: %d", pg.params.FillerRoomsMax)
	opts[idxWallFlavor].Text = fmt.Sprintf("WallFlavor: %s", safeFlavor(pg.params.WallFlavor))
	opts[idxFloorFlavor].Text = fmt.Sprintf("FloorFlavor: %s", safeFlavor(pg.params.FloorFlavor))

	pg.List.SetOptions(opts) // also recomputes visible window
}

func (pg *ProcGenMenu) adjust(idx, delta int) {
	clampInt := func(v, lo, hi int) int {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}
	clampFloat := func(v, lo, hi float64) float64 {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}
	indexOf := func(arr []string, s string) int {
		for i, v := range arr {
			if v == s {
				return i
			}
		}
		return -1
	}
	wrap := func(i, n int) int {
		if n <= 0 {
			return 0
		}
		if i < 0 {
			return n - 1
		}
		if i >= n {
			return 0
		}
		return i
	}

	switch idx {
	case idxSeed:
		pg.params.Seed += int64(delta)

	case idxRoomCountMin:
		pg.params.RoomCountMin = clampInt(pg.params.RoomCountMin+delta, 1, pg.params.RoomCountMax)

	case idxRoomCountMax:
		pg.params.RoomCountMax = clampInt(pg.params.RoomCountMax+delta, pg.params.RoomCountMin, 30)

	case idxRoomWMin:
		pg.params.RoomWMin = clampInt(pg.params.RoomWMin+delta, 2, pg.params.RoomWMax)

	case idxRoomWMax:
		pg.params.RoomWMax = clampInt(pg.params.RoomWMax+delta, pg.params.RoomWMin, 64)

	case idxRoomHMin:
		pg.params.RoomHMin = clampInt(pg.params.RoomHMin+delta, 2, pg.params.RoomHMax)

	case idxRoomHMax:
		pg.params.RoomHMax = clampInt(pg.params.RoomHMax+delta, pg.params.RoomHMin, 64)

	case idxCorridorWidth:
		pg.params.CorridorWidth = clampInt(pg.params.CorridorWidth+delta, 1, 6)

	case idxDashLaneMinLen:
		pg.params.DashLaneMinLen = clampInt(pg.params.DashLaneMinLen+delta, 1, 32)

	case idxGrappleRange:
		pg.params.GrappleRange = clampInt(pg.params.GrappleRange+delta, 0, 32)

	case idxExtras:
		pg.params.Extras = clampInt(pg.params.Extras+delta, 0, 10)

	case idxCoverageTarget:
		step := 0.05
		pg.params.CoverageTarget = clampFloat(pg.params.CoverageTarget+float64(delta)*step, 0.10, 0.90)

	case idxFillerRoomsMax:
		pg.params.FillerRoomsMax = clampInt(pg.params.FillerRoomsMax+delta, 0, 10)

	case idxWallFlavor:
		cur := safeFlavor(pg.params.WallFlavor)
		i := indexOf(wallFlavors, cur)
		if i < 0 {
			i = 0
		}
		pg.params.WallFlavor = wallFlavors[wrap(i+delta, len(wallFlavors))]

	case idxFloorFlavor:
		cur := safeFlavor(pg.params.FloorFlavor)
		i := indexOf(floorFlavors, cur)
		if i < 0 {
			i = 0
		}
		pg.params.FloorFlavor = floorFlavors[wrap(i+delta, len(floorFlavors))]
	}

	pg.refresh()
}

func safeFlavor(f string) string {
	if f == "" {
		return "normal"
	}
	return f
}
