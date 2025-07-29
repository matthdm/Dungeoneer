package menumanager

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Menu interface {
	Show()
	Hide()
	IsVisible() bool
}

type MenuManager struct {
	Active    Menu
	PauseMenu Menu
}

var defaultManager *MenuManager

func Init(pause Menu) {
	defaultManager = &MenuManager{PauseMenu: pause}
}

func Manager() *MenuManager {
	if defaultManager == nil {
		defaultManager = &MenuManager{}
	}
	return defaultManager
}

func (mm *MenuManager) Open(menu Menu) {
	if mm.Active != nil && mm.Active != menu {
		mm.Active.Hide()
	}
	mm.Active = menu
	if menu != nil {
		menu.Show()
	}
}

func (mm *MenuManager) CloseActiveMenu() {
	if mm.Active != nil {
		mm.Active.Hide()
		mm.Active = nil
	}
}

func (mm *MenuManager) HandleEscapePress() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return
	}
	if mm.Active != nil {
		if mm.Active == mm.PauseMenu {
			mm.CloseActiveMenu()
		} else {
			mm.CloseActiveMenu()
		}
	} else if mm.PauseMenu != nil {
		mm.Open(mm.PauseMenu)
	}
}

func (mm *MenuManager) IsMenuOpen() bool {
	return mm.Active != nil && mm.Active.IsVisible()
}
