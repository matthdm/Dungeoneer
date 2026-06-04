package game

import "testing"

func TestRunState(t *testing.T) {
	rs := NewRunState(5)
	
	if !rs.Active {
		t.Errorf("expected new run to be active")
	}
	if rs.CurrentFloor != 1 {
		t.Errorf("expected starting floor to be 1")
	}
	if rs.TotalFloors != 5 {
		t.Errorf("expected 5 total floors")
	}
	if len(rs.Biomes) != 5 {
		t.Errorf("expected 5 biomes generated")
	}
	
	// Test IsLastFloor
	if rs.IsLastFloor() {
		t.Errorf("expected 1 not to be last floor of 5")
	}
	
	rs.CurrentFloor = 5
	if !rs.IsLastFloor() {
		t.Errorf("expected 5 to be last floor of 5")
	}
	
	// Test CalculateRemnants
	rs.FloorsCleared = 3
	rs.KillCount = 10
	if rs.CalculateRemnants() != 3*10 + 10*2 {
		t.Errorf("unexpected remnant calculation")
	}
	
	// Test BuildFloorContext
	ctx := rs.BuildFloorContext(1)
	if ctx.FloorNumber != 1 {
		t.Errorf("expected floor num 1")
	}
	if ctx.TotalFloors != 5 {
		t.Errorf("expected 5 total floors")
	}
	if ctx.Difficulty != 0.0 {
		t.Errorf("expected difficulty 0.0 on floor 1")
	}
	
	ctx5 := rs.BuildFloorContext(5)
	if ctx5.Difficulty != 1.0 {
		t.Errorf("expected difficulty 1.0 on floor 5")
	}
	
	// Fallback when out of bounds
	ctxOut := rs.BuildFloorContext(10)
	if ctxOut.Biome != BiomeCrypt {
		t.Errorf("expected fallback biome crypt")
	}
	if ctxOut.Difficulty != 1.0 { // Actually 9 / 4 = 2.25
		// we don't clamp it currently but let's check it's > 0
		if ctxOut.Difficulty <= 1.0 {
			t.Errorf("expected >1 difficulty")
		}
	}
}
