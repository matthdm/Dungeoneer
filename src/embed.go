package main

import "embed"

// dataFS is the copy of the read-only game data baked into the binary. It is
// handed to gamedata, which prefers a file on disk and falls back to this.
//
//go:embed dialogues/*.json data/*.json levels/hub.json dev_settings.json
var dataFS embed.FS
