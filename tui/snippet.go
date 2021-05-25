package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"
)

type snippetView struct {
	*DetailPane
	// Represents the current node shown in the detail
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
		case tcell.KeyEnter:
			s.rootUI.Statusf("Enter pressed")
		case tcell.KeyCtrlL:
			s.rootUI.Statusf("Ctrl + L pressed")
		case tcell.KeyCtrlR:
			s.rootUI.Statusf("Ctrl + R pressed")
			if s.curNodeView == nil {
				break
			}
			result, err := s.rootUI.Execute(s.curNodeView.Node, s.rootUI.Execution, s.rootUI.Execution)
			if err != nil {
				s.rootUI.Statusf("Run: Execute(node): err = %v", err)
			} else {
				body := fmt.Sprintf("requestID = %v err = %v", result.RequestID, result.Err)
				if _, err := s.Write([]byte(body)); err != nil {
					s.rootUI.Statusf("Run: write(result) err = %v", err)
				}
			}
		}

		switch event.Rune() {
		case 'c':
			s.rootUI.Statusf("time to edit shit")
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

func (s *snippetView) executeCommand() {

}
