package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/events"
	"github.com/go-resty/resty/v2"
)

// {
//    "apiVersion": "v1beta1",
//    "timestampMilliseconds": 1616107997573,
//    "kind": "ClipboardCopy",
//    "metadata": {
//        "sessionId": "aslkdjf;lsdauvovjwlkmvweufolwejre",
//        "title": "My Fancy Pants"
//    },
//    "details": {
//        "xmlString": "<h3>Take These Important Steps</h3>"
//    }
//}

func main() {
	e := events.Event{
		Kind:    "PageOpen",
		Details: json.RawMessage("{\"hello\":\"world\"}"),
		Metadata: events.Metadata{
			SessionID: "abcde",
			Title:     "resty - golang REST ",
			URL:       "https://github.com/go-resty/resty",
		},
	}

	// Create a Resty Client
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resultE := events.Event{}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(&e).
		SetResult(&resultE). // or SetResult(AuthSuccess{}).
		Post("https://foobar.aardvarklabs.com:9443/api/v1/events")
	if err != nil {
		fmt.Println("err = ", err)
	}

	fmt.Printf("resp = %v", resp)
	fmt.Printf("resultE %v", string(resultE.Details))
}
