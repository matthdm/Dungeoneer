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
	Active      bool
	Face        font.Face
	EquipSlots  map[string]image.Rectangle
	GridOrigin  image.Point
	CellSize    image.Point
	HoverGridX  int
	HoverGridY  int
	Dragging    bool
	DragItem    *items.Item
	DragFromX   int
	DragFromY   int
	menuActive  bool
	menuPos     image.Point
	menuOpts    []string
	menuTargetX int
	menuTargetY int
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
		DragFromX:  -1,
		DragFromY:  -1,
	}
}

func (s *InventoryScreen) Open()  { s.Active = true }
func (s *InventoryScreen) Close() { s.Active = false }

// Update handles mouse and keyboard input while the screen is open.
func (s *InventoryScreen) Update(p *entities.Player, hint func(string)) {
	if !s.Active || p == nil || p.Inventory == nil {
		return
	}
	mx, my := ebiten.CursorPosition()

	if s.menuActive {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			for i, opt := range s.menuOpts {
				r := image.Rect(s.menuPos.X, s.menuPos.Y+i*16, s.menuPos.X+80, s.menuPos.Y+(i+1)*16)
				if mx >= r.Min.X && mx <= r.Max.X && my >= r.Min.Y && my <= r.Max.Y {
					switch opt {
					case "Equip":
						if !p.Equip(autoSlot(p.Inventory.Grid[s.menuTargetY][s.menuTargetX]), s.menuTargetX, s.menuTargetY) && hint != nil {
							hint("Inventory full")
						}
					case "Drop":
						p.DropFromInventory(s.menuTargetX, s.menuTargetY, 1)
					case "Destroy":
						p.Inventory.Grid[s.menuTargetY][s.menuTargetX] = nil
					}
					s.menuActive = false
					return
				}
			}
			s.menuActive = false
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			s.menuActive = false
		}
		return
	}

	gx := (mx - s.GridOrigin.X) / s.CellSize.X
	gy := (my - s.GridOrigin.Y) / s.CellSize.Y
	s.HoverGridX, s.HoverGridY = -1, -1
	if gx >= 0 && gy >= 0 && gx < p.Inventory.Width && gy < p.Inventory.Height {
		s.HoverGridX, s.HoverGridY = gx, gy
	}

	if s.Dragging {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if s.HoverGridX >= 0 && s.HoverGridY >= 0 {
				dest := p.Inventory.Grid[s.HoverGridY][s.HoverGridX]
				if dest == nil {
					p.Inventory.Grid[s.HoverGridY][s.HoverGridX] = s.DragItem
				} else if dest.ID == s.DragItem.ID && dest.Stackable && dest.Count < dest.MaxStack {
					space := dest.MaxStack - dest.Count
					if s.DragItem.Count <= space {
						dest.Count += s.DragItem.Count
					} else {
						dest.Count = dest.MaxStack
						s.DragItem.Count -= space
						p.Inventory.Grid[s.DragFromY][s.DragFromX] = s.DragItem
					}
				} else {
					p.Inventory.Grid[s.DragFromY][s.DragFromX] = dest
					p.Inventory.Grid[s.HoverGridY][s.HoverGridX] = s.DragItem
				}
			} else {
				equipped := false
				for slot, r := range s.EquipSlots {
					if mx >= r.Min.X && mx <= r.Max.X && my >= r.Min.Y && my <= r.Max.Y {
						p.Inventory.Grid[s.DragFromY][s.DragFromX] = s.DragItem
						if !p.Equip(slot, s.DragFromX, s.DragFromY) && hint != nil {
							hint("Inventory full")
						}
						equipped = true
						break
					}
				}
				if !equipped {
					p.Inventory.Grid[s.DragFromY][s.DragFromX] = s.DragItem
				}
			}
			s.Dragging = false
			s.DragItem = nil
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			p.Inventory.Grid[s.DragFromY][s.DragFromX] = s.DragItem
			s.Dragging = false
			s.DragItem = nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			p.Inventory.Grid[s.DragFromY][s.DragFromX] = s.DragItem
			s.Dragging = false
			s.DragItem = nil
		}
		return
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && s.HoverGridX >= 0 {
		it := p.Inventory.Grid[s.HoverGridY][s.HoverGridX]
		if it != nil {
			s.Dragging = true
			s.DragItem = it
			s.DragFromX, s.DragFromY = s.HoverGridX, s.HoverGridY
			p.Inventory.Grid[s.DragFromY][s.DragFromX] = nil
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && s.HoverGridX >= 0 {
		it := p.Inventory.Grid[s.HoverGridY][s.HoverGridX]
		if it != nil {
			s.menuActive = true
			s.menuPos = image.Pt(mx, my)
			s.menuTargetX, s.menuTargetY = s.HoverGridX, s.HoverGridY
			s.menuOpts = s.menuOpts[:0]
			if it.Equippable {
				s.menuOpts = append(s.menuOpts, "Equip")
			}
			s.menuOpts = append(s.menuOpts, "Drop", "Destroy")
		}
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
	mx, my := ebiten.CursorPosition()
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
			if x == s.HoverGridX && y == s.HoverGridY {
				vector.StrokeRect(dst, float32(px), float32(py), float32(s.CellSize.X), float32(s.CellSize.Y), 3, color.RGBA{255, 255, 0, 255}, false)
			}
		}
	}

	// Tooltip
	if s.HoverGridX >= 0 && s.HoverGridY >= 0 && !s.menuActive {
		if it := p.Inventory.Grid[s.HoverGridY][s.HoverGridX]; it != nil {
			drawItemTooltip(dst, it, mx+16, my+16)
		}
	}

	// Dragged item
	if s.Dragging && s.DragItem != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(mx-s.CellSize.X/2), float64(my-s.CellSize.Y/2))
		if s.DragItem.Icon != nil {
			dst.DrawImage(s.DragItem.Icon, op)
		} else {
			ebitenutil.DebugPrintAt(dst, truncate(s.DragItem.Name, 8), mx-s.CellSize.X/2+2, my-s.CellSize.Y/2+2)
		}
	}

	// Context menu
	if s.menuActive {
		width := 80
		height := len(s.menuOpts) * 16
		vector.DrawFilledRect(dst, float32(s.menuPos.X), float32(s.menuPos.Y), float32(width), float32(height), color.RGBA{0, 0, 0, 200}, false)
		for i, opt := range s.menuOpts {
			ebitenutil.DebugPrintAt(dst, opt, s.menuPos.X+2, s.menuPos.Y+i*16+2)
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

func drawItemTooltip(dst *ebiten.Image, it *items.Item, x, y int) {
	lines := []string{it.Name}
	if it.Description != "" {
		lines = append(lines, it.Description)
	}
	if len(it.Stats) > 0 {
		for stat, v := range it.Stats {
			lines = append(lines, fmt.Sprintf("%s %+d", stat, v))
		}
	}
	if it.Effect != nil {
		line := fmt.Sprintf("%s %s %d%%", it.Effect.Trigger, it.Effect.Type, it.Effect.MagnitudePct)
		lines = append(lines, line)
		if it.Effect.ChancePct != 0 {
			lines = append(lines, fmt.Sprintf("Chance %d%%", it.Effect.ChancePct))
		}
	}
	width := 0
	for _, ln := range lines {
		if w := len(ln) * 7; w > width {
			width = w
		}
	}
	height := len(lines) * 14
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(width+4), float32(height+4), color.RGBA{0, 0, 0, 200}, false)
	for i, ln := range lines {
		ebitenutil.DebugPrintAt(dst, ln, x+2, y+2+i*14)
	}
}
