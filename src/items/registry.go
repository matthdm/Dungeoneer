package items

// Registry holds all loaded item templates keyed by ID.
var Registry = map[string]*ItemTemplate{}

// RegisterItem adds a new template to the global registry.
func RegisterItem(tmpl *ItemTemplate) {
	Registry[tmpl.ID] = tmpl
}

// NewItem creates an item instance from a template ID.
func NewItem(id string) *Item {
	tmpl, ok := Registry[id]
	if !ok {
		panic("Invalid item ID: " + id)
	}
	return &Item{ItemTemplate: tmpl, Count: 1}
}
