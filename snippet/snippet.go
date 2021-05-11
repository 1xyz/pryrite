package snippet

import (
	"errors"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"net/http"
	"net/url"
	"strings"
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
	id, err := getID(id)
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

func getID(idOrURL string) (string, error) {
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
