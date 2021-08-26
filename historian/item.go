package historian

import (
	"fmt"
	"time"
)

// TODO: add current working directory for each item (and pass along to slurp)
type Item struct {
	ID          uint64        `json:"id"`
	RecordedAt  time.Time     `json:"recorded_at"`
	CommandLine string        `json:"command_line,omitempty"`
	ParentPID   *int          `json:"parent_pid,omitempty"`
	ExitStatus  *int          `json:"exit_status,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
}

//------------------------------------------------------------------------

func (item *Item) String() string {
	tsFmt := time.RFC3339
	if time.Since(item.RecordedAt) < 24*time.Hour {
		tsFmt = "15:04" // or time.Kitchen?
	}

	ts := item.RecordedAt.Format(tsFmt)

	duration := "?"
	if item.Duration > 0 {
		duration = fmt.Sprint(item.Duration.Round(time.Millisecond))
	}

	exitStatus := "?"
	if item.ExitStatus != nil {
		exitStatus = fmt.Sprint(*item.ExitStatus)
	}

	return fmt.Sprintf("%d  %s %7s %3s %s",
		item.ID, ts, duration, exitStatus, item.CommandLine)
}
