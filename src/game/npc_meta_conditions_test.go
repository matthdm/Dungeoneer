package game

import (
	"dungeoneer/dialogue"
	"os"
	"testing"
)

func makeGameWithMeta(defeatCount, totalTrust, highestPhase int) *Game {
	g := &Game{
		Meta: &MetaSave{
			NPCMeta: map[string]*NPCMetaState{
				"varn": {
					DefeatCount:  defeatCount,
					TotalTrust:   totalTrust,
					HighestPhase: highestPhase,
				},
			},
		},
		RunState: NewRunState(1),
	}
	migrateMetaSave(g.Meta)
	return g
}

func TestMetaFlagGte_DefeatCount(t *testing.T) {
	g := makeGameWithMeta(2, 0, 0)
	cond := &dialogue.DialogueCondition{
		Type:  "meta_flag_gte",
		Flag:  "varn",
		Field: "defeat_count",
		Value: 2,
	}
	if !g.evalDialogueCondition(cond) {
		t.Error("meta_flag_gte defeat_count 2 should be true when DefeatCount=2")
	}
}

func TestMetaFlagGte_TotalTrust(t *testing.T) {
	g := makeGameWithMeta(0, 75, 0)
	cond := &dialogue.DialogueCondition{
		Type:  "meta_flag_gte",
		Flag:  "varn",
		Field: "total_trust",
		Value: 60,
	}
	if !g.evalDialogueCondition(cond) {
		t.Error("meta_flag_gte total_trust 60 should be true when TotalTrust=75")
	}
}

func TestMetaFlagEquals_HighestPhase(t *testing.T) {
	g := makeGameWithMeta(1, 0, 2)
	cond := &dialogue.DialogueCondition{
		Type:  "meta_flag_equals",
		Flag:  "varn",
		Field: "highest_phase",
		Value: 2,
	}
	if !g.evalDialogueCondition(cond) {
		t.Error("meta_flag_equals highest_phase 2 should be true when HighestPhase=2")
	}
}

func TestMetaFlagGte_NPCNotMet_ReturnsFalse(t *testing.T) {
	g := makeGameWithMeta(0, 0, 0)
	g.Meta.NPCMeta = make(map[string]*NPCMetaState)
	cond := &dialogue.DialogueCondition{
		Type:  "meta_flag_gte",
		Flag:  "seris",
		Field: "defeat_count",
		Value: 1,
	}
	if g.evalDialogueCondition(cond) {
		t.Error("meta_flag_gte for unknown NPC should return false")
	}
}

// --- unlock_lore action ---

func TestExecAction_UnlockLore_AddsToMetaAndQueuesToast(t *testing.T) {
	t.Cleanup(func() { os.Remove("meta.json") })
	origReg := LoreRegistry
	LoreRegistry = []LoreDef{{ID: "test_entry", Title: "Test Entry", Category: LoreCategoryCharacter, Body: "body"}}
	defer func() { LoreRegistry = origReg }()

	g := &Game{Meta: &MetaSave{}}
	migrateMetaSave(g.Meta)

	g.execDialogueAction(dialogue.DialogueAction{Type: "unlock_lore", LoreID: "test_entry"})

	if !IsLoreUnlocked(g.Meta, "test_entry") {
		t.Error("test_entry should be unlocked after unlock_lore action")
	}
	if len(g.pendingToasts) == 0 {
		t.Error("a toast should be queued on first lore unlock")
	}
}

func TestExecAction_UnlockLore_Idempotent(t *testing.T) {
	t.Cleanup(func() { os.Remove("meta.json") })
	origReg := LoreRegistry
	LoreRegistry = []LoreDef{{ID: "test_entry", Title: "Test Entry", Category: LoreCategoryCharacter, Body: "body"}}
	defer func() { LoreRegistry = origReg }()

	g := &Game{Meta: &MetaSave{LoreUnlocked: []string{"test_entry"}}}
	migrateMetaSave(g.Meta)

	g.execDialogueAction(dialogue.DialogueAction{Type: "unlock_lore", LoreID: "test_entry"})

	if len(g.Meta.LoreUnlocked) != 1 {
		t.Errorf("LoreUnlocked should stay at 1 entry, got %d", len(g.Meta.LoreUnlocked))
	}
	if len(g.pendingToasts) != 0 {
		t.Error("no toast should fire for already-unlocked lore")
	}
}

func TestExecAction_UnlockLore_NilMetaIsNoop(t *testing.T) {
	g := &Game{Meta: nil}
	// Should not panic.
	g.execDialogueAction(dialogue.DialogueAction{Type: "unlock_lore", LoreID: "anything"})
}

// --- set_betrayed persistence ---

func TestExecAction_SetBetrayedPersistsToMetaSave(t *testing.T) {
	t.Cleanup(func() { os.Remove("meta.json") })
	g := &Game{
		Meta: &MetaSave{
			NPCMeta: map[string]*NPCMetaState{"varn": {Met: true}},
		},
		RunState: NewRunState(1),
	}
	migrateMetaSave(g.Meta)

	g.execDialogueAction(dialogue.DialogueAction{Type: "set_betrayed", Flag: "varn"})

	if g.RunState.QuestFlags["varn_betrayed"] != 1 {
		t.Error("varn_betrayed flag should be set in RunState QuestFlags")
	}
	if !g.Meta.NPCMeta["varn"].Betrayed {
		t.Error("Betrayed should be persisted to MetaSave.NPCMeta[varn]")
	}
}

func TestExecAction_SetBetrayedCreatesNPCMetaIfAbsent(t *testing.T) {
	t.Cleanup(func() { os.Remove("meta.json") })
	g := &Game{
		Meta:     &MetaSave{},
		RunState: NewRunState(1),
	}
	migrateMetaSave(g.Meta)

	g.execDialogueAction(dialogue.DialogueAction{Type: "set_betrayed", Flag: "seris"})

	if g.Meta.NPCMeta["seris"] == nil {
		t.Fatal("NPCMeta entry should be created for seris")
	}
	if !g.Meta.NPCMeta["seris"].Betrayed {
		t.Error("Betrayed should be true for newly created NPCMeta entry")
	}
}
