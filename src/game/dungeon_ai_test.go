package game

import "testing"

func TestInferMood_SpitefulWinsOverDeceptive(t *testing.T) {
	p := PlayerProfile{DeathRate: 0.8, AvgRangedRatio: 0.9}
	if got := InferMood(p); got != MoodSpiteful {
		t.Errorf("want MoodSpiteful, got %v", got)
	}
}

func TestInferMood_DeceptiveWhenHighRanged(t *testing.T) {
	p := PlayerProfile{DeathRate: 0.3, AvgRangedRatio: 0.75}
	if got := InferMood(p); got != MoodDeceptive {
		t.Errorf("want MoodDeceptive, got %v", got)
	}
}

func TestInferMood_CautiousWhenLowKillRatio(t *testing.T) {
	p := PlayerProfile{DeathRate: 0.2, AvgRangedRatio: 0.3, AvgKillRatio: 0.4}
	if got := InferMood(p); got != MoodCautious {
		t.Errorf("want MoodCautious, got %v", got)
	}
}

func TestInferMood_ChaoticWhenManySpells(t *testing.T) {
	p := PlayerProfile{DeathRate: 0.1, AvgRangedRatio: 0.2, AvgKillRatio: 0.8, AvgSpellVariety: 5}
	if got := InferMood(p); got != MoodChaotic {
		t.Errorf("want MoodChaotic, got %v", got)
	}
}

func TestInferMood_NeutralDefault(t *testing.T) {
	p := PlayerProfile{DeathRate: 0.3, AvgRangedRatio: 0.3, AvgKillRatio: 0.7, AvgSpellVariety: 2}
	if got := InferMood(p); got != MoodNeutral {
		t.Errorf("want MoodNeutral, got %v", got)
	}
}

func TestBuildProfile_WeightsMostRecentRun(t *testing.T) {
	records := []BehaviorRecord{
		{TotalKills: 10, TotalEnemies: 20, RangedKills: 0, Deaths: 1, SpellUsage: map[string]int{"a": 1}},
		{TotalKills: 18, TotalEnemies: 20, RangedKills: 14, Deaths: 0, SpellUsage: map[string]int{"b": 3, "c": 2}},
	}
	p := BuildProfile(records)
	// Most recent has kill ratio 0.9 (weight 2), first has 0.5 (weight 1).
	// Weighted avg = (0.5*1 + 0.9*2) / 3 = 2.3/3 ≈ 0.767
	if p.AvgKillRatio < 0.75 || p.AvgKillRatio > 0.78 {
		t.Errorf("AvgKillRatio want ~0.767, got %.3f", p.AvgKillRatio)
	}
}

func TestMoodGenModifiers_DeceptiveBoostsRanged(t *testing.T) {
	delta := MoodGenModifiers(MoodDeceptive)
	if delta.RangedEnemyBias < 1.4 {
		t.Errorf("Deceptive mood should boost ranged bias >=1.5, got %.2f", delta.RangedEnemyBias)
	}
}

func TestMoodGenModifiers_SpitefulTightensCorridors(t *testing.T) {
	delta := MoodGenModifiers(MoodSpiteful)
	if delta.CorridorWidthMod <= 0 {
		t.Error("Spiteful mood should widen CorridorWidthMod to tighten corridors")
	}
}
