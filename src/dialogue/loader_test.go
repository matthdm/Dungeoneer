package dialogue

import "testing"

// stubRegistry replaces the global Registry for a test and restores it on cleanup.
func stubRegistry(t *testing.T, entries map[string]*DialogueTree) {
	t.Helper()
	orig := Registry
	Registry = entries
	t.Cleanup(func() { Registry = orig })
}

func tree(id string) *DialogueTree { return &DialogueTree{ID: id} }

func TestSelectTree_StandardPhase(t *testing.T) {
	stubRegistry(t, map[string]*DialogueTree{
		"varn_phase0": tree("varn_phase0"),
		"varn_phase1": tree("varn_phase1"),
	})
	flags := map[string]int{"varn_phase": 1}
	if got := SelectTree("varn", flags); got != "varn_phase1" {
		t.Errorf("want varn_phase1, got %s", got)
	}
}

func TestSelectTree_BetrayedTakesPriorityOverNGPlus(t *testing.T) {
	stubRegistry(t, map[string]*DialogueTree{
		"varn_phase0":   tree("varn_phase0"),
		"varn_betrayed": tree("varn_betrayed"),
		"varn_ng1":      tree("varn_ng1"),
	})
	flags := map[string]int{
		"varn_phase":        0,
		"varn_betrayed":     1,
		"varn_ng_plus":      1,
		"varn_defeat_count": 1,
	}
	if got := SelectTree("varn", flags); got != "varn_betrayed" {
		t.Errorf("want varn_betrayed, got %s", got)
	}
}

func TestSelectTree_BetrayedFlagButNoTreeFallsThrough(t *testing.T) {
	stubRegistry(t, map[string]*DialogueTree{
		"varn_phase0": tree("varn_phase0"),
		// betrayed tree absent
	})
	flags := map[string]int{"varn_phase": 0, "varn_betrayed": 1}
	if got := SelectTree("varn", flags); got != "varn_phase0" {
		t.Errorf("want varn_phase0 fallback, got %s", got)
	}
}

func TestSelectTree_NGPlusDefeatCountTree(t *testing.T) {
	stubRegistry(t, map[string]*DialogueTree{
		"varn_phase0": tree("varn_phase0"),
		"varn_ng1":    tree("varn_ng1"),
		"varn_ng2":    tree("varn_ng2"),
		"varn_ng3":    tree("varn_ng3"),
	})
	cases := []struct {
		defeats int
		want    string
	}{
		{1, "varn_ng1"},
		{2, "varn_ng2"},
		{3, "varn_ng3"},
		{5, "varn_ng3"}, // capped at 3
	}
	for _, c := range cases {
		flags := map[string]int{
			"varn_phase":        0,
			"varn_ng_plus":      1,
			"varn_defeat_count": c.defeats,
		}
		if got := SelectTree("varn", flags); got != c.want {
			t.Errorf("defeat_count=%d: want %s, got %s", c.defeats, c.want, got)
		}
	}
}

func TestSelectTree_NGPlusFallsBackToNGPhaseTree(t *testing.T) {
	stubRegistry(t, map[string]*DialogueTree{
		"varn_phase0":    tree("varn_phase0"),
		"varn_ng_phase0": tree("varn_ng_phase0"),
		// no varn_ng1
	})
	flags := map[string]int{
		"varn_phase":        0,
		"varn_ng_plus":      1,
		"varn_defeat_count": 0,
	}
	if got := SelectTree("varn", flags); got != "varn_ng_phase0" {
		t.Errorf("want varn_ng_phase0 fallback, got %s", got)
	}
}

func TestSelectTree_NGPlusNoSpecialTreeFallsToStandard(t *testing.T) {
	stubRegistry(t, map[string]*DialogueTree{
		"varn_phase1": tree("varn_phase1"),
		// no ng trees
	})
	flags := map[string]int{
		"varn_phase":        1,
		"varn_ng_plus":      1,
		"varn_defeat_count": 2,
	}
	if got := SelectTree("varn", flags); got != "varn_phase1" {
		t.Errorf("want varn_phase1 standard fallback, got %s", got)
	}
}
