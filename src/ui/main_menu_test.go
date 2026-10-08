package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestMainMenuOptions_WhenQuitIsPossible(t *testing.T) {
	got := mainMenuOptions(true)
	want := []string{"Continue", "New Game", "Load Game", "Options", "Exit Game"}
	if len(got) != len(want) {
		t.Fatalf("options = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("options = %v, want %v", got, want)
		}
	}
}

func TestMainMenuOptions_WhenQuitIsImpossible(t *testing.T) {
	got := mainMenuOptions(false)
	want := []string{"Continue", "New Game", "Load Game", "Options"}
	if len(got) != len(want) {
		t.Fatalf("options = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("options = %v, want %v", got, want)
		}
	}
}

// The draw code lays out one label per option and the input code indexes
// Options by the label's position, so the two must stay parallel.
func TestMainMenuLabels_ParallelToOptions(t *testing.T) {
	m := &MainMenu{
		ContinueGameLabel: ebiten.NewImage(1, 1),
		NewGameLabel:      ebiten.NewImage(1, 1),
		LoadGameLabel:     ebiten.NewImage(1, 1),
		OptionsLabel:      ebiten.NewImage(1, 1),
		ExitGameLabel:     ebiten.NewImage(1, 1),
	}

	m.Options = mainMenuOptions(true)
	labels := m.Labels()
	want := []*ebiten.Image{m.ContinueGameLabel, m.NewGameLabel, m.LoadGameLabel, m.OptionsLabel, m.ExitGameLabel}
	if len(labels) != len(want) {
		t.Fatalf("got %d labels, want %d", len(labels), len(want))
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Errorf("label %d does not match option %q", i, m.Options[i])
		}
	}

	m.Options = mainMenuOptions(false)
	labels = m.Labels()
	if len(labels) != 4 {
		t.Fatalf("got %d labels without quit, want 4", len(labels))
	}
	for i, l := range labels {
		if l == m.ExitGameLabel {
			t.Errorf("label %d is the Exit Game label, want it left out", i)
		}
	}
}
