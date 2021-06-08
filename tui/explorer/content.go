package explorer

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type contentView struct {
	*tview.TextView
	rootUI *UI
}

func (c *contentView) SetNode(n *graph.Node) {
	c.Clear()
	md := n.Markdown
	var out string
	r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
	if err != nil {
		c.rootUI.StatusErrorf("NewTermRenderer err =  %v", err)
		return
	}

	out, err = r.Render(md)
	if err != nil {
		c.rootUI.StatusErrorf("render markdown: err = %v", err)
		return
	}

	if _, err := c.Write([]byte(out)); err != nil {
		c.rootUI.StatusErrorf("Write: err = %v", err)
		return
	}

	c.ScrollToBeginning()
}

func (c *contentView) SetBlock(b *graph.Block) {
	c.Clear()
	c.SetText(b.Content)
	c.ScrollToBeginning()
}

func (c *contentView) NavHelp() [][]string {
	return [][]string{
		{"Down/Up", "Navigate through conent"},
		{"Tab", "Node Listing"},
	}
}

func newContentView(rootUI *UI) *contentView {
	c := &contentView{
		TextView: tview.NewTextView(),
		rootUI:   rootUI,
	}
	c.SetBorder(true)
	c.SetBorderColor(tcell.ColorDarkCyan)
	c.SetDoneFunc(c.rootUI.Navigate)
	return c
}
