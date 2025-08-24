package entities

import (
	"dungeoneer/items"

	"github.com/hajimehoshi/ebiten/v2"
)

// ItemDrop represents an item lying on a level tile.
type ItemDrop struct {
	TileX, TileY int
	Item         items.Item
}

// Draw renders the item's icon at its tile position.
func (d *ItemDrop) Draw(screen *ebiten.Image, tileSize int, camX, camY, camScale, cx, cy float64) {
	if d == nil || d.Item.Icon == nil {
		return
	}
	x, y := isoToScreenFloat(float64(d.TileX), float64(d.TileY), tileSize)
	b := d.Item.Icon.Bounds()
	w := float64(b.Dx())
	h := float64(b.Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(-w/4, -h/2)
	op.GeoM.Translate(x, y)
	op.GeoM.Translate(-camX, camY)
	op.GeoM.Scale(camScale, camScale)
	op.GeoM.Translate(cx, cy)
	screen.DrawImage(d.Item.Icon, op)
}

// Update is a no-op placeholder for ItemDrop.
func (d *ItemDrop) Update() {}
