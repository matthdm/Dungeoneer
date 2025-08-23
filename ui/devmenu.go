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
	list    *ScrollList
	itemIDs []string
	player  *entities.Player
	hint    func(string)
}

// NewDevMenu creates a new dev menu centered on the screen.
func NewDevMenu(w, h int, p *entities.Player, hint func(string)) *DevMenu {
	dm := &DevMenu{player: p, hint: hint}
	rect := image.Rect(w/2-150, h/2-150, w/2+150, h/2+150)
	dm.list = NewScrollList(rect, "Spawn Item", nil, DefaultMenuStyles())
	dm.list.SetInstructions([]string{"F2 Toggle", "Enter Spawn"})
	dm.refreshItemIDs()
	dm.list.SetOptions(dm.buildOptions())
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
	options := make([]MenuOption, 0, len(dm.itemIDs)+1)
	for _, id := range dm.itemIDs {
		name := items.Registry[id].Name
		itemID := id
		options = append(options, MenuOption{
			Text: name,
			Action: func() {
				if dm.player != nil {
					if !dm.player.AddToInventory(items.NewItem(itemID)) && dm.hint != nil {
						dm.hint("Inventory full")
					}
				}
			},
		})
	}
	options = append(options, MenuOption{Text: "Close", Action: func() { dm.list.Hide() }})
	return options
}

// Update handles input and menu logic.
func (dm *DevMenu) Update() {
	if !constants.DebugMode {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		dm.list.ToggleVisibility()
	}
	if dm.list.IsVisible() {
		dm.list.Update()
	}
}

// Draw renders the dev menu.
func (dm *DevMenu) Draw(screen *ebiten.Image) {
	if !constants.DebugMode {
		return
	}
	if dm.list.IsVisible() {
		dm.list.Draw(screen)
	}
}
