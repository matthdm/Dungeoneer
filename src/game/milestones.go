package game

const (
	MilestoneShop        = "shop_unlocked"
	MilestoneUpgrades    = "upgrades_unlocked"
	MilestoneEchoShrine  = "echo_shrine_unlocked"
	MilestoneLoreLibrary = "lore_library_unlocked"
)

// MilestoneMessages maps milestone IDs to the toast text shown on first unlock.
var MilestoneMessages = map[string]string{
	MilestoneShop:        "A merchant has arrived at the hub.",
	MilestoneUpgrades:    "An upgrade station has appeared.",
	MilestoneEchoShrine:  "An echo shrine has manifested.",
	MilestoneLoreLibrary: "A lore library has opened.",
}

// CheckMilestones evaluates thresholds and sets HubState flags for newly-met
// milestones. Returns the IDs of milestones that unlocked this call only.
func CheckMilestones(meta *MetaSave) []string {
	if meta.HubState == nil {
		meta.HubState = make(map[string]bool)
	}
	var newly []string
	check := func(id string, cond bool) {
		if cond && !meta.HubState[id] {
			meta.HubState[id] = true
			newly = append(newly, id)
		}
	}
	check(MilestoneShop, meta.CompletedRuns >= 1)
	check(MilestoneUpgrades, meta.CompletedRuns >= 3)
	check(MilestoneEchoShrine, meta.TotalDeaths >= 1)
	for _, state := range meta.NPCMeta {
		if state.HighestPhase >= 1 {
			check(MilestoneLoreLibrary, true)
			break
		}
	}
	return newly
}
