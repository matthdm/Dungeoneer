package leveleditor

import (
	"dungeoneer/levels"
	"fmt"
	"image"
	"image/color"
	"strings"

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
	AutoTile           bool
}

func NewEditor(level *levels.Level, screenWidth, screenHeight int) *Editor {
	editor := &Editor{
		Active:      true,
		SelectedID:  "",
		PaletteOpen: false,
		level:       level,
		AutoTile:    true,
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
	id := e.SelectedID
	if e.AutoTile && strings.Contains(id, "_wall") {
		if parts := strings.SplitN(id, "_", 2); len(parts) == 2 {
			id = AutoSelectWallVariant(e.level, tx, ty, parts[0])
		}
	}

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
	if len(tile.Sprites) >= 2 {
		return // already has base + 1
	}

	tile.AddSpriteByID(id, meta.Image)
	tile.IsWalkable = meta.IsWalkable
}

// AutoSelectWallVariant chooses the correct wall sprite variant for the given
// flavor based on adjacent tiles. It returns the sprite ID that should be used
// at the provided location.
func AutoSelectWallVariant(level *levels.Level, x, y int, flavor string) string {
	isSame := func(tx, ty int) bool {
		t := level.Tile(tx, ty)
		if t == nil {
			return false
		}
		prefix := flavor + "_"
		for _, s := range t.Sprites {
			if strings.HasPrefix(s.ID, prefix) && strings.Contains(s.ID, "wall") {
				return true
			}
		}
		return false
	}

	up := isSame(x, y-1)
	down := isSame(x, y+1)
	left := isSame(x-1, y)
	right := isSame(x+1, y)

	// Very simple heuristic using three generic variants
	connections := 0
	if up {
		connections++
	}
	if down {
		connections++
	}
	if left {
		connections++
	}
	if right {
		connections++
	}

	base := flavor + "_"
	if connections >= 3 {
		return base + "wall_nesw"
	}
	if (up && down) || (left && right) {
		return base + "wall"
	}
	return base + "wall_nwse"
}
