package markdown

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/1xyz/pryrite/config"
	executor "github.com/1xyz/pryrite/executors"
	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/inspector"
	"github.com/1xyz/pryrite/internal/markdown"
	"github.com/1xyz/pryrite/snippet"
	"github.com/1xyz/pryrite/tools"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// MDFileInspect executes the provided ma rkdown file via the inspector REPL
func MDFileInspect(mdFile string) error {
	file, err := fetchFile(mdFile)
	if err != nil {
		return err
	}
	nodeID, err := ExtractIDFromFilePath(mdFile)
	if err != nil {
		return err
	}
	store, err := NewMDFileStore(nodeID, file)
	if err != nil {
		return err
	}

	cfg, err := config.Default()
	if err != nil {
		return err
	}

	entry, ok := cfg.GetDefaultEntry()
	if !ok {
		return fmt.Errorf("default not found")
	}

	graphCtx := snippet.Context{
		ConfigEntry: entry,
		Metadata:    nil,
	}
	graphCtx.SetStore(store)
	return inspector.InspectNode(&graphCtx, nodeID)
}

func CreateNodeFromMarkdownFile(id, mdFile string) (*graph.Node, error) {
	mdContent, err := ioutil.ReadFile(mdFile)
	if err != nil {
		return nil, fmt.Errorf("readfile %v %w", mdFile, err)
	}

	if id == "" {
		id, err = ExtractIDFromFilePath(mdFile)
		if err != nil {
			return nil, fmt.Errorf("extractID %v", err)
		}
	}
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

func fetchFile(filename string) (string, error) {
	u, err := url.Parse(filename)
	if err != nil {
		return "", fmt.Errorf("url.Parser %w", err)
	}
	switch u.Scheme {
	case "file", "":
		return filename, nil
	case "http", "https":
		return httpDownloadFile(filename)
	default:
		return "", fmt.Errorf("unsupported scheme %v", u.Scheme)
	}
}

func httpDownloadFile(url string) (string, error) {
	tools.LogStdout("fetching from %v\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "resp.body.close %v", err)
		}
	}()

	filename, err := tools.CreateTempFile("", "t_*.md")
	if err != nil {
		return "", err
	}
	tools.LogStdout("fetching from %v to %v\n", url, filename)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(filename, b, 0600); err != nil {
		return "", err
	}
	return filename, nil
}

func ExtractIDFromFilePath(filepath string) (string, error) {
	idOrURL := strings.TrimSpace(filepath)
	u, err := url.Parse(idOrURL)
	if err != nil {
		return "", err
	}
	tokens := strings.Split(u.Path, "/")
	if len(tokens) == 0 || len(idOrURL) == 0 {
		return "", fmt.Errorf("empty id")
	}
	return tokens[len(tokens)-1], nil
}
