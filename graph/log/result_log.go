package log

import (
	"errors"
	"fmt"
	"time"

	"github.com/aardlabs/terminal-poc/tools"
)

type ExecState string

const (
	ExecStateUnknown   ExecState = "Unknown"
	ExecStateQueued    ExecState = "Queued"
	ExecStateStarted   ExecState = "Started"
	ExecStateCanceled  ExecState = "Cancel-Requested"
	ExecStateCompleted ExecState = "Completed"
	ExecStateFailed    ExecState = "Failed"
)

var (
	ErrResultLogEntryNotFound = errors.New("result Log Entry not found")
	ErrResultLogNotFound      = errors.New("result log not found")
)

type ResultLogEntry struct {
	ID          string     `yaml:"id" json:"id"`
	ExecutionID string     `yaml:"execution_id" json:"execution_id"`
	NodeID      string     `yaml:"node_id" json:"node_id"`
	BlockID     string     `yaml:"block_id" json:"block_id"`
	RequestID   string     `yaml:"request_id" json:"request_id"`
	ExitStatus  string     `yaml:"exit_status" json:"exit_status"`
	Stdout      string     `yaml:"stdout" json:"stdout"`
	Stderr      string     `yaml:"stderr" json:"stderr"`
	ExecutedAt  *time.Time `yaml:"executed_at,omitempty" json:"executed_at,omitempty"`
	ExecutedBy  string     `yaml:"executed_by" json:"executed_by"`
	State       ExecState  `yaml:"state" json:"state"`

	// Err is a string representation of error
	Err string `yaml:"err,omitempty" json:"err,omitempty"`

	// The Content can change (in the referenced block)
	// so persist the original  command alongside
	Content string `yaml:"content" json:"content"`
}

func (e *ResultLogEntry) SetError(err error) {
	if err == nil {
		e.Err = ""
	} else {
		e.Err = err.Error()
	}
}

func NewResultLogEntry(executionID, nodeID, blockID, requestID, executedBy, content string) *ResultLogEntry {
	now := time.Now().UTC()
	res := &ResultLogEntry{
		ID:          tools.RandAlphaNumericStr(8),
		ExecutionID: executionID,
		NodeID:      nodeID,
		BlockID:     blockID,
		RequestID:   requestID,
		ExecutedAt:  &now,
		ExecutedBy:  executedBy,
		Err:         "",
		Content:     content,
		ExitStatus:  "",
		Stdout:      "",
		Stderr:      "",
		State:       ExecStateUnknown,
	}

	return res
}

type ResultLog interface {
	// Len returns the number of entries in the log
	Len() (int, error)

	// Each provides a callback fn (intended to be called as an iterator
	Each(cb func(int, *ResultLogEntry) bool) error

	// Find looks up a specific result log entry by ID
	Find(id string) (*ResultLogEntry, error)

	// Append an entry to the ResultLog
	Append(entry *ResultLogEntry) error
}

// ResultLogIndex provides an interface to provide Results primarily by nodeID
type ResultLogIndex interface {
	// Get the ResultLog associated with the nodeID
	Get(nodeID string) (ResultLog, error)

	// Append an entry to the ResultLog associated with the nodeID
	Append(*ResultLogEntry) error
}

type LogIndexType int

const (
	IndexUnknown    LogIndexType = 0
	IndexInMemory   LogIndexType = 1
	IndexFileSystem LogIndexType = 2
)

var ResultLogDir = tools.MyPathTo("result_log")

func NewResultLogIndex(typ LogIndexType) (ResultLogIndex, error) {
	switch typ {
	case IndexInMemory:
		return newInMemLogIndex(), nil
	case IndexFileSystem:
		fsIndex, err := newFSLogIndex(ResultLogDir)
		if err != nil {
			return nil, err
		}
		return fsIndex, nil
	default:
		return nil, fmt.Errorf("un-supported type %v", typ)
	}
}
