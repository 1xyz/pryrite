package tools

import (
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/app"
)

func FormatTime(t *time.Time) string {
	if t == nil {
		return "time(nil)"
	}
	return t.Format("2006/01/02 15:04:05")
}

func Examplef(format string, args ...string) string {
	args = append(args, "{AppName}", app.UsageName)
	r := strings.NewReplacer(args...)
	return heredoc.Doc(r.Replace(format))
}
