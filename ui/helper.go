package ui

import "regexp"

var validFilename = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.json$`)

func isValidFilename(name string) bool {
	return validFilename.MatchString(name)
}
