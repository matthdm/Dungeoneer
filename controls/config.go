package controls

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

const configFileName = "controls.json"

// KeyName is a string representation of ebiten.Key for JSON serialization
type KeyName string

// keyToName maps ebiten.Key to string names for persistence
var keyToName = map[ebiten.Key]string{
	ebiten.KeyArrowLeft:    "ArrowLeft",
	ebiten.KeyArrowRight:   "ArrowRight",
	ebiten.KeyArrowUp:      "ArrowUp",
	ebiten.KeyArrowDown:    "ArrowDown",
	ebiten.KeyW:            "W",
	ebiten.KeyA:            "A",
	ebiten.KeyS:            "S",
	ebiten.KeyD:            "D",
	ebiten.KeyEnter:        "Enter",
	ebiten.KeySpace:        "Space",
	ebiten.KeyEscape:       "Escape",
	ebiten.KeyTab:          "Tab",
	ebiten.KeyShift:        "Shift",
	ebiten.KeyControl:      "Control",
	ebiten.KeyAlt:          "Alt",
	ebiten.KeyH:            "H",
	ebiten.KeyI:            "I",
	ebiten.KeyM:            "M",
	ebiten.KeyN:            "N",
	ebiten.KeyR:            "R",
	ebiten.KeyF1:           "F1",
	ebiten.KeyF2:           "F2",
	ebiten.KeyF3:           "F3",
	ebiten.KeyF5:           "F5",
	ebiten.KeyF6:           "F6",
	ebiten.KeyF7:           "F7",
	ebiten.KeyF8:           "F8",
	ebiten.KeyF10:          "F10",
	ebiten.KeyF12:          "F12",
	ebiten.Key0:            "0",
	ebiten.Key1:            "1",
	ebiten.Key2:            "2",
	ebiten.Key3:            "3",
	ebiten.Key4:            "4",
	ebiten.Key5:            "5",
	ebiten.Key6:            "6",
	ebiten.Key7:            "7",
	ebiten.Key8:            "8",
	ebiten.Key9:            "9",
	ebiten.KeyComma:        "Comma",
	ebiten.KeyPeriod:       "Period",
	ebiten.KeyBracketLeft:  "BracketLeft",
	ebiten.KeyBracketRight: "BracketRight",
	ebiten.KeyPageUp:       "PageUp",
	ebiten.KeyPageDown:     "PageDown",
}

// nameToKey maps string names back to ebiten.Key
var nameToKey map[string]ebiten.Key

func init() {
	nameToKey = make(map[string]ebiten.Key)
	for key, name := range keyToName {
		nameToKey[name] = key
	}
	// Auto-populate any keys not in the explicit list using ebiten's String().
	// This ensures every pressable key (e.g. KeyF, KeyT, KeyMeta) can round-trip
	// through save/load without corrupting to KeyMax.
	for key := ebiten.Key(0); key < ebiten.KeyMax; key++ {
		if _, ok := keyToName[key]; !ok {
			s := key.String()
			if s != "" {
				keyToName[key] = s
				nameToKey[s] = key
			}
		}
	}
}

// SavedBinding represents a binding in JSON format
type SavedBinding struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// SavedConfig represents the entire configuration file
type SavedConfig struct {
	Version  int                     `json:"version"`
	Bindings map[string]SavedBinding `json:"bindings"`
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	// Save to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(cwd, configFileName), nil
}

// SaveBindings saves the current control bindings to a config file
func (c *Controls) SaveBindings() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Convert bindings to saveable format
	saved := SavedConfig{
		Version:  1,
		Bindings: make(map[string]SavedBinding),
	}

	for actionID, binding := range c.bindings {
		primaryName := keyToNameString(binding.Primary)
		secondaryName := keyToNameString(binding.Secondary)
		saved.Bindings[string(actionID)] = SavedBinding{
			Primary:   primaryName,
			Secondary: secondaryName,
		}
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bindings: %w", err)
	}

	// Write to file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadBindings loads control bindings from a config file
func (c *Controls) LoadBindings() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Check if file exists
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		// Config doesn't exist yet - use defaults
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON
	var saved SavedConfig
	err = json.Unmarshal(data, &saved)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bindings: %w", err)
	}

	// Apply loaded bindings
	for actionStr, savedBinding := range saved.Bindings {
		actionID := ActionID(actionStr)

		// Only load if action is valid
		if _, exists := c.bindings[actionID]; !exists {
			continue
		}

		// Convert key names back to ebiten.Key
		primaryKey := stringToKeyName(savedBinding.Primary)
		secondaryKey := stringToKeyName(savedBinding.Secondary)

		binding := c.bindings[actionID]
		binding.Primary = primaryKey
		binding.Secondary = secondaryKey
		c.bindings[actionID] = binding
	}

	return nil
}

// keyToNameString converts an ebiten.Key to its string name
func keyToNameString(key ebiten.Key) string {
	if name, ok := keyToName[key]; ok {
		return name
	}
	return ""
}

// stringToKeyName converts a string name back to ebiten.Key
func stringToKeyName(name string) ebiten.Key {
	if name == "" {
		return ebiten.KeyMax // Invalid key
	}
	if key, ok := nameToKey[name]; ok {
		return key
	}
	return ebiten.KeyMax // Invalid key
}
