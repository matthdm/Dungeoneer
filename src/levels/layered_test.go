package levels

import "testing"

func TestLayeredLevelBasic(t *testing.T) {
	l1 := NewEmptyLevel(10, 10)
	l2 := NewEmptyLevel(20, 20)
	
	ll := NewLayeredLevel(l1)
	if ll.ActiveLayer() != l1 {
		t.Errorf("expected l1 as active layer")
	}
	
	ll.AddLayer(l2)
	if len(ll.Layers) != 2 {
		t.Errorf("expected 2 layers")
	}
	
	ll.SwitchToLayer(1, Point{0, 0})
	if ll.ActiveLayer() != l2 {
		t.Errorf("expected l2 as active layer")
	}
	
	// switch to invalid layer
	ll.SwitchToLayer(5, Point{0, 0})
	if ll.ActiveLayer() != l2 {
		t.Errorf("expected l2 to remain active")
	}
	
	ll.RemoveLastLayer()
	if len(ll.Layers) != 1 {
		t.Errorf("expected 1 layer")
	}
	if ll.ActiveIndex != 0 {
		t.Errorf("expected active index to revert to 0")
	}
	if ll.ActiveLayer() != l1 {
		t.Errorf("expected l1 as active layer")
	}
	
	// Try removing last layer when only 1 exists
	ll.RemoveLastLayer()
	if len(ll.Layers) != 1 {
		t.Errorf("expected 1 layer to remain")
	}
}

func TestLayeredLevelNilSafe(t *testing.T) {
	var ll *LayeredLevel
	if ll.ActiveLayer() != nil {
		t.Errorf("expected nil")
	}
	ll.AddLayer(nil)
	ll.RemoveLastLayer()
	
	ll = NewLayeredLevel(nil)
	if ll.ActiveLayer() != nil {
		t.Errorf("expected nil")
	}
	ll.SwitchToLayer(0, Point{0,0})
	ll.AddLayer(nil)
}
