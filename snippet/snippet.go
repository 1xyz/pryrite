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
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
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
	id, err := GetID(id)
	if err != nil {
		return nil, err
	}

	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}
	n, err := store.GetNode(id)
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

func AddSnippetNode(ctx *Context, content string, contentType string) (*graph.Node, error) {
	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	n, err := graph.NewNode(graph.Command, content, contentType, *ctx.Metadata)
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
		if len(n.Blocks) > 1 {
			tools.Log.Error().Msgf("EditSnippetNode (%s). TODO: handle more than one snippet", n.ID)
		}
		if err := ioutil.WriteFile(filename, []byte(n.Blocks[0].Content), 0600); err != nil {
			return nil, err
		}
		tools.Log.Info().Msgf("EditSnippetNode (%s). wrote n=%d bytes to %s",
			n.ID, len(n.Blocks), filename)
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
	n.Blocks[0].Content = string(newContent)

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

func GetSnippetNodeViewWithChildren(ctx *Context, id string) (*graph.NodeView, error) {
	id, err := GetID(id)
	if err != nil {
		return nil, err
	}

	store, err := ctx.GetStore()
	if err != nil {
		return nil, err
	}

	nv, err := GetSnippetNodeView(store, id)
	if err != nil {
		return nil, err
	}

	nv.Children = []*graph.NodeView{}
	childNodes, err := store.GetChildren(id)
	if err != nil {
		return nil, err
	}
	for i := range childNodes {
		nv.Children = append(nv.Children, &graph.NodeView{Node: &childNodes[i], View: childNodes[i].View})
	}
	return nv, nil
}

func GetNodeViewURL(ctx *Context, node *graph.Node) *url.URL {
	u, err := url.Parse(ctx.ConfigEntry.ServiceUrl)
	if err != nil {
		return nil
	}

	u.Path = "nodes/" + node.ID
	return u
}

func GetSnippetNodeView(store graph.Store, id string) (*graph.NodeView, error) {
	startAt := time.Now()
	n, err := store.GetNodeView(id)
	if err != nil {
		ctxMsg := fmt.Sprintf("GetSnippetNodeView(%s) = %v", id, err)
		var ghe *graph.HttpError
		if errors.As(err, &ghe) {
			return nil, handleGraphHTTPErr(ghe, ctxMsg)
		}
		tools.Log.Err(err).Msg(ctxMsg)
		return nil, err
	}
	duration := time.Since(startAt)
	tools.Log.Info().Msgf("GetSnippetNodeView: took %v", duration)
	return n, nil
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

func GetID(idOrURL string) (string, error) {
	idOrURL = strings.TrimSpace(idOrURL)
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
