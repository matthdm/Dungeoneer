package ui

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"image"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	paletteVisible bool
	paletteRect    = image.Rect(100, 100, 540, 420)
	paletteScroll  int
	itemIDs        []string
)

const (
	paletteColumns = 4
	paletteRows    = 3
	spriteSize     = 64
	padding        = 5
)

func ensureIDs() {
	if len(itemIDs) == len(items.Registry) {
		return
	}
	itemIDs = itemIDs[:0]
	for id := range items.Registry {
		itemIDs = append(itemIDs, id)
	}
	sort.Strings(itemIDs)
}

// HandleItemPaletteInput updates palette state and handles clicks.
func HandleItemPaletteInput(p *entities.Player) {
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		paletteVisible = !paletteVisible
	}
	if !paletteVisible {
		return
	}

	_, wheel := ebiten.Wheel()
	if wheel != 0 {
		paletteScroll -= int(wheel) * paletteColumns
		if paletteScroll < 0 {
			paletteScroll = 0
		}
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	x, y := ebiten.CursorPosition()
	if !pointInRect(x, y, paletteRect) {
		paletteVisible = false
		return
	}

	ox := x - paletteRect.Min.X - padding
	oy := y - paletteRect.Min.Y - padding
	col := ox / (spriteSize + padding)
	row := oy / (spriteSize + padding)
	if col < 0 || col >= paletteColumns || row < 0 || row >= paletteRows {
		return
	}

	idx := paletteScroll + row*paletteColumns + col
	ensureIDs()
	if idx >= 0 && idx < len(itemIDs) {
		id := itemIDs[idx]
		if p != nil && p.Inventory != nil {
			p.Inventory.AddItem(items.NewItem(id))
		}
	}
}

// DrawItemPalette renders the developer item palette.
func DrawItemPalette(screen *ebiten.Image) {
	if !paletteVisible {
		return
	}
	ensureIDs()
	DrawMenuOverlay(screen, DefaultOverlayColor)
	vector.DrawFilledRect(screen, float32(paletteRect.Min.X), float32(paletteRect.Min.Y),
		float32(paletteRect.Dx()), float32(paletteRect.Dy()), DefaultBackgroundColor, false)

	startX := paletteRect.Min.X + padding
	startY := paletteRect.Min.Y + padding
	perPage := paletteColumns * paletteRows

	for i := 0; i < perPage && i+paletteScroll < len(itemIDs); i++ {
		id := itemIDs[i+paletteScroll]
		itm := items.Registry[id]
		col := i % paletteColumns
		row := i / paletteColumns
		x := startX + col*(spriteSize+padding)
		y := startY + row*(spriteSize+padding)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(itm.Icon, op)
	}

	mx, my := ebiten.CursorPosition()
	if pointInRect(mx, my, paletteRect) {
		ox := mx - startX
		oy := my - startY
		col := ox / (spriteSize + padding)
		row := oy / (spriteSize + padding)
		if col >= 0 && col < paletteColumns && row >= 0 && row < paletteRows {
			idx := paletteScroll + row*paletteColumns + col
			if idx >= 0 && idx < len(itemIDs) {
				name := items.Registry[itemIDs[idx]].Name
				ebitenutil.DebugPrintAt(screen, name, paletteRect.Min.X, paletteRect.Max.Y-15)
			}
		}
	}
}
