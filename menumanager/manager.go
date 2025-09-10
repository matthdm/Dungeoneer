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
	stack     []Menu
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
	if menu == nil {
		return
	}
	if len(mm.stack) > 0 {
		mm.stack[len(mm.stack)-1].Hide()
	}
	mm.stack = append(mm.stack, menu)
	menu.Show()
}

func (mm *MenuManager) CloseActiveMenu() {
	if len(mm.stack) == 0 {
		return
	}
	top := mm.stack[len(mm.stack)-1]
	top.Hide()
	mm.stack = mm.stack[:len(mm.stack)-1]
	if len(mm.stack) > 0 {
		mm.stack[len(mm.stack)-1].Show()
	}
}

func (mm *MenuManager) HandleEscapePress() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return
	}
	if len(mm.stack) > 0 {
		mm.CloseActiveMenu()
	} else if mm.PauseMenu != nil {
		mm.Open(mm.PauseMenu)
	}
}

func (mm *MenuManager) IsMenuOpen() bool {
	return len(mm.stack) > 0 && mm.stack[len(mm.stack)-1].IsVisible()
}

func (mm *MenuManager) Active() Menu {
	if len(mm.stack) == 0 {
		return nil
	}
	return mm.stack[len(mm.stack)-1]
}
