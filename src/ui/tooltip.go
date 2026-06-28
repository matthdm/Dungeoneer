package ui

import (
	"dungeoneer/items"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// QualityColor returns the display color for a given quality tier.
func QualityColor(quality string) color.RGBA {
	switch quality {
	case items.RarityUncommon:
		return color.RGBA{80, 220, 80, 255}
	case items.RarityRare:
		return color.RGBA{80, 140, 255, 255}
	case items.RarityLegendary:
		return color.RGBA{255, 160, 0, 255}
	default: // common or empty
		return color.RGBA{200, 200, 200, 255}
	}
}

// DrawItemTooltip renders an item tooltip anchored at (x, y), clamped to the
// screen so it never overflows the window bounds.
func DrawItemTooltip(dst *ebiten.Image, it *items.Item, x, y int) {
	const (
		lineH  = 15
		padX   = 6
		padY   = 6
		charW  = 7
		minW   = 120
		ascent = 11 // basicfont.Face7x13 ascent in pixels
	)

	type tline struct {
		text string
		clr  color.RGBA
	}
	var lines []tline

	nameClr := QualityColor(it.Quality)
	lines = append(lines, tline{it.Name, nameClr})

	typeStr := string(it.Type)
	if it.GrantsAbility != "" {
		typeStr += " — grants " + it.GrantsAbility
	}
	if typeStr != "" {
		lines = append(lines, tline{typeStr, color.RGBA{160, 160, 160, 255}})
	}

	if it.Description != "" {
		lines = append(lines, tline{it.Description, color.RGBA{220, 220, 180, 255}})
	}

	if len(it.Stats) > 0 {
		order := []string{"Strength", "Dexterity", "Vitality", "Intelligence", "Luck"}
		for _, stat := range order {
			if v, ok := it.Stats[stat]; ok {
				clr := color.RGBA{80, 220, 80, 255}
				if v < 0 {
					clr = color.RGBA{220, 80, 80, 255}
				}
				lines = append(lines, tline{fmt.Sprintf("%s %+d", stat, v), clr})
			}
		}
	}

	if it.Effect != nil {
		eff := it.Effect
		var effectDesc string
		if eff.MagnitudePct != 0 {
			effectDesc = fmt.Sprintf("%d%% %s", eff.MagnitudePct, formatEffectType(eff.Type))
		} else if eff.MagnitudeFlat != 0 {
			effectDesc = fmt.Sprintf("%+d %s", eff.MagnitudeFlat, formatEffectType(eff.Type))
		} else {
			effectDesc = formatEffectType(eff.Type)
		}

		if eff.Element != "" {
			effectDesc += fmt.Sprintf(" (%s)", eff.Element)
		}

		triggerDesc := formatTrigger(eff.Trigger)
		if eff.Trigger == "on_low_hp" {
			threshold := eff.ThresholdPct
			if threshold <= 0 {
				threshold = 20
			}
			triggerDesc = fmt.Sprintf("At <%d%% HP", threshold)
		}

		txt := fmt.Sprintf("%s: %s", triggerDesc, effectDesc)
		if eff.ChancePct != 0 {
			txt += fmt.Sprintf(" (%d%% chance)", eff.ChancePct)
		}
		if eff.DurationSec != 0 {
			txt += fmt.Sprintf(" for %.1fs", eff.DurationSec)
		}
		if eff.IntervalSec != 0 {
			txt += fmt.Sprintf(" every %.1fs", eff.IntervalSec)
		}
		if eff.CooldownSec != 0 {
			txt += fmt.Sprintf(" (cooldown %.1fs)", eff.CooldownSec)
		}

		lines = append(lines, tline{txt, color.RGBA{200, 180, 255, 255}})
	}

	// Show set membership if this item belongs to a set.
	for _, set := range items.SetRegistry {
		for _, mid := range set.MemberIDs {
			if mid == it.ID {
				lines = append(lines, tline{
					fmt.Sprintf("Set: %s (%d pieces)", set.Name, len(set.MemberIDs)),
					color.RGBA{255, 200, 80, 255},
				})
				break
			}
		}
	}

	if it.FlavorText != "" {
		lines = append(lines, tline{it.FlavorText, color.RGBA{160, 140, 120, 255}})
	}

	// Measure width and height.
	w := minW
	for _, ln := range lines {
		if cw := len(ln.text)*charW + padX*2; cw > w {
			w = cw
		}
	}
	h := len(lines)*lineH + padY*2

	// Clamp to screen edges.
	sw, sh := dst.Bounds().Dx(), dst.Bounds().Dy()
	if x+w > sw {
		x = sw - w - 4
	}
	if y+h > sh {
		y = sh - h - 4
	}

	// Background panel.
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{10, 10, 20, 220}, false)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, nameClr, false)

	// Draw each line with its color. text.Draw baseline = y + ascent.
	for i, ln := range lines {
		ty := y + padY + i*lineH + ascent
		text.Draw(dst, ln.text, basicfont.Face7x13, x+padX, ty, ln.clr)
	}
}

func formatEffectType(t string) string {
	switch t {
	case "damage_reduction":
		return "Damage Reduction"
	case "all_resistance":
		return "All Resistance"
	case "lifesteal":
		return "Lifesteal"
	case "crit_multiplier":
		return "Crit Multiplier"
	case "gold_find":
		return "Gold Find"
	case "mana_cost_reduction":
		return "Mana Cost Reduction"
	case "cooldown_reduction":
		return "Cooldown Reduction"
	case "regen_hp":
		return "HP Regen"
	case "heal":
		return "Heal"
	case "bonus_healing":
		return "Bonus Healing"
	case "counterpulse":
		return "Counterpulse"
	case "damage_reduction_buff":
		return "Damage Reduction Buff"
	default:
		return t
	}
}

func formatTrigger(t string) string {
	switch t {
	case "passive":
		return "Passive"
	case "on_kill":
		return "On Kill"
	case "on_hit":
		return "On Hit"
	case "on_low_hp":
		return "On Low HP"
	case "on_block":
		return "On Block"
	case "on_potion_use":
		return "On Potion Use"
	case "regen_hp":
		return "Regen HP"
	default:
		return t
	}
}
