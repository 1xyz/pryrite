package tools

import (
	"time"
)

func FormatTime(t *time.Time) string {
	if t == nil {
		return "time(nil)"
	}
	return t.Format("2006/01/02 15:04:05")
}
