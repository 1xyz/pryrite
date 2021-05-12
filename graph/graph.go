package graph

import (
	"fmt"
	"time"
)

type Metadata struct {
	SourceTitle string `json:"source_title"`
	SourceURI   string `json:"source_uri"`
	Agent       string `json:"Agent"`
}

func NewMetadata(agent, version string) *Metadata {
	return &Metadata{
		Agent: fmt.Sprintf("%s:%s", agent, version),
	}
}

type Node struct {
	ID               string    `json:"id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	OccurredAt       time.Time `json:"occurred_at,omitempty"`
	DeletedAt        time.Time `json:"deleted_at,omitempty"`
	Kind             Kind      `json:"Kind"`
	Metadata         Metadata  `json:"Metadata"`
	Title            string    `json:"title,omitempty"`
	IsTitleGenerated bool      `json:"title_was_generated"`
	Description      string    `json:"description,omitempty"`
	Content          string    `json:"content,omitempty"`
	ContentLanguage  string    `json:"content_language,omitempty"`
	Children         string    `json:"children,omitempty"`
}

func NewNode(kind Kind, title, description, content string, metadata Metadata) (*Node, error) {
	now := time.Now().UTC()
	return &Node{
		CreatedAt:   now,
		OccurredAt:  now,
		Kind:        kind,
		Title:       title,
		Description: description,
		Content:     content,
		Metadata:    metadata,
	}, nil
}
