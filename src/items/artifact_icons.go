package items

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// This file draws original, hand-composed glyph icons for the CR1 artifacts —
// one small silhouette per item that reads its shape from the item's
// mechanical identity (a gauntlet for a slam skill, a scythe for an execute,
// a chain for a root), replacing the flat domain-colored squares. Everything
// here is procedural vector art; nothing is traced or derived from any
// existing game's assets.

const iconSize = 32

// bespokeArtifactIcon returns a hand-composed icon for known artifact IDs, or
// nil if the ID has no bespoke icon yet (caller falls back to the domain
// square).
func bespokeArtifactIcon(id string) *ebiten.Image {
	gen, ok := iconGenerators[id]
	if !ok {
		return nil
	}
	img := ebiten.NewImage(iconSize, iconSize)
	drawIconFrame(img, gen.domain)
	gen.draw(img)
	return img
}

type iconGen struct {
	domain string
	draw   func(img *ebiten.Image)
}

var iconGenerators = map[string]iconGen{
	"ironbreaker_gauntlets": {"iron", drawGauntletIcon},
	"shroud_cloak":          {"shadow", drawHoodIcon},
	"wardens_medallion":     {"nature", drawShieldIcon},
	"ashbound_chain":        {"void", drawChainIcon},
	"grave_reaper":          {"void", drawScytheIcon},
	"ember_mantle":          {"flame", drawFlameIcon},
	"stone_skin_idol":       {"nature", drawTabletIcon},
	"shadows_return":        {"shadow", drawCrescentIcon},
	"void_mirror_pendant":   {"void", drawMirrorIcon},
	"arcane_surge":          {"arcane", drawStarburstIcon},
	"arcane_tempo_ring":     {"arcane", drawClockRingIcon},
	"blood_price":           {"void", drawDaggerDropIcon},
	"soul_harvest":          {"void", drawWispIcon},
	"voidweave_wraps":       {"void", drawWrapBandsIcon},
	"arcane_tempo_belt":     {"arcane", drawBeltIcon},
	"blood_vow_amulet":      {"void", drawGemAmuletIcon},
	"resonance_crystal":     {"arcane", drawCrystalIcon},
	"lifedrinker_robe":      {"nature", drawCollarIcon},
	"iron_will_band":        {"iron", drawBandIcon},
	"marrow_ring":           {"void", drawBoneRingIcon},
	"hollow_sigil":          {"void", drawSigilIcon},
	"thornweave_vest":       {"nature", drawThornVestIcon},
	"quicksilver_talisman":  {"iron", drawTalismanIcon},
	"void_rift_catalyst":    {"void", drawRiftIcon},
	"voidbound_pendant":     {"void", drawPendantIcon},
}

// ─── shared helpers ─────────────────────────────────────────────────────────

func domainAccent(domain string) color.NRGBA {
	switch domain {
	case "iron":
		return color.NRGBA{200, 215, 230, 255}
	case "shadow":
		return color.NRGBA{170, 110, 230, 255}
	case "flame":
		return color.NRGBA{255, 140, 60, 255}
	case "void":
		return color.NRGBA{190, 130, 255, 255}
	case "nature":
		return color.NRGBA{120, 220, 140, 255}
	case "arcane":
		return color.NRGBA{130, 180, 255, 255}
	default:
		return color.NRGBA{220, 220, 220, 255}
	}
}

// drawIconFrame paints the shared background: a dark domain-tinted square
// with a bright accent border, so every icon reads as part of one family.
func drawIconFrame(img *ebiten.Image, domain string) {
	accent := domainAccent(domain)
	bg := color.NRGBA{
		R: uint8(float64(accent.R) * 0.16),
		G: uint8(float64(accent.G) * 0.16),
		B: uint8(float64(accent.B) * 0.16),
		A: 255,
	}
	img.Fill(bg)
	vector.StrokeRect(img, 0.5, 0.5, iconSize-1, iconSize-1, 1, accent, false)
}

const c = iconSize / 2 // 16: shared center for all icon glyphs

// ─── Wave 1: melee artifacts ────────────────────────────────────────────────

// drawGauntletIcon: a clenched fist with knuckle ridges and radiating crack
// lines — ironbreaker_gauntlets' ground-slam shockwave, at rest.
func drawGauntletIcon(img *ebiten.Image) {
	steel := color.NRGBA{210, 220, 235, 255}
	shade := color.NRGBA{130, 140, 155, 255}
	// Fist body: rounded block.
	vector.DrawFilledRect(img, 9, 12, 14, 12, steel, true)
	vector.StrokeRect(img, 9, 12, 14, 12, 1, shade, true)
	// Knuckle ridges.
	for i := 0; i < 3; i++ {
		x := float32(11 + i*4)
		vector.StrokeLine(img, x, 12, x, 16, 1.5, shade, true)
	}
	// Thumb.
	vector.DrawFilledRect(img, 6, 17, 5, 6, steel, true)
	// Crack lines radiating from beneath the fist (impact).
	crack := color.NRGBA{160, 130, 100, 220}
	vector.StrokeLine(img, float32(c)-6, 25, float32(c)-10, 30, 1, crack, true)
	vector.StrokeLine(img, float32(c), 25, float32(c), 30, 1, crack, true)
	vector.StrokeLine(img, float32(c)+6, 25, float32(c)+10, 30, 1, crack, true)
}

// drawHoodIcon: a hooded silhouette with two glowing eyes — shroud_cloak.
func drawHoodIcon(img *ebiten.Image) {
	dark := color.NRGBA{35, 20, 55, 255}
	rim := color.NRGBA{140, 90, 200, 255}
	// Hood: a triangle-ish cloak silhouette.
	path := [][2]float32{{16, 6}, {24, 26}, {8, 26}}
	drawFilledTriangle(img, path[0], path[1], path[2], dark)
	vector.StrokeLine(img, 16, 6, 24, 26, 1, rim, true)
	vector.StrokeLine(img, 16, 6, 8, 26, 1, rim, true)
	// Glowing eyes.
	eyeCol := color.NRGBA{200, 160, 255, 255}
	vector.DrawFilledCircle(img, 13, 17, 1.6, eyeCol, true)
	vector.DrawFilledCircle(img, 19, 17, 1.6, eyeCol, true)
}

// drawShieldIcon: a heraldic shield with a cross emblem — wardens_medallion.
func drawShieldIcon(img *ebiten.Image) {
	gold := color.NRGBA{230, 195, 110, 255}
	dark := color.NRGBA{80, 65, 30, 255}
	pts := [][2]float32{{16, 6}, {24, 9}, {24, 18}, {16, 27}, {8, 18}, {8, 9}}
	for i := 1; i < len(pts)-1; i++ {
		drawFilledTriangle(img, pts[0], pts[i], pts[i+1], gold)
	}
	for i := 0; i < len(pts); i++ {
		p1 := pts[i]
		p2 := pts[(i+1)%len(pts)]
		vector.StrokeLine(img, p1[0], p1[1], p2[0], p2[1], 1, dark, true)
	}
	vector.StrokeLine(img, 16, 11, 16, 22, 2, dark, true)
	vector.StrokeLine(img, 11, 16, 21, 16, 2, dark, true)
}

// drawChainIcon: two interlocked chain links — ashbound_chain's root bind.
func drawChainIcon(img *ebiten.Image) {
	rust := color.NRGBA{170, 110, 60, 255}
	vector.StrokeCircle(img, 12, 14, 5, 2.5, rust, true)
	vector.StrokeCircle(img, 20, 19, 5, 2.5, rust, true)
}

// drawScytheIcon: a curved blade on a long handle — grave_reaper's execute.
func drawScytheIcon(img *ebiten.Image) {
	handle := color.NRGBA{90, 70, 60, 255}
	blade := color.NRGBA{190, 150, 230, 255}
	vector.StrokeLine(img, 11, 27, 20, 6, 2, handle, true)
	// Blade: an arc approximated by short segments.
	const segs = 10
	for i := 0; i < segs; i++ {
		t1 := float64(i) / segs
		t2 := float64(i+1) / segs
		a1 := math.Pi*0.15 + t1*math.Pi*0.9
		a2 := math.Pi*0.15 + t2*math.Pi*0.9
		r := float32(9)
		ox, oy := float32(20), float32(9)
		x1 := ox + r*float32(math.Cos(a1))
		y1 := oy + r*float32(math.Sin(a1))
		x2 := ox + r*float32(math.Cos(a2))
		y2 := oy + r*float32(math.Sin(a2))
		vector.StrokeLine(img, x1, y1, x2, y2, 2.2, blade, true)
	}
}

// drawFlameIcon: a teardrop flame — ember_mantle's burn.
func drawFlameIcon(img *ebiten.Image) {
	outer := color.NRGBA{255, 130, 40, 255}
	inner := color.NRGBA{255, 220, 90, 255}
	drawFlameShape(img, 16, 24, 8, outer)
	drawFlameShape(img, 16, 23, 4.2, inner)
}

func drawFlameShape(img *ebiten.Image, baseX, baseY float32, size float32, col color.NRGBA) {
	const segs = 12
	for i := 0; i < segs; i++ {
		t1 := float64(i) / segs
		t2 := float64(i+1) / segs
		y1 := baseY - float32(t1)*size*2.2
		y2 := baseY - float32(t2)*size*2.2
		widthAt := func(t float64) float32 {
			// Wide in the middle, tapering to a point at top and base.
			return size * float32(math.Sin(t*math.Pi)) * (1 - float32(t)*0.3)
		}
		w1, w2 := widthAt(t1), widthAt(t2)
		vector.StrokeLine(img, baseX-w1, y1, baseX-w2, y2, 2, col, true)
		vector.StrokeLine(img, baseX+w1, y1, baseX+w2, y2, 2, col, true)
	}
}

// ─── Wave 2: meta-build artifacts ───────────────────────────────────────────

// drawTabletIcon: a stone tablet with slit eyes — stone_skin_idol.
func drawTabletIcon(img *ebiten.Image) {
	stone := color.NRGBA{150, 165, 150, 255}
	dark := color.NRGBA{70, 80, 65, 255}
	vector.DrawFilledRect(img, 9, 7, 14, 19, stone, true)
	vector.StrokeRect(img, 9, 7, 14, 19, 1.5, dark, true)
	vector.StrokeLine(img, 12, 14, 15, 14, 2, dark, true)
	vector.StrokeLine(img, 17, 14, 20, 14, 2, dark, true)
	vector.StrokeLine(img, 12, 20, 20, 20, 1.5, dark, true)
}

// drawCrescentIcon: a crescent moon with a trailing swirl — shadows_return.
func drawCrescentIcon(img *ebiten.Image) {
	violet := color.NRGBA{170, 120, 230, 255}
	vector.StrokeCircle(img, 15, 16, 8, 2.5, violet, true)
	// Mask the near side to fake a crescent by drawing a background-colored
	// circle offset over part of it.
	bg := color.NRGBA{
		R: uint8(float64(violet.R) * 0.16), G: uint8(float64(violet.G) * 0.16), B: uint8(float64(violet.B) * 0.16), A: 255,
	}
	vector.DrawFilledCircle(img, 19, 14, 7, bg, true)
}

// drawMirrorIcon: a diamond mirror frame with a watching eye — void_mirror_pendant.
func drawMirrorIcon(img *ebiten.Image) {
	frame := color.NRGBA{150, 100, 220, 255}
	pts := [][2]float32{{16, 5}, {26, 16}, {16, 27}, {6, 16}}
	for i := 0; i < len(pts); i++ {
		p1 := pts[i]
		p2 := pts[(i+1)%len(pts)]
		vector.StrokeLine(img, p1[0], p1[1], p2[0], p2[1], 1.5, frame, true)
	}
	vector.StrokeCircle(img, 16, 16, 4, 1.5, color.NRGBA{210, 180, 255, 255}, true)
	vector.DrawFilledCircle(img, 16, 16, 1.6, color.NRGBA{230, 210, 255, 255}, true)
}

// drawStarburstIcon: bolts converging into a bright center — arcane_surge.
func drawStarburstIcon(img *ebiten.Image) {
	bolt := color.NRGBA{150, 190, 255, 255}
	for i := 0; i < 8; i++ {
		ang := float64(i) / 8 * 2 * math.Pi
		x1 := float32(c) + float32(math.Cos(ang))*10
		y1 := float32(c) + float32(math.Sin(ang))*10
		x2 := float32(c) + float32(math.Cos(ang))*4
		y2 := float32(c) + float32(math.Sin(ang))*4
		vector.StrokeLine(img, x1, y1, x2, y2, 1.5, bolt, true)
	}
	vector.DrawFilledCircle(img, c, c, 3.5, color.NRGBA{225, 235, 255, 255}, true)
}

// drawClockRingIcon: a ring with tick marks — arcane_tempo_ring's CDR.
func drawClockRingIcon(img *ebiten.Image) {
	ring := color.NRGBA{140, 180, 255, 255}
	vector.StrokeCircle(img, c, c, 9, 2, ring, true)
	for i := 0; i < 4; i++ {
		ang := float64(i) / 4 * 2 * math.Pi
		x1 := float32(c) + float32(math.Cos(ang))*6
		y1 := float32(c) + float32(math.Sin(ang))*6
		x2 := float32(c) + float32(math.Cos(ang))*9
		y2 := float32(c) + float32(math.Sin(ang))*9
		vector.StrokeLine(img, x1, y1, x2, y2, 1.5, ring, true)
	}
}

// drawDaggerDropIcon: a blood droplet pierced by a dagger — blood_price.
func drawDaggerDropIcon(img *ebiten.Image) {
	blood := color.NRGBA{200, 30, 50, 255}
	drawFlameShape(img, 16, 25, 6, blood) // droplet reuses the flame taper shape
	dagger := color.NRGBA{200, 200, 210, 255}
	vector.StrokeLine(img, 10, 22, 22, 10, 2, dagger, true)
	vector.StrokeLine(img, 10, 22, 13, 22, 2, dagger, true)
	vector.StrokeLine(img, 10, 22, 10, 19, 2, dagger, true)
}

// drawWispIcon: a soft looping spirit trail — soul_harvest.
func drawWispIcon(img *ebiten.Image) {
	wisp := color.NRGBA{210, 180, 255, 230}
	const segs = 20
	for i := 0; i < segs; i++ {
		t := float64(i) / segs * 2 * math.Pi * 1.6
		r := float32(2 + 5*t/(2*math.Pi*1.6))
		x := float32(c) + r*float32(math.Cos(t))
		y := float32(c+2) - r*float32(math.Sin(t))*0.7
		vector.DrawFilledCircle(img, x, y, 1.3, wisp, true)
	}
}

// ─── Stat-only build-enabler items ─────────────────────────────────────────

// drawWrapBandsIcon: diagonal wrap bandages — voidweave_wraps.
func drawWrapBandsIcon(img *ebiten.Image) {
	band := color.NRGBA{130, 90, 180, 255}
	for i := -1; i < 3; i++ {
		x := float32(6 + i*7)
		vector.StrokeLine(img, x, 26, x+10, 6, 2.5, band, true)
	}
}

// drawBeltIcon: a belt strap with a buckle — arcane_tempo_belt.
func drawBeltIcon(img *ebiten.Image) {
	strap := color.NRGBA{120, 170, 255, 255}
	vector.DrawFilledRect(img, 6, 14, 20, 5, strap, true)
	vector.StrokeCircle(img, 16, 16.5, 4, 1.5, color.NRGBA{230, 240, 255, 255}, true)
}

// drawGemAmuletIcon: a diamond gem on a chain — blood_vow_amulet.
func drawGemAmuletIcon(img *ebiten.Image) {
	chain := color.NRGBA{150, 100, 220, 200}
	vector.StrokeLine(img, 10, 6, 16, 12, 1, chain, true)
	vector.StrokeLine(img, 22, 6, 16, 12, 1, chain, true)
	gem := color.NRGBA{210, 30, 60, 255}
	pts := [][2]float32{{16, 12}, {21, 18}, {16, 26}, {11, 18}}
	for i := 1; i < len(pts)-1; i++ {
		drawFilledTriangle(img, pts[0], pts[i], pts[i+1], gem)
	}
}

// drawCrystalIcon: a faceted crystal cluster — resonance_crystal.
func drawCrystalIcon(img *ebiten.Image) {
	crystal := color.NRGBA{140, 190, 255, 255}
	drawFilledTriangle(img, [2]float32{16, 6}, [2]float32{22, 26}, [2]float32{10, 26}, crystal)
	drawFilledTriangle(img, [2]float32{9, 12}, [2]float32{13, 26}, [2]float32{5, 24}, color.NRGBA{170, 210, 255, 200})
	dark := color.NRGBA{70, 110, 170, 255}
	vector.StrokeLine(img, 16, 6, 16, 26, 1, dark, true)
}

// drawCollarIcon: a robe collar/vestment shape — lifedrinker_robe.
func drawCollarIcon(img *ebiten.Image) {
	robe := color.NRGBA{110, 200, 130, 255}
	pts := [][2]float32{{10, 8}, {22, 8}, {26, 26}, {6, 26}}
	for i := 1; i < len(pts)-1; i++ {
		drawFilledTriangle(img, pts[0], pts[i], pts[i+1], robe)
	}
	dark := color.NRGBA{50, 110, 65, 255}
	vector.StrokeLine(img, 13, 8, 16, 13, 1.5, dark, true)
	vector.StrokeLine(img, 19, 8, 16, 13, 1.5, dark, true)
}

// drawBandIcon: a plain reinforced band ring — iron_will_band.
func drawBandIcon(img *ebiten.Image) {
	steel := color.NRGBA{210, 220, 235, 255}
	vector.StrokeCircle(img, c, c, 8, 3, steel, true)
}

// drawBoneRingIcon: a bone-pale ring with cross-hatch ticks — marrow_ring.
func drawBoneRingIcon(img *ebiten.Image) {
	bone := color.NRGBA{220, 210, 190, 255}
	vector.StrokeCircle(img, c, c, 8, 2.5, bone, true)
	for i := 0; i < 6; i++ {
		ang := float64(i) / 6 * 2 * math.Pi
		x1 := float32(c) + float32(math.Cos(ang))*6
		y1 := float32(c) + float32(math.Sin(ang))*6
		x2 := float32(c) + float32(math.Cos(ang))*10
		y2 := float32(c) + float32(math.Sin(ang))*10
		vector.StrokeLine(img, x1, y1, x2, y2, 1, bone, true)
	}
}

// drawSigilIcon: a triangular rune with a hollow center — hollow_sigil.
func drawSigilIcon(img *ebiten.Image) {
	rune_ := color.NRGBA{160, 110, 220, 255}
	pts := [][2]float32{{16, 6}, {25, 24}, {7, 24}}
	for i := 0; i < 3; i++ {
		p1, p2 := pts[i], pts[(i+1)%3]
		vector.StrokeLine(img, p1[0], p1[1], p2[0], p2[1], 2, rune_, true)
	}
	vector.StrokeCircle(img, 16, 19, 3.5, 1.5, rune_, true)
}

// drawThornVestIcon: a vest silhouette ringed with thorn spikes — thornweave_vest.
func drawThornVestIcon(img *ebiten.Image) {
	vest := color.NRGBA{100, 190, 120, 255}
	vector.DrawFilledRect(img, 10, 9, 12, 17, vest, true)
	thorn := color.NRGBA{60, 130, 70, 255}
	for i := 0; i < 4; i++ {
		y := float32(10 + i*4)
		vector.StrokeLine(img, 10, y, 6, y-2, 1.5, thorn, true)
		vector.StrokeLine(img, 22, y, 26, y-2, 1.5, thorn, true)
	}
}

// drawTalismanIcon: a teardrop talisman with a lightning zigzag — quicksilver_talisman.
func drawTalismanIcon(img *ebiten.Image) {
	steel := color.NRGBA{215, 225, 240, 255}
	drawFlameShape(img, 16, 26, 7, steel)
	bolt := color.NRGBA{90, 130, 190, 255}
	pts := [][2]float32{{18, 12}, {14, 18}, {17, 18}, {13, 25}}
	for i := 0; i < len(pts)-1; i++ {
		vector.StrokeLine(img, pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1], 1.5, bolt, true)
	}
}

// drawRiftIcon: a jagged dark tear ringed by a bright event-horizon —
// void_rift_catalyst.
func drawRiftIcon(img *ebiten.Image) {
	rim := color.NRGBA{190, 130, 255, 255}
	dark := color.NRGBA{35, 15, 55, 255}
	pts := [][2]float32{{16, 6}, {21, 13}, {26, 16}, {20, 19}, {16, 27}, {12, 19}, {6, 16}, {11, 13}}
	for i := 1; i < len(pts)-1; i++ {
		drawFilledTriangle(img, pts[0], pts[i], pts[i+1], dark)
	}
	for i := 0; i < len(pts); i++ {
		p1 := pts[i]
		p2 := pts[(i+1)%len(pts)]
		vector.StrokeLine(img, p1[0], p1[1], p2[0], p2[1], 1.2, rim, true)
	}
	vector.DrawFilledCircle(img, 16, 16, 2, color.NRGBA{230, 210, 255, 255}, true)
}

// drawPendantIcon: a teardrop pendant on a chain, warping the space around
// it — voidbound_pendant.
func drawPendantIcon(img *ebiten.Image) {
	chain := color.NRGBA{150, 100, 220, 200}
	vector.StrokeLine(img, 12, 5, 16, 11, 1, chain, true)
	vector.StrokeLine(img, 20, 5, 16, 11, 1, chain, true)
	gem := color.NRGBA{130, 60, 190, 255}
	drawFlameShape(img, 16, 27, 6.5, gem)
	vector.StrokeCircle(img, 16, 18, 9, 1, color.NRGBA{190, 130, 255, 140}, true)
}

// ─── low-level primitives ──────────────────────────────────────────────────

// drawFilledTriangle fills a triangle using ebiten's vertex/triangle path,
// since the vector package has no direct filled-polygon helper.
func drawFilledTriangle(img *ebiten.Image, p1, p2, p3 [2]float32, col color.NRGBA) {
	r := float32(col.R) / 255
	g := float32(col.G) / 255
	b := float32(col.B) / 255
	a := float32(col.A) / 255
	vs := []ebiten.Vertex{
		{DstX: p1[0], DstY: p1[1], ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: p2[0], DstY: p2[1], ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: p3[0], DstY: p3[1], ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	is := []uint16{0, 1, 2}
	whitePixel := ebiten.NewImage(1, 1)
	whitePixel.Fill(color.White)
	img.DrawTriangles(vs, is, whitePixel, nil)
}
