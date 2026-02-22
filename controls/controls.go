package controls

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// ActionID defines available game actions
type ActionID string

const (
	// Movement
	ActionMoveLeft  ActionID = "move_left"
	ActionMoveRight ActionID = "move_right"
	ActionMoveUp    ActionID = "move_up"
	ActionMoveDown  ActionID = "move_down"

	// UI
	ActionMenuUp      ActionID = "menu_up"
	ActionMenuDown    ActionID = "menu_down"
	ActionMenuConfirm ActionID = "menu_confirm"
	ActionMenuCancel  ActionID = "menu_cancel"

	// Combat
	ActionSpell1 ActionID = "spell_1"
	ActionSpell2 ActionID = "spell_2"
	ActionSpell3 ActionID = "spell_3"
	ActionSpell4 ActionID = "spell_4"
	ActionSpell5 ActionID = "spell_5"
	ActionSpell6 ActionID = "spell_6"

	// Game
	ActionInventory     ActionID = "inventory"
	ActionHeroPanel     ActionID = "hero_panel"
	ActionTogglePause   ActionID = "toggle_pause"
	ActionShowHUD       ActionID = "show_hud"
	ActionToggleEditor  ActionID = "toggle_editor"
	ActionToggleKeybind ActionID = "toggle_keybind"

	// Level generation
	ActionGenMaze   ActionID = "gen_maze"
	ActionGenForest ActionID = "gen_forest"

	// Spells Debug
	ActionSpellDebug ActionID = "spell_debug"
	ActionShowRays   ActionID = "show_rays"
)

// KeyBinding stores primary and secondary keys for an action
type KeyBinding struct {
	Primary   ebiten.Key
	Secondary ebiten.Key
}

// Controls manages all keybindings
type Controls struct {
	bindings map[ActionID]KeyBinding
}

var defaultBindings = map[ActionID]KeyBinding{
	// Movement
	ActionMoveLeft:  {Primary: ebiten.KeyArrowLeft},
	ActionMoveRight: {Primary: ebiten.KeyArrowRight},
	ActionMoveUp:    {Primary: ebiten.KeyArrowUp},
	ActionMoveDown:  {Primary: ebiten.KeyArrowDown},

	// UI Navigation
	ActionMenuUp:      {Primary: ebiten.KeyW, Secondary: ebiten.KeyArrowUp},
	ActionMenuDown:    {Primary: ebiten.KeyS, Secondary: ebiten.KeyArrowDown},
	ActionMenuConfirm: {Primary: ebiten.KeyEnter, Secondary: ebiten.KeySpace},
	ActionMenuCancel:  {Primary: ebiten.KeyEscape},

	// Combat Spells
	ActionSpell1: {Primary: ebiten.Key1},
	ActionSpell2: {Primary: ebiten.Key2},
	ActionSpell3: {Primary: ebiten.Key3},
	ActionSpell4: {Primary: ebiten.Key4},
	ActionSpell5: {Primary: ebiten.Key5},
	ActionSpell6: {Primary: ebiten.Key6},

	// Game
	ActionInventory:     {Primary: ebiten.KeyI},
	ActionHeroPanel:     {Primary: ebiten.KeyH},
	ActionTogglePause:   {Primary: ebiten.KeyEscape},
	ActionShowHUD:       {Primary: ebiten.KeyF10},
	ActionToggleEditor:  {Primary: ebiten.KeyF3},
	ActionToggleKeybind: {Primary: ebiten.KeyF1},

	// Level generation
	ActionGenMaze:   {Primary: ebiten.KeyM},
	ActionGenForest: {Primary: ebiten.KeyN},

	// Debug
	ActionSpellDebug: {Primary: ebiten.Key8},
	ActionShowRays:   {Primary: ebiten.Key9},
}

// New creates a new Controls manager with default bindings
func New() *Controls {
	c := &Controls{
		bindings: make(map[ActionID]KeyBinding),
	}
	// Copy default bindings
	for action, binding := range defaultBindings {
		c.bindings[action] = binding
	}
	return c
}

// GetBinding returns the keybinding for an action
func (c *Controls) GetBinding(action ActionID) KeyBinding {
	if binding, ok := c.bindings[action]; ok {
		return binding
	}
	return KeyBinding{}
}

// SetBinding sets the primary key for an action
func (c *Controls) SetBinding(action ActionID, key ebiten.Key) {
	if binding, ok := c.bindings[action]; ok {
		binding.Primary = key
		c.bindings[action] = binding
	}
}

// SetSecondaryBinding sets the secondary key for an action
func (c *Controls) SetSecondaryBinding(action ActionID, key ebiten.Key) {
	if binding, ok := c.bindings[action]; ok {
		binding.Secondary = key
		c.bindings[action] = binding
	}
}

// Reset resets all bindings to defaults
func (c *Controls) Reset() {
	c.bindings = make(map[ActionID]KeyBinding)
	for action, binding := range defaultBindings {
		c.bindings[action] = binding
	}
}

// ResetBinding resets a single action binding to its default
func (c *Controls) ResetBinding(action ActionID) {
	if defaultBinding, ok := defaultBindings[action]; ok {
		c.bindings[action] = defaultBinding
	}
}

// GetAllActionIDs returns all available action IDs
func GetAllActionIDs() []ActionID {
	return []ActionID{
		ActionMoveLeft,
		ActionMoveRight,
		ActionMoveUp,
		ActionMoveDown,
		ActionMenuUp,
		ActionMenuDown,
		ActionMenuConfirm,
		ActionMenuCancel,
		ActionSpell1,
		ActionSpell2,
		ActionSpell3,
		ActionSpell4,
		ActionSpell5,
		ActionSpell6,
		ActionInventory,
		ActionHeroPanel,
		ActionTogglePause,
		ActionShowHUD,
		ActionToggleEditor,
		ActionToggleKeybind,
		ActionGenMaze,
		ActionGenForest,
		ActionSpellDebug,
		ActionShowRays,
	}
}

// GetActionLabel returns a user-friendly name for an action
func GetActionLabel(action ActionID) string {
	labels := map[ActionID]string{
		ActionMoveLeft:      "Move Left",
		ActionMoveRight:     "Move Right",
		ActionMoveUp:        "Move Up",
		ActionMoveDown:      "Move Down",
		ActionMenuUp:        "Menu Up",
		ActionMenuDown:      "Menu Down",
		ActionMenuConfirm:   "Menu Confirm",
		ActionMenuCancel:    "Menu Cancel",
		ActionSpell1:        "Spell 1 - Fireball",
		ActionSpell2:        "Spell 2 - Chaos Ray",
		ActionSpell3:        "Spell 3 - Lightning Strike",
		ActionSpell4:        "Spell 4 - Lightning Storm",
		ActionSpell5:        "Spell 5 - Fractal Bloom",
		ActionSpell6:        "Spell 6 - Fractal Canopy",
		ActionInventory:     "Open Inventory",
		ActionHeroPanel:     "Toggle Hero Panel",
		ActionTogglePause:   "Pause Game",
		ActionShowHUD:       "Toggle HUD",
		ActionToggleEditor:  "Toggle Level Editor",
		ActionToggleKeybind: "Toggle Controls Info",
		ActionGenMaze:       "Generate Maze Level",
		ActionGenForest:     "Generate Forest Level",
		ActionSpellDebug:    "Spell Debug",
		ActionShowRays:      "Show Rays",
	}
	if label, ok := labels[action]; ok {
		return label
	}
	return string(action)
}

// GetKeyName returns a user-friendly name for a key
func GetKeyName(key ebiten.Key) string {
	names := map[ebiten.Key]string{
		ebiten.KeyArrowLeft:  "←",
		ebiten.KeyArrowRight: "→",
		ebiten.KeyArrowUp:    "↑",
		ebiten.KeyArrowDown:  "↓",
		ebiten.KeyW:          "W",
		ebiten.KeyA:          "A",
		ebiten.KeyS:          "S",
		ebiten.KeyD:          "D",
		ebiten.KeyEnter:      "Enter",
		ebiten.KeySpace:      "Space",
		ebiten.KeyEscape:     "Esc",
		ebiten.KeyTab:        "Tab",
		ebiten.KeyShift:      "Shift",
		ebiten.KeyControl:    "Ctrl",
		ebiten.KeyAlt:        "Alt",
		ebiten.KeyH:          "H",
		ebiten.KeyI:          "I",
		ebiten.KeyM:          "M",
		ebiten.KeyN:          "N",
		ebiten.KeyF1:         "F1",
		ebiten.KeyF2:         "F2",
		ebiten.KeyF3:         "F3",
		ebiten.KeyF10:        "F10",
		ebiten.Key1:          "1",
		ebiten.Key2:          "2",
		ebiten.Key3:          "3",
		ebiten.Key4:          "4",
		ebiten.Key5:          "5",
		ebiten.Key6:          "6",
		ebiten.Key8:          "8",
		ebiten.Key9:          "9",
	}
	if name, ok := names[key]; ok {
		return name
	}
	return "?"
}
