package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ScrollList displays a vertical list of menu options with scrolling support.
type ScrollList struct {
	rect         image.Rectangle
	title        string
	options      []MenuOption
	style        MenuStyle
	instructions []string
	visible      bool

	offset      int
	selected    int
	visibleRows int
}

// NewScrollList creates a new scrollable list.
func NewScrollList(rect image.Rectangle, title string, options []MenuOption, style MenuStyle) *ScrollList {
	sl := &ScrollList{rect: rect, title: title, style: style, options: options}
	sl.calcVisible()
	sl.selected = 0
	return sl
}

func (sl *ScrollList) calcVisible() {
	sl.visibleRows = int((float32(sl.rect.Dy()) - sl.style.titleHeight - sl.style.instructionTextHeight) / sl.style.optionTextHeight)
	if sl.visibleRows < 1 {
		sl.visibleRows = 1
	}
	if sl.offset > len(sl.options)-sl.visibleRows {
		sl.offset = len(sl.options) - sl.visibleRows
	}
	if sl.offset < 0 {
		sl.offset = 0
	}
}

// SetOptions replaces the option list.
func (sl *ScrollList) SetOptions(opts []MenuOption) {
	sl.options = opts
	if sl.selected >= len(sl.options) {
		sl.selected = len(sl.options) - 1
	}
	if sl.selected < 0 {
		sl.selected = 0
	}
	sl.calcVisible()
}

// SetInstructions defines footer instructions.
func (sl *ScrollList) SetInstructions(instr []string) { sl.instructions = instr }

func (sl *ScrollList) Show()             { sl.visible = true }
func (sl *ScrollList) Hide()             { sl.visible = false }
func (sl *ScrollList) ToggleVisibility() { sl.visible = !sl.visible }
func (sl *ScrollList) IsVisible() bool   { return sl.visible }

// Update handles input for selection and scrolling.
func (sl *ScrollList) Update() {
	if !sl.visible {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if sl.selected > 0 {
			sl.selected--
			if sl.selected < sl.offset {
				sl.offset = sl.selected
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if sl.selected < len(sl.options)-1 {
			sl.selected++
			if sl.selected >= sl.offset+sl.visibleRows {
				sl.offset = sl.selected - sl.visibleRows + 1
			}
		}
	}

	if _, wheel := ebiten.Wheel(); wheel != 0 {
		sl.offset -= int(wheel)
		if sl.offset < 0 {
			sl.offset = 0
		}
		if max := len(sl.options) - sl.visibleRows; sl.offset > max {
			sl.offset = max
		}
	}

	mx, my := ebiten.CursorPosition()
	optionsStartY := float32(sl.rect.Min.Y) + sl.style.titleHeight
	clickableRegionXStart := float32(sl.rect.Min.X) + sl.style.optionPaddingX/2.0
	clickableRegionWidth := float32(sl.rect.Dx()) - sl.style.optionPaddingX

	for i := 0; i < sl.visibleRows && sl.offset+i < len(sl.options); i++ {
		optionPosY := optionsStartY + float32(i)*sl.style.optionTextHeight
		mouseHoverMinY := optionPosY - sl.style.optionInternalPaddingY
		mouseHoverMaxY := optionPosY - sl.style.optionInternalPaddingY + sl.style.optionTextHeight
		if float32(mx) >= clickableRegionXStart && float32(mx) <= clickableRegionXStart+clickableRegionWidth &&
			float32(my) >= mouseHoverMinY && float32(my) <= mouseHoverMaxY {
			sl.selected = sl.offset + i
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if sl.options[sl.selected].Action != nil {
					sl.options[sl.selected].Action()
				}
			}
			break
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if sl.selected >= 0 && sl.selected < len(sl.options) {
			if sl.options[sl.selected].Action != nil {
				sl.options[sl.selected].Action()
			}
		}
	}
}

// Draw renders the scroll list.
func (sl *ScrollList) Draw(screen *ebiten.Image) {
	if !sl.visible {
		return
	}
	DrawMenuOverlay(screen, sl.style.overlayColor)
	menuX, menuY := float32(sl.rect.Min.X), float32(sl.rect.Min.Y)
	menuW, menuH := float32(sl.rect.Dx()), float32(sl.rect.Dy())
	DrawMenuWindow(screen, &sl.style, menuX, menuY, menuW, menuH)
	DrawMenuTitleText(screen, sl.title, menuX, menuY, menuW, menuH)

	optionsStartY := menuY + sl.style.titleHeight
	for i := 0; i < sl.visibleRows && sl.offset+i < len(sl.options); i++ {
		idx := sl.offset + i
		option := sl.options[idx]
		textPosY := optionsStartY + float32(i)*sl.style.optionTextHeight
		textPosX := menuX + sl.style.optionPaddingX + sl.style.optionTextOffsetX
		textToShow := "  " + option.Text
		if idx == sl.selected {
			textToShow = "> " + option.Text
			bgX := menuX + sl.style.optionPaddingX/2
			bgY := textPosY - sl.style.optionInternalPaddingY
			bgW := menuW - sl.style.optionPaddingX
			bgH := sl.style.optionTextHeight
			vector.DrawFilledRect(screen, bgX, bgY, bgW, bgH, sl.style.selectionBGColor, false)
		}
		ebitenutil.DebugPrintAt(screen, textToShow, int(textPosX), int(textPosY))
	}

	if len(sl.instructions) > 0 {
		instructionPosY := menuY + menuH - sl.style.instructionTextHeight
		instructionPosX := menuX + sl.style.optionPaddingX
		for i, line := range sl.instructions {
			ebitenutil.DebugPrintAt(screen, line, int(instructionPosX), int(instructionPosY+float32(i*15)))
		}
	}
}
