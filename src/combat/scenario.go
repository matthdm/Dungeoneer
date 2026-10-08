package combat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Scenario defines one benchmarker test case.
type Scenario struct {
	Name          string    `json:"name"`
	Class         string    `json:"class"`     // "knight" | "mage"
	Artifacts     []string  `json:"artifacts"` // artifact IDs, up to 7
	Stats         BaseStats `json:"stats"`
	Floor         int       `json:"floor"`
	Biome         string    `json:"biome"`
	EnemyPool     []string  `json:"enemy_pool"` // enemy role names
	Iterations    int       `json:"iterations"`
	SkillRotation []string  `json:"skill_rotation"` // "auto" | "slot_N"
	EnemyBuildID  string    `json:"enemy_build"`    // optional: counter-meta enemy archetype ID
	// StatPriority defines which stat gets each of the 3 points awarded on a level-up.
	// Cycles: first entry gets point 1, second entry gets point 2, third gets point 3.
	// Valid values: "str" "vit" "int" "dex". Defaults to ["str","vit","int"] if empty.
	StatPriority []string `json:"stat_priority"`
	// Leveling lets the player gain levels and stat points from kills during
	// the run. Off by default so a result describes the build at the stats the
	// scenario states, not at whatever it snowballed into.
	Leveling bool `json:"leveling,omitempty"`
	// Distance is how far (tiles) the player stands from the enemy. 0 picks the
	// class default: melee reach for a knight, casting range for a mage.
	Distance float64 `json:"distance,omitempty"`
	// Pack is how many enemies attack at once (a wave). The player fights one
	// of them; the rest keep hitting. Area skills splash onto the rest. 0 and
	// 1 both mean a single enemy.
	Pack int `json:"pack,omitempty"`
	// RampSec raises the enemy floor by one every RampSec seconds, so enemies
	// get tougher the longer the fight goes on. 0 keeps the floor fixed.
	RampSec float64 `json:"ramp_sec,omitempty"`
}

// BaseStats mirrors entities.BaseStats but is defined here for engine independence.
type BaseStats struct {
	Strength     int `json:"Strength"`
	Dexterity    int `json:"Dexterity"`
	Vitality     int `json:"Vitality"`
	Intelligence int `json:"Intelligence"`
	Luck         int `json:"Luck"`
}

// LoadScenario reads a JSON scenario file from disk. Unknown fields and
// anything Validate rejects are errors, so a typo cannot silently run as a
// different build.
func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("combat: LoadScenario: %w", err)
	}
	var s Scenario
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return Scenario{}, fmt.Errorf("combat: LoadScenario %s: %w", path, err)
	}
	if err := s.Validate(); err != nil {
		return Scenario{}, fmt.Errorf("combat: LoadScenario %s: %w", path, err)
	}
	return s, nil
}

// Validate reports everything wrong with a scenario, joined into one error.
func (s Scenario) Validate() error {
	var errs []error
	bad := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	if s.Name == "" {
		bad("name is empty")
	}
	if s.Class != "knight" && s.Class != "mage" {
		bad("class %q is not \"knight\" or \"mage\"", s.Class)
	}
	if s.Floor < 1 {
		bad("floor %d must be at least 1", s.Floor)
	}
	if s.Iterations < 1 {
		bad("iterations %d must be at least 1", s.Iterations)
	}
	if s.Distance < 0 {
		bad("distance %g must not be negative", s.Distance)
	}
	if s.Pack < 0 || s.Pack > maxPack {
		bad("pack %d must be between 0 and %d", s.Pack, maxPack)
	}
	if s.RampSec < 0 {
		bad("ramp_sec %g must not be negative", s.RampSec)
	}
	if len(s.Artifacts) > maxArtifactSlots {
		bad("%d artifacts given, at most %d fit", len(s.Artifacts), maxArtifactSlots)
	}
	for i, id := range s.Artifacts {
		if id == "" {
			continue
		}
		if _, ok := ArtifactEffects[id]; !ok {
			bad("artifacts[%d]: unknown artifact %q", i, id)
		}
	}
	for i, entry := range s.SkillRotation {
		if entry == "auto" {
			continue
		}
		slot := parseSlotIdx(entry)
		if slot < 0 || slot >= maxArtifactSlots {
			bad("skill_rotation[%d]: %q is not \"auto\" or \"slot_1\"..\"slot_%d\"", i, entry, maxArtifactSlots)
			continue
		}
		if slot >= len(s.Artifacts) || s.Artifacts[slot] == "" {
			bad("skill_rotation[%d]: %s is empty", i, entry)
			continue
		}
		if eff, ok := ArtifactEffects[s.Artifacts[slot]]; ok && eff.IsPassive && eff.Cooldown == 0 {
			bad("skill_rotation[%d]: %s holds %q, a passive with nothing to activate", i, entry, s.Artifacts[slot])
		}
	}
	for i, stat := range s.StatPriority {
		switch stat {
		case "str", "vit", "int", "dex":
		default:
			bad("stat_priority[%d]: %q is not str, vit, int or dex", i, stat)
		}
	}
	if s.EnemyBuildID != "" {
		if _, ok := EnemyBuilds[s.EnemyBuildID]; !ok {
			bad("enemy_build: unknown enemy build %q", s.EnemyBuildID)
		}
	}
	return errors.Join(errs...)
}

// maxPack is the largest wave a scenario may ask for.
const maxPack = 20

// maxArtifactSlots is the number of artifact slots in CombatState.
const maxArtifactSlots = len(CombatState{}.EquippedArtifacts)
