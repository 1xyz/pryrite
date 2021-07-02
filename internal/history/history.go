package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aardlabs/terminal-poc/tools"
)

var HistoryDir = os.ExpandEnv("$HOME/.aardvark/history")

type History interface {
	// GetAll retrieves all the items and returns only the commands as a string slice
	GetAll() ([]string, error)

	// Append a command to the history
	Append(command string) error
}

func New(file string) (History, error) {
	return newLocalHistory(file)
}

type item struct {
	CreatedAt *time.Time `json:"createdAt"`
	Command   string     `json:"command"`
}

type localHistory struct {
	historyFile string
}

func (h *localHistory) GetAll() ([]string, error) {
	items, err := h.getAll()
	if err != nil {
		return nil, err
	}

	result := make([]string, len(items))
	for i := range items {
		result[i] = items[i].Command
	}
	return result, nil
}

func (h *localHistory) Append(command string) error {
	now := time.Now().UTC()
	h.append(&item{
		CreatedAt: &now,
		Command:   command,
	})
	return nil
}

func newLocalHistory(filename string) (*localHistory, error) {
	fp, err := tools.OpenFile(filename, os.O_WRONLY|os.O_CREATE)
	if err != nil {
		return nil, fmt.Errorf("filename = %s, err = %v", filename, err)
	}
	defer tools.CloseFile(fp)
	return &localHistory{historyFile: filename}, nil
}

func (h *localHistory) append(item *item) error {
	fp, err := tools.OpenFile(h.historyFile, os.O_APPEND|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("filename = %s, err = %v", h.historyFile, err)
	}
	defer tools.CloseFile(fp)
	return json.NewEncoder(fp).Encode(item)
}

func (h *localHistory) getAll() ([]item, error) {
	fp, err := tools.OpenFile(h.historyFile, os.O_RDONLY)
	if err != nil {
		return nil, fmt.Errorf("filename = %s, err = %v", h.historyFile, err)
	}
	defer tools.CloseFile(fp)

	result := make([]item, 0)
	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		l := sc.Text()
		var item item
		if err := json.Unmarshal([]byte(l), &item); err != nil {
			return nil, fmt.Errorf("json.Unmarshal line = %s err = %v", l, err)
		}
		result = append(result, item)
	}
	return result, nil
}
