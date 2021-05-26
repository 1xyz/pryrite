package tui

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"
)

type snippetView struct {
	*DetailPane

	// Represents the current node snippet to be shown in the detail
	curNodeView *graph.NodeView
}

func newSnippetView(rootUI *Tui) *snippetView {
	s := &snippetView{
		DetailPane: NewDetailPane("Snippet", rootUI),
	}
	s.setKeybinding()
	return s
}

func (s *snippetView) SetCurrentNodeView(nodeView *graph.NodeView) {
	s.curNodeView = nodeView
	s.Clear()
	s.updateDetailsContent()
}

func (s *snippetView) updateDetailsContent() {
	if s.curNodeView == nil {
		return
	}
	var out string
	r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
	if err != nil {
		s.rootUI.Statusf("SetSelectedFunc: NewTermRenderer err =  %v", err)
		return
	}

	out, err = r.Render(s.curNodeView.ContentMarkdown)
	if err != nil {
		s.rootUI.Statusf("SetSelectedFunc: render markdown: err = %v", err)
		return
	}
	if _, err := s.Write([]byte(out)); err != nil {
		s.rootUI.Statusf("SetSelectedFunc: render markdown: err = %v", err)
		return
	}
}

func (s *snippetView) setKeybinding() {
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//g.setGlobalKeybinding(event)
		switch event.Key() {
		case tcell.KeyCtrlR:
			if s.curNodeView == nil {
				break
			}

			tools.Log.Info().Msgf("Ctrl+R. request to execute node = %v", s.curNodeView.Node.ID)
			s.rootUI.SetExecutionInProgress()
			result, err := s.rootUI.Execute(s.curNodeView.Node, s.rootUI.execOutView, s.rootUI.execOutView)
			if err != nil {
				s.rootUI.Statusf("Run: Execute(node): err = %v", err)
				break
			}
			s.rootUI.SetCurrentNodeExecutionResult(result)
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
