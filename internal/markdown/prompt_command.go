package markdown

import (
	"fmt"
	"mime"
	"regexp"

	"github.com/rs/zerolog/log"
)

var promptRE = regexp.MustCompile(`(?m)^\s*(\w+)(=?)>\s*(.+)$`)

type promptCommand struct {
	prompt   position
	isAssign bool
	command  position
}

type position struct {
	start int
	stop  int
}

func (pos position) Join(char string) string {
	if pos.stop > -1 {
		return fmt.Sprint(pos.start, char, pos.stop)
	}

	return fmt.Sprint(pos.start)
}

func (pos position) String() string {
	if pos.stop > -1 {
		return fmt.Sprint(pos.start, ":", pos.stop)
	}

	return fmt.Sprint(pos.start, ":")
}

func extractPromptCommand(content string) *promptCommand {
	locations := promptRE.FindAllStringSubmatchIndex(content, -1)
	if locations == nil {
		return nil
	}

	// FIXME: consider how to propagate content-type prompt extraction errors to a customer

	if len(locations) > 1 {
		log.Error().Str("content", content).Msg("Found more than one prompt commands within the content")
	}

	// for now, we only support the first one found
	loc := locations[0]

	return &promptCommand{
		prompt:   position{loc[2], loc[3]},      // 1st capture group
		isAssign: content[loc[4]:loc[5]] == "=", // 2nd capture group
		command:  position{loc[6], -1},          // 3rd capture group (NOTE: no "stop"--use remaining content)
	}
}

func extractContentType(content, language string) string {
	typeSubtype := "text/" + language

	pc := extractPromptCommand(content)
	if pc == nil {
		return typeSubtype
	}

	params := map[string]string{
		"command": pc.command.String(),
	}

	prompt := pc.prompt.String()
	if pc.isAssign {
		params["prompt-assign"] = prompt
	} else {
		params["prompt"] = prompt
	}

	return mime.FormatMediaType(typeSubtype, params)
}
