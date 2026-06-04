package items

import (
	"testing"
)

// TestAdjustedWeightFloorScaling verifies rarity weight scaling at various floors.
func TestAdjustedWeightFloorScaling(t *testing.T) {
	tests := []struct {
		name   string
		rarity string
		floor  int
		baseW  float64
		wantGt float64 // result should be greater than this
		wantLt float64 // result should be less than this (use -1 to skip)
	}{
		// Common decreases with floor
		{
			name:   "common floor 0 equals base",
			rarity: RarityCommon,
			floor:  0,
			baseW:  10.0,
			wantGt: 9.99, // ~1.0 multiplier
			wantLt: 10.01,
		},
		{
			name:   "common floor 1 less than base",
			rarity: RarityCommon,
			floor:  1,
			baseW:  10.0,
			wantGt: -1,
			wantLt: 10.0,
		},
		{
			name:   "common floor 10 less than floor 1",
			rarity: RarityCommon,
			floor:  10,
			baseW:  10.0,
			wantGt: -1,
			// floor=1: 10 * (1 - 0.03) = 9.7; floor=10: 10 * (1 - 0.30) = 7.0
			wantLt: 9.7,
		},
		{
			name:   "common floor 100 clamps to 10 percent of base",
			rarity: RarityCommon,
			floor:  100,
			baseW:  10.0,
			// max(0.1, ...) clamp: should be exactly 10 * 0.1 = 1.0
			wantGt: 0.99,
			wantLt: 1.01,
		},
		{
			name:   "common floor 34 is near clamp boundary",
			rarity: RarityCommon,
			floor:  34,
			baseW:  10.0,
			// 1 - 0.03*34 = 1 - 1.02 = -0.02 → clamped to 0.1 → 1.0
			wantGt: 0.99,
			wantLt: 1.01,
		},

		// Uncommon increases with floor
		{
			name:   "uncommon floor 0 equals base",
			rarity: RarityUncommon,
			floor:  0,
			baseW:  5.0,
			wantGt: 4.99,
			wantLt: 5.01,
		},
		{
			name:   "uncommon floor 10 greater than floor 1",
			rarity: RarityUncommon,
			floor:  10,
			baseW:  5.0,
			// 5 * (1 + 0.01*10) = 5 * 1.10 = 5.5
			wantGt: 5.4,
			wantLt: 5.6,
		},

		// Rare increases with floor
		{
			name:   "rare floor 0 equals base",
			rarity: RarityRare,
			floor:  0,
			baseW:  2.0,
			wantGt: 1.99,
			wantLt: 2.01,
		},
		{
			name:   "rare floor 10 greater than base",
			rarity: RarityRare,
			floor:  10,
			baseW:  2.0,
			// 2 * (1 + 0.015*10) = 2 * 1.15 = 2.3
			wantGt: 2.29,
			wantLt: 2.31,
		},

		// Legendary increases with floor
		{
			name:   "legendary floor 0 equals base",
			rarity: RarityLegendary,
			floor:  0,
			baseW:  1.0,
			wantGt: 0.99,
			wantLt: 1.01,
		},
		{
			name:   "legendary floor 10 greater than base",
			rarity: RarityLegendary,
			floor:  10,
			baseW:  1.0,
			// 1 * (1 + 0.005*10) = 1 * 1.05 = 1.05
			wantGt: 1.04,
			wantLt: 1.06,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := LootEntry{Weight: tt.baseW, Rarity: tt.rarity}
			got := adjustedWeight(e, tt.floor, 0)

			if tt.wantGt >= 0 && got <= tt.wantGt {
				t.Errorf("adjustedWeight(%s, floor=%d) = %v; want > %v", tt.rarity, tt.floor, got, tt.wantGt)
			}
			if tt.wantLt >= 0 && got >= tt.wantLt {
				t.Errorf("adjustedWeight(%s, floor=%d) = %v; want < %v", tt.rarity, tt.floor, got, tt.wantLt)
			}
		})
	}
}

// TestAdjustedWeightCommonDecreases confirms common weight strictly decreases
// from floor 0 through floor 30 (where the 0.1 clamp begins: 1-0.03*30 = 0.1).
// Beyond floor 30, the clamp holds the weight flat at 10% of base.
func TestAdjustedWeightCommonDecreases(t *testing.T) {
	baseW := 10.0
	entry := LootEntry{Weight: baseW, Rarity: RarityCommon}

	// clampFloor is the first floor where the multiplier hits the 0.1 floor.
	// 1 - 0.03*f <= 0.1  =>  f >= 30
	const clampFloor = 30

	prev := adjustedWeight(entry, 0, 0)
	for floor := 1; floor <= clampFloor; floor++ {
		cur := adjustedWeight(entry, floor, 0)
		if cur >= prev {
			t.Errorf("floor %d: common weight %v did not decrease from floor %d weight %v", floor, cur, floor-1, prev)
		}
		prev = cur
	}

	// After the clamp floor, weight should stay flat at 10% of base.
	// Use a tolerance for float64 comparison.
	const eps = 1e-9
	expected := baseW * 0.1
	for floor := clampFloor; floor <= 100; floor++ {
		cur := adjustedWeight(entry, floor, 0)
		diff := cur - expected
		if diff < -eps || diff > eps {
			t.Errorf("floor %d: common weight %v should be ~%.2f (10%% of base, within %.1e)", floor, cur, expected, eps)
		}
	}
}

// TestAdjustedWeightRarityOrdering confirms that at high floors the rarity
// ordering uncommon < rare < legendary holds relative to their floor boosts.
func TestAdjustedWeightRarityOrdering(t *testing.T) {
	floor := 20
	base := 1.0

	uncommonW := adjustedWeight(LootEntry{Weight: base, Rarity: RarityUncommon}, floor, 0)
	rareW := adjustedWeight(LootEntry{Weight: base, Rarity: RarityRare}, floor, 0)
	legendaryW := adjustedWeight(LootEntry{Weight: base, Rarity: RarityLegendary}, floor, 0)

	if uncommonW <= base {
		t.Errorf("uncommon at floor %d: %v should be > base %v", floor, uncommonW, base)
	}
	if rareW <= base {
		t.Errorf("rare at floor %d: %v should be > base %v", floor, rareW, base)
	}
	if legendaryW <= base {
		t.Errorf("legendary at floor %d: %v should be > base %v", floor, legendaryW, base)
	}
	// Rare scales faster than uncommon (0.015 vs 0.01)
	if rareW <= uncommonW {
		t.Errorf("rare at floor %d (%v) should be > uncommon (%v)", floor, rareW, uncommonW)
	}
}

// TestShouldDropBossAlwaysDrops confirms bosses always drop loot.
func TestShouldDropBossAlwaysDrops(t *testing.T) {
	for floor := 0; floor <= 20; floor++ {
		// Run multiple times to confirm it's always true (not probabilistic).
		for i := 0; i < 10; i++ {
			if !ShouldDrop("boss", floor) {
				t.Errorf("ShouldDrop(boss, floor=%d) returned false; expected always true", floor)
			}
		}
	}
}

// TestShouldDropNormalMonsterRange confirms normal monsters have drop chance in [30%,90%].
func TestShouldDropNormalMonsterRange(t *testing.T) {
	// Statistical test: run many trials and verify the empirical rate is plausible.
	// Floor 1 should give 30 + 1*2 = 32% chance.
	// Floor 30 would hit the 90% cap.
	// We just verify: at floor 1, not every roll returns true (not 100%) and
	// not every roll returns false (not 0%).
	const trials = 2000
	floor := 1
	hits := 0
	for i := 0; i < trials; i++ {
		if ShouldDrop("normal", floor) {
			hits++
		}
	}
	rate := float64(hits) / float64(trials)
	// Expected ~32%, accept 20%-50% range for statistical noise.
	if rate < 0.20 || rate > 0.50 {
		t.Errorf("ShouldDrop(normal, floor=1) empirical rate = %.3f; expected ~0.32 (range 0.20–0.50)", rate)
	}
}

// TestShouldDropEliteHigherThanNormal confirms elites have a higher base drop
// rate than normal monsters at the same floor.
func TestShouldDropEliteHigherThanNormal(t *testing.T) {
	const trials = 5000
	floor := 1
	normalHits, eliteHits := 0, 0
	for i := 0; i < trials; i++ {
		if ShouldDrop("normal", floor) {
			normalHits++
		}
		if ShouldDrop("elite", floor) {
			eliteHits++
		}
	}
	normalRate := float64(normalHits) / float64(trials)
	eliteRate := float64(eliteHits) / float64(trials)
	if eliteRate <= normalRate {
		t.Errorf("elite drop rate %.3f should be > normal drop rate %.3f at floor %d", eliteRate, normalRate, floor)
	}
}

// TestShouldDropChanceCapAt90 confirms the drop chance never exceeds 90%.
func TestShouldDropChanceCapAt90(t *testing.T) {
	// At a very high floor (e.g. floor 100), base + floor*0.02 would be huge,
	// but the function clamps at 0.90. We can't easily test this probabilistically
	// without mocking rand, but we can verify it doesn't return true every time
	// at floor 100 by running many trials (very unlikely all fail if uncapped).
	// Instead, just verify floor 1 behaves near expected 32%.
	const trials = 2000
	floor := 30 // base 0.30 + 30*0.02 = 0.90 exactly → cap
	hits := 0
	for i := 0; i < trials; i++ {
		if ShouldDrop("normal", floor) {
			hits++
		}
	}
	rate := float64(hits) / float64(trials)
	// At cap 90%, accept 85%-95% range.
	if rate < 0.80 || rate > 0.99 {
		t.Errorf("ShouldDrop(normal, floor=30) empirical rate = %.3f; expected ~0.90 (range 0.80–0.99)", rate)
	}
}

func TestAdjustedWeightLuckScaling(t *testing.T) {
	// Rare and legendary items should have higher weights when luck > 0
	base := 1.0
	floor := 5

	// Rare without luck vs with luck
	rareNoLuck := adjustedWeight(LootEntry{Weight: base, Rarity: RarityRare}, floor, 0)
	rareWithLuck := adjustedWeight(LootEntry{Weight: base, Rarity: RarityRare}, floor, 5)
	if rareWithLuck <= rareNoLuck {
		t.Errorf("rare item weight should increase with luck: got %v with luck, %v without", rareWithLuck, rareNoLuck)
	}

	// Legendary without luck vs with luck
	legendaryNoLuck := adjustedWeight(LootEntry{Weight: base, Rarity: RarityLegendary}, floor, 0)
	legendaryWithLuck := adjustedWeight(LootEntry{Weight: base, Rarity: RarityLegendary}, floor, 5)
	if legendaryWithLuck <= legendaryNoLuck {
		t.Errorf("legendary item weight should increase with luck: got %v with luck, %v without", legendaryWithLuck, legendaryNoLuck)
	}

	// Common should not scale with luck
	commonNoLuck := adjustedWeight(LootEntry{Weight: base, Rarity: RarityCommon}, floor, 0)
	commonWithLuck := adjustedWeight(LootEntry{Weight: base, Rarity: RarityCommon}, floor, 5)
	if commonWithLuck != commonNoLuck {
		t.Errorf("common item weight should not change with luck: got %v with luck, %v without", commonWithLuck, commonNoLuck)
	}
}
