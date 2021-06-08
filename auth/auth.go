package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1xyz/sseclient"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/cristalhq/jwt/v3"
)

const ClientTimeout = 30 * time.Second

func AuthUser(entry *config.Entry, serviceUrl string) error {
	tools.Log.Info().Msgf("AuthUser entry=%v, serviceUR=%s", entry, serviceUrl)
	if serviceUrl == "" {
		serviceUrl = entry.ServiceUrl
	}

	loginUrl, err := url.Parse(serviceUrl)
	if err != nil {
		return err
	}

	loginUrl.Path = "/api/v1/login"
	tools.Log.Info().Msgf("AuthUser serviceURL=%s loginURL=%s", serviceUrl, loginUrl)

	if entry.SkipSSLCheck {
		tools.Log.Warn().Msg("Warning: SSL check is disabled")
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: entry.SkipSSLCheck},
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 25 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}}

	events, err := sseclient.HttpClientOpenURL(client, loginUrl.String())
	//events, err := sseclient.OpenURL(loginUrl.String())
	if err != nil {
		return fmt.Errorf("authUser err = %v", err)
	}

	var tokenString string
	for event := range events {
		if url, ok := event.Data["authorize_url"]; ok {
			fmt.Println("Please open this link to authorize this app:", url)
		} else if t, ok := event.Data["bearer_token"]; ok {
			fmt.Println("Authorization is complete!")
			tokenString = strings.TrimSpace(t.(string))
			break
		} else {
			fmt.Println("Unknown event", event.Name, event.Data)
		}
	}

	// NOTE: this does _NOT_ verify since we don't download the issuer's key
	//       this is purely to snag the user info out of the jwt
	token, err := jwt.ParseString(tokenString)
	if err == nil {
		var claims map[string]interface{}
		err = json.Unmarshal(token.RawClaims(), &claims)
		if err != nil {
			return err
		}
		entry.Email = claims["email"].(string)
	} else {
		// FIXME: remove this when we stop supporting Silly auth
		entry.Email = tokenString
	}

	entry.ServiceUrl = serviceUrl
	entry.User = tokenString
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
