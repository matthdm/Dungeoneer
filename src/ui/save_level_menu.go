package ui

import (
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

type SaveLevelMenu struct {
	Menu     *Menu
	Level    *levels.Level
	OnCancel func()
}

func NewSaveLevelMenu(w, h int, level *levels.Level, onCancel func()) *SaveLevelMenu {
	menuStyle := DefaultMenuStyles()
	menuWidth := 400
	menuHeight := 300
	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	menuRect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)

	sm := &SaveLevelMenu{
		Level:    level,
		OnCancel: onCancel,
	}

	options := []MenuOption{
		{Text: "Save as level1.json", Action: sm.saveAs("level1.json")},
		{Text: "Save as level2.json", Action: sm.saveAs("level2.json")},
		{Text: "Cancel", Action: sm.OnCancel},
	}

	menu := NewMenu(menuRect, "SAVE LEVEL", options, menuStyle)
	menu.SetInstructions([]string{"Choose filename", "Enter/Space to confirm", "Esc to cancel"})
	sm.Menu = menu
	return sm
}

func (sm *SaveLevelMenu) saveAs(filename string) func() {
	return func() {
		path := filepath.Join("levels", filename)
		err := leveleditor.SaveLevelToFile(sm.Level, path)
		if err != nil {
			fmt.Println("Error saving level:", err)
			return
		}
		fmt.Println("Saved level:", filename)
		sm.Menu.Hide()
	}
}

func (sm *SaveLevelMenu) Show()                     { sm.Menu.Show() }
func (sm *SaveLevelMenu) Hide()                     { sm.Menu.Hide() }
func (sm *SaveLevelMenu) IsVisible() bool           { return sm.Menu.IsVisible() }
func (sm *SaveLevelMenu) Update()                   { sm.Menu.Update() }
func (sm *SaveLevelMenu) Draw(screen *ebiten.Image) { sm.Menu.Draw(screen) }
func (sm *SaveLevelMenu) SetRect(r image.Rectangle) { sm.Menu.SetRect(r) }
