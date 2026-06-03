package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// EchoEntry is the display model for one stored echo run.
type EchoEntry struct {
	Path       string
	RunIndex   int
	DeathCause string
	DeathFloor int
}

// EchoShrine is the hub UI for viewing and banishing stored echo runs.
type EchoShrine struct {
	Active      bool
	Entries     []EchoEntry
	SelectedIdx int
	screenW     int
	screenH     int
	OnBanish    func(path string) // called when player banishes an echo
}

// NewEchoShrine creates a new, closed EchoShrine sized to the given screen dimensions.
func NewEchoShrine(w, h int) *EchoShrine {
	return &EchoShrine{screenW: w, screenH: h}
}

// Open shows the shrine with the given entries.
func (e *EchoShrine) Open(entries []EchoEntry) {
	e.Entries = entries
	e.Active = true
	e.SelectedIdx = 0
}

// Close hides the shrine.
func (e *EchoShrine) Close() { e.Active = false }

// Resize updates screen dimensions (call on window resize).
func (e *EchoShrine) Resize(w, h int) { e.screenW = w; e.screenH = h }

// Update handles input. Call only when shrine is active.
func (e *EchoShrine) Update() {
	if !e.Active {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		e.Active = false
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && e.SelectedIdx < len(e.Entries)-1 {
		e.SelectedIdx++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && e.SelectedIdx > 0 {
		e.SelectedIdx--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) && len(e.Entries) > 0 {
		if e.OnBanish != nil {
			e.OnBanish(e.Entries[e.SelectedIdx].Path)
		}
	}
}

// Draw renders the echo shrine UI.
func (e *EchoShrine) Draw(screen *ebiten.Image) {
	if !e.Active {
		return
	}
	w, h := e.screenW, e.screenH

	bg := ebiten.NewImage(w, h)
	bg.Fill(color.NRGBA{8, 5, 18, 240})
	screen.DrawImage(bg, &ebiten.DrawImageOptions{})

	ebitenutil.DebugPrintAt(screen, "ECHO SHRINE  [Esc] Close  [Up/Down] Select  [X] Banish", w/2-120, 12)

	if len(e.Entries) == 0 {
		ebitenutil.DebugPrintAt(screen, "No echoes recorded yet. Complete a run to leave an echo.", w/2-150, h/2)
		return
	}

	y := 36
	for i, entry := range e.Entries {
		rowH := 40
		rowBg := ebiten.NewImage(w-40, rowH-2)
		if i == e.SelectedIdx {
			rowBg.Fill(color.NRGBA{80, 40, 120, 200})
		} else {
			rowBg.Fill(color.NRGBA{25, 15, 40, 180})
		}
		rop := &ebiten.DrawImageOptions{}
		rop.GeoM.Translate(20, float64(y))
		screen.DrawImage(rowBg, rop)

		label := fmt.Sprintf("Run %d | Floor %d | %s", entry.RunIndex, entry.DeathFloor, entry.DeathCause)
		if entry.DeathCause == "" {
			label = fmt.Sprintf("Run %d | Floor %d | victory", entry.RunIndex, entry.DeathFloor)
		}
		ebitenutil.DebugPrintAt(screen, label, 28, y+4)
		if i == e.SelectedIdx {
			ebitenutil.DebugPrintAt(screen, "[X] Banish this echo", 28, y+18)
		}
		y += rowH
		if y > h-20 {
			break
		}
	}
}
