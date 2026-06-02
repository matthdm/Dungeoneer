package game

import (
	"os"
	"testing"
)

// minimalGame builds a Game with just enough state to run endRunDeath/endRunVictory.
func minimalGame(t *testing.T) *Game {
	t.Helper()
	t.Cleanup(func() {
		os.Remove("meta.json")
		os.Remove("runsave.json")
	})
	g := &Game{
		Meta: &MetaSave{
			NPCMeta: make(map[string]*NPCMetaState),
		},
		RunState: &RunState{
			Active:        true,
			KillCount:     5,
			FloorsCleared: 3,
			TotalFloors:   7,
			QuestFlags:    make(map[string]int),
		},
	}
	migrateMetaSave(g.Meta)
	return g
}

func TestEndRunDeath_IncrementsTotalDeaths(t *testing.T) {
	g := minimalGame(t)
	g.endRunDeath()
	if g.Meta.TotalDeaths != 1 {
		t.Errorf("TotalDeaths want 1, got %d", g.Meta.TotalDeaths)
	}
}

func TestEndRunDeath_AccumulatesTotalKills(t *testing.T) {
	g := minimalGame(t)
	g.Meta.TotalKills = 10
	g.endRunDeath()
	if g.Meta.TotalKills != 15 {
		t.Errorf("TotalKills want 15, got %d", g.Meta.TotalKills)
	}
}

func TestEndRunDeath_UpdatesBestFloor(t *testing.T) {
	g := minimalGame(t)
	g.endRunDeath()
	if g.Meta.BestFloor != 3 {
		t.Errorf("BestFloor want 3, got %d", g.Meta.BestFloor)
	}
}

func TestEndRunDeath_DoesNotOverwriteBetterBestFloor(t *testing.T) {
	g := minimalGame(t)
	g.Meta.BestFloor = 10
	g.endRunDeath()
	if g.Meta.BestFloor != 10 {
		t.Errorf("BestFloor should not regress: want 10, got %d", g.Meta.BestFloor)
	}
}

func TestEndRunDeath_AccumulatesTotalRemnants(t *testing.T) {
	g := minimalGame(t)
	g.endRunDeath()
	earned := g.Meta.TotalRemnants
	if earned <= 0 {
		t.Errorf("TotalRemnants should be positive after run, got %d", earned)
	}
}

func TestEndRunVictory_IncrementsCompletedRuns(t *testing.T) {
	g := minimalGame(t)
	g.endRunVictory()
	if g.Meta.CompletedRuns != 1 {
		t.Errorf("CompletedRuns want 1, got %d", g.Meta.CompletedRuns)
	}
}

func TestEndRunVictory_DoesNotIncrementTotalDeaths(t *testing.T) {
	g := minimalGame(t)
	g.endRunVictory()
	if g.Meta.TotalDeaths != 0 {
		t.Errorf("TotalDeaths should stay 0 on victory, got %d", g.Meta.TotalDeaths)
	}
}

func TestEndRunVictory_SetsBestFloorToTotalFloors(t *testing.T) {
	g := minimalGame(t)
	g.endRunVictory()
	if g.Meta.BestFloor != 7 {
		t.Errorf("BestFloor want 7 (TotalFloors), got %d", g.Meta.BestFloor)
	}
}

func TestEndRunVictory_UnlocksShopMilestoneOnFirstWin(t *testing.T) {
	g := minimalGame(t)
	g.endRunVictory()
	if !g.Meta.HubState[MilestoneShop] {
		t.Error("shop milestone should unlock after first victory")
	}
}

func TestQueueMilestoneToasts_AppendsMessages(t *testing.T) {
	g := &Game{}
	g.queueMilestoneToasts([]string{MilestoneShop, MilestoneUpgrades})
	if len(g.pendingToasts) != 2 {
		t.Fatalf("want 2 pending toasts, got %d", len(g.pendingToasts))
	}
	if g.pendingToasts[0] != MilestoneMessages[MilestoneShop] {
		t.Errorf("first toast want %q, got %q", MilestoneMessages[MilestoneShop], g.pendingToasts[0])
	}
}

func TestQueueMilestoneToasts_NilSliceAddsNothing(t *testing.T) {
	g := &Game{}
	g.queueMilestoneToasts(nil)
	if len(g.pendingToasts) != 0 {
		t.Errorf("nil milestones should add no toasts, got %d", len(g.pendingToasts))
	}
}

func TestQueueMilestoneToasts_UnknownIDSkipped(t *testing.T) {
	g := &Game{}
	g.queueMilestoneToasts([]string{"nonexistent_milestone"})
	if len(g.pendingToasts) != 0 {
		t.Errorf("unknown milestone ID should not add toast, got %d toasts", len(g.pendingToasts))
	}
}
