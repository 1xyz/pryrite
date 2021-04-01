package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/events"
	"github.com/go-resty/resty/v2"
	"io"
	"log"
	"time"
)

type eventsResult struct {
	E []events.Event `json:"Events"`
}

func fromJSON(rdr io.Reader) (*eventsResult, error) {
	result := eventsResult{E: []events.Event{}}
	dec := json.NewDecoder(rdr)
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	// Create a Resty Client
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	client.SetDoNotParseResponse(true)
	client.SetHostURL("https://foobar.aardvarklabs.com:9443")

	// curl -vvv https://foobar.aardvarklabs.com:9443/api/v1/events\?Kind\=PageOpen\&Limit\=2
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"Kind":  "PageOpen",
			"Limit": "25",
		}).
		SetHeader("Accept", "application/json").
		Get("api/v1/events")
	if err != nil {
		log.Fatalf("err = %v", err)
	}

	result, err := fromJSON(resp.RawBody())
	if err != nil {
		log.Fatalf("fromJson err = %v", err)
	}
	for _, e := range result.E {
		fmt.Printf("%v %v %s %s\n", e.ID, e.CreatedAt.Format(time.Stamp), e.Metadata.Title, e.Kind)
	}

	//b, err := ioutil.ReadAll(resp.RawBody())
	//if err != nil {
	//	log.Fatalf("Err = %v", err)
	//}
	//fmt.Printf("%s", string(b))
	//fmt.Printf("query took %v", time.Since(st))
}
