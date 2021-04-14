package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

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
	if len(t.Text) > 0 {
		return t.Text
	} else if len(t.Url) > 0 {
		return t.Url
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

func NewMetadata(clientID, agent, version string) *Metadata {
	ppId := os.Getppid()
	return &Metadata{
		SessionID: fmt.Sprintf("%s:%d", clientID, ppId),
		Agent:     fmt.Sprintf("%s:%s", agent, version),
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

func NewNode(kind Kind, description string, details Details, metadata *Metadata) (*Node, error) {
	d, err := details.EncodeJSON()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
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
			return nil, err
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
