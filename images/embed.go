package images

import (
	_ "embed"
)

var (
	//go:embed spritesheet.png
	Spritesheet_png []byte

	//go:embed smoke.png
	Smoke_png []byte
)
