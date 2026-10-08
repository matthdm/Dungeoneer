//go:build !js

package storage

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

// chdirTemp moves the test into an empty temp directory, which is where the
// desktop backend keeps saves.
func chdirTemp(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestRoundTrip(t *testing.T) {
	chdirTemp(t)

	if err := WriteFile("a.json", []byte(`{"x":1}`)); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got, err := ReadFile("a.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != `{"x":1}` {
		t.Errorf("got %q", got)
	}
}

func TestReadMissing(t *testing.T) {
	chdirTemp(t)

	_, err := ReadFile("missing.json")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("errors.Is(err, fs.ErrNotExist) = false, err = %v", err)
	}
	// Existing save code checks with os.IsNotExist.
	if !os.IsNotExist(err) {
		t.Errorf("os.IsNotExist(err) = false, err = %v", err)
	}
}

func TestWriteCreatesParentDir(t *testing.T) {
	chdirTemp(t)

	if err := WriteFile("echoes/run_1.json", []byte("echo")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got, err := ReadFile("echoes/run_1.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "echo" {
		t.Errorf("got %q", got)
	}
}

func TestOverwrite(t *testing.T) {
	chdirTemp(t)

	if err := WriteFile("a.json", []byte("first, and longer")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := WriteFile("a.json", []byte("second")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got, err := ReadFile("a.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}

func TestRemove(t *testing.T) {
	chdirTemp(t)

	if err := WriteFile("a.json", []byte("x")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := Remove("a.json"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := ReadFile("a.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("after Remove, ReadFile err = %v, want fs.ErrNotExist", err)
	}
}

func TestRemoveMissingIsNotAnError(t *testing.T) {
	chdirTemp(t)

	if err := Remove("never_existed.json"); err != nil {
		t.Errorf("Remove of a missing name = %v, want nil", err)
	}
}

func TestDirectoryInTheWayFails(t *testing.T) {
	chdirTemp(t)
	if err := os.Mkdir("controls.json", 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := WriteFile("controls.json", []byte("x")); err == nil {
		t.Error("WriteFile over a directory = nil, want error")
	}
	_, err := ReadFile("controls.json")
	if err == nil {
		t.Error("ReadFile of a directory = nil, want error")
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile of a directory reported not-exist: %v", err)
	}
}
