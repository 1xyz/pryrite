package slurp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/1xyz/pryrite/shells"
	"github.com/1xyz/pryrite/tools"

	"github.com/itchyny/timefmt-go"
)

type HistorySlurper struct{}

func (bhs *HistorySlurper) Slurp(shell string, reader io.Reader, digester Digester) error {
	if shell == "" {
		parent, err := shells.GetParent()
		if err != nil {
			return err
		}

		shell = parent.Executable()
	}

	parseDuration := false
	timestampFormat := os.Getenv("HISTTIMEFORMAT")
	timestampHelp := ""
	switch shell {
	case "bash":
		timestampHelp = `Consider setting this environment variable: export HISTTIMEFORMAT='%F %T '`
	case "zsh":
		unknown := ""
		zshHistOpt := os.Getenv("AARDY_ZSH_HISTOPTS")
		for _, char := range zshHistOpt {
			switch char {
			case '-', ' ':
				// ignore
			case 'd':
				if timestampFormat == "" {
					timestampFormat = "%H:%M"
				}
			case 'i':
				timestampFormat = "%Y-%m-%d %H:%M "
			case 'D':
				parseDuration = true
			default:
				unknown += string(char)
			}
		}
		timestampHelp = `Consider using "history -iD" and setting this environment variable: export AARDY_ZSH_HISTOPTS='-iD '`
		if unknown != "" {
			tools.Log.Warn().Str("unknown", unknown).Msg("Provided unknown zsh history options")
		}
	default:
		if timestampFormat == "" {
			tools.LogStdError("WARN: Attempting to parse %s history without timestamp support", shell)
		}
	}

	if timestampFormat == "" && timestampHelp != "" {
		tools.LogStdError(
			"\nSnippets slurped will have more complete contextual info with timestamps.\n%s\n",
			timestampHelp)
	}

	// use "bash" as language since most highlighters don't support other types of shells (e.g. zsh)
	slurpFac := newSlurpFactory(shell, "bash")
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		slurp, err := bhs.parse(slurpFac, scanner.Text(), timestampFormat, parseDuration)
		if err != nil {
			return err
		}

		if slurp != nil {
			err = digester(slurp)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//--------------------------------------------------------------------------------

func (bhs *HistorySlurper) parse(slurpFac slurpFactory, line, timestampFormat string, parseDuration bool) (*Slurp, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	// split into the history index value and the remaining portion of the line
	splits := strings.SplitN(line, "  ", 2)
	if len(splits) != 2 {
		tools.LogStdError("WARN: Skipping history entry without a command: " + line)
		return nil, nil
	}

	line = strings.TrimSpace(splits[1])

	slurp := slurpFac()

	if timestampFormat != "" {
		var timestamp *time.Time
		timestamp, line = bhs.parseTimestamp(timestampFormat, line)
		if timestamp == nil {
			tools.LogStdError("WARN: Skipping history entry with invalid timestamp: " + line)
			return nil, nil
		}

		slurp.ExecutedAt = timestamp
		line = strings.TrimSpace(line)
	}

	if parseDuration {
		splits = strings.SplitN(line, " ", 2)
		if len(splits) != 2 {
			tools.LogStdError("WARN: Skipping history entry without an expected duration: " + line)
			return nil, nil
		}

		var hours, seconds uint
		n, err := fmt.Sscanf(splits[0], "%d:%d", &hours, &seconds)
		if n != 2 || err != nil {
			tools.LogStderr(err, "WARN: Skipping history entry with an invalid duration: "+line)
			return nil, nil
		}

		line = splits[1]
		duration := float64(hours*3600 + seconds)
		slurp.ExecuteSeconds = &duration
		line = strings.TrimSpace(line)
	}

	if line == "" {
		tools.LogStdError("WARN: Skipping history entry with an empty command: " + line)
		return nil, nil
	}

	slurp.Commandline = line

	return slurp, nil
}

func (bhs *HistorySlurper) parseTimestamp(format, line string) (*time.Time, string) {
	var foundTimestamp *time.Time
	var foundRemain string

	gobble(line, func(target, remain string) bool {
		timestamp, err := timefmt.ParseInLocation(target, format, time.Local)
		if foundTimestamp == nil {
			if err == nil {
				// once one is found, we need to continue until we get a failure to
				// ensure we capture some "optional" parts of dates at the end (e.g. AM/PM)
				foundTimestamp = &timestamp
				foundRemain = remain
			}
		} else {
			if err != nil {
				return true
			}

			foundTimestamp = &timestamp
			foundRemain = remain
		}

		return false
	})

	if foundTimestamp == nil {
		return nil, line
	}

	return foundTimestamp, foundRemain
}

// gobble up words off the front of a line for a handler to process
func gobble(line string, handler func(target, remain string) bool) {
	target := ""
	remain := line

	for {
		splits := strings.SplitN(remain, " ", 2)
		if len(splits) != 2 {
			return
		}

		target += splits[0] + " "
		remain = splits[1]

		if handler(target, remain) {
			return
		}
	}
}
