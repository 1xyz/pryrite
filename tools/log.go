package tools

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"io"
	"time"
)

var (
	Log = zlog.Logger
)

func InitLogger(w io.Writer, level zerolog.Level) {
	Log = zlog.Output(w).Level(level)
}

func TrimLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func FmtTime(t time.Time) string {
	return t.In(time.Local).Format("Jan _2 3:04PM")
}
