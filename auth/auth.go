package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"

	"github.com/cristalhq/jwt/v3"
	"github.com/pkg/errors"
)

const keyringInUseMarker = "-"

func AuthUser(entry *config.Entry) error {
	auth0 := NewAuth0(entry)

	if auth0 == nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("email? ")
		email, _ := reader.ReadString('\n')
		entry.Mode = "silly"
		entry.User = strings.TrimSpace(email)
		return config.SetEntry(entry)
	}

	url, err := auth0.GetAuthUrl()
	if err != nil {
		return err
	}

	fmt.Println("Please open this link to authorize this app:", url)

	tokens, err := auth0.GetTokens()
	if err != nil {
		return err
	}

	return processTokens(entry, tokens)
}

func LogoutUser(entry *config.Entry) error {
	if entry.User == keyringInUseMarker {
		RemoveTokens(entry.ServiceUrl)
	}

	entry.Email = ""
	entry.User = ""
	entry.AuthScheme = ""

	return config.SetEntry(entry)
}

func GetLoggedInUser(entry *config.Entry) (string, bool) {
	var tokens *Tokens

	if entry.User == keyringInUseMarker {
		var err error
		tokens, err = GetTokens(entry.ServiceUrl)
		if err != nil {
			return "", false
		}
	} else {
		tokens = &Tokens{Id: entry.User}
	}

	isExpired := false
	if entry.UserExpiresAt != nil {
		isExpired = time.Now().After(*entry.UserExpiresAt)
		if isExpired {
			if tokens.Refresh == "" {
				tools.Log.Info().Time("exp", *entry.UserExpiresAt).Msg("credentials have expired")
			} else {
				tools.Log.Info().Time("exp", *entry.UserExpiresAt).Msg("credentials have expired--refreshing")
				auth0 := NewAuth0(entry)
				if auth0 != nil {
					err := auth0.Refresh(tokens)
					if err == nil {
						err = processTokens(entry, tokens)
						if err != nil {
							tools.Log.Err(err).Msg("processing refresh tokens failed")
						}
						isExpired = time.Now().After(*entry.UserExpiresAt)
					} else {
						tools.LogStderr(err, "refresh attempt failed--will need to re-authorize this device")
					}
				}
			}
		}
	}

	return tokens.Id, tokens.Id != "" && !isExpired
}

func processTokens(entry *config.Entry, tokens *Tokens) error {
	// NOTE: this does _NOT_ verify since we don't download the issuer's key
	//       this is purely to snag the user info out of the jwt
	token, err := jwt.ParseString(tokens.Id)
	if err != nil {
		return errors.Wrap(err, "unable to parse the ID token")
	}

	var claims map[string]interface{}
	err = json.Unmarshal(token.RawClaims(), &claims)
	if err != nil {
		return errors.Wrap(err, "unable to get claims from the ID token")
	}

	entry.AuthScheme = "Bearer"
	entry.Email = claims["email"].(string)

	exp := claims["exp"].(float64)
	expTime := time.Unix(int64(math.Floor(exp)), 0)
	entry.UserExpiresAt = &expTime

	err = SaveTokens(entry.ServiceUrl, tokens)
	if err == nil {
		entry.User = keyringInUseMarker
	} else {
		println("Warning: no keyring manager found--storing your identity token insecurely")
		entry.User = tokens.Id
	}

	return config.SetEntry(entry)
}
