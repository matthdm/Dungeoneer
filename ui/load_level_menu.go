package ui

import (
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

type LoadLevelMenu struct {
	Menu        *Menu
	OnLevelLoad func(*levels.Level) // Callback to set loaded level
	OnCancel    func()              // Callback to return to previous menu
}

func NewLoadLevelMenu(w, h int, onLoad func(*levels.Level), onCancel func()) *LoadLevelMenu {
	menuStyle := DefaultMenuStyles()

	menuWidth := 400
	menuHeight := 400
	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	menuRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)

	menu := NewMenu(menuRect, "LOAD LEVEL", nil, menuStyle)
	menu.SetInstructions([]string{"W/S/Arrow keys to navigate", "Enter/Space to select", "Esc to cancel"})

	llm := &LoadLevelMenu{
		Menu:        menu,
		OnLevelLoad: onLoad,
		OnCancel:    onCancel,
	}
	llm.populateMenuOptions()
	return llm
}

func (llm *LoadLevelMenu) populateMenuOptions() {
	files, err := listSavedLevels("levels")
	if err != nil {
		llm.Menu.options = []MenuOption{
			{Text: "Error: " + err.Error(), Action: func() {}},
			{Text: "Back", Action: llm.OnCancel},
		}
		return
	}

	var options []MenuOption
	for _, file := range files {
		filename := file // capture for closure
		options = append(options, MenuOption{
			Text: filename,
			Action: func() {
				level, err := leveleditor.LoadLevelFromFile(filepath.Join("levels", filename))
				if err != nil {
					fmt.Println("Failed to load level:", err)
					return
				}
				if llm.OnLevelLoad != nil {
					llm.OnLevelLoad(level)
				}
				llm.Menu.Hide()
			},
		})
	}

	options = append(options, MenuOption{
		Text:   "Back",
		Action: llm.OnCancel,
	})

	llm.Menu.options = options
	llm.Menu.SetSelectedIndex(0)
}

func (llm *LoadLevelMenu) Show() {
	llm.populateMenuOptions() // refresh file list each time
	llm.Menu.Show()
}

func (llm *LoadLevelMenu) Hide() {
	llm.Menu.Hide()
}

func (llm *LoadLevelMenu) Update() {
	llm.Menu.Update()
}

func (llm *LoadLevelMenu) Draw(screen *ebiten.Image) {
	llm.Menu.Draw(screen)
}

func listSavedLevels(dir string) ([]string, error) {
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
func (llm *LoadLevelMenu) SetRect(newRect image.Rectangle) {
	llm.Menu.rect = newRect
}
