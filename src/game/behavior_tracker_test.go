package game

import "testing"

func TestBehaviorTracker(t *testing.T) {
	bt := &BehaviorTracker{}
	
	// Test inactive state (before start)
	bt.RecordSpellCast("fireball")
	bt.RecordKill(true)
	bt.RecordEnemySpawned()
	
	if bt.Record.TotalKills > 0 || bt.Record.TotalEnemies > 0 || len(bt.Record.SpellUsage) > 0 {
		t.Errorf("tracker should ignore records when inactive")
	}
	
	// Test start
	bt.Start(10)
	if bt.Record.TotalEnemies != 10 {
		t.Errorf("expected 10 initial enemies")
	}
	
	// Test active state
	bt.RecordSpellCast("fireball")
	bt.RecordSpellCast("fireball")
	bt.RecordSpellCast("heal")
	
	bt.RecordKill(true) // ranged
	bt.RecordKill(false) // melee
	bt.RecordKill(false) // melee
	
	bt.RecordEnemySpawned()
	bt.RecordEnemySpawned()
	
	rec := bt.Finalize(true)
	
	if rec.Deaths != 1 {
		t.Errorf("expected 1 death")
	}
	if rec.TotalKills != 3 {
		t.Errorf("expected 3 total kills")
	}
	if rec.MeleeKills != 2 {
		t.Errorf("expected 2 melee kills")
	}
	if rec.RangedKills != 1 {
		t.Errorf("expected 1 ranged kill")
	}
	if rec.TotalEnemies != 12 {
		t.Errorf("expected 12 total enemies")
	}
	if rec.SpellUsage["fireball"] != 2 {
		t.Errorf("expected 2 fireball casts")
	}
	if rec.SpellUsage["heal"] != 1 {
		t.Errorf("expected 1 heal cast")
	}
	
	// Test ratios
	if rec.KillRatio() != 3.0/12.0 {
		t.Errorf("unexpected kill ratio")
	}
	if rec.RangedRatio() != 1.0/3.0 {
		t.Errorf("unexpected ranged ratio")
	}
	if rec.SpellVariety() != 2 {
		t.Errorf("expected 2 distinct spells")
	}
	
	// Test zero values
	zeroRec := BehaviorRecord{}
	if zeroRec.KillRatio() != 1.0 {
		t.Errorf("expected kill ratio of 1.0 when no enemies")
	}
	if zeroRec.RangedRatio() != 0.0 {
		t.Errorf("expected ranged ratio of 0.0 when no kills")
	}
}
