package explorer

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type contentView struct {
	*tview.TextView
	rootUI *UI
}

func (c *contentView) SetNode(n *graph.Node) {
	c.Clear()

	gCtx := c.rootUI.GetContext()
	mr, err := tools.NewMarkdownRenderer(gCtx.ConfigEntry.Style)
	if err != nil {
		c.rootUI.StatusErrorf("NewTermRenderer err =  %v", err)
		return
	}
	out, err := mr.Render(n.Markdown)
	if err != nil {
		c.rootUI.StatusErrorf("render markdown: err = %v", err)
		return
	}

	out = tview.TranslateANSI(out)
	c.SetText(out)
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
	c.SetDynamicColors(true)
	c.SetBorderColor(tcell.ColorDarkCyan)
	c.SetTitle("Content Preview")
	c.SetTitleAlign(tview.AlignLeft)
	c.SetDoneFunc(c.rootUI.Navigate)
	return c
}
