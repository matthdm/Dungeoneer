package hud

import (
	"fmt"
	"image"
	"image/color"

	"dungeoneer/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// SkillSlot represents a single skill slot in the HUD.
type SkillSlot struct {
	Icon     *ebiten.Image
	Cooldown float64
}

// HUD renders a bottom-screen interface similar to classic action RPGs.
type HUD struct {
	HealthPercent float64
	ManaPercent   float64
	DashCharges   int
	DashCooldown  float64
	ExpCurrent    int
	ExpNeeded     int
	SkillSlots    [5]SkillSlot
	ActiveSkill   int

	OrbFrame      *ebiten.Image
	HUDBackground *ebiten.Image

	orbSize int
	orbFill *ebiten.Image
}

// New creates a HUD with default sizes.
func New() *HUD {
	h := &HUD{orbSize: 96}
	h.initOrbResources()
	return h
}

func (h *HUD) initOrbResources() {
	h.orbFill = ebiten.NewImage(h.orbSize, h.orbSize)
	h.orbFill.Fill(color.Transparent)
}

// Update decrements any cooldown timers.
func (h *HUD) Update(dt float64) {
	if h.DashCooldown > 0 {
		h.DashCooldown -= dt
		if h.DashCooldown < 0 {
			h.DashCooldown = 0
		}
	}
	for i := range h.SkillSlots {
		if h.SkillSlots[i].Cooldown > 0 {
			h.SkillSlots[i].Cooldown -= dt
			if h.SkillSlots[i].Cooldown < 0 {
				h.SkillSlots[i].Cooldown = 0
			}
		}
	}
}

// Draw renders the HUD anchored to the bottom of the screen.
func (h *HUD) Draw(screen *ebiten.Image, w, hgt int) {
	if h.orbFill == nil || h.orbFill.Bounds().Dx() != h.orbSize {
		h.initOrbResources()
	}

	if h.HUDBackground != nil {
		bw, bh := h.HUDBackground.Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64((w-bw)/2), float64(hgt-bh))
		screen.DrawImage(h.HUDBackground, op)
	}

	margin := 10
	y := hgt - h.orbSize - margin
	drawOrb(screen, margin, y, h.orbSize, h.HealthPercent, color.RGBA{200, 0, 0, 255}, h.OrbFrame, h.orbFill)
	drawOrb(screen, w-h.orbSize-margin, y, h.orbSize, h.ManaPercent, color.RGBA{0, 0, 200, 255}, h.OrbFrame, h.orbFill)

	h.drawSkillBar(screen, w, hgt)
}

func (h *HUD) drawSkillBar(screen *ebiten.Image, w, hgt int) {
	slot := 64
	pad := 6
	barW := slot*5 + pad*4
	x := (w - barW) / 2
	y := hgt - slot - 20

	for i := 0; i < 5; i++ {
		sx := x + i*(slot+pad)
		vector.StrokeRect(screen, float32(sx), float32(y), float32(slot), float32(slot), 2, color.White, false)

		if ic := h.SkillSlots[i].Icon; ic != nil {
			iw, ih := ic.Size()
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(slot)/float64(iw), float64(slot)/float64(ih))
			op.GeoM.Translate(float64(sx), float64(y))
			screen.DrawImage(ic, op)
		}

		if cd := h.SkillSlots[i].Cooldown; cd > 0 {
			overlay := float32(slot) * float32(cd) / 5
			vector.DrawFilledRect(screen, float32(sx), float32(y)+float32(slot)-overlay, float32(slot), overlay, color.RGBA{0, 0, 0, 150}, false)
		}

		if i == h.ActiveSkill {
			vector.StrokeRect(screen, float32(sx-2), float32(y-2), float32(slot+4), float32(slot+4), 2, color.RGBA{255, 220, 60, 255}, false)
		}

		text.Draw(screen, fmt.Sprintf("%d", i+1), basicfont.Face7x13, sx+slot/2-4, y+slot+12, color.White)
	}

	h.drawDashCharges(screen, x, barW, y)
	h.drawEXPBar(screen, x, barW, y)
}

func (h *HUD) drawDashCharges(screen *ebiten.Image, barX, barW, barY int) {
	size := 18
	pad := 4
	total := constants.MaxDashCharges
	start := barX + barW/2 - ((size+pad)*total-pad)/2
	y := barY - size - 6

	for i := 0; i < total; i++ {
		dx := start + i*(size+pad)
		vector.StrokeRect(screen, float32(dx), float32(y), float32(size), float32(size), 2, color.White, false)
		if i < h.DashCharges {
			vector.DrawFilledRect(screen, float32(dx), float32(y), float32(size), float32(size), color.RGBA{140, 0, 230, 255}, false)
		}
	}

	if h.DashCharges < total {
		txt := fmt.Sprintf("\u231B %.1fs", h.DashCooldown)
		b := text.BoundString(basicfont.Face7x13, txt)
		tx := start + (size*total+pad*(total-1)-b.Dx())/2
		text.Draw(screen, txt, basicfont.Face7x13, tx, y-4, color.White)
	}
}

func (h *HUD) drawEXPBar(screen *ebiten.Image, barX, barW, barY int) {
	if h.ExpNeeded <= 0 {
		return
	}
	dashSize := 18
	y := barY - dashSize - 6 - 10
	barH := 8
	filled := int(float64(h.ExpCurrent) / float64(h.ExpNeeded) * float64(barW))
	vector.DrawFilledRect(screen, float32(barX), float32(y), float32(barW), float32(barH), color.RGBA{80, 80, 80, 255}, false)
	vector.DrawFilledRect(screen, float32(barX), float32(y), float32(filled), float32(barH), color.RGBA{0, 200, 0, 255}, false)
}

func drawOrb(dst *ebiten.Image, x, y, size int, percent float64, clr color.Color, frame *ebiten.Image, buf *ebiten.Image) {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	buf.Clear()
	vector.DrawFilledCircle(buf, float32(size)/2, float32(size)/2, float32(size)/2, clr, true)

	sy := int(float64(size) * (1 - percent))
	src := buf.SubImage(image.Rect(0, sy, size, size)).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y+sy))
	dst.DrawImage(src, op)

	if frame != nil {
		fw, fh := frame.Size()
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Scale(float64(size)/float64(fw), float64(size)/float64(fh))
		op2.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(frame, op2)
	}
}
