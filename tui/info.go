package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type info struct {
	*tview.TextView
	rootUI *Tui
	gCtx   *snippet.Context
}

func newInfo(rootUI *Tui, gCtx *snippet.Context) *info {
	i := &info{
		TextView: tview.NewTextView(),
		rootUI:   rootUI,
		gCtx:     gCtx,
	}

	i.display()
	return i
}

func (i *info) display() {
	agentInfo := fmt.Sprintf("ver:%s", i.gCtx.Metadata.Agent)
	e, found := i.gCtx.Config.GetDefaultEntry()
	if !found {
		i.rootUI.Statusf("info.Display: config.GetDefaultEntry not found")
		return
	}

	serviceEndpoint := fmt.Sprintf("endpoint:%s", e.ServiceUrl)
	userName := fmt.Sprintf("%s", "<foobar@aardvarklabs.com>")

	i.SetTextColor(tcell.ColorYellow)
	i.SetText(fmt.Sprintf(" aardy \t| %s \n server\t| %s\n user  \t| %s\n",
		agentInfo, serviceEndpoint, userName))
	i.SetTextAlign(tview.AlignLeft)
}
