package game

import (
	"os"
	"testing"
)

func TestOptions(t *testing.T) {
	// Clean up any existing options.json
	os.Remove("options.json")
	defer os.Remove("options.json")

	// Test DefaultOptions
	opts := DefaultOptions()
	if opts.Fullscreen {
		t.Errorf("default should not be fullscreen")
	}
	if opts.MasterVolume != 0.8 {
		t.Errorf("default volume should be 0.8")
	}

	// Test LoadOptions with missing file
	loadedOpts := LoadOptions()
	if loadedOpts.MasterVolume != 0.8 {
		t.Errorf("load should return defaults when missing file")
	}

	// Test Save and LoadOptions
	opts.Fullscreen = true
	opts.MasterVolume = 0.5
	opts.Save()

	loadedOpts2 := LoadOptions()
	if !loadedOpts2.Fullscreen {
		t.Errorf("expected fullscreen=true after save/load")
	}
	if loadedOpts2.MasterVolume != 0.5 {
		t.Errorf("expected volume=0.5 after save/load")
	}

	// Test clamp volume < 0
	opts.MasterVolume = -1.0
	opts.Save()
	loadedOpts3 := LoadOptions()
	if loadedOpts3.MasterVolume != 0.0 {
		t.Errorf("expected volume to clamp to 0, got %f", loadedOpts3.MasterVolume)
	}

	// Test clamp volume > 1
	opts.MasterVolume = 2.0
	opts.Save()
	loadedOpts4 := LoadOptions()
	if loadedOpts4.MasterVolume != 1.0 {
		t.Errorf("expected volume to clamp to 1, got %f", loadedOpts4.MasterVolume)
	}

	// Test corrupted file
	os.WriteFile("options.json", []byte("invalid json"), 0644)
	loadedOpts5 := LoadOptions()
	if loadedOpts5.MasterVolume != 0.8 {
		t.Errorf("expected defaults after corrupt file")
	}
	
	// Apply (just testing it doesn't panic)
	opts.Apply()
}
