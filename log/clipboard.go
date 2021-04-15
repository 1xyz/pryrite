package log

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/atotto/clipboard"
)

func clipTo(d graph.Details) error {
	text := ""
	if body := d.GetBody(); len(body) > 0 {
		text = body
	} else if url := d.GetUrl(); len(url) > 0 {
		text = url
	} else {
		text = d.GetTitle()
	}
	if len(text) > 0 {
		return fmt.Errorf("no body/title/url to copy")
	}
	return clipboard.WriteAll(text)
}

func getClip() (string, error) {
	return clipboard.ReadAll()
}
