package auth

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
)

var scopes = strings.Join([]string{"profile", "email", "openid", "offline_access"}, " ")

const grantType = "urn:ietf:params:oauth:grant-type:device_code"

//--------------------------------------------------------------------------------

type Auth0 struct {
	config *auth0Config
	client *resty.Client

	deviceCode   string
	pollInterval time.Duration
}

type auth0Config struct {
	baseUrl  url.URL
	clientId string
}

type genericResult map[string]interface{}

var auth0Configs = map[string]*auth0Config{
	"production": {
		baseUrl:  url.URL{Scheme: "https", Host: "aardy.us.auth0.com"},
		clientId: "yhuJdYfScDKexUYOxTRy5EGA0PhrJC8T",
	},
	"staging": {
		baseUrl:  url.URL{Scheme: "https", Host: "aardy-staging.us.auth0.com"},
		clientId: "y4SgI4bcU37hmaz28oyXdMurq2tQ8vDm",
	},
	"development": {
		baseUrl:  url.URL{Scheme: "https", Host: "dev-2fc95ql0.us.auth0.com"},
		clientId: "T8Tw6io5O8rGiLJBtKCrGGhS6PUMg580",
	},
}

//--------------------------------------------------------------------------------

func NewAuth0(entry *config.Entry) *Auth0 {
	if cfg, ok := auth0Configs[entry.Mode]; ok {
		return &Auth0{
			config: cfg,
			client: resty.New().SetHeaders(map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
				"User-Agent":   fmt.Sprintf("%s/%s (%s)", app.Name, app.Version, app.CommitHash),
			}),
		}
	}

	return nil
}

func (a0 *Auth0) GetAuthUrl() (string, error) {
	url := a0.config.baseUrl // copy of base
	url.Path = "/oauth/device/code"

	resp, err := a0.client.R().
		SetFormData(map[string]string{
			"client_id": a0.config.clientId,
			"scope":     scopes,
		}).
		SetResult(&genericResult{}).
		Post(url.String())
	if err != nil {
		return "", err
	}

	if !resp.IsSuccess() {
		return "", fmt.Errorf("unable to start device authorization: %v", resp)
	}

	res := *resp.Result().(*genericResult)

	a0.deviceCode = res["device_code"].(string)

	interval := res["interval"].(float64)
	a0.pollInterval = time.Duration(interval) * time.Second

	return res["verification_uri_complete"].(string), nil
}

func (a0 *Auth0) GetTokens() (*Tokens, error) {
	url := a0.config.baseUrl // copy of base
	url.Path = "/oauth/token"

	req := a0.client.R().
		SetFormData(map[string]string{
			"client_id":   a0.config.clientId,
			"grant_type":  grantType,
			"device_code": a0.deviceCode,
		}).
		SetResult(&genericResult{}).
		SetError(&genericResult{})

	for {
		resp, err := req.Post(url.String())
		if err != nil {
			return nil, err
		}

		statusClass := resp.StatusCode() / 100
		switch statusClass {
		case 2:
			tokens := extractTokens(resp.Result())
			return tokens, nil
		case 3:
			err = fmt.Errorf("unexpected redirect from Auth0: %s", resp.Status())
			return nil, err
		case 4:
			// logic handled below...
		case 5:
			err = fmt.Errorf("unexpected error from Auth0: %s %s", resp.Status(), resp)
			return nil, err
		default:
			err = fmt.Errorf("unknown HTTP status from Auth0: %s", resp.Status())
			return nil, err
		}

		// 4xx response...
		res := *resp.Error().(*genericResult)
		auth0Err := res["error"].(string)

		// https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#token-responses
		switch auth0Err {
		case "authorization_pending":
			time.Sleep(a0.pollInterval)
		case "slow_down":
			newInterval := float64(a0.pollInterval) * 1.2
			a0.pollInterval = time.Duration(newInterval)
			tools.Log.Warn().Interface("auth0Err", auth0Err).Dur("pollInterval", a0.pollInterval).Msg("slowed down")
		default:
			err = fmt.Errorf("%s: %s", auth0Err, res["error_description"])
			return nil, err
		}
	}
}

func (a0 *Auth0) Refresh(tokens *Tokens) error {
	url := a0.config.baseUrl // copy of base
	url.Path = "/oauth/token"

	resp, err := a0.client.R().
		SetFormData(map[string]string{
			"client_id":     a0.config.clientId,
			"grant_type":    "refresh_token",
			"refresh_token": tokens.Refresh,
		}).
		SetResult(&genericResult{}).
		Post(url.String())
	if err != nil {
		return errors.Wrap(err, "unable to refresh device authorization")
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("unable to refresh device authorization: %v", resp)
	}

	newTokens := extractTokens(resp.Result())
	*tokens = *newTokens

	return nil
}

func extractTokens(respRes interface{}) *Tokens {
	res := *respRes.(*genericResult)
	return &Tokens{
		Id:      res["id_token"].(string),
		Refresh: res["refresh_token"].(string),
	}
}
