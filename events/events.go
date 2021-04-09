package events

import (
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/google/uuid"
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
	Body() string
}

type RawDetails struct {
	Raw string `json:"raw"`
}

func (r *RawDetails) Body() string                { return r.Raw }
func (r *RawDetails) EncodeJSON() ([]byte, error) { return json.Marshal(&RawDetails{Raw: r.Raw}) }

type Metadata struct {
	SessionID string `json:"SessionID"`
	Title     string `json:"Title"`
	URL       string `json:"URL"`
}

type Event struct {
	ID        int64           `json:"ID"`
	CreatedAt time.Time       `json:"CreatedAt"`
	Kind      Kind            `json:"Kind"`
	Details   json.RawMessage `json:"Details,omitempty"`
	Metadata  Metadata        `json:"Metadata"`
}

func New(kind Kind, title, url string, details Details) (*Event, error) {
	d, err := details.EncodeJSON()
	if err != nil {
		return nil, err
	}
	return &Event{
		CreatedAt: time.Now().UTC(),
		Kind:      kind,
		Details:   d,
		Metadata: Metadata{
			SessionID: uuid.New().String(),
			Title:     tools.TrimLength(title, maxColumnLen),
			URL:       url,
		},
	}, nil
}

func (e *Event) EncodeDetails(d Details) error {
	switch e.Kind {
	case "Console":
		b, err := d.EncodeJSON()
		if err != nil {
			return err
		}
		e.Details = b
	}
	return nil
}

func (e *Event) DecodeDetails() (Details, error) {
	switch e.Kind {
	case Console, AsciiCast:
		raw := RawDetails{}
		if err := json.Unmarshal(e.Details, &raw); err != nil {
			return nil, fmt.Errorf("unmarshall ConsoleEvent %v", err)
		}
		return &raw, nil
	case PageClose, PageOpen:
		return &RawDetails{Raw: e.Metadata.URL}, nil
	default:
		return &RawDetails{Raw: string(e.Details)}, nil
	}
}

func (e *Event) WriteBody(w io.Writer) (int, error) {
	d, err := e.DecodeDetails()
	if err != nil {
		return 0, err
	}
	body := []byte(d.Body())
	return w.Write(body)
}

//

func AddConsoleEvent(entry *config.Entry, content, message string, doRender bool) (*Event, error) {
	return AddEvent(entry, Console, content, message, doRender)
}

func AddEventFromFile(entry *config.Entry, kind Kind, filename, message string, doRender bool) (*Event, error) {
	// ToDo: *maybe* a better way
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	content := string(b)
	return AddEvent(entry, kind, content, message, doRender)
}

func AddEvent(entry *config.Entry, kind Kind, content, message string, doRender bool) (*Event, error) {
	store := NewStore(entry)
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if message == "None" {
		message = tools.TrimLength(content, maxColumnLen)
	}

	event, err := New(kind, message, "", &RawDetails{Raw: content})
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

func GetEvent(entry *config.Entry, eventID string) (*Event, error) {
	store := NewStore(entry)
	return store.GetEvent(eventID)
}

func WriteEventDetailsToFile(event *Event, filename string, overwrite bool) error {
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
