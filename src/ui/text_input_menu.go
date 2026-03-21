package ui

import (
	"image"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type TextInputMenu struct {
	rect         image.Rectangle
	Title        string
	Prompt       string
	input        string
	visible      bool
	onSubmit     func(text string)
	onCancel     func()
	Instructions []string
}

func NewTextInputMenu(rect image.Rectangle, title, prompt string, onSubmit func(string), onCancel func()) *TextInputMenu {
	return &TextInputMenu{
		rect:     rect,
		Title:    title,
		Prompt:   prompt,
		onSubmit: onSubmit,
		onCancel: onCancel,
		Instructions: []string{
			"Type and press Enter to confirm",
			"Esc to cancel",
		},
	}
}

func (t *TextInputMenu) Update() {
	if !t.visible {
		return
	}

	// Accept letters, digits, dashes, underscores, dots
	for _, r := range ebiten.InputChars() {
		if r == '\n' || r == '\r' {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-_.", r) {
			t.input += string(r)
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(t.input) > 0 {
		t.input = t.input[:len(t.input)-1]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		t.Hide()
		if t.onCancel != nil {
			t.onCancel()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(t.input) > 0 {
		if t.Title != "NEW LAYER" {
			if !isValidFilename(t.input) {
				t.Instructions = []string{"Invalid filename. Use a-z, 0-9, _, -, and end with .json"}
				return
			}
		}

		t.Hide()
		if t.onSubmit != nil {
			t.onSubmit(t.input)
		}
	}
}

func (t *TextInputMenu) Draw(screen *ebiten.Image) {
	if !t.visible {
		return
	}
	style := DefaultMenuStyles()
	DrawMenuOverlay(screen, DefaultOverlayColor)
	DrawMenuWindow(screen, &style, float32(t.rect.Min.X), float32(t.rect.Min.Y), float32(t.rect.Dx()), float32(t.rect.Dy()))

	ebitenutil.DebugPrintAt(screen, t.Title, t.rect.Min.X+20, t.rect.Min.Y+20)
	ebitenutil.DebugPrintAt(screen, t.Prompt, t.rect.Min.X+20, t.rect.Min.Y+60)
	ebitenutil.DebugPrintAt(screen, t.input+"|", t.rect.Min.X+20, t.rect.Min.Y+80)

	for i, inst := range t.Instructions {
		ebitenutil.DebugPrintAt(screen, inst, t.rect.Min.X+20, t.rect.Max.Y-40+i*15)
	}
}

func (t *TextInputMenu) Show()           { t.visible = true }
func (t *TextInputMenu) Hide()           { t.visible = false }
func (t *TextInputMenu) IsVisible() bool { return t.visible }
func (t *TextInputMenu) SetRect(r image.Rectangle) {
	t.rect = r
}
