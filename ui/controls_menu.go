package ui

import (
	"dungeoneer/controls"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ControlsMenu shows all key bindings and allows remapping
type ControlsMenu struct {
	visible        bool
	ctrl           *controls.Controls
	rect           image.Rectangle
	actions        []controls.ActionID
	selectedIndex  int
	remappingIndex int
	isRemapping    bool
	scrollOffset   int
	itemsPerPage   int
	style          MenuStyle
	onCancel       func()
}

// NewControlsMenu creates a new controls menu
func NewControlsMenu(w, h int, ctrl *controls.Controls, onCancel func()) *ControlsMenu {
	mw, mh := 600, 500
	mx := (w - mw) / 2
	my := (h - mh) / 2

	cm := &ControlsMenu{
		visible:        false,
		ctrl:           ctrl,
		rect:           image.Rect(mx, my, mx+mw, my+mh),
		actions:        controls.GetAllActionIDs(),
		selectedIndex:  0,
		remappingIndex: -1,
		isRemapping:    false,
		scrollOffset:   0,
		itemsPerPage:   14,
		style:          DefaultMenuStyles(),
		onCancel:       onCancel,
	}

	return cm
}

func (cm *ControlsMenu) Show()           { cm.visible = true }
func (cm *ControlsMenu) Hide()           { cm.visible = false }
func (cm *ControlsMenu) IsVisible() bool { return cm.visible }

func (cm *ControlsMenu) Update() {
	if !cm.visible {
		return
	}

	// If remapping, listen for key press
	if cm.isRemapping {
		// Get any key pressed
		for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
			if inpututil.IsKeyJustPressed(key) {
				if key != ebiten.KeyEscape { // Don't allow Escape as a binding
					action := cm.actions[cm.remappingIndex]
					cm.ctrl.SetBinding(action, key)
					// Save bindings to file immediately
					cm.ctrl.SaveBindings()
					cm.isRemapping = false
				} else {
					cm.isRemapping = false
				}
				return
			}
		}
		return
	}

	// Navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		cm.Hide()
		if cm.onCancel != nil {
			cm.onCancel()
		}
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		cm.selectedIndex--
		if cm.selectedIndex < 0 {
			cm.selectedIndex = len(cm.actions) - 1
		}
		cm.ensureVisible()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		cm.selectedIndex++
		if cm.selectedIndex >= len(cm.actions) {
			cm.selectedIndex = 0
		}
		cm.ensureVisible()
	}

	// Enter to remap
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		cm.remappingIndex = cm.selectedIndex
		cm.isRemapping = true
	}

	// R to reset
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		action := cm.actions[cm.selectedIndex]
		cm.ctrl.ResetBinding(action)
		cm.ctrl.SaveBindings()
	}
}

func (cm *ControlsMenu) ensureVisible() {
	if cm.selectedIndex < cm.scrollOffset {
		cm.scrollOffset = cm.selectedIndex
	}
	if cm.selectedIndex >= cm.scrollOffset+cm.itemsPerPage {
		cm.scrollOffset = cm.selectedIndex - cm.itemsPerPage + 1
	}
}

func (cm *ControlsMenu) Draw(screen *ebiten.Image) {
	if !cm.visible {
		return
	}

	// Draw overlay
	DrawMenuOverlay(screen, cm.style.overlayColor)

	// Draw window
	DrawMenuWindow(screen, &cm.style, float32(cm.rect.Min.X), float32(cm.rect.Min.Y),
		float32(cm.rect.Dx()), float32(cm.rect.Dy()))

	// Draw title
	ebitenutil.DebugPrintAt(screen, "CONTROLS", cm.rect.Min.X+20, cm.rect.Min.Y+15)

	// Draw instructions
	instructionsY := cm.rect.Min.Y + 40
	ebitenutil.DebugPrintAt(screen, "↑/↓ Navigate | Enter/Space Remap | R Reset | Esc Cancel", cm.rect.Min.X+15, instructionsY)

	// Draw control items
	itemHeight := 25
	startY := instructionsY + 30
	contentX := cm.rect.Min.X + 20
	contentWidth := cm.rect.Dx() - 40

	end := cm.scrollOffset + cm.itemsPerPage
	if end > len(cm.actions) {
		end = len(cm.actions)
	}

	for i := cm.scrollOffset; i < end; i++ {
		action := cm.actions[i]
		itemY := startY + (i-cm.scrollOffset)*itemHeight

		// Highlight selected
		if i == cm.selectedIndex {
			vector.DrawFilledRect(screen, float32(contentX), float32(itemY-5),
				float32(contentWidth), float32(itemHeight), color.RGBA{100, 100, 150, 100}, false)

			if cm.isRemapping {
				ebitenutil.DebugPrintAt(screen, "Press a key...", contentX+300, itemY)
			}
		}

		// Draw action label
		label := controls.GetActionLabel(action)
		ebitenutil.DebugPrintAt(screen, label, contentX, itemY)

		// Draw current binding
		binding := cm.ctrl.GetBinding(action)
		keyName := controls.GetKeyName(binding.Primary)
		ebitenutil.DebugPrintAt(screen, keyName, contentX+280, itemY)
	}

	// Draw scroll indicator
	if len(cm.actions) > cm.itemsPerPage {
		scrollIndicator := fmt.Sprintf("%d/%d", cm.scrollOffset/cm.itemsPerPage+1, (len(cm.actions)+cm.itemsPerPage-1)/cm.itemsPerPage)
		ebitenutil.DebugPrintAt(screen, scrollIndicator, cm.rect.Max.X-60, cm.rect.Max.Y-20)
	}
}
