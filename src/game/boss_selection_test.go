package game

import "testing"

func TestSelectBoss(t *testing.T) {
	if SelectBoss(nil) != BossGeneric {
		t.Errorf("expected generic boss on nil run state")
	}
	
	rs := &RunState{
		QuestFlags: make(map[string]int),
	}
	
	if SelectBoss(rs) != BossGeneric {
		t.Errorf("expected generic boss when flag not set")
	}
	
	rs.QuestFlags["varn_p3_done"] = 1
	if SelectBoss(rs) != BossVarn {
		t.Errorf("expected Varn boss when flag is set")
	}
}
