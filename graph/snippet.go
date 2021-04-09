package graph

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"os"
	"strings"
)

func AddCommandSnippet(entry *config.Entry, sessionID, agent, version, content, description string) (*Node, error) {
	return AddSnippet(entry, Command, sessionID, agent, version, content, description)
}

func AddSnippetFromFile(entry *config.Entry, kind Kind, sessionID, agent, version, filename, message string) (*Node, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	content := string(b)
	return AddSnippet(entry, kind, sessionID, agent, version, content, message)
}

func AddSnippet(entry *config.Entry, kind Kind, sessionID, agent, version, content, description string) (*Node, error) {
	store := NewStore(entry)
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if len(description) == 0 {
		description = tools.TrimLength(content, maxSummaryLen)
	}
	event, err := NewNode(kind, description, &TextDetails{Text: content}, NewMetadata(sessionID, agent, version))
	if err != nil {
		return nil, err
	}
	event, err = store.AddNode(event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func GetSnippet(entry *config.Entry, eventID string) (*Node, error) {
	store := NewStore(entry)
	return store.GetNode(eventID)
}

func WriteSnippetDetails(event *Node, filename string, overwrite bool) error {
	exists, err := tools.StatExists(filename)
	if err != nil {
		return err
	}
	if exists && !overwrite {
		return fmt.Errorf("cannot overwrite file = %v", filename)
	}
	fw, err := tools.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer tools.CloseFile(fw)
	if _, err := event.WriteBody(fw); err != nil {
		return err
	}
	return nil
}
