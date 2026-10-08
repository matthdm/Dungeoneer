// Package storage persists the player's saves (meta progression, run save,
// options, key bindings, echoes).
//
// Saves are addressed by a slash-separated relative name such as "meta.json"
// or "echoes/run_3.json". On desktop a name is a file path relative to the
// working directory; in a browser it is a localStorage key. Reading a name
// that was never written returns an error for which
// errors.Is(err, fs.ErrNotExist) is true on both.
package storage
