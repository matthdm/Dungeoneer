package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Default colors and constants for styling
var (
	DefaultOverlayColor     = color.RGBA{0, 0, 0, 128}
	DefaultBackgroundColor  = color.RGBA{10, 10, 10, 255}
	DefaultBorderColor      = color.RGBA{150, 20, 15, 255}
	DefaultTextColor        = color.White
	DefaultSelectionBGColor = color.RGBA{80, 75, 70, 255}
)

const (
	DefaultBorderThickness       float32 = 3
	DefaultTitleHeight           float32 = 40 // Space reserved for title from top of menu
	DefaultOptionTextHeight      float32 = 35 // Height of each option line
	DefaultOptionPaddingX        float32 = 20 // Horizontal padding for option text from menu edge
	DefaultOptionTextOffsetX     float32 = 5  // Small offset for text like "> " or "  "
	DefaultOptionInternalPadY    float32 = 5  // Vertical padding for selection background around text
	DefaultInstructionTextHeight float32 = 60 // Space at the bottom for instructions
)

func DefaultMenuStyles() MenuStyle {
	return MenuStyle{
		overlayColor:            DefaultOverlayColor,
		backgroundColor:         DefaultBackgroundColor,
		borderColor:             DefaultBorderColor,
		titleColor:              DefaultTextColor,
		optionTextColor:         DefaultTextColor,
		selectedOptionTextColor: DefaultTextColor,
		selectionBGColor:        DefaultSelectionBGColor,
		borderThickness:         DefaultBorderThickness,
		titleHeight:             DefaultTitleHeight,
		optionTextHeight:        DefaultOptionTextHeight,
		optionPaddingX:          DefaultOptionPaddingX,
		optionTextOffsetX:       DefaultOptionTextOffsetX,
		optionInternalPaddingY:  DefaultOptionInternalPadY,
		instructionTextHeight:   DefaultInstructionTextHeight,
	}
}

// Draws a semi-transparent overlay over the entire screen
func DrawMenuOverlay(screen *ebiten.Image, overlayColor color.Color) {
	screenBounds := screen.Bounds()
	vector.DrawFilledRect(screen, 0, 0, float32(screenBounds.Dx()), float32(screenBounds.Dy()), overlayColor, false)
}

// Draws a basic window with a background and border
func DrawMenuWindow(screen *ebiten.Image, m *Menu, menuX, menuY, menuW, menuH float32) {
	vector.DrawFilledRect(screen, menuX, menuY, menuW, menuH, m.style.backgroundColor, false)
	vector.StrokeRect(screen, menuX, menuY, menuW, menuH, m.style.borderThickness, m.style.borderColor, false)

}

// Draws the vertically-centered, top-anchored title text
func DrawMenuTitleText(screen *ebiten.Image, title string, menuX, menuY, menuW, menuH float32) {
	if title != "" {
		titleTextX := menuX + (menuW-float32(len(title)*8))/2 // Approx centering
		titleTextY := menuY + 10
		ebitenutil.DebugPrintAt(screen, title, int(titleTextX), int(titleTextY))
	}
}

// Draws the list of selectable options
func DrawMenuOptions(screen *ebiten.Image, m *Menu,
	menuX, menuY, menuW float32) {

	if len(m.options) > 0 && m.selectedOptionIndex != -1 {
		optionsContentStartY := menuY + m.style.titleHeight
		for i, option := range m.options {
			// Y position for the start of the text line
			textPosY := optionsContentStartY + float32(i)*m.style.optionTextHeight
			// X position for the start of the text ("  " or "> ")
			textPosX := menuX + m.style.optionPaddingX + m.style.optionTextOffsetX

			textToShow := "  " + option.Text
			if i == m.selectedOptionIndex {
				textToShow = "> " + option.Text
				bgX := menuX + m.style.optionPaddingX/2
				bgY := textPosY - m.style.optionInternalPaddingY
				bgW := menuW - m.style.optionPaddingX
				bgH := m.style.optionTextHeight
				vector.DrawFilledRect(screen, bgX, bgY, bgW, bgH, m.style.selectionBGColor, false)
				ebitenutil.DebugPrintAt(screen, textToShow, int(textPosX), int(textPosY))
			} else {
				ebitenutil.DebugPrintAt(screen, textToShow, int(textPosX), int(textPosY))
			}
		}
	}

}
