package items

import "testing"

func TestRecalculateSetBonuses_PartialSetNoActiveTier(t *testing.T) {
	// Only 1 piece of Stormcaller (needs 2) — no bonus tier, but shows as partial.
	active := RecalculateSetBonuses([]string{"item_0_26"}, false)
	if len(active) != 1 {
		t.Fatalf("want 1 partial entry, got %d", len(active))
	}
	if active[0].Tier.PiecesRequired != 0 {
		t.Error("partial set should have zero Tier (no qualifying tier)")
	}
}

func TestRecalculateSetBonuses_FullSetAppliesBonus(t *testing.T) {
	active := RecalculateSetBonuses([]string{"item_0_26", "item_0_35"}, false)
	if len(active) == 0 {
		t.Fatal("expected at least 1 active set bonus")
	}
	found := false
	for _, a := range active {
		if a.SetID == "stormcaller" && a.Tier.PiecesRequired == 2 {
			found = true
		}
	}
	if !found {
		t.Error("stormcaller 2pc bonus should be active")
	}
}

func TestRecalculateSetBonuses_QuestLockedSetSkippedWithoutQuest(t *testing.T) {
	// All 3 Chainbreaker pieces, but quest not complete.
	active := RecalculateSetBonuses([]string{"item_1_12", "item_0_3", "item_0_1"}, false)
	for _, a := range active {
		if a.SetID == "chainbreaker" {
			t.Error("chainbreaker should be skipped when quest not complete")
		}
	}
}

func TestRecalculateSetBonuses_QuestLockedSetActivatesWithQuest(t *testing.T) {
	active := RecalculateSetBonuses([]string{"item_1_12", "item_0_3", "item_0_1"}, true)
	found := false
	for _, a := range active {
		if a.SetID == "chainbreaker" {
			found = true
		}
	}
	if !found {
		t.Error("chainbreaker should be active when quest complete and all pieces equipped")
	}
}

func TestRecalculateSetBonuses_EmptyInventory(t *testing.T) {
	active := RecalculateSetBonuses(nil, false)
	if len(active) != 0 {
		t.Errorf("empty inventory should yield no active sets, got %d", len(active))
	}
}
