package leveleditor

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
)

type EntitiesPalette struct {
	Visible    bool
	Rect       image.Rectangle
	Entries    []string
	OnSelect   func(id string)
	columns    int
	rows       int
	spriteSize int
	padding    int
}

func NewEntitiesPalette(w, h int, entries []string, onSelect func(id string)) *EntitiesPalette {
	ep := &EntitiesPalette{
		Visible:    false,
		Rect:       image.Rect(w/4, h/4, 3*w/4, 3*h/4),
		Entries:    entries,
		OnSelect:   onSelect,
		columns:    4,
		rows:       3,
		spriteSize: 64,
		padding:    5,
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
	if !PointInRect(x, y, ep.Rect) {
		ep.Visible = false
		return
	}

	ox := x - ep.Rect.Min.X
	oy := y - ep.Rect.Min.Y

	col := ox / (ep.spriteSize + ep.padding)
	row := oy / (ep.spriteSize + ep.padding)
	if col >= ep.columns || row >= ep.rows {
		return
	}

	idx := row*ep.columns + col
	if idx >= 0 && idx < len(ep.Entries) {
		id := ep.Entries[idx]
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
	yStart := ep.Rect.Min.Y + ep.padding

	for i, id := range ep.Entries {
		img := SpriteRegistry[id].Image
		col := i % ep.columns
		row := i / ep.columns
		x := xStart + col*(ep.spriteSize+ep.padding)
		y := yStart + row*(ep.spriteSize+ep.padding)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(img, op)
		ebitenutil.DebugPrintAt(screen, id, x, y+ep.spriteSize-8)
	}
}
