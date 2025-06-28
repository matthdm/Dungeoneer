package leveleditor

import (
	"dungeoneer/levels"
	"fmt"
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
	EntityPaletteOpen  bool
	level              *levels.Level
	cursorX            int
	cursorY            int
	cursorScreen       image.Point
	Palette            *SpritePalette
	EntitiesPalette    *EntitiesPalette
	JustSelectedSprite bool
	JustSelectedEntity bool
	SelectedEntityID   string
	AutoTile           bool
}

func NewEditor(level *levels.Level, screenWidth, screenHeight int) *Editor {
	editor := &Editor{
		Active:            true,
		SelectedID:        "",
		PaletteOpen:       false,
		EntityPaletteOpen: false,
		level:             level,
		AutoTile:          true,
	}

	// Create the palette with a callback to set the selected sprite
	editor.Palette = NewSpritePalette(screenWidth, screenHeight, editor.SetSelectedSprite)
	// Entities palette with some default monster sprites
	entries := []string{"Statue", "BlueMan"}
	editor.EntitiesPalette = NewEntitiesPalette(screenWidth, screenHeight, entries, editor.SetSelectedEntity)

	return editor
}
func (e *Editor) TogglePalette() {
	e.PaletteOpen = !e.PaletteOpen
	e.Palette.Visible = e.PaletteOpen
}
func (e *Editor) ToggleEntityPalette() {
	e.EntityPaletteOpen = !e.EntityPaletteOpen
	e.EntitiesPalette.Visible = e.EntityPaletteOpen
}

func (e *Editor) SetSelectedSprite(id string) {
	e.SelectedID = id
	e.PaletteOpen = false
	e.JustSelectedSprite = true
}
func (e *Editor) SetSelectedEntity(id string) {
	e.SelectedEntityID = id
	e.EntityPaletteOpen = false
	e.JustSelectedEntity = true
}
func (e *Editor) Update(screenToTile func() (int, int)) {
	if !e.Active {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		e.AutoTile = !e.AutoTile
		if e.AutoTile {
			fmt.Println("Auto-tiling enabled")
		} else {
			fmt.Println("Auto-tiling disabled")
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		e.TogglePalette()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		e.ToggleEntityPalette()
	}

	if e.PaletteOpen {
		e.Palette.Update()
		return
	}
	if e.EntityPaletteOpen {
		e.EntitiesPalette.Update()
		return
	}

	// Reset justSelectedSprite so we allow clicks next frame
	e.JustSelectedSprite = false
	e.JustSelectedEntity = false
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
	id := e.SelectedID

	meta, ok := SpriteRegistry[id]
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
	if len(tile.Sprites) >= 3 {
		return // already has base + 1
	}

	tile.AddSpriteByID(id, meta.Image)
	tile.IsWalkable = meta.IsWalkable
}

// PlaceSelectedEntityAt places the chosen entity on the given tile.
// Only one entity is allowed per tile; existing entity is overwritten.
func (e *Editor) PlaceSelectedEntityAt(tx, ty int) {
	if e.SelectedEntityID == "" {
		return
	}

	// ensure coordinates are inside the level bounds
	if tx < 0 || ty < 0 || tx >= e.level.W || ty >= e.level.H {
		return
	}

	// remove existing entity on this tile if any
	replaced := false
	for i, ent := range e.level.Entities {
		if ent.X == tx && ent.Y == ty {
			e.level.Entities[i].Type = "AmbushMonster"
			e.level.Entities[i].SpriteID = e.SelectedEntityID
			replaced = true
			break
		}
	}
	if !replaced {
		e.level.Entities = append(e.level.Entities, levels.PlacedEntity{
			X:        tx,
			Y:        ty,
			Type:     "AmbushMonster",
			SpriteID: e.SelectedEntityID,
		})
	}
}
