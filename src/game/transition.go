package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	transStateIdle    = 0
	transStateFadeOut = 1
	transStateHold    = 2
	transStateFadeIn  = 3
)

// Transition manages a fade-to-black screen transition.
type Transition struct {
	state      int
	timer      float64
	fadeOutDur float64
	holdDur    float64
	fadeInDur  float64
	OnMidpoint func() // called once when fade-out completes (do the actual floor change here)
	overlay    *ebiten.Image
}

// NewTransition creates a transition with a pre-allocated overlay image.
func NewTransition(screenW, screenH int) *Transition {
	return &Transition{
		overlay:    ebiten.NewImage(screenW, screenH),
		fadeOutDur: 0.4,
		holdDur:    0.05,
		fadeInDur:  0.4,
	}
}

// Resize recreates the overlay for the new screen size.
func (t *Transition) Resize(w, h int) {
	t.overlay = ebiten.NewImage(w, h)
}

// Start begins a fade-out → hold → fade-in cycle.
// onMidpoint is called once when the screen is fully black.
func (t *Transition) Start(onMidpoint func()) {
	t.state = transStateFadeOut
	t.timer = 0
	t.OnMidpoint = onMidpoint
}

// StartFadeIn begins only the fade-in phase (use when arriving at hub/new floor
// with no outgoing transition).
func (t *Transition) StartFadeIn() {
	t.state = transStateFadeIn
	t.timer = 0
	t.OnMidpoint = nil
}

// Active returns true while a transition is running.
func (t *Transition) Active() bool {
	return t.state != transStateIdle
}

// Update advances the transition. dt is delta time in seconds.
func (t *Transition) Update(dt float64) {
	if t.state == transStateIdle {
		return
	}
	t.timer += dt
	switch t.state {
	case transStateFadeOut:
		if t.timer >= t.fadeOutDur {
			t.timer = 0
			t.state = transStateHold
			if t.OnMidpoint != nil {
				t.OnMidpoint()
			}
		}
	case transStateHold:
		if t.timer >= t.holdDur {
			t.timer = 0
			t.state = transStateFadeIn
		}
	case transStateFadeIn:
		if t.timer >= t.fadeInDur {
			t.timer = 0
			t.state = transStateIdle
		}
	}
}

// Draw renders the fade overlay on top of everything else.
func (t *Transition) Draw(screen *ebiten.Image) {
	if t.state == transStateIdle {
		return
	}
	if t.overlay == nil {
		return
	}
	var alpha float32
	switch t.state {
	case transStateFadeOut:
		alpha = float32(t.timer / t.fadeOutDur)
	case transStateHold:
		alpha = 1.0
	case transStateFadeIn:
		alpha = float32(1.0 - t.timer/t.fadeInDur)
	}
	if alpha <= 0 {
		return
	}
	t.overlay.Fill(color.Black)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(t.overlay, op)
}
