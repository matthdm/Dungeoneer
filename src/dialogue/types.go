package dialogue

import "fmt"

// DialogueTree is a complete conversation graph loaded from JSON.
type DialogueTree struct {
	ID    string                    `json:"id"`
	Root  string                    `json:"root"`
	Nodes map[string]*DialogueNode  `json:"nodes"`
}

// DialogueNode is a single point in a conversation.
type DialogueNode struct {
	ID        string             `json:"id"`
	Speaker   string             `json:"speaker"`
	Text      string             `json:"text"`
	Portrait  string             `json:"portrait"`
	Responses []DialogueResponse `json:"responses"`
	OnEnter   []DialogueAction   `json:"on_enter,omitempty"`
}

// DialogueResponse is a player-selectable reply.
type DialogueResponse struct {
	Text      string             `json:"text"`
	NextNode  string             `json:"next_node"`
	Condition *DialogueCondition `json:"condition,omitempty"`
	OnSelect  []DialogueAction   `json:"on_select,omitempty"`
}

// DialogueCondition gates whether a response is shown.
type DialogueCondition struct {
	Type   string `json:"type"`    // "flag_equals", "flag_gte", "flag_lte", "has_item", "not_flag"
	Flag   string `json:"flag"`
	Value  int    `json:"value"`
	ItemID string `json:"item_id,omitempty"`
}

// DialogueAction is a side effect triggered by entering a node or selecting a response.
type DialogueAction struct {
	Type   string `json:"type"` // "set_flag", "add_flag", "give_item", "take_item", "give_exp"
	Flag   string `json:"flag,omitempty"`
	Value  int    `json:"value,omitempty"`
	ItemID string `json:"item_id,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

// SimpleDialogue is a linear (non-branching) conversation for minor NPCs.
type SimpleDialogue struct {
	Speaker  string   `json:"speaker"`
	Portrait string   `json:"portrait"`
	Lines    []string `json:"lines"`
}

// ToTree converts a SimpleDialogue into a linear DialogueTree so the same
// panel can render both branching and non-branching conversations.
func (sd *SimpleDialogue) ToTree(id string) *DialogueTree {
	tree := &DialogueTree{
		ID:    id,
		Root:  "line_0",
		Nodes: make(map[string]*DialogueNode, len(sd.Lines)),
	}
	for i, line := range sd.Lines {
		nodeID := fmt.Sprintf("line_%d", i)
		node := &DialogueNode{
			ID:       nodeID,
			Speaker:  sd.Speaker,
			Text:     line,
			Portrait: sd.Portrait,
		}
		if i < len(sd.Lines)-1 {
			nextID := fmt.Sprintf("line_%d", i+1)
			node.Responses = []DialogueResponse{
				{Text: "[Continue]", NextNode: nextID},
			}
		}
		// Last node has no responses — panel closes on click.
		tree.Nodes[nodeID] = node
	}
	return tree
}
