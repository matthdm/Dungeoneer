package levels

// Point represents tile coordinates within a level.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// LayerLink represents a connection between two layers.
type LayerLink struct {
	FromLayerIndex int    `json:"from_layer"`
	FromTile       Point  `json:"from_tile"`
	ToLayerIndex   int    `json:"to_layer"`
	ToTile         Point  `json:"to_tile"`
	TriggerSprite  string `json:"trigger_sprite"`
	IsOneWay       bool   `json:"is_one_way"`
}

// LayeredLevel stores multiple stacked levels.
type LayeredLevel struct {
	Layers      []*Level     `json:"layers"`
	ActiveIndex int          `json:"active_index"`
	Stairwells  []*LayerLink `json:"stairwells"`
}

// ActiveLayer returns the currently active Level.
func (ll *LayeredLevel) ActiveLayer() *Level {
	if ll == nil || ll.ActiveIndex < 0 || ll.ActiveIndex >= len(ll.Layers) {
		return nil
	}
	return ll.Layers[ll.ActiveIndex]
}

// SwitchToLayer changes the active layer and positions the player at the given tile.
func (ll *LayeredLevel) SwitchToLayer(index int, entry Point) {
	if index < 0 || index >= len(ll.Layers) {
		return
	}
	ll.ActiveIndex = index
	layer := ll.Layers[index]
	if layer == nil {
		return
	}
	// Move the player's starting tile if a player is associated later.
	// This simply ensures the target tile exists.
	if entry.X >= 0 && entry.Y >= 0 && entry.X < layer.W && entry.Y < layer.H {
		// Placeholder for player reposition handled by the game.
	}
}

// NewLayeredLevel creates a layered level with a single base layer.
func NewLayeredLevel(base *Level) *LayeredLevel {
	return &LayeredLevel{Layers: []*Level{base}, ActiveIndex: 0, Stairwells: []*LayerLink{}}
}

// AddLayer appends the given level as a new layer.
func (ll *LayeredLevel) AddLayer(l *Level) {
	if ll == nil || l == nil {
		return
	}
	ll.Layers = append(ll.Layers, l)
}

// RemoveLastLayer removes the final layer if there is more than one.
func (ll *LayeredLevel) RemoveLastLayer() {
	if ll == nil || len(ll.Layers) <= 1 {
		return
	}
	ll.Layers = ll.Layers[:len(ll.Layers)-1]
	if ll.ActiveIndex >= len(ll.Layers) {
		ll.ActiveIndex = len(ll.Layers) - 1
	}
}
