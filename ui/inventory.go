package ui

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// InventoryScreen renders and manages the player's inventory and equipment.
type InventoryScreen struct {
	Active     bool
	Face       font.Face
	EquipSlots map[string]image.Rectangle
	GridOrigin image.Point
	CellSize   image.Point
	HoverGridX int
	HoverGridY int
	SelX       int
	SelY       int
}

// NewInventoryScreen creates a screen with default layout values.
func NewInventoryScreen() *InventoryScreen {
	slots := map[string]image.Rectangle{
		"Head":    image.Rect(200, 40, 264, 104),
		"Chest":   image.Rect(200, 110, 264, 174),
		"Weapon":  image.Rect(130, 110, 194, 174),
		"Offhand": image.Rect(270, 110, 334, 174),
		"Ring1":   image.Rect(130, 180, 194, 244),
		"Ring2":   image.Rect(270, 180, 334, 244),
	}
	return &InventoryScreen{
		Face:       basicfont.Face7x13,
		EquipSlots: slots,
		GridOrigin: image.Pt(350, 40),
		CellSize:   image.Pt(64, 64),
		HoverGridX: -1,
		HoverGridY: -1,
		SelX:       -1,
		SelY:       -1,
	}
}

func (s *InventoryScreen) Open()  { s.Active = true }
func (s *InventoryScreen) Close() { s.Active = false }

// Update handles mouse and keyboard input while the screen is open.
func (s *InventoryScreen) Update(p *entities.Player) {
	if !s.Active || p == nil || p.Inventory == nil {
		return
	}
	mx, my := ebiten.CursorPosition()
	gx := (mx - s.GridOrigin.X) / s.CellSize.X
	gy := (my - s.GridOrigin.Y) / s.CellSize.Y
	s.HoverGridX, s.HoverGridY = -1, -1
	if gx >= 0 && gy >= 0 && gx < p.Inventory.Width && gy < p.Inventory.Height {
		s.HoverGridX, s.HoverGridY = gx, gy
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			s.SelX, s.SelY = gx, gy
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) && s.SelX >= 0 {
		it := p.Inventory.Grid[s.SelY][s.SelX]
		if it != nil && it.Equippable {
			slot := autoSlot(it)
			if slot != "" {
				p.Equip(slot, s.SelX, s.SelY)
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyU) && s.SelX >= 0 {
		it := p.Inventory.Grid[s.SelY][s.SelX]
		if it != nil && it.Usable && it.OnUse != nil {
			it.OnUse(p)
			it.Count--
			if it.Count <= 0 {
				p.Inventory.Grid[s.SelY][s.SelX] = nil
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && s.SelX >= 0 {
		p.DropFromInventory(s.SelX, s.SelY, 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		s.Close()
	}
}

// Draw renders the inventory screen.
func (s *InventoryScreen) Draw(dst *ebiten.Image, p *entities.Player) {
	if !s.Active || p == nil || p.Inventory == nil {
		return
	}
	DrawMenuOverlay(dst, DefaultOverlayColor)
	// Stats column
	x := 20
	y := 40
	ebitenutil.DebugPrintAt(dst, p.Name, x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("Level: %d", p.Level), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("EXP: %d", p.EXP), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("HP: %d/%d", p.HP, p.MaxHP), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("MP: %d/%d", p.Mana, p.MaxMana), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("Gold: %d", p.Gold), x, y)
	y += 25
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("STR %d", p.Stats.Strength), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("DEX %d", p.Stats.Dexterity), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("VIT %d", p.Stats.Vitality), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("INT %d", p.Stats.Intelligence), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("LUK %d", p.Stats.Luck), x, y)

	// Equipment slots
	for slot, r := range s.EquipSlots {
		vector.StrokeRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 2, color.White, false)
		if it := p.Equipment[slot]; it != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
			if it.Icon != nil {
				dst.DrawImage(it.Icon, op)
			} else {
				ebitenutil.DebugPrintAt(dst, truncate(it.Name, 8), r.Min.X+2, r.Min.Y+2)
			}
		}
	}

	// Inventory grid
	for y := 0; y < p.Inventory.Height; y++ {
		for x := 0; x < p.Inventory.Width; x++ {
			px := s.GridOrigin.X + x*s.CellSize.X
			py := s.GridOrigin.Y + y*s.CellSize.Y
			vector.StrokeRect(dst, float32(px), float32(py), float32(s.CellSize.X), float32(s.CellSize.Y), 2, color.White, false)
			it := p.Inventory.Grid[y][x]
			if it != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(px), float64(py))
				if it.Icon != nil {
					dst.DrawImage(it.Icon, op)
				} else {
					ebitenutil.DebugPrintAt(dst, truncate(it.Name, 8), px+2, py+2)
				}
				if it.Count > 1 {
					ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%dx", it.Count), px+2, py+s.CellSize.Y-12)
				}
			}
			if x == s.SelX && y == s.SelY {
				vector.StrokeRect(dst, float32(px), float32(py), float32(s.CellSize.X), float32(s.CellSize.Y), 3, color.RGBA{255, 255, 0, 255}, false)
			}
		}
	}
}

func autoSlot(it *items.Item) string {
	switch it.Type {
	case items.ItemWeapon:
		return "Weapon"
	case items.ItemArmor:
		return "Chest"
	default:
		return ""
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
