package leveleditor

import (
	"dungeoneer/levels"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Editor struct {
	Active             bool
	SelectedID         string
	PaletteOpen        bool
	level              *levels.Level
	cursorX            int
	cursorY            int
	cursorScreen       image.Point
	Palette            *SpritePalette
	JustSelectedSprite bool
}

func NewEditor(level *levels.Level, screenWidth, screenHeight int) *Editor {
	editor := &Editor{
		Active:      false,
		SelectedID:  "",
		PaletteOpen: false,
		level:       level,
	}

	// Create the palette with a callback to set the selected sprite
	editor.Palette = NewSpritePalette(screenWidth, screenHeight, editor.SetSelectedSprite)

	return editor
}
func (e *Editor) TogglePalette() {
	e.PaletteOpen = !e.PaletteOpen
	e.Palette.Visible = e.PaletteOpen
}

func (e *Editor) SetSelectedSprite(id string) {
	e.SelectedID = id
	e.PaletteOpen = false
	e.JustSelectedSprite = true
}
func (e *Editor) Update(screenToTile func() (int, int)) {
	if !e.Active {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		e.TogglePalette()
	}

	if e.PaletteOpen {
		e.Palette.Update()
		return
	}

	// Reset justSelectedSprite so we allow clicks next frame
	e.JustSelectedSprite = false
}
func (e *Editor) Draw(screen *ebiten.Image, tileSize int, camX, camY float64, camScale float64) {
	if !e.Active {
		return
	}
	x := float64(e.cursorX * tileSize)
	y := float64(e.cursorY * tileSize)

	// Apply camera
	screenX := (x - camX) * camScale
	screenY := (y - camY) * camScale

	vector.StrokeRect(screen, float32(screenX), float32(screenY),
		float32(tileSize)*float32(camScale), float32(tileSize)*float32(camScale),
		2, color.RGBA{255, 255, 0, 255}, false)
}

func (e *Editor) PlaceSelectedSpriteAt(tx, ty int) {
	if e.SelectedID == "" {
		return
	}
	meta, ok := SpriteRegistry[e.SelectedID]
	if !ok {
		return
	}
	tile := e.level.Tile(tx, ty)
	if tile == nil {
		return
	}

	// Don't add if already present
	if tile.HasSpriteID(e.SelectedID) {
		return
	}

	// Allow base floor + 1 extra
	if len(tile.Sprites) >= 2 {
		return // already has base + 1
	}

	tile.AddSpriteByID(e.SelectedID, meta.Image)
	tile.IsWalkable = meta.IsWalkable
}
