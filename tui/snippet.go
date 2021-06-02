package tui

import (
	"fmt"
	"github.com/charmbracelet/glamour"
	"os"
	"strconv"
	"strings"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
)

type snippetView struct {
	*detailView
	nCodeBlocks int
}

// Refresh refreshes the view with the provided nodeview object
func (s *snippetView) Refresh(view *graph.NodeView) {
	s.Clear()
	s.nCodeBlocks = 0
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

	nCodeBlocks := 0
	// construct the markdown manually but inject blocks
	mdBuf := strings.Builder{}
	if n.HasBlocks() {
		for _, b := range n.Blocks {
			var blockContent string
			if b.IsCode() {
				blockContent = fmt.Sprintf(`["%d"]%s[""]`, nCodeBlocks, strings.TrimSpace(b.Content))
				blockContent += "\n"
				nCodeBlocks++
			} else {
				blockContent = b.Content
			}
			mdBuf.WriteString(blockContent)
		}
	}
	md := mdBuf.String()
	os.WriteFile("/tmp/foo.md", []byte(md), 0600)

	var out string
	r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
	if err != nil {
		s.rootUI.StatusErrorf("SetSelectedFunc: NewTermRenderer err =  %v", err)
		return
	}
	out, err = r.Render(md)
	if err != nil {
		s.rootUI.StatusErrorf("updateDetailsContent: render markdown: err = %v", err)
		return
	}

	s.nCodeBlocks = nCodeBlocks

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
			s.rootUI.ExecuteCurrentNode()
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
				if index == s.nCodeBlocks-1 {
					s.Highlight()
					s.rootUI.Navigate(key)
					return
				}
				// This is not the last selected one
				index = (index + 1) % s.nCodeBlocks
			} else if key == tcell.KeyBacktab {
				// this is the first selected block
				// BackTab should take it to the previous pane
				// Toggle the highlight and send it to the root UI
				if index == 0 {
					s.Highlight()
					s.rootUI.Navigate(key)
					return
				}
				// THis is not the first selected one
				index = (index - 1 + s.nCodeBlocks) % s.nCodeBlocks
			} else {
				return
			}
			// Hight the current selected one
			s.Highlight(strconv.Itoa(index)).ScrollToHighlight()
		} else if s.nCodeBlocks > 0 {
			// Nothing is selected so let us find the zero element and tab
			s.Highlight("0").ScrollToHighlight()
		} else {
			s.rootUI.Navigate(key)
			return
		}
	})
}

func newSnippetView(rootUI *Tui) *snippetView {
	s := &snippetView{
		detailView:  newDetailView("selected snippet", true, rootUI),
		nCodeBlocks: 0,
	}
	s.setKeybinding()
	s.setDoneFn()
	return s
}
