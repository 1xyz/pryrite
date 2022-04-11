package auth

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/1xyz/pryrite/tools"
	"github.com/zalando/go-keyring"
)

type Tokens struct {
	Id      string
	Refresh string
}

const secretsUser = "aardy-cli"

//--------------------------------------------------------------------------------

func SaveTokens(serviceUrl string, tokens *Tokens) error {
	service, err := getService(serviceUrl)
	if err != nil {
		return err
	}

	secrets := fmt.Sprintf("%s:%s", tokens.Id, tokens.Refresh)

	err = keyring.Set(service, secretsUser, secrets)
	if err != nil {
		tools.Log.Err(err).Msg("unable to set tokens in the keyring")
		return err
	}

	return nil
}

func GetTokens(serviceUrl string) (tokens *Tokens, err error) {
	service, err := getService(serviceUrl)
	if err != nil {
		return nil, err
	}

	secrets, err := keyring.Get(service, secretsUser)
	if err != nil {
		tools.Log.Err(err).Msg("unable to get tokens from the keyring")
		return nil, err
	}

	parts := strings.SplitN(secrets, ":", 2)
	return &Tokens{Id: parts[0], Refresh: parts[1]}, nil
}

func RemoveTokens(serviceUrl string) error {
	service, err := getService(serviceUrl)
	if err != nil {
		return err
	}

	err = keyring.Delete(service, secretsUser)
	if err != nil {
		tools.Log.Err(err).Msg("unable to delete tokens from the keyring")
	}

	return err
}

//--------------------------------------------------------------------------------

func getService(serviceUrl string) (string, error) {
	u, err := url.Parse(serviceUrl)
	if err != nil {
		tools.Log.Err(err).Str("serviceUrl", serviceUrl).Msg("unable to parse service URL for keyring")
		return "", err
	}

	return u.Hostname(), nil
}
