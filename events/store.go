package events

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)

type Store interface {
	// GetEvents returns the most recent n events
	GetEvents(n int) ([]Event, error)

	// Get the Event associated with this id
	GetEvent(id string) (*Event, error)

	// Add a new event to this store
	AddEvent(*Event) (*Event, error)
}

// remoteStore represents the remote event store backed by the service
type remoteStore struct {
	serviceUrl string
}

func NewStore(serviceUrl string) Store {
	return &remoteStore{
		serviceUrl: serviceUrl,
	}
}

type getEventsResponse struct {
	E []Event `json:"Events"`
}

func (r *remoteStore) GetEvents(n int) ([]Event, error) {
	client := r.newHTTPClient(false)
	resp, err := client.R().
		SetQueryParams(map[string]string{
			//"Kind":  "PageOpen", // query either PageOpen or PageClose events for now
			"Limit": strconv.Itoa(n),
		}).
		SetHeader("Accept", "application/json").
		Get("/api/v1/events")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v", err)
	}

	result := getEventsResponse{E: []Event{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}

	return result.E, nil
}

func (r *remoteStore) GetEvent(id string) (*Event, error) {
	client := r.newHTTPClient(false)
	resp, err := client.R().
		SetPathParam("eventId", id).
		SetHeader("Accept", "application/json").
		Get("/api/v1/events/{eventId}")
	if err != nil {
		return nil, fmt.Errorf("http.get err: %v", err)
	}
	result := getEventsResponse{E: []Event{}}
	if err := json.NewDecoder(resp.RawBody()).Decode(&result); err != nil {
		return nil, fmt.Errorf("json.Decode err: %v", err)
	}
	if len(result.E) == 0 {
		return nil, fmt.Errorf("no entry found with id = %s", id)
	}
	return &result.E[0], nil
}

func (r *remoteStore) AddEvent(e *Event) (*Event, error) {
	client := r.newHTTPClient(true)

	result := Event{}
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(e).
		SetResult(&result).
		Post("/api/v1/events")
	if err != nil {
		return nil, fmt.Errorf("http.post err: %v", err)
	}
	return &result, nil
}

func (r *remoteStore) newHTTPClient(parseResponse bool) *resty.Client {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetDoNotParseResponse(!parseResponse)
	client.SetHostURL(r.serviceUrl)
	client.SetTimeout(clientTimeout)
	return client
}
