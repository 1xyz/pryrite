package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"os"
	"time"
)

type Item struct {
	CreatedAt time.Time `json:"createdAt"`
	Command   string    `json:"command"`
}

type History interface {
	// Retrieve all the items from history
	GetAll() ([]Item, error)

	// Append an item to the history
	Append(*Item) error

	// Return the history item by index (zero based index)
	GetByIndex(index int) (*Item, error)
}

type localHistory struct {
	historyFile string
}

func newLocalHistory(filename string) (*localHistory, error) {
	// ensure that the file is created
	fp, err := tools.OpenFile(filename, os.O_WRONLY|os.O_CREATE)
	if err != nil {
		return nil, fmt.Errorf("filename = %s, err = %v", filename, err)
	}
	defer tools.CloseFile(fp)
	return &localHistory{historyFile: filename}, nil
}

func (h *localHistory) Append(item *Item) error {
	fp, err := tools.OpenFile(h.historyFile, os.O_APPEND|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("filename = %s, err = %v", h.historyFile, err)
	}
	defer tools.CloseFile(fp)
	return json.NewEncoder(fp).Encode(item)
}

func (h *localHistory) GetAll() ([]Item, error) {
	fp, err := tools.OpenFile(h.historyFile, os.O_RDONLY)
	if err != nil {
		return nil, fmt.Errorf("filename = %s, err = %v", h.historyFile, err)
	}
	defer tools.CloseFile(fp)

	result := make([]Item, 0)
	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		l := sc.Text()
		var item Item
		if err := json.Unmarshal([]byte(l), &item); err != nil {
			return nil, fmt.Errorf("json.Unmarshal line = %s err = %v", l, err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (h *localHistory) GetByIndex(index int) (*Item, error) {
	items, err := h.GetAll()
	if err != nil {
		return nil, err
	}
	if index >= len(items) {
		return nil, fmt.Errorf("index (%d) exceeds number of items (%d)", index, len(items))
	}
	return &items[index], nil
}

var (
	historyLogFile = os.ExpandEnv("$HOME/.pruney/history.json")
)

func New() (History, error) {
	return newLocalHistory(historyLogFile)
}
