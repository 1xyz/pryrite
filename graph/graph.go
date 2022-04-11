package graph

import (
	"fmt"
	"strings"
	"time"

	executor "github.com/1xyz/pryrite/executors"
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
	ID             string     `json:"id,omitempty"`
	CreatedAt      *time.Time `json:"created_at"`
	OccurredAt     *time.Time `json:"occurred_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	Kind           Kind       `json:"kind"`
	Metadata       Metadata   `json:"metadata"`
	Title          string     `json:"title,omitempty"`
	Markdown       string     `json:"markdown,omitempty"`
	View           string     `json:"view"`
	Blocks         []*Block   `json:"blocks,omitempty"`
	Children       string     `json:"children,omitempty"`
	IsShared       bool       `json:"is_shared"`
	LastExecutedAt *time.Time `json:"last_executed_at"`
	LastExecutedBy string     `json:"last_executed_by"`

	// ChildNodes is the actual reference to the node(s)
	// Note: currently, this relation is not persisted in the remote store
	//       so skip this from json encoding..
	ChildNodes []*Node `json:"-"`
}

func (n *Node) GetBlock(blockID string) (*Block, bool) {
	if n.Blocks == nil {
		return nil, false
	}
	for _, b := range n.Blocks {
		if b.ID == blockID {
			return b, true
		}
	}
	return nil, false
}

type Block struct {
	ID          string                `json:"id,omitempty"`
	CreatedAt   *time.Time            `json:"created_at"`
	Content     string                `json:"content"`
	ContentType *executor.ContentType `json:"content_type"`
	MD5         string                `json:"md5"`

	LastExecutedAt    *time.Time `json:"last_executed_at,omitempty"`
	LastExecutedBy    string     `json:"last_executed_by,omitempty"`
	LastExitStatus    string     `json:"last_exit_status,omitempty"`
	LastExecutionInfo string     `json:"last_execution_info,omitempty"`
}

func (block *Block) IsCode() bool {
	return block.ContentType != nil &&
		block.ContentType.Type == "text" && block.ContentType.Subtype != "markdown"
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

func (n *Node) LoadChildNodes(store Store, force bool) error {
	if n.ChildNodes != nil && !force {
		// the childNodes are already loaded and we are not forcing it
		return nil
	}
	children, err := store.GetChildren(n.ID)
	if err != nil {
		return err
	}
	n.ChildNodes = make([]*Node, len(children))
	for i := range children {
		n.ChildNodes[i] = &children[i]
	}
	return nil
}

func (n *Node) HasBlocks() bool {
	return n.Blocks != nil && len(n.Blocks) > 0
}

func (n *Node) HasChildren() bool {
	return len(n.Children) > 0
}

func (n Node) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("ID=%s kind=%v Title=%s MD=%s Children=%s",
		n.ID, n.Kind, n.Title, n.Markdown, n.Children))
	return sb.String()
}

func NewNode(kind Kind, title, content, contentType string, metadata Metadata) (*Node, error) {
	now := time.Now().UTC()
	var language string
	vals := strings.Split(contentType, "/")
	if len(vals) > 1 {
		language = vals[1]
	}
	var markdown string
	if len(title) == 0 {
		markdown = fmt.Sprintf("```%s\n%s\n```\n", language, content)
	} else {
		markdown = fmt.Sprintf("# %s\n```%s\n%s\n```\n", title, language, content)
	}

	return &Node{
		Title:      title,
		OccurredAt: &now,
		Kind:       kind,
		Markdown:   markdown,
		Metadata:   metadata,
		ChildNodes: nil,
	}, nil
}

func NewNodeFromBlocks(kind Kind, blocks []*Block, metadata Metadata) (*Node, error) {
	now := time.Now().UTC()
	return &Node{
		OccurredAt: &now,
		Kind:       kind,
		Metadata:   metadata,
		Blocks:     blocks,
	}, nil
}
