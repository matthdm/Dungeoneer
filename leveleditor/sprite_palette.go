package leveleditor

import (
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type SpritePalette struct {
	Visible        bool
	Rect           image.Rectangle
	Entries        []string
	OnSelect       func(id string)
	selectedID     string
	columns        int
	rows           int
	spriteSize     int
	padding        int
	currentPage    int
	spritesPerPage int
	lastClickFrame int
	prevButton     image.Rectangle
	nextButton     image.Rectangle
}

func NewSpritePalette(w, h int, onSelect func(id string)) *SpritePalette {
	keys := make([]string, 0, len(SpriteRegistry))
	for k := range SpriteRegistry {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	spriteSize := 64
	padding := 10
	columns := 4
	rows := 3 // set rows per page
	spritesPerPage := columns * rows

	return &SpritePalette{
		Visible:        false,
		Entries:        keys,
		Rect:           image.Rect(w/4, h/4, 3*w/4, 3*h/4),
		OnSelect:       onSelect,
		spriteSize:     spriteSize,
		columns:        columns,
		rows:           rows,
		padding:        padding,
		spritesPerPage: spritesPerPage,
	}
}

func (sp *SpritePalette) Toggle() { sp.Visible = !sp.Visible }

func (sp *SpritePalette) Update() {
	if !sp.Visible {
		return
	}

	x, y := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if !clicked {
		return
	}

	// --- Handle Button Clicks First ---
	if PointInRect(x, y, sp.prevButton) {
		if sp.currentPage > 0 {
			sp.currentPage--
		}
		return // <-- always return to prevent palette closure
	}
	if PointInRect(x, y, sp.nextButton) {
		maxPage := (len(sp.Entries) - 1) / sp.spritesPerPage
		if sp.currentPage < maxPage {
			sp.currentPage++
		}
		return
	}

	// --- Close if clicked outside main palette area ---
	if !PointInRect(x, y, sp.Rect) {
		sp.Visible = false
		return
	}

	// --- Handle Sprite Selection ---
	ox := x - sp.Rect.Min.X
	oy := y - sp.Rect.Min.Y

	// Only consider clicks within the grid area (not buttons)
	gridHeight := sp.rows*(sp.spriteSize+sp.padding) - sp.padding
	if oy >= gridHeight {
		return // click was in the button area below sprites
	}

	col := ox / (sp.spriteSize + sp.padding)
	row := oy / (sp.spriteSize + sp.padding)

	if col >= sp.columns || row >= sp.rows {
		return // outside valid grid
	}

	indexOnPage := row*sp.columns + col
	index := sp.currentPage*sp.spritesPerPage + indexOnPage

	if index >= 0 && index < len(sp.Entries) {
		id := sp.Entries[index]
		sp.selectedID = id
		if sp.OnSelect != nil {
			sp.OnSelect(id)
		}
		sp.Visible = false
	}
}
func PointInRect(x, y int, r image.Rectangle) bool {
	return x >= r.Min.X && x < r.Max.X && y >= r.Min.Y && y < r.Max.Y
}

func (sp *SpritePalette) Draw(screen *ebiten.Image) {
	if !sp.Visible {
		return
	}

	DrawMenuOverlay(screen, DefaultOverlayColor)
	vector.DrawFilledRect(screen, float32(sp.Rect.Min.X), float32(sp.Rect.Min.Y),
		float32(sp.Rect.Dx()), float32(sp.Rect.Dy()), DefaultBackgroundColor, false)

	xStart := sp.Rect.Min.X + sp.padding
	yStart := sp.Rect.Min.Y + sp.padding

	start := sp.currentPage * sp.spritesPerPage
	end := min(start+sp.spritesPerPage, len(sp.Entries))
	entries := sp.Entries[start:end]

	for i, id := range entries {
		img := SpriteRegistry[id].Image
		col := i % sp.columns
		row := i / sp.columns
		x := xStart + col*(sp.spriteSize+sp.padding)
		y := yStart + row*(sp.spriteSize+sp.padding)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(img, op)
	}

	// Place buttons *below* the grid
	gridHeight := sp.rows*(sp.spriteSize+sp.padding) + sp.padding
	buttonY := sp.Rect.Min.Y + gridHeight + 10

	// Center horizontally
	buttonWidth := 80
	buttonHeight := 30

	prevRect := image.Rect(xStart, buttonY, xStart+buttonWidth, buttonY+buttonHeight)
	nextRect := image.Rect(sp.Rect.Max.X-buttonWidth-sp.padding, buttonY, sp.Rect.Max.X-sp.padding, buttonY+buttonHeight)

	// Draw rectangles
	vector.DrawFilledRect(screen, float32(prevRect.Min.X), float32(prevRect.Min.Y), float32(prevRect.Dx()), float32(prevRect.Dy()), color.RGBA{30, 30, 30, 200}, false)
	vector.DrawFilledRect(screen, float32(nextRect.Min.X), float32(nextRect.Min.Y), float32(nextRect.Dx()), float32(nextRect.Dy()), color.RGBA{30, 30, 30, 200}, false)

	ebitenutil.DebugPrintAt(screen, "< Prev", prevRect.Min.X+5, prevRect.Min.Y+8)
	ebitenutil.DebugPrintAt(screen, "Next >", nextRect.Min.X+5, nextRect.Min.Y+8)

	// Save for input detection
	sp.prevButton = prevRect
	sp.nextButton = nextRect
}

// Draws a semi-transparent overlay over the entire screen
func DrawMenuOverlay(screen *ebiten.Image, overlayColor color.Color) {
	screenBounds := screen.Bounds()
	vector.DrawFilledRect(screen, 0, 0, float32(screenBounds.Dx()), float32(screenBounds.Dy()), overlayColor, false)
}

// Default colors and constants for styling
var (
	DefaultOverlayColor    = color.RGBA{0, 0, 0, 128}
	DefaultBackgroundColor = color.RGBA{10, 10, 10, 255}
)
