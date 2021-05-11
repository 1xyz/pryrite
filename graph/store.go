package graph

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/go-resty/resty/v2"
	"strconv"
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
			"Limit": strconv.Itoa(limit),
			"Kind":  kindStr,
		})
	resp, err := req.Get("/api/v1/nodes")
	if err != nil {
		return nil, err
	}
	if err := checkHTTP2XX("getNodes(%s)", resp.StatusCode()); err != nil {
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
		SetHeader("Accept", "application/json")
	resp, err := req.Get("/api/v1/nodes/{nodeId}")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v", err)
	}
	if err := checkHTTP2XX(fmt.Sprintf("getNode(%s)", id), resp.StatusCode()); err != nil {
		return nil, err
	}
	result := Node{}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, &Error{"GetNode", err}
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
	if err := checkHTTP2XX(fmt.Sprintf("addNode(%v)", n), resp.StatusCode()); err != nil {
		return nil, err
	}
	return &result, nil
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
			"Q":     query,
			"Limit": strconv.Itoa(limit),
			"Kind":  kindStr,
		}).
		SetHeader("Accept", "application/json")

	resp, err := req.Get("/api/v1/search/nodes")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v req: %s", err, req.URL)
	}
	if err := checkHTTP2XX(fmt.Sprintf("searchNodes(%s, %d, %v)", query, limit, kind), resp.StatusCode()); err != nil {
		return nil, err
	}

	result := getNodesResponse{N: []Node{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}

	return result.N, nil
}

func (r *remoteStore) newHTTPClient(parseResponse bool) *resty.Client {
	skipSSLCheck := r.configEntry.SkipSSLCheck
	if skipSSLCheck {
		tools.LogStdout("Warning! SSL check is disabled")
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

func checkHTTP2XX(message string, statusCode int) error {
	if statusCode < 200 || statusCode > 299 {
		return &HttpError{
			Err:      errors.New(message),
			HTTPCode: statusCode,
		}
	}
	return nil
}
