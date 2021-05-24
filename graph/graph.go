package graph

import (
	"fmt"
	"strings"
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
	ID               string     `json:"id,omitempty"`
	CreatedAt        *time.Time `json:"created_at"`
	OccurredAt       *time.Time `json:"occurred_at,omitempty"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	Kind             Kind       `json:"Kind"`
	Metadata         Metadata   `json:"Metadata"`
	Title            string     `json:"title,omitempty"`
	IsTitleGenerated bool       `json:"title_was_generated"`
	Description      string     `json:"description,omitempty"`
	Content          string     `json:"content,omitempty"`
	ContentLanguage  string     `json:"content_language,omitempty"`
	Children         string     `json:"children,omitempty"`
}

func (n *Node) GetChildIDs() []string {
	ids := strings.Split(n.Children, ",")
	result := []string{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if len(id) > 0 {
			result = append(result, id)
		}
	}
	return result
}

func (n Node) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("ID=%s kind=%v Title=%s Desc=%s Children=%s",
		n.ID, n.Kind, n.Title, n.Description, n.Children))
	return sb.String()
}

func NewNode(kind Kind, title, description, content string, metadata Metadata) (*Node, error) {
	now := time.Now().UTC()
	return &Node{
		CreatedAt:   &now,
		OccurredAt:  &now,
		Kind:        kind,
		Title:       title,
		Description: description,
		Content:     content,
		Metadata:    metadata,
	}, nil
}

// NodeView is a terminal specific rendered response
type NodeView struct {
	Node            *Node  `json:"node"`
	ContentMarkdown string `json:"content_markdown"`
	Children        []*NodeView
}

func (n NodeView) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Node: %v\n", n.Node))
	if n.Children != nil && len(n.Children) > 0 {
		for i, c := range n.Children {
			sb.WriteString(fmt.Sprintf("Child[%d] = %v\n", i, c))
		}
	}
	return sb.String()
}
