package tools

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"io"
)

var (
	Log = zlog.Logger
)

func InitLogger(w io.Writer, level zerolog.Level) {
	Log = zlog.Output(w).Level(level)
}
