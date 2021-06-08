package graph

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/go-resty/resty/v2"
)

type Store interface {
	// GetNodes returns the most recent n entries
	GetNodes(limit int, kind Kind) ([]Node, error)

	// GetNode returns the Node associated with this id
	GetNode(id string) (*Node, error)

	// AddNode adds a new node to this store
	AddNode(*Node) (*Node, error)

	// SearchNodes searches for nodes for provided query
	SearchNodes(query string, limit int, kind Kind) ([]Node, error)

	// UpdateNode updates the content of an existing node
	UpdateNode(node *Node) error

	// GetNodeView asks the server for a terminal render-able view.
	GetNodeView(id string) (*NodeView, error)

	// UpdateNodeBlock updates the block of an existing node at the server
	UpdateNodeBlock(node *Node, block *Block) error
}

type ErrorResponse struct {
	OccurredAt time.Time `json:"occurred_at"`
	Status     int       `json:"status"`
	Error      string    `json:"error"`
	Message    string    `json:"message"`
	Path       string    `json:"path"`
}

// remoteStore represents the remote event store backed by the service
type remoteStore struct {
	configEntry *config.Entry
	m           *Metadata
}

func NewStore(configEntry *config.Entry, metadata *Metadata) Store {
	return &remoteStore{
		configEntry: configEntry,
		m:           metadata,
	}
}

type getNodesResponse struct {
	N []Node `json:"Nodes"`
}

func (r *remoteStore) GetNodes(limit int, kind Kind) ([]Node, error) {
	client := r.newHTTPClient(false)
	kindStr := string(kind)
	if kind == Unknown {
		kindStr = ""
	}
	req := client.R().
		SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{
			"limit":   strconv.Itoa(limit),
			"kind":    kindStr,
			"include": "blocks",
		})
	resp, err := req.Get("/api/v1/nodes")
	if err != nil {
		return nil, err
	}
	if err := checkHTTP2XX("getNodes(%s)", resp); err != nil {
		return nil, err
	}
	result := getNodesResponse{N: []Node{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, &Error{"GetNodes", err}
	}
	return result.N, nil
}

func (r *remoteStore) GetNode(id string) (*Node, error) {
	client := r.newHTTPClient(false)
	req := client.R().
		SetPathParam("nodeId", id).
		SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{
			"include": "blocks",
			"view":    "text",
		})
	resp, err := req.Get("/api/v1/nodes/{nodeId}")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v", err)
	}
	if err := checkHTTP2XX(fmt.Sprintf("getNode(%s)", id), resp); err != nil {
		return nil, err
	}
	result := Node{}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, &Error{"GetNode", err}
	}
	return &result, nil
}

func (r *remoteStore) GetNodeView(id string) (*NodeView, error) {
	node, err := r.GetNode(id)
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v", err)
	}
	result := NodeView{
		Node: node,
		View: node.View,
	}
	return &result, nil
}

func (r *remoteStore) AddNode(n *Node) (*Node, error) {
	client := r.newHTTPClient(true)

	result := Node{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(n).
		SetResult(&result).
		Post("/api/v1/nodes")
	if err != nil {
		return nil, fmt.Errorf("http.post err: %v", err)
	}
	if err := checkHTTP2XX(fmt.Sprintf("addNode(%v)", n), resp); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *remoteStore) UpdateNode(n *Node) error {
	client := r.newHTTPClient(true)
	resp, err := client.R().
		SetPathParam("nodeId", n.ID).
		SetHeader("Content-Type", "application/json").
		SetBody(n).
		Put("/api/v1/nodes/{nodeId}")
	if err != nil {
		return fmt.Errorf("http.put err: %v", err)
	}
	if err := checkHTTP2XX(fmt.Sprintf("UpdateNode(%v)", n), resp); err != nil {
		return err
	}
	return nil
}

func (r *remoteStore) SearchNodes(query string, limit int, kind Kind) ([]Node, error) {
	client := r.newHTTPClient(false)
	kindStr := ""
	if kind != Unknown {
		kindStr = string(kind)
	}
	req := client.R().
		SetQueryParams(map[string]string{
			//"Kind":  "PageOpen", // query either PageOpen or PageClose events for now
			"Q":       query,
			"Limit":   strconv.Itoa(limit),
			"Kind":    kindStr,
			"Include": "blocks",
		}).
		SetHeader("Accept", "application/json")

	resp, err := req.Get("/api/v1/search/nodes")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v req: %s", err, req.URL)
	}
	if err := checkHTTP2XX(fmt.Sprintf("searchNodes(%s, %d, %v)", query, limit, kind), resp); err != nil {
		return nil, err
	}

	result := getNodesResponse{N: []Node{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}

	return result.N, nil
}

func (r *remoteStore) UpdateNodeBlock(n *Node, b *Block) error {
	if !b.IsCode() {
		return fmt.Errorf("currently only code-blocks can be updated")
	}

	client := r.newHTTPClient(true)
	resp, err := client.R().
		SetPathParam("nodeId", n.ID).
		SetPathParam("snippetId", b.ID).
		SetHeader("Content-Type", "application/json").
		SetBody(b).
		Put("/api/v1/nodes/{nodeId}/snippet/{snippetId}")
	if err != nil {
		return fmt.Errorf("http.put err: %v", err)
	}
	if err := checkHTTP2XX(fmt.Sprintf("UpdateNodeBlock(%v)", n), resp); err != nil {
		return err
	}
	return nil
}

func (r *remoteStore) newHTTPClient(parseResponse bool) *resty.Client {
	skipSSLCheck := r.configEntry.SkipSSLCheck
	if skipSSLCheck {
		tools.Log.Warn().Msg("Warning: SSL check is disabled")
	}

	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: skipSSLCheck})
	client.SetDoNotParseResponse(!parseResponse)
	client.SetHostURL(r.configEntry.ServiceUrl)
	client.SetTimeout(clientTimeout)
	client.SetHeader("Authorization",
		fmt.Sprintf("%s %s", r.configEntry.AuthScheme, r.configEntry.User))
	client.SetHeader("Accept", "application/json")
	return client
}

func checkHTTP2XX(message string, resp *resty.Response) error {
	statusCode := resp.StatusCode()
	if statusCode == 401 {
		return &HttpError{
			Err:      errors.New("your credentials have expired: please run auth login"),
			HTTPCode: statusCode,
		}
	} else if statusCode < 200 || statusCode > 299 {
		if resp.RawResponse.ContentLength > 0 {
			errResp := ErrorResponse{}
			if err := json.NewDecoder(resp.RawBody()).Decode(&errResp); err == nil {
				tools.Log.Error().Interface("error", errResp).Msg(message)
				message += ": " + errResp.Message
			} else {
				body, err := ioutil.ReadAll(resp.RawResponse.Body)
				if err == nil {
					message += ": " + string(body)
				} else {
					tools.Log.Err(err).Msg("failed to read error response body")
				}
			}
		}
		return &HttpError{
			Err:      errors.New(message),
			HTTPCode: statusCode,
		}
	}
	return nil
}
