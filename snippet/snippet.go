package snippet

import (
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/tools"
)

func NewContext(cfg *config.Config, agent string) *Context {
	entry, found := cfg.GetDefaultEntry()
	if !found {
		panic("default config is nil")
	}
	return &Context{
		ConfigEntry: entry,
		Metadata: &graph.Metadata{
			Agent: agent,
		},
	}
}

type Context struct {
	ConfigEntry *config.Entry
	Metadata    *graph.Metadata
	store       graph.Store
}

func (c *Context) SetStore(store graph.Store) { c.store = store }
func (c *Context) GetStore() (graph.Store, error) {
	if c.store == nil {
		s, err := NewStoreFromContext(c)
		if err != nil {
			return nil, err
		}
		c.store = s
	}
	return c.store, nil
}

func GetSnippetNode(ctx *Context, id string) (*graph.Node, error) {
	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	nodeID, err := store.ExtractID(id)
	if err != nil {
		return nil, err
	}

	n, err := store.GetNode(nodeID)
	if err != nil {
		ctxMsg := fmt.Sprintf("GetSnippetNode(%s) = %v", id, err)
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return nil, handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return nil, err
	}

	return n, nil
}

func GetSnippetNodeWithChildren(ctx *Context, id string) (*graph.Node, error) {
	n, err := GetSnippetNode(ctx, id)
	if err != nil {
		return nil, err
	}
	s, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	if err := n.LoadChildNodes(s, false /*force */); err != nil {
		return nil, err
	}
	return n, nil
}

func UpdateSnippetNode(ctx *Context, n *graph.Node) error {
	store, err := ctx.GetStore()
	if err != nil {
		return err
	}
	err = store.UpdateNode(n)
	if err != nil {
		ctxMsg := fmt.Sprintf("UpdateSnippetNode(%s) = %v", n.ID, err)
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return err
	}

	return nil
}

func UpdateNodeBlock(ctx *Context, n *graph.Node, b *graph.Block) error {
	store, err := ctx.GetStore()
	if err != nil {
		return err
	}
	err = store.UpdateNodeBlock(n, b)
	if err != nil {
		ctxMsg := fmt.Sprintf("UpdateNodeBlock(%s) = %v", n.ID, err)
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return err
	}

	return nil
}

type ExecutionInfo struct {
	ExecutedAt    *time.Time
	ExecutedBy    string
	ExitStatus    string
	ExecutionInfo string
}

func UpdateNodeBlockExecution(ctx *Context, n *graph.Node, b *graph.Block, e *ExecutionInfo) error {
	store, err := ctx.GetStore()
	if err != nil {
		return err
	}

	b.LastExecutedAt = e.ExecutedAt
	b.LastExecutionInfo = e.ExecutionInfo
	b.LastExitStatus = e.ExitStatus
	b.LastExecutedBy = e.ExecutedBy

	tools.Log.Info().Msgf("UpdateNodeBlockExecution node %s block %s", n.ID, b.ID)
	err = store.UpdateNodeBlockExecution(n, b)
	if err != nil {
		ctxMsg := fmt.Sprintf("UpdateNodeBlockExecution(%s) = %v", n.ID, err)
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return err
	}

	return nil
}

func SearchSnippetNodes(ctx *Context, query string, limit int, kind graph.Kind) ([]graph.Node, error) {
	ctxMsg := fmt.Sprintf("SearchSnippetNodes query=%s(limit=%d, kind=%v)",
		query, limit, kind)

	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	n, err := store.SearchNodes(query, limit, kind)
	if err != nil {
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return nil, handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return nil, err
	}

	tools.Log.Info().Msgf("%s, %d nodes found", ctxMsg, len(n))
	return n, nil
}

func GetSnippetNodes(ctx *Context, limit int, kind graph.Kind) ([]graph.Node, error) {
	ctxMsg := fmt.Sprintf("GetSnippetNodes(limit=%d, kind=%v)", limit, kind)

	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	n, err := store.GetNodes(limit, kind)
	if err != nil {
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return nil, handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return nil, err
	}

	tools.Log.Info().Msgf("%s, %d nodes found", ctxMsg, len(n))
	return n, nil
}

func AddSnippetNode(ctx *Context, title, content string, contentType string) (*graph.Node, error) {
	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	n, err := graph.NewNode(graph.Command, title, content, contentType, *ctx.Metadata)
	if err != nil {
		return nil, err
	}

	result, err := store.AddNode(n)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func EditSnippetNode(ctx *Context, id string, save bool) (*graph.Node, error) {
	n, err := GetSnippetNode(ctx, id)
	if err != nil {
		return nil, err
	}

	// create a temporary file - current we do not know the type.
	// but hopefully in the future we can guess this.
	filename, err := tools.CreateTempFile("", "t_*.txt")
	if err != nil {
		return nil, err
	}

	// Write the content to temporary file
	if len(n.Blocks) > 0 {
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}

		bytes := 0
		for _, block := range n.Blocks {
			n, err := file.WriteString(block.Content)
			if err != nil {
				file.Close()
				return nil, err
			}

			bytes += n
		}

		file.Close()
		tools.Log.Info().Msgf("EditSnippetNode (%s). wrote %d bytes to %s",
			n.ID, bytes, filename)
	} else {
		tools.Log.Info().Msgf("EditSnippetNode (%s). no content to write to file", n.ID)
	}

	// Compute the file hash before
	h0, err := computeFileHash("sha256", filename)
	if err != nil {
		return nil, err
	}

	// open this file in editor
	if err := openFileInEditor(filename); err != nil {
		return nil, err
	}

	// Compute the filehash after edit
	h1, err := computeFileHash("sha256", filename)
	if err != nil {
		return nil, err
	}

	// check to see if the content should be sent
	if areEqual(h0, h1) {
		tools.LogStdout("no changes are detected. abandoning edit")
		return n, nil
	}

	newContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	n.Blocks = nil
	n.Markdown = string(newContent)

	if save {
		tools.Log.Info().Msgf("EditSnippetNode: Save %s to remote service.", n.ID)
		if err := UpdateSnippetNode(ctx, n); err != nil {
			return nil, err
		}
	}

	return n, nil
}

func EditNodeBlock(ctx *Context, n *graph.Node, b *graph.Block, save bool) (*graph.Block, error) {
	// Write content to a temporary file
	filename, err := tools.CreateTempFile("", "block_*.txt")
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filename, []byte(b.Content), 0600); err != nil {
		tools.Log.Err(err).Msgf("EditNodeBlock: [%s:%s] WriteFile failed", n.ID, b.ID)
		return nil, err
	}

	// Compute the file hash before
	h0, err := computeFileHash("md5", filename)
	if err != nil {
		return nil, err
	}

	// open this file in editor
	if err := openFileInEditor(filename); err != nil {
		return nil, err
	}

	// Compute the filehash after edit
	h1, err := computeFileHash("md5", filename)
	if err != nil {
		return nil, err
	}

	// check to see if the content should be sent
	if areEqual(h0, h1) {
		tools.LogStdout("no changes are detected. abandoning edit")
		return b, nil
	}

	// save the old content and hash in case save fails!
	oldContent := b.Content
	oldMD5 := b.MD5

	// update the block with the new content and hash
	newContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	b.Content = string(newContent)
	b.MD5 = hexString(h1)

	if save {
		tools.Log.Info().Msgf("EditSnippetNode: Save %s to remote service.", n.ID)
		if err := UpdateNodeBlock(ctx, n, b); err != nil {
			// restore the old content and hash
			b.Content = oldContent
			b.MD5 = oldMD5
			return nil, err
		}
	}
	return b, err
}

func handleGraphHTTPErr(ghe *graph.HttpError, ctxMessage string) error {
	tools.Log.Err(ghe).Msgf("%s HTTP error = %v reason = %s",
		ctxMessage, ghe.HTTPCode, ghe.Error())
	if ghe.HTTPCode == http.StatusUnauthorized {
		return ghe
	}
	if ghe.HTTPCode == http.StatusNotFound {
		return fmt.Errorf("snippet not found")
	}
	return fmt.Errorf("%s code = %v", ctxMessage, ghe.HTTPCode)
}

func NewStoreFromContext(ctx *Context) (graph.Store, error) {
	store := graph.NewStore(ctx.ConfigEntry, ctx.Metadata)
	return store, nil
}

func openFileInEditor(filename string) error {
	// ToDo: have a better solution to support windows etc
	const DefaultEditor = "nano"
	editor := os.Getenv("EDITOR")
	if editor == "" {
		tools.LogStdout(fmt.Sprintf("No EDITOR variable defined, using %s", DefaultEditor))
		editor = DefaultEditor
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("could not find path for the editor %v", err)
	}
	tools.Log.Info().Msgf("found editor=%s path=%s", editor, executable)

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func computeFileHash(algo, filename string) ([]byte, error) {
	fr, err := tools.OpenFile(filename, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer tools.CloseFile(fr)

	var h hash.Hash = nil
	if algo == "sha256" {
		h = sha256.New()
	} else if algo == "md5" {
		h = md5.New()
	} else {
		return nil, fmt.Errorf("unsupport slgo %s", algo)
	}

	if _, err := io.Copy(h, fr); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func areEqual(b0, b1 []byte) bool {
	if len(b0) != len(b1) {
		return false
	}
	for i := 0; i < len(b0); i++ {
		if b0[i] != b1[i] {
			return false
		}
	}
	return true
}

func hexString(hash []byte) string { return fmt.Sprintf("%x", hash) }
