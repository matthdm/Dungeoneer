//go:build js

package ui

// CanQuit reports whether the game can close itself. A browser tab cannot,
// so menus leave out their exit entry.
const CanQuit = false
