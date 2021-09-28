package markdown

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/inspector"
	"github.com/aardlabs/terminal-poc/internal/markdown"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"io/ioutil"
	"path/filepath"
	"time"
)

// MDFileInspect executes the provided ma rkdown file via the inspector REPL
func MDFileInspect(mdFile string) error {
	store, err := NewMDFileStore(mdFile)
	if err != nil {
		return err
	}
	nodeID, err := store.ExtractID(mdFile)
	if err != nil {
		return err
	}
	graphCtx := snippet.Context{
		ConfigEntry: &config.Entry{
			Name:             "Unknown",
			LastUpdateCheck:  time.Now(),
			Email:            "unknown@bar.com",
			ExecutionTimeout: tools.MarshalledDuration{Duration: 30 * time.Second},
			HideInspectIntro: false,
		},
		Metadata: nil,
	}
	graphCtx.SetStore(store)
	return inspector.InspectNode(&graphCtx, nodeID)
}

func CreateNodeFromMarkdownFile(mdFile string) (*graph.Node, error) {
	mdContent, err := ioutil.ReadFile(mdFile)
	if err != nil {
		return nil, fmt.Errorf("readfile %v %w", mdFile, err)
	}

	id := filepath.Base(mdFile)
	return CreateNodeFromMarkdown(id, mdFile, string(mdContent))
}

func CreateNodeFromMarkdown(id, sourceURI, mdContent string) (*graph.Node, error) {
	now := time.Now().UTC()
	i := 0
	blocks := make([]*graph.Block, 0)
	title, err := markdown.Split(mdContent, func(chunk string, chunkType markdown.ChunkType, language string) error {
		i++
		contentType, err := executor.Parse(language)
		if err != nil {
			return fmt.Errorf("executor.Parse language = %v %w", language, err)
		}
		b := graph.Block{
			ID:                fmt.Sprintf("%s/%d", id, i),
			CreatedAt:         &now,
			Content:           chunk,
			ContentType:       contentType,
			MD5:               createMD5Hash(chunk),
			LastExecutedAt:    nil,
			LastExecutedBy:    "",
			LastExitStatus:    "",
			LastExecutionInfo: "",
		}
		blocks = append(blocks, &b)
		tools.Log.Info().
			Str("BlockID", b.ID).
			Str("Lang", language).
			Str("IsCode", fmt.Sprintf("%v", b.IsCode())).
			Str("ContentType", contentType.String()).
			Str("MD5", b.MD5).
			Msg("block created")
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("markdown.split err = %w", err)
	}

	return &graph.Node{
		ID:         id,
		Title:      title,
		CreatedAt:  &now,
		OccurredAt: &now,
		Metadata:   graph.Metadata{SourceURI: sourceURI},
		Markdown:   mdContent,
		ChildNodes: []*graph.Node{},
		Blocks:     blocks,
	}, nil
}

func createMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
