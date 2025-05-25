package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Menu struct {
	rect                image.Rectangle
	title               string
	options             []MenuOption
	selectedOptionIndex int
	visible             bool
	style               MenuStyle
	instructions        []string // Optional: for custom instructions below options
}

// MenuOption defines a single item in a Menu
// The Action is a function that is called when selected
type MenuOption struct {
	Text   string
	Action func()
}

// MenuStyle defines the visual appearance of a Menu
type MenuStyle struct {
	overlayColor            color.Color
	backgroundColor         color.Color
	borderColor             color.Color
	titleColor              color.Color
	optionTextColor         color.Color
	selectedOptionTextColor color.Color
	selectionBGColor        color.Color
	borderThickness         float32
	titleHeight             float32 // Space reserved for title from top of menu
	optionTextHeight        float32 // Height of each option line
	optionPaddingX          float32 // Horizontal padding for option text from menu edge
	optionTextOffsetX       float32 // Small offset for text like "> " or "  "
	optionInternalPaddingY  float32 // Vertical padding for selection background around text
	instructionTextHeight   float32 // Space at the bottom for instructions
}

func NewMenu(rect image.Rectangle, title string, options []MenuOption, style MenuStyle) *Menu {
	m := &Menu{
		rect:                rect,
		title:               title,
		options:             options,
		selectedOptionIndex: 0,
		visible:             false,
		style:               style,
	}
	// Ensure selected index is valid if options are initially empty
	if len(options) == 0 {
		m.selectedOptionIndex = -1
	}
	return m
}

// Update handles input for the menu (navigation, selection)
func (m *Menu) Update() {
	if !m.visible || len(m.options) == 0 {
		return
	}

	// Keyboard navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedOptionIndex--
		if m.selectedOptionIndex < 0 {
			m.selectedOptionIndex = len(m.options) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedOptionIndex++
		if m.selectedOptionIndex >= len(m.options) {
			m.selectedOptionIndex = 0
		}
	}

	// Mouse navigation and selection
	mouseX, mouseY := ebiten.CursorPosition()
	optionsContentStartY := float32(m.rect.Min.Y) + m.style.titleHeight
	clickableRegionXStart := float32(m.rect.Min.X) + m.style.optionPaddingX/2.0
	clickableRegionWidth := float32(m.rect.Dx()) - m.style.optionPaddingX

	for i := range m.options {
		optionPosY := optionsContentStartY + float32(i)*m.style.optionTextHeight
		mouseHoverMinY := optionPosY - m.style.optionInternalPaddingY
		mouseHoverMaxY := optionPosY - m.style.optionInternalPaddingY + m.style.optionTextHeight

		if float32(mouseX) >= clickableRegionXStart && float32(mouseX) <= clickableRegionXStart+clickableRegionWidth &&
			float32(mouseY) >= mouseHoverMinY && float32(mouseY) <= mouseHoverMaxY {
			m.selectedOptionIndex = i
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if m.options[i].Action != nil {
					m.options[i].Action()
					return // Action might hide menu or change state
				}
			}
			break
		}
	}

	// Keyboard selection
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if m.selectedOptionIndex >= 0 && m.selectedOptionIndex < len(m.options) && m.options[m.selectedOptionIndex].Action != nil {
			m.options[m.selectedOptionIndex].Action()
			return // Action might hide menu or change state
		}
	}
}

func (m *Menu) Draw(screen *ebiten.Image) {
	if !m.visible {
		return
	}

	DrawMenuOverlay(screen, m.style.overlayColor)

	menuX, menuY := float32(m.rect.Min.X), float32(m.rect.Min.Y)
	menuW, menuH := float32(m.rect.Dx()), float32(m.rect.Dy())

	DrawMenuWindow(screen, m, menuX, menuY, menuW, menuH)
	DrawMenuTitleText(screen, m.title, menuX, menuY, menuW, menuH)
	DrawMenuOptions(screen, m, menuX, menuY, menuW)

	// Draw instructions
	if len(m.instructions) > 0 {
		instructionPosY := menuY + menuH - m.style.instructionTextHeight
		instructionPosX := menuX + m.style.optionPaddingX
		for i, line := range m.instructions {
			ebitenutil.DebugPrintAt(screen, line, int(instructionPosX), int(instructionPosY+float32(i*15))) // 15px line spacing
		}
	}
}

func (m *Menu) Show()             { m.visible = true }
func (m *Menu) Hide()             { m.visible = false }
func (m *Menu) IsVisible() bool   { return m.visible }
func (m *Menu) ToggleVisibility() { m.visible = !m.visible }

func (m *Menu) SetSelectedIndex(index int) {
	m.selectedOptionIndex = index
}
func (m *Menu) SetRect(newRect image.Rectangle) {
	m.rect = newRect
}
func (m *Menu) SetInstructions(newInstructions []string) {
	m.instructions = newInstructions
}
