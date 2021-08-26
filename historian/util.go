package historian

import (
	"time"

	"github.com/pkg/errors"
)

// returns a comparable date string (useful for time range keys)
func dateBytes(date time.Time) []byte {
	return []byte(date.UTC().Format(time.RFC3339Nano))
}

func wrap(err error, msg string) error {
	msg = "historian failed to " + msg
	if err == nil {
		return errors.New(msg)
	}

	return errors.Wrap(err, msg)
}
