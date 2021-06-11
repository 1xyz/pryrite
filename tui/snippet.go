package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
)

type snippetView struct {
	*detailView
	codeBlocks      *blockIDs // An index of code blocks IDs
	selectedBlockID string    // Indicates the blockID currently selected block
}

// Refresh refreshes the view with the provided nodeview object
func (s *snippetView) Refresh(view *graph.NodeView) {
	s.Clear()
	s.codeBlocks.clear()
	s.updateDetailsContent(view.Node)
	s.updateTitle(view.Node)
}

func (b *snippetView) NavHelp() [][]string {
	return [][]string{
		{"Ctrl + E", "Edit selected code block"},
		{"Ctrl + R", "Run selected node or code-block"},
		{"Tab", "Navigate to the next pane"},
		{"Shift + Tab", "Navigate to the previous pane"},
	}
}

func (s *snippetView) updateTitle(n *graph.Node) {
	if n == nil {
		s.SetTitle("")
		return
	}

	s.SetTitle(fmt.Sprintf("Node: %s (%s)", n.Title, n.ID))
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
				s.codeBlocks.add(b.ID)
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

	if _, err := s.Write([]byte(out)); err != nil {
		s.rootUI.StatusErrorf("updateDetailsContent: render markdown: err = %v", err)
		return
	}

	s.setDoneFn()
}

func (s *snippetView) setKeybinding() {
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = s.rootUI.GlobalKeyBindings(event)

		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown:
			s.navigateBlocks(event.Key())
		case tcell.KeyCtrlR:
			if s.selectedBlockID != noBlock {
				if err := s.rootUI.ExecuteSelectedBlock(s.selectedBlockID); err != nil {
					s.rootUI.StatusErrorf("ExecuteSelectedBlock: err = %v", err)
				}
			} else {
				if err := s.rootUI.ExecuteCurrentNode(); err != nil {
					s.rootUI.StatusErrorf("ExecuteCurrentNode (%s) failed err = %v", s.rootUI.curNodeID, err)
				}
			}
		case tcell.KeyCtrlE:
			tools.Log.Info().Msgf("snippetView: Ctrl+E request to edit node")
			if s.selectedBlockID != noBlock {
				if err := s.rootUI.EditSelectedBlock(s.selectedBlockID); err != nil {
					s.rootUI.StatusErrorf("EditSelectedBlock (%s) failed err = %v", s.selectedBlockID, err)
				}
			}
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

const noBlock = ""

// setDoneFn is the primary navigate function used to intercept Enter/ESC or Tab/BackTab
func (s *snippetView) setDoneFn() {
	s.SetDoneFunc(s.navigateBlocks)
}

func (s *snippetView) navigateBlocks(key tcell.Key) {
	// Returns back to see if any block is selected
	currentSelection := s.GetHighlights()

	// Check to see if anything is selected
	if len(currentSelection) > 0 {
		// Indicates a block is selected. get the index of that block selected
		index, _ := strconv.Atoi(currentSelection[0])
		if key == tcell.KeyTab || key == tcell.KeyDown {
			// if that block is the last selected one
			// Tab should take it to the next pane.
			// Toggle the highlight and send it to the rootUI
			if index == s.codeBlocks.count()-1 {
				s.Highlight()
				s.rootUI.Navigate(key)
				s.selectedBlockID = noBlock
				return
			}
			// This is not the last selected one so let us handle it
			index = (index + 1) % s.codeBlocks.count()
		} else if key == tcell.KeyBacktab || key == tcell.KeyUp {
			// This is the first selected block
			// BackTab should take it to the previous pane
			// Toggle the highlight and send it to the root UI
			if index == 0 {
				s.Highlight()
				s.rootUI.Navigate(key)
				s.selectedBlockID = noBlock
				return
			}
			// THis is not the first selected one
			index = (index - 1 + s.codeBlocks.count()) % s.codeBlocks.count()
		} else {
			// We don't handle any keys other than Tab/BackTab here
			return
		}
		// Highlight the current selected one
		s.Highlight(strconv.Itoa(index)).ScrollToHighlight()
		// the current highlighted block is the selected one
		s.selectedBlockID = s.codeBlocks.get(index)
	} else if s.codeBlocks.count() > 0 {
		// Nothing is selected so let us find the element and tab
		index := 0
		if key == tcell.KeyBacktab {
			// we have back tabbed into this so select the last one
			index = s.codeBlocks.count() - 1
		}
		s.Highlight(strconv.Itoa(index)).ScrollToHighlight()
		s.selectedBlockID = s.codeBlocks.get(index)
	} else {
		s.rootUI.Navigate(key)
		s.selectedBlockID = noBlock
		return
	}
}

func newSnippetView(rootUI *Tui) *snippetView {
	s := &snippetView{
		detailView:      newDetailView("", true, rootUI),
		codeBlocks:      newBlockIndex(),
		selectedBlockID: noBlock,
	}
	s.setKeybinding()
	s.setDoneFn()
	return s
}

type blockIDs struct{ index []string }

func newBlockIndex() *blockIDs           { return &blockIDs{[]string{}} }
func (b *blockIDs) add(blockID string)   { b.index = append(b.index, blockID) }
func (b *blockIDs) clear()               { b.index = []string{} }
func (b *blockIDs) get(index int) string { return b.index[index] }
func (b *blockIDs) count() int           { return len(b.index) }
