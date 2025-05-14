package leveleditor

import (
	"dungeoneer/tiles"

	"github.com/hajimehoshi/ebiten/v2"
)

type Editor struct {
	SelectedSprite *ebiten.Image
}

func NewEditor() *Editor {
	return &Editor{}
}

func (e *Editor) SetSelectedSprite(s *ebiten.Image) {
	e.SelectedSprite = s
}

func (e *Editor) GetSelectedSprite() *ebiten.Image {
	return e.SelectedSprite
}

func (e *Editor) PlaceTile(t *tiles.Tile) {
	if e.SelectedSprite != nil && t != nil {
		t.AddSprite(e.SelectedSprite)
	}
}
