package items

import (
	"dungeoneer/images"
	"encoding/json"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// sheetEntry represents a single icon entry in the reverse map JSON.
type sheetEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Pos  struct {
		Row int `json:"row"`
		Col int `json:"col"`
	} `json:"position"`
}

// LoadItemSheet registers items from the provided sheet and mapping.
func LoadItemSheet(img *ebiten.Image, entries []sheetEntry) {
	for _, e := range entries {
		x0 := e.Pos.Col * 32
		y0 := e.Pos.Row * 32
		sub := img.SubImage(image.Rect(x0, y0, x0+32, y0+32)).(*ebiten.Image)
		scaled := ebiten.NewImage(64, 64)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2)
		scaled.DrawImage(sub, op)

		tmpl := &ItemTemplate{
			ID:         e.ID,
			Name:       e.Name,
			Type:       ItemMisc,
			Stackable:  false,
			MaxStack:   1,
			Usable:     false,
			Equippable: false,
			Stats:      map[string]int{},
			Icon:       scaled,
		}
		RegisterItem(tmpl)
	}
}

// LoadDefaultItems loads the bundled item sheet and mapping.
func LoadDefaultItems() error {
	img, err := images.LoadEmbeddedImage(images.Item_subset_png)
	if err != nil {
		return err
	}
	var entries []sheetEntry
	if err := json.Unmarshal(images.Item_reversemap_rows_0_1_2_json, &entries); err != nil {
		return err
	}
	LoadItemSheet(img, entries)
	return nil
}
