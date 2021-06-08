package explorer

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type statusView struct {
	*tview.TextView
	rootUI *UI
}

func newStatusView(rootUI *UI) *statusView {
	return &statusView{
		TextView: tview.NewTextView(),
		rootUI:   rootUI,
	}
}

func (s *statusView) Infof(format string, v ...interface{}) {
	tools.Log.Info().Msgf(format, v)
	s.msgf(tcell.ColorYellow, format, v...)
}

func (s *statusView) Errorf(format string, v ...interface{}) {
	tools.Log.Error().Msgf(format, v)
	s.msgf(tcell.ColorRed, format, v...)
}

func (s *statusView) msgf(color tcell.Color, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v)
	s.Clear()
	s.SetTextColor(color)
	s.SetText(msg)
}
