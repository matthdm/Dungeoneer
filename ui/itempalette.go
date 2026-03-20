package ui

import (
	"dungeoneer/entities"
	"dungeoneer/items"
	"fmt"
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ipColumns    = 4
	ipRows       = 3
	ipSpriteSize = 64
	ipPadding    = 5
	ipNameHeight = 12
)

// ItemPalette is a page-based icon+name grid for spawning items.
// Navigate pages with Prev/Next buttons; no mouse-wheel scroll to avoid
// conflicting with the camera zoom.
type ItemPalette struct {
	visible      bool
	rect         image.Rectangle
	currentPage  int
	itemsPerPage int
	itemIDs      []string
	prevButton   image.Rectangle
	nextButton   image.Rectangle
	player       *entities.Player
	hint         func(string)
}

// NewItemPalette creates an ItemPalette centered on the screen.
func NewItemPalette(screenW, screenH int, player *entities.Player, hint func(string)) *ItemPalette {
	ip := &ItemPalette{
		itemsPerPage: ipColumns * ipRows,
		player:       player,
		hint:         hint,
	}
	ip.rect = ip.computeRect(screenW, screenH)
	return ip
}

func (ip *ItemPalette) computeRect(screenW, screenH int) image.Rectangle {
	cellW := ipSpriteSize + ipPadding
	cellH := ipSpriteSize + ipNameHeight + ipPadding
	gridW := ipColumns*cellW - ipPadding
	gridH := ipRows*cellH - ipPadding
	menuW := gridW + 20            // 10px inner padding left+right
	menuH := 20 + 30 + gridH + 14 // title + nav row + grid + bottom padding
	mx := (screenW - menuW) / 2
	my := (screenH - menuH) / 2
	return image.Rect(mx, my, mx+menuW, my+menuH)
}

func (ip *ItemPalette) refreshItemIDs() {
	ip.itemIDs = ip.itemIDs[:0]
	for id := range items.Registry {
		ip.itemIDs = append(ip.itemIDs, id)
	}
	sort.Strings(ip.itemIDs)
}

func (ip *ItemPalette) Show() {
	ip.refreshItemIDs()
	ip.visible = true
}

func (ip *ItemPalette) Hide()           { ip.visible = false }
func (ip *ItemPalette) Toggle()         { if ip.visible { ip.Hide() } else { ip.Show() } }
func (ip *ItemPalette) IsVisible() bool { return ip.visible }

// Resize recomputes the palette rect for the new screen dimensions.
func (ip *ItemPalette) Resize(screenW, screenH int) {
	ip.rect = ip.computeRect(screenW, screenH)
}

func (ip *ItemPalette) maxPage() int {
	if len(ip.itemIDs) == 0 {
		return 0
	}
	return (len(ip.itemIDs) - 1) / ip.itemsPerPage
}

// Update handles mouse input for page navigation and item spawning.
func (ip *ItemPalette) Update() {
	if !ip.visible {
		return
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	x, y := ebiten.CursorPosition()

	if pointInRect(x, y, ip.prevButton) {
		if ip.currentPage > 0 {
			ip.currentPage--
		}
		return
	}
	if pointInRect(x, y, ip.nextButton) {
		if ip.currentPage < ip.maxPage() {
			ip.currentPage++
		}
		return
	}

	// Click outside the menu window → close
	if !pointInRect(x, y, ip.rect) {
		ip.Hide()
		return
	}

	// Grid click
	gridX := ip.rect.Min.X + ipPadding
	gridY := ip.rect.Min.Y + 20 + 30 // title height + nav bar height
	ox := x - gridX
	oy := y - gridY
	if ox < 0 || oy < 0 {
		return
	}
	cellW := ipSpriteSize + ipPadding
	cellH := ipSpriteSize + ipNameHeight + ipPadding
	col := ox / cellW
	row := oy / cellH
	if col >= ipColumns || row >= ipRows {
		return
	}

	idx := ip.currentPage*ip.itemsPerPage + row*ipColumns + col
	if idx >= 0 && idx < len(ip.itemIDs) {
		id := ip.itemIDs[idx]
		if ip.player != nil && ip.player.Inventory != nil {
			if !ip.player.AddToInventory(items.NewItem(id)) && ip.hint != nil {
				ip.hint("Inventory full")
			}
		}
	}
}

// Draw renders the item palette.
func (ip *ItemPalette) Draw(screen *ebiten.Image) {
	if !ip.visible {
		return
	}

	DrawMenuOverlay(screen, DefaultOverlayColor)

	r := ip.rect
	vector.DrawFilledRect(screen,
		float32(r.Min.X), float32(r.Min.Y),
		float32(r.Dx()), float32(r.Dy()),
		DefaultBackgroundColor, false)
	vector.StrokeRect(screen,
		float32(r.Min.X), float32(r.Min.Y),
		float32(r.Dx()), float32(r.Dy()),
		DefaultBorderThickness, DefaultBorderColor, false)

	// Title
	title := "SPAWN ITEM"
	titleX := r.Min.X + (r.Dx()-len(title)*6)/2
	ebitenutil.DebugPrintAt(screen, title, titleX, r.Min.Y+5)

	// Pagination nav bar
	navY := r.Min.Y + 20
	btnW, btnH := 50, 20
	ip.prevButton = image.Rect(r.Min.X+ipPadding, navY, r.Min.X+ipPadding+btnW, navY+btnH)
	ip.nextButton = image.Rect(r.Max.X-ipPadding-btnW, navY, r.Max.X-ipPadding, navY+btnH)

	prevCol := color.RGBA{60, 60, 80, 200}
	if ip.currentPage == 0 {
		prevCol = color.RGBA{30, 30, 40, 150}
	}
	nextCol := color.RGBA{60, 60, 80, 200}
	if ip.currentPage >= ip.maxPage() {
		nextCol = color.RGBA{30, 30, 40, 150}
	}

	pb, nb := ip.prevButton, ip.nextButton
	vector.DrawFilledRect(screen, float32(pb.Min.X), float32(pb.Min.Y), float32(pb.Dx()), float32(pb.Dy()), prevCol, false)
	vector.DrawFilledRect(screen, float32(nb.Min.X), float32(nb.Min.Y), float32(nb.Dx()), float32(nb.Dy()), nextCol, false)
	ebitenutil.DebugPrintAt(screen, "< Prev", pb.Min.X+4, pb.Min.Y+5)
	ebitenutil.DebugPrintAt(screen, "Next >", nb.Min.X+4, nb.Min.Y+5)

	pageStr := fmt.Sprintf("Page %d/%d", ip.currentPage+1, ip.maxPage()+1)
	pageX := (pb.Max.X+nb.Min.X)/2 - len(pageStr)*3
	ebitenutil.DebugPrintAt(screen, pageStr, pageX, navY+5)

	// Grid — clipped to the grid region to prevent image spillage
	gridX := r.Min.X + ipPadding
	gridY := navY + btnH + ipPadding
	cellW := ipSpriteSize + ipPadding
	cellH := ipSpriteSize + ipNameHeight + ipPadding

	clipRect := image.Rect(r.Min.X, gridY, r.Max.X, r.Max.Y-12)
	clipTarget := screen.SubImage(clipRect).(*ebiten.Image)

	start := ip.currentPage * ip.itemsPerPage
	end := start + ip.itemsPerPage
	if end > len(ip.itemIDs) {
		end = len(ip.itemIDs)
	}

	for i := start; i < end; i++ {
		indexOnPage := i - start
		col := indexOnPage % ipColumns
		row := indexOnPage / ipColumns
		cx := gridX + col*cellW
		cy := gridY + row*cellH

		id := ip.itemIDs[i]
		itm := items.Registry[id]
		if itm.Icon != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(cx), float64(cy))
			clipTarget.DrawImage(itm.Icon, op)
		}

		// Truncate name to fit within the cell width
		name := itm.Name
		maxChars := ipSpriteSize / 6
		if len(name) > maxChars {
			name = name[:maxChars-1] + "."
		}
		ebitenutil.DebugPrintAt(screen, name, cx, cy+ipSpriteSize+1)
	}

	// Instructions
	ebitenutil.DebugPrintAt(screen, "Click to spawn | Outside to close", r.Min.X+ipPadding, r.Max.Y-11)
}
