package ui

import (
	"dungeoneer/entities"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// LoadPlayerMenu allows selecting a saved player profile to load.
type LoadPlayerMenu struct {
	Menu         *Menu
	OnPlayerLoad func(*entities.Player)
	OnCancel     func()
}

func NewLoadPlayerMenu(w, h int, onLoad func(*entities.Player), onCancel func()) *LoadPlayerMenu {
	style := DefaultMenuStyles()
	menuWidth := 400
	menuHeight := 400
	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	rect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)

	menu := NewMenu(rect, "LOAD PLAYER", nil, style)
	menu.SetInstructions([]string{"W/S/Arrows Navigate", "Enter/Space Select", "Esc to cancel"})

	lpm := &LoadPlayerMenu{
		Menu:         menu,
		OnPlayerLoad: onLoad,
		OnCancel:     onCancel,
	}
	lpm.populateMenuOptions()
	return lpm
}

func (lpm *LoadPlayerMenu) populateMenuOptions() {
	files, err := listSavedProfiles("players")
	if err != nil {
		lpm.Menu.options = []MenuOption{
			{Text: "Error: " + err.Error(), Action: func() {}},
			{Text: "Back", Action: lpm.OnCancel},
		}
		return
	}
	var options []MenuOption
	for _, file := range files {
		filename := file
		options = append(options, MenuOption{
			Text: filename,
			Action: func() {
				player, err := entities.LoadPlayerFromFile(filepath.Join("players", filename))
				if err != nil {
					fmt.Println("Failed to load player:", err)
					return
				}
				if lpm.OnPlayerLoad != nil {
					lpm.OnPlayerLoad(player)
				}
				lpm.Menu.Hide()
			},
		})
	}
	options = append(options, MenuOption{Text: "Back", Action: lpm.OnCancel})
	lpm.Menu.options = options
	lpm.Menu.SetSelectedIndex(0)
}

func (lpm *LoadPlayerMenu) Show()                     { lpm.populateMenuOptions(); lpm.Menu.Show() }
func (lpm *LoadPlayerMenu) Hide()                     { lpm.Menu.Hide() }
func (lpm *LoadPlayerMenu) Update()                   { lpm.Menu.Update() }
func (lpm *LoadPlayerMenu) Draw(screen *ebiten.Image) { lpm.Menu.Draw(screen) }
func (lpm *LoadPlayerMenu) SetRect(r image.Rectangle) { lpm.Menu.rect = r }

func listSavedProfiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
