package game

import (
	"encoding/json"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const optionsSavePath = "options.json"

// OptionsData holds user-configurable display and audio settings.
type OptionsData struct {
	Fullscreen   bool    `json:"fullscreen"`
	MasterVolume float64 `json:"master_volume"` // 0.0–1.0; wired to audio in Phase 9C
}

// DefaultOptions returns sensible defaults: windowed, volume at 80%.
func DefaultOptions() *OptionsData {
	return &OptionsData{
		Fullscreen:   false,
		MasterVolume: 0.8,
	}
}

// LoadOptions reads options.json and returns the stored settings.
// Returns defaults on missing or corrupt file — never returns an error to callers.
func LoadOptions() *OptionsData {
	data, err := os.ReadFile(optionsSavePath)
	if err != nil {
		return DefaultOptions()
	}
	var o OptionsData
	if err := json.Unmarshal(data, &o); err != nil {
		return DefaultOptions()
	}
	// Clamp volume to valid range.
	if o.MasterVolume < 0 {
		o.MasterVolume = 0
	}
	if o.MasterVolume > 1 {
		o.MasterVolume = 1
	}
	return &o
}

// Save writes the current options to options.json.
func (o *OptionsData) Save() {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(optionsSavePath, data, 0644)
}

// Apply pushes options values into the Ebiten runtime.
func (o *OptionsData) Apply() {
	ebiten.SetFullscreen(o.Fullscreen)
}
