package controls

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestInitKeys(t *testing.T) {
	if len(keyToName) == 0 {
		t.Error("expected keyToName map to be populated")
	}
	if len(nameToKey) == 0 {
		t.Error("expected nameToKey map to be populated")
	}
}

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("expected no error from GetConfigPath, got %v", err)
	}
	if filepath.Base(path) != configFileName {
		t.Errorf("expected filename %q, got %q", configFileName, filepath.Base(path))
	}
}

func TestGetConfigPath_Error(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping working directory deletion test on Windows due to file locking")
	}

	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()

	// Delete tmpDir so os.Getwd() fails
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Fatalf("failed to remove tmpDir: %v", err)
	}

	_, err = GetConfigPath()
	if err == nil {
		t.Error("expected error from GetConfigPath when working directory is deleted, got nil")
	}

	// Test SaveBindings/LoadBindings handling GetConfigPath error
	c := New()
	if err := c.SaveBindings(); err == nil {
		t.Error("expected SaveBindings to fail when GetConfigPath fails, got nil")
	}
	if err := c.LoadBindings(); err == nil {
		t.Error("expected LoadBindings to fail when GetConfigPath fails, got nil")
	}
}

func TestSaveAndLoadBindings(t *testing.T) {
	// Create temporary directory and change current working directory to it
	// so GetConfigPath() returns a path in the temp dir.
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// 1. Load when file does not exist (should return nil and not error)
	c := New()
	err = c.LoadBindings()
	if err != nil {
		t.Fatalf("expected no error loading non-existent bindings, got %v", err)
	}

	// 2. Modify defaults, save bindings, then verify
	c.SetBinding(ActionMoveLeft, ebiten.KeySpace)
	c.SetSecondaryBinding(ActionMoveLeft, ebiten.KeyEnter)
	err = c.SaveBindings()
	if err != nil {
		t.Fatalf("failed to save bindings: %v", err)
	}

	// Create a new controls instance and load from the file
	c2 := New()
	// Ensure c2 is different before loading
	if c2.GetBinding(ActionMoveLeft).Primary == ebiten.KeySpace {
		t.Fatal("expected c2 primary key for MoveLeft to be default, not KeySpace")
	}
	err = c2.LoadBindings()
	if err != nil {
		t.Fatalf("failed to load bindings: %v", err)
	}
	// Verify it loaded correctly
	loaded := c2.GetBinding(ActionMoveLeft)
	if loaded.Primary != ebiten.KeySpace || loaded.Secondary != ebiten.KeyEnter {
		t.Errorf("expected loaded binding to match saved binding, got %+v", loaded)
	}

	// 3. Test loading invalid JSON content
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	err = os.WriteFile(configPath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid config file: %v", err)
	}

	err = c.LoadBindings()
	if err == nil {
		t.Fatal("expected error loading invalid json, got nil")
	}

	// 4. Test loading valid JSON containing invalid action IDs
	validJSON := `{
		"version": 1,
		"bindings": {
			"non_existent_action": {
				"primary": "Space",
				"secondary": "Enter"
			},
			"move_left": {
				"primary": "ArrowLeft",
				"secondary": ""
			}
		}
	}`
	err = os.WriteFile(configPath, []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write valid config file: %v", err)
	}

	c3 := New()
	err = c3.LoadBindings()
	if err != nil {
		t.Fatalf("failed to load bindings with invalid actions: %v", err)
	}
	// Verify that move_left was loaded but non_existent_action was skipped safely
	if c3.GetBinding(ActionMoveLeft).Primary != ebiten.KeyArrowLeft {
		t.Errorf("expected MoveLeft primary to be ArrowLeft, got %v", GetKeyName(c3.GetBinding(ActionMoveLeft).Primary))
	}
}

func TestSaveAndLoadBindings_FileErrors(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()

	// Create a directory named controls.json so os.ReadFile/os.WriteFile fail
	err = os.Mkdir(configFileName, 0755)
	if err != nil {
		t.Fatalf("failed to create directory controls.json: %v", err)
	}

	c := New()
	err = c.SaveBindings()
	if err == nil {
		t.Error("expected error saving bindings when controls.json is a directory, got nil")
	}

	err = c.LoadBindings()
	if err == nil {
		t.Error("expected error loading bindings when controls.json is a directory, got nil")
	}
}

func TestKeyToNameStringAndStringToKeyName(t *testing.T) {
	// Test keyToNameString
	if keyToNameString(ebiten.KeyArrowLeft) != "ArrowLeft" {
		t.Errorf("expected ArrowLeft, got %s", keyToNameString(ebiten.KeyArrowLeft))
	}
	if keyToNameString(ebiten.Key(9999)) != "" {
		t.Errorf("expected empty string for invalid key, got %q", keyToNameString(ebiten.Key(9999)))
	}

	// Test stringToKeyName
	if stringToKeyName("ArrowLeft") != ebiten.KeyArrowLeft {
		t.Errorf("expected KeyArrowLeft, got %v", stringToKeyName("ArrowLeft"))
	}
	if stringToKeyName("") != ebiten.KeyMax {
		t.Errorf("expected KeyMax for empty string, got %v", stringToKeyName(""))
	}
	if stringToKeyName("InvalidKeyName") != ebiten.KeyMax {
		t.Errorf("expected KeyMax for invalid key name, got %v", stringToKeyName("InvalidKeyName"))
	}
}
