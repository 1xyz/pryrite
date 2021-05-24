package tui

import "github.com/rivo/tview"

// NodeDetails represents the content and title of the selected node)
type NodeDetails struct {
	// UI component that displays the text
	TextArea *tview.TextView
}

func (nd *NodeDetails) Clear() {
	nd.TextArea.Clear()
}

func (nd *NodeDetails) Write(b []byte) error {
	_, err := nd.TextArea.Write(b)
	return err
}

func NewNodeDetails() (*NodeDetails, error) {
	textArea, err := createTextView()
	if err != nil {
		return nil, err
	}

	return &NodeDetails{
		TextArea: textArea,
	}, nil
}
