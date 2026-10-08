package gamedata

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// useEmbedded registers fsys for the duration of the test.
func useEmbedded(t *testing.T, fsys fs.FS) {
	t.Helper()
	prev := embedded
	t.Cleanup(func() { embedded = prev })
	Register(fsys)
}

// chdirTemp moves the test into an empty temp directory and returns it.
func chdirTemp(t *testing.T) string {
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
	return dir
}

func writeDisk(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestReadFile_EmbeddedFallback(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{"data/lore.json": {Data: []byte("embedded")}})

	got, err := ReadFile("data/lore.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "embedded" {
		t.Errorf("got %q, want %q", got, "embedded")
	}
}

func TestReadFile_DiskWins(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{"data/lore.json": {Data: []byte("embedded")}})
	writeDisk(t, filepath.Join("data", "lore.json"), "disk")

	got, err := ReadFile("data/lore.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "disk" {
		t.Errorf("got %q, want %q", got, "disk")
	}
}

func TestReadFile_BackslashPath(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{"data/lore.json": {Data: []byte("embedded")}})

	got, err := ReadFile(`data\lore.json`)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "embedded" {
		t.Errorf("got %q, want %q", got, "embedded")
	}
}

func TestReadFile_MissingEverywhere(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{})

	_, err := ReadFile("data/nope.json")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestReadFile_NothingRegistered(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, nil)

	_, err := ReadFile("data/nope.json")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestReadDir_EmbeddedFallback(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{
		"dialogues/a.json": {Data: []byte("{}")},
		"dialogues/b.json": {Data: []byte("{}")},
	})

	entries, err := ReadDir("dialogues")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 || entries[0].Name() != "a.json" || entries[1].Name() != "b.json" {
		t.Errorf("entries = %v, want a.json and b.json", entries)
	}
}

func TestReadDir_DiskWins(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{
		"dialogues/a.json": {Data: []byte("{}")},
		"dialogues/b.json": {Data: []byte("{}")},
	})
	writeDisk(t, filepath.Join("dialogues", "only_on_disk.json"), "{}")

	entries, err := ReadDir("dialogues")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "only_on_disk.json" {
		t.Errorf("entries = %v, want only_on_disk.json", entries)
	}
}

func TestReadDir_MissingEverywhere(t *testing.T) {
	chdirTemp(t)
	useEmbedded(t, fstest.MapFS{})

	_, err := ReadDir("dialogues")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}
