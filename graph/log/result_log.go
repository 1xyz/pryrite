package log

import (
	"github.com/aardlabs/terminal-poc/tools"
	"time"
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

	// Err can be marshalled to json or yaml
	Err *tools.MarshalledError `yaml:"err,omitempty" json:"err,omitempty"`

	// The Content can change (in the referenced block)
	// so persist the original  command alongside
	Content string `yaml:"content" json:"content"`
}

type ResultLog interface {
	// Len returns the number of entries in the log
	Len() int

	// Each provides a callback fn (intended to be called as an iterator
	Each(cb func(int, *ResultLogEntry) bool)

	// Find looks up a specific result log entry by ID
	Find(id string) (*ResultLogEntry, bool)

	// Append an entry to the ResultLog
	Append(entry *ResultLogEntry)
}

// ResultLogIndex provides an interface to provide Results primarily by nodeID
type ResultLogIndex interface {
	// Get the ResultLog associated with the nodeID
	Get(nodeID string) (ResultLog, error)

	// Append an entry to the ResultLog associated with the nodeID
	Append(*ResultLogEntry) error
}
