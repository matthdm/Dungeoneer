package leveleditor

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type EntitiesPalette struct {
	Visible         bool
	Rect            image.Rectangle
	Entries         []string
	OnSelect        func(id string)
	columns         int
	rows            int
	spriteSize      int
	padding         int
	currentPage     int
	entitiesPerPage int
	prevButton      image.Rectangle
	nextButton      image.Rectangle
}

func NewEntitiesPalette(w, h int, entries []string, onSelect func(id string)) *EntitiesPalette {
	columns := 4
	rows := 3 // rows per page
	entitiesPerPage := columns * rows

	ep := &EntitiesPalette{
		Visible:         false,
		Rect:            image.Rect(w/4, h/4, 3*w/4, 3*h/4),
		Entries:         entries,
		OnSelect:        onSelect,
		columns:         columns,
		rows:            rows,
		spriteSize:      64,
		padding:         5,
		entitiesPerPage: entitiesPerPage,
		currentPage:     0,
	}
	return ep
}

func (ep *EntitiesPalette) Update() {
	if !ep.Visible {
		return
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	x, y := ebiten.CursorPosition()

	// --- Handle pagination button clicks first ---
	if PointInRect(x, y, ep.prevButton) {
		if ep.currentPage > 0 {
			ep.currentPage--
		}
		return // always return to prevent palette closure
	}
	if PointInRect(x, y, ep.nextButton) {
		maxPage := (len(ep.Entries) - 1) / ep.entitiesPerPage
		if ep.currentPage < maxPage {
			ep.currentPage++
		}
		return
	}

	// --- Close if clicked outside main palette area ---
	if !PointInRect(x, y, ep.Rect) {
		ep.Visible = false
		return
	}

	// --- Handle entity selection ---
	ox := x - ep.Rect.Min.X
	oy := y - ep.Rect.Min.Y

	gridOffset := ep.padding*2 + 20
	if oy < gridOffset {
		return // click was in button UI
	}

	oy -= gridOffset

	// Only consider clicks within the grid area (not buttons)
	gridHeight := ep.rows*(ep.spriteSize+ep.padding) - ep.padding
	if oy >= gridHeight {
		return // click was in the button area below sprites
	}

	col := ox / (ep.spriteSize + ep.padding)
	row := oy / (ep.spriteSize + ep.padding)

	if col >= ep.columns || row >= ep.rows {
		return // outside valid grid
	}

	indexOnPage := row*ep.columns + col
	index := ep.currentPage*ep.entitiesPerPage + indexOnPage

	if index >= 0 && index < len(ep.Entries) {
		id := ep.Entries[index]
		if ep.OnSelect != nil {
			ep.OnSelect(id)
		}
		ep.Visible = false
	}
}

func (ep *EntitiesPalette) Draw(screen *ebiten.Image) {
	if !ep.Visible {
		return
	}

	DrawMenuOverlay(screen, DefaultOverlayColor)
	vector.DrawFilledRect(screen, float32(ep.Rect.Min.X), float32(ep.Rect.Min.Y),
		float32(ep.Rect.Dx()), float32(ep.Rect.Dy()), DefaultBackgroundColor, false)

	xStart := ep.Rect.Min.X + ep.padding
	yStart := ep.Rect.Min.Y + ep.padding*2 + 20

	// Draw pagination buttons
	buttonY := ep.Rect.Min.Y + ep.padding
	arrowW := 20
	arrowH := 20
	prevBtn := image.Rect(xStart, buttonY, xStart+arrowW, buttonY+arrowH)
	nextBtn := image.Rect(ep.Rect.Max.X-ep.padding-arrowW, buttonY, ep.Rect.Max.X-ep.padding, buttonY+arrowH)

	vector.DrawFilledRect(screen, float32(prevBtn.Min.X), float32(prevBtn.Min.Y), float32(prevBtn.Dx()), float32(prevBtn.Dy()), color.RGBA{30, 30, 30, 200}, false)
	vector.DrawFilledRect(screen, float32(nextBtn.Min.X), float32(nextBtn.Min.Y), float32(nextBtn.Dx()), float32(nextBtn.Dy()), color.RGBA{30, 30, 30, 200}, false)
	ebitenutil.DebugPrintAt(screen, "<", prevBtn.Min.X+6, prevBtn.Min.Y+4)
	ebitenutil.DebugPrintAt(screen, ">", nextBtn.Min.X+6, nextBtn.Min.Y+4)

	// Store button rectangles for click detection
	ep.prevButton = prevBtn
	ep.nextButton = nextBtn

	// Display page info
	currentPageStr := ""
	if len(ep.Entries) > 0 {
		maxPage := (len(ep.Entries) - 1) / ep.entitiesPerPage
		currentPageStr = string(rune('1' + ep.currentPage))
		if maxPage > 0 {
			currentPageStr += " / " + string(rune('1'+maxPage))
		}
	}
	ebitenutil.DebugPrintAt(screen, currentPageStr, prevBtn.Max.X+5, buttonY+4)

	// Get entities for current page
	start := ep.currentPage * ep.entitiesPerPage
	end := start + ep.entitiesPerPage
	if end > len(ep.Entries) {
		end = len(ep.Entries)
	}

	// Draw entities on current page
	for i := start; i < end; i++ {
		id := ep.Entries[i]
		img := SpriteRegistry[id].Image
		indexOnPage := i - start
		col := indexOnPage % ep.columns
		row := indexOnPage / ep.columns
		x := xStart + col*(ep.spriteSize+ep.padding)
		y := yStart + row*(ep.spriteSize+ep.padding)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(img, op)
		ebitenutil.DebugPrintAt(screen, id, x, y+ep.spriteSize-8)
	}
}
