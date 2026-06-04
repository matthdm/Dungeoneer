package images

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmbeddedImage(t *testing.T) {
	// Normal_png is embedded, we can test loading it
	img, err := LoadEmbeddedImage(Normal_png)
	if err != nil {
		t.Fatalf("LoadEmbeddedImage failed: %v", err)
	}
	if img == nil {
		t.Fatalf("expected non-nil image")
	}

	// Test with invalid data
	_, err = LoadEmbeddedImage([]byte("invalid png data"))
	if err == nil {
		t.Fatalf("expected error for invalid image data")
	}
}

func TestLoadImage(t *testing.T) {
	// We need to write a temporary PNG file to test this
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.png")
	
	err := os.WriteFile(tempFile, Normal_png, 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	
	img, err := LoadImage(tempFile)
	if err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}
	if img == nil {
		t.Fatalf("expected non-nil image")
	}
	
	// Test with non-existent file
	_, err = LoadImage(filepath.Join(tempDir, "non_existent.png"))
	if err == nil {
		t.Fatalf("expected error for non-existent file")
	}
	
	// Test with invalid image file
	invalidFile := filepath.Join(tempDir, "invalid.png")
	err = os.WriteFile(invalidFile, []byte("invalid data"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}
	
	_, err = LoadImage(invalidFile)
	if err == nil {
		t.Fatalf("expected error for invalid image file")
	}
}

func TestSetDefaultWindowIcon(t *testing.T) {
	// This function sets the ebiten window icon, we just want to ensure it doesn't panic
	// Note: We're running in tests, so an ebiten window may not actually exist.
	// We just want coverage for the code path.
	SetDefaultWindowIcon()
}
