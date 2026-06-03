package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// UpgradeEntry is the display model for one upgrade.
type UpgradeEntry struct {
	ID           string
	Name         string
	Description  string
	CurrentLevel int
	MaxLevel     int
	NextCost     int
	CanAfford    bool
}

// UpgradeStation is the hub upgrade station UI.
type UpgradeStation struct {
	Active      bool
	Entries     []UpgradeEntry
	SelectedIdx int
	Remnants    int
	screenW     int
	screenH     int
	OnUpgrade   func(upgradeID string, cost int) bool
}

// NewUpgradeStation creates a closed upgrade station sized to the given screen dimensions.
func NewUpgradeStation(w, h int) *UpgradeStation {
	return &UpgradeStation{screenW: w, screenH: h}
}

// Open shows the upgrade station with the given entries and current remnant balance.
func (u *UpgradeStation) Open(entries []UpgradeEntry, remnants int) {
	u.Entries = entries
	u.Remnants = remnants
	u.Active = true
	u.SelectedIdx = 0
}

// Close hides the upgrade station.
func (u *UpgradeStation) Close() { u.Active = false }

// Resize updates screen dimensions (call on window resize).
func (u *UpgradeStation) Resize(w, h int) { u.screenW = w; u.screenH = h }

// Update handles input. Returns early if upgrade station is not active.
func (u *UpgradeStation) Update() {
	if !u.Active {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		u.Active = false
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && u.SelectedIdx < len(u.Entries)-1 {
		u.SelectedIdx++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && u.SelectedIdx > 0 {
		u.SelectedIdx--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) && len(u.Entries) > 0 {
		entry := &u.Entries[u.SelectedIdx]
		if entry.CanAfford && entry.CurrentLevel < entry.MaxLevel {
			if u.OnUpgrade != nil && u.OnUpgrade(entry.ID, entry.NextCost) {
				entry.CurrentLevel++
				u.Remnants -= entry.NextCost
				for i := range u.Entries {
					u.Entries[i].CanAfford = u.Remnants >= u.Entries[i].NextCost
				}
			}
		}
	}
}

// Draw renders the upgrade station UI.
func (u *UpgradeStation) Draw(screen *ebiten.Image) {
	if !u.Active {
		return
	}
	w, h := u.screenW, u.screenH

	bg := ebiten.NewImage(w, h)
	bg.Fill(color.NRGBA{20, 5, 5, 240})
	screen.DrawImage(bg, &ebiten.DrawImageOptions{})

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("UPGRADES   Remnants: %d   [E] Upgrade  [Esc] Close", u.Remnants), 20, 12)

	y := 32
	for i, entry := range u.Entries {
		rowH := 40
		rowBg := ebiten.NewImage(w-40, rowH-2)
		if i == u.SelectedIdx {
			rowBg.Fill(color.NRGBA{80, 40, 20, 200})
		} else {
			rowBg.Fill(color.NRGBA{30, 15, 10, 180})
		}
		rop := &ebiten.DrawImageOptions{}
		rop.GeoM.Translate(20, float64(y))
		screen.DrawImage(rowBg, rop)

		maxedStr := ""
		costStr := fmt.Sprintf("  Next: %d Remnants", entry.NextCost)
		if entry.CurrentLevel >= entry.MaxLevel {
			maxedStr = " [MAX]"
			costStr = ""
		}
		label := fmt.Sprintf("%s (Lv %d/%d)%s%s", entry.Name, entry.CurrentLevel, entry.MaxLevel, maxedStr, costStr)
		ebitenutil.DebugPrintAt(screen, label, 28, y+4)
		ebitenutil.DebugPrintAt(screen, entry.Description, 28, y+18)
		y += rowH
		if y > h-20 {
			break
		}
	}
}
