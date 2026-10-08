package ui

import "testing"

func optionTexts(opts []MenuOption) []string {
	texts := make([]string, len(opts))
	for i, o := range opts {
		texts[i] = o.Text
	}
	return texts
}

func TestAppendExitOption_WhenQuitIsPossible(t *testing.T) {
	called := false
	base := []MenuOption{{Text: "Resume"}, {Text: "Settings"}}

	got := appendExitOption(base, func() { called = true }, true)

	if len(got) != 3 || got[2].Text != "Exit Game" {
		t.Fatalf("options = %v, want Exit Game appended last", optionTexts(got))
	}
	got[2].Action()
	if !called {
		t.Error("Exit Game action did not call onExit")
	}
}

func TestAppendExitOption_WhenQuitIsImpossible(t *testing.T) {
	base := []MenuOption{{Text: "Resume"}, {Text: "Settings"}}

	got := appendExitOption(base, func() {}, false)

	for _, text := range optionTexts(got) {
		if text == "Exit Game" {
			t.Fatalf("options = %v, want no Exit Game entry", optionTexts(got))
		}
	}
	if len(got) != 2 {
		t.Errorf("options = %v, want the base options unchanged", optionTexts(got))
	}
}
