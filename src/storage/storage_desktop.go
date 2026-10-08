//go:build !js

package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// ReadFile returns the save called name.
func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.FromSlash(name))
}

// WriteFile stores data under name, creating parent directories as needed.
func WriteFile(name string, data []byte) error {
	path := filepath.FromSlash(name)
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}

// Remove deletes the save called name. Removing a name that does not exist
// is not an error.
func Remove(name string) error {
	err := os.Remove(filepath.FromSlash(name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
