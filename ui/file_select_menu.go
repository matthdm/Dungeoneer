package ui

import (
	"dungeoneer/entities"
	"dungeoneer/leveleditor"
	"dungeoneer/levels"
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// FileSelectMenu is a generic menu for selecting JSON files from a directory.
type FileSelectMenu struct {
	Menu     *Menu
	dir      string
	onSelect func(string)
	onCancel func()
}

func NewFileSelectMenu(w, h int, title, dir string, onSelect func(string), onCancel func()) *FileSelectMenu {
	style := DefaultMenuStyles()
	menuWidth := 400
	menuHeight := 400
	menuX := (w - menuWidth) / 2
	menuY := (h - menuHeight) / 2
	rect := image.Rect(menuX, menuY, menuX+menuWidth, menuY+menuHeight)
	menu := NewMenu(rect, title, nil, style)
	menu.SetInstructions([]string{"W/S/Arrow keys to navigate", "Enter/Space to select", "Esc to cancel"})
	fsm := &FileSelectMenu{
		Menu:     menu,
		dir:      dir,
		onSelect: onSelect,
		onCancel: onCancel,
	}
	fsm.populate()
	return fsm
}

func (fsm *FileSelectMenu) populate() {
	files, err := listJSONFiles(fsm.dir)
	if err != nil {
		fsm.Menu.options = []MenuOption{{Text: "Error: " + err.Error(), Action: func() {}}, {Text: "Back", Action: fsm.onCancel}}
		return
	}
	var options []MenuOption
	for _, file := range files {
		filename := file
		options = append(options, MenuOption{
			Text: filename,
			Action: func() {
				if fsm.onSelect != nil {
					fsm.onSelect(filepath.Join(fsm.dir, filename))
				}
				fsm.Menu.Hide()
			},
		})
	}
	options = append(options, MenuOption{Text: "Back", Action: fsm.onCancel})
	fsm.Menu.options = options
	fsm.Menu.SetSelectedIndex(0)
}

func (fsm *FileSelectMenu) Show()                     { fsm.populate(); fsm.Menu.Show() }
func (fsm *FileSelectMenu) Hide()                     { fsm.Menu.Hide() }
func (fsm *FileSelectMenu) IsVisible() bool           { return fsm.Menu.IsVisible() }
func (fsm *FileSelectMenu) Update()                   { fsm.Menu.Update() }
func (fsm *FileSelectMenu) Draw(screen *ebiten.Image) { fsm.Menu.Draw(screen) }
func (fsm *FileSelectMenu) SetRect(r image.Rectangle) { fsm.Menu.rect = r }

// NewLoadLevelMenu creates a FileSelectMenu specialized for loading levels.
func NewLoadLevelMenu(w, h int, onLoad func(*levels.Level), onCancel func()) *FileSelectMenu {
	return NewFileSelectMenu(w, h, "LOAD LEVEL", "levels", func(path string) {
		level, err := leveleditor.LoadLevelFromFile(path)
		if err != nil {
			fmt.Println("Failed to load level:", err)
			return
		}
		if onLoad != nil {
			onLoad(level)
		}
	}, onCancel)
}

// NewLoadPlayerMenu creates a FileSelectMenu specialized for loading players.
func NewLoadPlayerMenu(w, h int, onLoad func(*entities.Player), onCancel func()) *FileSelectMenu {
	return NewFileSelectMenu(w, h, "LOAD PLAYER", "players", func(path string) {
		player, err := entities.LoadPlayerFromFile(path)
		if err != nil {
			fmt.Println("Failed to load player:", err)
			return
		}
		if onLoad != nil {
			onLoad(player)
		}
	}, onCancel)
}
