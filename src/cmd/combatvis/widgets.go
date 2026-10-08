package main

import (
	"dungeoneer/combat"
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	arenaW  = 680
	panelW  = 500
	totalW  = arenaW + panelW
	totalH  = 700
	playerX = float32(190)
	enemyX  = float32(490)
	dotY    = float32(300)

	maxLogLines = 12

	lineH = 15 // pixels per text line (DebugPrint internal font height ≈ 13px)
)

var (
	colBG       = color.RGBA{14, 14, 22, 255}
	colArenaBG  = color.RGBA{20, 20, 34, 255}
	colPanelBG  = color.RGBA{16, 16, 26, 255}
	colHeader   = color.RGBA{30, 30, 52, 255}
	colDivider  = color.RGBA{50, 50, 80, 255}
	colText     = color.RGBA{210, 210, 230, 255}
	colDim      = color.RGBA{130, 130, 160, 255}
	colGreen    = color.RGBA{80, 200, 120, 255}
	colRed      = color.RGBA{200, 70, 70, 255}
	colOrange   = color.RGBA{220, 140, 50, 255}
	colBlue     = color.RGBA{100, 160, 255, 255}
	colPurple   = color.RGBA{160, 90, 220, 255}
	colYellow   = color.RGBA{240, 210, 60, 255}
	colCDReady  = color.RGBA{70, 190, 110, 255}
	colCDActive = color.RGBA{180, 120, 40, 255}
	colCDBar    = color.RGBA{55, 55, 80, 255}
)

// domainColor maps domain name to a color for item display.
func domainColor(domain string) color.RGBA {
	switch domain {
	case "iron":
		return color.RGBA{180, 190, 200, 255}
	case "shadow":
		return color.RGBA{130, 80, 200, 255}
	case "flame":
		return color.RGBA{220, 100, 40, 255}
	case "void":
		return color.RGBA{110, 60, 170, 255}
	case "nature":
		return color.RGBA{80, 180, 100, 255}
	case "arcane":
		return color.RGBA{80, 160, 230, 255}
	default:
		return color.RGBA{160, 160, 180, 255}
	}
}

// effectDesc produces a human-readable description of an ArtifactEffect.
func effectDesc(eff combat.ArtifactEffect) string {
	var parts []string
	if eff.IsPassive {
		if eff.HealOnKillPct > 0 {
			parts = append(parts, fmt.Sprintf("heal %d%% maxHP on kill", eff.HealOnKillPct))
		}
		if eff.DamageCapPct > 0 {
			parts = append(parts, fmt.Sprintf("cap incoming dmg at %d%% maxHP", eff.DamageCapPct))
		}
		if eff.CooldownReductionPct > 0 {
			parts = append(parts, fmt.Sprintf("-%d%% all cooldowns", eff.CooldownReductionPct))
		}
		if eff.SkillDurationPct > 0 {
			parts = append(parts, fmt.Sprintf("+%d%% skill duration", eff.SkillDurationPct))
		}
		if eff.AttackSpeedPct > 0 {
			parts = append(parts, fmt.Sprintf("+%d%% attack speed", eff.AttackSpeedPct))
		}
		if eff.ShroudCooldownReset > 0 {
			parts = append(parts, fmt.Sprintf("shadow hit: -%.0fs shroud CD", eff.ShroudCooldownReset))
		}
		if eff.BurnDPS > 0 {
			parts = append(parts, fmt.Sprintf("auto hits: burn %d dps for %.0fs", eff.BurnDPS, eff.BurnDurationSec))
		}
		if len(parts) == 0 {
			parts = append(parts, "passive")
		}
		return strings.Join(parts, " · ")
	}
	// Active skill effects
	if eff.DamageMultiplier > 0 {
		parts = append(parts, fmt.Sprintf("%.1f× weapon dmg", eff.DamageMultiplier))
	}
	if eff.AOERadius > 0 {
		parts = append(parts, fmt.Sprintf("AoE r%.1f", eff.AOERadius))
	}
	if eff.IsBlinkStrike {
		parts = append(parts, "blink + guarantee crit")
	}
	if eff.IsTaunt {
		parts = append(parts, fmt.Sprintf("taunt: %d%% DR for %.0fs", eff.DamageReductionPct, eff.DurationSec))
	}
	if eff.IsRoot {
		parts = append(parts, fmt.Sprintf("root enemy for %.0fs", eff.DurationSec))
	}
	if eff.IsExecute {
		parts = append(parts, fmt.Sprintf("execute enemy below %d%% HP", eff.ExecuteThresholdPct))
	}
	if eff.SurgeDmgPerCooldown > 0 {
		parts = append(parts, fmt.Sprintf("%d dmg per artifact on CD", eff.SurgeDmgPerCooldown))
	}
	if eff.HPCostPct > 0 && eff.DamageFlat > 0 {
		parts = append(parts, fmt.Sprintf("spend %d%% HP → %d void dmg", eff.HPCostPct, eff.DamageFlat))
	}
	if eff.SpellDamageBase > 0 {
		s := fmt.Sprintf("%d base", eff.SpellDamageBase)
		if eff.SpellDamagePerINT > 0 {
			s += fmt.Sprintf(" +%.1f/INT", eff.SpellDamagePerINT)
		}
		if eff.SpellDamagePerSTR > 0 {
			s += fmt.Sprintf(" +%.1f/STR", eff.SpellDamagePerSTR)
		}
		if eff.IsAOEField {
			s += fmt.Sprintf(" field (×%.0fs)", eff.DurationSec)
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		parts = append(parts, "no registered effect")
	}
	return strings.Join(parts, " · ")
}

// ── Helper types and functions ────────────────────────────────────────────

type panelCursor struct {
	screen *ebiten.Image
	x, y   int
}

func (c *panelCursor) gap() { c.y += 6 }

func (c *panelCursor) header(title string) {
	ebitenutil.DrawRect(c.screen, float64(c.x-8), float64(c.y), float64(panelW), 1, colDivider)
	c.y += 3
	ebitenutil.DebugPrintAt(c.screen, title, c.x, c.y)
	c.y += lineH + 2
	ebitenutil.DrawRect(c.screen, float64(c.x-8), float64(c.y-3), float64(panelW), 1, colDivider)
}

func (c *panelCursor) text(s string) {
	ebitenutil.DebugPrintAt(c.screen, s, c.x, c.y)
	c.y += lineH
}

func (c *panelCursor) textCol(s string, _ color.RGBA) {
	// DebugPrintAt doesn't support per-call color; use as-is.
	ebitenutil.DebugPrintAt(c.screen, s, c.x, c.y)
	c.y += lineH
}

func (c *panelCursor) stat(label, value string, _ color.RGBA) {
	row := fmt.Sprintf("  %-18s %s", label, value)
	ebitenutil.DebugPrintAt(c.screen, row, c.x, c.y)
	c.y += lineH
}

func (c *panelCursor) statColored(label, value string, _ color.RGBA) {
	c.stat(label, value, colText)
}

func (c *panelCursor) itemRow(slot, name string, _ color.RGBA, cdStatus string, _ color.RGBA) {
	row := fmt.Sprintf("  [%s] %-26s %s", slot, name, cdStatus)
	ebitenutil.DebugPrintAt(c.screen, row, c.x, c.y)
	c.y += lineH
}

func drawBar(screen *ebiten.Image, x, y, w, h float32, cur, max int, col color.RGBA) {
	if max <= 0 {
		return
	}
	pct := float64(cur) / float64(max)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), colCDBar)
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w)*pct, float64(h), col)
}

func drawBarF(screen *ebiten.Image, x, y, w, h float32, pct float64, col color.RGBA) {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), colCDBar)
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w)*pct, float64(h), col)
}

func clampI(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func wrapText(s string, maxChars int) []string {
	if len(s) <= maxChars {
		return []string{s}
	}
	var lines []string
	for len(s) > maxChars {
		cut := maxChars
		// Try to break at a space.
		for i := maxChars - 1; i > maxChars/2; i-- {
			if s[i] == ' ' || s[i] == '·' {
				cut = i + 1
				break
			}
		}
		lines = append(lines, s[:cut])
		s = s[cut:]
	}
	if len(s) > 0 {
		lines = append(lines, s)
	}
	return lines
}
