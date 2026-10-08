package leveleditor

import (
	"os"
	"testing"

	"dungeoneer/gamedata"
)

// The working directory here is src/leveleditor, so "levels/hub.json" is not
// on disk; the repo's real hub file stands in for the embedded copy.
func TestLoadLevelFromFile_EmbeddedFallback(t *testing.T) {
	gamedata.Register(os.DirFS(".."))
	t.Cleanup(func() { gamedata.Register(nil) })

	l, err := LoadLevelFromFile("levels/hub.json")
	if err != nil {
		t.Fatalf("LoadLevelFromFile: %v", err)
	}
	if l.W == 0 || l.H == 0 {
		t.Errorf("hub loaded with size %dx%d", l.W, l.H)
	}
}
