package auth

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"net/url"

	"github.com/aardlabs/terminal-poc/config"
	"github.com/peterhellberg/sseclient"
)

func AuthUser(entry *config.Entry, serviceUrl string) error {
	if serviceUrl == "" {
		serviceUrl = entry.ServiceUrl
	}

	loginUrl, err := url.Parse(serviceUrl)
	if err != nil {
		return err
	}

	loginUrl.Path = "/api/v1/login"
	tools.Log.Info().Msgf("AuthUser serviceURL=%s loginURL=%s", serviceUrl, loginUrl)
	events, err := sseclient.OpenURL(loginUrl.String())
	if err != nil {
		return fmt.Errorf("authUser err = %v", err)
	}

	var token string
	for event := range events {
		if url, ok := event.Data["authorize_url"]; ok {
			fmt.Println("Please open this link to authorize this app:", url)
		} else if t, ok := event.Data["bearer_token"]; ok {
			fmt.Println("Authorization is complete!")
			token = t.(string)
			break
		} else {
			fmt.Println("Unknown event", event.Name, event.Data)
		}
	}

	entry.ServiceUrl = serviceUrl
	entry.User = token
	entry.AuthScheme = "Bearer"
	return config.SetEntry(entry)
}

func LogoutUser(entry *config.Entry) error {
	entry.User = ""
	entry.AuthScheme = ""
	return config.SetEntry(entry)
}

func GetLoggedInUser(entry *config.Entry) (string, bool) {
	return entry.User, entry.User != ""
}
