package tui

import (
	"fmt"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/charmbracelet/glamour"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
)

type snippetView struct {
	*detailView
}

// Refresh refreshes the view with the provided nodeview object
func (s *snippetView) Refresh(view *graph.NodeView) {
	s.Clear()
	s.updateDetailsContent(view)
}

func (s *snippetView) NavHelp() string {
	help := " ctrl+r: run snippet, ctrl+e: edit the selected snippet"
	navigate := " tab: next pane, shift+tab: previous pane"
	navHelp := fmt.Sprintf(" commands \t| %s\n navigate \t| %s\n", help, navigate)
	return navHelp
}

func (s *snippetView) updateDetailsContent(view *graph.NodeView) {
	if view == nil {
		return
	}

	var out string
	r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
	if err != nil {
		s.rootUI.StatusErrorf("SetSelectedFunc: NewTermRenderer err =  %v", err)
		return
	}

	out, err = r.Render(view.ContentMarkdown)
	if err != nil {
		s.rootUI.StatusErrorf("updateDetailsContent: render markdown: err = %v", err)
		return
	}

	if view.Node.LastExecutedAt != nil {
		// include last execution info (FIXME: there is certainly a better way to include this)
		out += fmt.Sprintf("\n\n[ Most recently run by %s %s. ]\n",
			view.Node.LastExecutedBy, humanize.Time(*view.Node.LastExecutedAt))
	}

	if _, err := s.Write([]byte(out)); err != nil {
		s.rootUI.StatusErrorf("updateDetailsContent: render markdown: err = %v", err)
		return
	}
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

func newSnippetView(rootUI *Tui) *snippetView {
	s := &snippetView{
		detailView: newDetailView("selected snippet", true, rootUI),
	}
	s.setKeybinding()
	return s
}
