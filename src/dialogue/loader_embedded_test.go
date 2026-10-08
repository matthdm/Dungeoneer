package dialogue

import (
	"testing"
	"testing/fstest"

	"dungeoneer/gamedata"
)

// The working directory for these tests is src/dialogue, where no
// "dialogues" directory exists on disk, so every read must come from the
// registered embedded copy.

const embeddedTreeJSON = `{"id":"embedded_tree","root":"start","nodes":{"start":{"text":"hi"}}}`
const embeddedSimpleJSON = `{"speaker":"Monk","lines":["one","two"]}`

func useEmbeddedDialogues(t *testing.T) {
	t.Helper()
	gamedata.Register(fstest.MapFS{
		"dialogues/embedded_tree.json":   {Data: []byte(embeddedTreeJSON)},
		"dialogues/embedded_simple.json": {Data: []byte(embeddedSimpleJSON)},
	})
	t.Cleanup(func() {
		gamedata.Register(nil)
		delete(Registry, "embedded_tree")
		delete(Registry, "embedded_simple")
	})
}

func TestLoadAll_EmbeddedFallback(t *testing.T) {
	useEmbeddedDialogues(t)

	if err := LoadAll("dialogues"); err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if _, ok := Registry["embedded_tree"]; !ok {
		t.Error("embedded_tree not in Registry")
	}
	if _, ok := Registry["embedded_simple"]; !ok {
		t.Error("embedded_simple not in Registry")
	}
}

func TestLoadTree_EmbeddedFallback(t *testing.T) {
	useEmbeddedDialogues(t)

	tree, err := LoadTree("dialogues/embedded_tree.json")
	if err != nil {
		t.Fatalf("LoadTree: %v", err)
	}
	if tree.ID != "embedded_tree" {
		t.Errorf("tree.ID = %q, want embedded_tree", tree.ID)
	}
}

func TestLoadSimple_EmbeddedFallback(t *testing.T) {
	useEmbeddedDialogues(t)

	tree, err := LoadSimple("dialogues/embedded_simple.json")
	if err != nil {
		t.Fatalf("LoadSimple: %v", err)
	}
	if tree.ID != "embedded_simple" {
		t.Errorf("tree.ID = %q, want embedded_simple", tree.ID)
	}
}
