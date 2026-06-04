package dialogue

import (
	"os"
	"testing"
)

func TestToTree(t *testing.T) {
	sd := &SimpleDialogue{
		Speaker:  "Speaker1",
		Portrait: "portrait1",
		Lines:    []string{"Line 1", "Line 2"},
	}
	
	tree := sd.ToTree("test_tree_id")
	if tree.ID != "test_tree_id" {
		t.Errorf("expected test_tree_id, got %v", tree.ID)
	}
	if tree.Root != "line_0" {
		t.Errorf("expected root line_0")
	}
	if len(tree.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %v", len(tree.Nodes))
	}
	
	n0 := tree.Nodes["line_0"]
	if n0 == nil || n0.Speaker != "Speaker1" || n0.Text != "Line 1" || n0.Portrait != "portrait1" {
		t.Errorf("node 0 fields did not map correctly")
	}
	if len(n0.Responses) != 1 || n0.Responses[0].NextNode != "line_1" {
		t.Errorf("node 0 responses incorrect")
	}
	
	n1 := tree.Nodes["line_1"]
	if len(n1.Responses) != 0 {
		t.Errorf("expected last node to have no responses")
	}
}

func TestLoadTree(t *testing.T) {
	treeJSON := `{"id": "test_tree", "root": "node1", "nodes": {"node1": {"id": "node1", "speaker": "A", "text": "B"}}}`
	
	// Create a temporary file
	f, err := os.CreateTemp("", "dialogue_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	
	if _, err := f.Write([]byte(treeJSON)); err != nil {
		t.Fatal(err)
	}
	f.Close()
	
	_, err = LoadTree(f.Name())
	if err != nil {
		t.Fatalf("LoadTree failed: %v", err)
	}
	
	if _, ok := Registry["test_tree"]; !ok {
		t.Errorf("expected test_tree to be registered")
	}
}

func TestLoadSimple(t *testing.T) {
	simpleJSON := `{"speaker": "A", "lines": ["line 1", "line 2"]}`
	
	f, err := os.CreateTemp("", "simple_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	
	if _, err := f.Write([]byte(simpleJSON)); err != nil {
		t.Fatal(err)
	}
	f.Close()
	
	_, err = LoadSimple(f.Name())
	if err != nil {
		t.Fatalf("LoadSimple failed: %v", err)
	}
	
	// ID is derived from filename. Let's find it in the Registry.
	// Since filename is simple_XXX.json, we can iterate over the registry
	found := false
	for k, tree := range Registry {
		if len(tree.Nodes) == 2 && tree.Nodes["line_0"].Text == "line 1" && tree.Nodes["line_1"].Text == "line 2" {
			found = true
			if tree.Nodes["line_0"].Responses[0].NextNode != "line_1" {
				t.Errorf("expected response target to be line_1")
			}
			break
		}
		_ = k
	}
	if !found {
		t.Fatalf("expected simple_tree to be registered")
	}
}

func TestLoadAll(t *testing.T) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "dialogue_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	
	treeJSON := `{"id": "test_loadall_tree", "root": "node1", "nodes": {"node1": {"id": "node1", "speaker": "A", "text": "B"}}}`
	simpleJSON := `{"speaker": "A", "lines": ["line 1"]}`
	
	err = os.WriteFile(dir+"/tree.json", []byte(treeJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(dir+"/simple.json", []byte(simpleJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	err = LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	
	if _, ok := Registry["test_loadall_tree"]; !ok {
		t.Errorf("expected test_loadall_tree")
	}
	if _, ok := Registry["simple"]; !ok {
		t.Errorf("expected simple_tree to be registered from basename")
	}
}

func TestLoaderErrors(t *testing.T) {
	_, err := LoadTree("nonexistent.json")
	if err == nil {
		t.Errorf("expected error reading nonexistent file")
	}
	
	f, _ := os.CreateTemp("", "badjson_*.json")
	defer os.Remove(f.Name())
	f.Write([]byte(`{bad json`))
	f.Close()
	
	_, err = LoadTree(f.Name())
	if err == nil {
		t.Errorf("expected error parsing bad json in LoadTree")
	}
	
	_, err = LoadSimple(f.Name())
	if err == nil {
		t.Errorf("expected error parsing bad json in LoadSimple")
	}
	
	err = LoadAll("nonexistent_dir")
	if err == nil {
		t.Errorf("expected error reading nonexistent dir")
	}
}
