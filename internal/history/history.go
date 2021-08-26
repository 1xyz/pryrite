package history

import (
	"github.com/aardlabs/terminal-poc/historian"
	"github.com/aardlabs/terminal-poc/tools"
)

var HistoryDir = tools.MyPathTo("history")

type History interface {
	// GetAll retrieves all the items and returns only the commands as a string slice
	GetAll() ([]string, error)

	// Append a command to the history
	Append(command string) error
}

func New(file string) (History, error) {
	return newLocalHistory(file)
}

type localHistory struct {
	historian.Historian
}

func (h *localHistory) GetAll() ([]string, error) {
	items := []string{}
	err := h.Each(nil, func(item *historian.Item) error {
		items = append(items, item.CommandLine)
		return nil
	})
	return items, err
}

func (h *localHistory) Append(command string) error {
	return h.Put(&historian.Item{CommandLine: command})
}

func newLocalHistory(filename string) (*localHistory, error) {
	hist, err := historian.Open(filename+".db", false)
	if err != nil {
		return nil, err
	}

	return &localHistory{*hist}, nil
}
