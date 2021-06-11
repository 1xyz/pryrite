package tui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Level string

const (
	Info  Level = "Info"
	Error Level = "Error"
)

type activity struct {
	Level Level     `yaml:"level"`
	Msg   string    `yaml:"msg"`
	At    time.Time `yaml:"at"`
}

type activityView struct {
	*tview.Table
	rootUI     *Tui
	index      int
	activities []*activity
}

func newActivityView(rootUI *Tui) *activityView {
	a := &activityView{
		Table: tview.NewTable().
			SetSelectable(true, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI:     rootUI,
		index:      1,
		activities: []*activity{},
	}
	a.SetDoneFunc(rootUI.Navigate)
	a.setKeybinding()
	a.Table.SetBorder(true).
		SetTitle("Activity Log").
		SetTitleAlign(tview.AlignLeft)
	a.display()
	return a
}

func (a *activityView) display() {
	table := a.Clear()

	headers := []string{
		"#",
		"Level",
		"Message",
		"Time",
	}

	for i, header := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
	}
	a.Log(Info, "Initialized terminal console")
}

func (a *activityView) Log(level Level, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	entry := &activity{
		Level: level,
		Msg:   msg,
		At:    time.Now().UTC(),
	}
	showColor := tcell.ColorYellow
	if entry.Level == Error {
		showColor = tcell.ColorRed
	}

	a.SetCell(a.index, 0, tview.NewTableCell(strconv.Itoa(a.index)).
		SetTextColor(showColor).
		SetMaxWidth(1).
		SetExpansion(1))
	a.SetCell(a.index, 1, tview.NewTableCell(string(entry.Level)).
		SetTextColor(showColor).
		SetMaxWidth(3).
		SetExpansion(3))
	a.SetCell(a.index, 2, tview.NewTableCell(entry.Msg).
		SetTextColor(showColor).
		SetMaxWidth(20).
		SetExpansion(20))
	a.SetCell(a.index, 3, tview.NewTableCell(entry.At.Format("2006/01/02 15:04:05")).
		SetTextColor(showColor).
		SetMaxWidth(5).
		SetExpansion(5))
	a.Select(a.index, 0)
	a.activities = append(a.activities, entry)
	a.index++
}

func (a *activityView) setKeybinding() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = a.rootUI.GlobalKeyBindings(event)
		switch event.Key() {
		case tcell.KeyEnter:
			r, _ := a.GetSelection()
			entry := a.activities[r-1]
			a.rootUI.InspectActivity(entry)
		}
		return event
	})
}

func (a *activityView) NavHelp() [][]string {
	return [][]string{
		{"Enter", "Inspect Activity Detail"},
		{"⇵ Down/Up", "Navigate through executions"},
		{"Tab", "Navigate to the next pane"},
		{"Shift + Tab", "Navigate to the previous pane"},
	}
}

func (a *activityView) Focus(delegate func(p tview.Primitive)) {
	a.SetTitleColor(a.rootUI.focusColor)
	a.SetBorderColor(a.rootUI.focusColor)
	a.Table.Focus(delegate)
}

func (a *activityView) Blur() {
	a.SetTitleColor(tcell.ColorDefault)
	a.SetBorderColor(tcell.ColorDefault)
	a.Table.Blur()
}
