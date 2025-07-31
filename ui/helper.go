package ui

import (
	"image"
	"regexp"
)

var validFilename = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.json$`)

func isValidFilename(name string) bool {
	return validFilename.MatchString(name)
}

func pointInRect(x, y int, r image.Rectangle) bool {
	return x >= r.Min.X && x < r.Max.X && y >= r.Min.Y && y < r.Max.Y
}
