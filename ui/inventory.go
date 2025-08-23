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
	Active         bool
	Face           font.Face
	EquipSlots     map[string]image.Rectangle
	GridOrigin     image.Point
	CellSize       image.Point
	HoverGridX     int
	HoverGridY     int
	Dragging       bool
	DragItem       *items.Item
	DragFromX      int
	DragFromY      int
	menuActive     bool
	menuPos        image.Point
	menuOpts       []string
	menuTargetX    int
	menuTargetY    int
	menuSlot       string
	menuHover      int
	confirmActive  bool
	confirmItem    *items.Item
	confirmTargetX int
	confirmTargetY int
	confirmSlot    string
	confirmPos     image.Point
	confirmHover   int
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
		Face:           basicfont.Face7x13,
		EquipSlots:     slots,
		GridOrigin:     image.Pt(350, 40),
		CellSize:       image.Pt(64, 64),
		HoverGridX:     -1,
		HoverGridY:     -1,
		DragFromX:      -1,
		DragFromY:      -1,
		menuHover:      -1,
		confirmTargetX: -1,
		confirmTargetY: -1,
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

	if s.confirmActive {
		s.confirmHover = -1
		yes := image.Rect(s.confirmPos.X+20, s.confirmPos.Y+40, s.confirmPos.X+80, s.confirmPos.Y+56)
		no := image.Rect(s.confirmPos.X+110, s.confirmPos.Y+40, s.confirmPos.X+170, s.confirmPos.Y+56)
		if mx >= yes.Min.X && mx <= yes.Max.X && my >= yes.Min.Y && my <= yes.Max.Y {
			s.confirmHover = 0
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if s.confirmSlot != "" {
					p.DropEquipped(s.confirmSlot)
				} else if s.confirmTargetY >= 0 {
					p.Inventory.Grid[s.confirmTargetY][s.confirmTargetX] = nil
				}
				s.confirmActive = false
				s.confirmItem = nil
				s.confirmSlot = ""
				s.confirmTargetX, s.confirmTargetY = -1, -1
				return
			}
		} else if mx >= no.Min.X && mx <= no.Max.X && my >= no.Min.Y && my <= no.Max.Y {
			s.confirmHover = 1
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				s.confirmActive = false
				s.confirmItem = nil
				s.confirmSlot = ""
				s.confirmTargetX, s.confirmTargetY = -1, -1
				return
			}
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			s.confirmActive = false
			s.confirmItem = nil
			s.confirmSlot = ""
			s.confirmTargetX, s.confirmTargetY = -1, -1
		}
		return
	}

	if s.menuActive {
		s.menuHover = -1
		for i, opt := range s.menuOpts {
			r := image.Rect(s.menuPos.X, s.menuPos.Y+i*16, s.menuPos.X+80, s.menuPos.Y+(i+1)*16)
			if mx >= r.Min.X && mx <= r.Max.X && my >= r.Min.Y && my <= r.Max.Y {
				s.menuHover = i
				if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
					switch opt {
					case "Equip":
						it := p.Inventory.Grid[s.menuTargetY][s.menuTargetX]
						slot := autoSlot(p, it)
						if slot == "" || !p.Equip(slot, s.menuTargetX, s.menuTargetY) {
							if hint != nil {
								hint("Cannot equip")
							}
						}
					case "Drop":
						if s.menuSlot != "" {
							p.DropEquipped(s.menuSlot)
						} else {
							p.DropFromInventory(s.menuTargetX, s.menuTargetY, 1)
						}
					case "Destroy":
						if s.menuSlot != "" {
							s.confirmItem = p.Equipment[s.menuSlot]
							s.confirmSlot = s.menuSlot
							s.confirmTargetX, s.confirmTargetY = -1, -1
						} else {
							s.confirmItem = p.Inventory.Grid[s.menuTargetY][s.menuTargetX]
							s.confirmTargetX, s.confirmTargetY = s.menuTargetX, s.menuTargetY
							s.confirmSlot = ""
						}
						s.confirmPos = s.menuPos
						s.confirmActive = true
					case "Unequip":
						if !p.Unequip(s.menuSlot) && hint != nil {
							hint("Inventory full")
						}
					}
					s.menuActive = false
					return
				}
			}
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
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
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if s.HoverGridX >= 0 {
			it := p.Inventory.Grid[s.HoverGridY][s.HoverGridX]
			if it != nil {
				s.menuActive = true
				s.menuPos = image.Pt(mx, my)
				s.menuTargetX, s.menuTargetY = s.HoverGridX, s.HoverGridY
				s.menuSlot = ""
				s.menuOpts = s.menuOpts[:0]
				if it.Equippable {
					s.menuOpts = append(s.menuOpts, "Equip")
				}
				s.menuOpts = append(s.menuOpts, "Drop", "Destroy")
			}
		} else {
			for slot, r := range s.EquipSlots {
				if mx >= r.Min.X && mx <= r.Max.X && my >= r.Min.Y && my <= r.Max.Y {
					it := p.Equipment[slot]
					if it != nil {
						s.menuActive = true
						s.menuPos = image.Pt(mx, my)
						s.menuSlot = slot
						s.menuTargetX, s.menuTargetY = -1, -1
						s.menuOpts = []string{"Unequip", "Drop", "Destroy"}
					}
					break
				}
			}
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
	eff := p.EffectiveStats()
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("STR %d", eff.Strength), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("DEX %d", eff.Dexterity), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("VIT %d", eff.Vitality), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("INT %d", eff.Intelligence), x, y)
	y += 15
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("LUK %d", eff.Luck), x, y)

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
	if s.HoverGridX >= 0 && s.HoverGridY >= 0 && !s.menuActive && !s.confirmActive {
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
			if i == s.menuHover {
				vector.DrawFilledRect(dst, float32(s.menuPos.X), float32(s.menuPos.Y+i*16), float32(width), 16, color.RGBA{80, 80, 80, 255}, false)
			}
			ebitenutil.DebugPrintAt(dst, opt, s.menuPos.X+2, s.menuPos.Y+i*16+2)
		}
	}

	if s.confirmActive {
		width := 200
		height := 60
		x := s.confirmPos.X
		y := s.confirmPos.Y
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(width), float32(height), color.RGBA{0, 0, 0, 200}, false)
		msg := fmt.Sprintf("Are you sure you want to destroy %s?", s.confirmItem.Name)
		ebitenutil.DebugPrintAt(dst, truncate(msg, 28), x+4, y+4)
		yes := image.Rect(x+20, y+40, x+80, y+56)
		no := image.Rect(x+110, y+40, x+170, y+56)
		if s.confirmHover == 0 {
			vector.DrawFilledRect(dst, float32(yes.Min.X), float32(yes.Min.Y), float32(yes.Dx()), float32(yes.Dy()), color.RGBA{80, 80, 80, 255}, false)
		} else {
			vector.DrawFilledRect(dst, float32(yes.Min.X), float32(yes.Min.Y), float32(yes.Dx()), float32(yes.Dy()), color.RGBA{40, 40, 40, 255}, false)
		}
		if s.confirmHover == 1 {
			vector.DrawFilledRect(dst, float32(no.Min.X), float32(no.Min.Y), float32(no.Dx()), float32(no.Dy()), color.RGBA{80, 80, 80, 255}, false)
		} else {
			vector.DrawFilledRect(dst, float32(no.Min.X), float32(no.Min.Y), float32(no.Dx()), float32(no.Dy()), color.RGBA{40, 40, 40, 255}, false)
		}
		ebitenutil.DebugPrintAt(dst, "Yes", yes.Min.X+2, yes.Min.Y+2)
		ebitenutil.DebugPrintAt(dst, "No", no.Min.X+2, no.Min.Y+2)
	}
}

func autoSlot(p *entities.Player, it *items.Item) string {
	switch it.Type {
	case items.ItemWeapon:
		return "Weapon"
	case items.ItemArmor:
		return "Chest"
	}
	order := []string{"Head", "Chest", "Weapon", "Offhand", "Ring1", "Ring2"}
	for _, slot := range order {
		if p.Equipment[slot] == nil {
			return slot
		}
	}
	return ""
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
		order := []string{"Strength", "Dexterity", "Vitality", "Intelligence", "Luck"}
		for _, stat := range order {
			if v, ok := it.Stats[stat]; ok {
				lines = append(lines, fmt.Sprintf("%s %+d", stat, v))
			}
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
