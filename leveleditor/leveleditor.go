package leveleditor

import (
	"dungeoneer/levels"
	"dungeoneer/sprites"
	"dungeoneer/tiles"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type EditorMode int

const (
	ModeSpawner EditorMode = iota
	ModeDelete
)

type Editor struct {
	Active             bool
	SelectedID         string
	PaletteOpen        bool
	EntityPaletteOpen  bool
	level              *levels.Level
	cursorX            int
	cursorY            int
	layered            *levels.LayeredLevel
	layerIndex         int
	cursorScreen       image.Point
	Palette            *SpritePalette
	EntitiesPalette    *EntitiesPalette
	JustSelectedSprite bool
	JustSelectedEntity bool
	SelectedEntityID   string
	EntityMode         EditorMode
	spawnerButtonRect  image.Rectangle
	deleteButtonRect   image.Rectangle
	clearButtonRect    image.Rectangle
	// OnLayerChange is called whenever the active layer changes.
	OnLayerChange func(*levels.Level)
	// OnStairPlaced is called when a stairwell sprite is placed.
	OnStairPlaced func(x, y int, spriteID string)
}

func NewEditor(level *levels.Level, screenWidth, screenHeight int) *Editor {
	editor := &Editor{
		Active:            true,
		SelectedID:        "",
		PaletteOpen:       false,
		EntityPaletteOpen: false,
		EntityMode:        ModeSpawner,
		level:             level,
	}

	// Create the palette with a callback to set the selected sprite
	editor.Palette = NewSpritePalette(screenWidth, screenHeight, editor.SetSelectedSprite)
	// Entities palette with some default monster sprites
	entries := []string{
		"GreyKnight",
		"Chimera",
		"Sentinel",
		"Sorcerer",
		"Duchess",
		"Absolem",
		"Death",
		"Oracle",
		"Jester",
		"GreaterDemon",
		"DemonKnight",
		"Abomination",
		"QueenOfDarkness",
		"LesserDemon",
		"TheTerror",
		"Celestial",
		"Demon",
		"Apparition",
		"Griffon",
		"Manticore",
		"Minotaur",
		"TorturedSoul",
		"ChaosTotem",
		"HydraBase",
		"Hydra2",
		"Hydra3",
		"Hydra4",
		"Hydra5",
		"Hydra6",
		"HydraFinal",
		"LesserDragon",
		"GoldenDragon",
		"Wyvern",
		"LesserCaveDragon",
		"GhostWyvern",
		"AncientDragonRed",
		"AncientDragonBlack",
		"AncientDragon",
		"PetrifiedDragon",
		"GreaterDragon",
		"GreaterRedDragon",
		"BlueMan",
		"Cyclops",
		"TwoHeadedOgre",
		"RedChampion",
		"BlueChampion",
		"Caveman",
		"RockCollector",
		"RedMan"}
	editor.EntitiesPalette = NewEntitiesPalette(screenWidth, screenHeight, entries, editor.SetSelectedEntity)

	return editor
}

// NewLayeredEditor creates an editor for a layered level.
func NewLayeredEditor(ll *levels.LayeredLevel, w, h int) *Editor {
	ed := NewEditor(ll.ActiveLayer(), w, h)
	ed.layered = ll
	ed.layerIndex = ll.ActiveIndex
	return ed
}

// LinkNewLayer creates a blank layer and appends it to the layered level.
func (e *Editor) LinkNewLayer() {
	if e.layered == nil || e.level == nil {
		return
	}
	ss, err := sprites.LoadSpriteSheet(e.level.TileSize)
	if err != nil {
		fmt.Println("failed to load spritesheet:", err)
		return
	}
	newL := levels.CreateNewBlankLevel(e.level.W, e.level.H, e.level.TileSize, ss)
	e.layered.AddLayer(newL)
	e.layerIndex = len(e.layered.Layers) - 1
	e.layered.ActiveIndex = e.layerIndex
	e.level = newL
	if e.OnLayerChange != nil {
		e.OnLayerChange(e.level)
	}
	fmt.Println("Linked new layer", e.layerIndex)
}

// UnlinkLastLayer removes the last layer from the layered level.
func (e *Editor) UnlinkLastLayer() {
	if e.layered == nil {
		return
	}
	if len(e.layered.Layers) <= 1 {
		fmt.Println("Cannot unlink base layer")
		return
	}
	e.layered.RemoveLastLayer()
	if e.layerIndex >= len(e.layered.Layers) {
		e.layerIndex = len(e.layered.Layers) - 1
	}
	e.level = e.layered.ActiveLayer()
	if e.OnLayerChange != nil {
		e.OnLayerChange(e.level)
	}
	fmt.Println("Unlinked last layer, remaining:", len(e.layered.Layers))
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

	if e.layered != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
			e.PrevLayer()
			fmt.Println("Switched to layer", e.layerIndex)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
			e.NextLayer()
			fmt.Println("Switched to layer", e.layerIndex)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
			e.LinkNewLayer()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
			e.UnlinkLastLayer()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		e.TogglePalette()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		e.ToggleEntityPalette()
	}

	// Handle mode button clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if PointInRect(mx, my, e.spawnerButtonRect) {
			e.EntityMode = ModeSpawner
			return
		}
		if PointInRect(mx, my, e.deleteButtonRect) {
			e.EntityMode = ModeDelete
			return
		}
		if PointInRect(mx, my, e.clearButtonRect) {
			e.SelectedEntityID = ""
			return
		}
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

// IsMenuOpen returns true if any palette menu is currently visible
func (e *Editor) IsMenuOpen() bool {
	return e.PaletteOpen || e.EntityPaletteOpen || e.Palette.Visible || e.EntitiesPalette.Visible
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
	// If placing a door sprite, enforce door state and clear other door sprites.
	if isDoorSpriteID(id) {
		// Remove any existing door sprites so there is only one active state.
		pruned := tile.Sprites[:0]
		for _, s := range tile.Sprites {
			if !isDoorSpriteID(s.ID) {
				pruned = append(pruned, s)
			}
		}
		pruned = append(pruned, tiles.SpriteRef{ID: id, Image: meta.Image})
		tile.Sprites = pruned
		applyDoorState(tile, id)
	}
	lower := strings.ToLower(id)
	if strings.Contains(lower, "stairsascending") || strings.Contains(lower, "stairsdecending") || strings.Contains(lower, "stairsdescending") {
		if e.OnStairPlaced != nil {
			e.OnStairPlaced(tx, ty, id)
		}
	}
}

func isDoorSpriteID(id string) bool {
	lower := strings.ToLower(id)
	return strings.Contains(lower, "door_locked") || strings.Contains(lower, "door_unlocked") ||
		strings.Contains(lower, "lockeddoor") || strings.Contains(lower, "unlockeddoor")
}

func applyDoorState(tile *tiles.Tile, id string) {
	if tile == nil {
		return
	}
	lower := strings.ToLower(id)
	tile.SetTag(tiles.TagDoor)
	tile.DoorSpriteID = id
	if strings.Contains(lower, "door_unlocked") || strings.Contains(lower, "unlockeddoor") {
		tile.DoorState = 1 // open
		tile.IsWalkable = true
		return
	}
	if strings.Contains(lower, "door_locked") || strings.Contains(lower, "lockeddoor") {
		tile.DoorState = 3 // locked
		tile.IsWalkable = false
	}
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

// DeleteEntityAt removes an entity at the given tile coordinates if one exists.
func (e *Editor) DeleteEntityAt(tx, ty int) {
	if e.level == nil || e.level.Entities == nil {
		return
	}

	// Find and remove entity at this position
	for i, ent := range e.level.Entities {
		if ent.X == tx && ent.Y == ty {
			// Remove this entity by slicing
			e.level.Entities = append(e.level.Entities[:i], e.level.Entities[i+1:]...)
			return
		}
	}
}

// NextLayer switches the editor to the next layer if a layered level is loaded.
func (e *Editor) NextLayer() {
	if e.layered == nil || len(e.layered.Layers) == 0 {
		return
	}
	e.layerIndex = (e.layerIndex + 1) % len(e.layered.Layers)
	e.layered.ActiveIndex = e.layerIndex
	e.level = e.layered.ActiveLayer()
	if e.OnLayerChange != nil {
		e.OnLayerChange(e.level)
	}
}

// PrevLayer switches the editor to the previous layer.
func (e *Editor) PrevLayer() {
	if e.layered == nil || len(e.layered.Layers) == 0 {
		return
	}
	e.layerIndex--
	if e.layerIndex < 0 {
		e.layerIndex = len(e.layered.Layers) - 1
	}
	e.layered.ActiveIndex = e.layerIndex
	e.level = e.layered.ActiveLayer()
	if e.OnLayerChange != nil {
		e.OnLayerChange(e.level)
	}
}

// SetLayeredLevel replaces the editor's layered level and refreshes its active layer.
func (e *Editor) SetLayeredLevel(ll *levels.LayeredLevel) {
	if ll == nil {
		return
	}
	e.layered = ll
	e.layerIndex = ll.ActiveIndex
	e.level = ll.ActiveLayer()
	if e.OnLayerChange != nil {
		e.OnLayerChange(e.level)
	}
}

// SetActiveLayer changes the editor's active layer by index.
func (e *Editor) SetActiveLayer(idx int) {
	if e.layered == nil || idx < 0 || idx >= len(e.layered.Layers) {
		return
	}
	e.layerIndex = idx
	e.layered.ActiveIndex = idx
	e.level = e.layered.ActiveLayer()
	if e.OnLayerChange != nil {
		e.OnLayerChange(e.level)
	}
}

// SetActiveLayerSilently updates the active layer without invoking the
// OnLayerChange callback. This avoids recursive updates when the game code is
// the one triggering the layer change.
func (e *Editor) SetActiveLayerSilently(idx int) {
	if e.layered == nil || idx < 0 || idx >= len(e.layered.Layers) {
		return
	}
	e.layerIndex = idx
	e.layered.ActiveIndex = idx
	e.level = e.layered.ActiveLayer()
}

// DrawModeButtons draws the entity mode buttons at the top-left of the screen
func (e *Editor) DrawModeButtons(screen *ebiten.Image) {
	buttonY := 5
	buttonSpacing := 5
	buttonW := 80
	buttonH := 20
	x := 5

	// Spawner Mode button
	spawnerRect := image.Rect(x, buttonY, x+buttonW, buttonY+buttonH)
	spawnerColor := color.RGBA{100, 100, 100, 200}
	if e.EntityMode == ModeSpawner {
		spawnerColor = color.RGBA{0, 200, 0, 200}
	}
	vector.DrawFilledRect(screen, float32(spawnerRect.Min.X), float32(spawnerRect.Min.Y),
		float32(spawnerRect.Dx()), float32(spawnerRect.Dy()), spawnerColor, false)
	ebitenutil.DebugPrintAt(screen, "Spawn", spawnerRect.Min.X+10, spawnerRect.Min.Y+4)

	x += buttonW + buttonSpacing

	// Delete Mode button
	deleteRect := image.Rect(x, buttonY, x+buttonW, buttonY+buttonH)
	deleteColor := color.RGBA{100, 100, 100, 200}
	if e.EntityMode == ModeDelete {
		deleteColor = color.RGBA{200, 0, 0, 200}
	}
	vector.DrawFilledRect(screen, float32(deleteRect.Min.X), float32(deleteRect.Min.Y),
		float32(deleteRect.Dx()), float32(deleteRect.Dy()), deleteColor, false)
	ebitenutil.DebugPrintAt(screen, "Delete", deleteRect.Min.X+5, deleteRect.Min.Y+4)

	x += buttonW + buttonSpacing*3

	// Clear Selection button
	clearRect := image.Rect(x, buttonY, x+90, buttonY+buttonH)
	vector.DrawFilledRect(screen, float32(clearRect.Min.X), float32(clearRect.Min.Y),
		float32(clearRect.Dx()), float32(clearRect.Dy()), color.RGBA{50, 50, 50, 200}, false)

	if e.SelectedEntityID != "" {
		ebitenutil.DebugPrintAt(screen, "Clear ["+e.SelectedEntityID+"]", clearRect.Min.X+5, clearRect.Min.Y+4)
	} else {
		ebitenutil.DebugPrintAt(screen, "Clear", clearRect.Min.X+30, clearRect.Min.Y+4)
	}

	// Store button rects on editor for click detection in Update
	e.spawnerButtonRect = spawnerRect
	e.deleteButtonRect = deleteRect
	e.clearButtonRect = clearRect
}
