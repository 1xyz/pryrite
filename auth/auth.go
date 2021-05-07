package auth

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"regexp"
)

var (
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func AuthUser(entry *config.Entry, email string) error {
	// Ref: validate email -- https://golangcode.com/validate-an-email-address/
	if !isEmailValid(email) {
		return fmt.Errorf("email %v is not valid", email)
	}

	entry.User = email
	entry.AuthScheme = "Silly"
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

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
