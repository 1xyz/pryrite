package events

import (
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"os"
	"strings"
	"time"
)

const clientTimeout = 10 * time.Second

type Details interface {
	// Encode this to rawMessage
	EncodeJSON() ([]byte, error)

	// The body representation of this
	Summary() string
}

type TextDetails struct {
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
	Text  string `json:"text,omitempty"`
}

func (t *TextDetails) Summary() string {
	if len(t.Title) > 0 {
		return t.Title
	} else if len(t.Text) > 0 {
		return t.Text
	} else if len(t.Url) > 0 {
		return t.Url
	} else {
		return ""
	}
}

func (t *TextDetails) EncodeJSON() ([]byte, error) {
	return json.Marshal(t)
}

type Metadata struct {
	SessionID string `json:"SessionID"`
	Agent     string `json:"Agent"`
}

func NewMetadata(sessionID, agent string) *Metadata {
	ppId := os.Getppid()
	return &Metadata{
		SessionID: fmt.Sprintf("%s:%d", sessionID, ppId),
		Agent:     agent,
	}
}

type User struct {
	Email    string `json:"Email"`
	Username string `json:"Username"`
	Color    string `json:"Color"`
}

type Node struct {
	ID                    int64           `json:"ID,omitempty"`
	CreatedAt             time.Time       `json:"CreatedAt"`
	OccurredAt            time.Time       `json:"OccurredAt,omitempty"`
	TimestampMilliseconds int64           `json:"timestampMilliseconds,omitempty"`
	Kind                  Kind            `json:"Kind"`
	Metadata              Metadata        `json:"Metadata"`
	Description           string          `json:"description,omitempty"`
	Details               json.RawMessage `json:"Details,omitempty"`
	UserID                int64           `json:"UserID,omitempty"`
	User                  User            `json:"User,omitempty"`
}

func New(kind Kind, description string, details Details, metadata *Metadata) (*Node, error) {
	d, err := details.EncodeJSON()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	fmt.Printf("now ms %v\n", now.Unix())
	return &Node{
		CreatedAt:             now,
		OccurredAt:            now,
		TimestampMilliseconds: now.UnixNano() / int64(time.Millisecond),
		Kind:                  kind,
		Details:               d,
		Description:           description,
		Metadata:              *metadata,
	}, nil
}

func (e *Node) EncodeDetails(d Details) error {
	switch e.Kind {
	case "Command":
		b, err := d.EncodeJSON()
		if err != nil {
			return err
		}
		e.Details = b
	}
	return nil
}

func (e *Node) DecodeDetails() (Details, error) {
	switch e.Kind {
	case Command, AsciiCast, PageClose, PageOpen, TextSelect:
		raw := TextDetails{}
		if err := json.Unmarshal(e.Details, &raw); err != nil {
			return nil, fmt.Errorf("unmarshall ConsoleEvent %v", err)
		}
		return &raw, nil
	default:
		return &TextDetails{Text: string(e.Details)}, nil
	}
}

func (e *Node) WriteBody(w io.Writer) (int, error) {
	d, err := e.DecodeDetails()
	if err != nil {
		return 0, err
	}
	body := []byte(d.Summary())
	return w.Write(body)
}

//

func AddConsoleEvent(entry *config.Entry, sessionID, agent, content, description string, doRender bool) (*Node, error) {
	return AddEvent(entry, Command, sessionID, agent, content, description, doRender)
}

func AddEventFromFile(entry *config.Entry, kind Kind, sessionID, agent, filename, message string, doRender bool) (*Node, error) {
	// ToDo: *maybe* a better way
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	content := string(b)
	return AddEvent(entry, kind, sessionID, agent,
		content, message, doRender)
}

func AddEvent(entry *config.Entry, kind Kind, sessionID, agent, content, description string, doRender bool) (*Node, error) {
	store := NewStore(entry)
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if len(description) == 0 {
		description = tools.TrimLength(content, maxColumnLen)
	}

	event, err := New(kind, description, &TextDetails{Text: content}, NewMetadata(sessionID, agent))
	if err != nil {
		return nil, err
	}
	event, err = store.AddEvent(event)
	if err != nil {
		return nil, err
	}

	if doRender {
		evtRender := &eventRender{E: event, renderDetail: false}
		evtRender.Render()
	}
	return event, nil
}

func GetEvent(entry *config.Entry, eventID string) (*Node, error) {
	store := NewStore(entry)
	return store.GetEvent(eventID)
}

func WriteEventDetailsToFile(event *Node, filename string, overwrite bool) error {
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
