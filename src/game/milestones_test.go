package game

import "testing"

func TestCheckMilestones_ShopUnlocksAfterFirstRun(t *testing.T) {
	meta := &MetaSave{}
	migrateMetaSave(meta)
	meta.CompletedRuns = 1
	newly := CheckMilestones(meta)
	if !meta.HubState[MilestoneShop] {
		t.Error("shop should unlock after 1 completed run")
	}
	if len(newly) == 0 {
		t.Error("should return newly unlocked milestones")
	}
}

func TestCheckMilestones_UpgradesAfterThreeRuns(t *testing.T) {
	meta := &MetaSave{}
	migrateMetaSave(meta)
	meta.CompletedRuns = 3
	meta.HubState[MilestoneShop] = true
	newly := CheckMilestones(meta)
	if !meta.HubState[MilestoneUpgrades] {
		t.Error("upgrades should unlock after 3 runs")
	}
	found := false
	for _, id := range newly {
		if id == MilestoneUpgrades {
			found = true
		}
	}
	if !found {
		t.Error("MilestoneUpgrades should be in newly list")
	}
}

func TestCheckMilestones_EchoShrineOnFirstDeath(t *testing.T) {
	meta := &MetaSave{}
	migrateMetaSave(meta)
	meta.TotalDeaths = 1
	CheckMilestones(meta)
	if !meta.HubState[MilestoneEchoShrine] {
		t.Error("echo shrine should unlock on first death")
	}
}

func TestCheckMilestones_AlreadyUnlockedNotReturned(t *testing.T) {
	meta := &MetaSave{}
	migrateMetaSave(meta)
	meta.CompletedRuns = 5
	meta.HubState[MilestoneShop] = true
	meta.HubState[MilestoneUpgrades] = true
	newly := CheckMilestones(meta)
	for _, id := range newly {
		if id == MilestoneShop || id == MilestoneUpgrades {
			t.Errorf("already-unlocked %q should not appear in newly", id)
		}
	}
}

func TestCheckMilestones_LoreLibraryOnNPCPhase1(t *testing.T) {
	meta := &MetaSave{}
	migrateMetaSave(meta)
	meta.NPCMeta["varn"] = &NPCMetaState{HighestPhase: 1}
	CheckMilestones(meta)
	if !meta.HubState[MilestoneLoreLibrary] {
		t.Error("lore library should unlock when any NPC reaches phase 1")
	}
}
