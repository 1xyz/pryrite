package snippet

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

func NewContext(c *config.Config, agent string) *Context {
	return &Context{
		Config: c,
		Metadata: &graph.Metadata{
			Agent: agent,
		},
	}
}

type Context struct {
	Config   *config.Config
	Metadata *graph.Metadata
}

func GetSnippetNode(ctx *Context, id string) (*graph.Node, error) {
	id, err := GetID(id)
	if err != nil {
		return nil, err
	}
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return nil, fmt.Errorf("default config is nil")
	}

	store := graph.NewStore(entry, ctx.Metadata)
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
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return fmt.Errorf("default config is nil")
	}

	store := graph.NewStore(entry, ctx.Metadata)
	err := store.UpdateNode(n)
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

func SearchSnippetNodes(ctx *Context, query string, limit int, kind graph.Kind) ([]graph.Node, error) {
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return nil, fmt.Errorf("default config is nil")
	}

	ctxMsg := fmt.Sprintf("SearchSnippetNodes query=%s(limit=%d, kind=%v)", query, limit, kind)
	store := graph.NewStore(entry, ctx.Metadata)
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
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return nil, fmt.Errorf("default config is nil")
	}

	ctxMsg := fmt.Sprintf("GetSnippetNodes(limit=%d, kind=%v)", limit, kind)
	store := graph.NewStore(entry, ctx.Metadata)
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

func AddSnippetNode(ctx *Context, content string) (*graph.Node, error) {
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return nil, fmt.Errorf("default config is nil")
	}

	store := graph.NewStore(entry, ctx.Metadata)
	n, err := graph.NewNode(graph.Command, "", "", content, *ctx.Metadata)
	if err != nil {
		return nil, err
	}

	result, err := store.AddNode(n)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func EditSnippetNode(ctx *Context, id string) (*graph.Node, error) {
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
	if len(n.Content) > 0 {
		if err := ioutil.WriteFile(filename, []byte(n.Content), 0600); err != nil {
			return nil, err
		}
		tools.Log.Info().Msgf("EditSnippetNode (%s). wrote n=%d bytes to %s",
			n.ID, len(n.Content), filename)
	} else {
		tools.Log.Info().Msgf("EditSnippetNode (%s). no content to write to file",
			n.ID, filename)
	}

	// Compute the file hash before
	h0, err := computeFileHash(filename)
	if err != nil {
		return nil, err
	}

	// open this file in editor
	if err := openFileInEditor(filename); err != nil {
		return nil, err
	}

	// Compute the filehash after edit
	h1, err := computeFileHash(filename)
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
	n.Content = string(newContent)

	if err := UpdateSnippetNode(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func GetSnippetNodeViewWithChildren(ctx *Context, id string) (*graph.NodeView, error) {
	id, err := GetID(id)
	if err != nil {
		return nil, err
	}
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return nil, fmt.Errorf("default config is nil")
	}

	store := graph.NewStore(entry, ctx.Metadata)
	nv, err := GetSnippetNodeView(store, id)
	if err != nil {
		return nil, err
	}

	nv.Children = []*graph.NodeView{}
	childIDs := nv.Node.GetChildIDs()
	for _, childID := range childIDs {
		cnv, err := GetSnippetNodeView(store, childID)
		if err != nil {
			return nil, fmt.Errorf("error fetching childId = %v for node %v. err = %v",
				childID, id, err)
		}
		nv.Children = append(nv.Children, cnv)
	}

	return nv, nil
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
		return fmt.Errorf("not authorized to get snippet")
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
	entry, found := ctx.Config.GetDefaultEntry()
	if !found {
		return nil, fmt.Errorf("default config is nil")
	}

	store := graph.NewStore(entry, ctx.Metadata)
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

func computeFileHash(filename string) ([]byte, error) {
	fr, err := tools.OpenFile(filename, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer tools.CloseFile(fr)

	h := sha256.New()
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
