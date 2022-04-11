package slurp

import (
	"io"
	"net/url"
	"os"
	"time"

	"github.com/1xyz/pryrite/tools"
)

type Slurp struct {
	ExecutedAt     *time.Time
	ExecuteSeconds *float64
	Location       *url.URL
	Language       string
	Commandline    string
	ExitStatus     string
}

type Slurper interface {
	Slurp(shell string, reader io.Reader, digester Digester) error
}

type Digester func(*Slurp) error

//--------------------------------------------------------------------------------

type slurpFactory func() *Slurp

func newSlurpFactory(sourceScheme, language string) slurpFactory {
	baseLocation := url.URL{
		Scheme: sourceScheme,
	}

	var err error

	baseLocation.Host, err = os.Hostname()
	if err != nil {
		tools.Log.Err(err).Msg("Unable to get hostname")
		baseLocation.Host = "localhost"
	}

	return func() *Slurp {
		location := baseLocation // copy
		return &Slurp{
			Location: &location,
			Language: language,
		}
	}
}
