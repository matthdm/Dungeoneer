package game

import (
	"dungeoneer/items"
	"dungeoneer/levels"
	"encoding/json"
	"image/color"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// EventEffect describes a single mechanical consequence of an event choice.
type EventEffect struct {
	Type  string `json:"type"`  // "lose_hp", "gain_remnants", "give_item", "set_flag", "curse_item"
	Value int    `json:"value"` // HP amount, remnant amount, etc.
	ID    string `json:"id"`    // item ID for give_item; flag key for set_flag
}

// EventChoice is one option the player may select in an event panel.
type EventChoice struct {
	Label   string        `json:"label"`
	Preview string        `json:"preview"`
	Effects []EventEffect `json:"effects"`
}

// EventDef is the data definition for a single event loaded from events.json.
type EventDef struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Trigger string        `json:"trigger"` // "entry" only in v1
	Text    string        `json:"text"`
	Choices []EventChoice `json:"choices"`
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

// EventDefs holds all loaded event definitions.
var EventDefs []EventDef

// LoadEventDefs reads and parses events.json into EventDefs.
func LoadEventDefs(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &EventDefs)
}

// ---------------------------------------------------------------------------
// Panel state (package-level; acceptable for single-instance game)
// ---------------------------------------------------------------------------

var eventPanelOpen bool
var eventPanelSelectedIdx int

// ---------------------------------------------------------------------------
// assignEventRoom — called by hub.go after floor generation
// ---------------------------------------------------------------------------

// assignEventRoom picks an unseen event, arms it on the event-tagged room, and
// marks it seen immediately so it will not repeat even if the player skips it.
// g.ActiveEvent is set here but the panel only opens when the player enters the
// room (checkEventRoomEntry sets eventPanelOpen).
func (g *Game) assignEventRoom() {
	g.EventRoom = nil
	g.ActiveEvent = nil
	eventPanelOpen = false

	if len(EventDefs) == 0 || g.RunState == nil || g.currentLevel == nil {
		return
	}

	// Pick the first unseen event.
	var chosen *EventDef
	for i := range EventDefs {
		seen := false
		for _, id := range g.RunState.EventsSeen {
			if id == EventDefs[i].ID {
				seen = true
				break
			}
		}
		if !seen {
			chosen = &EventDefs[i]
			break
		}
	}
	if chosen == nil {
		return
	}

	// Find the event-tagged room in the current level.
	for i := range g.currentLevel.Rooms {
		if g.currentLevel.Rooms[i].HasTag(levels.TagEvent) {
			g.EventRoom = &g.currentLevel.Rooms[i]
			defCopy := *chosen
			g.ActiveEvent = &defCopy
			// Mark as seen immediately — the event is consumed even if skipped.
			g.RunState.EventsSeen = append(g.RunState.EventsSeen, chosen.ID)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// checkEventRoomEntry — called from game.go update loop
// ---------------------------------------------------------------------------

// checkEventRoomEntry opens the event panel when the player steps into the
// event room. Safe to call every tick; exits early if already open.
func (g *Game) checkEventRoomEntry() {
	if g.EventRoom == nil || g.ActiveEvent == nil || eventPanelOpen {
		return
	}
	if g.player == nil {
		return
	}
	if g.EventRoom.Contains(g.player.TileX, g.player.TileY) {
		eventPanelOpen = true
		eventPanelSelectedIdx = 0
	}
}

// ---------------------------------------------------------------------------
// executeEventEffect — applies a single EventEffect to the game state
// ---------------------------------------------------------------------------

func (g *Game) executeEventEffect(e EventEffect) {
	switch e.Type {
	case "lose_hp":
		if g.player != nil {
			dmg := e.Value
			if dmg >= g.player.HP {
				dmg = g.player.HP - 1 // cannot kill
			}
			if dmg > 0 {
				g.player.HP -= dmg
			}
		}
	case "gain_remnants":
		if g.RunState != nil {
			g.RunState.RemnantEarned += e.Value
		}
	case "give_item":
		if e.ID != "" {
			tmpl, ok := items.Registry[e.ID]
			if ok {
				it := &items.Item{ItemTemplate: tmpl, Count: 1}
				g.AddItemToPlayer(it)
			}
		}
	case "set_flag":
		if g.RunState != nil && e.ID != "" {
			g.RunState.QuestFlags[e.ID] = 1
		}
	case "curse_item":
		g.curseRandomEquippedItem()
	}
}

// curseRandomEquippedItem downgrades the Quality of one randomly chosen
// equipped item by one tier (legendary → rare → uncommon → common).
func (g *Game) curseRandomEquippedItem() {
	if g.player == nil || len(g.player.Equipment) == 0 {
		return
	}
	downgrades := map[string]string{
		items.RarityLegendary: items.RarityRare,
		items.RarityRare:      items.RarityUncommon,
		items.RarityUncommon:  items.RarityCommon,
	}
	// Collect equipped items that have a downgradeable quality.
	var candidates []*items.ItemTemplate
	for _, it := range g.player.Equipment {
		if it == nil || it.ItemTemplate == nil {
			continue
		}
		if _, ok := downgrades[it.Quality]; ok {
			candidates = append(candidates, it.ItemTemplate)
		}
	}
	if len(candidates) == 0 {
		return
	}
	target := candidates[rand.IntN(len(candidates))]
	target.Quality = downgrades[target.Quality]
}

// ---------------------------------------------------------------------------
// closeEventPanel
// ---------------------------------------------------------------------------

func (g *Game) closeEventPanel() {
	eventPanelOpen = false
	g.ActiveEvent = nil
	g.EventRoom = nil
}

// ---------------------------------------------------------------------------
// drawEventPanel — modal overlay rendered when eventPanelOpen is true
// ---------------------------------------------------------------------------

// drawEventPanel renders the event choice panel and handles keyboard input.
// It is called from Draw (draw.game.go) only when eventPanelOpen && ActiveEvent != nil.
func (g *Game) drawEventPanel(screen *ebiten.Image) {
	// --- Input ---
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if eventPanelSelectedIdx > 0 {
			eventPanelSelectedIdx--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if eventPanelSelectedIdx < len(g.ActiveEvent.Choices)-1 {
			eventPanelSelectedIdx++
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		// Escape selects the last choice (walk away).
		eventPanelSelectedIdx = len(g.ActiveEvent.Choices) - 1
		g.commitEventChoice()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.commitEventChoice()
		return
	}

	// --- Layout constants ---
	sw, sh := g.w, g.h

	// Full-screen dark overlay.
	overlay := ebiten.NewImage(sw, sh)
	overlay.Fill(color.NRGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, &ebiten.DrawImageOptions{})

	const (
		panelW    = 520
		padX      = 24
		padY      = 20
		charW     = 7  // ebitenutil.DebugPrint character width
		charH     = 13 // ebitenutil.DebugPrint line height
		lineWrap  = 70 // chars per line for body text
		titleScale = 2.0
	)

	// Wrap body text into lines.
	bodyLines := wrapText(g.ActiveEvent.Text, lineWrap)
	// Height: title + gap + body lines + gap + choices.
	numChoices := len(g.ActiveEvent.Choices)
	panelH := padY + int(charH*titleScale) + 12 +
		len(bodyLines)*charH + 16 +
		numChoices*(charH+6) + padY

	panelX := (sw - panelW) / 2
	panelY := (sh - panelH) / 2

	// --- Panel background ---
	bg := ebiten.NewImage(panelW, panelH)
	bg.Fill(color.NRGBA{15, 10, 20, 240})
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(float64(panelX), float64(panelY))
	screen.DrawImage(bg, bgOp)

	// --- Border ---
	borderCol := color.NRGBA{180, 150, 100, 220}
	drawBorderRect(screen, panelX, panelY, panelW, panelH, borderCol)

	// --- Title ---
	titleW := len(g.ActiveEvent.Title) * charW * int(titleScale)
	titleImg := ebiten.NewImage(titleW+4, int(charH*titleScale))
	ebitenutil.DebugPrintAt(titleImg, g.ActiveEvent.Title, 0, 0)
	titleOp := &ebiten.DrawImageOptions{}
	titleOp.GeoM.Scale(titleScale, titleScale)
	titleOp.GeoM.Translate(float64(panelX+padX), float64(panelY+padY))
	titleOp.ColorScale.ScaleWithColor(color.NRGBA{240, 220, 160, 255})
	screen.DrawImage(titleImg, titleOp)

	// --- Body text ---
	curY := panelY + padY + int(charH*titleScale) + 12
	bodyCol := color.NRGBA{210, 200, 185, 255}
	for _, line := range bodyLines {
		lineImg := ebiten.NewImage(panelW-padX*2+4, charH)
		ebitenutil.DebugPrintAt(lineImg, line, 0, 0)
		lineOp := &ebiten.DrawImageOptions{}
		lineOp.GeoM.Translate(float64(panelX+padX), float64(curY))
		lineOp.ColorScale.ScaleWithColor(bodyCol)
		screen.DrawImage(lineImg, lineOp)
		curY += charH
	}

	// --- Divider ---
	curY += 8
	vector.StrokeLine(screen,
		float32(panelX+padX), float32(curY),
		float32(panelX+panelW-padX), float32(curY),
		1, color.NRGBA{120, 100, 70, 180}, false)
	curY += 8

	// --- Choices ---
	for i, ch := range g.ActiveEvent.Choices {
		selected := i == eventPanelSelectedIdx
		label := ch.Label
		if selected {
			label = "> " + label
		} else {
			label = "  " + label
		}
		// Choice row background highlight.
		if selected {
			hiBg := ebiten.NewImage(panelW-padX*2, charH+4)
			hiBg.Fill(color.NRGBA{80, 60, 30, 160})
			hiOp := &ebiten.DrawImageOptions{}
			hiOp.GeoM.Translate(float64(panelX+padX), float64(curY-2))
			screen.DrawImage(hiBg, hiOp)
		}
		choiceImg := ebiten.NewImage(panelW-padX*2+4, charH)
		ebitenutil.DebugPrintAt(choiceImg, label, 0, 0)
		choiceOp := &ebiten.DrawImageOptions{}
		choiceOp.GeoM.Translate(float64(panelX+padX), float64(curY))
		if selected {
			choiceOp.ColorScale.ScaleWithColor(color.NRGBA{255, 220, 100, 255})
		} else {
			choiceOp.ColorScale.ScaleWithColor(color.NRGBA{180, 170, 155, 255})
		}
		screen.DrawImage(choiceImg, choiceOp)

		// Preview line (smaller, indented) if present.
		if ch.Preview != "" && selected {
			prevImg := ebiten.NewImage(panelW-padX*2+4, charH)
			ebitenutil.DebugPrintAt(prevImg, "    "+ch.Preview, 0, 0)
			prevOp := &ebiten.DrawImageOptions{}
			prevOp.GeoM.Translate(float64(panelX+padX), float64(curY+charH))
			prevOp.ColorScale.ScaleWithColor(color.NRGBA{150, 140, 120, 200})
			screen.DrawImage(prevImg, prevOp)
			curY += charH + 6
		} else {
			curY += charH + 6
		}
	}

	// --- Footer hint ---
	hint := "[Enter] Confirm   [Esc] Walk away"
	hintImg := ebiten.NewImage(len(hint)*charW+4, charH)
	ebitenutil.DebugPrintAt(hintImg, hint, 0, 0)
	hintOp := &ebiten.DrawImageOptions{}
	hintOp.GeoM.Translate(float64(panelX+padX), float64(panelY+panelH-padY-charH))
	hintOp.ColorScale.ScaleWithColor(color.NRGBA{120, 110, 90, 180})
	screen.DrawImage(hintImg, hintOp)
}

// commitEventChoice executes the currently selected choice and closes the panel.
func (g *Game) commitEventChoice() {
	if g.ActiveEvent == nil || eventPanelSelectedIdx >= len(g.ActiveEvent.Choices) {
		g.closeEventPanel()
		return
	}
	chosen := g.ActiveEvent.Choices[eventPanelSelectedIdx]
	for _, eff := range chosen.Effects {
		g.executeEventEffect(eff)
	}
	g.closeEventPanel()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// wrapText wraps s into lines of at most maxChars characters, breaking on
// word boundaries.
func wrapText(s string, maxChars int) []string {
	words := strings.Fields(s)
	var lines []string
	var current strings.Builder
	for _, w := range words {
		if current.Len() == 0 {
			current.WriteString(w)
		} else if current.Len()+1+len(w) <= maxChars {
			current.WriteByte(' ')
			current.WriteString(w)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(w)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

// drawBorderRect draws a 1-pixel border rectangle around (x,y,w,h).
func drawBorderRect(screen *ebiten.Image, x, y, w, h int, col color.NRGBA) {
	// Top
	vector.StrokeLine(screen, float32(x), float32(y), float32(x+w), float32(y), 1, col, false)
	// Bottom
	vector.StrokeLine(screen, float32(x), float32(y+h), float32(x+w), float32(y+h), 1, col, false)
	// Left
	vector.StrokeLine(screen, float32(x), float32(y), float32(x), float32(y+h), 1, col, false)
	// Right
	vector.StrokeLine(screen, float32(x+w), float32(y), float32(x+w), float32(y+h), 1, col, false)
}
