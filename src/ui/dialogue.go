package ui

import (
	"dungeoneer/dialogue"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

var (
	dialogueSpeakerColor  = color.RGBA{220, 180, 60, 255}  // gold
	dialogueTextColor     = color.RGBA{220, 220, 220, 255} // light grey
	dialogueResponseColor = color.RGBA{200, 200, 200, 255}
	dialogueHoverColor    = color.RGBA{255, 220, 100, 255} // bright gold on hover
	dialogueBGColor       = color.RGBA{12, 12, 18, 240}
	dialogueBorderColor   = color.RGBA{150, 20, 15, 255}
	dialogueOverlayColor  = color.RGBA{0, 0, 0, 80}
)

const (
	dialoguePadX       = 20
	dialoguePadY       = 14
	dialogueLineH      = 18
	dialoguePortraitSz = 128 // 64x64 sprite drawn at 2x
	typewriterSpeed    = 1.5 // characters per tick at 60 fps
)

// DialoguePanel renders a BG1-style dialogue window at the bottom of the screen.
type DialoguePanel struct {
	Active   bool
	screenW  int
	screenH  int

	// Current conversation state
	Tree        *dialogue.DialogueTree
	CurrentNode *dialogue.DialogueNode
	Responses   []dialogue.DialogueResponse // filtered by conditions

	// Typewriter
	TextProgress float64
	TextDone     bool

	// Mouse
	HoverIndex    int
	ResponseRects []image.Rectangle

	// Portrait
	Portrait *ebiten.Image

	// Callbacks (set by game)
	OnClose       func()
	EvalCondition func(*dialogue.DialogueCondition) bool
	ExecAction    func(dialogue.DialogueAction)
}

// NewDialoguePanel creates an inactive dialogue panel sized for the given screen.
func NewDialoguePanel(w, h int) *DialoguePanel {
	return &DialoguePanel{
		screenW:    w,
		screenH:    h,
		HoverIndex: -1,
	}
}

// Open starts a dialogue with the given tree and optional portrait.
func (dp *DialoguePanel) Open(tree *dialogue.DialogueTree, portrait *ebiten.Image) {
	dp.Tree = tree
	dp.Portrait = portrait
	dp.Active = true
	dp.advanceToNode(tree.Root)
}

// Close ends the dialogue.
func (dp *DialoguePanel) Close() {
	dp.Active = false
	dp.Tree = nil
	dp.CurrentNode = nil
	dp.Responses = nil
	if dp.OnClose != nil {
		dp.OnClose()
	}
}

// Resize updates the screen dimensions (called from Layout).
func (dp *DialoguePanel) Resize(w, h int) {
	dp.screenW = w
	dp.screenH = h
}

// Update handles input for the dialogue panel.
func (dp *DialoguePanel) Update() {
	if !dp.Active || dp.CurrentNode == nil {
		return
	}

	// Typewriter advance
	if !dp.TextDone {
		dp.TextProgress += typewriterSpeed
		if int(dp.TextProgress) >= len(dp.CurrentNode.Text) {
			dp.TextDone = true
		}
	}

	mx, my := ebiten.CursorPosition()

	// Click handling
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// If typewriter still going, skip to end
		if !dp.TextDone {
			dp.TextProgress = float64(len(dp.CurrentNode.Text))
			dp.TextDone = true
			return
		}

		// If no responses, close on click
		if len(dp.Responses) == 0 {
			dp.Close()
			return
		}

		// Check response clicks
		for i, rect := range dp.ResponseRects {
			if mx >= rect.Min.X && mx <= rect.Max.X && my >= rect.Min.Y && my <= rect.Max.Y {
				dp.selectResponse(i)
				return
			}
		}
	}

	// Escape to close (only when no mandatory responses)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		dp.Close()
		return
	}

	// Hover detection
	dp.HoverIndex = -1
	for i, rect := range dp.ResponseRects {
		if mx >= rect.Min.X && mx <= rect.Max.X && my >= rect.Min.Y && my <= rect.Max.Y {
			dp.HoverIndex = i
			break
		}
	}
}

// Draw renders the dialogue panel.
func (dp *DialoguePanel) Draw(screen *ebiten.Image) {
	if !dp.Active || dp.CurrentNode == nil {
		return
	}

	sw := float32(dp.screenW)
	sh := float32(dp.screenH)

	// Semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, sw, sh, dialogueOverlayColor, false)

	// Panel dimensions: bottom-center, 80% width, 35% height
	panelW := sw * 0.80
	panelH := sh * 0.35
	panelX := (sw - panelW) / 2
	panelY := sh - panelH - 10

	// Background + border
	vector.DrawFilledRect(screen, panelX, panelY, panelW, panelH, dialogueBGColor, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 3, dialogueBorderColor, false)

	// Content area
	contentX := int(panelX) + dialoguePadX
	contentY := int(panelY) + dialoguePadY
	contentW := int(panelW) - dialoguePadX*2

	// Portrait (2x scale, top-left of panel)
	portraitEndX := contentX
	if dp.Portrait != nil {
		pOp := &ebiten.DrawImageOptions{}
		pOp.GeoM.Scale(2, 2)
		pOp.GeoM.Translate(float64(contentX), float64(contentY))
		screen.DrawImage(dp.Portrait, pOp)
		portraitEndX = contentX + dialoguePortraitSz + dialoguePadX
	}

	textX := portraitEndX
	textW := contentX + contentW - textX

	// Speaker name
	curY := contentY + 4
	if dp.CurrentNode.Speaker != "" {
		DrawBigText(screen, dp.CurrentNode.Speaker, textX, curY, 2, dialogueSpeakerColor)
		curY += 30
	}

	// Dialogue text (typewriter)
	visibleLen := len(dp.CurrentNode.Text)
	if !dp.TextDone {
		visibleLen = int(dp.TextProgress)
		if visibleLen > len(dp.CurrentNode.Text) {
			visibleLen = len(dp.CurrentNode.Text)
		}
	}
	visibleText := dp.CurrentNode.Text[:visibleLen]

	maxChars := textW / 7 // basicfont is 7px wide
	if maxChars < 20 {
		maxChars = 20
	}
	lines := WrapText(visibleText, maxChars)
	for _, line := range lines {
		text.Draw(screen, line, basicfont.Face7x13, textX, curY+13, dialogueTextColor)
		curY += dialogueLineH
	}

	// Responses (only when typewriter done)
	dp.ResponseRects = dp.ResponseRects[:0]
	if dp.TextDone && len(dp.Responses) > 0 {
		curY += 8
		for i, resp := range dp.Responses {
			prefix := "  "
			clr := dialogueResponseColor
			if i == dp.HoverIndex {
				prefix = "> "
				clr = dialogueHoverColor
			}
			display := prefix + resp.Text
			rx := textX
			ry := curY
			text.Draw(screen, display, basicfont.Face7x13, rx, ry+13, clr)

			rw := len(display) * 7
			if rw < textW {
				rw = textW // make the whole row clickable
			}
			dp.ResponseRects = append(dp.ResponseRects, image.Rect(rx, ry, rx+rw, ry+dialogueLineH))
			curY += dialogueLineH + 2
		}
	}

	// "Click to continue" hint when no responses and text done
	if dp.TextDone && len(dp.Responses) == 0 {
		curY += 8
		hint := "[Click to continue]"
		text.Draw(screen, hint, basicfont.Face7x13, textX, curY+13, color.RGBA{120, 120, 120, 200})
	}
}

// advanceToNode moves to the specified dialogue node.
func (dp *DialoguePanel) advanceToNode(nodeID string) {
	if dp.Tree == nil {
		dp.Close()
		return
	}
	node, ok := dp.Tree.Nodes[nodeID]
	if !ok {
		dp.Close()
		return
	}
	dp.CurrentNode = node
	dp.TextProgress = 0
	dp.TextDone = false
	dp.HoverIndex = -1
	dp.ResponseRects = nil

	// Fire OnEnter actions
	for _, a := range node.OnEnter {
		if dp.ExecAction != nil {
			dp.ExecAction(a)
		}
	}

	// Filter responses by conditions
	dp.Responses = nil
	for _, r := range node.Responses {
		if r.Condition == nil || (dp.EvalCondition != nil && dp.EvalCondition(r.Condition)) {
			dp.Responses = append(dp.Responses, r)
		}
	}
}

// selectResponse handles the player clicking a response.
func (dp *DialoguePanel) selectResponse(idx int) {
	if idx < 0 || idx >= len(dp.Responses) {
		return
	}
	resp := dp.Responses[idx]

	// Fire OnSelect actions
	for _, a := range resp.OnSelect {
		if dp.ExecAction != nil {
			dp.ExecAction(a)
		}
	}

	// Navigate
	if resp.NextNode == "" {
		dp.Close()
		return
	}
	dp.advanceToNode(resp.NextNode)
}
