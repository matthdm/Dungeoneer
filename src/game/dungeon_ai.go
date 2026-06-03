package game

// DungeonMood represents the dungeon's adaptive personality for the current run.
type DungeonMood int

const (
	MoodNeutral   DungeonMood = iota
	MoodSpiteful              // player dies often
	MoodDeceptive             // player uses ranged heavily
	MoodCautious              // player avoids fights (low kill ratio)
	MoodChaotic               // player uses many different spells
)

// PlayerProfile aggregates behavior across recent runs (weighted average).
type PlayerProfile struct {
	AvgKillRatio    float64 `json:"avg_kill_ratio"`
	AvgRangedRatio  float64 `json:"avg_ranged_ratio"`
	AvgSpellVariety float64 `json:"avg_spell_variety"`
	DeathRate       float64 `json:"death_rate"` // fraction of runs that ended in death
}

// BuildProfile computes a weighted average profile from recent behavior records.
// Most recent run is weighted 2×; older runs weight 1× each.
func BuildProfile(records []BehaviorRecord) PlayerProfile {
	if len(records) == 0 {
		return PlayerProfile{AvgKillRatio: 1.0}
	}
	// Weights: most recent (last in slice) = 2.0, others = 1.0
	totalWeight := 0.0
	var p PlayerProfile
	for i, r := range records {
		w := 1.0
		if i == len(records)-1 {
			w = 2.0
		}
		p.AvgKillRatio += r.KillRatio() * w
		p.AvgRangedRatio += r.RangedRatio() * w
		p.AvgSpellVariety += float64(r.SpellVariety()) * w
		p.DeathRate += float64(r.Deaths) * w
		totalWeight += w
	}
	p.AvgKillRatio /= totalWeight
	p.AvgRangedRatio /= totalWeight
	p.AvgSpellVariety /= totalWeight
	p.DeathRate /= totalWeight
	return p
}

// InferMood derives a DungeonMood from a PlayerProfile.
// Priority: Spiteful > Deceptive > Cautious > Chaotic > Neutral.
func InferMood(p PlayerProfile) DungeonMood {
	switch {
	case p.DeathRate > 0.6:
		return MoodSpiteful
	case p.AvgRangedRatio > 0.7:
		return MoodDeceptive
	case p.AvgKillRatio < 0.5:
		return MoodCautious
	case p.AvgSpellVariety > 4:
		return MoodChaotic
	default:
		return MoodNeutral
	}
}

// GenParamsDelta holds additive modifiers to apply on top of base GenParams.
type GenParamsDelta struct {
	RoomCountDelta   int     // +/- rooms
	CorridorWidthMod int     // +/- corridor width (base is 1)
	RangedEnemyBias  float64 // multiplier on ranged encounter weight (1.0 = no change)
	AmbushBias       float64 // multiplier on ambush encounter weight
}

// MoodGenModifiers returns the generation modifier for the given mood.
func MoodGenModifiers(mood DungeonMood) GenParamsDelta {
	switch mood {
	case MoodSpiteful:
		return GenParamsDelta{RoomCountDelta: -2, CorridorWidthMod: 1, RangedEnemyBias: 1.0, AmbushBias: 1.2}
	case MoodDeceptive:
		return GenParamsDelta{RoomCountDelta: 0, CorridorWidthMod: 0, RangedEnemyBias: 1.5, AmbushBias: 1.0}
	case MoodCautious:
		return GenParamsDelta{RoomCountDelta: 1, CorridorWidthMod: 0, RangedEnemyBias: 1.0, AmbushBias: 1.5}
	case MoodChaotic:
		return GenParamsDelta{RoomCountDelta: 2, CorridorWidthMod: 0, RangedEnemyBias: 1.0, AmbushBias: 1.0}
	default:
		return GenParamsDelta{}
	}
}

// DungeonWhispers maps mood to 3 flavor strings shown on floor 1 entry.
var DungeonWhispers = map[DungeonMood][]string{
	MoodNeutral:   {"The dungeon is quiet.", "Stone and shadow. Nothing watches.", "Each step echoes as it always has."},
	MoodSpiteful:  {"The walls press closer this time.", "Something remembers you.", "It has been waiting."},
	MoodDeceptive: {"The shadows are patient.", "Distance is the dungeon's language.", "Eyes in the dark. Always watching from far."},
	MoodCautious:  {"The dungeon knows you linger.", "Every room passed is a debt.", "Hesitation has a sound. The dungeon hears it."},
	MoodChaotic:   {"Something revels in the noise.", "The dungeon is restless tonight.", "Chaos calls to chaos."},
}

// PickWhisper returns a deterministic whisper for the given mood and run index.
func PickWhisper(mood DungeonMood, runIndex int) string {
	lines := DungeonWhispers[mood]
	if len(lines) == 0 {
		return ""
	}
	return lines[runIndex%len(lines)]
}

// MoodName returns a human-readable mood label for debug/display.
func MoodName(mood DungeonMood) string {
	names := []string{"Neutral", "Spiteful", "Deceptive", "Cautious", "Chaotic"}
	idx := int(mood)
	if idx < 0 || idx >= len(names) {
		return "Unknown"
	}
	return names[idx]
}
