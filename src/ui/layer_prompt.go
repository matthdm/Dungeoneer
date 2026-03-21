package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type LayerPrompt struct {
	dimsInput  *TextInputMenu
	flavorMenu *Menu
	flavors    []string
	width      int
	height     int
	onCreate   func(w, h int, flavor string)
	onCancel   func()
}

func NewLayerPrompt(w, h int, flavors []string, onCreate func(int, int, string), onCancel func()) *LayerPrompt {
	rect := image.Rect(w/2-150, h/2-100, w/2+150, h/2+100)
	lp := &LayerPrompt{
		flavors:  flavors,
		onCreate: onCreate,
		onCancel: onCancel,
	}
	lp.dimsInput = NewTextInputMenu(rect, "NEW LAYER", "Enter WxH (e.g. 64x64):", lp.handleDims, lp.cancel)
	return lp
}

func (lp *LayerPrompt) handleDims(text string) {
	parts := strings.FieldsFunc(text, func(r rune) bool { return r == 'x' || r == ',' })
	if len(parts) != 2 {
		lp.dimsInput.Instructions = []string{"Format must be WxH"}
		lp.dimsInput.Show()
		lp.dimsInput.input = ""
		return
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		lp.dimsInput.Instructions = []string{"Numbers only, format WxH"}
		lp.dimsInput.Show()
		lp.dimsInput.input = ""
		return
	}
	lp.width = w
	lp.height = h
	style := DefaultMenuStyles()
	rect := lp.dimsInput.rect
	options := make([]MenuOption, len(lp.flavors)+1)
	for i, fl := range lp.flavors {
		f := fl
		options[i] = MenuOption{Text: strings.Title(fl), Action: func() { lp.selectFlavor(f) }}
	}
	options[len(lp.flavors)] = MenuOption{Text: "Cancel", Action: lp.cancel}
	lp.flavorMenu = NewMenu(rect, "SELECT FLAVOR", options, style)
	lp.flavorMenu.SetInstructions([]string{"W/S Navigate", "Enter/Space Select", "Esc Cancel"})
	lp.flavorMenu.Show()
}

func (lp *LayerPrompt) selectFlavor(fl string) {
	if lp.onCreate != nil {
		lp.onCreate(lp.width, lp.height, fl)
	}
	lp.hideAll()
}

func (lp *LayerPrompt) cancel() {
	lp.hideAll()
	if lp.onCancel != nil {
		lp.onCancel()
	}
}

func (lp *LayerPrompt) hideAll() {
	if lp.dimsInput != nil {
		lp.dimsInput.Hide()
	}
	if lp.flavorMenu != nil {
		lp.flavorMenu.Hide()
	}
}

func (lp *LayerPrompt) Show() { lp.dimsInput.Show() }
func (lp *LayerPrompt) IsVisible() bool {
	return (lp.dimsInput != nil && lp.dimsInput.IsVisible()) || (lp.flavorMenu != nil && lp.flavorMenu.IsVisible())
}

func (lp *LayerPrompt) Update() {
	if lp.dimsInput != nil && lp.dimsInput.IsVisible() {
		lp.dimsInput.Update()
		return
	}
	if lp.flavorMenu != nil && lp.flavorMenu.IsVisible() {
		lp.flavorMenu.Update()
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			lp.cancel()
		}
	}
}

func (lp *LayerPrompt) Draw(screen *ebiten.Image) {
	if lp.dimsInput != nil && lp.dimsInput.IsVisible() {
		lp.dimsInput.Draw(screen)
		return
	}
	if lp.flavorMenu != nil && lp.flavorMenu.IsVisible() {
		lp.flavorMenu.Draw(screen)
	}
}
