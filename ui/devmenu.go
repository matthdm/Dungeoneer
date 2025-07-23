package ui

import (
	"dungeoneer/constants"
	"dungeoneer/entities"
	"dungeoneer/items"
	"image"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// DevMenu provides a simple debug UI for spawning items.
type DevMenu struct {
	menu    *Menu
	itemIDs []string
	player  *entities.Player
}

// NewDevMenu creates a new dev menu centered on the screen.
func NewDevMenu(w, h int, p *entities.Player) *DevMenu {
	dm := &DevMenu{player: p}
	dm.refreshItemIDs()
	rect := image.Rect(w/2-150, h/2-150, w/2+150, h/2+150)
	dm.menu = NewMenu(rect, "Spawn Item", dm.buildOptions(), DefaultMenuStyles())
	dm.menu.SetInstructions([]string{"F2 Toggle", "Enter Spawn"})
	return dm
}

func (dm *DevMenu) refreshItemIDs() {
	dm.itemIDs = dm.itemIDs[:0]
	for id := range items.Registry {
		dm.itemIDs = append(dm.itemIDs, id)
	}
	sort.Strings(dm.itemIDs)
}

func (dm *DevMenu) buildOptions() []MenuOption {
	options := []MenuOption{}
	max := 10
	if len(dm.itemIDs) < max {
		max = len(dm.itemIDs)
	}
	for i := 0; i < max; i++ {
		id := dm.itemIDs[i]
		name := items.Registry[id].Name
		itemID := id
		options = append(options, MenuOption{
			Text: name,
			Action: func() {
				if dm.player != nil {
					dm.player.Inventory.AddItem(items.NewItem(itemID))
				}
			},
		})
	}
	options = append(options, MenuOption{Text: "Close", Action: func() { dm.menu.Hide() }})
	return options
}

// Update handles input and menu logic.
func (dm *DevMenu) Update() {
	if !constants.DebugMode {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		dm.menu.ToggleVisibility()
	}
	if dm.menu.IsVisible() {
		dm.menu.Update()
	}
}

// Draw renders the dev menu.
func (dm *DevMenu) Draw(screen *ebiten.Image) {
	if !constants.DebugMode {
		return
	}
	if dm.menu.IsVisible() {
		dm.menu.Draw(screen)
	}
}
