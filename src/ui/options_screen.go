package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// OptionsScreen renders fullscreen toggle and volume slider.
// Key bindings are handled by ControlsMenu shown via onControls.
type OptionsScreen struct {
	Fullscreen  *bool
	Volume      *float64
	OnSave      func() // called whenever a setting changes (to persist)
	OnControls  func() // called when Controls row is activated
	OnBack      func()
	rect        image.Rectangle
	selectedRow int // 0=Fullscreen, 1=Volume, 2=Controls, 3=Back
	style       MenuStyle
}

func NewOptionsScreen(w, h int, fullscreen *bool, volume *float64, onSave, onControls, onBack func()) *OptionsScreen {
	sw, sh := 500, 300
	return &OptionsScreen{
		Fullscreen:  fullscreen,
		Volume:      volume,
		OnSave:      onSave,
		OnControls:  onControls,
		OnBack:      onBack,
		rect:        image.Rect((w-sw)/2, (h-sh)/2, (w+sw)/2, (h+sh)/2),
		selectedRow: 0,
		style:       DefaultMenuStyles(),
	}
}

func (os *OptionsScreen) Resize(w, h int) {
	sw, sh := 500, 300
	os.rect = image.Rect((w-sw)/2, (h-sh)/2, (w+sw)/2, (h+sh)/2)
}

func (os *OptionsScreen) Update() {
	// Row navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		os.selectedRow = (os.selectedRow + 1) % 4
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		os.selectedRow = (os.selectedRow + 3) % 4
	}

	// Activate selected row
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		os.activateRow(os.selectedRow)
	}

	// Shortcut: F toggles fullscreen, +/- adjust volume
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		*os.Fullscreen = !*os.Fullscreen
		if os.OnSave != nil {
			os.OnSave()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		*os.Volume = min(1.0, *os.Volume+0.1)
		if os.OnSave != nil {
			os.OnSave()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		*os.Volume = max(0.0, *os.Volume-0.1)
		if os.OnSave != nil {
			os.OnSave()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if os.OnBack != nil {
			os.OnBack()
		}
	}
}

func (os *OptionsScreen) activateRow(row int) {
	switch row {
	case 0:
		*os.Fullscreen = !*os.Fullscreen
		if os.OnSave != nil {
			os.OnSave()
		}
	case 1:
		// Volume row: Enter nudges volume up.
		*os.Volume = min(1.0, *os.Volume+0.1)
		if os.OnSave != nil {
			os.OnSave()
		}
	case 2:
		if os.OnControls != nil {
			os.OnControls()
		}
	case 3:
		if os.OnBack != nil {
			os.OnBack()
		}
	}
}

func (os *OptionsScreen) Draw(screen *ebiten.Image) {
	DrawMenuOverlay(screen, color.NRGBA{0, 0, 0, 180})
	DrawMenuWindow(screen, &os.style,
		float32(os.rect.Min.X), float32(os.rect.Min.Y),
		float32(os.rect.Dx()), float32(os.rect.Dy()))

	x := os.rect.Min.X + 30
	y := os.rect.Min.Y + 20

	ebitenutil.DebugPrintAt(screen, "OPTIONS", x, y)
	y += 30
	ebitenutil.DebugPrintAt(screen, "──────────────────────────────", x, y)
	y += 20

	// Static row labels
	rowLabels := [4]string{"Fullscreen", "Volume    ", "Controls  ", "Back      "}

	// Dynamic row values
	fullscreenVal := "[OFF]"
	if *os.Fullscreen {
		fullscreenVal = "[ON] "
	}
	rowValues := [4]string{
		fullscreenVal,
		volumeBar(*os.Volume),
		">",
		"ESC",
	}

	for i := 0; i < 4; i++ {
		prefix := "  "
		if i == os.selectedRow {
			prefix = "> "
			vector.DrawFilledRect(screen,
				float32(x-5), float32(y-2),
				float32(os.rect.Dx()-50), 18,
				color.RGBA{100, 100, 150, 80}, false)
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%-12s %s", prefix, rowLabels[i], rowValues[i]), x, y)
		y += 22
	}

	y += 10
	ebitenutil.DebugPrintAt(screen, "W/S Navigate | Enter Select | F Toggle Fullscreen | +/- Volume", x, y)
}

func volumeBar(v float64) string {
	const total = 10
	filled := int(v * total)
	if filled > total {
		filled = total
	}
	if filled < 0 {
		filled = 0
	}
	var buf [total + 2]byte
	buf[0] = '['
	for i := 0; i < total; i++ {
		if i < filled {
			buf[i+1] = '='
		} else {
			buf[i+1] = ' '
		}
	}
	buf[total+1] = ']'
	return string(buf[:])
}
