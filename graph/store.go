package graph

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/go-resty/resty/v2"
	"strconv"
)

type Store interface {
	// GetNodes returns the most recent n events
	GetNodes(n int) ([]Node, error)

	// GetNode returns the Node associated with this id
	GetNode(id string) (*Node, error)

	// AddNode adds a new node to this store
	AddNode(*Node) (*Node, error)

	// SearchNodes searches for nodes for provided query
	SearchNodes(string) ([]Node, error)
}

// remoteStore represents the remote event store backed by the service
type remoteStore struct {
	configEntry *config.Entry
}

func NewStore(configEntry *config.Entry) Store {
	return &remoteStore{
		configEntry: configEntry,
	}
}

type getEventsResponse struct {
	N []Node `json:"Nodes"`
}

func (r *remoteStore) GetNodes(n int) ([]Node, error) {
	client := r.newHTTPClient(false)
	req := client.R().
		SetQueryParams(map[string]string{
			//"Kind":  "PageOpen", // query either PageOpen or PageClose events for now
			"Limit": strconv.Itoa(n),
		}).
		SetHeader("Accept", "application/json")
	resp, err := req.Get("/api/v1/nodes")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v req: %s", err, req.URL)
	}

	result := getEventsResponse{N: []Node{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}
	if err := checkHTTP2XX(resp.StatusCode()); err != nil {
		return nil, fmt.Errorf("checkHTTP2XX url: %s err: %v", req.URL, err)
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
	result := getEventsResponse{N: []Node{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}
	if len(result.N) == 0 {
		return nil, fmt.Errorf("no entry found with id = %s", id)
	}
	if err := checkHTTP2XX(resp.StatusCode()); err != nil {
		return nil, fmt.Errorf("checkHTTP2XX url: %s err: %v", req.URL, err)
	}
	return &result.N[0], nil
}

func (r *remoteStore) AddNode(e *Node) (*Node, error) {
	client := r.newHTTPClient(true)

	result := Node{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(e).
		SetResult(&result).
		Post("/api/v1/nodes")
	if err != nil {
		return nil, fmt.Errorf("http.post err: %v", err)
	}
	if err := checkHTTP2XX(resp.StatusCode()); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *remoteStore) SearchNodes(query string) ([]Node, error) {
	client := r.newHTTPClient(false)
	const limit = 25
	req := client.R().
		SetQueryParams(map[string]string{
			//"Kind":  "PageOpen", // query either PageOpen or PageClose events for now
			"Q":     query,
			"Limit": strconv.Itoa(limit),
		}).
		SetHeader("Accept", "application/json")
	resp, err := req.Get("/api/v1/search/nodes")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v req: %s", err, req.URL)
	}

	result := getEventsResponse{N: []Node{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}
	if err := checkHTTP2XX(resp.StatusCode()); err != nil {
		return nil, fmt.Errorf("checkHTTP2XX url: %s err: %v", req.URL, err)
	}
	return result.N, nil
}

func (r *remoteStore) newHTTPClient(parseResponse bool) *resty.Client {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetDoNotParseResponse(!parseResponse)
	client.SetHostURL(r.configEntry.ServiceUrl)
	client.SetTimeout(clientTimeout)
	client.SetHeader("Authorization",
		fmt.Sprintf("%s %s", r.configEntry.AuthScheme, r.configEntry.User))
	client.SetHeader("Accept", "application/json")
	return client
}

func checkHTTP2XX(statusCode int) error {
	if statusCode < 200 || statusCode > 299 {
		return fmt.Errorf("service returned non-2XX status code %d", statusCode)
	}
	return nil
}
