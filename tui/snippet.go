package tui

import (
	"fmt"
	"github.com/charmbracelet/glamour"
	"strconv"
	"strings"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
)

type snippetView struct {
	*detailView
	codeBlocks    *blockIndex  // An index of code blocks
	selectedBlock *graph.Block // Indicates the currently selected block
}

// Refresh refreshes the view with the provided nodeview object
func (s *snippetView) Refresh(view *graph.NodeView) {
	s.Clear()
	s.codeBlocks.clear()
	s.updateDetailsContent(view.Node)
}

func (s *snippetView) NavHelp() string {
	help := " ctrl+r: run code snippet, ctrl+e: edit the selected snippet"
	navigate := " tab: select snippet or pane, shift+tab: previous snippet/pane"
	navHelp := fmt.Sprintf(" commands \t| %s\n navigate \t| %s\n", help, navigate)
	return navHelp
}

func (s *snippetView) updateDetailsContent(n *graph.Node) {
	if n == nil {
		return
	}

	s.codeBlocks.clear()
	// construct the markdown manually but inject blocks
	mdBuf := strings.Builder{}
	if n.HasBlocks() {
		prevCodeBlock := false
		for _, b := range n.Blocks {
			var blockContent string
			if b.IsCode() {
				// Inject region blocks. each region block is of the format ["0"]...[""]
				blockContent = fmt.Sprintf(`["%d"]%s[""]`, s.codeBlocks.count(), strings.TrimSpace(b.Content))
				s.codeBlocks.add(b)
				prevCodeBlock = true
			} else {
				if prevCodeBlock {
					// only add a new line if the content starts with ```
					s := strings.TrimSpace(b.Content)
					strings.HasPrefix(s, "```")
					blockContent = "\n" + b.Content
				} else {
					blockContent = b.Content
				}
				prevCodeBlock = false
			}
			mdBuf.WriteString(blockContent)
		}
	}
	md := mdBuf.String()
	//os.WriteFile("/tmp/foo.md", []byte(md), 0600)

	var out string
	r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
	if err != nil {
		s.rootUI.StatusErrorf("SetSelectedFunc: NewTermRenderer err =  %v", err)
		s.codeBlocks.clear()
		return
	}
	out, err = r.Render(md)
	if err != nil {
		s.rootUI.StatusErrorf("updateDetailsContent: render markdown: err = %v", err)
		s.codeBlocks.clear()
		return
	}

	////if n.LastExecutedAt != nil {
	////	// include last execution info (FIXME: there is certainly a better way to include this)
	////	out += fmt.Sprintf("\n\n[ Most recently run by %s %s. ]\n",
	////		n.LastExecutedBy, humanize.Time(*n.LastExecutedAt))
	////}

	if _, err := s.Write([]byte(out)); err != nil {
		s.rootUI.StatusErrorf("updateDetailsContent: render markdown: err = %v", err)
		return
	}

	s.setDoneFn()
}

func (s *snippetView) setKeybinding() {
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//g.setGlobalKeybinding(event)
		switch event.Key() {
		case tcell.KeyCtrlR:
			tools.Log.Info().Msgf("snippetView: Ctrl+R request to run node")
			s.rootUI.ExecuteSelectedBlock(s.selectedBlock)
		case tcell.KeyCtrlE:
			tools.Log.Info().Msgf("snippetView: Ctrl+E request to edit node")
			s.rootUI.EditCurrentNode()
		}

		switch event.Rune() {
		case 'c':
			//s.rootUI.Statusf("time to edit shit")
			//case 'p':
			//	g.pullImageForm()
			//case 'd':
			//	g.removeImage()
			//case 'i':
			//	g.importImageForm()
			//case 's':
			//	g.saveImageForm()
			//case 'f':
			//	newSearchInputField(g)
		}

		return event
	})
}

func (s *snippetView) setDoneFn() {
	s.SetDoneFunc(func(key tcell.Key) {
		// Returns back to see if any block is selected
		currentSelection := s.GetHighlights()

		if len(currentSelection) > 0 {
			// Indicates a block is selected. get the index of that block
			index, _ := strconv.Atoi(currentSelection[0])
			if key == tcell.KeyTab {
				// if that block is the last selected one
				// Tab should take it to the next pane.
				// Toggle the highlight and send it to the rootUI
				if index == s.codeBlocks.count()-1 {
					s.Highlight()
					s.rootUI.Navigate(key)
					s.selectedBlock = nil
					return
				}
				// This is not the last selected one
				index = (index + 1) % s.codeBlocks.count()
			} else if key == tcell.KeyBacktab {
				// this is the first selected block
				// BackTab should take it to the previous pane
				// Toggle the highlight and send it to the root UI
				if index == 0 {
					s.Highlight()
					s.rootUI.Navigate(key)
					s.selectedBlock = nil
					return
				}
				// THis is not the first selected one
				index = (index - 1 + s.codeBlocks.count()) % s.codeBlocks.count()
			} else {
				return
			}
			// Hight the current selected one
			s.Highlight(strconv.Itoa(index)).ScrollToHighlight()
			s.selectedBlock = s.codeBlocks.get(index)
		} else if s.codeBlocks.count() > 0 {
			// Nothing is selected so let us find the zero element and tab
			s.Highlight("0").ScrollToHighlight()
			s.selectedBlock = s.codeBlocks.get(0)
		} else {
			s.rootUI.Navigate(key)
			s.selectedBlock = nil
			return
		}
	})
}

func newSnippetView(rootUI *Tui) *snippetView {
	s := &snippetView{
		detailView:    newDetailView("selected snippet", true, rootUI),
		codeBlocks:    newBlockIndex(),
		selectedBlock: nil,
	}
	s.setKeybinding()
	s.setDoneFn()
	return s
}

type blockIndex struct{ index []*graph.Block }

func newBlockIndex() *blockIndex                 { return &blockIndex{[]*graph.Block{}} }
func (b *blockIndex) add(blk *graph.Block)       { b.index = append(b.index, blk) }
func (b *blockIndex) clear()                     { b.index = []*graph.Block{} }
func (b *blockIndex) get(index int) *graph.Block { return b.index[index] }
func (b *blockIndex) count() int                 { return len(b.index) }
