// Package gamedata reads the game's read-only data files (dialogues, lore,
// flavor text, the hub level, dev settings).
//
// A file on disk wins, so JSON can be edited during development without a
// rebuild. When it is not on disk the copy embedded in the binary is used,
// which is what lets a lone executable or a browser build find its data.
package gamedata

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

var embedded fs.FS

// Register installs the embedded copy of the game's data files.
func Register(fsys fs.FS) {
	embedded = fsys
}

// ReadFile returns the file at p. p may use either slash direction.
func ReadFile(p string) ([]byte, error) {
	data, err := os.ReadFile(p)
	if err == nil || embedded == nil {
		return data, err
	}
	return fs.ReadFile(embedded, embeddedName(p))
}

// ReadDir lists dir from disk if it exists there, otherwise from the
// embedded copy.
func ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err == nil || embedded == nil {
		return entries, err
	}
	return fs.ReadDir(embedded, embeddedName(dir))
}

// embeddedName converts an OS-style relative path to the slash-separated,
// cleaned form fs.FS requires.
func embeddedName(p string) string {
	return path.Clean(strings.ReplaceAll(p, `\`, "/"))
}
